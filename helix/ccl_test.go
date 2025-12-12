package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetContentClassificationLabels(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/content_classification_labels" {
			t.Errorf("expected /content_classification_labels, got %s", r.URL.Path)
		}

		locale := r.URL.Query().Get("locale")
		if locale != "en-US" {
			t.Errorf("expected locale 'en-US', got %s", locale)
		}

		resp := Response[ContentClassificationLabel]{
			Data: []ContentClassificationLabel{
				{
					ID:          "DrugsIntoxication",
					Description: "Excessive tobacco glorification or promotion",
					Name:        "Drugs, Intoxication, or Excessive Tobacco Use",
				},
				{
					ID:          "SexualThemes",
					Description: "Content that focuses on sexualized physical attributes",
					Name:        "Sexual Themes",
				},
				{
					ID:          "ViolentGraphic",
					Description: "Realistic or excessive violence or gore",
					Name:        "Violent and Graphic Depictions",
				},
				{
					ID:          "Gambling",
					Description: "Participating in online or in-person gambling",
					Name:        "Gambling",
				},
				{
					ID:          "ProfanityVulgarity",
					Description: "Excessive or extreme use of profanity",
					Name:        "Significant Profanity or Vulgarity",
				},
				{
					ID:          "MatureGame",
					Description: "Games rated 18+ by official rating boards",
					Name:        "Mature-rated game",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetContentClassificationLabels(context.Background(), &GetContentClassificationLabelsParams{
		Locale: "en-US",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 6 {
		t.Fatalf("expected 6 labels, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "DrugsIntoxication" {
		t.Errorf("expected first label ID 'DrugsIntoxication', got %s", resp.Data[0].ID)
	}
}

func TestClient_GetContentClassificationLabels_NoParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		locale := r.URL.Query().Get("locale")
		if locale != "" {
			t.Errorf("expected no locale param, got %s", locale)
		}

		resp := Response[ContentClassificationLabel]{
			Data: []ContentClassificationLabel{
				{
					ID:   "DrugsIntoxication",
					Name: "Drugs, Intoxication, or Excessive Tobacco Use",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetContentClassificationLabels(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 label, got %d", len(resp.Data))
	}
}
