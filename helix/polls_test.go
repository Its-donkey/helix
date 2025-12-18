package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetPolls(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/polls" {
			t.Errorf("expected /polls, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[Poll]{
			Data: []Poll{
				{
					ID:               "poll1",
					BroadcasterID:    "12345",
					BroadcasterName:  "Streamer",
					BroadcasterLogin: "streamer",
					Title:            "What should I play next?",
					Choices: []PollChoice{
						{ID: "1", Title: "Game A", Votes: 100, ChannelPointsVotes: 500, BitsVotes: 50},
						{ID: "2", Title: "Game B", Votes: 150, ChannelPointsVotes: 300, BitsVotes: 25},
					},
					BitsVotingEnabled:          true,
					BitsPerVote:                10,
					ChannelPointsVotingEnabled: true,
					ChannelPointsPerVote:       100,
					Status:                     "ACTIVE",
					Duration:                   300,
					StartedAt:                  time.Now(),
				},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetPolls(context.Background(), &GetPollsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 poll, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "What should I play next?" {
		t.Errorf("unexpected title: %s", resp.Data[0].Title)
	}
	if len(resp.Data[0].Choices) != 2 {
		t.Errorf("expected 2 choices, got %d", len(resp.Data[0].Choices))
	}
}

func TestClient_GetPolls_ByIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[Poll]{
			Data: []Poll{
				{ID: "poll1", Title: "Poll 1"},
				{ID: "poll2", Title: "Poll 2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetPolls(context.Background(), &GetPollsParams{
		BroadcasterID: "12345",
		IDs:           []string{"poll1", "poll2"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 polls, got %d", len(resp.Data))
	}
}

func TestClient_CreatePoll(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/polls" {
			t.Errorf("expected /polls, got %s", r.URL.Path)
		}

		var params CreatePollParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Title != "Best game?" {
			t.Errorf("expected title 'Best game?', got %s", params.Title)
		}
		if len(params.Choices) != 3 {
			t.Errorf("expected 3 choices, got %d", len(params.Choices))
		}
		if params.Duration != 300 {
			t.Errorf("expected duration 300, got %d", params.Duration)
		}

		resp := Response[Poll]{
			Data: []Poll{
				{
					ID:                         "newpoll",
					BroadcasterID:              params.BroadcasterID,
					Title:                      params.Title,
					Status:                     "ACTIVE",
					Duration:                   params.Duration,
					StartedAt:                  time.Now(),
					ChannelPointsVotingEnabled: params.ChannelPointsVotingEnabled,
					ChannelPointsPerVote:       params.ChannelPointsPerVote,
					Choices: []PollChoice{
						{ID: "1", Title: "Option A"},
						{ID: "2", Title: "Option B"},
						{ID: "3", Title: "Option C"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreatePoll(context.Background(), &CreatePollParams{
		BroadcasterID: "12345",
		Title:         "Best game?",
		Choices: []CreatePollChoice{
			{Title: "Option A"},
			{Title: "Option B"},
			{Title: "Option C"},
		},
		Duration:                   300,
		ChannelPointsVotingEnabled: true,
		ChannelPointsPerVote:       500,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != "newpoll" {
		t.Errorf("expected poll ID 'newpoll', got %s", result.ID)
	}
	if result.Status != "ACTIVE" {
		t.Errorf("expected status ACTIVE, got %s", result.Status)
	}
}

func TestClient_EndPoll_Terminate(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/polls" {
			t.Errorf("expected /polls, got %s", r.URL.Path)
		}

		var params EndPollParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Status != "TERMINATED" {
			t.Errorf("expected status TERMINATED, got %s", params.Status)
		}

		resp := Response[Poll]{
			Data: []Poll{
				{
					ID:      params.ID,
					Status:  "TERMINATED",
					EndedAt: time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.EndPoll(context.Background(), &EndPollParams{
		BroadcasterID: "12345",
		ID:            "poll123",
		Status:        "TERMINATED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "TERMINATED" {
		t.Errorf("expected status TERMINATED, got %s", result.Status)
	}
}

func TestClient_EndPoll_Archive(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params EndPollParams
		_ = json.NewDecoder(r.Body).Decode(&params)
		if params.Status != "ARCHIVED" {
			t.Errorf("expected status ARCHIVED, got %s", params.Status)
		}

		resp := Response[Poll]{
			Data: []Poll{
				{
					ID:     params.ID,
					Status: "ARCHIVED",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.EndPoll(context.Background(), &EndPollParams{
		BroadcasterID: "12345",
		ID:            "poll123",
		Status:        "ARCHIVED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ARCHIVED" {
		t.Errorf("expected status ARCHIVED, got %s", result.Status)
	}
}

func TestClient_GetPolls_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetPolls(context.Background(), &GetPollsParams{
		BroadcasterID: "12345",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreatePoll_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.CreatePoll(context.Background(), &CreatePollParams{
		BroadcasterID: "12345",
		Title:         "Test Poll",
		Choices:       []CreatePollChoice{{Title: "A"}, {Title: "B"}},
		Duration:      60,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_EndPoll_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"poll not found"}`))
	})
	defer server.Close()

	_, err := client.EndPoll(context.Background(), &EndPollParams{
		BroadcasterID: "12345",
		ID:            "poll123",
		Status:        "TERMINATED",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreatePoll_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Poll]{Data: []Poll{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreatePoll(context.Background(), &CreatePollParams{
		BroadcasterID: "12345",
		Title:         "Test Poll",
		Choices:       []CreatePollChoice{{Title: "A"}, {Title: "B"}},
		Duration:      60,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_EndPoll_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Poll]{Data: []Poll{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.EndPoll(context.Background(), &EndPollParams{
		BroadcasterID: "12345",
		ID:            "poll123",
		Status:        "TERMINATED",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}
