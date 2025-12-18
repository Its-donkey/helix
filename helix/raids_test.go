package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_StartRaid(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/raids" {
			t.Errorf("expected /raids, got %s", r.URL.Path)
		}

		fromBroadcasterID := r.URL.Query().Get("from_broadcaster_id")
		toBroadcasterID := r.URL.Query().Get("to_broadcaster_id")

		if fromBroadcasterID != "12345" {
			t.Errorf("expected from_broadcaster_id=12345, got %s", fromBroadcasterID)
		}
		if toBroadcasterID != "67890" {
			t.Errorf("expected to_broadcaster_id=67890, got %s", toBroadcasterID)
		}

		resp := Response[Raid]{
			Data: []Raid{
				{
					CreatedAt: time.Now(),
					IsMature:  false,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.StartRaid(context.Background(), &StartRaidParams{
		FromBroadcasterID: "12345",
		ToBroadcasterID:   "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.IsMature {
		t.Error("expected is_mature to be false")
	}
}

func TestClient_StartRaid_MatureContent(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Raid]{
			Data: []Raid{
				{
					CreatedAt: time.Now(),
					IsMature:  true,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.StartRaid(context.Background(), &StartRaidParams{
		FromBroadcasterID: "12345",
		ToBroadcasterID:   "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsMature {
		t.Error("expected is_mature to be true")
	}
}

func TestClient_StartRaid_NoResult(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Raid]{
			Data: []Raid{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.StartRaid(context.Background(), &StartRaidParams{
		FromBroadcasterID: "12345",
		ToBroadcasterID:   "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_CancelRaid(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/raids" {
			t.Errorf("expected /raids, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.CancelRaid(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StartRaid_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"conflict"}`))
	})
	defer server.Close()

	_, err := client.StartRaid(context.Background(), &StartRaidParams{
		FromBroadcasterID: "12345",
		ToBroadcasterID:   "67890",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CancelRaid_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	err := client.CancelRaid(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
