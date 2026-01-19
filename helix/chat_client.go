package helix

import (
	"context"
	"errors"
	"sync"
)

// ChatBotClient provides a high-level interface for Twitch chat bots.
// It wraps IRCClient with convenience methods and automatic token handling.
type ChatBotClient struct {
	irc        *IRCClient
	authClient *AuthClient
	nick       string
	ircURL     string // custom IRC URL for testing

	// Event handlers
	onMessage    func(*ChatMessage)
	onSub        func(*UserNotice)
	onResub      func(*UserNotice)
	onSubGift    func(*UserNotice)
	onRaid       func(*UserNotice)
	onCheer      func(*ChatMessage)
	onJoin       func(channel, user string)
	onPart       func(channel, user string)
	onRoomState  func(*RoomState)
	onNotice     func(*Notice)
	onClearChat  func(*ClearChat)
	onWhisper    func(*Whisper)
	onConnect    func()
	onDisconnect func()
	onError      func(error)

	mu sync.RWMutex
}

// ChatBotOption configures the ChatBotClient.
type ChatBotOption func(*ChatBotClient)

// NewChatBotClient creates a new high-level chat bot client.
// The nick should be the bot's username, and authClient should have a valid user access token.
func NewChatBotClient(nick string, authClient *AuthClient, opts ...ChatBotOption) *ChatBotClient {
	c := &ChatBotClient{
		authClient: authClient,
		nick:       nick,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithChatBotURL sets a custom IRC WebSocket URL (for testing).
func WithChatBotURL(url string) ChatBotOption {
	return func(c *ChatBotClient) {
		c.ircURL = url
	}
}

// OnMessage sets the handler for all chat messages.
func (c *ChatBotClient) OnMessage(fn func(*ChatMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMessage = fn
}

// OnSub sets the handler for new subscription events.
func (c *ChatBotClient) OnSub(fn func(*UserNotice)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onSub = fn
}

// OnResub sets the handler for resubscription events.
func (c *ChatBotClient) OnResub(fn func(*UserNotice)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onResub = fn
}

// OnSubGift sets the handler for gift subscription events.
func (c *ChatBotClient) OnSubGift(fn func(*UserNotice)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onSubGift = fn
}

// OnRaid sets the handler for raid events.
func (c *ChatBotClient) OnRaid(fn func(*UserNotice)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRaid = fn
}

// OnCheer sets the handler for cheer (bits) messages.
func (c *ChatBotClient) OnCheer(fn func(*ChatMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onCheer = fn
}

// OnJoin sets the handler for user join events.
func (c *ChatBotClient) OnJoin(fn func(channel, user string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onJoin = fn
}

// OnPart sets the handler for user part events.
func (c *ChatBotClient) OnPart(fn func(channel, user string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onPart = fn
}

// OnRoomState sets the handler for room state changes.
func (c *ChatBotClient) OnRoomState(fn func(*RoomState)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRoomState = fn
}

// OnNotice sets the handler for server notices.
func (c *ChatBotClient) OnNotice(fn func(*Notice)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onNotice = fn
}

// OnClearChat sets the handler for chat clear/timeout/ban events.
func (c *ChatBotClient) OnClearChat(fn func(*ClearChat)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onClearChat = fn
}

// OnWhisper sets the handler for whisper messages.
func (c *ChatBotClient) OnWhisper(fn func(*Whisper)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onWhisper = fn
}

// OnConnect sets the handler for successful connections.
func (c *ChatBotClient) OnConnect(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onConnect = fn
}

// OnDisconnect sets the handler for disconnections.
func (c *ChatBotClient) OnDisconnect(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onDisconnect = fn
}

// OnError sets the handler for errors.
func (c *ChatBotClient) OnError(fn func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onError = fn
}

// Connect establishes a connection to Twitch chat.
func (c *ChatBotClient) Connect(ctx context.Context) error {
	token := ""
	if c.authClient != nil {
		if t := c.authClient.GetToken(); t != nil {
			token = t.AccessToken
		}
	}

	if token == "" {
		return errors.New("chatbot: no authentication token available")
	}

	ircOpts := []IRCOption{
		WithMessageHandler(c.handleMessage),
		WithUserNoticeHandler(c.handleUserNotice),
		WithJoinHandler(c.handleJoin),
		WithPartHandler(c.handlePart),
		WithRoomStateHandler(c.handleRoomState),
		WithNoticeHandler(c.handleNotice),
		WithClearChatHandler(c.handleClearChat),
		WithWhisperHandler(c.handleWhisper),
		WithConnectHandler(c.handleConnect),
		WithDisconnectHandler(c.handleDisconnect),
		WithIRCErrorHandler(c.handleError),
	}
	if c.ircURL != "" {
		ircOpts = append(ircOpts, WithIRCURL(c.ircURL))
	}
	c.irc = NewIRCClient(c.nick, token, ircOpts...)

	if c.irc == nil {
		return errors.New("chatbot: failed to create IRC client (invalid nick or token)")
	}

	return c.irc.Connect(ctx)
}

// Close closes the chat connection.
func (c *ChatBotClient) Close() error {
	if c.irc != nil {
		return c.irc.Close()
	}
	return nil
}

// IsConnected returns whether the client is connected.
func (c *ChatBotClient) IsConnected() bool {
	if c.irc == nil {
		return false
	}
	return c.irc.IsConnected()
}

// Join joins one or more channels.
func (c *ChatBotClient) Join(channels ...string) error {
	if c.irc == nil {
		return ErrIRCNotConnected
	}
	return c.irc.Join(channels...)
}

// Part leaves one or more channels.
func (c *ChatBotClient) Part(channels ...string) error {
	if c.irc == nil {
		return ErrIRCNotConnected
	}
	return c.irc.Part(channels...)
}

// Say sends a message to a channel.
func (c *ChatBotClient) Say(channel, message string) error {
	if c.irc == nil {
		return ErrIRCNotConnected
	}
	return c.irc.Say(channel, message)
}

// Reply sends a reply to a specific message.
func (c *ChatBotClient) Reply(channel, parentMsgID, message string) error {
	if c.irc == nil {
		return ErrIRCNotConnected
	}
	return c.irc.Reply(channel, parentMsgID, message)
}

// Whisper sends a whisper to a user.
func (c *ChatBotClient) Whisper(user, message string) error {
	if c.irc == nil {
		return ErrIRCNotConnected
	}
	return c.irc.Whisper(user, message)
}

// GetJoinedChannels returns the list of joined channels.
func (c *ChatBotClient) GetJoinedChannels() []string {
	if c.irc == nil {
		return nil
	}
	return c.irc.GetJoinedChannels()
}

// IRC returns the underlying IRC client for advanced usage.
func (c *ChatBotClient) IRC() *IRCClient {
	return c.irc
}

// Internal handlers

func (c *ChatBotClient) handleMessage(msg *ChatMessage) {
	c.mu.RLock()
	onMessage := c.onMessage
	onCheer := c.onCheer
	c.mu.RUnlock()

	// Check for cheers
	if msg.Bits > 0 && onCheer != nil {
		onCheer(msg)
	}

	if onMessage != nil {
		onMessage(msg)
	}
}

func (c *ChatBotClient) handleUserNotice(notice *UserNotice) {
	c.mu.RLock()
	onSub := c.onSub
	onResub := c.onResub
	onSubGift := c.onSubGift
	onRaid := c.onRaid
	c.mu.RUnlock()

	switch notice.Type {
	case UserNoticeTypeSub:
		if onSub != nil {
			onSub(notice)
		}
	case UserNoticeTypeResub:
		if onResub != nil {
			onResub(notice)
		}
	case UserNoticeTypeSubGift, UserNoticeTypeAnonSubGift, UserNoticeTypeSubMysteryGift:
		if onSubGift != nil {
			onSubGift(notice)
		}
	case UserNoticeTypeRaid:
		if onRaid != nil {
			onRaid(notice)
		}
	}
}

func (c *ChatBotClient) handleJoin(channel, user string) {
	c.mu.RLock()
	fn := c.onJoin
	c.mu.RUnlock()

	if fn != nil {
		fn(channel, user)
	}
}

func (c *ChatBotClient) handlePart(channel, user string) {
	c.mu.RLock()
	fn := c.onPart
	c.mu.RUnlock()

	if fn != nil {
		fn(channel, user)
	}
}

func (c *ChatBotClient) handleRoomState(state *RoomState) {
	c.mu.RLock()
	fn := c.onRoomState
	c.mu.RUnlock()

	if fn != nil {
		fn(state)
	}
}

func (c *ChatBotClient) handleNotice(notice *Notice) {
	c.mu.RLock()
	fn := c.onNotice
	c.mu.RUnlock()

	if fn != nil {
		fn(notice)
	}
}

func (c *ChatBotClient) handleClearChat(clear *ClearChat) {
	c.mu.RLock()
	fn := c.onClearChat
	c.mu.RUnlock()

	if fn != nil {
		fn(clear)
	}
}

func (c *ChatBotClient) handleWhisper(whisper *Whisper) {
	c.mu.RLock()
	fn := c.onWhisper
	c.mu.RUnlock()

	if fn != nil {
		fn(whisper)
	}
}

func (c *ChatBotClient) handleConnect() {
	c.mu.RLock()
	fn := c.onConnect
	c.mu.RUnlock()

	if fn != nil {
		fn()
	}
}

func (c *ChatBotClient) handleDisconnect() {
	c.mu.RLock()
	fn := c.onDisconnect
	c.mu.RUnlock()

	if fn != nil {
		fn()
	}
}

func (c *ChatBotClient) handleError(err error) {
	c.mu.RLock()
	fn := c.onError
	c.mu.RUnlock()

	if fn != nil {
		fn(err)
	}
}
