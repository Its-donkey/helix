package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetVideos_ByIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/videos" {
			t.Errorf("expected /videos, got %s", r.URL.Path)
		}

		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[Video]{
			Data: []Video{
				{
					ID:           "video1",
					UserID:       "12345",
					UserLogin:    "streamer",
					UserName:     "Streamer",
					Title:        "Amazing Stream",
					Description:  "Great content",
					CreatedAt:    time.Now(),
					PublishedAt:  time.Now(),
					URL:          "https://twitch.tv/videos/video1",
					ThumbnailURL: "https://example.com/thumb.jpg",
					Viewable:     "public",
					ViewCount:    1000,
					Language:     "en",
					Type:         "archive",
					Duration:     "3h2m1s",
				},
				{
					ID:       "video2",
					Title:    "Highlight",
					Type:     "highlight",
					Duration: "30s",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetVideos(context.Background(), &GetVideosParams{
		IDs: []string{"video1", "video2"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(resp.Data))
	}
	if resp.Data[0].Type != "archive" {
		t.Errorf("expected archive, got %s", resp.Data[0].Type)
	}
}

func TestClient_GetVideos_ByUserID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := Response[Video]{
			Data: []Video{
				{ID: "video1", UserID: "12345"},
				{ID: "video2", UserID: "12345"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetVideos(context.Background(), &GetVideosParams{
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(resp.Data))
	}
}

func TestClient_GetVideos_ByGameID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		gameID := r.URL.Query().Get("game_id")
		if gameID != "game123" {
			t.Errorf("expected game_id=game123, got %s", gameID)
		}

		resp := Response[Video]{
			Data: []Video{
				{ID: "video1"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetVideos(context.Background(), &GetVideosParams{
		GameID: "game123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetVideos_WithFilters(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		language := r.URL.Query().Get("language")
		period := r.URL.Query().Get("period")
		sort := r.URL.Query().Get("sort")
		videoType := r.URL.Query().Get("type")

		if language != "en" {
			t.Errorf("expected language=en, got %s", language)
		}
		if period != "week" {
			t.Errorf("expected period=week, got %s", period)
		}
		if sort != "views" {
			t.Errorf("expected sort=views, got %s", sort)
		}
		if videoType != "highlight" {
			t.Errorf("expected type=highlight, got %s", videoType)
		}

		resp := Response[Video]{Data: []Video{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetVideos(context.Background(), &GetVideosParams{
		UserID:   "12345",
		Language: "en",
		Period:   "week",
		Sort:     "views",
		Type:     "highlight",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetVideos_WithMutedSegments(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Video]{
			Data: []Video{
				{
					ID:    "video1",
					Title: "Video with muted audio",
					MutedSegments: []MutedSegment{
						{Duration: 30, Offset: 120},
						{Duration: 60, Offset: 300},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetVideos(context.Background(), &GetVideosParams{
		IDs: []string{"video1"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data[0].MutedSegments) != 2 {
		t.Errorf("expected 2 muted segments, got %d", len(resp.Data[0].MutedSegments))
	}
}

func TestClient_DeleteVideos(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/videos" {
			t.Errorf("expected /videos, got %s", r.URL.Path)
		}

		ids := r.URL.Query()["id"]
		if len(ids) != 3 {
			t.Errorf("expected 3 ids, got %d", len(ids))
		}

		resp := struct {
			Data []string `json:"data"`
		}{
			Data: []string{"video1", "video2", "video3"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	deletedIDs, err := client.DeleteVideos(context.Background(), []string{"video1", "video2", "video3"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deletedIDs) != 3 {
		t.Errorf("expected 3 deleted IDs, got %d", len(deletedIDs))
	}
}

func TestClient_GetVideos_WithPagination(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		first := r.URL.Query().Get("first")
		if first != "20" {
			t.Errorf("expected first=20, got %s", first)
		}

		resp := Response[Video]{
			Data:       []Video{{ID: "1"}},
			Pagination: &Pagination{Cursor: "nextpage"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetVideos(context.Background(), &GetVideosParams{
		UserID:           "12345",
		PaginationParams: &PaginationParams{First: 20},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Pagination.Cursor != "nextpage" {
		t.Errorf("expected cursor 'nextpage', got %s", resp.Pagination.Cursor)
	}
}
