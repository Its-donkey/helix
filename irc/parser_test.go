package irc

import (
	"testing"
	"time"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected *Message
	}{
		{
			name: "simple command",
			raw:  "PING :tmi.twitch.tv",
			expected: &Message{
				Raw:      "PING :tmi.twitch.tv",
				Tags:     map[string]string{},
				Command:  "PING",
				Trailing: "tmi.twitch.tv",
			},
		},
		{
			name: "privmsg with tags",
			raw:  "@badge-info=;badges=broadcaster/1;color=#FF0000;display-name=TestUser;emotes=;id=abc123;mod=0;room-id=12345;subscriber=0;tmi-sent-ts=1234567890123;turbo=0;user-id=12345;user-type= :testuser!testuser@testuser.tmi.twitch.tv PRIVMSG #testchannel :Hello World",
			expected: &Message{
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
			expected: &Message{
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
			expected: &Message{
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
			result := parseMessage(tt.raw)

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
		expected []Emote
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name:  "single emote",
			input: "25:0-4",
			expected: []Emote{
				{ID: "25", Start: 0, End: 4, Count: 1},
			},
		},
		{
			name:  "multiple positions",
			input: "25:0-4,6-10",
			expected: []Emote{
				{ID: "25", Start: 0, End: 4, Count: 1},
				{ID: "25", Start: 6, End: 10, Count: 1},
			},
		},
		{
			name:  "multiple emotes",
			input: "25:0-4/1902:6-10",
			expected: []Emote{
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

	msg := parseMessage(raw)
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

	msg := parseMessage(raw)
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

	msg := parseMessage(raw)
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
		name             string
		raw              string
		expectedUser     string
		expectedDuration int
	}{
		{
			name:             "timeout",
			raw:              "@ban-duration=600;room-id=12345;target-user-id=67890;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARCHAT #testchannel :baduser",
			expectedUser:     "baduser",
			expectedDuration: 600,
		},
		{
			name:             "ban",
			raw:              "@room-id=12345;target-user-id=67890;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARCHAT #testchannel :baduser",
			expectedUser:     "baduser",
			expectedDuration: 0,
		},
		{
			name:             "clear chat",
			raw:              "@room-id=12345;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARCHAT #testchannel",
			expectedUser:     "",
			expectedDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := parseMessage(tt.raw)
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

func TestParseNotice(t *testing.T) {
	raw := "@msg-id=subs_on :tmi.twitch.tv NOTICE #testchannel :This room is now in subscribers-only mode."

	msg := parseMessage(raw)
	notice := parseNotice(msg)

	if notice.Channel != "testchannel" {
		t.Errorf("Channel: got %q, want %q", notice.Channel, "testchannel")
	}

	if notice.MsgID != "subs_on" {
		t.Errorf("MsgID: got %q, want %q", notice.MsgID, "subs_on")
	}

	if notice.Message != "This room is now in subscribers-only mode." {
		t.Errorf("Message: got %q, want %q", notice.Message, "This room is now in subscribers-only mode.")
	}
}

func TestParseClearMessage(t *testing.T) {
	raw := "@login=testuser;room-id=;target-msg-id=abc123-def456;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARMSG #testchannel :This is the deleted message"

	msg := parseMessage(raw)
	clear := parseClearMessage(msg)

	if clear.Channel != "testchannel" {
		t.Errorf("Channel: got %q, want %q", clear.Channel, "testchannel")
	}

	if clear.User != "testuser" {
		t.Errorf("User: got %q, want %q", clear.User, "testuser")
	}

	if clear.TargetMsgID != "abc123-def456" {
		t.Errorf("TargetMsgID: got %q, want %q", clear.TargetMsgID, "abc123-def456")
	}

	if clear.Message != "This is the deleted message" {
		t.Errorf("Message: got %q, want %q", clear.Message, "This is the deleted message")
	}
}

func TestParseWhisper(t *testing.T) {
	raw := "@badges=;color=#FF0000;display-name=TestSender;emotes=;message-id=1;thread-id=12345_67890;user-id=12345 :testsender!testsender@testsender.tmi.twitch.tv WHISPER testrecipient :Hello there!"

	msg := parseMessage(raw)
	whisper := parseWhisper(msg)

	if whisper.From != "testsender" {
		t.Errorf("From: got %q, want %q", whisper.From, "testsender")
	}

	if whisper.FromID != "12345" {
		t.Errorf("FromID: got %q, want %q", whisper.FromID, "12345")
	}

	if whisper.To != "testrecipient" {
		t.Errorf("To: got %q, want %q", whisper.To, "testrecipient")
	}

	if whisper.Message != "Hello there!" {
		t.Errorf("Message: got %q, want %q", whisper.Message, "Hello there!")
	}

	if whisper.DisplayName != "TestSender" {
		t.Errorf("DisplayName: got %q, want %q", whisper.DisplayName, "TestSender")
	}

	if whisper.Color != "#FF0000" {
		t.Errorf("Color: got %q, want %q", whisper.Color, "#FF0000")
	}

	if whisper.MessageID != "1" {
		t.Errorf("MessageID: got %q, want %q", whisper.MessageID, "1")
	}

	if whisper.ThreadID != "12345_67890" {
		t.Errorf("ThreadID: got %q, want %q", whisper.ThreadID, "12345_67890")
	}
}

func TestParseGlobalUserState(t *testing.T) {
	raw := "@badge-info=subscriber/12;badges=subscriber/12;color=#FF0000;display-name=TestUser;emote-sets=0,1,2,300;user-id=12345 :tmi.twitch.tv GLOBALUSERSTATE"

	msg := parseMessage(raw)
	state := parseGlobalUserState(msg)

	if state.UserID != "12345" {
		t.Errorf("UserID: got %q, want %q", state.UserID, "12345")
	}

	if state.DisplayName != "TestUser" {
		t.Errorf("DisplayName: got %q, want %q", state.DisplayName, "TestUser")
	}

	if state.Color != "#FF0000" {
		t.Errorf("Color: got %q, want %q", state.Color, "#FF0000")
	}

	if len(state.EmoteSets) != 4 {
		t.Errorf("EmoteSets length: got %d, want %d", len(state.EmoteSets), 4)
	}

	if state.Badges["subscriber"] != "12" {
		t.Errorf("Badges subscriber: got %q, want %q", state.Badges["subscriber"], "12")
	}
}

func TestParseUserState(t *testing.T) {
	raw := "@badge-info=subscriber/6;badges=subscriber/6,premium/1;color=#00FF00;display-name=TestUser;emote-sets=0,1,2;mod=1;subscriber=1 :tmi.twitch.tv USERSTATE #testchannel"

	msg := parseMessage(raw)
	state := parseUserState(msg)

	if state.Channel != "testchannel" {
		t.Errorf("Channel: got %q, want %q", state.Channel, "testchannel")
	}

	if state.DisplayName != "TestUser" {
		t.Errorf("DisplayName: got %q, want %q", state.DisplayName, "TestUser")
	}

	if state.Color != "#00FF00" {
		t.Errorf("Color: got %q, want %q", state.Color, "#00FF00")
	}

	if !state.IsMod {
		t.Error("IsMod should be true")
	}

	if !state.IsSubscriber {
		t.Error("IsSubscriber should be true")
	}

	if len(state.EmoteSets) != 3 {
		t.Errorf("EmoteSets length: got %d, want %d", len(state.EmoteSets), 3)
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

func TestParseMessageEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"empty", ""},
		{"tags only no space", "@tag=value"},
		{"prefix only", ":prefix"},
		{"command only", "PING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := parseMessage(tt.raw)
			if result == nil {
				t.Error("parseMessage should never return nil")
			}
		})
	}
}

func TestParseTagsEdgeCases(t *testing.T) {
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
			name:  "key only no value",
			input: "key",
			expected: map[string]string{
				"key": "",
			},
		},
		{
			name:  "empty pairs",
			input: ";key=value;;",
			expected: map[string]string{
				"key": "value",
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

func TestUnescapeTagValueEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`\r`, "\r"},
		{`\n`, "\n"},
		{`\\`, "\\"},
		{`\x`, "x"},         // unknown escape
		{`test\`, "test\\"}, // trailing backslash
	}

	for _, tt := range tests {
		result := unescapeTagValue(tt.input)
		if result != tt.expected {
			t.Errorf("unescapeTagValue(%q): got %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseEmotesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty parts", "//", 0},
		{"no colon", "25", 0},
		{"no dash", "25:04", 0},
		{"invalid start", "25:abc-4", 0},
		{"invalid end", "25:0-def", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmotes(tt.input)
			if len(result) != tt.expected {
				t.Errorf("Expected %d emotes, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestParseBadgesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "no slash",
			input: "badge",
			expected: map[string]string{
				"badge": "",
			},
		},
		{
			name:     "empty parts",
			input:    ",,",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBadges(tt.input)
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Badge %q: got %q, want %q", k, result[k], v)
				}
			}
		})
	}
}

func TestParseTimestampEdgeCases(t *testing.T) {
	// Empty timestamp should return current time (approximately)
	before := time.Now()
	result := parseTimestamp("")
	after := time.Now()

	if result.Before(before) || result.After(after) {
		t.Error("Empty timestamp should return current time")
	}

	// Invalid timestamp should return current time
	before = time.Now()
	result = parseTimestamp("invalid")
	after = time.Now()

	if result.Before(before) || result.After(after) {
		t.Error("Invalid timestamp should return current time")
	}
}

func TestParseIntEdgeCases(t *testing.T) {
	if parseInt("") != 0 {
		t.Error("Empty string should return 0")
	}

	if parseInt("invalid") != 0 {
		t.Error("Invalid int should return 0")
	}

	if parseInt("42") != 42 {
		t.Error("'42' should return 42")
	}
}

func TestParseMessageMoreEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"message with trailing only", "PRIVMSG :message"},
		{"message with multiple colons", ":prefix PRIVMSG #channel :hello:world:test"},
		{"message with empty trailing", ":prefix PRIVMSG #channel :"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMessage(tt.raw)
			if result == nil {
				t.Error("parseMessage should never return nil")
			}
		})
	}
}
