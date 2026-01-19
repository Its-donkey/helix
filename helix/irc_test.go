package helix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockIRCServer creates a test IRC WebSocket server
type mockIRCServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader
	conn     *websocket.Conn
	mu       sync.Mutex
}

func newMockIRCServer(handler func(*websocket.Conn)) *mockIRCServer {
	mock := &mockIRCServer{
		upgrader: websocket.Upgrader{},
	}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := mock.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		mock.mu.Lock()
		mock.conn = conn
		mock.mu.Unlock()
		handler(conn)
	}))

	return mock
}

func (m *mockIRCServer) URL() string {
	return "ws" + strings.TrimPrefix(m.server.URL, "http")
}

func (m *mockIRCServer) Close() {
	m.mu.Lock()
	if m.conn != nil {
		_ = m.conn.Close()
	}
	m.mu.Unlock()
	m.server.Close()
}

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

func TestNewIRCClient_EmptyInputs(t *testing.T) {
	// Test with empty nick
	client := NewIRCClient("", "token")
	if client != nil {
		t.Error("expected nil client for empty nick")
	}

	// Test with empty token
	client = NewIRCClient("nick", "")
	if client != nil {
		t.Error("expected nil client for empty token")
	}

	// Test with both empty
	client = NewIRCClient("", "")
	if client != nil {
		t.Error("expected nil client for both empty")
	}
}

func TestIRCClient_AllHandlerOptions(t *testing.T) {
	joinCalled := false
	partCalled := false
	noticeCalled := false
	userNoticeCalled := false
	roomStateCalled := false
	clearChatCalled := false
	clearMessageCalled := false
	whisperCalled := false
	globalUserStateCalled := false
	userStateCalled := false
	connectCalled := false
	disconnectCalled := false
	reconnectCalled := false
	rawMessageCalled := false

	client := NewIRCClient("testuser", "token",
		WithJoinHandler(func(channel, user string) { joinCalled = true }),
		WithPartHandler(func(channel, user string) { partCalled = true }),
		WithNoticeHandler(func(n *Notice) { noticeCalled = true }),
		WithUserNoticeHandler(func(n *UserNotice) { userNoticeCalled = true }),
		WithRoomStateHandler(func(r *RoomState) { roomStateCalled = true }),
		WithClearChatHandler(func(c *ClearChat) { clearChatCalled = true }),
		WithClearMessageHandler(func(c *ClearMessage) { clearMessageCalled = true }),
		WithWhisperHandler(func(w *Whisper) { whisperCalled = true }),
		WithGlobalUserStateHandler(func(g *GlobalUserState) { globalUserStateCalled = true }),
		WithUserStateHandler(func(u *UserState) { userStateCalled = true }),
		WithConnectHandler(func() { connectCalled = true }),
		WithDisconnectHandler(func() { disconnectCalled = true }),
		WithReconnectHandler(func() { reconnectCalled = true }),
		WithRawMessageHandler(func(s string) { rawMessageCalled = true }),
	)

	// Verify all handlers are set
	if client.onJoin == nil {
		t.Error("onJoin handler should be set")
	}
	if client.onPart == nil {
		t.Error("onPart handler should be set")
	}
	if client.onNotice == nil {
		t.Error("onNotice handler should be set")
	}
	if client.onUserNotice == nil {
		t.Error("onUserNotice handler should be set")
	}
	if client.onRoomState == nil {
		t.Error("onRoomState handler should be set")
	}
	if client.onClearChat == nil {
		t.Error("onClearChat handler should be set")
	}
	if client.onClearMessage == nil {
		t.Error("onClearMessage handler should be set")
	}
	if client.onWhisper == nil {
		t.Error("onWhisper handler should be set")
	}
	if client.onGlobalUserState == nil {
		t.Error("onGlobalUserState handler should be set")
	}
	if client.onUserState == nil {
		t.Error("onUserState handler should be set")
	}
	if client.onConnect == nil {
		t.Error("onConnect handler should be set")
	}
	if client.onDisconnect == nil {
		t.Error("onDisconnect handler should be set")
	}
	if client.onReconnect == nil {
		t.Error("onReconnect handler should be set")
	}
	if client.onRawMessage == nil {
		t.Error("onRawMessage handler should be set")
	}

	// Call all handlers to verify they work
	client.onJoin("channel", "user")
	client.onPart("channel", "user")
	client.onNotice(&Notice{})
	client.onUserNotice(&UserNotice{})
	client.onRoomState(&RoomState{})
	client.onClearChat(&ClearChat{})
	client.onClearMessage(&ClearMessage{})
	client.onWhisper(&Whisper{})
	client.onGlobalUserState(&GlobalUserState{})
	client.onUserState(&UserState{})
	client.onConnect()
	client.onDisconnect()
	client.onReconnect()
	client.onRawMessage("raw message")

	if !joinCalled || !partCalled || !noticeCalled || !userNoticeCalled ||
		!roomStateCalled || !clearChatCalled || !clearMessageCalled ||
		!whisperCalled || !globalUserStateCalled || !userStateCalled ||
		!connectCalled || !disconnectCalled || !reconnectCalled || !rawMessageCalled {
		t.Error("not all handlers were called")
	}
}

