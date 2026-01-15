package helix

import (
	"encoding/json"
	"time"
)

// EventSub Event Types - Common structures that appear in event payloads

// EventSubUser represents a user in EventSub events.
type EventSubUser struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// EventSubBroadcaster represents a broadcaster in EventSub events.
type EventSubBroadcaster struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// EventSubModerator represents a moderator in EventSub events.
type EventSubModerator struct {
	ModeratorUserID    string `json:"moderator_user_id"`
	ModeratorUserLogin string `json:"moderator_user_login"`
	ModeratorUserName  string `json:"moderator_user_name"`
}

// Channel Events

// ChannelUpdateEvent is sent when a broadcaster updates channel properties.
type ChannelUpdateEvent struct {
	EventSubBroadcaster
	Title                       string   `json:"title"`
	Language                    string   `json:"language"`
	CategoryID                  string   `json:"category_id"`
	CategoryName                string   `json:"category_name"`
	ContentClassificationLabels []string `json:"content_classification_labels"`
}

// ChannelFollowEvent is sent when a user follows a channel.
type ChannelFollowEvent struct {
	EventSubUser
	EventSubBroadcaster
	FollowedAt time.Time `json:"followed_at"`
}

// ChannelSubscribeEvent is sent when a user subscribes to a channel.
type ChannelSubscribeEvent struct {
	EventSubUser
	EventSubBroadcaster
	Tier   string `json:"tier"` // 1000, 2000, 3000
	IsGift bool   `json:"is_gift"`
}

// ChannelSubscriptionEndEvent is sent when a subscription ends.
type ChannelSubscriptionEndEvent struct {
	EventSubUser
	EventSubBroadcaster
	Tier   string `json:"tier"`
	IsGift bool   `json:"is_gift"`
}

// ChannelSubscriptionGiftEvent is sent when a user gifts subscriptions.
type ChannelSubscriptionGiftEvent struct {
	EventSubUser
	EventSubBroadcaster
	Total           int    `json:"total"`
	Tier            string `json:"tier"`
	CumulativeTotal int    `json:"cumulative_total,omitempty"` // Only if not anonymous
	IsAnonymous     bool   `json:"is_anonymous"`
}

// ChannelSubscriptionMessageEvent is sent when a user sends a resubscription message.
type ChannelSubscriptionMessageEvent struct {
	EventSubUser
	EventSubBroadcaster
	Tier             string              `json:"tier"`
	Message          SubscriptionMessage `json:"message"`
	CumulativeMonths int                 `json:"cumulative_months"`
	StreakMonths     int                 `json:"streak_months,omitempty"`
	DurationMonths   int                 `json:"duration_months"`
}

// SubscriptionMessage represents the message in a subscription event.
type SubscriptionMessage struct {
	Text   string              `json:"text"`
	Emotes []SubscriptionEmote `json:"emotes,omitempty"`
}

// SubscriptionEmote represents an emote in a subscription message.
type SubscriptionEmote struct {
	Begin int    `json:"begin"`
	End   int    `json:"end"`
	ID    string `json:"id"`
}

// ChannelCheerEvent is sent when a user cheers in a channel.
type ChannelCheerEvent struct {
	IsAnonymous bool   `json:"is_anonymous"`
	UserID      string `json:"user_id,omitempty"` // Empty if anonymous
	UserLogin   string `json:"user_login,omitempty"`
	UserName    string `json:"user_name,omitempty"`
	EventSubBroadcaster
	Message string `json:"message"`
	Bits    int    `json:"bits"`
}

// ChannelRaidEvent is sent when a broadcaster raids another channel.
type ChannelRaidEvent struct {
	FromBroadcasterUserID    string `json:"from_broadcaster_user_id"`
	FromBroadcasterUserLogin string `json:"from_broadcaster_user_login"`
	FromBroadcasterUserName  string `json:"from_broadcaster_user_name"`
	ToBroadcasterUserID      string `json:"to_broadcaster_user_id"`
	ToBroadcasterUserLogin   string `json:"to_broadcaster_user_login"`
	ToBroadcasterUserName    string `json:"to_broadcaster_user_name"`
	Viewers                  int    `json:"viewers"`
}

// ChannelBanEvent is sent when a user is banned from a channel.
type ChannelBanEvent struct {
	EventSubUser
	EventSubBroadcaster
	EventSubModerator
	Reason      string     `json:"reason"`
	BannedAt    time.Time  `json:"banned_at"`
	EndsAt      *time.Time `json:"ends_at,omitempty"` // nil for permanent bans
	IsPermanent bool       `json:"is_permanent"`
}

// ChannelUnbanEvent is sent when a user is unbanned from a channel.
type ChannelUnbanEvent struct {
	EventSubUser
	EventSubBroadcaster
	EventSubModerator
}

// ChannelModeratorAddEvent is sent when a user is added as a moderator.
type ChannelModeratorAddEvent struct {
	EventSubUser
	EventSubBroadcaster
}

// ChannelModeratorRemoveEvent is sent when a user is removed as a moderator.
type ChannelModeratorRemoveEvent struct {
	EventSubUser
	EventSubBroadcaster
}

// Channel Points Events

// ChannelPointsRewardAddEvent is sent when a custom reward is created.
type ChannelPointsRewardAddEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	IsEnabled                        bool                   `json:"is_enabled"`
	IsPaused                         bool                   `json:"is_paused"`
	IsInStock                        bool                   `json:"is_in_stock"`
	Title                            string                 `json:"title"`
	Cost                             int                    `json:"cost"`
	Prompt                           string                 `json:"prompt"`
	IsUserInputRequired              bool                   `json:"is_user_input_required"`
	ShouldRedemptionsSkipQueue       bool                   `json:"should_redemptions_skip_request_queue"`
	MaxPerStream                     EventSubMaxPerStream   `json:"max_per_stream"`
	MaxPerUserPerStream              EventSubMaxPerStream   `json:"max_per_user_per_stream"`
	BackgroundColor                  string                 `json:"background_color"`
	Image                            *EventSubImage         `json:"image"`
	DefaultImage                     EventSubImage          `json:"default_image"`
	GlobalCooldown                   EventSubGlobalCooldown `json:"global_cooldown"`
	CooldownExpiresAt                *time.Time             `json:"cooldown_expires_at"`
	RedemptionsRedeemedCurrentStream *int                   `json:"redemptions_redeemed_current_stream"`
}

// EventSubMaxPerStream represents max redemptions per stream settings.
type EventSubMaxPerStream struct {
	IsEnabled bool `json:"is_enabled"`
	Value     int  `json:"value"`
}

// EventSubGlobalCooldown represents global cooldown settings.
type EventSubGlobalCooldown struct {
	IsEnabled bool `json:"is_enabled"`
	Seconds   int  `json:"seconds"`
}

// EventSubImage represents image URLs for rewards.
type EventSubImage struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

// ChannelPointsRewardUpdateEvent is sent when a custom reward is updated.
type ChannelPointsRewardUpdateEvent = ChannelPointsRewardAddEvent

