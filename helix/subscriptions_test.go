package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetBroadcasterSubscriptions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions" {
			t.Errorf("expected /subscriptions, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := SubscriptionsResponse{
			Data: []Subscription{
				{
					BroadcasterID:    "12345",
					BroadcasterLogin: "broadcaster",
					BroadcasterName:  "Broadcaster",
					UserID:           "67890",
					UserLogin:        "subscriber",
					UserName:         "Subscriber",
					Tier:             "1000",
					PlanName:         "Channel Subscription",
					IsGift:           false,
				},
				{
					BroadcasterID: "12345",
					UserID:        "11111",
					Tier:          "3000",
					IsGift:        true,
					GifterID:      "22222",
					GifterLogin:   "gifter",
					GifterName:    "Gifter",
				},
			},
			Total:  100,
			Points: 150,
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBroadcasterSubscriptions(context.Background(), &GetBroadcasterSubscriptionsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(resp.Data))
	}
	if resp.Total != 100 {
		t.Errorf("expected total 100, got %d", resp.Total)
	}
	if resp.Points != 150 {
		t.Errorf("expected points 150, got %d", resp.Points)
	}
	if resp.Data[1].IsGift != true {
		t.Error("expected second subscription to be a gift")
	}
}

func TestClient_GetBroadcasterSubscriptions_FilterByUsers(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userIDs := r.URL.Query()["user_id"]
		if len(userIDs) != 2 {
			t.Errorf("expected 2 user_ids, got %d", len(userIDs))
		}

		resp := SubscriptionsResponse{
			Data: []Subscription{
				{UserID: "user1", Tier: "1000"},
				{UserID: "user2", Tier: "2000"},
			},
			Total: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBroadcasterSubscriptions(context.Background(), &GetBroadcasterSubscriptionsParams{
		BroadcasterID: "12345",
		UserIDs:       []string{"user1", "user2"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(resp.Data))
	}
}

func TestClient_GetBroadcasterSubscriptions_WithPagination(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		first := r.URL.Query().Get("first")
		after := r.URL.Query().Get("after")

		if first != "50" {
			t.Errorf("expected first=50, got %s", first)
		}
		if after != "cursor123" {
			t.Errorf("expected after=cursor123, got %s", after)
		}

		resp := SubscriptionsResponse{
			Data:       []Subscription{},
			Total:      0,
			Pagination: &Pagination{Cursor: "nextcursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetBroadcasterSubscriptions(context.Background(), &GetBroadcasterSubscriptionsParams{
		BroadcasterID: "12345",
		PaginationParams: &PaginationParams{
			First: 50,
			After: "cursor123",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CheckUserSubscription(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions/user" {
			t.Errorf("expected /subscriptions/user, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		userID := r.URL.Query().Get("user_id")

		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}
		if userID != "67890" {
			t.Errorf("expected user_id=67890, got %s", userID)
		}

		resp := Response[UserSubscription]{
			Data: []UserSubscription{
				{
					BroadcasterID:    "12345",
					BroadcasterLogin: "broadcaster",
					BroadcasterName:  "Broadcaster",
					Tier:             "2000",
					IsGift:           false,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	sub, err := client.CheckUserSubscription(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub == nil {
		t.Fatal("expected subscription, got nil")
	}
	if sub.Tier != "2000" {
		t.Errorf("expected tier 2000, got %s", sub.Tier)
	}
}

func TestClient_CheckUserSubscription_NotSubscribed(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[UserSubscription]{
			Data: []UserSubscription{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	sub, err := client.CheckUserSubscription(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub != nil {
		t.Errorf("expected nil subscription, got %+v", sub)
	}
}

func TestClient_CheckUserSubscription_GiftedSub(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[UserSubscription]{
			Data: []UserSubscription{
				{
					BroadcasterID: "12345",
					Tier:          "1000",
					IsGift:        true,
					GifterID:      "gifter123",
					GifterLogin:   "giftgiver",
					GifterName:    "Gift Giver",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	sub, err := client.CheckUserSubscription(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sub.IsGift {
		t.Error("expected subscription to be a gift")
	}
	if sub.GifterID != "gifter123" {
		t.Errorf("expected gifter ID gifter123, got %s", sub.GifterID)
	}
}
