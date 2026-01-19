package helix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// PubSub compatibility errors.
var (
	ErrPubSubNotConnected     = errors.New("pubsub: not connected")
	ErrPubSubInvalidTopic     = errors.New("pubsub: invalid topic format")
	ErrPubSubUnsupportedTopic = errors.New("pubsub: topic type not supported")
)

// ParsedTopic represents a parsed PubSub topic string.
type ParsedTopic struct {
	Type        string // e.g., "channel-bits-events", "channel-points-channel"
	ChannelID   string // broadcaster/channel ID
	UserID      string // user ID (for whispers, moderator actions)
	ModeratorID string // moderator ID (for automod)
}

// TopicMapping defines how a PubSub topic maps to EventSub subscriptions.
type TopicMapping struct {
	EventSubTypes []string                                    // EventSub subscription types
	Condition     func(parsed *ParsedTopic) map[string]string // Build condition from parsed topic
}

// Topic patterns for parsing PubSub topic strings.
var topicPatterns = map[string]*regexp.Regexp{
	"channel-bits-events":    regexp.MustCompile(`^channel-bits-events-v[12]\.(\d+)$`),
	"channel-bits-badge":     regexp.MustCompile(`^channel-bits-badge-unlocks\.(\d+)$`),
	"channel-points-channel": regexp.MustCompile(`^channel-points-channel-v1\.(\d+)$`),
	"channel-subscribe":      regexp.MustCompile(`^channel-subscribe-events-v1\.(\d+)$`),
	"automod-queue":          regexp.MustCompile(`^automod-queue\.(\d+)\.(\d+)$`),
	"chat-moderator-actions": regexp.MustCompile(`^chat_moderator_actions\.(\d+)\.(\d+)$`),
	"whispers":               regexp.MustCompile(`^whispers\.(\d+)$`),
}

// topicMappings maps PubSub topic types to EventSub subscriptions.
var topicMappings = map[string]TopicMapping{
	"channel-bits-events": {
		EventSubTypes: []string{EventSubTypeChannelCheer},
		Condition: func(p *ParsedTopic) map[string]string {
			return BroadcasterCondition(p.ChannelID)
		},
	},
	"channel-bits-badge": {
		// EventSub uses channel.chat.notification for bits badge unlocks
		EventSubTypes: []string{EventSubTypeChannelChatNotification},
		Condition: func(p *ParsedTopic) map[string]string {
			return BroadcasterCondition(p.ChannelID)
		},
	},
	"channel-points-channel": {
		EventSubTypes: []string{EventSubTypeChannelPointsRedemptionAdd},
		Condition: func(p *ParsedTopic) map[string]string {
			return BroadcasterCondition(p.ChannelID)
		},
	},
	"channel-subscribe": {
		// Multiple EventSub types for one PubSub topic
		EventSubTypes: []string{
			EventSubTypeChannelSubscribe,
			EventSubTypeChannelSubscriptionGift,
			EventSubTypeChannelSubscriptionMessage,
		},
		Condition: func(p *ParsedTopic) map[string]string {
			return BroadcasterCondition(p.ChannelID)
		},
	},
	"automod-queue": {
		EventSubTypes: []string{EventSubTypeAutomodMessageHold},
		Condition: func(p *ParsedTopic) map[string]string {
			return BroadcasterModeratorCondition(p.ChannelID, p.ModeratorID)
		},
	},
	"chat-moderator-actions": {
		EventSubTypes: []string{EventSubTypeChannelModerate},
		Condition: func(p *ParsedTopic) map[string]string {
			return BroadcasterModeratorCondition(p.ChannelID, p.UserID)
		},
	},
	"whispers": {
		EventSubTypes: []string{EventSubTypeUserWhisperMessage},
		Condition: func(p *ParsedTopic) map[string]string {
			return UserCondition(p.UserID)
		},
	},
}

