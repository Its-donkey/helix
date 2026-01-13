package helix

import (
	"context"
	"net/url"
	"time"
)

// EventSubSubscription represents an EventSub subscription.
type EventSubSubscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition map[string]string `json:"condition"`
	CreatedAt time.Time         `json:"created_at"`
	Transport EventSubTransport `json:"transport"`
	Cost      int               `json:"cost"`
}

// EventSubTransport represents the transport for an EventSub subscription.
type EventSubTransport struct {
	Method         string `json:"method"` // webhook, websocket, conduit
	Callback       string `json:"callback,omitempty"`
	Secret         string `json:"secret,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	ConduitID      string `json:"conduit_id,omitempty"`
	ConnectedAt    string `json:"connected_at,omitempty"`
	DisconnectedAt string `json:"disconnected_at,omitempty"`
}

// GetEventSubSubscriptionsParams contains parameters for GetEventSubSubscriptions.
type GetEventSubSubscriptionsParams struct {
	Status string // Filter by status
	Type   string // Filter by subscription type
	UserID string // Filter by user ID
	*PaginationParams
}

// EventSubResponse represents the response from EventSub operations.
type EventSubResponse struct {
	Data         []EventSubSubscription `json:"data"`
	Total        int                    `json:"total"`
	TotalCost    int                    `json:"total_cost"`
	MaxTotalCost int                    `json:"max_total_cost"`
	Pagination   *Pagination            `json:"pagination,omitempty"`
}

// GetEventSubSubscriptions gets EventSub subscriptions.
// Requires: App access token.
func (c *Client) GetEventSubSubscriptions(ctx context.Context, params *GetEventSubSubscriptionsParams) (*EventSubResponse, error) {
	q := url.Values{}
	if params != nil {
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if params.Type != "" {
			q.Set("type", params.Type)
		}
		if params.UserID != "" {
			q.Set("user_id", params.UserID)
		}
		addPaginationParams(q, params.PaginationParams)
	}

	var resp EventSubResponse
	if err := c.get(ctx, "/eventsub/subscriptions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateEventSubSubscriptionParams contains parameters for CreateEventSubSubscription.
type CreateEventSubSubscriptionParams struct {
	Type      string                  `json:"type"`
	Version   string                  `json:"version"`
	Condition map[string]string       `json:"condition"`
	Transport CreateEventSubTransport `json:"transport"`
}

// CreateEventSubTransport represents the transport when creating a subscription.
type CreateEventSubTransport struct {
	Method    string `json:"method"` // webhook, websocket, conduit
	Callback  string `json:"callback,omitempty"`
	Secret    string `json:"secret,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	ConduitID string `json:"conduit_id,omitempty"`
}

