package helix

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// TwitchIRCWebSocket is the WebSocket URL for Twitch IRC.
	TwitchIRCWebSocket = "wss://irc-ws.chat.twitch.tv:443"

	// TwitchIRCTCP is the TCP address for Twitch IRC.
	TwitchIRCTCP = "irc.chat.twitch.tv:6697"
)

// IRC command constants
const (
	ircCAP             = "CAP"
	ircJOIN            = "JOIN"
	ircPART            = "PART"
	ircPRIVMSG         = "PRIVMSG"
	ircWHISPER         = "WHISPER"
	ircPING            = "PING"
	ircPONG            = "PONG"
	ircNOTICE          = "NOTICE"
	ircUSERNOTICE      = "USERNOTICE"
	ircROOMSTATE       = "ROOMSTATE"
	ircCLEARCHAT       = "CLEARCHAT"
	ircCLEARMSG        = "CLEARMSG"
	ircGLOBALUSERSTATE = "GLOBALUSERSTATE"
	ircUSERSTATE       = "USERSTATE"
	ircRECONNECT       = "RECONNECT"
)

// IRC errors
var (
	ErrIRCNotConnected     = errors.New("irc: not connected")
	ErrIRCAlreadyConnected = errors.New("irc: already connected")
	ErrIRCAuthFailed       = errors.New("irc: authentication failed")
	ErrIRCInvalidNick      = errors.New("irc: nick is required")
	ErrIRCInvalidToken     = errors.New("irc: token is required")
)

// sanitizeIRCMessage removes CR/LF characters to prevent IRC command injection.
func sanitizeIRCMessage(msg string) string {
	msg = strings.ReplaceAll(msg, "\r", "")
	msg = strings.ReplaceAll(msg, "\n", "")
	return msg
}

// IRCClient manages a connection to Twitch IRC.
type IRCClient struct {
	url   string
	conn  *websocket.Conn
	nick  string
	token string

	// Channel tracking
	channels map[string]bool

	// Handlers
	onMessage         func(*ChatMessage)
	onJoin            func(channel, user string)
	onPart            func(channel, user string)
	onNotice          func(*Notice)
	onUserNotice      func(*UserNotice)
	onRoomState       func(*RoomState)
	onClearChat       func(*ClearChat)
	onClearMessage    func(*ClearMessage)
	onWhisper         func(*Whisper)
	onGlobalUserState func(*GlobalUserState)
	onUserState       func(*UserState)
	onError           func(error)
	onConnect         func()
	onDisconnect      func()
	onReconnect       func()
	onRawMessage      func(string)

	// State
	mu           sync.RWMutex
	connected    bool
	connecting   bool
	stopChan     chan struct{}
	stopOnce     sync.Once      // ensures stopChan is closed only once
	wg           sync.WaitGroup // tracks readLoop goroutine
	writeMu      sync.Mutex
	globalState  *GlobalUserState
	pongReceived chan struct{}

	// Options
	autoReconnect  bool
	reconnectDelay time.Duration
	capabilities   []string
}

// IRCOption configures the IRC client.
type IRCOption func(*IRCClient)

// NewIRCClient creates a new IRC client.
// Deprecated: Use NewIRCClientE instead which returns an error for invalid inputs.
// This function returns nil if nick or token is empty.
func NewIRCClient(nick, token string, opts ...IRCOption) *IRCClient {
	client, _ := NewIRCClientE(nick, token, opts...)
	return client
}

