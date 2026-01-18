package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetCreatorGoals(t *testing.T) {
	// Using official Twitch API response sample from docs
	createdAt, _ := time.Parse(time.RFC3339, "2021-08-16T17:22:23Z")

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/goals" {
			t.Errorf("expected /goals, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "141981764" {
			t.Errorf("expected broadcaster_id=141981764, got %s", broadcasterID)
		}

		resp := Response[CreatorGoal]{
			Data: []CreatorGoal{
				{
					ID:               "1woowvbkiNv8BRxEWSqmQz6Zk92",
					BroadcasterID:    "141981764",
					BroadcasterName:  "TwitchDev",
					BroadcasterLogin: "twitchdev",
					Type:             "follower",
					Description:      "Follow goal for Helix testing",
					CurrentAmount:    27062,
					TargetAmount:     30000,
					CreatedAt:        createdAt,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCreatorGoals(context.Background(), "141981764")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "1woowvbkiNv8BRxEWSqmQz6Zk92" {
		t.Errorf("expected id 1woowvbkiNv8BRxEWSqmQz6Zk92, got %s", resp.Data[0].ID)
	}
	if resp.Data[0].BroadcasterID != "141981764" {
		t.Errorf("expected broadcaster_id 141981764, got %s", resp.Data[0].BroadcasterID)
	}
	if resp.Data[0].BroadcasterName != "TwitchDev" {
		t.Errorf("expected broadcaster_name TwitchDev, got %s", resp.Data[0].BroadcasterName)
	}
	if resp.Data[0].Type != "follower" {
		t.Errorf("expected type follower, got %s", resp.Data[0].Type)
	}
	if resp.Data[0].Description != "Follow goal for Helix testing" {
		t.Errorf("expected description 'Follow goal for Helix testing', got %s", resp.Data[0].Description)
	}
	if resp.Data[0].CurrentAmount != 27062 {
		t.Errorf("expected current_amount 27062, got %d", resp.Data[0].CurrentAmount)
	}
	if resp.Data[0].TargetAmount != 30000 {
		t.Errorf("expected target_amount 30000, got %d", resp.Data[0].TargetAmount)
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