// CreateEventSubSubscription creates an EventSub subscription.
// Requires: App access token.
func (c *Client) CreateEventSubSubscription(ctx context.Context, params *CreateEventSubSubscriptionParams) (*EventSubSubscription, error) {
	var resp EventSubResponse
	if err := c.post(ctx, "/eventsub/subscriptions", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// DeleteEventSubSubscription deletes an EventSub subscription.
// Requires: App access token.
func (c *Client) DeleteEventSubSubscription(ctx context.Context, subscriptionID string) error {
	q := url.Values{}
	q.Set("id", subscriptionID)

	return c.delete(ctx, "/eventsub/subscriptions", q, nil)
}

// EventSub subscription types - Automod
const (
	EventSubTypeAutomodMessageHold    = "automod.message.hold"
	EventSubTypeAutomodMessageUpdate  = "automod.message.update"
	EventSubTypeAutomodSettingsUpdate = "automod.settings.update"
	EventSubTypeAutomodTermsUpdate    = "automod.terms.update"
)

// EventSub subscription types - Channel
const (
	EventSubTypeChannelUpdate              = "channel.update"
	EventSubTypeChannelFollow              = "channel.follow"
	EventSubTypeChannelAdBreakBegin        = "channel.ad_break.begin"
	EventSubTypeChannelBitsUse             = "channel.bits.use"
	EventSubTypeChannelSubscribe           = "channel.subscribe"
	EventSubTypeChannelSubscriptionEnd     = "channel.subscription.end"
	EventSubTypeChannelSubscriptionGift    = "channel.subscription.gift"
	EventSubTypeChannelSubscriptionMessage = "channel.subscription.message"
	EventSubTypeChannelCheer               = "channel.cheer"
	EventSubTypeChannelRaid                = "channel.raid"
	EventSubTypeChannelBan                 = "channel.ban"
	EventSubTypeChannelUnban               = "channel.unban"
	EventSubTypeChannelUnbanRequestCreate  = "channel.unban_request.create"
	EventSubTypeChannelUnbanRequestResolve = "channel.unban_request.resolve"
	EventSubTypeChannelModerate            = "channel.moderate"
	EventSubTypeChannelModeratorAdd        = "channel.moderator.add"
	EventSubTypeChannelModeratorRemove     = "channel.moderator.remove"
	EventSubTypeChannelVIPAdd              = "channel.vip.add"
	EventSubTypeChannelVIPRemove           = "channel.vip.remove"
	EventSubTypeChannelWarningSend         = "channel.warning.send"
	EventSubTypeChannelWarningAcknowledge  = "channel.warning.acknowledge"
)

// EventSub subscription types - Channel Chat
const (
	EventSubTypeChannelChatClear             = "channel.chat.clear"
	EventSubTypeChannelChatClearUserMessages = "channel.chat.clear_user_messages"
	EventSubTypeChannelChatMessage           = "channel.chat.message"
	EventSubTypeChannelChatMessageDelete     = "channel.chat.message_delete"
	EventSubTypeChannelChatNotification      = "channel.chat.notification"
	EventSubTypeChannelChatSettingsUpdate    = "channel.chat_settings.update"
	EventSubTypeChannelChatUserMessageHold   = "channel.chat.user_message_hold"
	EventSubTypeChannelChatUserMessageUpdate = "channel.chat.user_message_update"
)

// EventSub subscription types - Channel Shared Chat
const (
	EventSubTypeChannelSharedChatBegin  = "channel.shared_chat.begin"
	EventSubTypeChannelSharedChatUpdate = "channel.shared_chat.update"
	EventSubTypeChannelSharedChatEnd    = "channel.shared_chat.end"
)

// EventSub subscription types - Channel Points
const (
	EventSubTypeChannelPointsAutomaticRewardRedemptionAdd = "channel.channel_points_automatic_reward_redemption.add"
	EventSubTypeChannelPointsRewardAdd                    = "channel.channel_points_custom_reward.add"
	EventSubTypeChannelPointsRewardUpdate                 = "channel.channel_points_custom_reward.update"
	EventSubTypeChannelPointsRewardRemove                 = "channel.channel_points_custom_reward.remove"
	EventSubTypeChannelPointsRedemptionAdd                = "channel.channel_points_custom_reward_redemption.add"
	EventSubTypeChannelPointsRedemptionUpdate             = "channel.channel_points_custom_reward_redemption.update"
)

// EventSub subscription types - Polls & Predictions
const (
	EventSubTypeChannelPollBegin          = "channel.poll.begin"
	EventSubTypeChannelPollProgress       = "channel.poll.progress"
	EventSubTypeChannelPollEnd            = "channel.poll.end"
	EventSubTypeChannelPredictionBegin    = "channel.prediction.begin"
	EventSubTypeChannelPredictionProgress = "channel.prediction.progress"
	EventSubTypeChannelPredictionLock     = "channel.prediction.lock"
	EventSubTypeChannelPredictionEnd      = "channel.prediction.end"
)

// EventSub subscription types - Hype Train (default v2)
const (
	EventSubTypeChannelHypeTrainBegin    = "channel.hype_train.begin"
	EventSubTypeChannelHypeTrainProgress = "channel.hype_train.progress"
	EventSubTypeChannelHypeTrainEnd      = "channel.hype_train.end"
)

// EventSub Hype Train version constants.
const (
	// Deprecated: EventSubVersionHypeTrainV1 is deprecated by Twitch. Use EventSubVersionHypeTrainV2.
	EventSubVersionHypeTrainV1 = "1"
	EventSubVersionHypeTrainV2 = "2"
)

// EventSub subscription types - Charity
const (
	EventSubTypeChannelCharityCampaignDonate   = "channel.charity_campaign.donate"
	EventSubTypeChannelCharityCampaignStart    = "channel.charity_campaign.start"
	EventSubTypeChannelCharityCampaignProgress = "channel.charity_campaign.progress"
	EventSubTypeChannelCharityCampaignStop     = "channel.charity_campaign.stop"
)

// EventSub subscription types - Goals
const (
	EventSubTypeChannelGoalBegin    = "channel.goal.begin"
	EventSubTypeChannelGoalProgress = "channel.goal.progress"
	EventSubTypeChannelGoalEnd      = "channel.goal.end"
)

// EventSub subscription types - Shield Mode
const (
	EventSubTypeChannelShieldModeBegin = "channel.shield_mode.begin"
	EventSubTypeChannelShieldModeEnd   = "channel.shield_mode.end"
)

// EventSub subscription types - Shoutout
const (
	EventSubTypeChannelShoutoutCreate  = "channel.shoutout.create"
	EventSubTypeChannelShoutoutReceive = "channel.shoutout.receive"
)

// EventSub subscription types - Suspicious User
const (
	EventSubTypeChannelSuspiciousUserMessage = "channel.suspicious_user.message"
	EventSubTypeChannelSuspiciousUserUpdate  = "channel.suspicious_user.update"
)

// EventSub subscription types - Guest Star (Beta)
const (
	EventSubTypeChannelGuestStarSessionBegin   = "channel.guest_star_session.begin"
	EventSubTypeChannelGuestStarSessionEnd     = "channel.guest_star_session.end"
	EventSubTypeChannelGuestStarGuestUpdate    = "channel.guest_star_guest.update"
	EventSubTypeChannelGuestStarSettingsUpdate = "channel.guest_star_settings.update"
)

// EventSub subscription types - Conduit
const (
	EventSubTypeConduitShardDisabled = "conduit.shard.disabled"
)

// EventSub subscription types - Drop
const (
	EventSubTypeDropEntitlementGrant = "drop.entitlement.grant"
)

// EventSub subscription types - Extension
const (
	EventSubTypeExtensionBitsTransactionCreate = "extension.bits_transaction.create"
)

// EventSub subscription types - Stream
const (
	EventSubTypeStreamOnline  = "stream.online"
	EventSubTypeStreamOffline = "stream.offline"
)

// EventSub subscription types - User
const (
	EventSubTypeUserAuthorizationGrant  = "user.authorization.grant"
	EventSubTypeUserAuthorizationRevoke = "user.authorization.revoke"
	EventSubTypeUserUpdate              = "user.update"
	EventSubTypeUserWhisperMessage      = "user.whisper.message"
)

// EventSub subscription statuses
const (
	EventSubStatusEnabled                            = "enabled"
	EventSubStatusWebhookCallbackVerificationPending = "webhook_callback_verification_pending"
	EventSubStatusWebhookCallbackVerificationFailed  = "webhook_callback_verification_failed"
	EventSubStatusNotificationFailuresExceeded       = "notification_failures_exceeded"
	EventSubStatusAuthorizationRevoked               = "authorization_revoked"
	EventSubStatusModeratorRemoved                   = "moderator_removed"
	EventSubStatusUserRemoved                        = "user_removed"
	EventSubStatusVersionRemoved                     = "version_removed"
	EventSubStatusBetaMaintenance                    = "beta_maintenance"
	EventSubStatusWebsocketDisconnected              = "websocket_disconnected"
	EventSubStatusWebsocketFailedPingPong            = "websocket_failed_ping_pong"
	EventSubStatusWebsocketReceivedInboundTraffic    = "websocket_received_inbound_traffic"
	EventSubStatusWebsocketConnectionUnused          = "websocket_connection_unused"
	EventSubStatusWebsocketInternalError             = "websocket_internal_error"
	EventSubStatusWebsocketNetworkTimeout            = "websocket_network_timeout"
	EventSubStatusWebsocketNetworkError              = "websocket_network_error"
	EventSubStatusWebsocketFailedToReconnect         = "websocket_failed_to_reconnect"
)

// EventSub transport methods
const (
	EventSubTransportWebhook   = "webhook"
	EventSubTransportWebSocket = "websocket"
	EventSubTransportConduit   = "conduit"
)

// EventSubTypeVersion maps subscription types to their latest versions.
var EventSubTypeVersion = map[string]string{
	// Automod
	EventSubTypeAutomodMessageHold:    "2",
	EventSubTypeAutomodMessageUpdate:  "2",
	EventSubTypeAutomodSettingsUpdate: "1",
	EventSubTypeAutomodTermsUpdate:    "1",
	// Channel
	EventSubTypeChannelUpdate:              "2",
	EventSubTypeChannelFollow:              "2",
	EventSubTypeChannelAdBreakBegin:        "1",
	EventSubTypeChannelBitsUse:             "1",
	EventSubTypeChannelSubscribe:           "1",
	EventSubTypeChannelSubscriptionEnd:     "1",
	EventSubTypeChannelSubscriptionGift:    "1",
	EventSubTypeChannelSubscriptionMessage: "1",
	EventSubTypeChannelCheer:               "1",
	EventSubTypeChannelRaid:                "1",
	EventSubTypeChannelBan:                 "1",
	EventSubTypeChannelUnban:               "1",
	EventSubTypeChannelUnbanRequestCreate:  "1",
	EventSubTypeChannelUnbanRequestResolve: "1",
	EventSubTypeChannelModerate:            "2",
	EventSubTypeChannelModeratorAdd:        "1",
	EventSubTypeChannelModeratorRemove:     "1",
	EventSubTypeChannelVIPAdd:              "1",
	EventSubTypeChannelVIPRemove:           "1",
	EventSubTypeChannelWarningSend:         "1",
	EventSubTypeChannelWarningAcknowledge:  "1",
	// Chat
	EventSubTypeChannelChatClear:             "1",
	EventSubTypeChannelChatClearUserMessages: "1",
	EventSubTypeChannelChatMessage:           "1",
	EventSubTypeChannelChatMessageDelete:     "1",
	EventSubTypeChannelChatNotification:      "1",
	EventSubTypeChannelChatSettingsUpdate:    "1",
	EventSubTypeChannelChatUserMessageHold:   "1",
	EventSubTypeChannelChatUserMessageUpdate: "1",
	// Shared Chat
	EventSubTypeChannelSharedChatBegin:  "1",
	EventSubTypeChannelSharedChatUpdate: "1",
	EventSubTypeChannelSharedChatEnd:    "1",
	// Channel Points
	EventSubTypeChannelPointsAutomaticRewardRedemptionAdd: "2",
	EventSubTypeChannelPointsRewardAdd:                    "1",
	EventSubTypeChannelPointsRewardUpdate:                 "1",
	EventSubTypeChannelPointsRewardRemove:                 "1",
	EventSubTypeChannelPointsRedemptionAdd:                "1",
	EventSubTypeChannelPointsRedemptionUpdate:             "1",
	// Polls & Predictions
	EventSubTypeChannelPollBegin:          "1",
	EventSubTypeChannelPollProgress:       "1",
	EventSubTypeChannelPollEnd:            "1",
	EventSubTypeChannelPredictionBegin:    "1",
	EventSubTypeChannelPredictionProgress: "1",
	EventSubTypeChannelPredictionLock:     "1",
	EventSubTypeChannelPredictionEnd:      "1",
	// Hype Train (v1 is deprecated by Twitch)
	EventSubTypeChannelHypeTrainBegin:    "2",
	EventSubTypeChannelHypeTrainProgress: "2",
	EventSubTypeChannelHypeTrainEnd:      "2",
	// Charity
	EventSubTypeChannelCharityCampaignDonate:   "1",
	EventSubTypeChannelCharityCampaignStart:    "1",
	EventSubTypeChannelCharityCampaignProgress: "1",
	EventSubTypeChannelCharityCampaignStop:     "1",
	// Goals
	EventSubTypeChannelGoalBegin:    "1",
	EventSubTypeChannelGoalProgress: "1",
	EventSubTypeChannelGoalEnd:      "1",
	// Shield Mode
	EventSubTypeChannelShieldModeBegin: "1",
	EventSubTypeChannelShieldModeEnd:   "1",
	// Shoutout
	EventSubTypeChannelShoutoutCreate:  "1",
	EventSubTypeChannelShoutoutReceive: "1",
	// Suspicious User
	EventSubTypeChannelSuspiciousUserMessage: "1",
	EventSubTypeChannelSuspiciousUserUpdate:  "1",
	// Guest Star (Beta)
	EventSubTypeChannelGuestStarSessionBegin:   "beta",
	EventSubTypeChannelGuestStarSessionEnd:     "beta",
	EventSubTypeChannelGuestStarGuestUpdate:    "beta",
	EventSubTypeChannelGuestStarSettingsUpdate: "beta",
	// Conduit
	EventSubTypeConduitShardDisabled: "1",
	// Drop
	EventSubTypeDropEntitlementGrant: "1",
	// Extension
	EventSubTypeExtensionBitsTransactionCreate: "1",
	// Stream
	EventSubTypeStreamOnline:  "1",
	EventSubTypeStreamOffline: "1",
	// User
	EventSubTypeUserAuthorizationGrant:  "1",
	EventSubTypeUserAuthorizationRevoke: "1",
	EventSubTypeUserUpdate:              "1",
	EventSubTypeUserWhisperMessage:      "1",
}

// GetEventSubVersion returns the latest version for a subscription type.
// Returns "1" if the type is not found.
func GetEventSubVersion(subType string) string {
	if v, ok := EventSubTypeVersion[subType]; ok {
		return v
	}
	return "1"
}

// EventSubCondition helpers for building common conditions.

// BroadcasterCondition returns a condition with broadcaster_user_id.
func BroadcasterCondition(broadcasterID string) map[string]string {
	return map[string]string{"broadcaster_user_id": broadcasterID}
}

// BroadcasterModeratorCondition returns a condition with broadcaster and moderator IDs.
func BroadcasterModeratorCondition(broadcasterID, moderatorID string) map[string]string {
	return map[string]string{
		"broadcaster_user_id": broadcasterID,
		"moderator_user_id":   moderatorID,
	}
}

// UserCondition returns a condition with user_id.
func UserCondition(userID string) map[string]string {
	return map[string]string{"user_id": userID}
}

// FromToBroadcasterCondition returns a condition for raid events.
func FromToBroadcasterCondition(fromBroadcasterID, toBroadcasterID string) map[string]string {
	cond := make(map[string]string)
	if fromBroadcasterID != "" {
		cond["from_broadcaster_user_id"] = fromBroadcasterID
	}
	if toBroadcasterID != "" {
		cond["to_broadcaster_user_id"] = toBroadcasterID
	}
	return cond
}

// ClientCondition returns a condition with client_id for extension events.
func ClientCondition(clientID string) map[string]string {
	return map[string]string{"client_id": clientID}
}

// ConduitCondition returns a condition with conduit_id.
func ConduitCondition(conduitID string) map[string]string {
	return map[string]string{"conduit_id": conduitID}
}

// RewardCondition returns a condition for channel points reward events.
func RewardCondition(broadcasterID, rewardID string) map[string]string {
	cond := map[string]string{"broadcaster_user_id": broadcasterID}
	if rewardID != "" {
		cond["reward_id"] = rewardID
	}
	return cond
}

// SubscribeToChannel is a helper to subscribe to a channel event using the latest version.
func (c *Client) SubscribeToChannel(ctx context.Context, eventType, broadcasterID string, transport CreateEventSubTransport) (*EventSubSubscription, error) {
	return c.CreateEventSubSubscription(ctx, &CreateEventSubSubscriptionParams{
		Type:      eventType,
		Version:   GetEventSubVersion(eventType),
		Condition: BroadcasterCondition(broadcasterID),
		Transport: transport,
	})
}

// SubscribeToChannelWithModerator is a helper to subscribe to channel events that require moderator.
func (c *Client) SubscribeToChannelWithModerator(ctx context.Context, eventType, broadcasterID, moderatorID string, transport CreateEventSubTransport) (*EventSubSubscription, error) {
	return c.CreateEventSubSubscription(ctx, &CreateEventSubSubscriptionParams{
		Type:      eventType,
		Version:   GetEventSubVersion(eventType),
		Condition: BroadcasterModeratorCondition(broadcasterID, moderatorID),
		Transport: transport,
	})
}

// SubscribeToUser is a helper to subscribe to user events using the latest version.
func (c *Client) SubscribeToUser(ctx context.Context, eventType, userID string, transport CreateEventSubTransport) (*EventSubSubscription, error) {
	return c.CreateEventSubSubscription(ctx, &CreateEventSubSubscriptionParams{
		Type:      eventType,
		Version:   GetEventSubVersion(eventType),
		Condition: UserCondition(userID),
		Transport: transport,
	})
}

// GetAllSubscriptions returns all EventSub subscriptions, handling pagination automatically.
func (c *Client) GetAllSubscriptions(ctx context.Context, params *GetEventSubSubscriptionsParams) ([]EventSubSubscription, error) {
	var all []EventSubSubscription
	if params == nil {
		params = &GetEventSubSubscriptionsParams{}
	}
	if params.PaginationParams == nil {
		params.PaginationParams = &PaginationParams{}
	}

	for {
		resp, err := c.GetEventSubSubscriptions(ctx, params)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if resp.Pagination == nil || resp.Pagination.Cursor == "" {
			break
		}
		params.After = resp.Pagination.Cursor
	}
	return all, nil
}

// DeleteAllSubscriptions deletes all EventSub subscriptions matching the filter.
func (c *Client) DeleteAllSubscriptions(ctx context.Context, params *GetEventSubSubscriptionsParams) (int, error) {
	subs, err := c.GetAllSubscriptions(ctx, params)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, sub := range subs {
		if err := c.DeleteEventSubSubscription(ctx, sub.ID); err != nil {
			return deleted, err
		}
		deleted++
	}
	return deleted, nil
}
