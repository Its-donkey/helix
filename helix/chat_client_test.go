package helix

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Twitch IRC sample data based on actual Twitch output formats
// Reference: https://dev.twitch.tv/docs/irc/tags

const (
	// Sample Twitch welcome message
	twitchWelcome = ":tmi.twitch.tv 001 justinfan12345 :Welcome, GLHF!\r\n"

	// Sample CAP ACK from Twitch
	twitchCapAck = ":tmi.twitch.tv CAP * ACK :twitch.tv/membership twitch.tv/tags twitch.tv/commands\r\n"
)

func TestNewChatBotClient(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.nick != "justinfan12345" {
		t.Errorf("expected nick justinfan12345, got %s", client.nick)
	}
}

func TestNewChatBotClient_WithOptions(t *testing.T) {
	called := false
	opt := func(c *ChatBotClient) {
		called = true
	}
	client := NewChatBotClient("justinfan12345", nil, opt)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if !called {
		t.Error("option was not called")
	}
}

func TestChatBotClient_Connect_NoToken(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	err := client.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "chatbot: no authentication token available" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatBotClient_Connect_NilTokenFromAuthClient(t *testing.T) {
	// Auth client exists but GetToken returns nil
	authClient := &AuthClient{}
	// Don't set token - authClient.token is nil

	client := NewChatBotClient("justinfan12345", authClient)
	err := client.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "chatbot: no authentication token available" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatBotClient_Connect_EmptyNick(t *testing.T) {
	// Create a mock auth client with a token
	authClient := &AuthClient{}
	authClient.token = &Token{AccessToken: "oauth:abcdefghijklmnop123456789"}

	client := NewChatBotClient("", authClient)
	err := client.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "chatbot: failed to create IRC client (invalid nick or token)" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatBotClient_Close_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	err := client.Close()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestChatBotClient_IsConnected_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	if client.IsConnected() {
		t.Error("expected false for nil IRC client")
	}
}

func TestChatBotClient_Join_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	err := client.Join("#dallas")
	if err != ErrIRCNotConnected {
		t.Errorf("expected ErrIRCNotConnected, got %v", err)
	}
}

func TestChatBotClient_Part_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	err := client.Part("#dallas")
	if err != ErrIRCNotConnected {
		t.Errorf("expected ErrIRCNotConnected, got %v", err)
	}
}

func TestChatBotClient_Say_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	err := client.Say("#dallas", "Hello chat!")
	if err != ErrIRCNotConnected {
		t.Errorf("expected ErrIRCNotConnected, got %v", err)
	}
}

func TestChatBotClient_Reply_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	err := client.Reply("#dallas", "b34ccfc7-4977-403a-8a94-33c6bac34fb8", "Hello!")
	if err != ErrIRCNotConnected {
		t.Errorf("expected ErrIRCNotConnected, got %v", err)
	}
}

func TestChatBotClient_Whisper_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	err := client.Whisper("ronni", "secret message")
	if err != ErrIRCNotConnected {
		t.Errorf("expected ErrIRCNotConnected, got %v", err)
	}
}

func TestChatBotClient_GetJoinedChannels_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	channels := client.GetJoinedChannels()
	if channels != nil {
		t.Errorf("expected nil, got %v", channels)
	}
}

func TestChatBotClient_IRC_NilIRC(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	if client.IRC() != nil {
		t.Error("expected nil IRC client")
	}
}

func TestChatBotClient_OnMessage(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnMessage(func(msg *ChatMessage) {})
	if client.onMessage == nil {
		t.Error("expected onMessage to be set")
	}
}

func TestChatBotClient_OnSub(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnSub(func(notice *UserNotice) {})
	if client.onSub == nil {
		t.Error("expected onSub to be set")
	}
}

func TestChatBotClient_OnResub(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnResub(func(notice *UserNotice) {})
	if client.onResub == nil {
		t.Error("expected onResub to be set")
	}
}

func TestChatBotClient_OnSubGift(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnSubGift(func(notice *UserNotice) {})
	if client.onSubGift == nil {
		t.Error("expected onSubGift to be set")
	}
}