// ChannelPointsRewardRemoveEvent is sent when a custom reward is removed.
type ChannelPointsRewardRemoveEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	IsEnabled                        bool                   `json:"is_enabled"`
	IsPaused                         bool                   `json:"is_paused"`
	IsInStock                        bool                   `json:"is_in_stock"`
	Title                            string                 `json:"title"`
	Cost                             int                    `json:"cost"`
	Prompt                           string                 `json:"prompt"`
	IsUserInputRequired              bool                   `json:"is_user_input_required"`
	ShouldRedemptionsSkipQueue       bool                   `json:"should_redemptions_skip_request_queue"`
	CooldownExpiresAt                *time.Time             `json:"cooldown_expires_at"`
	RedemptionsRedeemedCurrentStream *int                   `json:"redemptions_redeemed_current_stream"`
	MaxPerStream                     EventSubMaxPerStream   `json:"max_per_stream"`
	MaxPerUserPerStream              EventSubMaxPerStream   `json:"max_per_user_per_stream"`
	GlobalCooldown                   EventSubGlobalCooldown `json:"global_cooldown"`
	BackgroundColor                  string                 `json:"background_color"`
	Image                            *EventSubImage         `json:"image"`
	DefaultImage                     EventSubImage          `json:"default_image"`
}

// ChannelPointsRedemptionAddEvent is sent when a reward is redeemed.
type ChannelPointsRedemptionAddEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	EventSubUser
	UserInput  string         `json:"user_input"`
	Status     string         `json:"status"` // unfulfilled, fulfilled, canceled
	Reward     EventSubReward `json:"reward"`
	RedeemedAt time.Time      `json:"redeemed_at"`
}

// EventSubReward represents reward details in redemption events.
type EventSubReward struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Cost   int    `json:"cost"`
	Prompt string `json:"prompt"`
}

// ChannelPointsRedemptionUpdateEvent is sent when a redemption is updated.
type ChannelPointsRedemptionUpdateEvent = ChannelPointsRedemptionAddEvent

// Poll Events

// ChannelPollBeginEvent is sent when a poll begins.
type ChannelPollBeginEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Title               string           `json:"title"`
	Choices             []EventSubChoice `json:"choices"`
	BitsVoting          EventSubVoting   `json:"bits_voting"`
	ChannelPointsVoting EventSubVoting   `json:"channel_points_voting"`
	StartedAt           time.Time        `json:"started_at"`
	EndsAt              time.Time        `json:"ends_at"`
}

// EventSubChoice represents a poll choice.
type EventSubChoice struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	BitsVotes          int    `json:"bits_votes"`
	ChannelPointsVotes int    `json:"channel_points_votes"`
	Votes              int    `json:"votes"`
}

// EventSubVoting represents voting settings.
type EventSubVoting struct {
	IsEnabled     bool `json:"is_enabled"`
	AmountPerVote int  `json:"amount_per_vote"`
}

// ChannelPollProgressEvent is sent when poll results update.
type ChannelPollProgressEvent = ChannelPollBeginEvent

// ChannelPollEndEvent is sent when a poll ends.
type ChannelPollEndEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Title               string           `json:"title"`
	Choices             []EventSubChoice `json:"choices"`
	BitsVoting          EventSubVoting   `json:"bits_voting"`
	ChannelPointsVoting EventSubVoting   `json:"channel_points_voting"`
	Status              string           `json:"status"` // completed, archived, terminated
	StartedAt           time.Time        `json:"started_at"`
	EndedAt             time.Time        `json:"ended_at"`
}

// Prediction Events

// ChannelPredictionBeginEvent is sent when a prediction begins.
type ChannelPredictionBeginEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Title     string            `json:"title"`
	Outcomes  []EventSubOutcome `json:"outcomes"`
	StartedAt time.Time         `json:"started_at"`
	LocksAt   time.Time         `json:"locks_at"`
}

// EventSubOutcome represents a prediction outcome.
type EventSubOutcome struct {
	ID            string              `json:"id"`
	Title         string              `json:"title"`
	Color         string              `json:"color"` // blue, pink
	Users         int                 `json:"users,omitempty"`
	ChannelPoints int                 `json:"channel_points,omitempty"`
	TopPredictors []EventSubPredictor `json:"top_predictors,omitempty"`
}

// EventSubPredictor represents a top predictor.
type EventSubPredictor struct {
	EventSubUser
	ChannelPointsUsed int  `json:"channel_points_used"`
	ChannelPointsWon  *int `json:"channel_points_won,omitempty"`
}

// ChannelPredictionProgressEvent is sent when prediction totals update.
type ChannelPredictionProgressEvent = ChannelPredictionBeginEvent

// ChannelPredictionLockEvent is sent when a prediction locks.
type ChannelPredictionLockEvent = ChannelPredictionBeginEvent

// ChannelPredictionEndEvent is sent when a prediction ends.
type ChannelPredictionEndEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Title            string            `json:"title"`
	WinningOutcomeID string            `json:"winning_outcome_id,omitempty"`
	Outcomes         []EventSubOutcome `json:"outcomes"`
	Status           string            `json:"status"` // resolved, canceled
	StartedAt        time.Time         `json:"started_at"`
	EndedAt          time.Time         `json:"ended_at"`
}

// Hype Train Events

// HypeTrainType represents the type of hype train (v2 only).
type HypeTrainType string

const (
	HypeTrainTypeRegular    HypeTrainType = "regular"
	HypeTrainTypeGoldenKappa HypeTrainType = "golden_kappa"
	HypeTrainTypeShared     HypeTrainType = "shared"
)

// HypeTrainParticipant represents a participant in a shared hype train (v2 only).
type HypeTrainParticipant struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
}

// ChannelHypeTrainBeginEvent is sent when a Hype Train begins.
// Note: Hype Train v1 is deprecated by Twitch. Use v2 fields (Type, IsSharedTrain, etc.).
type ChannelHypeTrainBeginEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Total            int                    `json:"total"`
	Progress         int                    `json:"progress"`
	Goal             int                    `json:"goal"`
	TopContributions []EventSubContribution `json:"top_contributions"`
	LastContribution EventSubContribution   `json:"last_contribution"`
	Level            int                    `json:"level"`
	StartedAt        time.Time              `json:"started_at"`
	ExpiresAt        time.Time              `json:"expires_at"`
	// Deprecated: IsGoldenKappaTrain is from v1 which is deprecated. Use Type == HypeTrainTypeGoldenKappa instead.
	// This field is auto-populated from Type during unmarshaling for migration convenience.
	IsGoldenKappaTrain bool `json:"is_golden_kappa_train,omitempty"`
	// V2 fields
	Type                    HypeTrainType          `json:"type,omitempty"`
	IsSharedTrain           bool                   `json:"is_shared_train,omitempty"`
	SharedTrainParticipants []HypeTrainParticipant `json:"shared_train_participants,omitempty"`
	AllTimeHighLevel        int                    `json:"all_time_high_level,omitempty"`
	AllTimeHighTotal        int                    `json:"all_time_high_total,omitempty"`
}

// EventSubContribution represents a Hype Train contribution.
type EventSubContribution struct {
	EventSubUser
	Type  string `json:"type"` // bits, subscription, other
	Total int    `json:"total"`
}

// ChannelHypeTrainProgressEvent is sent when a Hype Train progresses.
type ChannelHypeTrainProgressEvent = ChannelHypeTrainBeginEvent