// NewIRCClientE creates a new IRC client with error handling.
// Returns an error if nick or token is empty.
func NewIRCClientE(nick, token string, opts ...IRCOption) (*IRCClient, error) {
	if nick == "" {
		return nil, ErrIRCInvalidNick
	}
	if token == "" {
		return nil, ErrIRCInvalidToken
	}

	// Ensure token has oauth: prefix
	if !strings.HasPrefix(token, "oauth:") {
		token = "oauth:" + token
	}

	c := &IRCClient{
		url:            TwitchIRCWebSocket,
		nick:           strings.ToLower(nick),
		token:          token,
		channels:       make(map[string]bool),
		autoReconnect:  true,
		reconnectDelay: 5 * time.Second,
		capabilities: []string{
			"twitch.tv/tags",
			"twitch.tv/commands",
			"twitch.tv/membership",
		},
		pongReceived: make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// WithIRCURL sets a custom WebSocket URL.
func WithIRCURL(url string) IRCOption {
	return func(c *IRCClient) {
		c.url = url
	}
}

// WithAutoReconnect enables or disables auto-reconnect.
func WithAutoReconnect(enabled bool) IRCOption {
	return func(c *IRCClient) {
		c.autoReconnect = enabled
	}
}

// WithReconnectDelay sets the delay between reconnection attempts.
func WithReconnectDelay(d time.Duration) IRCOption {
	return func(c *IRCClient) {
		c.reconnectDelay = d
	}
}

// WithMessageHandler sets the handler for chat messages.
func WithMessageHandler(fn func(*ChatMessage)) IRCOption {
	return func(c *IRCClient) {
		c.onMessage = fn
	}
}

// WithJoinHandler sets the handler for join events.
func WithJoinHandler(fn func(channel, user string)) IRCOption {
	return func(c *IRCClient) {
		c.onJoin = fn
	}
}

// WithPartHandler sets the handler for part events.
func WithPartHandler(fn func(channel, user string)) IRCOption {
	return func(c *IRCClient) {
		c.onPart = fn
	}
}

// WithNoticeHandler sets the handler for notice messages.
func WithNoticeHandler(fn func(*Notice)) IRCOption {
	return func(c *IRCClient) {
		c.onNotice = fn
	}
}

// WithUserNoticeHandler sets the handler for user notices (subs, raids, etc.).
func WithUserNoticeHandler(fn func(*UserNotice)) IRCOption {
	return func(c *IRCClient) {
		c.onUserNotice = fn
	}
}

// WithRoomStateHandler sets the handler for room state changes.
func WithRoomStateHandler(fn func(*RoomState)) IRCOption {
	return func(c *IRCClient) {
		c.onRoomState = fn
	}
}

// WithClearChatHandler sets the handler for clear chat events.
func WithClearChatHandler(fn func(*ClearChat)) IRCOption {
	return func(c *IRCClient) {
		c.onClearChat = fn
	}
}

// WithClearMessageHandler sets the handler for clear message events.
func WithClearMessageHandler(fn func(*ClearMessage)) IRCOption {
	return func(c *IRCClient) {
		c.onClearMessage = fn
	}
}

// WithWhisperHandler sets the handler for whisper messages.
func WithWhisperHandler(fn func(*Whisper)) IRCOption {
	return func(c *IRCClient) {
		c.onWhisper = fn
	}
}

// WithGlobalUserStateHandler sets the handler for global user state.
func WithGlobalUserStateHandler(fn func(*GlobalUserState)) IRCOption {
	return func(c *IRCClient) {
		c.onGlobalUserState = fn
	}
}

// WithUserStateHandler sets the handler for user state.
func WithUserStateHandler(fn func(*UserState)) IRCOption {
	return func(c *IRCClient) {
		c.onUserState = fn
	}
}

// WithIRCErrorHandler sets the handler for errors.
func WithIRCErrorHandler(fn func(error)) IRCOption {
	return func(c *IRCClient) {
		c.onError = fn
	}
}

// WithConnectHandler sets the handler for successful connections.
func WithConnectHandler(fn func()) IRCOption {
	return func(c *IRCClient) {
		c.onConnect = fn
	}
}

// WithDisconnectHandler sets the handler for disconnections.
func WithDisconnectHandler(fn func()) IRCOption {
	return func(c *IRCClient) {
		c.onDisconnect = fn
	}
}

// WithReconnectHandler sets the handler for reconnection events.
func WithReconnectHandler(fn func()) IRCOption {
	return func(c *IRCClient) {
		c.onReconnect = fn
	}
}

// WithRawMessageHandler sets the handler for raw IRC messages.
func WithRawMessageHandler(fn func(string)) IRCOption {
	return func(c *IRCClient) {
		c.onRawMessage = fn
	}
}

// Connect establishes a connection to Twitch IRC.
func (c *IRCClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.connected {
		c.mu.Unlock()
		return ErrIRCAlreadyConnected
	}
	if c.connecting {
		c.mu.Unlock()
		return ErrIRCAlreadyConnected
	}
	c.connecting = true
	c.mu.Unlock()

	// Ensure connecting flag is cleared on exit
	defer func() {
		c.mu.Lock()
		c.connecting = false
		c.mu.Unlock()
	}()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("connecting to IRC: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.stopChan = make(chan struct{})
	c.stopOnce = sync.Once{} // reset for new connection
	c.mu.Unlock()

	// Request capabilities
	caps := strings.Join(c.capabilities, " ")
	if err := c.send(fmt.Sprintf("CAP REQ :%s", caps)); err != nil {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		_ = conn.Close()
		return fmt.Errorf("requesting capabilities: %w", err)
	}

	// Authenticate
	if err := c.send(fmt.Sprintf("PASS %s", c.token)); err != nil {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		_ = conn.Close()
		return fmt.Errorf("sending PASS: %w", err)
	}

	if err := c.send(fmt.Sprintf("NICK %s", c.nick)); err != nil {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		_ = conn.Close()
		return fmt.Errorf("sending NICK: %w", err)
	}

	// Wait for authentication response
	if err := c.waitForAuth(ctx); err != nil {
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		_ = conn.Close()
		return err
	}

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	// Start read loop
	c.wg.Add(1)
	go c.readLoop()

	// Rejoin channels
	c.mu.RLock()
	channels := make([]string, 0, len(c.channels))
	for ch := range c.channels {
		channels = append(channels, ch)
	}
	c.mu.RUnlock()

	if len(channels) > 0 {
		if err := c.Join(channels...); err != nil && c.onError != nil {
			c.onError(fmt.Errorf("rejoining channels: %w", err))
		}
	}

	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// waitForAuth waits for authentication confirmation.
func (c *IRCClient) waitForAuth(ctx context.Context) error {
	// Set read deadline based on context, defaulting to 30 seconds
	deadline := time.Now().Add(30 * time.Second)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	_ = c.conn.SetReadDeadline(deadline)
	defer func() { _ = c.conn.SetReadDeadline(time.Time{}) }()

	// Read messages until we get 001 (welcome) or NOTICE (auth failed)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			// Check if context was cancelled (deadline-related error)
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			return fmt.Errorf("reading auth response: %w", err)
		}

		lines := strings.Split(string(data), "\r\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			msg := parseIRCMessage(line)

			switch msg.Command {
			case "001": // RPL_WELCOME
				return nil
			case ircNOTICE:
				if strings.Contains(msg.Trailing, "Login authentication failed") ||
					strings.Contains(msg.Trailing, "Improperly formatted auth") {
					return ErrIRCAuthFailed
				}
			case ircGLOBALUSERSTATE:
				c.mu.Lock()
				c.globalState = parseGlobalUserState(msg)
				c.mu.Unlock()
				if c.onGlobalUserState != nil {
					c.onGlobalUserState(c.globalState)
				}
			case ircCAP:
				// CAP ACK - capabilities acknowledged
				continue
			}
		}
	}
}

// readLoop continuously reads messages from the WebSocket.
func (c *IRCClient) readLoop() {
	defer c.wg.Done()
	defer func() {
		c.mu.Lock()
		wasConnected := c.connected
		shouldReconnect := c.autoReconnect // capture under lock to avoid race
		c.connected = false
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.mu.Unlock()

		if wasConnected && c.onDisconnect != nil {
			c.onDisconnect()
		}

		// Auto-reconnect
		if wasConnected && shouldReconnect {
			go c.reconnect()
		}
	}()

	for {
		// Capture connection and stopChan under lock
		c.mu.RLock()
		conn := c.conn
		stopChan := c.stopChan
		c.mu.RUnlock()

		// Check if we should stop
		select {
		case <-stopChan:
			return
		default:
		}

		// Check for nil connection
		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			if c.onError != nil && !errors.Is(err, websocket.ErrCloseSent) {
				c.onError(fmt.Errorf("reading message: %w", err))
			}
			return
		}

		lines := strings.Split(string(data), "\r\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			// Recover from panics in handlers to prevent crashing the connection
			func() {
				defer func() {
					if r := recover(); r != nil {
						if c.onError != nil {
							c.onError(fmt.Errorf("handler panic: %v", r))
						}
					}
				}()

				if c.onRawMessage != nil {
					c.onRawMessage(line)
				}

				c.handleMessage(line)
			}()
		}
	}
}

// handleMessage processes a single IRC message.
func (c *IRCClient) handleMessage(raw string) {
	msg := parseIRCMessage(raw)

	switch msg.Command {
	case ircPING:
		_ = c.send("PONG :" + msg.Trailing)

	case ircPONG:
		select {
		case c.pongReceived <- struct{}{}:
		default:
		}

	case ircPRIVMSG:
		if c.onMessage != nil {
			c.onMessage(parseChatMessage(msg))
		}

	case ircWHISPER:
		if c.onWhisper != nil {
			c.onWhisper(parseWhisper(msg))
		}

	case ircUSERNOTICE:
		if c.onUserNotice != nil {
			c.onUserNotice(parseUserNotice(msg))
		}

	case ircNOTICE:
		if c.onNotice != nil {
			c.onNotice(parseNotice(msg))
		}

	case ircROOMSTATE:
		if c.onRoomState != nil {
			c.onRoomState(parseRoomState(msg))
		}

	case ircCLEARCHAT:
		if c.onClearChat != nil {
			c.onClearChat(parseClearChat(msg))
		}

	case ircCLEARMSG:
		if c.onClearMessage != nil {
			c.onClearMessage(parseClearMessage(msg))
		}

	case ircGLOBALUSERSTATE:
		state := parseGlobalUserState(msg)
		c.mu.Lock()
		c.globalState = state
		c.mu.Unlock()
		if c.onGlobalUserState != nil {
			c.onGlobalUserState(state)
		}

	case ircUSERSTATE:
		if c.onUserState != nil {
			c.onUserState(parseUserState(msg))
		}

	case ircJOIN:
		if c.onJoin != nil {
			channel := ""
			if len(msg.Params) > 0 {
				channel = parseChannel(msg.Params[0])
			}
			user := parseUserFromPrefix(msg.Prefix)
			c.onJoin(channel, user)
		}

	case ircPART:
		if c.onPart != nil {
			channel := ""
			if len(msg.Params) > 0 {
				channel = parseChannel(msg.Params[0])
			}
			user := parseUserFromPrefix(msg.Prefix)
			c.onPart(channel, user)
		}

	case ircRECONNECT:
		// Twitch is requesting we reconnect
		if c.onReconnect != nil {
			c.onReconnect()
		}
		c.mu.Lock()
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.mu.Unlock()
		// Note: connected stays true so readLoop's defer will trigger auto-reconnect
	}
}

// reconnect attempts to reconnect to IRC.
func (c *IRCClient) reconnect() {
	for {
		c.mu.RLock()
		stopChan := c.stopChan
		shouldReconnect := c.autoReconnect
		c.mu.RUnlock()

		// Check if we should stop before waiting
		if !shouldReconnect {
			return
		}

		select {
		case <-stopChan:
			return
		case <-time.After(c.reconnectDelay):
		}

		// Re-check if auto-reconnect was disabled during the delay
		c.mu.RLock()
		shouldReconnect = c.autoReconnect
		stopChan = c.stopChan
		c.mu.RUnlock()

		if !shouldReconnect {
			return
		}

		// Final check before Connect to minimize race window
		select {
		case <-stopChan:
			return
		default:
		}

		if c.onReconnect != nil {
			c.onReconnect()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := c.Connect(ctx)
		cancel()

		if err == nil {
			return
		}

		if c.onError != nil {
			c.onError(fmt.Errorf("reconnect failed: %w", err))
		}
	}
}

// send sends a raw IRC message.
func (c *IRCClient) send(message string) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return ErrIRCNotConnected
	}

	return conn.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
}

