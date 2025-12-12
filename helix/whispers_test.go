package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_SendWhisper(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/whispers" {
			t.Errorf("expected /whispers, got %s", r.URL.Path)
		}

		fromUserID := r.URL.Query().Get("from_user_id")
		toUserID := r.URL.Query().Get("to_user_id")

		if fromUserID != "12345" {
			t.Errorf("expected from_user_id=12345, got %s", fromUserID)
		}
		if toUserID != "67890" {
			t.Errorf("expected to_user_id=67890, got %s", toUserID)
		}

		var params SendWhisperParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Message != "Hello from the bot!" {
			t.Errorf("expected message 'Hello from the bot!', got %s", params.Message)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendWhisper(context.Background(), &SendWhisperParams{
		FromUserID: "12345",
		ToUserID:   "67890",
		Message:    "Hello from the bot!",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendWhisper_LongMessage(t *testing.T) {
	longMessage := ""
	for i := 0; i < 100; i++ {
		longMessage += "This is a long message. "
	}

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params SendWhisperParams
		_ = json.NewDecoder(r.Body).Decode(&params)
		if params.Message != longMessage {
			t.Errorf("message mismatch")
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendWhisper(context.Background(), &SendWhisperParams{
		FromUserID: "12345",
		ToUserID:   "67890",
		Message:    longMessage,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendWhisper_EmptyMessage(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params SendWhisperParams
		_ = json.NewDecoder(r.Body).Decode(&params)
		if params.Message != "" {
			t.Errorf("expected empty message, got %s", params.Message)
		}

		// In practice, Twitch would return an error for empty message
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendWhisper(context.Background(), &SendWhisperParams{
		FromUserID: "12345",
		ToUserID:   "67890",
		Message:    "",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendWhisper_SpecialCharacters(t *testing.T) {
	specialMessage := "Hello! 👋 How are you? <script>alert('test')</script> & @user"

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params SendWhisperParams
		_ = json.NewDecoder(r.Body).Decode(&params)
		if params.Message != specialMessage {
			t.Errorf("message mismatch: got %s", params.Message)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendWhisper(context.Background(), &SendWhisperParams{
		FromUserID: "12345",
		ToUserID:   "67890",
		Message:    specialMessage,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
