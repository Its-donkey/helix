package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// Official Twitch API example values from https://dev.twitch.tv/docs/api/reference/
const (
	// Get Banned Users example
	twitchBannedUserID         = "423374343"
	twitchBannedUserLogin      = "glowillig"
	twitchBannedUserName       = "glowillig"
	twitchBannedModeratorID    = "141981764"
	twitchBannedModeratorLogin = "twitchdev"
	twitchBannedModeratorName  = "TwitchDev"
	twitchBannedReason         = "Does not like pineapple on pizza."

	// Ban User example
	twitchBanBroadcasterID = "1234"
	twitchBanModeratorID   = "5678"
	twitchBanUserID        = "9876"

	// Get Moderators example
	twitchModeratorUserID    = "424596340"
	twitchModeratorUserLogin = "quotrok"
	twitchModeratorUserName  = "quotrok"

	// Blocked Terms example
	twitchBlockedTermBroadcasterID = "1234"
	twitchBlockedTermModeratorID   = "5678"
	twitchBlockedTermID            = "520e4d4e-0cda-49c7-821e-e5ef4f88c2f2"
	twitchBlockedTermText          = "A phrase I'm not fond of"

	// Shield Mode example
	twitchShieldModeModeratorID    = "98765"
	twitchShieldModeModeratorLogin = "simplysimple"
	twitchShieldModeModeratorName  = "SimplySimple"
)