// ChannelHypeTrainEndEvent is sent when a Hype Train ends.
// Note: Hype Train v1 is deprecated by Twitch. Use v2 fields (Type, IsSharedTrain, etc.).
type ChannelHypeTrainEndEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Level            int                    `json:"level"`
	Total            int                    `json:"total"`
	TopContributions []EventSubContribution `json:"top_contributions"`
	StartedAt        time.Time              `json:"started_at"`
	EndedAt          time.Time              `json:"ended_at"`
	CooldownEndsAt   time.Time              `json:"cooldown_ends_at"`
	// Deprecated: IsGoldenKappaTrain is from v1 which is deprecated. Use Type == HypeTrainTypeGoldenKappa instead.
	// This field is auto-populated from Type during unmarshaling for migration convenience.
	IsGoldenKappaTrain bool `json:"is_golden_kappa_train,omitempty"`
	// V2 fields
	Type                    HypeTrainType          `json:"type,omitempty"`
	IsSharedTrain           bool                   `json:"is_shared_train,omitempty"`
	SharedTrainParticipants []HypeTrainParticipant `json:"shared_train_participants,omitempty"`
}

// UnmarshalJSON implements automatic v1/v2 field conversion for ChannelHypeTrainBeginEvent.
// When receiving v1 events, it populates Type from IsGoldenKappaTrain.
// When receiving v2 events, it populates IsGoldenKappaTrain from Type.
func (e *ChannelHypeTrainBeginEvent) UnmarshalJSON(data []byte) error {
	type Alias ChannelHypeTrainBeginEvent
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Convert v1 -> v2: if Type is empty but IsGoldenKappaTrain is set
	if e.Type == "" {
		if e.IsGoldenKappaTrain {
			e.Type = HypeTrainTypeGoldenKappa
		} else {
			e.Type = HypeTrainTypeRegular
		}
	}

	// Convert v2 -> v1: if Type is set, populate IsGoldenKappaTrain
	if e.Type == HypeTrainTypeGoldenKappa {
		e.IsGoldenKappaTrain = true
	}

	return nil
}

// UnmarshalJSON implements automatic v1/v2 field conversion for ChannelHypeTrainEndEvent.
// When receiving v1 events, it populates Type from IsGoldenKappaTrain.
// When receiving v2 events, it populates IsGoldenKappaTrain from Type.
func (e *ChannelHypeTrainEndEvent) UnmarshalJSON(data []byte) error {
	type Alias ChannelHypeTrainEndEvent
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Convert v1 -> v2: if Type is empty but IsGoldenKappaTrain is set
	if e.Type == "" {
		if e.IsGoldenKappaTrain {
			e.Type = HypeTrainTypeGoldenKappa
		} else {
			e.Type = HypeTrainTypeRegular
		}
	}

	// Convert v2 -> v1: if Type is set, populate IsGoldenKappaTrain
	if e.Type == HypeTrainTypeGoldenKappa {
		e.IsGoldenKappaTrain = true
	}

	return nil
}

// V1 compatibility type aliases.
// These are provided for users who need to work with v1 payloads explicitly.

// ChannelHypeTrainBeginEventV1 is the v1 version of the hype train begin event.
// Deprecated: Use ChannelHypeTrainBeginEvent with v2 subscription instead.
type ChannelHypeTrainBeginEventV1 = ChannelHypeTrainBeginEvent

// ChannelHypeTrainProgressEventV1 is the v1 version of the hype train progress event.
// Deprecated: Use ChannelHypeTrainProgressEvent with v2 subscription instead.
type ChannelHypeTrainProgressEventV1 = ChannelHypeTrainProgressEvent

// ChannelHypeTrainEndEventV1 is the v1 version of the hype train end event.
// Deprecated: Use ChannelHypeTrainEndEvent with v2 subscription instead.
type ChannelHypeTrainEndEventV1 = ChannelHypeTrainEndEvent

// Stream Events

// StreamOnlineEvent is sent when a stream goes online.
type StreamOnlineEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Type      string    `json:"type"` // live, playlist, watch_party, premiere, rerun
	StartedAt time.Time `json:"started_at"`
}

// StreamOfflineEvent is sent when a stream goes offline.
type StreamOfflineEvent struct {
	EventSubBroadcaster
}

// User Events

// UserUpdateEvent is sent when a user updates their account.
type UserUpdateEvent struct {
	UserID        string `json:"user_id"`
	UserLogin     string `json:"user_login"`
	UserName      string `json:"user_name"`
	Email         string `json:"email,omitempty"` // Requires user:read:email scope
	EmailVerified bool   `json:"email_verified,omitempty"`
	Description   string `json:"description"`
}

// Chat Events (for EventSub, not IRC)

// ChannelChatMessageEvent is sent when a message is sent in chat.
type ChannelChatMessageEvent struct {
	EventSubBroadcaster
	ChatterUserID               string           `json:"chatter_user_id"`
	ChatterUserLogin            string           `json:"chatter_user_login"`
	ChatterUserName             string           `json:"chatter_user_name"`
	MessageID                   string           `json:"message_id"`
	Message                     ChatEventMessage `json:"message"`
	Color                       string           `json:"color"`
	Badges                      []ChatEventBadge `json:"badges"`
	MessageType                 string           `json:"message_type"` // text, channel_points_highlighted, channel_points_sub_only, user_intro, power_ups_message_effect, power_ups_gigantified_emote
	Cheer                       *ChatEventCheer  `json:"cheer,omitempty"`
	Reply                       *ChatEventReply  `json:"reply,omitempty"`
	ChannelPointsCustomRewardID string           `json:"channel_points_custom_reward_id,omitempty"`
	// Shared chat fields (null if not in shared chat or same channel)
	SourceBroadcasterUserID    *string          `json:"source_broadcaster_user_id,omitempty"`
	SourceBroadcasterUserLogin *string          `json:"source_broadcaster_user_login,omitempty"`
	SourceBroadcasterUserName  *string          `json:"source_broadcaster_user_name,omitempty"`
	SourceMessageID            *string          `json:"source_message_id,omitempty"`
	SourceBadges               []ChatEventBadge `json:"source_badges,omitempty"`
}

// ChatEventMessage represents a chat message structure.
type ChatEventMessage struct {
	Text      string              `json:"text"`
	Fragments []ChatEventFragment `json:"fragments"`
}

// ChatEventFragment represents a fragment of a chat message.
type ChatEventFragment struct {
	Type      string              `json:"type"` // text, cheermote, emote, mention
	Text      string              `json:"text"`
	Cheermote *ChatEventCheermote `json:"cheermote,omitempty"`
	Emote     *ChatEventEmote     `json:"emote,omitempty"`
	Mention   *ChatEventMention   `json:"mention,omitempty"`
}

// ChatEventCheermote represents a cheermote in a message.
type ChatEventCheermote struct {
	Prefix string `json:"prefix"`
	Bits   int    `json:"bits"`
	Tier   int    `json:"tier"`
}

// ChatEventEmote represents an emote in a message.
type ChatEventEmote struct {
	ID         string   `json:"id"`
	EmoteSetID string   `json:"emote_set_id"`
	OwnerID    string   `json:"owner_id"`
	Format     []string `json:"format"` // static, animated
}

// ChatEventMention represents a mention in a message.
type ChatEventMention struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// ChatEventBadge represents a chat badge.
type ChatEventBadge struct {
	SetID string `json:"set_id"`
	ID    string `json:"id"`
	Info  string `json:"info"`
}

// ChatEventCheer represents cheer info in a message.
type ChatEventCheer struct {
	Bits int `json:"bits"`
}

