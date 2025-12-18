package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetCreatorGoals(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/goals" {
			t.Errorf("expected /goals, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[CreatorGoal]{
			Data: []CreatorGoal{
				{
					ID:               "goal1",
					BroadcasterID:    "12345",
					BroadcasterName:  "Streamer",
					BroadcasterLogin: "streamer",
					Type:             "follower",
					Description:      "Reach 10000 followers!",
					CurrentAmount:    8500,
					TargetAmount:     10000,
					CreatedAt:        time.Now().Add(-24 * time.Hour),
				},
				{
					ID:               "goal2",
					BroadcasterID:    "12345",
					BroadcasterName:  "Streamer",
					BroadcasterLogin: "streamer",
					Type:             "subscription",
					Description:      "Get 500 subs this month!",
					CurrentAmount:    350,
					TargetAmount:     500,
					CreatedAt:        time.Now().Add(-7 * 24 * time.Hour),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCreatorGoals(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 goals, got %d", len(resp.Data))
	}
	if resp.Data[0].Type != "follower" {
		t.Errorf("expected type follower, got %s", resp.Data[0].Type)
	}
	if resp.Data[0].CurrentAmount != 8500 {
		t.Errorf("expected current_amount 8500, got %d", resp.Data[0].CurrentAmount)
	}
	if resp.Data[0].TargetAmount != 10000 {
		t.Errorf("expected target_amount 10000, got %d", resp.Data[0].TargetAmount)
	}
}

func TestClient_GetCreatorGoals_NoGoals(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[CreatorGoal]{
			Data: []CreatorGoal{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCreatorGoals(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 goals, got %d", len(resp.Data))
	}
}

func TestClient_GetCreatorGoals_DifferentTypes(t *testing.T) {
	goalTypes := []string{"follower", "subscription", "subscription_count", "new_subscription", "new_subscription_count"}

	for _, goalType := range goalTypes {
		t.Run(goalType, func(t *testing.T) {
			client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
				resp := Response[CreatorGoal]{
					Data: []CreatorGoal{
						{
							ID:            "goal1",
							Type:          goalType,
							CurrentAmount: 100,
							TargetAmount:  200,
						},
					},
				}
				_ = json.NewEncoder(w).Encode(resp)
			})
			defer server.Close()

			resp, err := client.GetCreatorGoals(context.Background(), "12345")

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Data[0].Type != goalType {
				t.Errorf("expected type %s, got %s", goalType, resp.Data[0].Type)
			}
		})
	}
}

func TestClient_GetHypeTrainEvents(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hypetrain/events" {
			t.Errorf("expected /hypetrain/events, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[HypeTrainEvent]{
			Data: []HypeTrainEvent{
				{
					ID:             "event1",
					EventType:      "hypetrain.progression",
					EventTimestamp: time.Now(),
					Version:        "1.0",
					EventData: HypeTrainEventData{
						ID:            "train1",
						BroadcasterID: "12345",
						Level:         3,
						Total:         5000,
						Goal:          6000,
						StartedAt:     time.Now().Add(-30 * time.Minute),
						ExpiresAt:     time.Now().Add(5 * time.Minute),
						LastContribution: HypeTrainContribution{
							Total: 500,
							Type:  "BITS",
							User:  "contributor123",
						},
						TopContributions: []HypeTrainContribution{
							{Total: 1000, Type: "SUBS", User: "topuser1"},
							{Total: 800, Type: "BITS", User: "topuser2"},
						},
					},
				},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetHypeTrainEvents(context.Background(), &GetHypeTrainEventsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Data))
	}
	if resp.Data[0].EventData.Level != 3 {
		t.Errorf("expected level 3, got %d", resp.Data[0].EventData.Level)
	}
	if resp.Data[0].EventData.Total != 5000 {
		t.Errorf("expected total 5000, got %d", resp.Data[0].EventData.Total)
	}
}

func TestClient_GetHypeTrainEvents_WithPagination(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		first := r.URL.Query().Get("first")
		after := r.URL.Query().Get("after")

		if first != "10" {
			t.Errorf("expected first=10, got %s", first)
		}
		if after != "cursor123" {
			t.Errorf("expected after=cursor123, got %s", after)
		}

		resp := Response[HypeTrainEvent]{
			Data:       []HypeTrainEvent{},
			Pagination: &Pagination{Cursor: "nextcursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetHypeTrainEvents(context.Background(), &GetHypeTrainEventsParams{
		BroadcasterID: "12345",
		PaginationParams: &PaginationParams{
			First: 10,
			After: "cursor123",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetHypeTrainEvents_NoEvents(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[HypeTrainEvent]{
			Data: []HypeTrainEvent{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetHypeTrainEvents(context.Background(), &GetHypeTrainEventsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 events, got %d", len(resp.Data))
	}
}

func TestClient_GetHypeTrainStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hypetrain/status" {
			t.Errorf("expected /hypetrain/status, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[HypeTrainStatus]{
			Data: []HypeTrainStatus{
				{
					ID:            "train123",
					BroadcasterID: "12345",
					Level:         3,
					Total:         5000,
					Goal:          6000,
					TopContributions: []HypeTrainContribution{
						{Total: 1000, Type: "SUBS", User: "topuser1"},
						{Total: 800, Type: "BITS", User: "topuser2"},
					},
					LastContribution: HypeTrainContribution{
						Total: 500,
						Type:  "BITS",
						User:  "lastcontrib",
					},
					StartedAt:       time.Now().Add(-30 * time.Minute),
					ExpiresAt:       time.Now().Add(5 * time.Minute),
					CooldownEndTime: time.Now().Add(1 * time.Hour),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetHypeTrainStatus(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected status, got nil")
	}
	if resp.Level != 3 {
		t.Errorf("expected level 3, got %d", resp.Level)
	}
	if resp.Total != 5000 {
		t.Errorf("expected total 5000, got %d", resp.Total)
	}
	if len(resp.TopContributions) != 2 {
		t.Errorf("expected 2 top contributions, got %d", len(resp.TopContributions))
	}
}

func TestClient_GetHypeTrainStatus_NoActiveTrain(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[HypeTrainStatus]{
			Data: []HypeTrainStatus{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetHypeTrainStatus(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil status, got non-nil")
	}
}

func TestClient_GetCreatorGoals_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetCreatorGoals(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetHypeTrainEvents_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetHypeTrainEvents(context.Background(), &GetHypeTrainEventsParams{
		BroadcasterID: "12345",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetHypeTrainStatus_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetHypeTrainStatus(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
