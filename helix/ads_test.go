package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_StartCommercial(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/channels/commercial" {
			t.Errorf("expected /channels/commercial, got %s", r.URL.Path)
		}

		var params StartCommercialParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.BroadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", params.BroadcasterID)
		}
		if params.Length != 60 {
			t.Errorf("expected length=60, got %d", params.Length)
		}

		resp := Response[Commercial]{
			Data: []Commercial{
				{
					Length:     60,
					Message:    "Commercial started",
					RetryAfter: 480,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.StartCommercial(context.Background(), &StartCommercialParams{
		BroadcasterID: "12345",
		Length:        60,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Length != 60 {
		t.Errorf("expected length 60, got %d", result.Length)
	}
	if result.RetryAfter != 480 {
		t.Errorf("expected retry_after 480, got %d", result.RetryAfter)
	}
}

func TestClient_StartCommercial_DifferentLengths(t *testing.T) {
	lengths := []int{30, 60, 90, 120, 150, 180}

	for _, length := range lengths {
		t.Run(string(rune(length)), func(t *testing.T) {
			client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
				resp := Response[Commercial]{
					Data: []Commercial{
						{Length: length, Message: "OK", RetryAfter: 480},
					},
				}
				_ = json.NewEncoder(w).Encode(resp)
			})
			defer server.Close()

			result, err := client.StartCommercial(context.Background(), &StartCommercialParams{
				BroadcasterID: "12345",
				Length:        length,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Length != length {
				t.Errorf("expected length %d, got %d", length, result.Length)
			}
		})
	}
}

func TestClient_GetAdSchedule(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/ads" {
			t.Errorf("expected /channels/ads, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		nextAdAt := time.Now().Add(10 * time.Minute)
		lastAdAt := time.Now().Add(-5 * time.Minute)
		snoozeRefreshAt := time.Now().Add(30 * time.Minute)

		resp := Response[AdSchedule]{
			Data: []AdSchedule{
				{
					NextAdAt:        nextAdAt,
					LastAdAt:        lastAdAt,
					Duration:        60,
					PrerollFreeTime: 300,
					SnoozeCount:     2,
					SnoozeRefreshAt: snoozeRefreshAt,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetAdSchedule(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Duration != 60 {
		t.Errorf("expected duration 60, got %d", result.Duration)
	}
	if result.SnoozeCount != 2 {
		t.Errorf("expected snooze_count 2, got %d", result.SnoozeCount)
	}
	if result.PrerollFreeTime != 300 {
		t.Errorf("expected preroll_free_time 300, got %d", result.PrerollFreeTime)
	}
}

func TestClient_GetAdSchedule_NoSchedule(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[AdSchedule]{
			Data: []AdSchedule{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetAdSchedule(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestClient_SnoozeNextAd(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ads/schedule/snooze" {
			t.Errorf("expected /channels/ads/schedule/snooze, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		nextAdAt := time.Now().Add(15 * time.Minute)
		snoozeRefreshAt := time.Now().Add(45 * time.Minute)

		resp := Response[SnoozeNextAdResponse]{
			Data: []SnoozeNextAdResponse{
				{
					SnoozeCount:     1,
					SnoozeRefreshAt: snoozeRefreshAt,
					NextAdAt:        nextAdAt,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SnoozeNextAd(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.SnoozeCount != 1 {
		t.Errorf("expected snooze_count 1, got %d", result.SnoozeCount)
	}
}

func TestClient_SnoozeNextAd_NoSnoozeAvailable(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[SnoozeNextAdResponse]{
			Data: []SnoozeNextAdResponse{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SnoozeNextAd(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestClient_StartCommercial_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Commercial]{
			Data: []Commercial{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.StartCommercial(context.Background(), &StartCommercialParams{
		BroadcasterID: "12345",
		Length:        60,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestClient_StartCommercial_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid access token",
		})
	})
	defer server.Close()

	_, err := client.StartCommercial(context.Background(), &StartCommercialParams{
		BroadcasterID: "12345",
		Length:        60,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetAdSchedule_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid access token",
		})
	})
	defer server.Close()

	_, err := client.GetAdSchedule(context.Background(), "12345")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SnoozeNextAd_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid access token",
		})
	})
	defer server.Close()

	_, err := client.SnoozeNextAd(context.Background(), "12345")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
