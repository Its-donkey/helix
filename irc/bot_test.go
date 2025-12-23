package irc

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewBot(t *testing.T) {
	bot := NewBot("testbot", "token123")

	if bot.nick != "testbot" {
		t.Errorf("nick: got %q, want %q", bot.nick, "testbot")
	}

	if bot.token != "token123" {
		t.Errorf("token: got %q, want %q", bot.token, "token123")
	}
}

func TestBotOptions(t *testing.T) {
	bot := NewBot("testbot", "token",
		WithBotURL("wss://custom.url"),
		WithBotAutoReconnect(false),
	)

	if bot.url != "wss://custom.url" {
		t.Errorf("url: got %q, want %q", bot.url, "wss://custom.url")
	}

	if bot.autoReconnect == nil || *bot.autoReconnect != false {
		t.Error("autoReconnect should be false")
	}
}

func TestBotEventHandlers(t *testing.T) {
	bot := NewBot("testbot", "token")

	var (
		msgCalled     bool
		subCalled     bool
		resubCalled   bool
		giftCalled    bool
		raidCalled    bool
		cheerCalled   bool
		joinCalled    bool
		partCalled    bool
		roomCalled    bool
		noticeCalled  bool
		clearCalled   bool
		whisperCalled bool
		connectCalled bool
		disconnCalled bool
		errorCalled   bool
	)

	bot.OnMessage(func(m *ChatMessage) { msgCalled = true })
	bot.OnSub(func(n *UserNotice) { subCalled = true })
	bot.OnResub(func(n *UserNotice) { resubCalled = true })
	bot.OnSubGift(func(n *UserNotice) { giftCalled = true })
	bot.OnRaid(func(n *UserNotice) { raidCalled = true })
	bot.OnCheer(func(m *ChatMessage) { cheerCalled = true })
	bot.OnJoin(func(ch, u string) { joinCalled = true })
	bot.OnPart(func(ch, u string) { partCalled = true })
	bot.OnRoomState(func(s *RoomState) { roomCalled = true })
	bot.OnNotice(func(n *Notice) { noticeCalled = true })
	bot.OnClearChat(func(c *ClearChat) { clearCalled = true })
	bot.OnWhisper(func(w *Whisper) { whisperCalled = true })
	bot.OnConnect(func() { connectCalled = true })
	bot.OnDisconnect(func() { disconnCalled = true })
	bot.OnError(func(err error) { errorCalled = true })

	// Test handlers via internal methods
	bot.handleMessage(&ChatMessage{})
	if !msgCalled {
		t.Error("OnMessage not called")
	}

	bot.handleMessage(&ChatMessage{Bits: 100})
	if !cheerCalled {
		t.Error("OnCheer not called")
	}

	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeSub})
	if !subCalled {
		t.Error("OnSub not called")
	}

	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeResub})
	if !resubCalled {
		t.Error("OnResub not called")
	}

	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeSubGift})
	if !giftCalled {
		t.Error("OnSubGift not called")
	}

	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeRaid})
	if !raidCalled {
		t.Error("OnRaid not called")
	}

	bot.handleJoin("channel", "user")
	if !joinCalled {
		t.Error("OnJoin not called")
	}

	bot.handlePart("channel", "user")
	if !partCalled {
		t.Error("OnPart not called")
	}

	bot.handleRoomState(&RoomState{})
	if !roomCalled {
		t.Error("OnRoomState not called")
	}

	bot.handleNotice(&Notice{})
	if !noticeCalled {
		t.Error("OnNotice not called")
	}

	bot.handleClearChat(&ClearChat{})
	if !clearCalled {
		t.Error("OnClearChat not called")
	}

	bot.handleWhisper(&Whisper{})
	if !whisperCalled {
		t.Error("OnWhisper not called")
	}

	bot.handleConnect()
	if !connectCalled {
		t.Error("OnConnect not called")
	}

	bot.handleDisconnect()
	if !disconnCalled {
		t.Error("OnDisconnect not called")
	}

	bot.handleError(errors.New("test"))
	if !errorCalled {
		t.Error("OnError not called")
	}
}

func TestBotUserNoticeTypes(t *testing.T) {
	bot := NewBot("testbot", "token")

	var anonGiftCalled, mysteryGiftCalled bool

	bot.OnSubGift(func(n *UserNotice) {
		if n.Type == UserNoticeTypeAnonSubGift {
			anonGiftCalled = true
		}
		if n.Type == UserNoticeTypeSubMysteryGift {
			mysteryGiftCalled = true
		}
	})

	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeAnonSubGift})
	if !anonGiftCalled {
		t.Error("AnonSubGift should trigger OnSubGift")
	}

	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeSubMysteryGift})
	if !mysteryGiftCalled {
		t.Error("SubMysteryGift should trigger OnSubGift")
	}
}

