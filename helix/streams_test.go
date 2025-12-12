package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetStreams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/streams" {
			t.Errorf("expected /streams, got %s", r.URL.Path)
		}

		resp := Response[Stream]{
			Data: []Stream{
				{
					ID:           "stream123",
					UserID:       "12345",
					UserLogin:    "streamer1",
					UserName:     "Streamer1",
					GameID:       "game123",
					GameName:     "Test Game",
					Type:         "live",
					Title:        "Test Stream",
					ViewerCount:  1000,
					StartedAt:    time.Now(),
					Language:     "en",
					ThumbnailURL: "https://example.com/thumb.jpg",
					Tags:         []string{"English", "Gaming"},
					IsMature:     false,
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetStreams(context.Background(), &GetStreamsParams{
		PaginationParams: &PaginationParams{First: 20},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(resp.Data))
	}
	if resp.Data[0].ViewerCount != 1000 {
		t.Errorf("expected viewer count 1000, got %d", resp.Data[0].ViewerCount)
	}
}

func TestClient_GetStreams_WithFilters(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userLogins := r.URL.Query()["user_login"]
		if len(userLogins) != 2 {
			t.Errorf("expected 2 user_logins, got %d", len(userLogins))
		}

		gameIDs := r.URL.Query()["game_id"]
		if len(gameIDs) != 1 {
			t.Errorf("expected 1 game_id, got %d", len(gameIDs))
		}

		streamType := r.URL.Query().Get("type")
		if streamType != "live" {
			t.Errorf("expected type=live, got %s", streamType)
		}

		resp := Response[Stream]{Data: []Stream{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetStreams(context.Background(), &GetStreamsParams{
		UserLogins: []string{"user1", "user2"},
		GameIDs:    []string{"game123"},
		Type:       "live",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetStreams_ByUserIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userIDs := r.URL.Query()["user_id"]
		if len(userIDs) != 3 {
			t.Errorf("expected 3 user_ids, got %d", len(userIDs))
		}

		resp := Response[Stream]{
			Data: []Stream{
				{ID: "stream1", UserID: "12345"},
				{ID: "stream2", UserID: "67890"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetStreams(context.Background(), &GetStreamsParams{
		UserIDs: []string{"12345", "67890", "11111"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 streams, got %d", len(resp.Data))
	}
}

func TestClient_GetFollowedStreams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/streams/followed" {
			t.Errorf("expected /streams/followed, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := Response[Stream]{
			Data: []Stream{
				{ID: "stream1", UserID: "11111", Title: "Followed Stream 1"},
				{ID: "stream2", UserID: "22222", Title: "Followed Stream 2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetFollowedStreams(context.Background(), &GetFollowedStreamsParams{
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 streams, got %d", len(resp.Data))
	}
}

func TestClient_GetStreamKey(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/streams/key" {
			t.Errorf("expected /streams/key, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[StreamKey]{
			Data: []StreamKey{
				{StreamKey: "live_12345_abcdefghij"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	key, err := client.GetStreamKey(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected stream key to not be nil")
	}
	if key.StreamKey != "live_12345_abcdefghij" {
		t.Errorf("expected stream key 'live_12345_abcdefghij', got %s", key.StreamKey)
	}
}

func TestClient_CreateStreamMarker(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/streams/markers" {
			t.Errorf("expected /streams/markers, got %s", r.URL.Path)
		}

		var body CreateStreamMarkerParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.UserID != "12345" {
			t.Errorf("expected user_id=12345, got %s", body.UserID)
		}
		if body.Description != "Highlight moment" {
			t.Errorf("expected description 'Highlight moment', got %s", body.Description)
		}

		resp := Response[StreamMarker]{
			Data: []StreamMarker{
				{
					ID:              "marker123",
					CreatedAt:       time.Now(),
					Description:     "Highlight moment",
					PositionSeconds: 3600,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	marker, err := client.CreateStreamMarker(context.Background(), &CreateStreamMarkerParams{
		UserID:      "12345",
		Description: "Highlight moment",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if marker == nil {
		t.Fatal("expected marker to not be nil")
	}
	if marker.PositionSeconds != 3600 {
		t.Errorf("expected position 3600, got %d", marker.PositionSeconds)
	}
}

func TestClient_GetStreamMarkers(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/streams/markers" {
			t.Errorf("expected /streams/markers, got %s", r.URL.Path)
		}

		videoID := r.URL.Query().Get("video_id")
		if videoID != "video123" {
			t.Errorf("expected video_id=video123, got %s", videoID)
		}

		resp := Response[VideoStreamMarkers]{
			Data: []VideoStreamMarkers{
				{
					UserID:    "12345",
					UserLogin: "streamer",
					UserName:  "Streamer",
					Videos: []struct {
						VideoID string         `json:"video_id"`
						Markers []StreamMarker `json:"markers"`
					}{
						{
							VideoID: "video123",
							Markers: []StreamMarker{
								{ID: "marker1", Description: "Start"},
								{ID: "marker2", Description: "End"},
							},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetStreamMarkers(context.Background(), &GetStreamMarkersParams{
		VideoID: "video123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Data))
	}
}