// ChatEventReply represents reply info in a message.
type ChatEventReply struct {
	ParentMessageID   string `json:"parent_message_id"`
	ParentMessageBody string `json:"parent_message_body"`
	ParentUserID      string `json:"parent_user_id"`
	ParentUserLogin   string `json:"parent_user_login"`
	ParentUserName    string `json:"parent_user_name"`
	ThreadMessageID   string `json:"thread_message_id"`
	ThreadUserID      string `json:"thread_user_id"`
	ThreadUserLogin   string `json:"thread_user_login"`
	ThreadUserName    string `json:"thread_user_name"`
}

// Automod Events

// AutomodMessageHoldEvent is sent when AutoMod holds a message for review.
// V2 adds reason, automod, and blocked_term fields.
type AutomodMessageHoldEvent struct {
	EventSubBroadcaster
	EventSubUser
	MessageID string             `json:"message_id"`
	Message   AutomodHeldMessage `json:"message"`
	HeldAt    time.Time          `json:"held_at"`
	// V2 fields
	Reason      string               `json:"reason,omitempty"` // "automod" or "blocked_term"
	Automod     *AutomodCategoryInfo `json:"automod,omitempty"`
	BlockedTerm *BlockedTermInfo     `json:"blocked_term,omitempty"`
	// V1 fields (deprecated in v2)
	Level     int               `json:"level,omitempty"`
	Category  string            `json:"category,omitempty"`
	Fragments []AutomodFragment `json:"fragments,omitempty"`
}

// AutomodHeldMessage represents the message structure in automod events.
type AutomodHeldMessage struct {
	Text      string            `json:"text"`
	Fragments []AutomodFragment `json:"fragments"`
}

// AutomodCategoryInfo contains automod category information (v2).
type AutomodCategoryInfo struct {
	Category   string            `json:"category"`
	Level      int               `json:"level"`
	Boundaries []AutomodBoundary `json:"boundaries"`
}

// AutomodBoundary represents text boundaries flagged by automod.
type AutomodBoundary struct {
	StartPos int `json:"start_pos"`
	EndPos   int `json:"end_pos"`
}

// BlockedTermInfo contains blocked term information (v2).
type BlockedTermInfo struct {
	TermsFound                []BlockedTermFound `json:"terms_found"`
	OwnerBroadcasterUserID    string             `json:"owner_broadcaster_user_id"`
	OwnerBroadcasterUserLogin string             `json:"owner_broadcaster_user_login"`
	OwnerBroadcasterUserName  string             `json:"owner_broadcaster_user_name"`
}

// BlockedTermFound represents a found blocked term.
type BlockedTermFound struct {
	TermID   string          `json:"term_id"`
	Boundary AutomodBoundary `json:"boundary"`
}

// AutomodFragment represents a fragment in an automod message.
type AutomodFragment struct {
	Type      string              `json:"type"`
	Text      string              `json:"text"`
	Automod   *AutomodDetails     `json:"automod,omitempty"`
	Emote     *ChatEventEmote     `json:"emote,omitempty"`
	Cheermote *ChatEventCheermote `json:"cheermote,omitempty"`
}

// AutomodDetails contains automod-specific details.
type AutomodDetails struct {
	Topics []AutomodTopic `json:"topics"`
}

// AutomodTopic represents a topic flagged by automod.
type AutomodTopic struct {
	Type  string `json:"type"`
	Score int    `json:"score"`
}

// AutomodMessageUpdateEvent is sent when a held message's status is updated.
type AutomodMessageUpdateEvent struct {
	EventSubBroadcaster
	EventSubUser
	EventSubModerator
	MessageID string `json:"message_id"`
	Message   string `json:"message"`
	Status    string `json:"status"` // approved, denied, expired
}

// AutomodSettingsUpdateEvent is sent when automod settings are updated.
type AutomodSettingsUpdateEvent struct {
	EventSubBroadcaster
	EventSubModerator
	BulliedUsers          int  `json:"bullying"`
	Disability            int  `json:"disability"`
	Misogyny              int  `json:"misogyny"`
	OverallLevel          *int `json:"overall_level"`
	RaceEthnicityReligion int  `json:"race_ethnicity_or_religion"`
	SexBasedTerms         int  `json:"sex_based_terms"`
	SexualitySexGender    int  `json:"sexuality_sex_or_gender"`
	Swearing              int  `json:"swearing"`
}

// AutomodTermsUpdateEvent is sent when automod terms are updated.
type AutomodTermsUpdateEvent struct {
	EventSubBroadcaster
	EventSubModerator
	Action      string   `json:"action"` // add_permitted, remove_permitted, add_blocked, remove_blocked
	FromAutomod bool     `json:"from_automod"`
	Terms       []string `json:"terms"`
}

// Ad Break Events

// ChannelAdBreakBeginEvent is sent when an ad break begins.
type ChannelAdBreakBeginEvent struct {
	DurationSeconds int       `json:"duration_seconds"`
	StartedAt       time.Time `json:"started_at"`
	IsAutomatic     bool      `json:"is_automatic"`
	EventSubBroadcaster
	RequesterUserID    string `json:"requester_user_id"`
	RequesterUserLogin string `json:"requester_user_login"`
	RequesterUserName  string `json:"requester_user_name"`
}

// Bits Events

// ChannelBitsUseEvent is sent when bits are used in a channel.
type ChannelBitsUseEvent struct {
	EventSubBroadcaster
	EventSubUser
	BitsUsed int       `json:"bits_used"`
	Type     string    `json:"type"` // cheer, power_up_celebration, power_up_gigantify, power_up_message_effect
	UsedAt   time.Time `json:"used_at"`
	Message  *string   `json:"message,omitempty"`
	PowerUp  *PowerUp  `json:"power_up,omitempty"`
}

// PowerUp represents a bits power-up.
type PowerUp struct {
	Type string `json:"type"`
}

// VIP Events

// ChannelVIPAddEvent is sent when a user is added as a VIP.
type ChannelVIPAddEvent struct {
	EventSubUser
	EventSubBroadcaster
}

// ChannelVIPRemoveEvent is sent when a user is removed as a VIP.
type ChannelVIPRemoveEvent struct {
	EventSubUser
	EventSubBroadcaster
}

// Warning Events

// ChannelWarningSendEvent is sent when a warning is sent to a user.
type ChannelWarningSendEvent struct {
	EventSubBroadcaster
	EventSubModerator
	EventSubUser
	Reason         *string  `json:"reason,omitempty"`
	ChatRulesCited []string `json:"chat_rules_cited,omitempty"`
}

// ChannelWarningAcknowledgeEvent is sent when a warning is acknowledged.
type ChannelWarningAcknowledgeEvent struct {
	EventSubBroadcaster
	EventSubUser
}

// Unban Request Events

// ChannelUnbanRequestCreateEvent is sent when an unban request is created.
type ChannelUnbanRequestCreateEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	EventSubUser
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// ChannelUnbanRequestResolveEvent is sent when an unban request is resolved.
type ChannelUnbanRequestResolveEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	EventSubModerator
	EventSubUser
	ResolutionText string `json:"resolution_text,omitempty"`
	Status         string `json:"status"` // approved, canceled, denied
}

// Moderate Event

