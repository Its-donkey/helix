package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_SearchCategories(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/categories" {
			t.Errorf("expected /search/categories, got %s", r.URL.Path)
		}

		query := r.URL.Query().Get("query")
		if query != "fort" {
			t.Errorf("expected query=fort, got %s", query)
		}

		resp := Response[SearchCategory]{
			Data: []SearchCategory{
				{ID: "33214", Name: "Fortnite", BoxArtURL: "https://example.com/fortnite.jpg"},
				{ID: "12345", Name: "Fort Boyard", BoxArtURL: "https://example.com/fortboyard.jpg"},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.SearchCategories(context.Background(), &SearchCategoriesParams{
		Query: "fort",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "Fortnite" {
		t.Errorf("expected Fortnite, got %s", resp.Data[0].Name)
	}
}

func TestClient_SearchCategories_WithPagination(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		first := r.URL.Query().Get("first")
		if first != "10" {
			t.Errorf("expected first=10, got %s", first)
		}

		resp := Response[SearchCategory]{
			Data: []SearchCategory{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.SearchCategories(context.Background(), &SearchCategoriesParams{
		Query:            "test",
		PaginationParams: &PaginationParams{First: 10},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SearchChannels(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/channels" {
			t.Errorf("expected /search/channels, got %s", r.URL.Path)
		}

		query := r.URL.Query().Get("query")
		if query != "ninja" {
			t.Errorf("expected query=ninja, got %s", query)
		}

		resp := Response[SearchChannel]{
			Data: []SearchChannel{
				{
					ID:                  "12345",
					DisplayName:         "Ninja",
					BroadcasterLogin:    "ninja",
					BroadcasterLanguage: "en",
					GameID:              "33214",
					GameName:            "Fortnite",
					IsLive:              true,
					Title:               "Playing Fortnite!",
					Tags:                []string{"English", "Gaming"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.SearchChannels(context.Background(), &SearchChannelsParams{
		Query: "ninja",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(resp.Data))
	}
	if !resp.Data[0].IsLive {
		t.Error("expected channel to be live")
	}
}

func TestClient_SearchChannels_LiveOnly(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		liveOnly := r.URL.Query().Get("live_only")
		if liveOnly != "true" {
			t.Errorf("expected live_only=true, got %s", liveOnly)
		}

		resp := Response[SearchChannel]{
			Data: []SearchChannel{
				{ID: "123", DisplayName: "LiveStreamer", IsLive: true},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.SearchChannels(context.Background(), &SearchChannelsParams{
		Query:    "stream",
		LiveOnly: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(resp.Data))
	}
}

func TestClient_SearchChannels_WithPagination(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		first := r.URL.Query().Get("first")
		after := r.URL.Query().Get("after")
		if first != "25" {
			t.Errorf("expected first=25, got %s", first)
		}
		if after != "cursor123" {
			t.Errorf("expected after=cursor123, got %s", after)
		}

		resp := Response[SearchChannel]{
			Data:       []SearchChannel{},
			Pagination: &Pagination{Cursor: "nextcursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.SearchChannels(context.Background(), &SearchChannelsParams{
		Query: "test",
		PaginationParams: &PaginationParams{
			First: 25,
			After: "cursor123",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
