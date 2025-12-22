package helix

import (
	"testing"
	"time"
)

func TestParseIRCMessage(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected *IRCMessage
	}{
		{
			name: "simple command",
			raw:  "PING :tmi.twitch.tv",
			expected: &IRCMessage{
				Raw:      "PING :tmi.twitch.tv",
				Tags:     map[string]string{},
				Command:  "PING",
				Trailing: "tmi.twitch.tv",
			},
		},
		{
			name: "privmsg with tags",
			raw:  "@badge-info=;badges=broadcaster/1;color=#FF0000;display-name=TestUser;emotes=;id=abc123;mod=0;room-id=12345;subscriber=0;tmi-sent-ts=1234567890123;turbo=0;user-id=12345;user-type= :testuser!testuser@testuser.tmi.twitch.tv PRIVMSG #testchannel :Hello World",
			expected: &IRCMessage{
				Raw: "@badge-info=;badges=broadcaster/1;color=#FF0000;display-name=TestUser;emotes=;id=abc123;mod=0;room-id=12345;subscriber=0;tmi-sent-ts=1234567890123;turbo=0;user-id=12345;user-type= :testuser!testuser@testuser.tmi.twitch.tv PRIVMSG #testchannel :Hello World",
				Tags: map[string]string{
					"badge-info":   "",
					"badges":       "broadcaster/1",
					"color":        "#FF0000",
					"display-name": "TestUser",
					"emotes":       "",
					"id":           "abc123",
					"mod":          "0",
					"room-id":      "12345",
					"subscriber":   "0",
					"tmi-sent-ts":  "1234567890123",
					"turbo":        "0",
					"user-id":      "12345",
					"user-type":    "",
				},
				Prefix:   "testuser!testuser@testuser.tmi.twitch.tv",
				Command:  "PRIVMSG",
				Params:   []string{"#testchannel"},
				Trailing: "Hello World",
			},
		},
		{
			name: "join message",
			raw:  ":testuser!testuser@testuser.tmi.twitch.tv JOIN #testchannel",
			expected: &IRCMessage{
				Raw:     ":testuser!testuser@testuser.tmi.twitch.tv JOIN #testchannel",
				Tags:    map[string]string{},
				Prefix:  "testuser!testuser@testuser.tmi.twitch.tv",
				Command: "JOIN",
				Params:  []string{"#testchannel"},
			},
		},
		{
			name: "cap ack",
			raw:  ":tmi.twitch.tv CAP * ACK :twitch.tv/tags twitch.tv/commands",
			expected: &IRCMessage{
				Raw:      ":tmi.twitch.tv CAP * ACK :twitch.tv/tags twitch.tv/commands",
				Tags:     map[string]string{},
				Prefix:   "tmi.twitch.tv",
				Command:  "CAP",
				Params:   []string{"*", "ACK"},
				Trailing: "twitch.tv/tags twitch.tv/commands",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIRCMessage(tt.raw)

			if result.Raw != tt.expected.Raw {
				t.Errorf("Raw mismatch: got %q, want %q", result.Raw, tt.expected.Raw)
			}

			if result.Command != tt.expected.Command {
				t.Errorf("Command mismatch: got %q, want %q", result.Command, tt.expected.Command)
			}

			if result.Prefix != tt.expected.Prefix {
				t.Errorf("Prefix mismatch: got %q, want %q", result.Prefix, tt.expected.Prefix)
			}

			if result.Trailing != tt.expected.Trailing {
				t.Errorf("Trailing mismatch: got %q, want %q", result.Trailing, tt.expected.Trailing)
			}

			if len(result.Params) != len(tt.expected.Params) {
				t.Errorf("Params length mismatch: got %d, want %d", len(result.Params), len(tt.expected.Params))
			} else {
				for i, p := range result.Params {
					if p != tt.expected.Params[i] {
						t.Errorf("Params[%d] mismatch: got %q, want %q", i, p, tt.expected.Params[i])
					}
				}
			}

			for k, v := range tt.expected.Tags {
				if result.Tags[k] != v {
					t.Errorf("Tag %q mismatch: got %q, want %q", k, result.Tags[k], v)
				}
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "basic tags",
			input: "color=#FF0000;display-name=TestUser",
			expected: map[string]string{
				"color":        "#FF0000",
				"display-name": "TestUser",
			},
		},
		{
			name:  "empty value",
			input: "badge-info=;badges=broadcaster/1",
			expected: map[string]string{
				"badge-info": "",
				"badges":     "broadcaster/1",
			},
		},
		{
			name:  "escaped values",
			input: `msg=hello\sworld\:\)`,
			expected: map[string]string{
				"msg": "hello world;)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Tag %q: got %q, want %q", k, result[k], v)
				}
			}
		})
	}
}

func TestParseEmotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []IRCEmote
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name:  "single emote",
			input: "25:0-4",
			expected: []IRCEmote{
				{ID: "25", Start: 0, End: 4, Count: 1},
			},
		},
		{
			name:  "multiple positions",
			input: "25:0-4,6-10",
			expected: []IRCEmote{
				{ID: "25", Start: 0, End: 4, Count: 1},
				{ID: "25", Start: 6, End: 10, Count: 1},
			},
		},
		{
			name:  "multiple emotes",
			input: "25:0-4/1902:6-10",
			expected: []IRCEmote{
				{ID: "25", Start: 0, End: 4, Count: 1},
				{ID: "1902", Start: 6, End: 10, Count: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmotes(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("Length mismatch: got %d, want %d", len(result), len(tt.expected))
			}

			for i, e := range result {
				if e.ID != tt.expected[i].ID {
					t.Errorf("Emote[%d].ID: got %q, want %q", i, e.ID, tt.expected[i].ID)
				}
				if e.Start != tt.expected[i].Start {
					t.Errorf("Emote[%d].Start: got %d, want %d", i, e.Start, tt.expected[i].Start)
				}
				if e.End != tt.expected[i].End {
					t.Errorf("Emote[%d].End: got %d, want %d", i, e.End, tt.expected[i].End)
				}
			}
		})
	}
}

func TestParseBadges(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "single badge",
			input: "broadcaster/1",
			expected: map[string]string{
				"broadcaster": "1",
			},
		},
		{
			name:  "multiple badges",
			input: "broadcaster/1,subscriber/12,premium/1",
			expected: map[string]string{
				"broadcaster": "1",
				"subscriber":  "12",
				"premium":     "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBadges(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("Length mismatch: got %d, want %d", len(result), len(tt.expected))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Badge %q: got %q, want %q", k, result[k], v)
				}
			}
		})
	}
}

func TestParseChatMessage(t *testing.T) {
	raw := "@badge-info=subscriber/12;badges=subscriber/12,premium/1;color=#FF0000;display-name=TestUser;emotes=25:0-4;first-msg=0;id=abc123-def456;mod=1;room-id=12345;subscriber=1;tmi-sent-ts=1234567890123;turbo=0;user-id=67890;user-type=mod :testuser!testuser@testuser.tmi.twitch.tv PRIVMSG #testchannel :Kappa Hello!"

	msg := parseIRCMessage(raw)
	chat := parseChatMessage(msg)

	if chat.ID != "abc123-def456" {
		t.Errorf("ID: got %q, want %q", chat.ID, "abc123-def456")
	}

	if chat.Channel != "testchannel" {
		t.Errorf("Channel: got %q, want %q", chat.Channel, "testchannel")
	}

	if chat.UserID != "67890" {
		t.Errorf("UserID: got %q, want %q", chat.UserID, "67890")
	}

	if chat.DisplayName != "TestUser" {
		t.Errorf("DisplayName: got %q, want %q", chat.DisplayName, "TestUser")
	}

	if chat.Message != "Kappa Hello!" {
		t.Errorf("Message: got %q, want %q", chat.Message, "Kappa Hello!")
	}

	if chat.Color != "#FF0000" {
		t.Errorf("Color: got %q, want %q", chat.Color, "#FF0000")
	}

	if !chat.IsMod {
		t.Error("IsMod should be true")
	}

	if !chat.IsSubscriber {
		t.Error("IsSubscriber should be true")
	}

	if len(chat.Emotes) != 1 {
		t.Fatalf("Expected 1 emote, got %d", len(chat.Emotes))
	}

	if chat.Emotes[0].ID != "25" {
		t.Errorf("Emote ID: got %q, want %q", chat.Emotes[0].ID, "25")
	}

	if chat.Badges["subscriber"] != "12" {
		t.Errorf("Subscriber badge: got %q, want %q", chat.Badges["subscriber"], "12")
	}
}