// ChannelModerateEvent is sent for various moderation actions.
type ChannelModerateEvent struct {
	EventSubBroadcaster
	EventSubModerator
	Action            string                     `json:"action"`
	Followers         *ModerateFollowers         `json:"followers,omitempty"`
	Slow              *ModerateSlow              `json:"slow,omitempty"`
	Vip               *ModerateUser              `json:"vip,omitempty"`
	Unvip             *ModerateUser              `json:"unvip,omitempty"`
	Mod               *ModerateUser              `json:"mod,omitempty"`
	Unmod             *ModerateUser              `json:"unmod,omitempty"`
	Ban               *ModerateBan               `json:"ban,omitempty"`
	Unban             *ModerateUser              `json:"unban,omitempty"`
	Timeout           *ModerateTimeout           `json:"timeout,omitempty"`
	Untimeout         *ModerateUser              `json:"untimeout,omitempty"`
	Raid              *ModerateRaid              `json:"raid,omitempty"`
	Unraid            *ModerateRaid              `json:"unraid,omitempty"`
	Delete            *ModerateDelete            `json:"delete,omitempty"`
	AutomodTerms      *ModerateAutomodTerms      `json:"automod_terms,omitempty"`
	UnbanRequest      *ModerateUnbanRequest      `json:"unban_request,omitempty"`
	Warn              *ModerateWarn              `json:"warn,omitempty"`
	SharedChatBan     *ModerateSharedChatBan     `json:"shared_chat_ban,omitempty"`
	SharedChatTimeout *ModerateSharedChatTimeout `json:"shared_chat_timeout,omitempty"`
	SharedChatDelete  *ModerateSharedChatDelete  `json:"shared_chat_delete,omitempty"`
}

// ModerateFollowers represents followers-only mode settings.
type ModerateFollowers struct {
	FollowDurationMinutes int `json:"follow_duration_minutes"`
}

// ModerateSlow represents slow mode settings.
type ModerateSlow struct {
	WaitTimeSeconds int `json:"wait_time_seconds"`
}

// ModerateUser represents a user in moderation actions.
type ModerateUser struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// ModerateBan represents ban details.
type ModerateBan struct {
	UserID    string  `json:"user_id"`
	UserLogin string  `json:"user_login"`
	UserName  string  `json:"user_name"`
	Reason    *string `json:"reason,omitempty"`
}

// ModerateTimeout represents timeout details.
type ModerateTimeout struct {
	UserID    string    `json:"user_id"`
	UserLogin string    `json:"user_login"`
	UserName  string    `json:"user_name"`
	Reason    *string   `json:"reason,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ModerateRaid represents raid details.
type ModerateRaid struct {
	UserID      string `json:"user_id"`
	UserLogin   string `json:"user_login"`
	UserName    string `json:"user_name"`
	ViewerCount int    `json:"viewer_count"`
}

// ModerateDelete represents message deletion details.
type ModerateDelete struct {
	UserID      string `json:"user_id"`
	UserLogin   string `json:"user_login"`
	UserName    string `json:"user_name"`
	MessageID   string `json:"message_id"`
	MessageBody string `json:"message_body"`
}

// ModerateAutomodTerms represents automod terms update.
type ModerateAutomodTerms struct {
	Action      string   `json:"action"`
	List        string   `json:"list"`
	Terms       []string `json:"terms"`
	FromAutomod bool     `json:"from_automod"`
}

// ModerateUnbanRequest represents unban request resolution.
type ModerateUnbanRequest struct {
	IsApproved       bool    `json:"is_approved"`
	UserID           string  `json:"user_id"`
	UserLogin        string  `json:"user_login"`
	UserName         string  `json:"user_name"`
	ModeratorMessage *string `json:"moderator_message,omitempty"`
}

// ModerateWarn represents warning details.
type ModerateWarn struct {
	UserID         string   `json:"user_id"`
	UserLogin      string   `json:"user_login"`
	UserName       string   `json:"user_name"`
	Reason         *string  `json:"reason,omitempty"`
	ChatRulesCited []string `json:"chat_rules_cited,omitempty"`
}

// ModerateSharedChatBan represents a ban in shared chat.
type ModerateSharedChatBan struct {
	UserID    string  `json:"user_id"`
	UserLogin string  `json:"user_login"`
	UserName  string  `json:"user_name"`
	Reason    *string `json:"reason,omitempty"`
}

// ModerateSharedChatTimeout represents a timeout in shared chat.
type ModerateSharedChatTimeout struct {
	UserID    string    `json:"user_id"`
	UserLogin string    `json:"user_login"`
	UserName  string    `json:"user_name"`
	Reason    *string   `json:"reason,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ModerateSharedChatDelete represents message deletion in shared chat.
type ModerateSharedChatDelete struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	MessageID string `json:"message_id"`
}

// Shared Chat Events

// ChannelSharedChatBeginEvent is sent when a shared chat session begins.
type ChannelSharedChatBeginEvent struct {
	SessionID string `json:"session_id"`
	EventSubBroadcaster
	HostBroadcasterUserID    string                  `json:"host_broadcaster_user_id"`
	HostBroadcasterUserLogin string                  `json:"host_broadcaster_user_login"`
	HostBroadcasterUserName  string                  `json:"host_broadcaster_user_name"`
	Participants             []SharedChatParticipant `json:"participants"`
}

// ChannelSharedChatUpdateEvent is sent when a shared chat session is updated.
type ChannelSharedChatUpdateEvent = ChannelSharedChatBeginEvent

// ChannelSharedChatEndEvent is sent when a shared chat session ends.
type ChannelSharedChatEndEvent struct {
	SessionID string `json:"session_id"`
	EventSubBroadcaster
	HostBroadcasterUserID    string `json:"host_broadcaster_user_id"`
	HostBroadcasterUserLogin string `json:"host_broadcaster_user_login"`
	HostBroadcasterUserName  string `json:"host_broadcaster_user_name"`
}

// Charity Events

// ChannelCharityDonationEvent is sent when a charity donation is made.
type ChannelCharityDonationEvent struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	EventSubBroadcaster
	EventSubUser
	CharityName        string        `json:"charity_name"`
	CharityDescription string        `json:"charity_description"`
	CharityLogo        string        `json:"charity_logo"`
	CharityWebsite     string        `json:"charity_website"`
	Amount             CharityAmount `json:"amount"`
}

// ChannelCharityCampaignStartEvent is sent when a charity campaign starts.
type ChannelCharityCampaignStartEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	CharityName        string        `json:"charity_name"`
	CharityDescription string        `json:"charity_description"`
	CharityLogo        string        `json:"charity_logo"`
	CharityWebsite     string        `json:"charity_website"`
	CurrentAmount      CharityAmount `json:"current_amount"`
	TargetAmount       CharityAmount `json:"target_amount"`
	StartedAt          time.Time     `json:"started_at"`
}

// ChannelCharityCampaignProgressEvent is sent when charity campaign progress updates.
type ChannelCharityCampaignProgressEvent = ChannelCharityCampaignStartEvent

// ChannelCharityCampaignStopEvent is sent when a charity campaign stops.
type ChannelCharityCampaignStopEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	CharityName        string        `json:"charity_name"`
	CharityDescription string        `json:"charity_description"`
	CharityLogo        string        `json:"charity_logo"`
	CharityWebsite     string        `json:"charity_website"`
	CurrentAmount      CharityAmount `json:"current_amount"`
	TargetAmount       CharityAmount `json:"target_amount"`
	StoppedAt          time.Time     `json:"stopped_at"`
}

// Goal Events

