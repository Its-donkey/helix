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

func TestGetEventSubVersion(t *testing.T) {
	// Test known types with specific versions
	if v := GetEventSubVersion(EventSubTypeChannelFollow); v != "2" {
		t.Errorf("expected version 2 for channel.follow, got %s", v)
	}
	if v := GetEventSubVersion(EventSubTypeStreamOnline); v != "1" {
		t.Errorf("expected version 1 for stream.online, got %s", v)
	}
	if v := GetEventSubVersion(EventSubTypeAutomodMessageHold); v != "2" {
		t.Errorf("expected version 2 for automod.message.hold, got %s", v)
	}

	// Test unknown type returns default "1"
	if v := GetEventSubVersion("unknown.type"); v != "1" {
		t.Errorf("expected version 1 for unknown type, got %s", v)
	}
}

func TestBroadcasterCondition(t *testing.T) {
	cond := BroadcasterCondition("12345")

	if cond["broadcaster_user_id"] != "12345" {
		t.Errorf("expected broadcaster_user_id=12345, got %s", cond["broadcaster_user_id"])
	}
	if len(cond) != 1 {
		t.Errorf("expected 1 key, got %d", len(cond))
	}
}

func TestBroadcasterModeratorCondition(t *testing.T) {
	cond := BroadcasterModeratorCondition("12345", "67890")

	if cond["broadcaster_user_id"] != "12345" {
		t.Errorf("expected broadcaster_user_id=12345, got %s", cond["broadcaster_user_id"])
	}
	if cond["moderator_user_id"] != "67890" {
		t.Errorf("expected moderator_user_id=67890, got %s", cond["moderator_user_id"])
	}
	if len(cond) != 2 {
		t.Errorf("expected 2 keys, got %d", len(cond))
	}
}

func TestUserCondition(t *testing.T) {
	cond := UserCondition("12345")

	if cond["user_id"] != "12345" {
		t.Errorf("expected user_id=12345, got %s", cond["user_id"])
	}
	if len(cond) != 1 {
		t.Errorf("expected 1 key, got %d", len(cond))
	}
}

func TestFromToBroadcasterCondition(t *testing.T) {
	// Test with both values
	cond := FromToBroadcasterCondition("12345", "67890")
	if cond["from_broadcaster_user_id"] != "12345" {
		t.Errorf("expected from_broadcaster_user_id=12345, got %s", cond["from_broadcaster_user_id"])
	}
	if cond["to_broadcaster_user_id"] != "67890" {
		t.Errorf("expected to_broadcaster_user_id=67890, got %s", cond["to_broadcaster_user_id"])
	}

	// Test with only from
	cond = FromToBroadcasterCondition("12345", "")
	if cond["from_broadcaster_user_id"] != "12345" {
		t.Errorf("expected from_broadcaster_user_id=12345, got %s", cond["from_broadcaster_user_id"])
	}
	if _, ok := cond["to_broadcaster_user_id"]; ok {
		t.Error("to_broadcaster_user_id should not be set")
	}

	// Test with only to
	cond = FromToBroadcasterCondition("", "67890")
	if _, ok := cond["from_broadcaster_user_id"]; ok {
		t.Error("from_broadcaster_user_id should not be set")
	}
	if cond["to_broadcaster_user_id"] != "67890" {
		t.Errorf("expected to_broadcaster_user_id=67890, got %s", cond["to_broadcaster_user_id"])
	}
}

func TestClientCondition(t *testing.T) {
	cond := ClientCondition("myClientId")

	if cond["client_id"] != "myClientId" {
		t.Errorf("expected client_id=myClientId, got %s", cond["client_id"])
	}
	if len(cond) != 1 {
		t.Errorf("expected 1 key, got %d", len(cond))
	}
}

func TestConduitCondition(t *testing.T) {
	cond := ConduitCondition("conduit123")

	if cond["conduit_id"] != "conduit123" {
		t.Errorf("expected conduit_id=conduit123, got %s", cond["conduit_id"])
	}
	if len(cond) != 1 {
		t.Errorf("expected 1 key, got %d", len(cond))
	}
}

func TestRewardCondition(t *testing.T) {
	// Test with reward ID
	cond := RewardCondition("12345", "reward123")
	if cond["broadcaster_user_id"] != "12345" {
		t.Errorf("expected broadcaster_user_id=12345, got %s", cond["broadcaster_user_id"])
	}
	if cond["reward_id"] != "reward123" {
		t.Errorf("expected reward_id=reward123, got %s", cond["reward_id"])
	}

	// Test without reward ID
	cond = RewardCondition("12345", "")
	if cond["broadcaster_user_id"] != "12345" {
		t.Errorf("expected broadcaster_user_id=12345, got %s", cond["broadcaster_user_id"])
	}
	if _, ok := cond["reward_id"]; ok {
		t.Error("reward_id should not be set")
	}
}