func TestParseUserNotice(t *testing.T) {
	raw := "@badge-info=subscriber/1;badges=subscriber/0;color=#FF0000;display-name=TestGifter;emotes=;id=abc123;login=testgifter;mod=0;msg-id=subgift;msg-param-gift-months=1;msg-param-months=1;msg-param-recipient-display-name=TestRecipient;msg-param-recipient-id=12345;msg-param-recipient-user-name=testrecipient;msg-param-sub-plan-name=Channel\\sSubscription;msg-param-sub-plan=1000;room-id=67890;subscriber=1;system-msg=TestGifter\\sgifted\\sa\\sTier\\s1\\ssub\\sto\\sTestRecipient!;tmi-sent-ts=1234567890123;user-id=11111;user-type= :tmi.twitch.tv USERNOTICE #testchannel"

	msg := parseIRCMessage(raw)
	notice := parseUserNotice(msg)

	if notice.Type != "subgift" {
		t.Errorf("Type: got %q, want %q", notice.Type, "subgift")
	}

	if notice.Channel != "testchannel" {
		t.Errorf("Channel: got %q, want %q", notice.Channel, "testchannel")
	}

	if notice.User != "testgifter" {
		t.Errorf("User: got %q, want %q", notice.User, "testgifter")
	}

	if notice.MsgParams["sub-plan"] != "1000" {
		t.Errorf("sub-plan: got %q, want %q", notice.MsgParams["sub-plan"], "1000")
	}

	if notice.MsgParams["recipient-user-name"] != "testrecipient" {
		t.Errorf("recipient-user-name: got %q, want %q", notice.MsgParams["recipient-user-name"], "testrecipient")
	}
}

func TestParseRoomState(t *testing.T) {
	raw := "@emote-only=0;followers-only=10;r9k=0;room-id=12345;slow=30;subs-only=1 :tmi.twitch.tv ROOMSTATE #testchannel"

	msg := parseIRCMessage(raw)
	state := parseRoomState(msg)

	if state.Channel != "testchannel" {
		t.Errorf("Channel: got %q, want %q", state.Channel, "testchannel")
	}

	if state.EmoteOnly {
		t.Error("EmoteOnly should be false")
	}

	if state.FollowersOnly != 10 {
		t.Errorf("FollowersOnly: got %d, want %d", state.FollowersOnly, 10)
	}

	if state.R9K {
		t.Error("R9K should be false")
	}

	if state.Slow != 30 {
		t.Errorf("Slow: got %d, want %d", state.Slow, 30)
	}

	if !state.SubsOnly {
		t.Error("SubsOnly should be true")
	}
}