func TestChatBotClient_OnRaid(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnRaid(func(notice *UserNotice) {})
	if client.onRaid == nil {
		t.Error("expected onRaid to be set")
	}
}

func TestChatBotClient_OnCheer(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnCheer(func(msg *ChatMessage) {})
	if client.onCheer == nil {
		t.Error("expected onCheer to be set")
	}
}

func TestChatBotClient_OnJoin(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnJoin(func(channel, user string) {})
	if client.onJoin == nil {
		t.Error("expected onJoin to be set")
	}
}

func TestChatBotClient_OnPart(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnPart(func(channel, user string) {})
	if client.onPart == nil {
		t.Error("expected onPart to be set")
	}
}

func TestChatBotClient_OnRoomState(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnRoomState(func(state *RoomState) {})
	if client.onRoomState == nil {
		t.Error("expected onRoomState to be set")
	}
}

func TestChatBotClient_OnNotice(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnNotice(func(notice *Notice) {})
	if client.onNotice == nil {
		t.Error("expected onNotice to be set")
	}
}

func TestChatBotClient_OnClearChat(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnClearChat(func(clear *ClearChat) {})
	if client.onClearChat == nil {
		t.Error("expected onClearChat to be set")
	}
}

func TestChatBotClient_OnWhisper(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnWhisper(func(whisper *Whisper) {})
	if client.onWhisper == nil {
		t.Error("expected onWhisper to be set")
	}
}

func TestChatBotClient_OnConnect(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnConnect(func() {})
	if client.onConnect == nil {
		t.Error("expected onConnect to be set")
	}
}

func TestChatBotClient_OnDisconnect(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnDisconnect(func() {})
	if client.onDisconnect == nil {
		t.Error("expected onDisconnect to be set")
	}
}

func TestChatBotClient_OnError(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	client.OnError(func(err error) {})
	if client.onError == nil {
		t.Error("expected onError to be set")
	}
}

func TestChatBotClient_handleMessage(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *ChatMessage
	client.OnMessage(func(msg *ChatMessage) {
		received = msg
	})

	// Simulate a real Twitch PRIVMSG parsed message
	msg := &ChatMessage{
		ID:          "b34ccfc7-4977-403a-8a94-33c6bac34fb8",
		Channel:     "dallas",
		UserID:      "12345678",
		User:        "ronni",
		DisplayName: "ronni",
		Message:     "Kappa Keepo Kappa",
		Color:       "#0000FF",
		Badges:      map[string]string{"broadcaster": "1"},
		Emotes: []IRCEmote{
			{ID: "25", Name: "Kappa", Start: 0, End: 4},
			{ID: "25", Name: "Kappa", Start: 12, End: 16},
			{ID: "1902", Name: "Keepo", Start: 6, End: 10},
		},
		Timestamp: time.Unix(1507246572, 675000000),
	}
	client.handleMessage(msg)

	if received == nil || received.Message != "Kappa Keepo Kappa" {
		t.Error("handleMessage did not call onMessage handler")
	}
	if received.ID != "b34ccfc7-4977-403a-8a94-33c6bac34fb8" {
		t.Error("message ID not preserved")
	}
}

func TestChatBotClient_handleMessage_WithCheer(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var cheerReceived *ChatMessage
	var messageReceived *ChatMessage
	client.OnCheer(func(msg *ChatMessage) {
		cheerReceived = msg
	})
	client.OnMessage(func(msg *ChatMessage) {
		messageReceived = msg
	})

	// Simulate a real Twitch cheer message
	msg := &ChatMessage{
		ID:          "b34ccfc7-4977-403a-8a94-33c6bac34fb8",
		Channel:     "dallas",
		UserID:      "12345678",
		User:        "ronni",
		DisplayName: "ronni",
		Message:     "cheer100",
		Bits:        100,
		Badges:      map[string]string{"staff": "1", "bits": "1000"},
		Timestamp:   time.Unix(1507246572, 675000000),
	}
	client.handleMessage(msg)

	if cheerReceived == nil || cheerReceived.Bits != 100 {
		t.Error("handleMessage did not call onCheer handler")
	}
	if messageReceived == nil {
		t.Error("handleMessage did not call onMessage handler")
	}
}

