package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetPredictions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/predictions" {
			t.Errorf("expected /predictions, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[Prediction]{
			Data: []Prediction{
				{
					ID:               "pred1",
					BroadcasterID:    "12345",
					BroadcasterName:  "Streamer",
					BroadcasterLogin: "streamer",
					Title:            "Will I win this game?",
					WinningOutcomeID: "",
					Outcomes: []PredictionOutcome{
						{
							ID:            "outcome1",
							Title:         "Yes",
							Users:         50,
							ChannelPoints: 10000,
							Color:         "BLUE",
							TopPredictors: []PredictionPredictor{
								{UserID: "1", UserLogin: "predictor1", UserName: "Predictor1", ChannelPointsUsed: 1000, ChannelPointsWon: 0},
							},
						},
						{
							ID:            "outcome2",
							Title:         "No",
							Users:         30,
							ChannelPoints: 5000,
							Color:         "PINK",
						},
					},
					PredictionWindow: 120,
					Status:           "ACTIVE",
					CreatedAt:        time.Now(),
				},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetPredictions(context.Background(), &GetPredictionsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 prediction, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "Will I win this game?" {
		t.Errorf("unexpected title: %s", resp.Data[0].Title)
	}
	if len(resp.Data[0].Outcomes) != 2 {
		t.Errorf("expected 2 outcomes, got %d", len(resp.Data[0].Outcomes))
	}
}

func TestClient_GetPredictions_ByIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[Prediction]{
			Data: []Prediction{
				{ID: "pred1", Title: "Prediction 1"},
				{ID: "pred2", Title: "Prediction 2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetPredictions(context.Background(), &GetPredictionsParams{
		BroadcasterID: "12345",
		IDs:           []string{"pred1", "pred2"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 predictions, got %d", len(resp.Data))
	}
}

func TestClient_CreatePrediction(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/predictions" {
			t.Errorf("expected /predictions, got %s", r.URL.Path)
		}

		var params CreatePredictionParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Title != "Will I clutch?" {
			t.Errorf("expected title 'Will I clutch?', got %s", params.Title)
		}
		if len(params.Outcomes) != 2 {
			t.Errorf("expected 2 outcomes, got %d", len(params.Outcomes))
		}
		if params.PredictionWindow != 60 {
			t.Errorf("expected prediction_window 60, got %d", params.PredictionWindow)
		}

		resp := Response[Prediction]{
			Data: []Prediction{
				{
					ID:               "newpred",
					BroadcasterID:    params.BroadcasterID,
					Title:            params.Title,
					Status:           "ACTIVE",
					PredictionWindow: params.PredictionWindow,
					CreatedAt:        time.Now(),
					Outcomes: []PredictionOutcome{
						{ID: "1", Title: "Yes", Color: "BLUE"},
						{ID: "2", Title: "No", Color: "PINK"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreatePrediction(context.Background(), &CreatePredictionParams{
		BroadcasterID: "12345",
		Title:         "Will I clutch?",
		Outcomes: []CreatePredictionOutcome{
			{Title: "Yes"},
			{Title: "No"},
		},
		PredictionWindow: 60,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != "newpred" {
		t.Errorf("expected prediction ID 'newpred', got %s", result.ID)
	}
	if result.Status != "ACTIVE" {
		t.Errorf("expected status ACTIVE, got %s", result.Status)
	}
}

func TestClient_EndPrediction_Resolve(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/predictions" {
			t.Errorf("expected /predictions, got %s", r.URL.Path)
		}

		var params EndPredictionParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Status != "RESOLVED" {
			t.Errorf("expected status RESOLVED, got %s", params.Status)
		}
		if params.WinningOutcomeID != "outcome1" {
			t.Errorf("expected winning_outcome_id=outcome1, got %s", params.WinningOutcomeID)
		}

		resp := Response[Prediction]{
			Data: []Prediction{
				{
					ID:               params.ID,
					Status:           "RESOLVED",
					WinningOutcomeID: params.WinningOutcomeID,
					EndedAt:          time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.EndPrediction(context.Background(), &EndPredictionParams{
		BroadcasterID:    "12345",
		ID:               "pred123",
		Status:           "RESOLVED",
		WinningOutcomeID: "outcome1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "RESOLVED" {
		t.Errorf("expected status RESOLVED, got %s", result.Status)
	}
	if result.WinningOutcomeID != "outcome1" {
		t.Errorf("expected winning_outcome_id=outcome1, got %s", result.WinningOutcomeID)
	}
}

func TestClient_EndPrediction_Cancel(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params EndPredictionParams
		_ = json.NewDecoder(r.Body).Decode(&params)
		if params.Status != "CANCELED" {
			t.Errorf("expected status CANCELED, got %s", params.Status)
		}

		resp := Response[Prediction]{
			Data: []Prediction{
				{
					ID:     params.ID,
					Status: "CANCELED",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.EndPrediction(context.Background(), &EndPredictionParams{
		BroadcasterID: "12345",
		ID:            "pred123",
		Status:        "CANCELED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "CANCELED" {
		t.Errorf("expected status CANCELED, got %s", result.Status)
	}
}

func TestClient_EndPrediction_Lock(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params EndPredictionParams
		json.NewDecoder(r.Body).Decode(&params)
		if params.Status != "LOCKED" {
			t.Errorf("expected status LOCKED, got %s", params.Status)
		}

		resp := Response[Prediction]{
			Data: []Prediction{
				{
					ID:       params.ID,
					Status:   "LOCKED",
					LockedAt: time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.EndPrediction(context.Background(), &EndPredictionParams{
		BroadcasterID: "12345",
		ID:            "pred123",
		Status:        "LOCKED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "LOCKED" {
		t.Errorf("expected status LOCKED, got %s", result.Status)
	}
}
