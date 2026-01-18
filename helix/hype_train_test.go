package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetHypeTrainEvents(t *testing.T) {
	// Using official Twitch API response sample from docs
	eventTimestamp, _ := time.Parse(time.RFC3339, "2020-04-24T20:07:24Z")
	startedAt, _ := time.Parse(time.RFC3339Nano, "2020-04-24T20:05:47.30473127Z")
	expiresAt, _ := time.Parse(time.RFC3339Nano, "2020-04-24T20:12:21.003802269Z")
	cooldownEndTime, _ := time.Parse(time.RFC3339Nano, "2020-04-24T20:13:21.003802269Z")

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hypetrain/events" {
			t.Errorf("expected /hypetrain/events, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "270954519" {
			t.Errorf("expected broadcaster_id=270954519, got %s", broadcasterID)
		}

		resp := Response[HypeTrainEvent]{
			Data: []HypeTrainEvent{
				{
					ID:             "1b0AsbInCHZW2SQFQkCzqN07Ib2",
					EventType:      "hypetrain.progression",
					EventTimestamp: eventTimestamp,
					Version:        "1.0",
					EventData: HypeTrainEventData{
						ID:              "70f0c7d8-ff60-4c50-b138-f3a352833b50",
						BroadcasterID:   "270954519",
						Level:           2,
						Total:           600,
						Goal:            1800,
						StartedAt:       startedAt,
						ExpiresAt:       expiresAt,
						CooldownEndTime: cooldownEndTime,
						LastContribution: HypeTrainContribution{
							Total: 200,
							Type:  "BITS",
							User:  "134247454",
						},
						TopContributions: []HypeTrainContribution{
							{Total: 600, Type: "BITS", User: "134247450"},
						},
					},
				},
			},
			Pagination: &Pagination{Cursor: "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjI3MDk1NDUxOToxNTg3NzU4ODQ0OjFiMEFzYkluQ0haVzJTUUZRa0N6cU4wN0liMiJ9fQ"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetHypeTrainEvents(context.Background(), &GetHypeTrainEventsParams{
		BroadcasterID: "270954519",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "1b0AsbInCHZW2SQFQkCzqN07Ib2" {
		t.Errorf("expected id 1b0AsbInCHZW2SQFQkCzqN07Ib2, got %s", resp.Data[0].ID)
	}
	if resp.Data[0].EventType != "hypetrain.progression" {
		t.Errorf("expected event_type hypetrain.progression, got %s", resp.Data[0].EventType)
	}
	if resp.Data[0].Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", resp.Data[0].Version)
	}
	if resp.Data[0].EventData.ID != "70f0c7d8-ff60-4c50-b138-f3a352833b50" {
		t.Errorf("expected event_data.id 70f0c7d8-ff60-4c50-b138-f3a352833b50, got %s", resp.Data[0].EventData.ID)
	}
	if resp.Data[0].EventData.BroadcasterID != "270954519" {
		t.Errorf("expected broadcaster_id 270954519, got %s", resp.Data[0].EventData.BroadcasterID)
	}
	if resp.Data[0].EventData.Level != 2 {
		t.Errorf("expected level 2, got %d", resp.Data[0].EventData.Level)
	}
	if resp.Data[0].EventData.Total != 600 {
		t.Errorf("expected total 600, got %d", resp.Data[0].EventData.Total)
	}
	if resp.Data[0].EventData.Goal != 1800 {
		t.Errorf("expected goal 1800, got %d", resp.Data[0].EventData.Goal)
	}
	if resp.Data[0].EventData.LastContribution.Total != 200 {
		t.Errorf("expected last_contribution.total 200, got %d", resp.Data[0].EventData.LastContribution.Total)
	}
	if resp.Data[0].EventData.LastContribution.Type != "BITS" {
		t.Errorf("expected last_contribution.type BITS, got %s", resp.Data[0].EventData.LastContribution.Type)
	}
	if resp.Data[0].EventData.LastContribution.User != "134247454" {
		t.Errorf("expected last_contribution.user 134247454, got %s", resp.Data[0].EventData.LastContribution.User)
	}
	if len(resp.Data[0].EventData.TopContributions) != 1 {
		t.Fatalf("expected 1 top contribution, got %d", len(resp.Data[0].EventData.TopContributions))
	}
	if resp.Data[0].EventData.TopContributions[0].Total != 600 {
		t.Errorf("expected top_contributions[0].total 600, got %d", resp.Data[0].EventData.TopContributions[0].Total)
	}
	if resp.Pagination.Cursor != "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjI3MDk1NDUxOToxNTg3NzU4ODQ0OjFiMEFzYkluQ0haVzJTUUZRa0N6cU4wN0liMiJ9fQ" {
		t.Errorf("expected pagination cursor, got %s", resp.Pagination.Cursor)
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
	// Using official Twitch API response sample values from docs
	startedAt, _ := time.Parse(time.RFC3339Nano, "2020-04-24T20:05:47.30473127Z")
	expiresAt, _ := time.Parse(time.RFC3339Nano, "2020-04-24T20:12:21.003802269Z")
	cooldownEndTime, _ := time.Parse(time.RFC3339Nano, "2020-04-24T20:13:21.003802269Z")

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hypetrain/status" {
			t.Errorf("expected /hypetrain/status, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "270954519" {
			t.Errorf("expected broadcaster_id=270954519, got %s", broadcasterID)
		}

		resp := Response[HypeTrainStatus]{
			Data: []HypeTrainStatus{
				{
					ID:            "70f0c7d8-ff60-4c50-b138-f3a352833b50",
					BroadcasterID: "270954519",
					Level:         2,
					Total:         600,
					Goal:          1800,
					TopContributions: []HypeTrainContribution{
						{Total: 600, Type: "BITS", User: "134247450"},
					},
					LastContribution: HypeTrainContribution{
						Total: 200,
						Type:  "BITS",
						User:  "134247454",
					},
					StartedAt:       startedAt,
					ExpiresAt:       expiresAt,
					CooldownEndTime: cooldownEndTime,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetHypeTrainStatus(context.Background(), "270954519")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected status, got nil")
	}
	if resp.ID != "70f0c7d8-ff60-4c50-b138-f3a352833b50" {
		t.Errorf("expected id 70f0c7d8-ff60-4c50-b138-f3a352833b50, got %s", resp.ID)
	}
	if resp.BroadcasterID != "270954519" {
		t.Errorf("expected broadcaster_id 270954519, got %s", resp.BroadcasterID)
	}
	if resp.Level != 2 {
		t.Errorf("expected level 2, got %d", resp.Level)
	}
	if resp.Total != 600 {
		t.Errorf("expected total 600, got %d", resp.Total)
	}
	if resp.Goal != 1800 {
		t.Errorf("expected goal 1800, got %d", resp.Goal)
	}
	if len(resp.TopContributions) != 1 {
		t.Fatalf("expected 1 top contribution, got %d", len(resp.TopContributions))
	}
	if resp.TopContributions[0].Total != 600 {
		t.Errorf("expected top_contributions[0].total 600, got %d", resp.TopContributions[0].Total)
	}
	if resp.TopContributions[0].Type != "BITS" {
		t.Errorf("expected top_contributions[0].type BITS, got %s", resp.TopContributions[0].Type)
	}
	if resp.TopContributions[0].User != "134247450" {
		t.Errorf("expected top_contributions[0].user 134247450, got %s", resp.TopContributions[0].User)
	}
	if resp.LastContribution.Total != 200 {
		t.Errorf("expected last_contribution.total 200, got %d", resp.LastContribution.Total)
	}
	if resp.LastContribution.Type != "BITS" {
		t.Errorf("expected last_contribution.type BITS, got %s", resp.LastContribution.Type)
	}
	if resp.LastContribution.User != "134247454" {
		t.Errorf("expected last_contribution.user 134247454, got %s", resp.LastContribution.User)
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