func TestChatBotClient_handleMessage_NoHandlers(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handlers - use realistic Twitch data
	client.handleMessage(&ChatMessage{
		ID:          "b34ccfc7-4977-403a-8a94-33c6bac34fb8",
		Channel:     "dallas",
		UserID:      "12345678",
		User:        "ronni",
		DisplayName: "ronni",
		Message:     "cheer100",
		Bits:        100,
		Timestamp:   time.Unix(1507246572, 675000000),
	})
}

func TestChatBotClient_handleUserNotice_Sub(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *UserNotice
	client.OnSub(func(notice *UserNotice) {
		received = notice
	})

	// Simulate a real Twitch subscription USERNOTICE
	notice := &UserNotice{
		Type:          UserNoticeTypeSub,
		Channel:       "dallas",
		UserID:        "87654321",
		User:          "ronni",
		DisplayName:   "ronni",
		SystemMessage: "ronni subscribed with Prime.",
		MsgParams: map[string]string{
			"msg-param-sub-plan":      "Prime",
			"msg-param-sub-plan-name": "Prime",
		},
		Timestamp: time.Unix(1507246572, 675000000),
	}
	client.handleUserNotice(notice)

	if received == nil {
		t.Error("handleUserNotice did not call onSub handler")
	}
	if received.MsgParams["msg-param-sub-plan"] != "Prime" {
		t.Error("subscription plan not preserved")
	}
}

func TestChatBotClient_handleUserNotice_Resub(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *UserNotice
	client.OnResub(func(notice *UserNotice) {
		received = notice
	})

	// Simulate a real Twitch resub USERNOTICE
	notice := &UserNotice{
		Type:          UserNoticeTypeResub,
		Channel:       "dallas",
		UserID:        "11111111",
		User:          "coolgamer",
		DisplayName:   "coolgamer",
		Message:       "Thank you for the stream!",
		SystemMessage: "coolgamer subscribed at Tier 1. They've subscribed for 36 months, currently on a 36 month streak!",
		MsgParams: map[string]string{
			"msg-param-cumulative-months": "36",
			"msg-param-streak-months":     "36",
			"msg-param-sub-plan":          "1000",
			"msg-param-sub-plan-name":     "Channel Subscription (streamer)",
		},
		Timestamp: time.Unix(1507246572, 675000000),
	}
	client.handleUserNotice(notice)

	if received == nil {
		t.Error("handleUserNotice did not call onResub handler")
	}
	if received.MsgParams["msg-param-cumulative-months"] != "36" {
		t.Error("cumulative months not preserved")
	}
}

func TestChatBotClient_handleUserNotice_SubGift(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *UserNotice
	client.OnSubGift(func(notice *UserNotice) {
		received = notice
	})

	// Test SubGift - real Twitch data format
	notice := &UserNotice{
		Type:          UserNoticeTypeSubGift,
		Channel:       "dallas",
		UserID:        "33333333",
		User:          "generousgifter",
		DisplayName:   "generousgifter",
		SystemMessage: "generousgifter gifted a Tier 1 sub to luckyviewer!",
		MsgParams: map[string]string{
			"msg-param-recipient-id":           "22222222",
			"msg-param-recipient-user-name":    "luckyviewer",
			"msg-param-recipient-display-name": "luckyviewer",
			"msg-param-sub-plan":               "1000",
			"msg-param-sub-plan-name":          "Channel Subscription (streamer)",
			"msg-param-sender-count":           "100",
		},
		Timestamp: time.Unix(1507246572, 675000000),
	}
	client.handleUserNotice(notice)
	if received == nil {
		t.Error("handleUserNotice did not call onSubGift handler for SubGift")
	}

	// Test AnonSubGift
	received = nil
	notice = &UserNotice{
		Type:          UserNoticeTypeAnonSubGift,
		Channel:       "dallas",
		SystemMessage: "An anonymous user gifted a Tier 1 sub to luckyviewer!",
		MsgParams: map[string]string{
			"msg-param-recipient-id":           "22222222",
			"msg-param-recipient-user-name":    "luckyviewer",
			"msg-param-recipient-display-name": "luckyviewer",
			"msg-param-sub-plan":               "1000",
		},
	}
	client.handleUserNotice(notice)
	if received == nil {
		t.Error("handleUserNotice did not call onSubGift handler for AnonSubGift")
	}

	// Test SubMysteryGift
	received = nil
	notice = &UserNotice{
		Type:          UserNoticeTypeSubMysteryGift,
		Channel:       "dallas",
		UserID:        "33333333",
		User:          "generousgifter",
		DisplayName:   "generousgifter",
		SystemMessage: "generousgifter is gifting 5 Tier 1 Subs to dallas's community!",
		MsgParams: map[string]string{
			"msg-param-mass-gift-count": "5",
			"msg-param-sub-plan":        "1000",
		},
	}
	client.handleUserNotice(notice)
	if received == nil {
		t.Error("handleUserNotice did not call onSubGift handler for SubMysteryGift")
	}
}

