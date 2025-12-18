package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetGames(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/games" {
			t.Errorf("expected /games, got %s", r.URL.Path)
		}

		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[Game]{
			Data: []Game{
				{ID: "123", Name: "Game 1", BoxArtURL: "https://example.com/1.jpg"},
				{ID: "456", Name: "Game 2", BoxArtURL: "https://example.com/2.jpg"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGames(ctx, &GetGamesParams{
		IDs: []string{"123", "456"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 games, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "Game 1" {
		t.Errorf("expected Game 1, got %s", resp.Data[0].Name)
	}
}

func TestClient_GetGames_ByName(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		names := r.URL.Query()["name"]
		if len(names) != 1 {
			t.Errorf("expected 1 name, got %d", len(names))
		}
		if names[0] != "Fortnite" {
			t.Errorf("expected Fortnite, got %s", names[0])
		}

		resp := Response[Game]{
			Data: []Game{
				{ID: "33214", Name: "Fortnite", BoxArtURL: "https://example.com/fortnite.jpg"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGames(ctx, &GetGamesParams{
		Names: []string{"Fortnite"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 game, got %d", len(resp.Data))
	}
}

func TestClient_GetGames_ByIGDB(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		igdbIDs := r.URL.Query()["igdb_id"]
		if len(igdbIDs) != 1 {
			t.Errorf("expected 1 igdb_id, got %d", len(igdbIDs))
		}

		resp := Response[Game]{
			Data: []Game{
				{ID: "123", Name: "Some Game", IGDBId: "12345"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGames(ctx, &GetGamesParams{
		IGDBIDs: []string{"12345"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 game, got %d", len(resp.Data))
	}
}

func TestClient_GetTopGames(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/games/top" {
			t.Errorf("expected /games/top, got %s", r.URL.Path)
		}

		first := r.URL.Query().Get("first")
		if first != "5" {
			t.Errorf("expected first=5, got %s", first)
		}

		resp := Response[Game]{
			Data: []Game{
				{ID: "1", Name: "Top Game 1"},
				{ID: "2", Name: "Top Game 2"},
				{ID: "3", Name: "Top Game 3"},
				{ID: "4", Name: "Top Game 4"},
				{ID: "5", Name: "Top Game 5"},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetTopGames(ctx, &GetTopGamesParams{
		PaginationParams: &PaginationParams{First: 5},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 5 {
		t.Fatalf("expected 5 games, got %d", len(resp.Data))
	}
	if resp.Pagination.Cursor != "next" {
		t.Errorf("expected cursor 'next', got %s", resp.Pagination.Cursor)
	}
}

func TestClient_GetTopGames_NilParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Game]{
			Data: []Game{
				{ID: "1", Name: "Top Game"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetTopGames(ctx, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 game, got %d", len(resp.Data))
	}
}

func TestClient_GetGames_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetGames(ctx, &GetGamesParams{IDs: []string{"123"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetTopGames_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetTopGames(ctx, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

var ctx = context.Background()
