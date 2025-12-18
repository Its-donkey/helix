package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetChannelInformation(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/channels" {
			t.Errorf("expected /channels, got %s", r.URL.Path)
		}

		ids := r.URL.Query()["broadcaster_id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 broadcaster_ids, got %d", len(ids))
		}

		resp := Response[Channel]{
			Data: []Channel{
				{
					BroadcasterID:       "12345",
					BroadcasterLogin:    "channel1",
					BroadcasterName:     "Channel1",
					BroadcasterLanguage: "en",
					GameID:              "game123",
					GameName:            "Test Game",
					Title:               "Test Stream",
					Delay:               0,
					Tags:                []string{"English", "Gaming"},
				},
				{
					BroadcasterID:    "67890",
					BroadcasterLogin: "channel2",
					BroadcasterName:  "Channel2",
					Title:            "Another Stream",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelInformation(context.Background(), &GetChannelInformationParams{
		BroadcasterIDs: []string{"12345", "67890"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "Test Stream" {
		t.Errorf("expected title 'Test Stream', got %s", resp.Data[0].Title)
	}
}

func TestClient_ModifyChannelInformation(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/channels" {
			t.Errorf("expected /channels, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id 12345, got %s", broadcasterID)
		}

		var body ModifyChannelInformationParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Title != "New Title" {
			t.Errorf("expected title 'New Title', got %s", body.Title)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.ModifyChannelInformation(context.Background(), &ModifyChannelInformationParams{
		BroadcasterID: "12345",
		Title:         "New Title",
		GameID:        "game456",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetChannelEditors(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/editors" {
			t.Errorf("expected /channels/editors, got %s", r.URL.Path)
		}

		resp := Response[ChannelEditor]{
			Data: []ChannelEditor{
				{
					UserID:    "11111",
					UserName:  "Editor1",
					CreatedAt: time.Now(),
				},
				{
					UserID:    "22222",
					UserName:  "Editor2",
					CreatedAt: time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelEditors(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 editors, got %d", len(resp.Data))
	}
}

func TestClient_GetFollowedChannels(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/followed" {
			t.Errorf("expected /channels/followed, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id 12345, got %s", userID)
		}

		resp := Response[FollowedChannel]{
			Data: []FollowedChannel{
				{
					BroadcasterID:    "11111",
					BroadcasterLogin: "followed1",
					BroadcasterName:  "Followed1",
					FollowedAt:       time.Now(),
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		total := 100
		resp.Total = &total
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetFollowedChannels(context.Background(), &GetFollowedChannelsParams{
		UserID: "12345",
		PaginationParams: &PaginationParams{
			First: 20,
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 followed channel, got %d", len(resp.Data))
	}
	if resp.Pagination == nil || resp.Pagination.Cursor != "next-cursor" {
		t.Error("expected pagination cursor")
	}
}

func TestClient_GetFollowedChannels_WithBroadcasterFilter(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "99999" {
			t.Errorf("expected broadcaster_id 99999, got %s", broadcasterID)
		}

		resp := Response[FollowedChannel]{
			Data: []FollowedChannel{
				{
					BroadcasterID: "99999",
					FollowedAt:    time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetFollowedChannels(context.Background(), &GetFollowedChannelsParams{
		UserID:        "12345",
		BroadcasterID: "99999",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 followed channel, got %d", len(resp.Data))
	}
}

func TestClient_GetChannelFollowers(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/followers" {
			t.Errorf("expected /channels/followers, got %s", r.URL.Path)
		}

		resp := Response[ChannelFollower]{
			Data: []ChannelFollower{
				{
					UserID:     "11111",
					UserLogin:  "follower1",
					UserName:   "Follower1",
					FollowedAt: time.Now(),
				},
				{
					UserID:     "22222",
					UserLogin:  "follower2",
					UserName:   "Follower2",
					FollowedAt: time.Now(),
				},
			},
		}
		total := 1000
		resp.Total = &total
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelFollowers(context.Background(), &GetChannelFollowersParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 followers, got %d", len(resp.Data))
	}
}

func TestClient_GetVIPs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/vips" {
			t.Errorf("expected /channels/vips, got %s", r.URL.Path)
		}

		resp := Response[VIP]{
			Data: []VIP{
				{UserID: "11111", UserLogin: "vip1", UserName: "VIP1"},
				{UserID: "22222", UserLogin: "vip2", UserName: "VIP2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetVIPs(context.Background(), &GetVIPsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 VIPs, got %d", len(resp.Data))
	}
}

func TestClient_AddChannelVIP(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/channels/vips" {
			t.Errorf("expected /channels/vips, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		userID := r.URL.Query().Get("user_id")

		if broadcasterID != "12345" || userID != "67890" {
			t.Errorf("expected broadcaster_id=12345, user_id=67890, got %s, %s", broadcasterID, userID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.AddChannelVIP(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_RemoveChannelVIP(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/channels/vips" {
			t.Errorf("expected /channels/vips, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.RemoveChannelVIP(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetChannelInformation_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetChannelInformation(context.Background(), &GetChannelInformationParams{BroadcasterIDs: []string{"12345"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetChannelEditors_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetChannelEditors(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetFollowedChannels_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetFollowedChannels(context.Background(), &GetFollowedChannelsParams{UserID: "12345"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetChannelFollowers_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetChannelFollowers(context.Background(), &GetChannelFollowersParams{BroadcasterID: "12345"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetVIPs_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetVIPs(context.Background(), &GetVIPsParams{BroadcasterID: "12345"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetChannelFollowers_WithUserID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID != "67890" {
			t.Errorf("expected user_id=67890, got %s", userID)
		}

		resp := Response[ChannelFollower]{
			Data: []ChannelFollower{
				{UserID: "67890", UserLogin: "follower1", UserName: "Follower1", FollowedAt: time.Now()},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelFollowers(context.Background(), &GetChannelFollowersParams{
		BroadcasterID: "12345",
		UserID:        "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 follower, got %d", len(resp.Data))
	}
}

func TestClient_GetVIPs_WithUserIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userIDs := r.URL.Query()["user_id"]
		if len(userIDs) != 2 {
			t.Errorf("expected 2 user_ids, got %d", len(userIDs))
		}

		resp := Response[VIP]{
			Data: []VIP{
				{UserID: "11111", UserLogin: "vip1", UserName: "VIP1"},
				{UserID: "22222", UserLogin: "vip2", UserName: "VIP2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetVIPs(context.Background(), &GetVIPsParams{
		BroadcasterID: "12345",
		UserIDs:       []string{"11111", "22222"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 VIPs, got %d", len(resp.Data))
	}
}