// ChannelGoalBeginEvent is sent when a goal begins.
type ChannelGoalBeginEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Type          string    `json:"type"` // follower, subscription, subscription_count, new_subscription, new_subscription_count, new_bit, new_cheerer
	Description   string    `json:"description"`
	CurrentAmount int       `json:"current_amount"`
	TargetAmount  int       `json:"target_amount"`
	StartedAt     time.Time `json:"started_at"`
}

// ChannelGoalProgressEvent is sent when goal progress updates.
type ChannelGoalProgressEvent = ChannelGoalBeginEvent

// ChannelGoalEndEvent is sent when a goal ends.
type ChannelGoalEndEvent struct {
	ID string `json:"id"`
	EventSubBroadcaster
	Type          string    `json:"type"`
	Description   string    `json:"description"`
	CurrentAmount int       `json:"current_amount"`
	TargetAmount  int       `json:"target_amount"`
	StartedAt     time.Time `json:"started_at"`
	EndedAt       time.Time `json:"ended_at"`
	IsAchieved    bool      `json:"is_achieved"`
}

// Shield Mode Events

// ChannelShieldModeBeginEvent is sent when shield mode begins.
type ChannelShieldModeBeginEvent struct {
	EventSubBroadcaster
	EventSubModerator
	StartedAt time.Time `json:"started_at"`
}

// ChannelShieldModeEndEvent is sent when shield mode ends.
type ChannelShieldModeEndEvent struct {
	EventSubBroadcaster
	EventSubModerator
	EndedAt time.Time `json:"ended_at"`
}

// Shoutout Events

// ChannelShoutoutCreateEvent is sent when a shoutout is created.
type ChannelShoutoutCreateEvent struct {
	EventSubBroadcaster
	EventSubModerator
	ToBroadcasterUserID    string    `json:"to_broadcaster_user_id"`
	ToBroadcasterUserLogin string    `json:"to_broadcaster_user_login"`
	ToBroadcasterUserName  string    `json:"to_broadcaster_user_name"`
	ViewerCount            int       `json:"viewer_count"`
	StartedAt              time.Time `json:"started_at"`
	CooldownEndsAt         time.Time `json:"cooldown_ends_at"`
	TargetCooldownEndsAt   time.Time `json:"target_cooldown_ends_at"`
}

// ChannelShoutoutReceiveEvent is sent when a shoutout is received.
type ChannelShoutoutReceiveEvent struct {
	EventSubBroadcaster
	FromBroadcasterUserID    string    `json:"from_broadcaster_user_id"`
	FromBroadcasterUserLogin string    `json:"from_broadcaster_user_login"`
	FromBroadcasterUserName  string    `json:"from_broadcaster_user_name"`
	ViewerCount              int       `json:"viewer_count"`
	StartedAt                time.Time `json:"started_at"`
}

// Suspicious User Events

// ChannelSuspiciousUserMessageEvent is sent when a suspicious user sends a message.
type ChannelSuspiciousUserMessageEvent struct {
	EventSubBroadcaster
	EventSubUser
	LowTrustStatus       string                `json:"low_trust_status"` // none, active_monitoring, restricted
	SharedBanChannelIDs  []string              `json:"shared_ban_channel_ids,omitempty"`
	Types                []string              `json:"types"`                  // manual, ban_evader_detector, shared_channel_ban
	BanEvasionEvaluation string                `json:"ban_evasion_evaluation"` // unknown, possible, likely
	Message              SuspiciousUserMessage `json:"message"`
}

// SuspiciousUserMessage represents a message from a suspicious user.
type SuspiciousUserMessage struct {
	MessageID string               `json:"message_id"`
	Text      string               `json:"text"`
	Fragments []SuspiciousFragment `json:"fragments"`
}

// SuspiciousFragment represents a fragment in a suspicious user message.
type SuspiciousFragment struct {
	Type      string              `json:"type"`
	Text      string              `json:"text"`
	Cheermote *ChatEventCheermote `json:"cheermote,omitempty"`
	Emote     *ChatEventEmote     `json:"emote,omitempty"`
}

// ChannelSuspiciousUserUpdateEvent is sent when a suspicious user's status is updated.
type ChannelSuspiciousUserUpdateEvent struct {
	EventSubBroadcaster
	EventSubModerator
	EventSubUser
	LowTrustStatus string `json:"low_trust_status"`
}

// Conduit Events

// ConduitShardDisabledEvent is sent when a conduit shard is disabled.
type ConduitShardDisabledEvent struct {
	ConduitID string                     `json:"conduit_id"`
	ShardID   string                     `json:"shard_id"`
	Status    string                     `json:"status"`
	Transport ConduitShardTransportEvent `json:"transport"`
}

// ConduitShardTransportEvent represents transport info in conduit events.
type ConduitShardTransportEvent struct {
	Method         string `json:"method"`
	SessionID      string `json:"session_id,omitempty"`
	Callback       string `json:"callback,omitempty"`
	ConnectedAt    string `json:"connected_at,omitempty"`
	DisconnectedAt string `json:"disconnected_at,omitempty"`
}

// Drop Events

// DropEntitlementGrantEvent is sent when a drop entitlement is granted.
type DropEntitlementGrantEvent struct {
	ID   string            `json:"id"`
	Data []DropEntitlement `json:"data"`
}

