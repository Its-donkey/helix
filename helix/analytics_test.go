package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetExtensionAnalytics(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/analytics/extensions" {
			t.Errorf("expected /analytics/extensions, got %s", r.URL.Path)
		}

		extensionID := r.URL.Query().Get("extension_id")
		if extensionID != "ext123" {
			t.Errorf("expected extension_id ext123, got %s", extensionID)
		}

		resp := Response[ExtensionAnalytics]{
			Data: []ExtensionAnalytics{
				{
					ExtensionID: "ext123",
					URL:         "https://example.com/report.csv",
					Type:        "overview_v2",
					DateRange: DateRange{
						StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						EndedAt:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensionAnalytics(context.Background(), &GetExtensionAnalyticsParams{
		ExtensionID: "ext123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 analytics report, got %d", len(resp.Data))
	}
	if resp.Data[0].ExtensionID != "ext123" {
		t.Errorf("expected extension_id ext123, got %s", resp.Data[0].ExtensionID)
	}
}

func TestClient_GetGameAnalytics(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/analytics/games" {
			t.Errorf("expected /analytics/games, got %s", r.URL.Path)
		}

		gameID := r.URL.Query().Get("game_id")
		if gameID != "game456" {
			t.Errorf("expected game_id game456, got %s", gameID)
		}

		resp := Response[GameAnalytics]{
			Data: []GameAnalytics{
				{
					GameID: "game456",
					URL:    "https://example.com/game_report.csv",
					Type:   "overview_v2",
					DateRange: DateRange{
						StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						EndedAt:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGameAnalytics(context.Background(), &GetGameAnalyticsParams{
		GameID: "game456",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 analytics report, got %d", len(resp.Data))
	}
	if resp.Data[0].GameID != "game456" {
		t.Errorf("expected game_id game456, got %s", resp.Data[0].GameID)
	}
}

func TestClient_GetExtensionAnalytics_NilParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[ExtensionAnalytics]{Data: []ExtensionAnalytics{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetExtensionAnalytics(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetExtensionAnalytics_AllParams(t *testing.T) {
	startedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endedAt := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("extension_id") != "ext123" {
			t.Errorf("expected extension_id ext123, got %s", q.Get("extension_id"))
		}
		if q.Get("type") != "overview_v2" {
			t.Errorf("expected type overview_v2, got %s", q.Get("type"))
		}
		if q.Get("started_at") == "" {
			t.Error("expected started_at to be set")
		}
		if q.Get("ended_at") == "" {
			t.Error("expected ended_at to be set")
		}
		if q.Get("first") != "10" {
			t.Errorf("expected first 10, got %s", q.Get("first"))
		}

		resp := Response[ExtensionAnalytics]{Data: []ExtensionAnalytics{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetExtensionAnalytics(context.Background(), &GetExtensionAnalyticsParams{
		ExtensionID:      "ext123",
		Type:             "overview_v2",
		StartedAt:        startedAt,
		EndedAt:          endedAt,
		PaginationParams: &PaginationParams{First: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetExtensionAnalytics_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid access token",
		})
	})
	defer server.Close()

	_, err := client.GetExtensionAnalytics(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetGameAnalytics_NilParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[GameAnalytics]{Data: []GameAnalytics{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetGameAnalytics(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetGameAnalytics_AllParams(t *testing.T) {
	startedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endedAt := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("game_id") != "game456" {
			t.Errorf("expected game_id game456, got %s", q.Get("game_id"))
		}
		if q.Get("type") != "overview_v2" {
			t.Errorf("expected type overview_v2, got %s", q.Get("type"))
		}
		if q.Get("started_at") == "" {
			t.Error("expected started_at to be set")
		}
		if q.Get("ended_at") == "" {
			t.Error("expected ended_at to be set")
		}

		resp := Response[GameAnalytics]{Data: []GameAnalytics{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetGameAnalytics(context.Background(), &GetGameAnalyticsParams{
		GameID:    "game456",
		Type:      "overview_v2",
		StartedAt: startedAt,
		EndedAt:   endedAt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetGameAnalytics_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid access token",
		})
	})
	defer server.Close()

	_, err := client.GetGameAnalytics(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