func TestParseClearChat(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		expectedUser  string
		expectedDuration int
	}{
		{
			name:          "timeout",
			raw:           "@ban-duration=600;room-id=12345;target-user-id=67890;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARCHAT #testchannel :baduser",
			expectedUser:  "baduser",
			expectedDuration: 600,
		},
		{
			name:          "ban",
			raw:           "@room-id=12345;target-user-id=67890;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARCHAT #testchannel :baduser",
			expectedUser:  "baduser",
			expectedDuration: 0,
		},
		{
			name:          "clear chat",
			raw:           "@room-id=12345;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARCHAT #testchannel",
			expectedUser:  "",
			expectedDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := parseIRCMessage(tt.raw)
			clear := parseClearChat(msg)

			if clear.User != tt.expectedUser {
				t.Errorf("User: got %q, want %q", clear.User, tt.expectedUser)
			}

			if clear.BanDuration != tt.expectedDuration {
				t.Errorf("BanDuration: got %d, want %d", clear.BanDuration, tt.expectedDuration)
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	ts := parseTimestamp("1234567890123")
	expected := time.UnixMilli(1234567890123)

	if !ts.Equal(expected) {
		t.Errorf("Timestamp: got %v, want %v", ts, expected)
	}
}

func TestParseUserFromPrefix(t *testing.T) {
	tests := []struct {
		prefix   string
		expected string
	}{
		{"testuser!testuser@testuser.tmi.twitch.tv", "testuser"},
		{"testuser", "testuser"},
		{"", ""},
	}

	for _, tt := range tests {
		result := parseUserFromPrefix(tt.prefix)
		if result != tt.expected {
			t.Errorf("parseUserFromPrefix(%q): got %q, want %q", tt.prefix, result, tt.expected)
		}
	}
}

func TestNewIRCClient(t *testing.T) {
	client := NewIRCClient("testuser", "token123")

	if client.nick != "testuser" {
		t.Errorf("nick: got %q, want %q", client.nick, "testuser")
	}

	if client.token != "oauth:token123" {
		t.Errorf("token: got %q, want %q", client.token, "oauth:token123")
	}

	if client.url != TwitchIRCWebSocket {
		t.Errorf("url: got %q, want %q", client.url, TwitchIRCWebSocket)
	}

	// Test with oauth: prefix already present
	client2 := NewIRCClient("testuser", "oauth:token456")
	if client2.token != "oauth:token456" {
		t.Errorf("token with prefix: got %q, want %q", client2.token, "oauth:token456")
	}
}

func TestIRCClientOptions(t *testing.T) {
	messageReceived := false
	errorReceived := false

	client := NewIRCClient("testuser", "token",
		WithIRCURL("wss://custom.url"),
		WithAutoReconnect(false),
		WithReconnectDelay(10*time.Second),
		WithMessageHandler(func(m *ChatMessage) {
			messageReceived = true
		}),
		WithIRCErrorHandler(func(err error) {
			errorReceived = true
		}),
	)

	if client.url != "wss://custom.url" {
		t.Errorf("url: got %q, want %q", client.url, "wss://custom.url")
	}

	if client.autoReconnect {
		t.Error("autoReconnect should be false")
	}

	if client.reconnectDelay != 10*time.Second {
		t.Errorf("reconnectDelay: got %v, want %v", client.reconnectDelay, 10*time.Second)
	}

	// Test that handlers are set
	if client.onMessage == nil {
		t.Error("onMessage handler should be set")
	}

	if client.onError == nil {
		t.Error("onError handler should be set")
	}

	// Call handlers to verify they work
	client.onMessage(&ChatMessage{})
	if !messageReceived {
		t.Error("message handler was not called")
	}

	client.onError(nil)
	if !errorReceived {
		t.Error("error handler was not called")
	}
}

func TestChannelManagement(t *testing.T) {
	client := NewIRCClient("testuser", "token")

	// Join channels while disconnected (should be queued)
	err := client.Join("channel1", "#channel2", "CHANNEL3")
	if err != nil {
		t.Errorf("Join error: %v", err)
	}

	channels := client.GetJoinedChannels()
	if len(channels) != 3 {
		t.Errorf("Expected 3 channels, got %d", len(channels))
	}

	// Verify channel names are normalized
	channelMap := make(map[string]bool)
	for _, ch := range channels {
		channelMap[ch] = true
	}

	if !channelMap["channel1"] {
		t.Error("channel1 should be in joined channels")
	}
	if !channelMap["channel2"] {
		t.Error("channel2 should be in joined channels")
	}
	if !channelMap["channel3"] {
		t.Error("channel3 should be in joined channels")
	}

	// Part a channel
	err = client.Part("channel1")
	if err != nil {
		t.Errorf("Part error: %v", err)
	}

	channels = client.GetJoinedChannels()
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels after part, got %d", len(channels))
	}
}