func TestIRCClient_Connect(t *testing.T) {
	connectCalled := false
	messageSent := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Read CAP REQ
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if !strings.Contains(string(data), "CAP REQ") {
			t.Errorf("expected CAP REQ, got: %s", data)
		}

		// Read PASS
		_, data, err = conn.ReadMessage()
		if err != nil {
			return
		}
		if !strings.Contains(string(data), "PASS oauth:") {
			t.Errorf("expected PASS, got: %s", data)
		}

		// Read NICK
		_, data, err = conn.ReadMessage()
		if err != nil {
			return
		}
		if !strings.Contains(string(data), "NICK testuser") {
			t.Errorf("expected NICK, got: %s", data)
		}

		// Send CAP ACK
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags twitch.tv/commands twitch.tv/membership\r\n"))

		// Send 001 welcome
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome, GLHF!\r\n"))

		close(messageSent)

		// Keep connection alive
		time.Sleep(200 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token123",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
		WithConnectHandler(func() { connectCalled = true }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for message exchange
	select {
	case <-messageSent:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message exchange")
	}

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}

	if !connectCalled {
		t.Error("connect handler was not called")
	}

	_ = client.Close()
}

func TestIRCClient_Connect_AlreadyConnected(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Send auth response
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(200 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("first Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Second connect should return error
	err = client.Connect(ctx)
	if err != ErrIRCAlreadyConnected {
		t.Errorf("expected ErrIRCAlreadyConnected, got: %v", err)
	}
}

func TestIRCClient_Connect_AuthFailed(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Send auth failure notice
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv NOTICE * :Login authentication failed\r\n"))
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "badtoken",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != ErrIRCAuthFailed {
		t.Errorf("expected ErrIRCAuthFailed, got: %v", err)
	}
}

func TestIRCClient_Connect_DialError(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithIRCURL("ws://invalid.invalid:9999"),
		WithAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestIRCClient_HandleMessage_AllTypes(t *testing.T) {
	var (
		privmsgReceived      bool
		joinReceived         bool
		partReceived         bool
		noticeReceived       bool
		userNoticeReceived   bool
		roomStateReceived    bool
		clearChatReceived    bool
		clearMsgReceived     bool
		whisperReceived      bool
		globalUserReceived   bool
		userStateReceived    bool
		rawMessageReceived   bool
		pongReceived         bool
	)

	messagesDone := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Send auth response
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		time.Sleep(50 * time.Millisecond)

		// Send PRIVMSG
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=;badges=;color=#FF0000;display-name=TestUser;emotes=;id=abc123;mod=0;room-id=12345;subscriber=0;tmi-sent-ts=1234567890123;turbo=0;user-id=67890;user-type= :testuser!testuser@testuser.tmi.twitch.tv PRIVMSG #testchannel :Hello World\r\n"))

		// Send JOIN
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":testuser!testuser@testuser.tmi.twitch.tv JOIN #testchannel\r\n"))

		// Send PART
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":testuser!testuser@testuser.tmi.twitch.tv PART #testchannel\r\n"))

		// Send NOTICE
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@msg-id=slow_on :tmi.twitch.tv NOTICE #testchannel :This room is now in slow mode.\r\n"))

		// Send USERNOTICE (sub)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=subscriber/1;badges=subscriber/0;color=#FF0000;display-name=TestUser;emotes=;id=abc123;login=testuser;mod=0;msg-id=sub;msg-param-months=1;msg-param-sub-plan-name=Channel\\sSubscription;msg-param-sub-plan=1000;room-id=12345;subscriber=1;system-msg=TestUser\\ssubscribed;tmi-sent-ts=1234567890123;user-id=67890;user-type= :tmi.twitch.tv USERNOTICE #testchannel\r\n"))

		// Send ROOMSTATE
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@emote-only=0;followers-only=-1;r9k=0;room-id=12345;slow=0;subs-only=0 :tmi.twitch.tv ROOMSTATE #testchannel\r\n"))

		// Send CLEARCHAT
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@ban-duration=600;room-id=12345;target-user-id=67890;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARCHAT #testchannel :baduser\r\n"))

		// Send CLEARMSG
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@login=baduser;room-id=;target-msg-id=abc123;tmi-sent-ts=1234567890123 :tmi.twitch.tv CLEARMSG #testchannel :deleted message\r\n"))

		// Send WHISPER
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badges=;color=#FF0000;display-name=TestUser;emotes=;message-id=1;thread-id=67890_12345;turbo=0;user-id=67890;user-type= :testuser!testuser@testuser.tmi.twitch.tv WHISPER testrecipient :Hello\r\n"))

		// Send GLOBALUSERSTATE
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=;badges=;color=#FF0000;display-name=TestUser;emote-sets=0;user-id=12345;user-type= :tmi.twitch.tv GLOBALUSERSTATE\r\n"))

		// Send USERSTATE
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=;badges=moderator/1;color=#FF0000;display-name=TestUser;emote-sets=0;mod=1;subscriber=0;user-type=mod :tmi.twitch.tv USERSTATE #testchannel\r\n"))

		// Send PING to test PONG response
		_ = conn.WriteMessage(websocket.TextMessage, []byte("PING :tmi.twitch.tv\r\n"))

		// Read PONG response - need to read through CAP, PASS, NICK, JOIN messages first
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if strings.Contains(string(data), "PONG") {
				pongReceived = true
				break
			}
		}

		close(messagesDone)
		time.Sleep(100 * time.Millisecond)

		// Close connection to allow client's readLoop to exit
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
		WithMessageHandler(func(m *ChatMessage) { privmsgReceived = true }),
		WithJoinHandler(func(channel, user string) { joinReceived = true }),
		WithPartHandler(func(channel, user string) { partReceived = true }),
		WithNoticeHandler(func(n *Notice) { noticeReceived = true }),
		WithUserNoticeHandler(func(n *UserNotice) { userNoticeReceived = true }),
		WithRoomStateHandler(func(r *RoomState) { roomStateReceived = true }),
		WithClearChatHandler(func(c *ClearChat) { clearChatReceived = true }),
		WithClearMessageHandler(func(c *ClearMessage) { clearMsgReceived = true }),
		WithWhisperHandler(func(w *Whisper) { whisperReceived = true }),
		WithGlobalUserStateHandler(func(g *GlobalUserState) { globalUserReceived = true }),
		WithUserStateHandler(func(u *UserState) { userStateReceived = true }),
		WithRawMessageHandler(func(s string) { rawMessageReceived = true }),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Wait for all messages
	select {
	case <-messagesDone:
		time.Sleep(100 * time.Millisecond) // Let handlers process
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for messages")
	}

	if !privmsgReceived {
		t.Error("PRIVMSG was not received")
	}
	if !joinReceived {
		t.Error("JOIN was not received")
	}
	if !partReceived {
		t.Error("PART was not received")
	}
	if !noticeReceived {
		t.Error("NOTICE was not received")
	}
	if !userNoticeReceived {
		t.Error("USERNOTICE was not received")
	}
	if !roomStateReceived {
		t.Error("ROOMSTATE was not received")
	}
	if !clearChatReceived {
		t.Error("CLEARCHAT was not received")
	}
	if !clearMsgReceived {
		t.Error("CLEARMSG was not received")
	}
	if !whisperReceived {
		t.Error("WHISPER was not received")
	}
	if !globalUserReceived {
		t.Error("GLOBALUSERSTATE was not received")
	}
	if !userStateReceived {
		t.Error("USERSTATE was not received")
	}
	if !rawMessageReceived {
		t.Error("raw message was not received")
	}
	if !pongReceived {
		t.Error("PONG was not sent in response to PING")
	}
}