func TestClient_SubscribeToChannel(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params CreateEventSubSubscriptionParams
		_ = json.NewDecoder(r.Body).Decode(&params)

		if params.Type != "stream.online" {
			t.Errorf("expected type stream.online, got %s", params.Type)
		}
		if params.Version != "1" {
			t.Errorf("expected version 1, got %s", params.Version)
		}
		if params.Condition["broadcaster_user_id"] != "12345" {
			t.Errorf("expected broadcaster_user_id=12345, got %s", params.Condition["broadcaster_user_id"])
		}

		resp := EventSubResponse{
			Data: []EventSubSubscription{
				{
					ID:        "sub123",
					Status:    "enabled",
					Type:      params.Type,
					Version:   params.Version,
					Condition: params.Condition,
					Cost:      1,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SubscribeToChannel(context.Background(), "stream.online", "12345", CreateEventSubTransport{
		Method:    "websocket",
		SessionID: "session123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "sub123" {
		t.Errorf("expected subscription ID sub123, got %s", result.ID)
	}
}

func TestClient_SubscribeToChannelWithModerator(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params CreateEventSubSubscriptionParams
		_ = json.NewDecoder(r.Body).Decode(&params)

		if params.Type != "channel.follow" {
			t.Errorf("expected type channel.follow, got %s", params.Type)
		}
		if params.Version != "2" {
			t.Errorf("expected version 2, got %s", params.Version)
		}
		if params.Condition["broadcaster_user_id"] != "12345" {
			t.Errorf("expected broadcaster_user_id=12345, got %s", params.Condition["broadcaster_user_id"])
		}
		if params.Condition["moderator_user_id"] != "67890" {
			t.Errorf("expected moderator_user_id=67890, got %s", params.Condition["moderator_user_id"])
		}

		resp := EventSubResponse{
			Data: []EventSubSubscription{
				{
					ID:        "sub123",
					Status:    "enabled",
					Type:      params.Type,
					Version:   params.Version,
					Condition: params.Condition,
					Cost:      1,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SubscribeToChannelWithModerator(context.Background(), "channel.follow", "12345", "67890", CreateEventSubTransport{
		Method:    "websocket",
		SessionID: "session123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "sub123" {
		t.Errorf("expected subscription ID sub123, got %s", result.ID)
	}
}

func TestClient_SubscribeToUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params CreateEventSubSubscriptionParams
		_ = json.NewDecoder(r.Body).Decode(&params)

		if params.Type != "user.update" {
			t.Errorf("expected type user.update, got %s", params.Type)
		}
		if params.Condition["user_id"] != "12345" {
			t.Errorf("expected user_id=12345, got %s", params.Condition["user_id"])
		}

		resp := EventSubResponse{
			Data: []EventSubSubscription{
				{
					ID:        "sub123",
					Status:    "enabled",
					Type:      params.Type,
					Version:   params.Version,
					Condition: params.Condition,
					Cost:      1,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SubscribeToUser(context.Background(), "user.update", "12345", CreateEventSubTransport{
		Method:    "websocket",
		SessionID: "session123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "sub123" {
		t.Errorf("expected subscription ID sub123, got %s", result.ID)
	}
}

func TestClient_GetAllSubscriptions(t *testing.T) {
	callCount := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp EventSubResponse
		if callCount == 1 {
			resp = EventSubResponse{
				Data: []EventSubSubscription{
					{ID: "sub1", Status: "enabled"},
					{ID: "sub2", Status: "enabled"},
				},
				Pagination: &Pagination{Cursor: "cursor1"},
			}
		} else {
			resp = EventSubResponse{
				Data: []EventSubSubscription{
					{ID: "sub3", Status: "enabled"},
				},
				Pagination: nil,
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	subs, err := client.GetAllSubscriptions(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 3 {
		t.Errorf("expected 3 subscriptions, got %d", len(subs))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestClient_GetAllSubscriptions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := client.GetAllSubscriptions(context.Background(), nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestClient_DeleteAllSubscriptions(t *testing.T) {
	callCount := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := EventSubResponse{
				Data: []EventSubSubscription{
					{ID: "sub1", Status: "enabled"},
					{ID: "sub2", Status: "enabled"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			callCount++
			w.WriteHeader(http.StatusNoContent)
		}
	})
	defer server.Close()

	deleted, err := client.DeleteAllSubscriptions(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", deleted)
	}
	if callCount != 2 {
		t.Errorf("expected 2 delete calls, got %d", callCount)
	}
}

func TestClient_DeleteAllSubscriptions_GetError(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := client.DeleteAllSubscriptions(context.Background(), nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestClient_DeleteAllSubscriptions_DeleteError(t *testing.T) {
	deleteCallCount := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := EventSubResponse{
				Data: []EventSubSubscription{
					{ID: "sub1", Status: "enabled"},
					{ID: "sub2", Status: "enabled"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			deleteCallCount++
			if deleteCallCount == 1 {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	})
	defer server.Close()

	deleted, err := client.DeleteAllSubscriptions(context.Background(), nil)
	if err == nil {
		t.Error("expected error on second delete")
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted before error, got %d", deleted)
	}
}

func TestClient_CreateEventSubSubscription_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request","message":"invalid condition"}`))
	})
	defer server.Close()

	_, err := client.CreateEventSubSubscription(context.Background(), &CreateEventSubSubscriptionParams{
		Type:    "channel.follow",
		Version: "2",
		Condition: map[string]string{
			"broadcaster_user_id": "12345",
		},
		Transport: CreateEventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/webhook",
			Secret:   "secret",
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateEventSubSubscription_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := EventSubResponse{
			Data: []EventSubSubscription{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateEventSubSubscription(context.Background(), &CreateEventSubSubscriptionParams{
		Type:    "channel.follow",
		Version: "2",
		Condition: map[string]string{
			"broadcaster_user_id": "12345",
		},
		Transport: CreateEventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/webhook",
			Secret:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for empty response")
	}
}

func TestClient_GetEventSubSubscriptions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetEventSubSubscriptions(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_DeleteEventSubSubscription_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	err := client.DeleteEventSubSubscription(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