func TestChatBotClient_handleUserNotice_Raid(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *UserNotice
	client.OnRaid(func(notice *UserNotice) {
		received = notice
	})

	// Simulate a real Twitch raid USERNOTICE
	notice := &UserNotice{
		Type:          UserNoticeTypeRaid,
		Channel:       "dallas",
		UserID:        "13405587",
		User:          "tww2",
		DisplayName:   "TWW2",
		SystemMessage: "15067 raiders from TWW2 have joined!",
		MsgParams: map[string]string{
			"msg-param-viewerCount":  "15067",
			"msg-param-displayName":  "TWW2",
			"msg-param-login":        "tww2",
		},
		Timestamp: time.Unix(1507246572, 675000000),
	}
	client.handleUserNotice(notice)

	if received == nil {
		t.Error("handleUserNotice did not call onRaid handler")
	}
	if received.MsgParams["msg-param-viewerCount"] != "15067" {
		t.Error("viewer count not preserved")
	}
}

func TestChatBotClient_handleUserNotice_NoHandlers(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handlers - use realistic data
	client.handleUserNotice(&UserNotice{Type: UserNoticeTypeSub, Channel: "dallas"})
	client.handleUserNotice(&UserNotice{Type: UserNoticeTypeResub, Channel: "dallas"})
	client.handleUserNotice(&UserNotice{Type: UserNoticeTypeSubGift, Channel: "dallas"})
	client.handleUserNotice(&UserNotice{Type: UserNoticeTypeRaid, Channel: "dallas"})
}

func TestChatBotClient_handleJoin(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var receivedChannel, receivedUser string
	client.OnJoin(func(channel, user string) {
		receivedChannel = channel
		receivedUser = user
	})

	// Simulate Twitch JOIN event
	client.handleJoin("dallas", "ronni")

	if receivedChannel != "dallas" || receivedUser != "ronni" {
		t.Error("handleJoin did not call onJoin handler correctly")
	}
}

func TestChatBotClient_handleJoin_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleJoin("dallas", "ronni")
}

func TestChatBotClient_handlePart(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var receivedChannel, receivedUser string
	client.OnPart(func(channel, user string) {
		receivedChannel = channel
		receivedUser = user
	})

	// Simulate Twitch PART event
	client.handlePart("dallas", "ronni")

	if receivedChannel != "dallas" || receivedUser != "ronni" {
		t.Error("handlePart did not call onPart handler correctly")
	}
}

func TestChatBotClient_handlePart_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handlePart("dallas", "ronni")
}

func TestChatBotClient_handleRoomState(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *RoomState
	client.OnRoomState(func(state *RoomState) {
		received = state
	})

	// Simulate real Twitch ROOMSTATE
	state := &RoomState{
		Channel:       "dallas",
		RoomID:        "12345678",
		EmoteOnly:     false,
		FollowersOnly: -1,
		R9K:           false,
		Slow:          0,
		SubsOnly:      false,
	}
	client.handleRoomState(state)

	if received == nil || received.Channel != "dallas" {
		t.Error("handleRoomState did not call onRoomState handler correctly")
	}
	if received.RoomID != "12345678" {
		t.Error("room ID not preserved")
	}
}