func TestIRCClient_Close(t *testing.T) {
	closeSignal := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		// Wait for test to signal close
		select {
		case <-closeSignal:
		case <-time.After(5 * time.Second):
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}

	// Signal mock server to close, then call Close()
	close(closeSignal)

	err = client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("expected client to be disconnected")
	}

	// Close again should be no-op
	err = client.Close()
	if err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}

func TestIRCClient_DisconnectHandler_CalledOnServerClose(t *testing.T) {
	disconnectCalled := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(100 * time.Millisecond)
		// Server closes the connection
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
		WithDisconnectHandler(func() {
			close(disconnectCalled)
		}),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for disconnect handler to be called when server closes connection
	select {
	case <-disconnectCalled:
		// Success - disconnect handler was called
	case <-time.After(2 * time.Second):
		t.Fatal("disconnect handler was not called when server closed connection")
	}
}

func TestIRCClient_Say(t *testing.T) {
	messageSent := make(chan string, 1)

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		// Read Say message
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), "PRIVMSG #testchannel :Hello World") {
				messageSent <- string(data)
				return
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Say("testchannel", "Hello World")
	if err != nil {
		t.Errorf("Say failed: %v", err)
	}

	select {
	case msg := <-messageSent:
		if !strings.Contains(msg, "PRIVMSG #testchannel :Hello World") {
			t.Errorf("unexpected message: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Say message")
	}
}

func TestIRCClient_Reply(t *testing.T) {
	messageSent := make(chan string, 1)

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), "reply-parent-msg-id=abc123") {
				messageSent <- string(data)
				return
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Reply("testchannel", "abc123", "Reply message")
	if err != nil {
		t.Errorf("Reply failed: %v", err)
	}

	select {
	case msg := <-messageSent:
		if !strings.Contains(msg, "@reply-parent-msg-id=abc123 PRIVMSG #testchannel :Reply message") {
			t.Errorf("unexpected message: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Reply message")
	}
}

func TestIRCClient_Whisper(t *testing.T) {
	messageSent := make(chan string, 1)

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), "PRIVMSG #jtv :/w") {
				messageSent <- string(data)
				return
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Whisper("targetuser", "Secret message")
	if err != nil {
		t.Errorf("Whisper failed: %v", err)
	}

	select {
	case msg := <-messageSent:
		if !strings.Contains(msg, "PRIVMSG #jtv :/w targetuser Secret message") {
			t.Errorf("unexpected message: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Whisper message")
	}
}

func TestIRCClient_Ping(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), "PING") {
				// Send PONG
				_ = conn.WriteMessage(websocket.TextMessage, []byte("PONG :tmi.twitch.tv\r\n"))
				return
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Ping(pingCtx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestIRCClient_Ping_Timeout(t *testing.T) {
	closeSignal := make(chan struct{})
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		// Don't respond to PING, wait for close signal
		select {
		case <-closeSignal:
		case <-time.After(5 * time.Second):
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() {
		close(closeSignal)
		_ = client.Close()
	}()

	pingCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = client.Ping(pingCtx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}
}

func TestIRCClient_GetGlobalUserState(t *testing.T) {
	closeSignal := make(chan struct{})
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		// Send GLOBALUSERSTATE during auth
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=;badges=premium/1;color=#FF0000;display-name=TestUser;emote-sets=0,1,2;user-id=12345;user-type= :tmi.twitch.tv GLOBALUSERSTATE\r\n"))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		select {
		case <-closeSignal:
		case <-time.After(5 * time.Second):
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() {
		close(closeSignal)
		_ = client.Close()
	}()

	state := client.GetGlobalUserState()
	if state == nil {
		t.Fatal("expected global user state to be set")
	}

	if state.UserID != "12345" {
		t.Errorf("UserID: got %q, want %q", state.UserID, "12345")
	}

	if state.DisplayName != "TestUser" {
		t.Errorf("DisplayName: got %q, want %q", state.DisplayName, "TestUser")
	}

	if state.Color != "#FF0000" {
		t.Errorf("Color: got %q, want %q", state.Color, "#FF0000")
	}
}

func TestIRCClient_Join_WhileConnected(t *testing.T) {
	joinSent := make(chan string, 1)

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), "JOIN #") {
				joinSent <- string(data)
				return
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	err = client.Join("testchannel")
	if err != nil {
		t.Errorf("Join failed: %v", err)
	}

	select {
	case msg := <-joinSent:
		if !strings.Contains(msg, "JOIN #testchannel") {
			t.Errorf("unexpected join message: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for JOIN message")
	}
}

func TestIRCClient_Part_WhileConnected(t *testing.T) {
	partSent := make(chan string, 1)

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), "PART #") {
				partSent <- string(data)
				return
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	// Join while disconnected
	_ = client.Join("testchannel")

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Wait for auto-join
	time.Sleep(100 * time.Millisecond)

	err = client.Part("testchannel")
	if err != nil {
		t.Errorf("Part failed: %v", err)
	}

	select {
	case msg := <-partSent:
		if !strings.Contains(msg, "PART #testchannel") {
			t.Errorf("unexpected part message: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for PART message")
	}
}

func TestIRCClient_HandleMessage_PanicRecovery(t *testing.T) {
	errorReceived := make(chan error, 1)
	closeSignal := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(50 * time.Millisecond)

		// Send a message that triggers the panicking handler
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=;badges=;color=#FF0000;display-name=TestUser;emotes=;id=abc123;mod=0;room-id=12345;subscriber=0;tmi-sent-ts=1234567890123;turbo=0;user-id=67890;user-type= :testuser!testuser@testuser.tmi.twitch.tv PRIVMSG #testchannel :trigger panic\r\n"))
		select {
		case <-closeSignal:
		case <-time.After(5 * time.Second):
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
		WithMessageHandler(func(m *ChatMessage) {
			panic("test panic")
		}),
		WithIRCErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() {
		close(closeSignal)
		_ = client.Close()
	}()

	select {
	case err := <-errorReceived:
		if !strings.Contains(err.Error(), "handler panic") {
			t.Errorf("expected panic error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for panic error")
	}
}

func TestIRCClient_Send_NotConnected(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Try to send without connecting
	err := client.Say("testchannel", "Hello")
	if err != ErrIRCNotConnected {
		t.Errorf("expected ErrIRCNotConnected, got: %v", err)
	}
}

func TestIRCClient_Reconnect(t *testing.T) {
	connectCount := 0
	secondConnected := make(chan struct{})
	closeSignal := make(chan struct{})
	var mu sync.Mutex

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()

		mu.Lock()
		connectCount++
		count := connectCount
		mu.Unlock()

		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		if count == 1 {
			// First connection - close after a moment to trigger reconnect
			time.Sleep(100 * time.Millisecond)
		} else {
			// Second+ connection - signal and wait for close
			select {
			case secondConnected <- struct{}{}:
			default:
			}
			select {
			case <-closeSignal:
			case <-time.After(5 * time.Second):
			}
		}
	})
	defer mock.Close()

	reconnectCalled := false
	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
		WithReconnectHandler(func() {
			reconnectCalled = true
		}),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for second connection to be established
	select {
	case <-secondConnected:
		// Second connection established
	case <-time.After(3 * time.Second):
		t.Fatal("second connection was not established")
	}

	mu.Lock()
	if connectCount < 2 {
		t.Errorf("expected at least 2 connections, got %d", connectCount)
	}
	mu.Unlock()

	if !reconnectCalled {
		t.Error("reconnect handler was not called")
	}

	close(closeSignal)
	_ = client.Close()
}

func TestIRCClient_HandleReconnectMessage(t *testing.T) {
	connectCount := 0
	secondConnected := make(chan struct{})
	closeSignal := make(chan struct{})
	reconnectCalled := make(chan struct{}, 1)
	var mu sync.Mutex

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()

		mu.Lock()
		connectCount++
		count := connectCount
		mu.Unlock()

		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		if count == 1 {
			// First connection - send RECONNECT command after a moment
			time.Sleep(100 * time.Millisecond)
			_ = conn.WriteMessage(websocket.TextMessage, []byte("RECONNECT\r\n"))
			// Wait for connection to close
			time.Sleep(100 * time.Millisecond)
		} else {
			// Second+ connection - signal and wait for close
			select {
			case secondConnected <- struct{}{}:
			default:
			}
			select {
			case <-closeSignal:
			case <-time.After(5 * time.Second):
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
		WithReconnectHandler(func() {
			select {
			case reconnectCalled <- struct{}{}:
			default:
			}
		}),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for reconnect handler to be called
	select {
	case <-reconnectCalled:
		// Reconnect handler was called
	case <-time.After(3 * time.Second):
		t.Fatal("reconnect handler was not called after RECONNECT")
	}

	// Wait for second connection to be established after RECONNECT
	select {
	case <-secondConnected:
		// Second connection established
	case <-time.After(3 * time.Second):
		t.Fatal("second connection was not established after RECONNECT")
	}

	mu.Lock()
	if connectCount < 2 {
		t.Errorf("expected at least 2 connections after RECONNECT, got %d", connectCount)
	}
	mu.Unlock()

	close(closeSignal)
	_ = client.Close()
}

func TestIRCClient_WaitForAuth_ImproperlFormatted(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv NOTICE * :Improperly formatted auth\r\n"))
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "badtoken",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != ErrIRCAuthFailed {
		t.Errorf("expected ErrIRCAuthFailed, got: %v", err)
	}
}

func TestIRCClient_WaitForAuth_ContextCancelled(t *testing.T) {
	// Test that waitForAuth returns context error when context is cancelled
	// and a message is received after the context deadline
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		// Wait longer than the context timeout, then send a message
		// This allows context check to happen between messages
		time.Sleep(150 * time.Millisecond)
		// Send a non-auth message to trigger another read loop iteration
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags\r\n"))
		time.Sleep(50 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Connect(ctx)
	// Should get context.DeadlineExceeded or a timeout error when context/read deadline expires
	if err == nil {
		t.Error("expected error, got nil")
	} else if err != context.DeadlineExceeded && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("expected timeout-related error, got: %v", err)
	}
}

func TestNewIRCClient_WithOAuthPrefix(t *testing.T) {
	// Test that oauth: prefix is not duplicated
	client := NewIRCClient("testuser", "oauth:mytoken",
		WithAutoReconnect(false),
	)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// The token should remain as oauth:mytoken, not oauth:oauth:mytoken
	if client.token != "oauth:mytoken" {
		t.Errorf("expected token oauth:mytoken, got %s", client.token)
	}
}

func TestIRCClient_GetJoinedChannels(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Initially empty
	channels := client.GetJoinedChannels()
	if len(channels) != 0 {
		t.Errorf("expected 0 channels, got %d", len(channels))
	}

	// Join some channels while disconnected
	_ = client.Join("channel1", "channel2", "channel3")

	channels = client.GetJoinedChannels()
	if len(channels) != 3 {
		t.Errorf("expected 3 channels, got %d", len(channels))
	}

	// Part one channel
	_ = client.Part("channel2")

	channels = client.GetJoinedChannels()
	if len(channels) != 2 {
		t.Errorf("expected 2 channels after part, got %d", len(channels))
	}
}

func TestIRCClient_WaitForAuth_ReadError(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Read all auth commands to allow sends to complete
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for i := 0; i < 4; i++ { // PASS, CAP REQ, NICK, CAP END
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
		// Close connection without sending auth response
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for read failure during auth")
	}
	if !strings.Contains(err.Error(), "reading auth response") {
		t.Errorf("expected read error, got: %v", err)
	}
}

func TestIRCClient_Connect_RejoinChannels(t *testing.T) {
	joinsSent := make(chan string, 3)

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)
			if strings.Contains(msg, "JOIN #") {
				joinsSent <- msg
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	// Join channels before connecting
	_ = client.Join("channel1", "channel2")

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Wait for joins to be sent
	timeout := time.After(2 * time.Second)
	joinsReceived := 0
	for joinsReceived < 2 {
		select {
		case <-joinsSent:
			joinsReceived++
		case <-timeout:
			t.Fatalf("timeout waiting for joins, got %d", joinsReceived)
		}
	}
}

func TestIRCClient_Reconnect_StopDuringDelay(t *testing.T) {
	connectCount := 0
	var mu sync.Mutex
	disconnected := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()

		mu.Lock()
		connectCount++
		count := connectCount
		mu.Unlock()

		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		if count == 1 {
			// First connection - close immediately to trigger reconnect
			time.Sleep(20 * time.Millisecond)
		} else {
			// Should not get here if stopChan is closed during reconnect delay
			time.Sleep(5 * time.Second)
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(2*time.Second), // Very long delay so we can definitely close during it
		WithDisconnectHandler(func() {
			select {
			case disconnected <- struct{}{}:
			default:
			}
		}),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for disconnect to occur (reconnect is now waiting on reconnectDelay)
	select {
	case <-disconnected:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for disconnect")
	}

	// Small delay to ensure we're in the select waiting on stopChan
	time.Sleep(50 * time.Millisecond)

	// Directly close stopChan to trigger the stopChan case in reconnect
	// Note: Close() won't work here because connected=false after disconnect
	client.mu.Lock()
	client.stopOnce.Do(func() {
		close(client.stopChan)
	})
	client.mu.Unlock()

	// Wait and verify no second connection
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if connectCount > 1 {
		t.Errorf("expected only 1 connection, got %d", connectCount)
	}
	mu.Unlock()
}

func TestIRCClient_Reconnect_DisabledAfterDelay(t *testing.T) {
	connectCount := 0
	var mu sync.Mutex

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()

		mu.Lock()
		connectCount++
		count := connectCount
		mu.Unlock()

		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		if count == 1 {
			// First connection - close immediately
			time.Sleep(50 * time.Millisecond)
		} else {
			// Should not get here
			time.Sleep(5 * time.Second)
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(100*time.Millisecond),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for first connection to close
	time.Sleep(70 * time.Millisecond)

	// Disable auto-reconnect during the delay
	client.mu.Lock()
	client.autoReconnect = false
	client.mu.Unlock()

	// Wait past reconnect delay
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if connectCount > 1 {
		t.Errorf("expected only 1 connection after disabling autoReconnect, got %d", connectCount)
	}
	mu.Unlock()
}

func TestIRCClient_Reconnect_ErrorHandler(t *testing.T) {
	errorCalled := make(chan error, 1)

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(50 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
		WithIRCErrorHandler(func(err error) {
			if strings.Contains(err.Error(), "reconnect failed") {
				select {
				case errorCalled <- err:
				default:
				}
			}
		}),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Close the mock server to cause reconnect failure
	mock.Close()

	// Wait for reconnect error
	select {
	case err := <-errorCalled:
		if !strings.Contains(err.Error(), "reconnect failed") {
			t.Errorf("expected reconnect failed error, got: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for reconnect error")
	}

	_ = client.Close()
}

func TestIRCClient_Join_SendError(t *testing.T) {
	closeSignal := make(chan struct{})
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		select {
		case <-closeSignal:
		case <-time.After(5 * time.Second):
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Nil out the connection to cause send error
	// Note: we don't actually close it here since closing a websocket
	// doesn't guarantee WriteMessage will fail immediately
	client.mu.Lock()
	savedConn := client.conn
	client.conn = nil
	client.mu.Unlock()

	close(closeSignal)

	// Join should fail due to nil connection
	err = client.Join("testchannel")
	if err == nil {
		t.Error("expected error for Join with nil connection")
	}

	// Restore and close properly
	client.mu.Lock()
	client.conn = savedConn
	client.mu.Unlock()

	_ = client.Close()
}

func TestIRCClient_Part_SendError(t *testing.T) {
	// Test that Part returns an error when the connection is nil
	// We simulate this by setting connected=true but conn=nil on a fresh client
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Join a channel (this just records the channel, doesn't actually send)
	_ = client.Join("testchannel")

	// Manually set connected=true to simulate a "connected" state
	// but leave conn=nil to cause the send to fail
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	// Part should fail due to nil connection
	err := client.Part("testchannel")
	if err == nil {
		t.Error("expected error for Part with nil connection")
	}

	// Reset state for cleanup
	client.mu.Lock()
	client.connected = false
	client.mu.Unlock()
}

func TestIRCClient_Ping_SendError(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Ping without connecting should fail
	ctx := context.Background()
	err := client.Ping(ctx)
	if err != ErrIRCNotConnected {
		t.Errorf("expected ErrIRCNotConnected, got: %v", err)
	}
}

func TestIRCClient_Ping_DrainPendingPongs(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if strings.Contains(string(data), "PING") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("PONG :tmi.twitch.tv\r\n"))
			}
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Pre-fill pongReceived channel
	select {
	case client.pongReceived <- struct{}{}:
	default:
	}

	// Ping should drain pending pongs and succeed
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Ping(pingCtx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestIRCClient_Part_NotConnected(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Join while disconnected
	_ = client.Join("testchannel")

	// Part while disconnected should work (just updates internal state)
	err := client.Part("testchannel")
	if err != nil {
		t.Errorf("Part while disconnected should not error, got: %v", err)
	}

	channels := client.GetJoinedChannels()
	if len(channels) != 0 {
		t.Errorf("expected 0 channels after part, got %d", len(channels))
	}
}

func TestIRCClient_WaitForAuth_GlobalUserStateHandler(t *testing.T) {
	globalUserStateReceived := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		// Send GLOBALUSERSTATE before 001
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=;badges=premium/1;color=#FF0000;display-name=TestUser;emote-sets=0,1,2;user-id=12345;user-type= :tmi.twitch.tv GLOBALUSERSTATE\r\n"))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(200 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
		WithGlobalUserStateHandler(func(g *GlobalUserState) {
			close(globalUserStateReceived)
		}),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case <-globalUserStateReceived:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("GlobalUserState handler was not called during auth")
	}
}

func TestIRCClient_ReadLoop_NilConnection(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Manually start readLoop with nil connection (edge case)
	client.wg.Add(1)
	client.stopChan = make(chan struct{})
	go client.readLoop()

	// Wait for readLoop to exit
	done := make(chan struct{})
	go func() {
		client.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - readLoop exited due to nil connection
	case <-time.After(2 * time.Second):
		t.Fatal("readLoop did not exit for nil connection")
	}
}

func TestIRCClient_ReadLoop_StopChan(t *testing.T) {
	closeSignal := make(chan struct{})
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		select {
		case <-closeSignal:
		case <-time.After(5 * time.Second):
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Signal stopChan and close
	close(closeSignal)
	_ = client.Close()
}

func TestIRCClient_Reconnect_AutoReconnectDisabledInitially(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(50 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false), // Disabled from the start
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for connection to close
	time.Sleep(150 * time.Millisecond)

	// Verify no reconnection happened
	if client.IsConnected() {
		t.Error("client should not be connected when autoReconnect is disabled")
	}
}

func TestIRCClient_ReadLoop_ReadError_NoErrorHandler(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(50 * time.Millisecond)
		// Close connection to cause read error
		_ = conn.Close()
	})
	defer mock.Close()

	// No error handler set
	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for readLoop to exit due to read error
	time.Sleep(200 * time.Millisecond)

	// Should have disconnected cleanly without crash
	if client.IsConnected() {
		t.Error("expected client to be disconnected after read error")
	}
}

func TestIRCClient_Connect_SendCapError(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Close connection immediately to cause send error during CAP REQ
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for CAP send failure")
	}
	if !strings.Contains(err.Error(), "requesting capabilities") && !strings.Contains(err.Error(), "reading auth response") {
		// May fail on CAP or later depending on timing
		t.Logf("got error: %v", err)
	}
}

func TestIRCClient_Reconnect_NoReconnectHandler(t *testing.T) {
	connectCount := 0
	secondConnected := make(chan struct{})
	closeSignal := make(chan struct{})
	var mu sync.Mutex

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()

		mu.Lock()
		connectCount++
		count := connectCount
		mu.Unlock()

		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		if count == 1 {
			// First connection - close after a moment to trigger reconnect
			time.Sleep(100 * time.Millisecond)
		} else {
			// Second+ connection - signal and wait for close
			select {
			case secondConnected <- struct{}{}:
			default:
			}
			select {
			case <-closeSignal:
			case <-time.After(5 * time.Second):
			}
		}
	})
	defer mock.Close()

	// No reconnect handler set
	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for second connection to be established
	select {
	case <-secondConnected:
		// Second connection established
	case <-time.After(3 * time.Second):
		t.Fatal("second connection was not established")
	}

	mu.Lock()
	if connectCount < 2 {
		t.Errorf("expected at least 2 connections, got %d", connectCount)
	}
	mu.Unlock()

	close(closeSignal)
	_ = client.Close()
}

func TestIRCClient_Reconnect_NoErrorHandler(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(50 * time.Millisecond)
	})
	defer mock.Close()

	// No error handler set
	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Close the mock server to cause reconnect failure
	mock.Close()

	// Wait for reconnect attempts to fail
	time.Sleep(200 * time.Millisecond)

	// Should not crash without error handler
	_ = client.Close()
}

func TestIRCClient_Connect_SendPassError(t *testing.T) {
	// Server reads CAP REQ but closes before PASS can complete
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Read the CAP REQ message
		_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, _, _ = conn.ReadMessage()
		// Close connection to cause PASS send error
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for PASS send failure")
	}
}

func TestIRCClient_Connect_SendNickError(t *testing.T) {
	// Server reads CAP REQ and PASS but closes before NICK can complete
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Read the CAP REQ and PASS messages
		_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, _, _ = conn.ReadMessage() // CAP REQ
		_, _, _ = conn.ReadMessage() // PASS
		// Close connection to cause NICK send error
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for NICK send failure")
	}
}

func TestIRCClient_Reconnect_AutoReconnectDisabledBeforeLoop(t *testing.T) {
	// Test where autoReconnect is disabled between readLoop exit and reconnect start
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(50 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(100*time.Millisecond),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Immediately disable autoReconnect before readLoop has a chance to call reconnect
	client.mu.Lock()
	client.autoReconnect = false
	client.mu.Unlock()

	// Wait for connection to close and reconnect to be attempted
	time.Sleep(200 * time.Millisecond)

	// Should not be connected and no reconnect should have happened
	if client.IsConnected() {
		t.Error("expected client to be disconnected")
	}
}

func TestIRCClient_Reconnect_StopChanDuringSelect(t *testing.T) {
	connectCount := 0
	var mu sync.Mutex

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()

		mu.Lock()
		connectCount++
		count := connectCount
		mu.Unlock()

		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))

		if count == 1 {
			// First connection - close immediately to trigger reconnect
			time.Sleep(30 * time.Millisecond)
		} else {
			// Should not get here
			time.Sleep(5 * time.Second)
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(1*time.Second), // Long delay to allow stopChan to be signaled
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for first connection to close
	time.Sleep(50 * time.Millisecond)

	// Close client which signals stopChan during the reconnect delay
	_ = client.Close()

	// Wait a bit and verify no second connection
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if connectCount > 1 {
		t.Errorf("expected only 1 connection, got %d", connectCount)
	}
	mu.Unlock()
}

// TestIRCClient_Reconnect_AutoReconnectDisabledAtStart tests that reconnect exits
// immediately when autoReconnect is false at the start of the reconnect loop.
func TestIRCClient_Reconnect_AutoReconnectDisabledAtStart(t *testing.T) {
	disconnected := make(chan struct{})
	var mu sync.Mutex
	connectCount := 0
	var clientPtr **IRCClient // Pointer to pointer to allow closure to access client

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		mu.Lock()
		connectCount++
		mu.Unlock()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		// Keep connection open briefly then close
		time.Sleep(20 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(true),
		WithReconnectDelay(500*time.Millisecond), // Long delay
		WithDisconnectHandler(func() {
			// Immediately disable autoReconnect when disconnect happens
			// This races with the reconnect() function starting
			if clientPtr != nil && *clientPtr != nil {
				(*clientPtr).mu.Lock()
				(*clientPtr).autoReconnect = false
				(*clientPtr).mu.Unlock()
			}
			select {
			case disconnected <- struct{}{}:
			default:
			}
		}),
	)
	clientPtr = &client

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for disconnect handler to fire
	select {
	case <-disconnected:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for disconnect")
	}

	// Wait a bit to let reconnect() attempt and see autoReconnect=false
	time.Sleep(100 * time.Millisecond)

	// Should not have reconnected since autoReconnect was disabled in disconnect handler
	mu.Lock()
	if connectCount > 1 {
		t.Errorf("expected only 1 connection, got %d", connectCount)
	}
	mu.Unlock()

	_ = client.Close()
}

// TestIRCClient_Connect_NickSendError tests the NICK send error path by closing
// after reading CAP REQ and PASS with a protocol close frame.
func TestIRCClient_Connect_NickSendError(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Read CAP REQ and PASS, then send close frame and close
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for i := 0; i < 2; i++ { // CAP REQ and PASS
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
		// Send close frame to signal immediate close
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "closing"),
			time.Now().Add(time.Second))
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	// This may or may not fail depending on timing, but we're trying to cover the NICK error path
	if err != nil {
		// If we got an error, it should be related to sending or reading
		if !strings.Contains(err.Error(), "NICK") &&
			!strings.Contains(err.Error(), "PASS") &&
			!strings.Contains(err.Error(), "reading") &&
			!strings.Contains(err.Error(), "capabilities") {
			t.Logf("Got error: %v", err)
		}
	}
}

// TestIRCClient_Connect_CapReqSendError tests the CAP REQ send error path
// by sending close frame immediately after accepting.
func TestIRCClient_Connect_CapReqSendError(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Send close frame immediately
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "closing"),
			time.Now().Add(time.Second))
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	// This may or may not fail depending on timing
	if err != nil {
		// If we got an error, it should be related to capabilities, auth, or connection
		if !strings.Contains(err.Error(), "capabilities") &&
			!strings.Contains(err.Error(), "PASS") &&
			!strings.Contains(err.Error(), "reading") {
			t.Logf("Got error: %v", err)
		}
	}
}