// Close closes the IRC connection.
func (c *IRCClient) Close() error {
	c.mu.Lock()
	// Close if connected OR if there's an in-progress connection (conn != nil)
	if !c.connected && c.conn == nil {
		c.mu.Unlock()
		return nil
	}

	c.autoReconnect = false
	c.connected = false

	// Signal readLoop to stop (only once to prevent panic)
	c.stopOnce.Do(func() {
		if c.stopChan != nil {
			close(c.stopChan)
		}
	})
	conn := c.conn
	c.mu.Unlock()

	// Close connection to unblock ReadMessage in readLoop
	if conn != nil {
		_ = conn.Close()
	}

	// Wait for readLoop to finish
	c.wg.Wait()

	return nil
}

// IsConnected returns whether the client is connected.
func (c *IRCClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Join joins one or more channels.
// Channel names are sanitized to prevent IRC command injection.
func (c *IRCClient) Join(channels ...string) error {
	c.mu.Lock()
	for _, ch := range channels {
		ch = sanitizeIRCMessage(strings.ToLower(strings.TrimPrefix(ch, "#")))
		c.channels[ch] = true
	}
	c.mu.Unlock()

	if !c.IsConnected() {
		return nil // Will join on connect
	}

	for _, ch := range channels {
		ch = sanitizeIRCMessage(strings.ToLower(strings.TrimPrefix(ch, "#")))
		if err := c.send(fmt.Sprintf("JOIN #%s", ch)); err != nil {
			return fmt.Errorf("joining %s: %w", ch, err)
		}
	}

	return nil
}

// Part leaves one or more channels.
// Channel names are sanitized to prevent IRC command injection.
func (c *IRCClient) Part(channels ...string) error {
	c.mu.Lock()
	for _, ch := range channels {
		ch = sanitizeIRCMessage(strings.ToLower(strings.TrimPrefix(ch, "#")))
		delete(c.channels, ch)
	}
	c.mu.Unlock()

	if !c.IsConnected() {
		return nil
	}

	for _, ch := range channels {
		ch = sanitizeIRCMessage(strings.ToLower(strings.TrimPrefix(ch, "#")))
		if err := c.send(fmt.Sprintf("PART #%s", ch)); err != nil {
			return fmt.Errorf("parting %s: %w", ch, err)
		}
	}

	return nil
}

// Say sends a message to a channel.
// The channel name and message are sanitized to prevent IRC command injection.
func (c *IRCClient) Say(channel, message string) error {
	channel = sanitizeIRCMessage(strings.ToLower(strings.TrimPrefix(channel, "#")))
	message = sanitizeIRCMessage(message)
	return c.send(fmt.Sprintf("PRIVMSG #%s :%s", channel, message))
}

// Reply sends a reply to a message.
// The channel name, parent message ID, and message are sanitized to prevent IRC command injection.
func (c *IRCClient) Reply(channel, parentMsgID, message string) error {
	channel = sanitizeIRCMessage(strings.ToLower(strings.TrimPrefix(channel, "#")))
	parentMsgID = sanitizeIRCMessage(parentMsgID)
	message = sanitizeIRCMessage(message)
	return c.send(fmt.Sprintf("@reply-parent-msg-id=%s PRIVMSG #%s :%s", parentMsgID, channel, message))
}

// Whisper sends a whisper to a user.
// Note: Whispers require verified bot status for high volume.
// The message is sanitized to prevent IRC command injection.
func (c *IRCClient) Whisper(user, message string) error {
	user = sanitizeIRCMessage(user)
	message = sanitizeIRCMessage(message)
	return c.send(fmt.Sprintf("PRIVMSG #jtv :/w %s %s", user, message))
}

// GetGlobalUserState returns the global user state.
func (c *IRCClient) GetGlobalUserState() *GlobalUserState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.globalState
}

// GetJoinedChannels returns the list of joined channels.
func (c *IRCClient) GetJoinedChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	channels := make([]string, 0, len(c.channels))
	for ch := range c.channels {
		channels = append(channels, ch)
	}
	return channels
}

// Ping sends a PING and waits for PONG.
func (c *IRCClient) Ping(ctx context.Context) error {
	// Clear any pending pong
	select {
	case <-c.pongReceived:
	default:
	}

	if err := c.send("PING :tmi.twitch.tv"); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.pongReceived:
		return nil
	}
}
