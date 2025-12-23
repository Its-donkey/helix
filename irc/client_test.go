package irc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Mock WebSocket server for testing

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func createMockIRCServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade: %v", err)
			return
		}
		defer func() { _ = conn.Close() }()
		handler(conn)
	}))
}

func TestNewClient(t *testing.T) {
	client := NewClient("testuser", "token123")

	if client.nick != "testuser" {
		t.Errorf("nick: got %q, want %q", client.nick, "testuser")
	}

	if client.token != "oauth:token123" {
		t.Errorf("token: got %q, want %q", client.token, "oauth:token123")
	}

	if client.url != TwitchWebSocket {
		t.Errorf("url: got %q, want %q", client.url, TwitchWebSocket)
	}

	// Test with oauth: prefix already present
	client2 := NewClient("testuser", "oauth:token456")
	if client2.token != "oauth:token456" {
		t.Errorf("token with prefix: got %q, want %q", client2.token, "oauth:token456")
	}
}

func TestClientOptions(t *testing.T) {
	messageReceived := false
	errorReceived := false

	client := NewClient("testuser", "token",
		WithURL("wss://custom.url"),
		WithAutoReconnect(false),
		WithReconnectDelay(10*time.Second),
		WithMessageHandler(func(m *ChatMessage) {
			messageReceived = true
		}),
		WithErrorHandler(func(err error) {
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
	client := NewClient("testuser", "token")

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

func TestAllClientOptions(t *testing.T) {
	var (
		joinCalled        bool
		partCalled        bool
		noticeCalled      bool
		userNoticeCalled  bool
		roomStateCalled   bool
		clearChatCalled   bool
		clearMsgCalled    bool
		whisperCalled     bool
		globalStateCalled bool
		userStateCalled   bool
		connectCalled     bool
		disconnectCalled  bool
		reconnectCalled   bool
		rawMsgCalled      bool
	)

	client := NewClient("testuser", "token",
		WithJoinHandler(func(channel, user string) { joinCalled = true }),
		WithPartHandler(func(channel, user string) { partCalled = true }),
		WithNoticeHandler(func(n *Notice) { noticeCalled = true }),
		WithUserNoticeHandler(func(n *UserNotice) { userNoticeCalled = true }),
		WithRoomStateHandler(func(s *RoomState) { roomStateCalled = true }),
		WithClearChatHandler(func(c *ClearChat) { clearChatCalled = true }),
		WithClearMessageHandler(func(c *ClearMessage) { clearMsgCalled = true }),
		WithWhisperHandler(func(w *Whisper) { whisperCalled = true }),
		WithGlobalUserStateHandler(func(s *GlobalUserState) { globalStateCalled = true }),
		WithUserStateHandler(func(s *UserState) { userStateCalled = true }),
		WithConnectHandler(func() { connectCalled = true }),
		WithDisconnectHandler(func() { disconnectCalled = true }),
		WithReconnectHandler(func() { reconnectCalled = true }),
		WithRawMessageHandler(func(s string) { rawMsgCalled = true }),
	)

	// Test all handlers are set and callable
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
	client.onRawMessage("raw")

	if !joinCalled {
		t.Error("join handler not called")
	}
	if !partCalled {
		t.Error("part handler not called")
	}
	if !noticeCalled {
		t.Error("notice handler not called")
	}
	if !userNoticeCalled {
		t.Error("userNotice handler not called")
	}
	if !roomStateCalled {
		t.Error("roomState handler not called")
	}
	if !clearChatCalled {
		t.Error("clearChat handler not called")
	}
	if !clearMsgCalled {
		t.Error("clearMessage handler not called")
	}
	if !whisperCalled {
		t.Error("whisper handler not called")
	}
	if !globalStateCalled {
		t.Error("globalUserState handler not called")
	}
	if !userStateCalled {
		t.Error("userState handler not called")
	}
	if !connectCalled {
		t.Error("connect handler not called")
	}
	if !disconnectCalled {
		t.Error("disconnect handler not called")
	}
	if !reconnectCalled {
		t.Error("reconnect handler not called")
	}
	if !rawMsgCalled {
		t.Error("rawMessage handler not called")
	}
}

func TestClientClose(t *testing.T) {
	client := NewClient("testuser", "token")

	// Close on unconnected client should not error
	err := client.Close()
	if err != nil {
		t.Errorf("Close on unconnected client: %v", err)
	}
}

func TestClientGetGlobalUserState(t *testing.T) {
	client := NewClient("testuser", "token")

	// Should return nil initially
	state := client.GetGlobalUserState()
	if state != nil {
		t.Error("GetGlobalUserState should return nil initially")
	}

	// Set a state
	client.globalState = &GlobalUserState{UserID: "12345"}
	state = client.GetGlobalUserState()
	if state == nil || state.UserID != "12345" {
		t.Error("GetGlobalUserState should return the set state")
	}
}

func TestClientSendNotConnected(t *testing.T) {
	client := NewClient("testuser", "token")

	err := client.Say("channel", "message")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Say should return ErrNotConnected, got: %v", err)
	}

	err = client.Reply("channel", "msgid", "message")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Reply should return ErrNotConnected, got: %v", err)
	}

	err = client.Whisper("user", "message")
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Whisper should return ErrNotConnected, got: %v", err)
	}
}

func TestClientConnect(t *testing.T) {
	server := createMockIRCServer(t, func(conn *websocket.Conn) {
		// Read and respond to connection sequence
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)

			if strings.HasPrefix(msg, "CAP REQ") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags twitch.tv/commands twitch.tv/membership\r\n"))
			} else if strings.HasPrefix(msg, "PASS") {
				// Continue
			} else if strings.HasPrefix(msg, "NICK") {
				// Send welcome
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome, GLHF!\r\n"))
				return
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	// Test double connect
	err = client.Connect(ctx)
	if !errors.Is(err, ErrAlreadyConnected) {
		t.Errorf("Second Connect should return ErrAlreadyConnected, got: %v", err)
	}

	_ = client.Close()
}

func TestClientAuthFailed(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv NOTICE * :Login authentication failed\r\n"))
				return
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "badtoken", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if !errors.Is(err, ErrAuthFailed) {
		t.Errorf("Connect should return ErrAuthFailed, got: %v", err)
	}
}

func TestClientHandleMessage(t *testing.T) {
	var (
		msgReceived         bool
		whisperReceived     bool
		noticeReceived      bool
		userNoticeReceived  bool
		roomStateReceived   bool
		clearChatReceived   bool
		clearMsgReceived    bool
		globalStateReceived bool
		userStateReceived   bool
		joinReceived        bool
		partReceived        bool
	)

	client := NewClient("testuser", "token",
		WithMessageHandler(func(m *ChatMessage) { msgReceived = true }),
		WithWhisperHandler(func(w *Whisper) { whisperReceived = true }),
		WithNoticeHandler(func(n *Notice) { noticeReceived = true }),
		WithUserNoticeHandler(func(n *UserNotice) { userNoticeReceived = true }),
		WithRoomStateHandler(func(s *RoomState) { roomStateReceived = true }),
		WithClearChatHandler(func(c *ClearChat) { clearChatReceived = true }),
		WithClearMessageHandler(func(c *ClearMessage) { clearMsgReceived = true }),
		WithGlobalUserStateHandler(func(s *GlobalUserState) { globalStateReceived = true }),
		WithUserStateHandler(func(s *UserState) { userStateReceived = true }),
		WithJoinHandler(func(ch, u string) { joinReceived = true }),
		WithPartHandler(func(ch, u string) { partReceived = true }),
	)

	// Initialize pongReceived channel
	client.pongReceived = make(chan struct{}, 1)

	// Test PING handling
	client.handleMessage("PING :tmi.twitch.tv")

	// Test PONG handling
	client.handleMessage("PONG :tmi.twitch.tv")
	select {
	case <-client.pongReceived:
		// Good
	default:
		t.Error("PONG should signal pongReceived")
	}

	// Test PRIVMSG
	client.handleMessage("@id=123 :user!user@user.tmi.twitch.tv PRIVMSG #channel :Hello")
	if !msgReceived {
		t.Error("PRIVMSG should trigger message handler")
	}

	// Test WHISPER
	client.handleMessage("@user-id=123 :user!user@user.tmi.twitch.tv WHISPER testuser :Hello")
	if !whisperReceived {
		t.Error("WHISPER should trigger whisper handler")
	}

	// Test NOTICE
	client.handleMessage("@msg-id=subs_on :tmi.twitch.tv NOTICE #channel :Subs only")
	if !noticeReceived {
		t.Error("NOTICE should trigger notice handler")
	}

	// Test USERNOTICE
	client.handleMessage("@msg-id=sub :tmi.twitch.tv USERNOTICE #channel")
	if !userNoticeReceived {
		t.Error("USERNOTICE should trigger userNotice handler")
	}

	// Test ROOMSTATE
	client.handleMessage("@room-id=123 :tmi.twitch.tv ROOMSTATE #channel")
	if !roomStateReceived {
		t.Error("ROOMSTATE should trigger roomState handler")
	}

	// Test CLEARCHAT
	client.handleMessage("@room-id=123 :tmi.twitch.tv CLEARCHAT #channel :user")
	if !clearChatReceived {
		t.Error("CLEARCHAT should trigger clearChat handler")
	}

	// Test CLEARMSG
	client.handleMessage("@target-msg-id=123 :tmi.twitch.tv CLEARMSG #channel :message")
	if !clearMsgReceived {
		t.Error("CLEARMSG should trigger clearMessage handler")
	}

	// Test GLOBALUSERSTATE
	client.handleMessage("@user-id=123 :tmi.twitch.tv GLOBALUSERSTATE")
	if !globalStateReceived {
		t.Error("GLOBALUSERSTATE should trigger globalUserState handler")
	}

	// Test USERSTATE
	client.handleMessage("@mod=1 :tmi.twitch.tv USERSTATE #channel")
	if !userStateReceived {
		t.Error("USERSTATE should trigger userState handler")
	}

	// Test JOIN
	client.handleMessage(":user!user@user.tmi.twitch.tv JOIN #channel")
	if !joinReceived {
		t.Error("JOIN should trigger join handler")
	}

	// Test PART
	client.handleMessage(":user!user@user.tmi.twitch.tv PART #channel")
	if !partReceived {
		t.Error("PART should trigger part handler")
	}
}

func TestClientPing(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
			} else if strings.HasPrefix(msg, "PING") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("PONG :tmi.twitch.tv\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Give readLoop time to start
	time.Sleep(100 * time.Millisecond)

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	_ = client.Close()
}

func TestClientHandleMessageWithNilHandlers(t *testing.T) {
	client := NewClient("testuser", "token")
	client.pongReceived = make(chan struct{}, 1)

	// All these should not panic with nil handlers
	client.handleMessage("PING :tmi.twitch.tv")
	client.handleMessage("PONG :tmi.twitch.tv")
	client.handleMessage("@id=123 :user!user@user.tmi.twitch.tv PRIVMSG #channel :Hello")
	client.handleMessage("@user-id=123 :user!user@user.tmi.twitch.tv WHISPER testuser :Hello")
	client.handleMessage("@msg-id=subs_on :tmi.twitch.tv NOTICE #channel :Subs only")
	client.handleMessage("@msg-id=sub :tmi.twitch.tv USERNOTICE #channel")
	client.handleMessage("@room-id=123 :tmi.twitch.tv ROOMSTATE #channel")
	client.handleMessage("@room-id=123 :tmi.twitch.tv CLEARCHAT #channel :user")
	client.handleMessage("@target-msg-id=123 :tmi.twitch.tv CLEARMSG #channel :message")
	client.handleMessage("@user-id=123 :tmi.twitch.tv GLOBALUSERSTATE")
	client.handleMessage("@mod=1 :tmi.twitch.tv USERSTATE #channel")
	client.handleMessage(":user!user@user.tmi.twitch.tv JOIN #channel")
	client.handleMessage(":user!user@user.tmi.twitch.tv PART #channel")
	client.handleMessage(":tmi.twitch.tv RECONNECT")
}

func TestClientReconnect(t *testing.T) {
	connectCount := 0
	reconnectCalled := false
	disconnectCalled := false

	server := createMockIRCServer(t, func(conn *websocket.Conn) {
		connectCount++
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)

			if strings.HasPrefix(msg, "CAP REQ") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags\r\n"))
			} else if strings.HasPrefix(msg, "NICK") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
				// On first connect, close immediately to trigger reconnect
				if connectCount == 1 {
					return
				}
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := NewClient("testuser", "token",
		WithURL(wsURL),
		WithAutoReconnect(true),
		WithReconnectDelay(100*time.Millisecond),
		WithReconnectHandler(func() { reconnectCalled = true }),
		WithDisconnectHandler(func() { disconnectCalled = true }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for reconnect
	time.Sleep(500 * time.Millisecond)

	if !disconnectCalled {
		t.Error("Disconnect handler should be called")
	}

	if !reconnectCalled {
		t.Error("Reconnect handler should be called")
	}

	if connectCount < 2 {
		t.Errorf("Should have connected at least twice, got %d", connectCount)
	}

	_ = client.Close()
}

func TestClientReconnectStopped(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
				return // Close connection
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := NewClient("testuser", "token",
		WithURL(wsURL),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait a bit for disconnect to happen
	time.Sleep(100 * time.Millisecond)

	// Close should stop reconnect loop
	_ = client.Close()

	// Verify reconnect loop stops
	time.Sleep(200 * time.Millisecond)
}

func TestClientJoinPartWhenConnected(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Test Join when connected
	err = client.Join("channel1", "channel2")
	if err != nil {
		t.Errorf("Join failed: %v", err)
	}

	channels := client.GetJoinedChannels()
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}

	// Test Part when connected
	err = client.Part("channel1")
	if err != nil {
		t.Errorf("Part failed: %v", err)
	}

	channels = client.GetJoinedChannels()
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel after Part, got %d", len(channels))
	}
}

func TestClientSendMethodsWhenConnected(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Test Say
	err = client.Say("channel", "Hello!")
	if err != nil {
		t.Errorf("Say failed: %v", err)
	}

	// Test Reply
	err = client.Reply("channel", "msg-123", "Reply!")
	if err != nil {
		t.Errorf("Reply failed: %v", err)
	}

	// Test Whisper
	err = client.Whisper("user", "Whisper!")
	if err != nil {
		t.Errorf("Whisper failed: %v", err)
	}
}

func TestClientCloseWhenConnected(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("Client should not be connected after Close")
	}
}

func TestWaitForAuthWithGlobalUserState(t *testing.T) {
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
				// Send GLOBALUSERSTATE before welcome
				_ = conn.WriteMessage(websocket.TextMessage, []byte("@user-id=12345;display-name=TestUser :tmi.twitch.tv GLOBALUSERSTATE\r\n"))
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
				return
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	globalStateCalled := false
	client := NewClient("testuser", "token",
		WithURL(wsURL),
		WithAutoReconnect(false),
		WithGlobalUserStateHandler(func(s *GlobalUserState) { globalStateCalled = true }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	if !globalStateCalled {
		t.Error("GlobalUserState handler should be called during auth")
	}

	state := client.GetGlobalUserState()
	if state == nil || state.UserID != "12345" {
		t.Error("GlobalUserState should be set")
	}
}

func TestWaitForAuthImproperFormat(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv NOTICE * :Improperly formatted auth\r\n"))
				return
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "badtoken", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if !errors.Is(err, ErrAuthFailed) {
		t.Errorf("Connect should return ErrAuthFailed for improper format, got: %v", err)
	}
}

func TestReconnectWithError(t *testing.T) {
	connectCount := 0
	errorReceived := false

	server := createMockIRCServer(t, func(conn *websocket.Conn) {
		connectCount++
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)

			if strings.HasPrefix(msg, "CAP REQ") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags\r\n"))
			} else if strings.HasPrefix(msg, "NICK") {
				if connectCount == 1 {
					_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
					return // Close to trigger reconnect
				} else {
					// On reconnect attempt, fail auth
					_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv NOTICE * :Login authentication failed\r\n"))
					return
				}
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := NewClient("testuser", "token",
		WithURL(wsURL),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
		WithErrorHandler(func(err error) { errorReceived = true }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for reconnect attempt
	time.Sleep(300 * time.Millisecond)

	_ = client.Close()

	if !errorReceived {
		t.Error("Error handler should be called on reconnect failure")
	}
}

func TestPingTimeout(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
			}
			// Don't respond to PING - let it timeout
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Ping with very short timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer pingCancel()

	err = client.Ping(pingCtx)
	if err == nil {
		t.Error("Ping should timeout")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got: %v", err)
	}
}

func TestConnectInvalidURL(t *testing.T) {
	client := NewClient("testuser", "token",
		WithURL("ws://localhost:99999/invalid"),
		WithAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Connect should fail with invalid URL")
	}
}

func TestJoinPartMultipleChannelsConnected(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token", WithURL(wsURL), WithAutoReconnect(false))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Test joining multiple channels at once
	err = client.Join("channel1", "#channel2", "CHANNEL3")
	if err != nil {
		t.Errorf("Join multiple channels failed: %v", err)
	}

	channels := client.GetJoinedChannels()
	if len(channels) != 3 {
		t.Errorf("Expected 3 channels, got %d", len(channels))
	}

	// Test parting multiple channels
	err = client.Part("channel1", "channel2")
	if err != nil {
		t.Errorf("Part multiple channels failed: %v", err)
	}

	channels = client.GetJoinedChannels()
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel after Part, got %d", len(channels))
	}
}

func TestReadLoopRawMessageHandler(t *testing.T) {
	rawMessages := []string{}

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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
				// Send a test message
				_ = conn.WriteMessage(websocket.TextMessage, []byte("@id=test :user!user@user.tmi.twitch.tv PRIVMSG #channel :test\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token",
		WithURL(wsURL),
		WithAutoReconnect(false),
		WithRawMessageHandler(func(raw string) {
			rawMessages = append(rawMessages, raw)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for messages to be received
	time.Sleep(100 * time.Millisecond)

	_ = client.Close()

	if len(rawMessages) == 0 {
		t.Error("Should have received raw messages")
	}
}

func TestCloseEdgeCases(t *testing.T) {
	client := NewClient("testuser", "token")

	// Close when stopChan is nil
	client.stopChan = nil
	client.connected = true
	err := client.Close()
	if err != nil {
		t.Errorf("Close with nil stopChan should not error: %v", err)
	}

	// Close when conn is nil but connected
	client2 := NewClient("testuser", "token")
	client2.connected = true
	client2.stopChan = make(chan struct{})
	err = client2.Close()
	if err != nil {
		t.Errorf("Close with nil conn should not error: %v", err)
	}
}

func TestHandleMessageReconnect(t *testing.T) {
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
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
				// Send RECONNECT command
				time.Sleep(50 * time.Millisecond)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv RECONNECT\r\n"))
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client := NewClient("testuser", "token",
		WithURL(wsURL),
		WithAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for RECONNECT to be processed and connection to close
	time.Sleep(200 * time.Millisecond)

	// After RECONNECT, client should not be connected
	if client.IsConnected() {
		t.Error("Client should be disconnected after RECONNECT command")
	}

	_ = client.Close()
}

func TestPingNotConnected(t *testing.T) {
	client := NewClient("testuser", "token")

	ctx := context.Background()
	err := client.Ping(ctx)
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("Ping should return ErrNotConnected, got: %v", err)
	}
}

func TestChannelsRejoinedOnReconnect(t *testing.T) {
	connectCount := 0
	joinCount := 0

	server := createMockIRCServer(t, func(conn *websocket.Conn) {
		connectCount++
		currentConnect := connectCount
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := string(data)

			if strings.HasPrefix(msg, "CAP REQ") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv CAP * ACK :twitch.tv/tags\r\n"))
			} else if strings.HasPrefix(msg, "NICK") {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(":tmi.twitch.tv 001 testuser :Welcome\r\n"))
			} else if strings.HasPrefix(msg, "JOIN") {
				joinCount++
				if currentConnect == 1 {
					// Close after first connect to trigger reconnect
					time.Sleep(10 * time.Millisecond)
					return
				}
			}
		}
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := NewClient("testuser", "token",
		WithURL(wsURL),
		WithAutoReconnect(true),
		WithReconnectDelay(50*time.Millisecond),
	)

	// Pre-join a channel (will be queued)
	_ = client.Join("testchannel")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Wait for initial join and reconnect
	time.Sleep(300 * time.Millisecond)

	_ = client.Close()

	if connectCount < 2 {
		t.Errorf("Should have connected at least twice, got %d", connectCount)
	}

	if joinCount < 2 {
		t.Errorf("JOIN should be sent at least twice (initial + reconnect), got %d", joinCount)
	}
}