func TestChatBotClient_handleRoomState_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleRoomState(&RoomState{Channel: "dallas", RoomID: "12345678"})
}

func TestChatBotClient_handleNotice(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *Notice
	client.OnNotice(func(notice *Notice) {
		received = notice
	})

	// Simulate real Twitch NOTICE
	notice := &Notice{
		Channel: "dallas",
		MsgID:   "slow_off",
		Message: "This room is no longer in slow mode.",
	}
	client.handleNotice(notice)

	if received == nil || received.Message != "This room is no longer in slow mode." {
		t.Error("handleNotice did not call onNotice handler correctly")
	}
	if received.MsgID != "slow_off" {
		t.Error("msg-id not preserved")
	}
}

func TestChatBotClient_handleNotice_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleNotice(&Notice{Channel: "dallas", MsgID: "slow_off"})
}

func TestChatBotClient_handleClearChat(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *ClearChat
	client.OnClearChat(func(clear *ClearChat) {
		received = clear
	})

	// Simulate real Twitch CLEARCHAT (timeout)
	clear := &ClearChat{
		Channel:      "dallas",
		TargetUserID: "87654321",
		User:         "ronni",
		BanDuration:  350,
		RoomID:       "12345678",
		Timestamp:    time.Unix(1642715756, 806000000),
	}
	client.handleClearChat(clear)

	if received == nil || received.Channel != "dallas" {
		t.Error("handleClearChat did not call onClearChat handler correctly")
	}
	if received.BanDuration != 350 {
		t.Error("ban duration not preserved")
	}
}

func TestChatBotClient_handleClearChat_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleClearChat(&ClearChat{Channel: "dallas", RoomID: "12345678"})
}

func TestChatBotClient_handleWhisper(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received *Whisper
	client.OnWhisper(func(whisper *Whisper) {
		received = whisper
	})

	// Simulate real Twitch WHISPER
	whisper := &Whisper{
		FromID:      "87654321",
		From:        "petsgomoo",
		To:          "dallas",
		DisplayName: "PetsgomOO",
		Message:     "hello",
		Color:       "#8A2BE2",
		Badges:      map[string]string{"staff": "1", "bits-charity": "1"},
		MessageID:   "306",
		ThreadID:    "12345678_87654321",
	}
	client.handleWhisper(whisper)

	if received == nil || received.Message != "hello" {
		t.Error("handleWhisper did not call onWhisper handler correctly")
	}
	if received.From != "petsgomoo" {
		t.Error("sender not preserved")
	}
}

func TestChatBotClient_handleWhisper_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleWhisper(&Whisper{From: "petsgomoo", To: "dallas", Message: "hello"})
}

func TestChatBotClient_handleConnect(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	called := false
	client.OnConnect(func() {
		called = true
	})

	client.handleConnect()

	if !called {
		t.Error("handleConnect did not call onConnect handler")
	}
}

func TestChatBotClient_handleConnect_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleConnect()
}

func TestChatBotClient_handleDisconnect(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	called := false
	client.OnDisconnect(func() {
		called = true
	})

	client.handleDisconnect()

	if !called {
		t.Error("handleDisconnect did not call onDisconnect handler")
	}
}

func TestChatBotClient_handleDisconnect_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleDisconnect()
}

func TestChatBotClient_handleError(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var received error
	client.OnError(func(err error) {
		received = err
	})

	testErr := ErrIRCNotConnected
	client.handleError(testErr)

	if received != testErr {
		t.Error("handleError did not call onError handler correctly")
	}
}

func TestChatBotClient_handleError_NoHandler(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)
	// Should not panic with nil handler
	client.handleError(ErrIRCNotConnected)
}

