package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetChatters(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/chatters" {
			t.Errorf("expected /chat/chatters, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		moderatorID := r.URL.Query().Get("moderator_id")

		if broadcasterID != "12345" || moderatorID != "67890" {
			t.Errorf("expected broadcaster_id=12345, moderator_id=67890")
		}

		resp := Response[Chatter]{
			Data: []Chatter{
				{UserID: "11111", UserLogin: "chatter1", UserName: "Chatter1"},
				{UserID: "22222", UserLogin: "chatter2", UserName: "Chatter2"},
				{UserID: "33333", UserLogin: "chatter3", UserName: "Chatter3"},
			},
		}
		total := 3
		resp.Total = &total
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChatters(context.Background(), &GetChattersParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 chatters, got %d", len(resp.Data))
	}
}

func TestClient_GetChannelEmotes(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/emotes" {
			t.Errorf("expected /chat/emotes, got %s", r.URL.Path)
		}

		resp := Response[Emote]{
			Data: []Emote{
				{
					ID:        "emote123",
					Name:      "TestEmote",
					Tier:      "1000",
					EmoteType: "subscriptions",
					Format:    []string{"static", "animated"},
					Scale:     []string{"1.0", "2.0", "3.0"},
					ThemeMode: []string{"light", "dark"},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelEmotes(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 emote, got %d", len(resp.Data))
	}
	if resp.Data[0].Tier != "1000" {
		t.Errorf("expected tier 1000, got %s", resp.Data[0].Tier)
	}
}

func TestClient_GetGlobalEmotes(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/emotes/global" {
			t.Errorf("expected /chat/emotes/global, got %s", r.URL.Path)
		}

		resp := Response[Emote]{
			Data: []Emote{
				{ID: "1", Name: "Kappa"},
				{ID: "2", Name: "PogChamp"},
				{ID: "3", Name: "LUL"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGlobalEmotes(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 emotes, got %d", len(resp.Data))
	}
}

func TestClient_GetEmoteSets(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/emotes/set" {
			t.Errorf("expected /chat/emotes/set, got %s", r.URL.Path)
		}

		emoteSetIDs := r.URL.Query()["emote_set_id"]
		if len(emoteSetIDs) != 2 {
			t.Errorf("expected 2 emote_set_ids, got %d", len(emoteSetIDs))
		}

		resp := Response[Emote]{
			Data: []Emote{
				{ID: "emote1", Name: "Emote1", EmoteSetID: "set1"},
				{ID: "emote2", Name: "Emote2", EmoteSetID: "set2"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetEmoteSets(context.Background(), []string{"set1", "set2"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 emotes, got %d", len(resp.Data))
	}
}

func TestClient_GetChannelChatBadges(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/badges" {
			t.Errorf("expected /chat/badges, got %s", r.URL.Path)
		}

		resp := Response[ChatBadge]{
			Data: []ChatBadge{
				{
					SetID: "subscriber",
					Versions: []BadgeVersion{
						{ID: "0", Title: "Subscriber", ImageURL1x: "https://example.com/sub0.png"},
						{ID: "3", Title: "3-Month Subscriber", ImageURL1x: "https://example.com/sub3.png"},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelChatBadges(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 badge set, got %d", len(resp.Data))
	}
	if len(resp.Data[0].Versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(resp.Data[0].Versions))
	}
}

func TestClient_GetGlobalChatBadges(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/badges/global" {
			t.Errorf("expected /chat/badges/global, got %s", r.URL.Path)
		}

		resp := Response[ChatBadge]{
			Data: []ChatBadge{
				{SetID: "moderator", Versions: []BadgeVersion{{ID: "1"}}},
				{SetID: "vip", Versions: []BadgeVersion{{ID: "1"}}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetGlobalChatBadges(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 badge sets, got %d", len(resp.Data))
	}
}

func TestClient_GetChatSettings(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/settings" {
			t.Errorf("expected /chat/settings, got %s", r.URL.Path)
		}

		resp := Response[ChatSettings]{
			Data: []ChatSettings{
				{
					BroadcasterID:         "12345",
					SlowMode:              true,
					SlowModeWaitTime:      30,
					FollowerMode:          true,
					FollowerModeDuration:  10,
					SubscriberMode:        false,
					EmoteMode:             false,
					UniqueChatMode:        false,
					NonModeratorChatDelay: true,
					NonModeratorChatDelayDuration: 2,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChatSettings(context.Background(), "12345", "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 setting, got %d", len(resp.Data))
	}
	if !resp.Data[0].SlowMode {
		t.Error("expected slow mode to be enabled")
	}
	if resp.Data[0].SlowModeWaitTime != 30 {
		t.Errorf("expected slow mode wait time 30, got %d", resp.Data[0].SlowModeWaitTime)
	}
}

func TestClient_UpdateChatSettings(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/chat/settings" {
			t.Errorf("expected /chat/settings, got %s", r.URL.Path)
		}

		var body UpdateChatSettingsParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		resp := Response[ChatSettings]{
			Data: []ChatSettings{
				{BroadcasterID: "12345", SlowMode: true, SlowModeWaitTime: 60},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	slowMode := true
	waitTime := 60
	resp, err := client.UpdateChatSettings(context.Background(), &UpdateChatSettingsParams{
		BroadcasterID:    "12345",
		ModeratorID:      "67890",
		SlowMode:         &slowMode,
		SlowModeWaitTime: &waitTime,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 setting, got %d", len(resp.Data))
	}
}

func TestClient_SendChatAnnouncement(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/announcements" {
			t.Errorf("expected /chat/announcements, got %s", r.URL.Path)
		}

		var body SendChatAnnouncementParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Message != "Important announcement!" {
			t.Errorf("expected message 'Important announcement!', got %s", body.Message)
		}
		if body.Color != "blue" {
			t.Errorf("expected color blue, got %s", body.Color)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendChatAnnouncement(context.Background(), &SendChatAnnouncementParams{
		BroadcasterID: "12345",
		ModeratorID:   "67890",
		Message:       "Important announcement!",
		Color:         "blue",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendShoutout(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/shoutouts" {
			t.Errorf("expected /chat/shoutouts, got %s", r.URL.Path)
		}

		fromID := r.URL.Query().Get("from_broadcaster_id")
		toID := r.URL.Query().Get("to_broadcaster_id")
		modID := r.URL.Query().Get("moderator_id")

		if fromID != "12345" || toID != "67890" || modID != "11111" {
			t.Errorf("unexpected query params: from=%s, to=%s, mod=%s", fromID, toID, modID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendShoutout(context.Background(), &SendShoutoutParams{
		FromBroadcasterID: "12345",
		ToBroadcasterID:   "67890",
		ModeratorID:       "11111",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetUserChatColor(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/color" {
			t.Errorf("expected /chat/color, got %s", r.URL.Path)
		}

		userIDs := r.URL.Query()["user_id"]
		if len(userIDs) != 2 {
			t.Errorf("expected 2 user_ids, got %d", len(userIDs))
		}

		resp := Response[UserChatColor]{
			Data: []UserChatColor{
				{UserID: "12345", UserLogin: "user1", UserName: "User1", Color: "#FF0000"},
				{UserID: "67890", UserLogin: "user2", UserName: "User2", Color: "#00FF00"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUserChatColor(context.Background(), []string{"12345", "67890"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(resp.Data))
	}
	if resp.Data[0].Color != "#FF0000" {
		t.Errorf("expected color #FF0000, got %s", resp.Data[0].Color)
	}
}

func TestClient_UpdateUserChatColor(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/chat/color" {
			t.Errorf("expected /chat/color, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		color := r.URL.Query().Get("color")

		if userID != "12345" || color != "blue" {
			t.Errorf("expected user_id=12345, color=blue, got %s, %s", userID, color)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.UpdateUserChatColor(context.Background(), "12345", "blue")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendChatMessage(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/messages" {
			t.Errorf("expected /chat/messages, got %s", r.URL.Path)
		}

		var body SendChatMessageParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Message != "Hello, chat!" {
			t.Errorf("expected message 'Hello, chat!', got %s", body.Message)
		}
		if body.BroadcasterID != "12345" {
			t.Errorf("expected broadcaster_id 12345, got %s", body.BroadcasterID)
		}

		resp := Response[SendChatMessageResponse]{
			Data: []SendChatMessageResponse{
				{MessageID: "msg123", IsSent: true},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SendChatMessage(context.Background(), &SendChatMessageParams{
		BroadcasterID: "12345",
		SenderID:      "67890",
		Message:       "Hello, chat!",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSent {
		t.Error("expected message to be sent")
	}
	if result.MessageID != "msg123" {
		t.Errorf("expected message ID msg123, got %s", result.MessageID)
	}
}

func TestClient_SendChatMessage_WithReply(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var body SendChatMessageParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.ReplyParentMessageID != "parent123" {
			t.Errorf("expected reply_parent_message_id parent123, got %s", body.ReplyParentMessageID)
		}

		resp := Response[SendChatMessageResponse]{
			Data: []SendChatMessageResponse{
				{MessageID: "reply456", IsSent: true},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SendChatMessage(context.Background(), &SendChatMessageParams{
		BroadcasterID:        "12345",
		SenderID:             "67890",
		Message:              "Reply message",
		ReplyParentMessageID: "parent123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MessageID != "reply456" {
		t.Errorf("expected message ID reply456, got %s", result.MessageID)
	}
}

func TestClient_SendChatMessage_Dropped(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[SendChatMessageResponse]{
			Data: []SendChatMessageResponse{
				{
					MessageID: "",
					IsSent:    false,
					DropReason: &struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					}{
						Code:    "channel_settings",
						Message: "Channel has subscriber-only mode enabled",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.SendChatMessage(context.Background(), &SendChatMessageParams{
		BroadcasterID: "12345",
		SenderID:      "67890",
		Message:       "This will be dropped",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsSent {
		t.Error("expected message to not be sent")
	}
	if result.DropReason == nil {
		t.Fatal("expected drop reason to not be nil")
	}
	if result.DropReason.Code != "channel_settings" {
		t.Errorf("expected drop reason code 'channel_settings', got %s", result.DropReason.Code)
	}
}

func TestClient_GetSharedChatSession(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shared_chat/session" {
			t.Errorf("expected /shared_chat/session, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[SharedChatSession]{
			Data: []SharedChatSession{
				{
					SessionID:         "session123",
					HostBroadcasterID: "12345",
					Participants: []SharedChatParticipant{
						{BroadcasterID: "12345"},
						{BroadcasterID: "67890"},
					},
					CreatedAt: "2024-01-15T12:00:00Z",
					UpdatedAt: "2024-01-15T12:30:00Z",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetSharedChatSession(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected session, got nil")
	}
	if resp.SessionID != "session123" {
		t.Errorf("expected session_id 'session123', got %s", resp.SessionID)
	}
	if len(resp.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(resp.Participants))
	}
}

func TestClient_GetSharedChatSession_NoSession(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[SharedChatSession]{
			Data: []SharedChatSession{},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetSharedChatSession(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil session, got non-nil")
	}
}

func TestClient_GetUserEmotes(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/emotes/user" {
			t.Errorf("expected /chat/emotes/user, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := UserEmotesResponse{
			Data: []UserEmote{
				{
					ID:         "emote1",
					Name:       "TestEmote",
					EmoteType:  "subscriptions",
					EmoteSetID: "set123",
					OwnerID:    "67890",
					Format:     []string{"static", "animated"},
					Scale:      []string{"1.0", "2.0", "3.0"},
					ThemeMode:  []string{"light", "dark"},
				},
			},
			Template:   "https://static-cdn.jtvnw.net/emoticons/v2/{{id}}/{{format}}/{{theme_mode}}/{{scale}}",
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUserEmotes(context.Background(), &GetUserEmotesParams{
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 emote, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "TestEmote" {
		t.Errorf("expected emote name 'TestEmote', got %s", resp.Data[0].Name)
	}
	if resp.Template == "" {
		t.Error("expected template to be set")
	}
}

func TestClient_GetUserEmotes_WithBroadcaster(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "67890" {
			t.Errorf("expected broadcaster_id=67890, got %s", broadcasterID)
		}

		resp := UserEmotesResponse{
			Data:     []UserEmote{},
			Template: "https://static-cdn.jtvnw.net/emoticons/v2/{{id}}/{{format}}/{{theme_mode}}/{{scale}}",
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetUserEmotes(context.Background(), &GetUserEmotesParams{
		UserID:        "12345",
		BroadcasterID: "67890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
