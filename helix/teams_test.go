package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetChannelTeams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/teams/channel" {
			t.Errorf("expected /teams/channel, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[ChannelTeam]{
			Data: []ChannelTeam{
				{
					ID:                 "team1",
					TeamName:           "coolteam",
					TeamDisplayName:    "Cool Team",
					BroadcasterID:      "12345",
					BroadcasterLogin:   "streamer",
					BroadcasterName:    "Streamer",
					Info:               "A cool team",
					ThumbnailURL:       "https://example.com/thumb.jpg",
					BackgroundImageURL: "https://example.com/bg.jpg",
					Banner:             "https://example.com/banner.jpg",
					CreatedAt:          time.Now(),
					UpdatedAt:          time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelTeams(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 team, got %d", len(resp.Data))
	}
	if resp.Data[0].TeamName != "coolteam" {
		t.Errorf("expected coolteam, got %s", resp.Data[0].TeamName)
	}
}

func TestClient_GetChannelTeams_NoTeams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[ChannelTeam]{
			Data: []ChannelTeam{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelTeams(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 teams, got %d", len(resp.Data))
	}
}

func TestClient_GetTeams_ByName(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/teams" {
			t.Errorf("expected /teams, got %s", r.URL.Path)
		}

		name := r.URL.Query().Get("name")
		if name != "coolteam" {
			t.Errorf("expected name=coolteam, got %s", name)
		}

		resp := Response[Team]{
			Data: []Team{
				{
					ID:                 "team1",
					TeamName:           "coolteam",
					TeamDisplayName:    "Cool Team",
					Info:               "We are a cool team",
					ThumbnailURL:       "https://example.com/thumb.jpg",
					BackgroundImageURL: "https://example.com/bg.jpg",
					Banner:             "https://example.com/banner.jpg",
					CreatedAt:          time.Now(),
					UpdatedAt:          time.Now(),
					Users: []TeamUser{
						{UserID: "1", UserLogin: "user1", UserName: "User 1"},
						{UserID: "2", UserLogin: "user2", UserName: "User 2"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetTeams(context.Background(), &GetTeamsParams{
		Name: "coolteam",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 team, got %d", len(resp.Data))
	}
	if len(resp.Data[0].Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(resp.Data[0].Users))
	}
}

func TestClient_GetTeams_ByID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id != "team123" {
			t.Errorf("expected id=team123, got %s", id)
		}

		resp := Response[Team]{
			Data: []Team{
				{
					ID:              "team123",
					TeamName:        "awesometeam",
					TeamDisplayName: "Awesome Team",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetTeams(context.Background(), &GetTeamsParams{
		ID: "team123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 team, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "team123" {
		t.Errorf("expected team123, got %s", resp.Data[0].ID)
	}
}

func TestClient_GetTeams_WithUsers(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Team]{
			Data: []Team{
				{
					ID:       "team1",
					TeamName: "bigteam",
					Users: []TeamUser{
						{UserID: "1", UserLogin: "streamer1", UserName: "Streamer 1"},
						{UserID: "2", UserLogin: "streamer2", UserName: "Streamer 2"},
						{UserID: "3", UserLogin: "streamer3", UserName: "Streamer 3"},
						{UserID: "4", UserLogin: "streamer4", UserName: "Streamer 4"},
						{UserID: "5", UserLogin: "streamer5", UserName: "Streamer 5"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetTeams(context.Background(), &GetTeamsParams{
		Name: "bigteam",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data[0].Users) != 5 {
		t.Errorf("expected 5 users, got %d", len(resp.Data[0].Users))
	}
}