func TestBotMethodsNotConnected(t *testing.T) {
	bot := NewBot("testbot", "token")

	if bot.IsConnected() {
		t.Error("Bot should not be connected")
	}

	if bot.GetJoinedChannels() != nil {
		t.Error("GetJoinedChannels should return nil")
	}

	if bot.Client() != nil {
		t.Error("Client should return nil")
	}

	err := bot.Join("channel")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Join should return ErrNotConnected, got: %v", err)
	}

	err = bot.Part("channel")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Part should return ErrNotConnected, got: %v", err)
	}

	err = bot.Say("channel", "message")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Say should return ErrNotConnected, got: %v", err)
	}

	err = bot.Reply("channel", "msgid", "message")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Reply should return ErrNotConnected, got: %v", err)
	}

	err = bot.Whisper("user", "message")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Whisper should return ErrNotConnected, got: %v", err)
	}

	// Close should not error
	err = bot.Close()
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}
}

func TestBotHandlersWithNilCallbacks(t *testing.T) {
	bot := NewBot("testbot", "token")

	// All handlers should be nil-safe
	bot.handleMessage(&ChatMessage{})
	bot.handleMessage(&ChatMessage{Bits: 100})
	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeSub})
	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeResub})
	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeSubGift})
	bot.handleUserNotice(&UserNotice{Type: UserNoticeTypeRaid})
	bot.handleJoin("channel", "user")
	bot.handlePart("channel", "user")
	bot.handleRoomState(&RoomState{})
	bot.handleNotice(&Notice{})
	bot.handleClearChat(&ClearChat{})
	bot.handleWhisper(&Whisper{})
	bot.handleConnect()
	bot.handleDisconnect()
	bot.handleError(nil)
}

func TestBotConnectWithMockServer(t *testing.T) {
	server := createMockIRCServer(t, func(conn *websocket.Conn) {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)

			if strings.HasPrefix(msg, "CAP REQ") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags\r\n"))
			} else if strings.HasPrefix(msg, "NICK") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testbot :Welcome\r\n"))
				return
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	connectCalled := false
	bot := NewBot("testbot", "token",
		WithBotURL(wsURL),
		WithBotAutoReconnect(false),
	)
	bot.OnConnect(func() { connectCalled = true })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bot.Connect(ctx)
	if err != nil {
		t.Fatalf("Bot.Connect failed: %v", err)
	}

	if !bot.IsConnected() {
		t.Error("Bot should be connected")
	}

	if !connectCalled {
		t.Error("OnConnect handler should be called")
	}

	if bot.Client() == nil {
		t.Error("Client should not be nil after connect")
	}

	_ = bot.Close()
}

func TestBotMethodsWhenConnected(t *testing.T) {
	server := createMockIRCServer(t, func(conn *websocket.Conn) {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)

			if strings.HasPrefix(msg, "CAP REQ") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags\r\n"))
			} else if strings.HasPrefix(msg, "NICK") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testbot :Welcome\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	bot := NewBot("testbot", "token",
		WithBotURL(wsURL),
		WithBotAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bot.Connect(ctx)
	if err != nil {
		t.Fatalf("Bot.Connect failed: %v", err)
	}
	defer func() { _ = bot.Close() }()

	// Test Join
	err = bot.Join("testchannel")
	if err != nil {
		t.Errorf("Join failed: %v", err)
	}

	// Test Part
	err = bot.Part("testchannel")
	if err != nil {
		t.Errorf("Part failed: %v", err)
	}

	// Test Say
	err = bot.Say("testchannel", "Hello!")
	if err != nil {
		t.Errorf("Say failed: %v", err)
	}

	// Test Reply
	err = bot.Reply("testchannel", "msg-123", "Hello!")
	if err != nil {
		t.Errorf("Reply failed: %v", err)
	}

	// Test Whisper
	err = bot.Whisper("someuser", "Hello!")
	if err != nil {
		t.Errorf("Whisper failed: %v", err)
	}

	// Test GetJoinedChannels
	channels := bot.GetJoinedChannels()
	if channels == nil {
		t.Error("GetJoinedChannels should not return nil when connected")
	}

	// Test IsConnected
	if !bot.IsConnected() {
		t.Error("Bot should still be connected")
	}
}

func TestBotIsConnectedWithClient(t *testing.T) {
	server := createMockIRCServer(t, func(conn *websocket.Conn) {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)

			if strings.HasPrefix(msg, "CAP REQ") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags\r\n"))
			} else if strings.HasPrefix(msg, "NICK") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testbot :Welcome\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	bot := NewBot("testbot", "token",
		WithBotURL(wsURL),
		WithBotAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := bot.Connect(ctx)
	if err != nil {
		t.Fatalf("Bot.Connect failed: %v", err)
	}
	defer func() { _ = bot.Close() }()

	// Test IsConnected returns true
	if !bot.IsConnected() {
		t.Error("Bot should be connected")
	}

	// Test GetJoinedChannels when client exists
	channels := bot.GetJoinedChannels()
	if channels == nil {
		t.Error("GetJoinedChannels should return empty slice, not nil")
	}
}