// DropEntitlement represents a single drop entitlement.
type DropEntitlement struct {
	OrganizationID string    `json:"organization_id"`
	CategoryID     string    `json:"category_id"`
	CategoryName   string    `json:"category_name"`
	CampaignID     string    `json:"campaign_id"`
	UserID         string    `json:"user_id"`
	UserLogin      string    `json:"user_login"`
	UserName       string    `json:"user_name"`
	EntitlementID  string    `json:"entitlement_id"`
	BenefitID      string    `json:"benefit_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// Extension Events

// ExtensionBitsTransactionCreateEvent is sent when an extension bits transaction occurs.
type ExtensionBitsTransactionCreateEvent struct {
	ID                string `json:"id"`
	ExtensionClientID string `json:"extension_client_id"`
	EventSubBroadcaster
	EventSubUser
	Product ExtensionProduct `json:"product"`
}

// ExtensionProduct represents an extension product.
type ExtensionProduct struct {
	Name          string `json:"name"`
	Sku           string `json:"sku"`
	Bits          int    `json:"bits"`
	InDevelopment bool   `json:"in_development"`
}

// User Authorization Events

// UserAuthorizationGrantEvent is sent when a user grants authorization.
type UserAuthorizationGrantEvent struct {
	ClientID  string `json:"client_id"`
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// UserAuthorizationRevokeEvent is sent when a user revokes authorization.
type UserAuthorizationRevokeEvent struct {
	ClientID  string  `json:"client_id"`
	UserID    string  `json:"user_id"`
	UserLogin *string `json:"user_login,omitempty"`
	UserName  *string `json:"user_name,omitempty"`
}

// Whisper Events

// UserWhisperMessageEvent is sent when a whisper is received.
type UserWhisperMessageEvent struct {
	FromUserID    string      `json:"from_user_id"`
	FromUserLogin string      `json:"from_user_login"`
	FromUserName  string      `json:"from_user_name"`
	ToUserID      string      `json:"to_user_id"`
	ToUserLogin   string      `json:"to_user_login"`
	ToUserName    string      `json:"to_user_name"`
	WhisperID     string      `json:"whisper_id"`
	Whisper       WhisperBody `json:"whisper"`
}

// WhisperBody represents the whisper message body.
type WhisperBody struct {
	Text string `json:"text"`
}

// Chat Notification Events

// ChannelChatNotificationEvent is sent for various chat notifications.
type ChannelChatNotificationEvent struct {
	EventSubBroadcaster
	ChatterUserID              string                            `json:"chatter_user_id"`
	ChatterUserLogin           string                            `json:"chatter_user_login"`
	ChatterUserName            string                            `json:"chatter_user_name"`
	ChatterIsAnonymous         bool                              `json:"chatter_is_anonymous"`
	Color                      string                            `json:"color"`
	Badges                     []ChatEventBadge                  `json:"badges"`
	SystemMessage              string                            `json:"system_message"`
	MessageID                  string                            `json:"message_id"`
	Message                    ChatEventMessage                  `json:"message"`
	NoticeType                 string                            `json:"notice_type"`
	Sub                        *ChatNotificationSub              `json:"sub,omitempty"`
	Resub                      *ChatNotificationResub            `json:"resub,omitempty"`
	SubGift                    *ChatNotificationSubGift          `json:"sub_gift,omitempty"`
	CommunitySubGift           *ChatNotificationCommunitySubGift `json:"community_sub_gift,omitempty"`
	GiftPaidUpgrade            *ChatNotificationGiftPaidUpgrade  `json:"gift_paid_upgrade,omitempty"`
	PrimePaidUpgrade           *ChatNotificationPrimePaidUpgrade `json:"prime_paid_upgrade,omitempty"`
	Raid                       *ChatNotificationRaid             `json:"raid,omitempty"`
	Unraid                     *ChatNotificationUnraid           `json:"unraid,omitempty"`
	PayItForward               *ChatNotificationPayItForward     `json:"pay_it_forward,omitempty"`
	Announcement               *ChatNotificationAnnouncement     `json:"announcement,omitempty"`
	BitsBadgeTier              *ChatNotificationBitsBadgeTier    `json:"bits_badge_tier,omitempty"`
	CharityDonation            *ChatNotificationCharityDonation  `json:"charity_donation,omitempty"`
	SharedChatSub              *ChatNotificationSub              `json:"shared_chat_sub,omitempty"`
	SharedChatResub            *ChatNotificationResub            `json:"shared_chat_resub,omitempty"`
	SharedChatSubGift          *ChatNotificationSubGift          `json:"shared_chat_sub_gift,omitempty"`
	SharedChatCommunitySubGift *ChatNotificationCommunitySubGift `json:"shared_chat_community_sub_gift,omitempty"`
	SharedChatGiftPaidUpgrade  *ChatNotificationGiftPaidUpgrade  `json:"shared_chat_gift_paid_upgrade,omitempty"`
	SharedChatPrimePaidUpgrade *ChatNotificationPrimePaidUpgrade `json:"shared_chat_prime_paid_upgrade,omitempty"`
	SharedChatRaid             *ChatNotificationRaid             `json:"shared_chat_raid,omitempty"`
	SharedChatPayItForward     *ChatNotificationPayItForward     `json:"shared_chat_pay_it_forward,omitempty"`
	SharedChatAnnouncement     *ChatNotificationAnnouncement     `json:"shared_chat_announcement,omitempty"`
	SourceBroadcasterUserID    *string                           `json:"source_broadcaster_user_id,omitempty"`
	SourceBroadcasterUserLogin *string                           `json:"source_broadcaster_user_login,omitempty"`
	SourceBroadcasterUserName  *string                           `json:"source_broadcaster_user_name,omitempty"`
	SourceMessageID            *string                           `json:"source_message_id,omitempty"`
	SourceBadges               []ChatEventBadge                  `json:"source_badges,omitempty"`
}

// ChatNotificationSub represents a sub notification.
type ChatNotificationSub struct {
	SubTier        string `json:"sub_tier"`
	IsPrime        bool   `json:"is_prime"`
	DurationMonths int    `json:"duration_months"`
}

// ChatNotificationResub represents a resub notification.
type ChatNotificationResub struct {
	CumulativeMonths  int     `json:"cumulative_months"`
	DurationMonths    int     `json:"duration_months"`
	StreakMonths      int     `json:"streak_months"`
	SubTier           string  `json:"sub_tier"`
	IsPrime           bool    `json:"is_prime"`
	IsGift            bool    `json:"is_gift"`
	GifterIsAnonymous bool    `json:"gifter_is_anonymous"`
	GifterUserID      *string `json:"gifter_user_id,omitempty"`
	GifterUserLogin   *string `json:"gifter_user_login,omitempty"`
	GifterUserName    *string `json:"gifter_user_name,omitempty"`
}

// ChatNotificationSubGift represents a sub gift notification.
type ChatNotificationSubGift struct {
	DurationMonths     int     `json:"duration_months"`
	CumulativeTotal    *int    `json:"cumulative_total,omitempty"`
	RecipientUserID    string  `json:"recipient_user_id"`
	RecipientUserLogin string  `json:"recipient_user_login"`
	RecipientUserName  string  `json:"recipient_user_name"`
	SubTier            string  `json:"sub_tier"`
	CommunityGiftID    *string `json:"community_gift_id,omitempty"`
}

// ChatNotificationCommunitySubGift represents a community sub gift notification.
type ChatNotificationCommunitySubGift struct {
	ID              string `json:"id"`
	Total           int    `json:"total"`
	SubTier         string `json:"sub_tier"`
	CumulativeTotal *int   `json:"cumulative_total,omitempty"`
}

// ChatNotificationGiftPaidUpgrade represents a gift paid upgrade notification.
type ChatNotificationGiftPaidUpgrade struct {
	GifterIsAnonymous bool    `json:"gifter_is_anonymous"`
	GifterUserID      *string `json:"gifter_user_id,omitempty"`
	GifterUserLogin   *string `json:"gifter_user_login,omitempty"`
	GifterUserName    *string `json:"gifter_user_name,omitempty"`
}

// ChatNotificationPrimePaidUpgrade represents a prime paid upgrade notification.
type ChatNotificationPrimePaidUpgrade struct {
	SubTier string `json:"sub_tier"`
}

// ChatNotificationRaid represents a raid notification.
type ChatNotificationRaid struct {
	UserID          string `json:"user_id"`
	UserLogin       string `json:"user_login"`
	UserName        string `json:"user_name"`
	ViewerCount     int    `json:"viewer_count"`
	ProfileImageURL string `json:"profile_image_url"`
}

// ChatNotificationUnraid represents an unraid notification.
type ChatNotificationUnraid struct{}

// ChatNotificationPayItForward represents a pay it forward notification.
type ChatNotificationPayItForward struct {
	GifterIsAnonymous bool    `json:"gifter_is_anonymous"`
	GifterUserID      *string `json:"gifter_user_id,omitempty"`
	GifterUserLogin   *string `json:"gifter_user_login,omitempty"`
	GifterUserName    *string `json:"gifter_user_name,omitempty"`
}

// ChatNotificationAnnouncement represents an announcement notification.
type ChatNotificationAnnouncement struct {
	Color string `json:"color"`
}

// ChatNotificationBitsBadgeTier represents a bits badge tier notification.
type ChatNotificationBitsBadgeTier struct {
	Tier int `json:"tier"`
}

// ChatNotificationCharityDonation represents a charity donation notification.
type ChatNotificationCharityDonation struct {
	CharityName string        `json:"charity_name"`
	Amount      CharityAmount `json:"amount"`
}

// Chat Clear Events

// ChannelChatClearEvent is sent when chat is cleared.
type ChannelChatClearEvent struct {
	EventSubBroadcaster
}

// ChannelChatClearUserMessagesEvent is sent when a user's messages are cleared.
type ChannelChatClearUserMessagesEvent struct {
	EventSubBroadcaster
	TargetUserID    string `json:"target_user_id"`
	TargetUserLogin string `json:"target_user_login"`
	TargetUserName  string `json:"target_user_name"`
	// Shared chat fields
	SourceBroadcasterUserID    *string `json:"source_broadcaster_user_id,omitempty"`
	SourceBroadcasterUserLogin *string `json:"source_broadcaster_user_login,omitempty"`
	SourceBroadcasterUserName  *string `json:"source_broadcaster_user_name,omitempty"`
}

// ChannelChatMessageDeleteEvent is sent when a message is deleted.
type ChannelChatMessageDeleteEvent struct {
	EventSubBroadcaster
	TargetUserID    string `json:"target_user_id"`
	TargetUserLogin string `json:"target_user_login"`
	TargetUserName  string `json:"target_user_name"`
	MessageID       string `json:"message_id"`
	// Shared chat fields
	SourceBroadcasterUserID    *string `json:"source_broadcaster_user_id,omitempty"`
	SourceBroadcasterUserLogin *string `json:"source_broadcaster_user_login,omitempty"`
	SourceBroadcasterUserName  *string `json:"source_broadcaster_user_name,omitempty"`
}

// Chat Settings Update Event

// ChannelChatSettingsUpdateEvent is sent when chat settings are updated.
type ChannelChatSettingsUpdateEvent struct {
	EventSubBroadcaster
	EmoteMode                   bool `json:"emote_mode"`
	FollowerMode                bool `json:"follower_mode"`
	FollowerModeDurationMinutes *int `json:"follower_mode_duration_minutes,omitempty"`
	SlowMode                    bool `json:"slow_mode"`
	SlowModeWaitTimeSeconds     int  `json:"slow_mode_wait_time_seconds"`
	SubscriberMode              bool `json:"subscriber_mode"`
	UniqueChatMode              bool `json:"unique_chat_mode"`
}

// Chat User Message Hold Events

// ChannelChatUserMessageHoldEvent is sent when a user's message is held.
type ChannelChatUserMessageHoldEvent struct {
	EventSubBroadcaster
	EventSubUser
	MessageID string              `json:"message_id"`
	Message   ChatUserHoldMessage `json:"message"`
	// Shared chat fields
	SourceBroadcasterUserID    *string `json:"source_broadcaster_user_id,omitempty"`
	SourceBroadcasterUserLogin *string `json:"source_broadcaster_user_login,omitempty"`
	SourceBroadcasterUserName  *string `json:"source_broadcaster_user_name,omitempty"`
}

// ChatUserHoldMessage represents a held message.
type ChatUserHoldMessage struct {
	Text      string                 `json:"text"`
	Fragments []ChatUserHoldFragment `json:"fragments"`
}

// ChatUserHoldFragment represents a fragment in a held message.
type ChatUserHoldFragment struct {
	Type      string              `json:"type"`
	Text      string              `json:"text"`
	Cheermote *ChatEventCheermote `json:"cheermote,omitempty"`
	Emote     *ChatEventEmote     `json:"emote,omitempty"`
}

// ChannelChatUserMessageUpdateEvent is sent when a held message status is updated.
type ChannelChatUserMessageUpdateEvent struct {
	EventSubBroadcaster
	EventSubUser
	Status    string              `json:"status"` // approved, denied, invalid
	MessageID string              `json:"message_id"`
	Message   ChatUserHoldMessage `json:"message"`
	// Shared chat fields
	SourceBroadcasterUserID    *string `json:"source_broadcaster_user_id,omitempty"`
	SourceBroadcasterUserLogin *string `json:"source_broadcaster_user_login,omitempty"`
	SourceBroadcasterUserName  *string `json:"source_broadcaster_user_name,omitempty"`
}

// Automatic Reward Redemption Events

// ChannelPointsAutomaticRewardRedemptionAddEvent is sent when an automatic reward is redeemed.
type ChannelPointsAutomaticRewardRedemptionAddEvent struct {
	EventSubBroadcaster
	EventSubUser
	ID         string                    `json:"id"`
	Reward     AutomaticRewardRedemption `json:"reward"`
	Message    AutomaticRewardMessage    `json:"message"`
	UserInput  string                    `json:"user_input"`
	RedeemedAt time.Time                 `json:"redeemed_at"`
}

// AutomaticRewardRedemption represents an automatic reward.
type AutomaticRewardRedemption struct {
	Type          string         `json:"type"` // single_message_bypass_sub_mode, send_highlighted_message, random_sub_emote_unlock, chosen_sub_emote_unlock, chosen_modified_sub_emote_unlock, message_effect, gigantify_an_emote, celebration
	Cost          int            `json:"cost"`
	UnlockedEmote *UnlockedEmote `json:"unlocked_emote,omitempty"`
}

// UnlockedEmote represents an unlocked emote reward.
type UnlockedEmote struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AutomaticRewardMessage represents an automatic reward message.
type AutomaticRewardMessage struct {
	Text   string                 `json:"text"`
	Emotes []AutomaticRewardEmote `json:"emotes"`
}

// AutomaticRewardEmote represents an emote in an automatic reward message.
type AutomaticRewardEmote struct {
	ID    string `json:"id"`
	Begin int    `json:"begin"`
	End   int    `json:"end"`
}

// Guest Star Events (Beta)

// ChannelGuestStarSessionBeginEvent is sent when a guest star session begins.
type ChannelGuestStarSessionBeginEvent struct {
	EventSubBroadcaster
	SessionID string    `json:"session_id"`
	StartedAt time.Time `json:"started_at"`
}

// ChannelGuestStarSessionEndEvent is sent when a guest star session ends.
type ChannelGuestStarSessionEndEvent struct {
	EventSubBroadcaster
	SessionID string    `json:"session_id"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// ChannelGuestStarGuestUpdateEvent is sent when a guest star guest is updated.
type ChannelGuestStarGuestUpdateEvent struct {
	EventSubBroadcaster
	EventSubModerator
	SessionID        string `json:"session_id"`
	GuestUserID      string `json:"guest_user_id"`
	GuestUserLogin   string `json:"guest_user_login"`
	GuestUserName    string `json:"guest_user_name"`
	SlotID           string `json:"slot_id"`
	State            string `json:"state"`
	HostVideoEnabled *bool  `json:"host_video_enabled,omitempty"`
	HostAudioEnabled *bool  `json:"host_audio_enabled,omitempty"`
	HostVolume       *int   `json:"host_volume,omitempty"`
}

// ChannelGuestStarSettingsUpdateEvent is sent when guest star settings are updated.
type ChannelGuestStarSettingsUpdateEvent struct {
	EventSubBroadcaster
	IsModeratorSendLiveEnabled  bool   `json:"is_moderator_send_live_enabled"`
	SlotCount                   int    `json:"slot_count"`
	IsBrowserSourceAudioEnabled bool   `json:"is_browser_source_audio_enabled"`
	GroupLayout                 string `json:"group_layout"`
}
