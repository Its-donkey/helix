package irc

import (
	"context"
	"sync"
)

// Bot provides a high-level interface for Twitch chat bots.
// It wraps Client with convenience methods and event routing.
type Bot struct {
	client *Client
	nick   string
	token  string

	// Client options
	url           string
	autoReconnect *bool

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

// BotOption configures the Bot.
type BotOption func(*Bot)

// WithBotURL sets a custom WebSocket URL for the bot.
func WithBotURL(url string) BotOption {
	return func(b *Bot) {
		b.url = url
	}
}

// WithBotAutoReconnect enables or disables auto-reconnect for the bot.
func WithBotAutoReconnect(enabled bool) BotOption {
	return func(b *Bot) {
		b.autoReconnect = &enabled
	}
}

// NewBot creates a new high-level chat bot.
// The nick should be the bot's username, and token should be a valid OAuth access token.
func NewBot(nick, token string, opts ...BotOption) *Bot {
	b := &Bot{
		nick:  nick,
		token: token,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// OnMessage sets the handler for all chat messages.
func (b *Bot) OnMessage(fn func(*ChatMessage)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onMessage = fn
}

// OnSub sets the handler for new subscription events.
func (b *Bot) OnSub(fn func(*UserNotice)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onSub = fn
}

// OnResub sets the handler for resubscription events.
func (b *Bot) OnResub(fn func(*UserNotice)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onResub = fn
}

// OnSubGift sets the handler for gift subscription events.
func (b *Bot) OnSubGift(fn func(*UserNotice)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onSubGift = fn
}

// OnRaid sets the handler for raid events.
func (b *Bot) OnRaid(fn func(*UserNotice)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onRaid = fn
}

// OnCheer sets the handler for cheer (bits) messages.
func (b *Bot) OnCheer(fn func(*ChatMessage)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onCheer = fn
}

// OnJoin sets the handler for user join events.
func (b *Bot) OnJoin(fn func(channel, user string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onJoin = fn
}

// OnPart sets the handler for user part events.
func (b *Bot) OnPart(fn func(channel, user string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onPart = fn
}

// OnRoomState sets the handler for room state changes.
func (b *Bot) OnRoomState(fn func(*RoomState)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onRoomState = fn
}

// OnNotice sets the handler for server notices.
func (b *Bot) OnNotice(fn func(*Notice)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onNotice = fn
}

// OnClearChat sets the handler for chat clear/timeout/ban events.
func (b *Bot) OnClearChat(fn func(*ClearChat)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onClearChat = fn
}

// OnWhisper sets the handler for whisper messages.
func (b *Bot) OnWhisper(fn func(*Whisper)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onWhisper = fn
}

// OnConnect sets the handler for successful connections.
func (b *Bot) OnConnect(fn func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onConnect = fn
}

// OnDisconnect sets the handler for disconnections.
func (b *Bot) OnDisconnect(fn func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onDisconnect = fn
}

// OnError sets the handler for errors.
func (b *Bot) OnError(fn func(error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onError = fn
}

// Connect establishes a connection to Twitch chat.
func (b *Bot) Connect(ctx context.Context) error {
	opts := []Option{
		WithMessageHandler(b.handleMessage),
		WithUserNoticeHandler(b.handleUserNotice),
		WithJoinHandler(b.handleJoin),
		WithPartHandler(b.handlePart),
		WithRoomStateHandler(b.handleRoomState),
		WithNoticeHandler(b.handleNotice),
		WithClearChatHandler(b.handleClearChat),
		WithWhisperHandler(b.handleWhisper),
		WithConnectHandler(b.handleConnect),
		WithDisconnectHandler(b.handleDisconnect),
		WithErrorHandler(b.handleError),
	}

	if b.url != "" {
		opts = append(opts, WithURL(b.url))
	}
	if b.autoReconnect != nil {
		opts = append(opts, WithAutoReconnect(*b.autoReconnect))
	}

	b.client = NewClient(b.nick, b.token, opts...)

	return b.client.Connect(ctx)
}

// Close closes the chat connection.
func (b *Bot) Close() error {
	if b.client != nil {
		return b.client.Close()
	}
	return nil
}

// IsConnected returns whether the bot is connected.
func (b *Bot) IsConnected() bool {
	if b.client == nil {
		return false
	}
	return b.client.IsConnected()
}

// Join joins one or more channels.
func (b *Bot) Join(channels ...string) error {
	if b.client == nil {
		return ErrNotConnected
	}
	return b.client.Join(channels...)
}

// Part leaves one or more channels.
func (b *Bot) Part(channels ...string) error {
	if b.client == nil {
		return ErrNotConnected
	}
	return b.client.Part(channels...)
}

// Say sends a message to a channel.
func (b *Bot) Say(channel, message string) error {
	if b.client == nil {
		return ErrNotConnected
	}
	return b.client.Say(channel, message)
}

// Reply sends a reply to a specific message.
func (b *Bot) Reply(channel, parentMsgID, message string) error {
	if b.client == nil {
		return ErrNotConnected
	}
	return b.client.Reply(channel, parentMsgID, message)
}

// Whisper sends a whisper to a user.
func (b *Bot) Whisper(user, message string) error {
	if b.client == nil {
		return ErrNotConnected
	}
	return b.client.Whisper(user, message)
}

// GetJoinedChannels returns the list of joined channels.
func (b *Bot) GetJoinedChannels() []string {
	if b.client == nil {
		return nil
	}
	return b.client.GetJoinedChannels()
}

// Client returns the underlying IRC client for advanced usage.
func (b *Bot) Client() *Client {
	return b.client
}

// Internal handlers

func (b *Bot) handleMessage(msg *ChatMessage) {
	b.mu.RLock()
	onMessage := b.onMessage
	onCheer := b.onCheer
	b.mu.RUnlock()

	// Check for cheers
	if msg.Bits > 0 && onCheer != nil {
		onCheer(msg)
	}

	if onMessage != nil {
		onMessage(msg)
	}
}

func (b *Bot) handleUserNotice(notice *UserNotice) {
	b.mu.RLock()
	onSub := b.onSub
	onResub := b.onResub
	onSubGift := b.onSubGift
	onRaid := b.onRaid
	b.mu.RUnlock()

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

func (b *Bot) handleJoin(channel, user string) {
	b.mu.RLock()
	fn := b.onJoin
	b.mu.RUnlock()

	if fn != nil {
		fn(channel, user)
	}
}

func (b *Bot) handlePart(channel, user string) {
	b.mu.RLock()
	fn := b.onPart
	b.mu.RUnlock()

	if fn != nil {
		fn(channel, user)
	}
}

func (b *Bot) handleRoomState(state *RoomState) {
	b.mu.RLock()
	fn := b.onRoomState
	b.mu.RUnlock()

	if fn != nil {
		fn(state)
	}
}

func (b *Bot) handleNotice(notice *Notice) {
	b.mu.RLock()
	fn := b.onNotice
	b.mu.RUnlock()

	if fn != nil {
		fn(notice)
	}
}

func (b *Bot) handleClearChat(clear *ClearChat) {
	b.mu.RLock()
	fn := b.onClearChat
	b.mu.RUnlock()

	if fn != nil {
		fn(clear)
	}
}

func (b *Bot) handleWhisper(whisper *Whisper) {
	b.mu.RLock()
	fn := b.onWhisper
	b.mu.RUnlock()

	if fn != nil {
		fn(whisper)
	}
}

func (b *Bot) handleConnect() {
	b.mu.RLock()
	fn := b.onConnect
	b.mu.RUnlock()

	if fn != nil {
		fn()
	}
}

func (b *Bot) handleDisconnect() {
	b.mu.RLock()
	fn := b.onDisconnect
	b.mu.RUnlock()

	if fn != nil {
		fn()
	}
}

func (b *Bot) handleError(err error) {
	b.mu.RLock()
	fn := b.onError
	b.mu.RUnlock()

	if fn != nil {
		fn(err)
	}
}