func TestChatBotClient_ConcurrentHandlerAccess(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil)

	var wg sync.WaitGroup
	wg.Add(4)

	// Concurrent writes
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			client.OnMessage(func(msg *ChatMessage) {})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			client.OnJoin(func(channel, user string) {})
		}
	}()

	// Concurrent reads via handlers - use realistic Twitch data
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			client.handleMessage(&ChatMessage{
				ID:          "b34ccfc7-4977-403a-8a94-33c6bac34fb8",
				Channel:     "dallas",
				UserID:      "12345678",
				User:        "ronni",
				DisplayName: "ronni",
				Message:     "Kappa",
			})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			client.handleJoin("dallas", "ronni")
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for concurrent operations")
	}
}

func TestChatBotClient_Connect_WithMockServer(t *testing.T) {
	// Create mock IRC server simulating Twitch TMI
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		// Read CAP REQ
		_, _, _ = conn.ReadMessage()
		// Read PASS
		_, _, _ = conn.ReadMessage()
		// Read NICK
		_, _, _ = conn.ReadMessage()
		// Send Twitch welcome sequence
		_ = conn.WriteMessage(websocket.TextMessage, []byte(twitchCapAck))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(twitchWelcome))
		// Keep connection open briefly
		time.Sleep(50 * time.Millisecond)
	})
	defer mock.Close()

	// Create auth client with token (Twitch OAuth format)
	authClient := &AuthClient{}
	authClient.token = &Token{AccessToken: "oauth:abcdefghijklmnop123456789"}

	// Use WithChatBotURL to inject mock server URL - tests full Connect path
	client := NewChatBotClient("justinfan12345", authClient, WithChatBotURL(mock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Call Connect through ChatBotClient (tests the full path including c.irc.Connect)
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = client.Close() }()

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}

	// Test IRC accessor
	if client.IRC() == nil {
		t.Error("expected IRC() to return non-nil")
	}
}

func TestWithChatBotURL(t *testing.T) {
	client := NewChatBotClient("justinfan12345", nil, WithChatBotURL("wss://irc-ws.chat.twitch.tv:443"))
	if client.ircURL != "wss://irc-ws.chat.twitch.tv:443" {
		t.Errorf("expected ircURL to be set, got %s", client.ircURL)
	}
}

func TestChatBotClient_OperationsWithMockConnection(t *testing.T) {
	// Create mock IRC server that accepts messages (simulating Twitch TMI)
	mock := newMockIRCServer(func(conn *websocket.Conn) {
		defer func() { _ = conn.Close() }()
		// Read CAP REQ, PASS, NICK
		_, _, _ = conn.ReadMessage()
		_, _, _ = conn.ReadMessage()
		_, _, _ = conn.ReadMessage()
		// Send Twitch welcome
		_ = conn.WriteMessage(websocket.TextMessage, []byte(twitchCapAck))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(twitchWelcome))
		// Read any additional messages (JOIN, PRIVMSG, etc.)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})
	defer mock.Close()

	client := NewChatBotClient("justinfan12345", nil)
	client.irc = NewIRCClient("justinfan12345", "oauth:abcdefghijklmnop123456789",
		WithIRCURL(mock.URL()),
		WithAutoReconnect(false),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.irc.Connect(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Test Join (Twitch channel format)
	err = client.Join("#dallas")
	if err != nil {
		t.Errorf("Join failed: %v", err)
	}

	// Test Part
	err = client.Part("#dallas")
	if err != nil {
		t.Errorf("Part failed: %v", err)
	}

	// Test Say
	err = client.Say("#dallas", "Hello chat! Kappa")
	if err != nil {
		t.Errorf("Say failed: %v", err)
	}

	// Test Reply (using real Twitch message ID format)
	err = client.Reply("#dallas", "b34ccfc7-4977-403a-8a94-33c6bac34fb8", "Thanks for the follow!")
	if err != nil {
		t.Errorf("Reply failed: %v", err)
	}

	// Test Whisper
	err = client.Whisper("ronni", "Hey, check out this new emote!")
	if err != nil {
		t.Errorf("Whisper failed: %v", err)
	}

	// Test GetJoinedChannels
	channels := client.GetJoinedChannels()
	if channels == nil {
		t.Error("GetJoinedChannels returned nil")
	}
}