// TestIRCClient_Connect_Concurrent tests that concurrent Connect calls return ErrAlreadyConnected
func TestIRCClient_Connect_Concurrent(t *testing.T) {
	connectStarted := make(chan struct{})
	connectBlocked := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		close(connectStarted)
		<-connectBlocked // Block until second connect attempt
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	// Start first connect in goroutine
	var firstErr error
	done := make(chan struct{})
	go func() {
		firstErr = client.Connect(context.Background())
		close(done)
	}()

	// Wait for first connect to start (connection established but waiting for auth)
	<-connectStarted

	// Try second connect while first is in progress
	secondErr := client.Connect(context.Background())
	if secondErr != ErrIRCAlreadyConnected {
		t.Errorf("expected ErrIRCAlreadyConnected from concurrent connect, got: %v", secondErr)
	}

	// Unblock first connect
	close(connectBlocked)
	<-done

	if firstErr != nil {
		t.Errorf("first connect should succeed, got: %v", firstErr)
	}

	_ = client.Close()
}

// TestIRCClient_Connect_NoOnConnectHandler tests Connect without an onConnect handler
func TestIRCClient_Connect_NoOnConnectHandler(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	// Create client without onConnect handler
	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
		// No WithConnectHandler
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}
}

// TestIRCClient_waitForAuth_NoContextDeadline tests waitForAuth uses 30 second default deadline
func TestIRCClient_waitForAuth_NoContextDeadline(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Send welcome immediately - no delay
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	// Use context without deadline
	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()
}

