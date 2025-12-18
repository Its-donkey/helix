package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetChannelGuestStarSettings(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/guest_star_settings" {
			t.Errorf("expected /channels/guest_star_settings, got %s", r.URL.Path)
		}

		resp := Response[GuestStarSettings]{
			Data: []GuestStarSettings{
				{
					IsModeratorSendLiveEnabled:  true,
					SlotCount:                   4,
					IsBrowserSourceAudioEnabled: true,
					GroupLayout:                 "TILED_LAYOUT",
					BrowserSourceToken:          "token123",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelGuestStarSettings(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected settings, got nil")
	}
	if resp.SlotCount != 4 {
		t.Errorf("expected slot_count 4, got %d", resp.SlotCount)
	}
}

func TestClient_UpdateChannelGuestStarSettings(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body UpdateChannelGuestStarSettingsParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if *body.SlotCount != 6 {
			t.Errorf("expected slot_count 6, got %d", *body.SlotCount)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	slotCount := 6
	err := client.UpdateChannelGuestStarSettings(context.Background(), &UpdateChannelGuestStarSettingsParams{
		BroadcasterID: "12345",
		SlotCount:     &slotCount,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetGuestStarSession(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guest_star/session" {
			t.Errorf("expected /guest_star/session, got %s", r.URL.Path)
		}

		resp := Response[GuestStarSession]{
			Data: []GuestStarSession{
				{
					ID: "session123",
					Guests: []GuestStarGuest{
						{
							SlotID:          "1",
							IsLive:          true,
							UserID:          "67890",
							UserLogin:       "guest1",
							UserDisplayName: "Guest1",
							Volume:          100,
							AssignedAt:      "2024-01-15T12:00:00Z",
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGuestStarSession(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected session, got nil")
	}
	if len(resp.Guests) != 1 {
		t.Fatalf("expected 1 guest, got %d", len(resp.Guests))
	}
}

func TestClient_CreateGuestStarSession(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		resp := Response[GuestStarSession]{
			Data: []GuestStarSession{
				{
					ID:     "newsession",
					Guests: []GuestStarGuest{},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CreateGuestStarSession(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected session, got nil")
	}
	if resp.ID != "newsession" {
		t.Errorf("expected session ID 'newsession', got %s", resp.ID)
	}
}

func TestClient_EndGuestStarSession(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		sessionID := r.URL.Query().Get("session_id")
		if sessionID != "session123" {
			t.Errorf("expected session_id 'session123', got %s", sessionID)
		}

		resp := Response[GuestStarSession]{
			Data: []GuestStarSession{
				{ID: "session123"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.EndGuestStarSession(context.Background(), "12345", "session123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected session, got nil")
	}
}

func TestClient_GetGuestStarInvites(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guest_star/invites" {
			t.Errorf("expected /guest_star/invites, got %s", r.URL.Path)
		}

		resp := Response[GuestStarInvite]{
			Data: []GuestStarInvite{
				{
					UserID:           "67890",
					InvitedAt:        "2024-01-15T12:00:00Z",
					Status:           "INVITED",
					IsAudioEnabled:   true,
					IsVideoEnabled:   true,
					IsAudioAvailable: true,
					IsVideoAvailable: true,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGuestStarInvites(context.Background(), "12345", "67890", "session123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 invite, got %d", len(resp.Data))
	}
}

func TestClient_SendGuestStarInvite(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		guestID := r.URL.Query().Get("guest_id")
		if guestID != "99999" {
			t.Errorf("expected guest_id '99999', got %s", guestID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendGuestStarInvite(context.Background(), "12345", "67890", "session123", "99999")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteGuestStarInvite(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteGuestStarInvite(context.Background(), "12345", "67890", "session123", "99999")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AssignGuestStarSlot(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/guest_star/slot" {
			t.Errorf("expected /guest_star/slot, got %s", r.URL.Path)
		}

		slotID := r.URL.Query().Get("slot_id")
		if slotID != "1" {
			t.Errorf("expected slot_id '1', got %s", slotID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.AssignGuestStarSlot(context.Background(), "12345", "67890", "session123", "99999", "1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UpdateGuestStarSlot(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.UpdateGuestStarSlot(context.Background(), &UpdateGuestStarSlotParams{
		BroadcasterID:     "12345",
		ModeratorID:       "67890",
		SessionID:         "session123",
		SourceSlotID:      "1",
		DestinationSlotID: "2",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteGuestStarSlot(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		slotID := r.URL.Query().Get("slot_id")
		if slotID != "1" {
			t.Errorf("expected slot_id '1', got %s", slotID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteGuestStarSlot(context.Background(), "12345", "67890", "session123", "99999", "1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UpdateGuestStarSlotSettings(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/guest_star/slot_settings" {
			t.Errorf("expected /guest_star/slot_settings, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	isAudioEnabled := true
	err := client.UpdateGuestStarSlotSettings(context.Background(), &UpdateGuestStarSlotSettingsParams{
		BroadcasterID:  "12345",
		ModeratorID:    "67890",
		SessionID:      "session123",
		SlotID:         "1",
		IsAudioEnabled: &isAudioEnabled,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetChannelGuestStarSettings_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetChannelGuestStarSettings(context.Background(), "12345", "67890")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetChannelGuestStarSettings_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[GuestStarSettings]{
			Data: []GuestStarSettings{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelGuestStarSettings(context.Background(), "12345", "67890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil, got settings")
	}
}

func TestClient_UpdateChannelGuestStarSettings_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	err := client.UpdateChannelGuestStarSettings(context.Background(), &UpdateChannelGuestStarSettingsParams{
		BroadcasterID: "12345",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetGuestStarSession_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	_, err := client.GetGuestStarSession(context.Background(), "12345", "67890")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetGuestStarSession_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[GuestStarSession]{
			Data: []GuestStarSession{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGuestStarSession(context.Background(), "12345", "67890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil, got session")
	}
}

func TestClient_CreateGuestStarSession_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.CreateGuestStarSession(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateGuestStarSession_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[GuestStarSession]{
			Data: []GuestStarSession{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CreateGuestStarSession(context.Background(), "12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil, got session")
	}
}

func TestClient_EndGuestStarSession_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	_, err := client.EndGuestStarSession(context.Background(), "12345", "session123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_EndGuestStarSession_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[GuestStarSession]{
			Data: []GuestStarSession{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.EndGuestStarSession(context.Background(), "12345", "session123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil, got session")
	}
}

func TestClient_GetGuestStarInvites_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetGuestStarInvites(context.Background(), "12345", "67890", "session123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SendGuestStarInvite_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	err := client.SendGuestStarInvite(context.Background(), "12345", "67890", "session123", "99999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_DeleteGuestStarInvite_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	err := client.DeleteGuestStarInvite(context.Background(), "12345", "67890", "session123", "99999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_AssignGuestStarSlot_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"conflict"}`))
	})
	defer server.Close()

	err := client.AssignGuestStarSlot(context.Background(), "12345", "67890", "session123", "99999", "1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateGuestStarSlot_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	err := client.UpdateGuestStarSlot(context.Background(), &UpdateGuestStarSlotParams{
		BroadcasterID:     "12345",
		ModeratorID:       "67890",
		SessionID:         "session123",
		SourceSlotID:      "1",
		DestinationSlotID: "2",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_DeleteGuestStarSlot_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	err := client.DeleteGuestStarSlot(context.Background(), "12345", "67890", "session123", "99999", "1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateGuestStarSlotSettings_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	err := client.UpdateGuestStarSlotSettings(context.Background(), &UpdateGuestStarSlotSettingsParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		SessionID:     "session123",
		SlotID:        "1",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
