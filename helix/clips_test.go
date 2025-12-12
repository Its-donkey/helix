package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_CreateClip(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/clips" {
			t.Errorf("expected /clips, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[CreateClipResponse]{
			Data: []CreateClipResponse{
				{
					ID:      "AwkwardHelplessSalamanderSwiftRage",
					EditURL: "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage/edit",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateClip(context.Background(), &CreateClipParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "AwkwardHelplessSalamanderSwiftRage" {
		t.Errorf("expected clip ID, got %s", result.ID)
	}
}

func TestClient_CreateClip_WithDelay(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		hasDelay := r.URL.Query().Get("has_delay")
		if hasDelay != "true" {
			t.Errorf("expected has_delay=true, got %s", hasDelay)
		}

		resp := Response[CreateClipResponse]{
			Data: []CreateClipResponse{
				{ID: "clip123", EditURL: "https://clips.twitch.tv/clip123/edit"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.CreateClip(context.Background(), &CreateClipParams{
		BroadcasterID: "12345",
		HasDelay:      true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetClips_ByBroadcaster(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clips" {
			t.Errorf("expected /clips, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[Clip]{
			Data: []Clip{
				{
					ID:              "clip1",
					URL:             "https://clips.twitch.tv/clip1",
					EmbedURL:        "https://clips.twitch.tv/embed?clip=clip1",
					BroadcasterID:   "12345",
					BroadcasterName: "Streamer",
					CreatorID:       "67890",
					CreatorName:     "Clipper",
					VideoID:         "video123",
					GameID:          "game123",
					Language:        "en",
					Title:           "Amazing Play",
					ViewCount:       1000,
					CreatedAt:       time.Now(),
					ThumbnailURL:    "https://example.com/thumb.jpg",
					Duration:        30.5,
					IsFeatured:      false,
				},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetClips(context.Background(), &GetClipsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(resp.Data))
	}
	if resp.Data[0].ViewCount != 1000 {
		t.Errorf("expected view count 1000, got %d", resp.Data[0].ViewCount)
	}
}

func TestClient_GetClips_ByGameID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		gameID := r.URL.Query().Get("game_id")
		if gameID != "game123" {
			t.Errorf("expected game_id=game123, got %s", gameID)
		}

		resp := Response[Clip]{
			Data: []Clip{
				{ID: "clip1", GameID: "game123"},
				{ID: "clip2", GameID: "game123"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetClips(context.Background(), &GetClipsParams{
		GameID: "game123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 clips, got %d", len(resp.Data))
	}
}

func TestClient_GetClips_ByIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		if len(ids) != 3 {
			t.Errorf("expected 3 ids, got %d", len(ids))
		}

		resp := Response[Clip]{
			Data: []Clip{
				{ID: "clip1"},
				{ID: "clip2"},
				{ID: "clip3"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetClips(context.Background(), &GetClipsParams{
		IDs: []string{"clip1", "clip2", "clip3"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 clips, got %d", len(resp.Data))
	}
}

func TestClient_GetClips_Featured(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		isFeatured := r.URL.Query().Get("is_featured")
		if isFeatured != "true" {
			t.Errorf("expected is_featured=true, got %s", isFeatured)
		}

		resp := Response[Clip]{
			Data: []Clip{
				{ID: "featured1", IsFeatured: true},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	featured := true
	resp, err := client.GetClips(context.Background(), &GetClipsParams{
		BroadcasterID: "12345",
		IsFeatured:    &featured,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(resp.Data))
	}
}

func TestClient_GetClips_DateRange(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		startedAt := r.URL.Query().Get("started_at")
		endedAt := r.URL.Query().Get("ended_at")

		if startedAt == "" {
			t.Error("expected started_at to be set")
		}
		if endedAt == "" {
			t.Error("expected ended_at to be set")
		}

		resp := Response[Clip]{Data: []Clip{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetClips(context.Background(), &GetClipsParams{
		BroadcasterID: "12345",
		StartedAt:     startTime,
		EndedAt:       endTime,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetClipsDownload(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clips/download" {
			t.Errorf("expected /clips/download, got %s", r.URL.Path)
		}

		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[ClipDownload]{
			Data: []ClipDownload{
				{
					ID:        "clip1",
					URL:       "https://example.com/download/clip1.mp4",
					ExpiresAt: "2024-01-15T12:00:00Z",
				},
				{
					ID:        "clip2",
					URL:       "https://example.com/download/clip2.mp4",
					ExpiresAt: "2024-01-15T12:00:00Z",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetClipsDownload(context.Background(), []string{"clip1", "clip2"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 downloads, got %d", len(resp.Data))
	}
	if resp.Data[0].URL == "" {
		t.Error("expected URL to be set")
	}
}