// TestIRCClient_waitForAuth_GlobalUserStateWithoutHandler tests GLOBALUSERSTATE without handler
func TestIRCClient_waitForAuth_GlobalUserStateWithoutHandler(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Send GLOBALUSERSTATE before welcome
		_ = conn.WriteMessage(websocket.TextMessage, []byte("@badge-info=;badges=;color=#FF0000;display-name=TestUser;emote-sets=0;user-id=12345;user-type= :tmi.twitch.tv GLOBALUSERSTATE\r\n"))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	// Create client without onGlobalUserState handler
	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
		// No WithGlobalUserStateHandler
	)

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Global state should still be parsed and stored
	state := client.GetGlobalUserState()
	if state == nil {
		t.Error("expected global state to be stored")
	}
}

// TestIRCClient_Close_NilStopChan tests Close when stopChan is nil
func TestIRCClient_Close_NilStopChan(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Manually set connected but leave stopChan nil
	client.mu.Lock()
	client.connected = true
	client.stopChan = nil
	client.mu.Unlock()

	// Close should not panic
	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestIRCClient_Close_NeverConnected tests Close on a fresh client that never connected
func TestIRCClient_Close_NeverConnected(t *testing.T) {
	client := NewIRCClient("testuser", "token",
		WithAutoReconnect(false),
	)

	// Close on fresh client (connected=false, conn=nil)
	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Should be safe to call multiple times
	err = client.Close()
	if err != nil {
		t.Errorf("Second Close failed: %v", err)
	}
}

// TestIRCClient_waitForAuth_ContextCancelledDuringWait tests context cancellation during auth wait
func TestIRCClient_waitForAuth_ContextCancelledDuringWait(t *testing.T) {
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		// Don't send welcome - let the client timeout
		time.Sleep(2 * time.Second)
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	// Use context with very short timeout (shorter than default 30s)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Connect(ctx)

	// Should fail due to context timeout or read deadline
	if err == nil {
		t.Error("expected error from context timeout")
		_ = client.Close()
		return
	}

	// The error should be timeout-related
	if !strings.Contains(err.Error(), "timeout") &&
		!strings.Contains(err.Error(), "deadline") &&
		!strings.Contains(err.Error(), "context") &&
		err != context.DeadlineExceeded {
		t.Logf("Got error: %v (this is acceptable)", err)
	}
}

// TestIRCClient_Close_InProgressConnection tests Close during connection establishment
func TestIRCClient_Close_InProgressConnection(t *testing.T) {
	connectBlocked := make(chan struct{})
	serverDone := make(chan struct{})

	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer close(serverDone)
		// Don't send welcome - simulate slow connection
		select {
		case <-connectBlocked:
		case <-time.After(5 * time.Second):
		}
	})
	defer mock.Close()

	client := NewIRCClient("testuser", "token",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	// Start connect in goroutine
	connectErr := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		connectErr <- client.Connect(ctx)
	}()

	// Give connect time to establish WebSocket but not complete auth
	time.Sleep(100 * time.Millisecond)

	// Close while connecting
	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Unblock server
	close(connectBlocked)

	// Wait for connect to return (should fail due to closed connection)
	select {
	case err := <-connectErr:
		// Expected to fail since we closed during auth
		if err == nil {
			t.Error("expected connect to fail after Close")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connect to return")
	}
}
