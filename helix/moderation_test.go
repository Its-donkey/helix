package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetBannedUsers(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/banned" {
			t.Errorf("expected /moderation/banned, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[BannedUser]{
			Data: []BannedUser{
				{
					UserID:         "11111",
					UserLogin:      "banned1",
					UserName:       "Banned1",
					ExpiresAt:      time.Now().Add(24 * time.Hour),
					CreatedAt:      time.Now(),
					Reason:         "Spam",
					ModeratorID:    "67890",
					ModeratorLogin: "mod1",
					ModeratorName:  "Mod1",
				},
				{
					UserID:      "22222",
					UserLogin:   "banned2",
					UserName:    "Banned2",
					CreatedAt:   time.Now(),
					Reason:      "Harassment",
					ModeratorID: "67890",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBannedUsers(context.Background(), &GetBannedUsersParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 banned users, got %d", len(resp.Data))
	}
}

func TestClient_GetBannedUsers_WithFilters(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userIDs := r.URL.Query()["user_id"]
		if len(userIDs) != 2 {
			t.Errorf("expected 2 user_ids, got %d", len(userIDs))
		}

		resp := Response[BannedUser]{Data: []BannedUser{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetBannedUsers(context.Background(), &GetBannedUsersParams{
		BroadcasterID: "12345",
		UserIDs:       []string{"11111", "22222"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_BanUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/bans" {
			t.Errorf("expected /moderation/bans, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		moderatorID := r.URL.Query().Get("moderator_id")

		if broadcasterID != "12345" || moderatorID != "67890" {
			t.Errorf("expected broadcaster_id=12345, moderator_id=67890")
		}

		var body BanUserParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Data.UserID != "11111" {
			t.Errorf("expected user_id=11111, got %s", body.Data.UserID)
		}
		if body.Data.Duration != 300 {
			t.Errorf("expected duration=300, got %d", body.Data.Duration)
		}

		resp := Response[BanUserResponse]{
			Data: []BanUserResponse{
				{
					BroadcasterID: "12345",
					ModeratorID:   "67890",
					UserID:        "11111",
					CreatedAt:     time.Now(),
					EndTime:       time.Now().Add(5 * time.Minute),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.BanUser(context.Background(), &BanUserParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		Data: BanUserData{
			UserID:   "11111",
			Duration: 300,
			Reason:   "Test ban",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != "11111" {
		t.Errorf("expected user_id 11111, got %s", result.UserID)
	}
}

func TestClient_BanUser_Permanent(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var body BanUserParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Data.Duration != 0 {
			t.Errorf("expected duration=0 for permanent ban, got %d", body.Data.Duration)
		}

		resp := Response[BanUserResponse]{
			Data: []BanUserResponse{
				{BroadcasterID: "12345", UserID: "11111"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.BanUser(context.Background(), &BanUserParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		Data: BanUserData{
			UserID:   "11111",
			Duration: 0, // Permanent
			Reason:   "Permanent ban",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UnbanUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/bans" {
			t.Errorf("expected /moderation/bans, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		moderatorID := r.URL.Query().Get("moderator_id")
		userID := r.URL.Query().Get("user_id")

		if broadcasterID != "12345" || moderatorID != "67890" || userID != "11111" {
			t.Errorf("unexpected query params")
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.UnbanUser(context.Background(), "12345", "67890", "11111")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetModerators(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/moderators" {
			t.Errorf("expected /moderation/moderators, got %s", r.URL.Path)
		}

		resp := Response[Moderator]{
			Data: []Moderator{
				{UserID: "11111", UserLogin: "mod1", UserName: "Mod1"},
				{UserID: "22222", UserLogin: "mod2", UserName: "Mod2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetModerators(context.Background(), &GetModeratorsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 moderators, got %d", len(resp.Data))
	}
}

func TestClient_AddChannelModerator(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/moderators" {
			t.Errorf("expected /moderation/moderators, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.AddChannelModerator(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_RemoveChannelModerator(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/moderators" {
			t.Errorf("expected /moderation/moderators, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.RemoveChannelModerator(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteChatMessages(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/chat" {
			t.Errorf("expected /moderation/chat, got %s", r.URL.Path)
		}

		messageID := r.URL.Query().Get("message_id")
		if messageID != "msg123" {
			t.Errorf("expected message_id=msg123, got %s", messageID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteChatMessages(context.Background(), &DeleteChatMessagesParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		MessageID:     "msg123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteChatMessages_All(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		messageID := r.URL.Query().Get("message_id")
		if messageID != "" {
			t.Errorf("expected no message_id, got %s", messageID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteChatMessages(context.Background(), &DeleteChatMessagesParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetBlockedTerms(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/blocked_terms" {
			t.Errorf("expected /moderation/blocked_terms, got %s", r.URL.Path)
		}

		resp := Response[BlockedTerm]{
			Data: []BlockedTerm{
				{
					BroadcasterID: "12345",
					ModeratorID:   "67890",
					ID:            "term1",
					Text:          "badword",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBlockedTerms(context.Background(), &GetBlockedTermsParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 blocked term, got %d", len(resp.Data))
	}
}

func TestClient_AddBlockedTerm(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/blocked_terms" {
			t.Errorf("expected /moderation/blocked_terms, got %s", r.URL.Path)
		}

		var body AddBlockedTermParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Text != "newbadword" {
			t.Errorf("expected text 'newbadword', got %s", body.Text)
		}

		resp := Response[BlockedTerm]{
			Data: []BlockedTerm{
				{ID: "newterm", Text: "newbadword"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.AddBlockedTerm(context.Background(), &AddBlockedTermParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		Text:          "newbadword",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "newbadword" {
		t.Errorf("expected text 'newbadword', got %s", result.Text)
	}
}

func TestClient_RemoveBlockedTerm(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/blocked_terms" {
			t.Errorf("expected /moderation/blocked_terms, got %s", r.URL.Path)
		}

		termID := r.URL.Query().Get("id")
		if termID != "term123" {
			t.Errorf("expected id=term123, got %s", termID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.RemoveBlockedTerm(context.Background(), "12345", "67890", "term123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetShieldModeStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/shield_mode" {
			t.Errorf("expected /moderation/shield_mode, got %s", r.URL.Path)
		}

		resp := Response[ShieldModeStatus]{
			Data: []ShieldModeStatus{
				{
					IsActive:        true,
					ModeratorID:     "67890",
					ModeratorLogin:  "mod1",
					ModeratorName:   "Mod1",
					LastActivatedAt: time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetShieldModeStatus(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsActive {
		t.Error("expected shield mode to be active")
	}
}

func TestClient_UpdateShieldModeStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/shield_mode" {
			t.Errorf("expected /moderation/shield_mode, got %s", r.URL.Path)
		}

		var body UpdateShieldModeStatusParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if !body.IsActive {
			t.Error("expected is_active to be true")
		}

		resp := Response[ShieldModeStatus]{
			Data: []ShieldModeStatus{
				{IsActive: true},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateShieldModeStatus(context.Background(), &UpdateShieldModeStatusParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		IsActive:      true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsActive {
		t.Error("expected shield mode to be active")
	}
}

func TestClient_WarnChatUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/warnings" {
			t.Errorf("expected /moderation/warnings, got %s", r.URL.Path)
		}

		var body WarnChatUserParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Data.UserID != "11111" {
			t.Errorf("expected user_id=11111, got %s", body.Data.UserID)
		}
		if body.Data.Reason != "First warning" {
			t.Errorf("expected reason 'First warning', got %s", body.Data.Reason)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.WarnChatUser(context.Background(), &WarnChatUserParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		Data: WarnChatUserData{
			UserID: "11111",
			Reason: "First warning",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CheckAutoModStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/enforcements/status" {
			t.Errorf("expected /moderation/enforcements/status, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[AutoModStatus]{
			Data: []AutoModStatus{
				{MsgID: "msg1", IsPermitted: true},
				{MsgID: "msg2", IsPermitted: false},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CheckAutoModStatus(context.Background(), &CheckAutoModStatusParams{
		BroadcasterID: "12345",
		Data: []AutoModStatusMessage{
			{MsgID: "msg1", MsgText: "Hello world"},
			{MsgID: "msg2", MsgText: "Some bad content"},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(resp.Data))
	}
	if !resp.Data[0].IsPermitted {
		t.Error("expected first message to be permitted")
	}
}

func TestClient_ManageHeldAutoModMessages(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/automod/message" {
			t.Errorf("expected /moderation/automod/message, got %s", r.URL.Path)
		}

		var body ManageHeldAutoModMessageParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.UserID != "67890" {
			t.Errorf("expected user_id=67890, got %s", body.UserID)
		}
		if body.MsgID != "msg123" {
			t.Errorf("expected msg_id=msg123, got %s", body.MsgID)
		}
		if body.Action != "ALLOW" {
			t.Errorf("expected action=ALLOW, got %s", body.Action)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.ManageHeldAutoModMessages(context.Background(), &ManageHeldAutoModMessageParams{
		UserID: "67890",
		MsgID:  "msg123",
		Action: "ALLOW",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetAutoModSettings(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/automod/settings" {
			t.Errorf("expected /moderation/automod/settings, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		moderatorID := r.URL.Query().Get("moderator_id")

		if broadcasterID != "12345" || moderatorID != "67890" {
			t.Errorf("unexpected query params")
		}

		overallLevel := 3
		resp := Response[AutoModSettings]{
			Data: []AutoModSettings{
				{
					BroadcasterID:           "12345",
					ModeratorID:             "67890",
					OverallLevel:            &overallLevel,
					Disability:              2,
					Aggression:              3,
					SexualitySexOrGender:    2,
					Misogyny:                2,
					Bullying:                3,
					Swearing:                1,
					RaceEthnicityOrReligion: 3,
					SexBasedTerms:           2,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetAutoModSettings(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected settings, got nil")
	}
	if result.OverallLevel == nil || *result.OverallLevel != 3 {
		t.Errorf("expected overall_level 3, got %v", result.OverallLevel)
	}
}

func TestClient_UpdateAutoModSettings(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		overallLevel := 4
		resp := Response[AutoModSettings]{
			Data: []AutoModSettings{
				{OverallLevel: &overallLevel},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	overallLevel := 4
	result, err := client.UpdateAutoModSettings(context.Background(), &UpdateAutoModSettingsParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		OverallLevel:  &overallLevel,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OverallLevel == nil || *result.OverallLevel != 4 {
		t.Errorf("expected overall_level 4, got %v", result.OverallLevel)
	}
}

func TestClient_GetUnbanRequests(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/unban_requests" {
			t.Errorf("expected /moderation/unban_requests, got %s", r.URL.Path)
		}

		status := r.URL.Query().Get("status")
		if status != "pending" {
			t.Errorf("expected status=pending, got %s", status)
		}

		resp := Response[UnbanRequest]{
			Data: []UnbanRequest{
				{
					ID:               "request123",
					BroadcasterID:    "12345",
					BroadcasterLogin: "streamer",
					BroadcasterName:  "Streamer",
					ModeratorID:      "",
					UserID:           "11111",
					UserLogin:        "banneduser",
					UserName:         "BannedUser",
					Text:             "Please unban me",
					Status:           "pending",
					CreatedAt:        "2024-01-14T12:00:00Z",
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUnbanRequests(context.Background(), &GetUnbanRequestsParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		Status:        "pending",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 unban request, got %d", len(resp.Data))
	}
	if resp.Data[0].Status != "pending" {
		t.Errorf("expected status 'pending', got %s", resp.Data[0].Status)
	}
}

func TestClient_ResolveUnbanRequest(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		unbanRequestID := r.URL.Query().Get("unban_request_id")
		status := r.URL.Query().Get("status")
		resolutionText := r.URL.Query().Get("resolution_text")

		if unbanRequestID != "request123" {
			t.Errorf("expected unban_request_id=request123, got %s", unbanRequestID)
		}
		if status != "approved" {
			t.Errorf("expected status=approved, got %s", status)
		}
		if resolutionText != "Appeal accepted" {
			t.Errorf("expected resolution_text='Appeal accepted', got %s", resolutionText)
		}

		resp := Response[UnbanRequest]{
			Data: []UnbanRequest{
				{
					ID:             "request123",
					Status:         "approved",
					ResolutionText: "Appeal accepted",
					ResolvedAt:     "2024-01-15T12:00:00Z",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.ResolveUnbanRequest(context.Background(), &ResolveUnbanRequestParams{
		BroadcasterID:   "12345",
		ModeratorID:     "67890",
		UnbanRequestID:  "request123",
		Status:          "approved",
		ResolutionText:  "Appeal accepted",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Status != "approved" {
		t.Errorf("expected status 'approved', got %s", result.Status)
	}
}

func TestClient_GetModeratedChannels(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/channels" {
			t.Errorf("expected /moderation/channels, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := Response[ModeratedChannel]{
			Data: []ModeratedChannel{
				{
					BroadcasterID:    "11111",
					BroadcasterLogin: "channel1",
					BroadcasterName:  "Channel1",
				},
				{
					BroadcasterID:    "22222",
					BroadcasterLogin: "channel2",
					BroadcasterName:  "Channel2",
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetModeratedChannels(context.Background(), &GetModeratedChannelsParams{
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 moderated channels, got %d", len(resp.Data))
	}
}