// ParseTopic parses a PubSub topic string into its components.
func ParseTopic(topic string) (*ParsedTopic, error) {
	// Handle channel-bits-events-v1.<channel_id> and v2
	if matches := topicPatterns["channel-bits-events"].FindStringSubmatch(topic); matches != nil {
		return &ParsedTopic{
			Type:      "channel-bits-events",
			ChannelID: matches[1],
		}, nil
	}

	// Handle channel-bits-badge-unlocks.<channel_id>
	if matches := topicPatterns["channel-bits-badge"].FindStringSubmatch(topic); matches != nil {
		return &ParsedTopic{
			Type:      "channel-bits-badge",
			ChannelID: matches[1],
		}, nil
	}

	// Handle channel-points-channel-v1.<channel_id>
	if matches := topicPatterns["channel-points-channel"].FindStringSubmatch(topic); matches != nil {
		return &ParsedTopic{
			Type:      "channel-points-channel",
			ChannelID: matches[1],
		}, nil
	}

	// Handle channel-subscribe-events-v1.<channel_id>
	if matches := topicPatterns["channel-subscribe"].FindStringSubmatch(topic); matches != nil {
		return &ParsedTopic{
			Type:      "channel-subscribe",
			ChannelID: matches[1],
		}, nil
	}

	// Handle automod-queue.<moderator_id>.<channel_id>
	if matches := topicPatterns["automod-queue"].FindStringSubmatch(topic); matches != nil {
		return &ParsedTopic{
			Type:        "automod-queue",
			ModeratorID: matches[1],
			ChannelID:   matches[2],
		}, nil
	}

	// Handle chat_moderator_actions.<user_id>.<channel_id>
	if matches := topicPatterns["chat-moderator-actions"].FindStringSubmatch(topic); matches != nil {
		return &ParsedTopic{
			Type:      "chat-moderator-actions",
			UserID:    matches[1],
			ChannelID: matches[2],
		}, nil
	}

	// Handle whispers.<user_id>
	if matches := topicPatterns["whispers"].FindStringSubmatch(topic); matches != nil {
		return &ParsedTopic{
			Type:   "whispers",
			UserID: matches[1],
		}, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrPubSubInvalidTopic, topic)
}

// PubSubMessage wraps EventSub events in a PubSub-style envelope.
type PubSubMessage struct {
	Type string          `json:"type"` // EventSub subscription type
	Data json.RawMessage `json:"data"` // EventSub event payload
}

// PubSubClient provides a PubSub-style API backed by EventSub.
// This enables migration from the deprecated Twitch PubSub to EventSub
// while maintaining familiar topic-based semantics.
type PubSubClient struct {
	client    *Client                  // Helix API client
	ws        *EventSubWebSocketClient // Internal WebSocket
	wsURL     string                   // Custom WebSocket URL (for testing)
	sessionID string

	// Topic tracking: maps PubSub topic -> EventSub subscription IDs
	topics     map[string][]string // topic -> subscription IDs
	subToTopic map[string]string   // subscription ID -> topic

	// Handlers
	onMessage   func(topic string, message json.RawMessage)
	onError     func(error)
	onConnect   func()
	onReconnect func()

	mu sync.RWMutex
	wg sync.WaitGroup // tracks background goroutines (e.g., reconnect)
}

// PubSubOption configures the PubSubClient.
type PubSubOption func(*PubSubClient)

// WithPubSubMessageHandler sets the handler for topic messages.
func WithPubSubMessageHandler(fn func(topic string, message json.RawMessage)) PubSubOption {
	return func(c *PubSubClient) {
		c.onMessage = fn
	}
}

// WithPubSubErrorHandler sets the handler for errors.
func WithPubSubErrorHandler(fn func(error)) PubSubOption {
	return func(c *PubSubClient) {
		c.onError = fn
	}
}

// WithPubSubConnectHandler sets the handler for connection events.
func WithPubSubConnectHandler(fn func()) PubSubOption {
	return func(c *PubSubClient) {
		c.onConnect = fn
	}
}

// WithPubSubReconnectHandler sets the handler for reconnection events.
func WithPubSubReconnectHandler(fn func()) PubSubOption {
	return func(c *PubSubClient) {
		c.onReconnect = fn
	}
}

// WithPubSubWSURL sets a custom WebSocket URL (useful for testing).
func WithPubSubWSURL(url string) PubSubOption {
	return func(c *PubSubClient) {
		c.wsURL = url
	}
}

