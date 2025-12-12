package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetBitsLeaderboard(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bits/leaderboard" {
			t.Errorf("expected /bits/leaderboard, got %s", r.URL.Path)
		}

		resp := BitsLeaderboardResponse{
			Data: []BitsLeaderboard{
				{UserID: "1", UserLogin: "topchatter", UserName: "Top Chatter", Rank: 1, Score: 10000},
				{UserID: "2", UserLogin: "second", UserName: "Second Place", Rank: 2, Score: 5000},
				{UserID: "3", UserLogin: "third", UserName: "Third Place", Rank: 3, Score: 2500},
			},
			DateRange: DateRange{
				StartedAt: time.Now().Add(-7 * 24 * time.Hour),
				EndedAt:   time.Now(),
			},
			Total: 3,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBitsLeaderboard(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(resp.Data))
	}
	if resp.Data[0].Score != 10000 {
		t.Errorf("expected score 10000, got %d", resp.Data[0].Score)
	}
	if resp.Total != 3 {
		t.Errorf("expected total 3, got %d", resp.Total)
	}
}

func TestClient_GetBitsLeaderboard_WithParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		period := r.URL.Query().Get("period")
		if period != "week" {
			t.Errorf("expected period=week, got %s", period)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := BitsLeaderboardResponse{
			Data: []BitsLeaderboard{
				{UserID: "12345", UserLogin: "specific", UserName: "Specific User", Rank: 5, Score: 1000},
			},
			Total: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBitsLeaderboard(context.Background(), &GetBitsLeaderboardParams{
		Period: "week",
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(resp.Data))
	}
}

func TestClient_GetBitsLeaderboard_WithStartTime(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		startedAt := r.URL.Query().Get("started_at")
		if startedAt == "" {
			t.Error("expected started_at to be set")
		}

		resp := BitsLeaderboardResponse{
			Data: []BitsLeaderboard{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetBitsLeaderboard(context.Background(), &GetBitsLeaderboardParams{
		StartedAt: startTime,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetCheermotes(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bits/cheermotes" {
			t.Errorf("expected /bits/cheermotes, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[Cheermote]{
			Data: []Cheermote{
				{
					Prefix:       "Cheer",
					Type:         "global_first_party",
					Order:        1,
					LastUpdated:  time.Now(),
					IsCharitable: false,
					Tiers: []CheermoteTier{
						{
							MinBits:        1,
							ID:             "1",
							Color:          "#979797",
							CanCheer:       true,
							ShowInBitsCard: true,
							Images: CheermoteImages{
								Dark: CheermoteTheme{
									Animated: map[string]string{
										"1":   "https://example.com/1.gif",
										"1.5": "https://example.com/1.5.gif",
									},
									Static: map[string]string{
										"1":   "https://example.com/1.png",
										"1.5": "https://example.com/1.5.png",
									},
								},
								Light: CheermoteTheme{
									Animated: map[string]string{},
									Static:   map[string]string{},
								},
							},
						},
						{
							MinBits: 100,
							ID:      "100",
							Color:   "#9c3ee8",
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCheermotes(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 cheermote, got %d", len(resp.Data))
	}
	if resp.Data[0].Prefix != "Cheer" {
		t.Errorf("expected prefix Cheer, got %s", resp.Data[0].Prefix)
	}
	if len(resp.Data[0].Tiers) != 2 {
		t.Errorf("expected 2 tiers, got %d", len(resp.Data[0].Tiers))
	}
}

func TestClient_GetCheermotes_Global(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "" {
			t.Errorf("expected no broadcaster_id, got %s", broadcasterID)
		}

		resp := Response[Cheermote]{
			Data: []Cheermote{
				{Prefix: "Cheer", Type: "global_first_party"},
				{Prefix: "BibleThump", Type: "global_third_party"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCheermotes(context.Background(), "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 cheermotes, got %d", len(resp.Data))
	}
}

func TestClient_GetBitsLeaderboard_NilParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := BitsLeaderboardResponse{
			Data:  []BitsLeaderboard{},
			Total: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetBitsLeaderboard(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetBitsLeaderboard_WithCount(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		count := r.URL.Query().Get("count")
		if count == "" {
			t.Error("expected count to be set")
		}

		resp := BitsLeaderboardResponse{
			Data:  []BitsLeaderboard{},
			Total: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetBitsLeaderboard(context.Background(), &GetBitsLeaderboardParams{
		Count: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetBitsLeaderboard_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid access token",
		})
	})
	defer server.Close()

	_, err := client.GetBitsLeaderboard(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetCheermotes_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid access token",
		})
	})
	defer server.Close()

	_, err := client.GetCheermotes(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