func TestClient_GetBannedUsers(t *testing.T) {
	// Using official Twitch API example from https://dev.twitch.tv/docs/api/reference/#get-banned-users
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/banned" {
			t.Errorf("expected /moderation/banned, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != twitchBannedModeratorID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchBannedModeratorID, broadcasterID)
		}

		// Official Twitch API response example
		resp := Response[BannedUser]{
			Data: []BannedUser{
				{
					UserID:         twitchBannedUserID,
					UserLogin:      twitchBannedUserLogin,
					UserName:       twitchBannedUserName,
					ExpiresAt:      mustParseTime("2022-03-15T02:00:28Z"),
					CreatedAt:      mustParseTime("2022-03-15T01:30:28Z"),
					Reason:         twitchBannedReason,
					ModeratorID:    twitchBannedModeratorID,
					ModeratorLogin: twitchBannedModeratorLogin,
					ModeratorName:  twitchBannedModeratorName,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBannedUsers(context.Background(), &GetBannedUsersParams{
		BroadcasterID: twitchBannedModeratorID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 banned user, got %d", len(resp.Data))
	}
	if resp.Data[0].UserID != twitchBannedUserID {
		t.Errorf("expected user_id %s, got %s", twitchBannedUserID, resp.Data[0].UserID)
	}
	if resp.Data[0].Reason != twitchBannedReason {
		t.Errorf("expected reason %q, got %q", twitchBannedReason, resp.Data[0].Reason)
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
		BroadcasterID: twitchBannedModeratorID,
		UserIDs:       []string{twitchBannedUserID, "999999999"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_BanUser(t *testing.T) {
	// Using official Twitch API example from https://dev.twitch.tv/docs/api/reference/#ban-user
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/bans" {
			t.Errorf("expected /moderation/bans, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		moderatorID := r.URL.Query().Get("moderator_id")

		if broadcasterID != twitchBanBroadcasterID || moderatorID != twitchBanModeratorID {
			t.Errorf("expected broadcaster_id=%s, moderator_id=%s", twitchBanBroadcasterID, twitchBanModeratorID)
		}

		var body BanUserParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Data.UserID != twitchBanUserID {
			t.Errorf("expected user_id=%s, got %s", twitchBanUserID, body.Data.UserID)
		}

		// Official Twitch API response example
		resp := Response[BanUserResponse]{
			Data: []BanUserResponse{
				{
					BroadcasterID: twitchBanBroadcasterID,
					ModeratorID:   twitchBanModeratorID,
					UserID:        twitchBanUserID,
					CreatedAt:     mustParseTime("2021-09-28T18:22:31Z"),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.BanUser(context.Background(), &BanUserParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		Data: BanUserData{
			UserID: twitchBanUserID,
			Reason: "Test ban",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != twitchBanUserID {
		t.Errorf("expected user_id %s, got %s", twitchBanUserID, result.UserID)
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
				{BroadcasterID: twitchBanBroadcasterID, UserID: twitchBanUserID},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.BanUser(context.Background(), &BanUserParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		Data: BanUserData{
			UserID:   twitchBanUserID,
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

		if broadcasterID != twitchBanBroadcasterID || moderatorID != twitchBanModeratorID || userID != twitchBanUserID {
			t.Errorf("unexpected query params")
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.UnbanUser(context.Background(), twitchBanBroadcasterID, twitchBanModeratorID, twitchBanUserID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetModerators(t *testing.T) {
	// Using official Twitch API example from https://dev.twitch.tv/docs/api/reference/#get-moderators
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/moderators" {
			t.Errorf("expected /moderation/moderators, got %s", r.URL.Path)
		}

		// Official Twitch API response example
		resp := Response[Moderator]{
			Data: []Moderator{
				{
					UserID:    twitchModeratorUserID,
					UserLogin: twitchModeratorUserLogin,
					UserName:  twitchModeratorUserName,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetModerators(context.Background(), &GetModeratorsParams{
		BroadcasterID: twitchBanBroadcasterID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 moderator, got %d", len(resp.Data))
	}
	if resp.Data[0].UserID != twitchModeratorUserID {
		t.Errorf("expected user_id %s, got %s", twitchModeratorUserID, resp.Data[0].UserID)
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

	err := client.AddChannelModerator(context.Background(), twitchBanBroadcasterID, twitchModeratorUserID)

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

	err := client.RemoveChannelModerator(context.Background(), twitchBanBroadcasterID, twitchModeratorUserID)

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
		if messageID != "abc-123-def" {
			t.Errorf("expected message_id=abc-123-def, got %s", messageID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteChatMessages(context.Background(), &DeleteChatMessagesParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		MessageID:     "abc-123-def",
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
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetBlockedTerms(t *testing.T) {
	// Using official Twitch API example from https://dev.twitch.tv/docs/api/reference/#get-blocked-terms
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/blocked_terms" {
			t.Errorf("expected /moderation/blocked_terms, got %s", r.URL.Path)
		}

		// Official Twitch API response example
		resp := Response[BlockedTerm]{
			Data: []BlockedTerm{
				{
					BroadcasterID: twitchBlockedTermBroadcasterID,
					ModeratorID:   twitchBlockedTermModeratorID,
					ID:            twitchBlockedTermID,
					Text:          twitchBlockedTermText,
					CreatedAt:     mustParseTime("2021-09-29T19:45:37Z"),
					UpdatedAt:     mustParseTime("2021-09-29T19:45:37Z"),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetBlockedTerms(context.Background(), &GetBlockedTermsParams{
		BroadcasterID: twitchBlockedTermBroadcasterID,
		ModeratorID:   twitchBlockedTermModeratorID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 blocked term, got %d", len(resp.Data))
	}
	if resp.Data[0].Text != twitchBlockedTermText {
		t.Errorf("expected text %q, got %q", twitchBlockedTermText, resp.Data[0].Text)
	}
}

func TestClient_AddBlockedTerm(t *testing.T) {
	// Using official Twitch API example from https://dev.twitch.tv/docs/api/reference/#add-blocked-term
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/blocked_terms" {
			t.Errorf("expected /moderation/blocked_terms, got %s", r.URL.Path)
		}

		var body AddBlockedTermParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Text != twitchBlockedTermText {
			t.Errorf("expected text %q, got %s", twitchBlockedTermText, body.Text)
		}

		// Official Twitch API response example
		resp := Response[BlockedTerm]{
			Data: []BlockedTerm{
				{
					BroadcasterID: twitchBlockedTermBroadcasterID,
					ModeratorID:   twitchBlockedTermModeratorID,
					ID:            twitchBlockedTermID,
					Text:          twitchBlockedTermText,
					CreatedAt:     mustParseTime("2021-09-29T19:45:37Z"),
					UpdatedAt:     mustParseTime("2021-09-29T19:45:37Z"),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.AddBlockedTerm(context.Background(), &AddBlockedTermParams{
		BroadcasterID: twitchBlockedTermBroadcasterID,
		ModeratorID:   twitchBlockedTermModeratorID,
		Text:          twitchBlockedTermText,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != twitchBlockedTermText {
		t.Errorf("expected text %q, got %s", twitchBlockedTermText, result.Text)
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
		if termID != twitchBlockedTermID {
			t.Errorf("expected id=%s, got %s", twitchBlockedTermID, termID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.RemoveBlockedTerm(context.Background(), twitchBlockedTermBroadcasterID, twitchBlockedTermModeratorID, twitchBlockedTermID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetShieldModeStatus(t *testing.T) {
	// Using official Twitch API example from https://dev.twitch.tv/docs/api/reference/#get-shield-mode-status
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderation/shield_mode" {
			t.Errorf("expected /moderation/shield_mode, got %s", r.URL.Path)
		}

		// Official Twitch API response example
		resp := Response[ShieldModeStatus]{
			Data: []ShieldModeStatus{
				{
					IsActive:        true,
					ModeratorID:     twitchShieldModeModeratorID,
					ModeratorLogin:  twitchShieldModeModeratorLogin,
					ModeratorName:   twitchShieldModeModeratorName,
					LastActivatedAt: mustParseTime("2022-07-26T17:16:03.123Z"),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetShieldModeStatus(context.Background(), twitchBanBroadcasterID, twitchShieldModeModeratorID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsActive {
		t.Error("expected shield mode to be active")
	}
	if result.ModeratorID != twitchShieldModeModeratorID {
		t.Errorf("expected moderator_id %s, got %s", twitchShieldModeModeratorID, result.ModeratorID)
	}
}

func TestClient_UpdateShieldModeStatus(t *testing.T) {
	// Using official Twitch API example from https://dev.twitch.tv/docs/api/reference/#update-shield-mode-status
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

		// Official Twitch API response example
		resp := Response[ShieldModeStatus]{
			Data: []ShieldModeStatus{
				{
					IsActive:        true,
					ModeratorID:     twitchShieldModeModeratorID,
					ModeratorLogin:  twitchShieldModeModeratorLogin,
					ModeratorName:   twitchShieldModeModeratorName,
					LastActivatedAt: mustParseTime("2022-07-26T17:16:03.123Z"),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateShieldModeStatus(context.Background(), &UpdateShieldModeStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchShieldModeModeratorID,
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

		if body.Data.UserID != twitchBanUserID {
			t.Errorf("expected user_id=%s, got %s", twitchBanUserID, body.Data.UserID)
		}
		if body.Data.Reason != "First warning" {
			t.Errorf("expected reason 'First warning', got %s", body.Data.Reason)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.WarnChatUser(context.Background(), &WarnChatUserParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		Data: WarnChatUserData{
			UserID: twitchBanUserID,
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
		if broadcasterID != twitchBanBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchBanBroadcasterID, broadcasterID)
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
		BroadcasterID: twitchBanBroadcasterID,
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

		if body.UserID != twitchBanModeratorID {
			t.Errorf("expected user_id=%s, got %s", twitchBanModeratorID, body.UserID)
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
		UserID: twitchBanModeratorID,
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

		if broadcasterID != twitchBanBroadcasterID || moderatorID != twitchBanModeratorID {
			t.Errorf("unexpected query params")
		}

		overallLevel := 3
		resp := Response[AutoModSettings]{
			Data: []AutoModSettings{
				{
					BroadcasterID:           twitchBanBroadcasterID,
					ModeratorID:             twitchBanModeratorID,
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

	result, err := client.GetAutoModSettings(context.Background(), twitchBanBroadcasterID, twitchBanModeratorID)

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
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
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
					BroadcasterID:    twitchBanBroadcasterID,
					BroadcasterLogin: "streamer",
					BroadcasterName:  "Streamer",
					ModeratorID:      "",
					UserID:           twitchBanUserID,
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
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
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
		BroadcasterID:  twitchBanBroadcasterID,
		ModeratorID:    twitchBanModeratorID,
		UnbanRequestID: "request123",
		Status:         "approved",
		ResolutionText: "Appeal accepted",
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
		if userID != twitchModeratorUserID {
			t.Errorf("expected user_id=%s, got %s", twitchModeratorUserID, userID)
		}

		resp := Response[ModeratedChannel]{
			Data: []ModeratedChannel{
				{
					BroadcasterID:    twitchBanBroadcasterID,
					BroadcasterLogin: "channel1",
					BroadcasterName:  "Channel1",
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetModeratedChannels(context.Background(), &GetModeratedChannelsParams{
		UserID: twitchModeratorUserID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 moderated channel, got %d", len(resp.Data))
	}
}

// Error handling tests

func TestClient_BanUser_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.BanUser(context.Background(), &BanUserParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		Data:          BanUserData{UserID: twitchBanUserID},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetModerators_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetModerators(context.Background(), &GetModeratorsParams{BroadcasterID: twitchBanBroadcasterID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_AddBlockedTerm_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.AddBlockedTerm(context.Background(), &AddBlockedTermParams{
		BroadcasterID: twitchBlockedTermBroadcasterID,
		ModeratorID:   twitchBlockedTermModeratorID,
		Text:          "badword",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetShieldModeStatus_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetShieldModeStatus(context.Background(), twitchBanBroadcasterID, twitchBanModeratorID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateShieldModeStatus_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.UpdateShieldModeStatus(context.Background(), &UpdateShieldModeStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		IsActive:      true,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetAutoModSettings_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetAutoModSettings(context.Background(), twitchBanBroadcasterID, twitchBanModeratorID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateAutoModSettings_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.UpdateAutoModSettings(context.Background(), &UpdateAutoModSettingsParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetBannedUsers_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetBannedUsers(context.Background(), &GetBannedUsersParams{BroadcasterID: twitchBanBroadcasterID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetBlockedTerms_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetBlockedTerms(context.Background(), &GetBlockedTermsParams{
		BroadcasterID: twitchBlockedTermBroadcasterID,
		ModeratorID:   twitchBlockedTermModeratorID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CheckAutoModStatus_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.CheckAutoModStatus(context.Background(), &CheckAutoModStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		Data:          []AutoModStatusMessage{{MsgID: "1", MsgText: "test"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetUnbanRequests_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetUnbanRequests(context.Background(), &GetUnbanRequestsParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		Status:        "pending",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ResolveUnbanRequest_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.ResolveUnbanRequest(context.Background(), &ResolveUnbanRequestParams{
		BroadcasterID:  twitchBanBroadcasterID,
		ModeratorID:    twitchBanModeratorID,
		UnbanRequestID: "req123",
		Status:         "approved",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetModeratedChannels_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetModeratedChannels(context.Background(), &GetModeratedChannelsParams{UserID: twitchModeratorUserID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Empty response tests

func TestClient_BanUser_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[BanUserResponse]{Data: []BanUserResponse{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.BanUser(context.Background(), &BanUserParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		Data:          BanUserData{UserID: twitchBanUserID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_GetModerators_WithUserIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userIDs := r.URL.Query()["user_id"]
		if len(userIDs) != 2 {
			t.Errorf("expected 2 user_ids, got %d", len(userIDs))
		}

		resp := Response[Moderator]{
			Data: []Moderator{
				{UserID: twitchModeratorUserID, UserLogin: twitchModeratorUserLogin, UserName: twitchModeratorUserName},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetModerators(context.Background(), &GetModeratorsParams{
		BroadcasterID: twitchBanBroadcasterID,
		UserIDs:       []string{twitchModeratorUserID, "999999999"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 moderator, got %d", len(resp.Data))
	}
}

func TestClient_AddBlockedTerm_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[BlockedTerm]{Data: []BlockedTerm{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.AddBlockedTerm(context.Background(), &AddBlockedTermParams{
		BroadcasterID: twitchBlockedTermBroadcasterID,
		ModeratorID:   twitchBlockedTermModeratorID,
		Text:          "badword",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_GetShieldModeStatus_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[ShieldModeStatus]{Data: []ShieldModeStatus{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetShieldModeStatus(context.Background(), twitchBanBroadcasterID, twitchBanModeratorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_UpdateShieldModeStatus_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[ShieldModeStatus]{Data: []ShieldModeStatus{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateShieldModeStatus(context.Background(), &UpdateShieldModeStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_GetAutoModSettings_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[AutoModSettings]{Data: []AutoModSettings{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetAutoModSettings(context.Background(), twitchBanBroadcasterID, twitchBanModeratorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_UpdateAutoModSettings_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[AutoModSettings]{Data: []AutoModSettings{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateAutoModSettings(context.Background(), &UpdateAutoModSettingsParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_ResolveUnbanRequest_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[UnbanRequest]{Data: []UnbanRequest{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.ResolveUnbanRequest(context.Background(), &ResolveUnbanRequestParams{
		BroadcasterID:  twitchBanBroadcasterID,
		ModeratorID:    twitchBanModeratorID,
		UnbanRequestID: "req123",
		Status:         "approved",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestClient_GetUnbanRequests_WithUserID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID != twitchBanUserID {
			t.Errorf("expected user_id=%s, got %s", twitchBanUserID, userID)
		}

		resp := Response[UnbanRequest]{
			Data: []UnbanRequest{
				{ID: "req1", UserID: twitchBanUserID, Status: "pending"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUnbanRequests(context.Background(), &GetUnbanRequestsParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		Status:        "pending",
		UserID:        twitchBanUserID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 request, got %d", len(resp.Data))
	}
}

// Suspicious user status tests

func TestClient_AddSuspiciousUserStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/suspicious_users" {
			t.Errorf("expected /moderation/suspicious_users, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		moderatorID := r.URL.Query().Get("moderator_id")

		if broadcasterID != twitchBanBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchBanBroadcasterID, broadcasterID)
		}
		if moderatorID != twitchBanModeratorID {
			t.Errorf("expected moderator_id=%s, got %s", twitchBanModeratorID, moderatorID)
		}

		var body AddSuspiciousUserStatusParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.UserID != twitchBanUserID {
			t.Errorf("expected user_id=%s, got %s", twitchBanUserID, body.UserID)
		}
		if body.Status != SuspiciousUserStatusRestricted {
			t.Errorf("expected status=restricted, got %s", body.Status)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.AddSuspiciousUserStatus(context.Background(), &AddSuspiciousUserStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		UserID:        twitchBanUserID,
		Status:        SuspiciousUserStatusRestricted,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AddSuspiciousUserStatus_Monitored(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var body AddSuspiciousUserStatusParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Status != SuspiciousUserStatusMonitored {
			t.Errorf("expected status=monitored, got %s", body.Status)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.AddSuspiciousUserStatus(context.Background(), &AddSuspiciousUserStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		UserID:        twitchBanUserID,
		Status:        SuspiciousUserStatusMonitored,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AddSuspiciousUserStatus_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	err := client.AddSuspiciousUserStatus(context.Background(), &AddSuspiciousUserStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		UserID:        twitchBanUserID,
		Status:        SuspiciousUserStatusRestricted,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_RemoveSuspiciousUserStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/moderation/suspicious_users" {
			t.Errorf("expected /moderation/suspicious_users, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		moderatorID := r.URL.Query().Get("moderator_id")
		userID := r.URL.Query().Get("user_id")

		if broadcasterID != twitchBanBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchBanBroadcasterID, broadcasterID)
		}
		if moderatorID != twitchBanModeratorID {
			t.Errorf("expected moderator_id=%s, got %s", twitchBanModeratorID, moderatorID)
		}
		if userID != twitchBanUserID {
			t.Errorf("expected user_id=%s, got %s", twitchBanUserID, userID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.RemoveSuspiciousUserStatus(context.Background(), &RemoveSuspiciousUserStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		UserID:        twitchBanUserID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_RemoveSuspiciousUserStatus_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	err := client.RemoveSuspiciousUserStatus(context.Background(), &RemoveSuspiciousUserStatusParams{
		BroadcasterID: twitchBanBroadcasterID,
		ModeratorID:   twitchBanModeratorID,
		UserID:        twitchBanUserID,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Helper function to parse time strings
func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try with milliseconds
		t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
		if err != nil {
			panic("failed to parse time: " + s)
		}
	}
	return t
}