// NewPubSubClient creates a new PubSub compatibility client.
// The client uses EventSub WebSocket internally but exposes a PubSub-style API.
// Returns nil if helixClient is nil.
func NewPubSubClient(helixClient *Client, opts ...PubSubOption) *PubSubClient {
	if helixClient == nil {
		return nil
	}

	c := &PubSubClient{
		client:     helixClient,
		topics:     make(map[string][]string),
		subToTopic: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect establishes the WebSocket connection.
func (c *PubSubClient) Connect(ctx context.Context) error {
	c.mu.Lock()

	if c.ws != nil && c.ws.IsConnected() {
		c.mu.Unlock()
		return nil // Already connected
	}

	// Build WebSocket options
	wsOpts := []EventSubWSOption{
		WithWSNotificationHandler(c.handleNotification),
		WithWSRevocationHandler(c.handleRevocation),
		WithWSReconnectHandler(c.handleReconnect),
		WithWSErrorHandler(c.handleError),
	}

	if c.wsURL != "" {
		wsOpts = append(wsOpts, WithWSURL(c.wsURL))
	}

	c.ws = NewEventSubWebSocketClient(wsOpts...)

	sessionID, err := c.ws.Connect(ctx)
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("connecting to EventSub: %w", err)
	}

	c.sessionID = sessionID
	onConnect := c.onConnect
	c.mu.Unlock()

	// Call handler outside of lock to prevent deadlock
	if onConnect != nil {
		onConnect()
	}

	return nil
}

// Listen subscribes to a PubSub topic.
// The topic is translated to the equivalent EventSub subscription(s).
func (c *PubSubClient) Listen(ctx context.Context, topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ws == nil || !c.ws.IsConnected() {
		return ErrPubSubNotConnected
	}

	// Check if already listening
	if _, exists := c.topics[topic]; exists {
		return nil // Already subscribed
	}

	// Parse topic
	parsed, err := ParseTopic(topic)
	if err != nil {
		return err
	}

	// Get mapping
	mapping, exists := topicMappings[parsed.Type]
	if !exists {
		return fmt.Errorf("%w: %s", ErrPubSubUnsupportedTopic, parsed.Type)
	}

	// Build condition
	condition := mapping.Condition(parsed)

	// Create EventSub subscriptions
	var subIDs []string
	transport := CreateEventSubTransport{
		Method:    EventSubTransportWebSocket,
		SessionID: c.sessionID,
	}

	for _, eventType := range mapping.EventSubTypes {
		sub, err := c.client.CreateEventSubSubscription(ctx, &CreateEventSubSubscriptionParams{
			Type:      eventType,
			Version:   GetEventSubVersion(eventType),
			Condition: condition,
			Transport: transport,
		})
		if err != nil {
			// Cleanup any created subscriptions on failure
			for _, id := range subIDs {
				_ = c.client.DeleteEventSubSubscription(ctx, id)
			}
			return fmt.Errorf("creating subscription for %s: %w", eventType, err)
		}

		subIDs = append(subIDs, sub.ID)
		c.subToTopic[sub.ID] = topic
	}

	c.topics[topic] = subIDs
	return nil
}

// Unlisten unsubscribes from a PubSub topic.
func (c *PubSubClient) Unlisten(ctx context.Context, topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	subIDs, exists := c.topics[topic]
	if !exists {
		return nil // Not listening
	}

	var lastErr error
	for _, id := range subIDs {
		if err := c.client.DeleteEventSubSubscription(ctx, id); err != nil {
			lastErr = err
		}
		delete(c.subToTopic, id)
	}

	delete(c.topics, topic)
	return lastErr
}

// Close closes the PubSub client and cleans up all subscriptions.
func (c *PubSubClient) Close(ctx context.Context) error {
	c.mu.Lock()

	// Delete all subscriptions, collecting errors
	var errs []error
	for _, subIDs := range c.topics {
		for _, id := range subIDs {
			if err := c.client.DeleteEventSubSubscription(ctx, id); err != nil {
				errs = append(errs, fmt.Errorf("deleting subscription %s: %w", id, err))
			}
		}
	}
	c.topics = make(map[string][]string)
	c.subToTopic = make(map[string]string)

	// Close WebSocket
	ws := c.ws
	c.mu.Unlock()

	// Wait for background goroutines to finish
	c.wg.Wait()

	if ws != nil {
		if err := ws.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing websocket: %w", err))
		}
	}

	// Return combined error if any
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// IsConnected returns whether the client is connected.
func (c *PubSubClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ws != nil && c.ws.IsConnected()
}

// Topics returns the list of topics currently being listened to.
func (c *PubSubClient) Topics() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	topics := make([]string, 0, len(c.topics))
	for topic := range c.topics {
		topics = append(topics, topic)
	}
	return topics
}

