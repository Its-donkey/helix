package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetEventSubSubscriptions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eventsub/subscriptions" {
			t.Errorf("expected /eventsub/subscriptions, got %s", r.URL.Path)
		}

		resp := EventSubResponse{
			Data: []EventSubSubscription{
				{
					ID:      "sub1",
					Status:  "enabled",
					Type:    "channel.follow",
					Version: "2",
					Condition: map[string]string{
						"broadcaster_user_id": "12345",
						"moderator_user_id":   "12345",
					},
					CreatedAt: time.Now(),
					Transport: EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/webhook",
					},
					Cost: 1,
				},
				{
					ID:      "sub2",
					Status:  "enabled",
					Type:    "stream.online",
					Version: "1",
					Condition: map[string]string{
						"broadcaster_user_id": "12345",
					},
					CreatedAt: time.Now(),
					Transport: EventSubTransport{
						Method:    "websocket",
						SessionID: "session123",
					},
					Cost: 1,
				},
			},
			Total:        2,
			TotalCost:    2,
			MaxTotalCost: 10000,
			Pagination:   &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetEventSubSubscriptions(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(resp.Data))
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
	if resp.TotalCost != 2 {
		t.Errorf("expected total_cost 2, got %d", resp.TotalCost)
	}
}

func TestClient_GetEventSubSubscriptions_WithFilters(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		subType := r.URL.Query().Get("type")
		userID := r.URL.Query().Get("user_id")

		if status != "enabled" {
			t.Errorf("expected status=enabled, got %s", status)
		}
		if subType != "channel.follow" {
			t.Errorf("expected type=channel.follow, got %s", subType)
		}
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := EventSubResponse{
			Data:         []EventSubSubscription{},
			Total:        0,
			TotalCost:    0,
			MaxTotalCost: 10000,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetEventSubSubscriptions(context.Background(), &GetEventSubSubscriptionsParams{
		Status: "enabled",
		Type:   "channel.follow",
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateEventSubSubscription_Webhook(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/eventsub/subscriptions" {
			t.Errorf("expected /eventsub/subscriptions, got %s", r.URL.Path)
		}

		var params CreateEventSubSubscriptionParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Type != "channel.follow" {
			t.Errorf("expected type channel.follow, got %s", params.Type)
		}
		if params.Version != "2" {
			t.Errorf("expected version 2, got %s", params.Version)
		}
		if params.Transport.Method != "webhook" {
			t.Errorf("expected method webhook, got %s", params.Transport.Method)
		}
		if params.Transport.Callback != "https://example.com/webhook" {
			t.Errorf("unexpected callback: %s", params.Transport.Callback)
		}

		resp := EventSubResponse{
			Data: []EventSubSubscription{
				{
					ID:        "newsub",
					Status:    "webhook_callback_verification_pending",
					Type:      params.Type,
					Version:   params.Version,
					Condition: params.Condition,
					Transport: EventSubTransport{
						Method:   params.Transport.Method,
						Callback: params.Transport.Callback,
					},
					Cost:      1,
					CreatedAt: time.Now(),
				},
			},
			Total:        1,
			TotalCost:    1,
			MaxTotalCost: 10000,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateEventSubSubscription(context.Background(), &CreateEventSubSubscriptionParams{
		Type:    "channel.follow",
		Version: "2",
		Condition: map[string]string{
			"broadcaster_user_id": "12345",
			"moderator_user_id":   "12345",
		},
		Transport: CreateEventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/webhook",
			Secret:   "mysecret",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != "newsub" {
		t.Errorf("expected subscription ID 'newsub', got %s", result.ID)
	}
	if result.Status != "webhook_callback_verification_pending" {
		t.Errorf("expected pending status, got %s", result.Status)
	}
}

func TestClient_CreateEventSubSubscription_WebSocket(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params CreateEventSubSubscriptionParams
		_ = json.NewDecoder(r.Body).Decode(&params)

		if params.Transport.Method != "websocket" {
			t.Errorf("expected method websocket, got %s", params.Transport.Method)
		}
		if params.Transport.SessionID != "session123" {
			t.Errorf("expected session_id session123, got %s", params.Transport.SessionID)
		}

		resp := EventSubResponse{
			Data: []EventSubSubscription{
				{
					ID:      "wssub",
					Status:  "enabled",
					Type:    params.Type,
					Version: params.Version,
					Transport: EventSubTransport{
						Method:    "websocket",
						SessionID: params.Transport.SessionID,
					},
					Cost: 1,
				},
			},
			Total:        1,
			TotalCost:    1,
			MaxTotalCost: 10000,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateEventSubSubscription(context.Background(), &CreateEventSubSubscriptionParams{
		Type:    "stream.online",
		Version: "1",
		Condition: map[string]string{
			"broadcaster_user_id": "12345",
		},
		Transport: CreateEventSubTransport{
			Method:    "websocket",
			SessionID: "session123",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Transport.Method != "websocket" {
		t.Errorf("expected websocket, got %s", result.Transport.Method)
	}
}

func TestClient_DeleteEventSubSubscription(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/eventsub/subscriptions" {
			t.Errorf("expected /eventsub/subscriptions, got %s", r.URL.Path)
		}

		subID := r.URL.Query().Get("id")
		if subID != "sub123" {
			t.Errorf("expected id=sub123, got %s", subID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteEventSubSubscription(context.Background(), "sub123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEventSubTypeConstants(t *testing.T) {
	// Verify some common EventSub type constants
	tests := []struct {
		constant string
		expected string
	}{
		{EventSubTypeChannelFollow, "channel.follow"},
		{EventSubTypeStreamOnline, "stream.online"},
		{EventSubTypeStreamOffline, "stream.offline"},
		{EventSubTypeChannelSubscribe, "channel.subscribe"},
		{EventSubTypeChannelCheer, "channel.cheer"},
		{EventSubTypeChannelRaid, "channel.raid"},
		{EventSubTypeChannelBan, "channel.ban"},
		{EventSubTypeChannelUnban, "channel.unban"},
		{EventSubTypeUserUpdate, "user.update"},
	}

	for _, tc := range tests {
		if tc.constant != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, tc.constant)
		}
	}
}