// SessionID returns the EventSub session ID.
func (c *PubSubClient) SessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

// handleNotification routes EventSub notifications to the appropriate topic handler.
func (c *PubSubClient) handleNotification(sub *EventSubSubscription, event json.RawMessage) {
	c.mu.RLock()
	topic, exists := c.subToTopic[sub.ID]
	handler := c.onMessage
	c.mu.RUnlock()

	if !exists || handler == nil {
		return
	}

	// Wrap event in a PubSub-style message envelope
	message := PubSubMessage{
		Type: sub.Type,
		Data: event,
	}

	msgJSON, err := json.Marshal(message)
	if err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("marshaling message: %w", err))
		}
		return
	}

	handler(topic, msgJSON)
}

// handleRevocation handles subscription revocations.
func (c *PubSubClient) handleRevocation(sub *EventSubSubscription) {
	c.mu.Lock()
	topic, exists := c.subToTopic[sub.ID]
	if exists {
		delete(c.subToTopic, sub.ID)
		// Remove from topics list
		if subIDs, ok := c.topics[topic]; ok {
			for i, id := range subIDs {
				if id == sub.ID {
					c.topics[topic] = append(subIDs[:i], subIDs[i+1:]...)
					break
				}
			}
			if len(c.topics[topic]) == 0 {
				delete(c.topics, topic)
			}
		}
	}
	c.mu.Unlock()

	if c.onError != nil {
		c.onError(fmt.Errorf("subscription revoked: %s (topic: %s, reason: %s)",
			sub.ID, topic, sub.Status))
	}
}

// handleReconnect handles WebSocket reconnection requests.
func (c *PubSubClient) handleReconnect(reconnectURL string) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		newSessionID, err := c.ws.Reconnect(ctx, reconnectURL)
		if err != nil {
			if c.onError != nil {
				c.onError(fmt.Errorf("reconnecting: %w", err))
			}
			return
		}

		c.mu.Lock()
		c.sessionID = newSessionID
		c.mu.Unlock()

		if c.onReconnect != nil {
			c.onReconnect()
		}
	}()
}

// handleError routes errors to the error handler.
func (c *PubSubClient) handleError(err error) {
	if c.onError != nil {
		c.onError(err)
	}
}

// SupportedTopics returns a list of supported PubSub topic patterns.
func SupportedTopics() []string {
	return []string{
		"channel-bits-events-v1.<channel_id>",
		"channel-bits-events-v2.<channel_id>",
		"channel-bits-badge-unlocks.<channel_id>",
		"channel-points-channel-v1.<channel_id>",
		"channel-subscribe-events-v1.<channel_id>",
		"automod-queue.<moderator_id>.<channel_id>",
		"chat_moderator_actions.<user_id>.<channel_id>",
		"whispers.<user_id>",
	}
}

// TopicEventSubTypes returns the EventSub types that a PubSub topic maps to.
// Returns nil if the topic is invalid or unsupported.
func TopicEventSubTypes(topic string) []string {
	parsed, err := ParseTopic(topic)
	if err != nil {
		return nil
	}

	mapping, exists := topicMappings[parsed.Type]
	if !exists {
		return nil
	}

	return mapping.EventSubTypes
}

// BuildTopic constructs a PubSub topic string from components.
// This is a helper for users migrating from PubSub who need to construct topics.
func BuildTopic(topicType string, ids ...string) string {
	switch topicType {
	case "channel-bits-events-v1", "channel-bits-events-v2":
		if len(ids) >= 1 {
			return topicType + "." + ids[0]
		}
	case "channel-bits-badge-unlocks", "channel-points-channel-v1", "channel-subscribe-events-v1":
		if len(ids) >= 1 {
			return topicType + "." + ids[0]
		}
	case "automod-queue":
		if len(ids) >= 2 {
			return "automod-queue." + ids[0] + "." + ids[1]
		}
	case "chat_moderator_actions":
		if len(ids) >= 2 {
			return "chat_moderator_actions." + ids[0] + "." + ids[1]
		}
	case "whispers":
		if len(ids) >= 1 {
			return "whispers." + ids[0]
		}
	}
	return ""
}

