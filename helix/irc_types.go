package helix

import "time"

// ChatMessage represents a PRIVMSG from Twitch IRC.
type ChatMessage struct {
	ID            string            // Unique message ID
	Channel       string            // Channel name (without #)
	User          string            // Username (login)
	UserID        string            // User's Twitch ID
	Message       string            // Message content
	Emotes        []IRCEmote        // Emotes used in the message
	Badges        map[string]string // User badges (badge-name -> version)
	BadgeInfo     map[string]string // Additional badge info (e.g., subscriber months)
	Color         string            // User's chat color (#RRGGBB)
	DisplayName   string            // User's display name
	IsMod         bool              // Is channel moderator
	IsVIP         bool              // Is VIP
	IsSubscriber  bool              // Is subscriber
	IsBroadcaster bool              // Is the channel broadcaster
	Bits          int               // Bits cheered (0 if none)
	FirstMessage  bool              // Is this user's first message in channel
	ReturningChatter bool           // Is a returning chatter
	ReplyParentMsgID   string       // Parent message ID if this is a reply
	ReplyParentUserID  string       // Parent message user ID
	ReplyParentUserLogin string     // Parent message username
	ReplyParentDisplayName string   // Parent message display name
	ReplyParentMsgBody string       // Parent message content
	Timestamp     time.Time         // Server timestamp
	Raw           string            // Raw IRC message
}

// IRCEmote represents an emote used in an IRC message.
type IRCEmote struct {
	ID     string // Emote ID
	Name   string // Emote name/code
	Start  int    // Start position in message
	End    int    // End position in message
	Count  int    // Number of times used
}

// UserNotice represents a USERNOTICE message (subs, raids, etc.).
type UserNotice struct {
	Type          string            // sub, resub, subgift, raid, ritual, etc.
	Channel       string            // Channel name
	User          string            // Username
	UserID        string            // User's Twitch ID
	DisplayName   string            // Display name
	Message       string            // Optional user message
	SystemMessage string            // System-generated message
	MsgParams     map[string]string // Type-specific parameters
	Badges        map[string]string // User badges
	BadgeInfo     map[string]string // Badge info
	Color         string            // User color
	Emotes        []IRCEmote        // Emotes in message
	Timestamp     time.Time         // Server timestamp
	Raw           string            // Raw IRC message
}

// Common MsgParams keys for UserNotice:
// - msg-param-cumulative-months: Total months subscribed
// - msg-param-months: Months in current streak
// - msg-param-multimonth-duration: Multi-month gift duration
// - msg-param-multimonth-tenure: Tenure of multi-month
// - msg-param-should-share-streak: Whether to share streak
// - msg-param-streak-months: Streak months
// - msg-param-sub-plan: Subscription plan (1000, 2000, 3000, Prime)
// - msg-param-sub-plan-name: Plan display name
// - msg-param-gift-months: Gifted months
// - msg-param-recipient-id: Gift recipient ID
// - msg-param-recipient-user-name: Gift recipient username
// - msg-param-recipient-display-name: Gift recipient display name
// - msg-param-sender-count: Sender's total gift count
// - msg-param-viewerCount: Raid viewer count
// - msg-param-displayName: Raid display name
// - msg-param-login: Raid login

// UserNotice types
const (
	UserNoticeTypeSub               = "sub"
	UserNoticeTypeResub             = "resub"
	UserNoticeTypeSubGift           = "subgift"
	UserNoticeTypeAnonSubGift       = "anonsubgift"
	UserNoticeTypeSubMysteryGift    = "submysterygift"
	UserNoticeTypeGiftPaidUpgrade   = "giftpaidupgrade"
	UserNoticeTypePrimePaidUpgrade  = "primepaidupgrade"
	UserNoticeTypeRaid              = "raid"
	UserNoticeTypeUnraid            = "unraid"
	UserNoticeTypeRitual            = "ritual"
	UserNoticeTypeBitsBadgeTier     = "bitsbadgetier"
	UserNoticeTypeCommunityPayForward = "communitypayforward"
	UserNoticeTypeStandardPayForward  = "standardpayforward"
	UserNoticeTypeAnnouncement      = "announcement"
)

// RoomState represents a ROOMSTATE message.
type RoomState struct {
	Channel       string // Channel name
	EmoteOnly     bool   // Emote-only mode
	FollowersOnly int    // Followers-only mode (-1 = off, 0+ = minutes required)
	R9K           bool   // R9K/unique mode
	Slow          int    // Slow mode (seconds between messages, 0 = off)
	SubsOnly      bool   // Subscribers-only mode
	RoomID        string // Channel/room ID
	Raw           string // Raw IRC message
}

// Notice represents a NOTICE message from the server.
type Notice struct {
	Channel   string // Channel name (may be empty for global notices)
	Message   string // Notice message
	MsgID     string // Notice type identifier
	Raw       string // Raw IRC message
}

// Common Notice MsgIDs
const (
	NoticeMsgIDSubsOn           = "subs_on"
	NoticeMsgIDSubsOff          = "subs_off"
	NoticeMsgIDEmoteOnlyOn      = "emote_only_on"
	NoticeMsgIDEmoteOnlyOff     = "emote_only_off"
	NoticeMsgIDSlowOn           = "slow_on"
	NoticeMsgIDSlowOff          = "slow_off"
	NoticeMsgIDFollowersOn      = "followers_on"
	NoticeMsgIDFollowersOff     = "followers_off"
	NoticeMsgIDR9KOn            = "r9k_on"
	NoticeMsgIDR9KOff           = "r9k_off"
	NoticeMsgIDHostOn           = "host_on"
	NoticeMsgIDHostOff          = "host_off"
	NoticeMsgIDMsgChannelSuspended = "msg_channel_suspended"
	NoticeMsgIDMsgBanned        = "msg_banned"
	NoticeMsgIDMsgRateLimit     = "msg_ratelimit"
	NoticeMsgIDMsgDuplicate     = "msg_duplicate"
	NoticeMsgIDMsgFollowersOnly = "msg_followersonly"
	NoticeMsgIDMsgSubsOnly      = "msg_subsonly"
	NoticeMsgIDMsgEmoteOnly     = "msg_emoteonly"
	NoticeMsgIDMsgSlowMode      = "msg_slowmode"
	NoticeMsgIDMsgR9K           = "msg_r9k"
	NoticeMsgIDNoPermission     = "no_permission"
	NoticeMsgIDUnrecognizedCmd  = "unrecognized_cmd"
)

// ClearChat represents a CLEARCHAT message (timeout/ban).
type ClearChat struct {
	Channel       string        // Channel name
	User          string        // User being cleared (empty = chat cleared)
	BanDuration   int           // Duration in seconds (0 = permanent ban)
	RoomID        string        // Channel ID
	TargetUserID  string        // Banned user's ID
	Timestamp     time.Time     // Server timestamp
	Raw           string        // Raw IRC message
}

// ClearMessage represents a CLEARMSG message (single message deletion).
type ClearMessage struct {
	Channel       string    // Channel name
	User          string    // Message author
	Message       string    // Deleted message content
	TargetMsgID   string    // ID of deleted message
	Timestamp     time.Time // Server timestamp
	Raw           string    // Raw IRC message
}

// Whisper represents a WHISPER message (direct message).
type Whisper struct {
	From        string            // Sender username
	FromID      string            // Sender user ID
	To          string            // Recipient username
	Message     string            // Message content
	DisplayName string            // Sender display name
	Color       string            // Sender color
	Badges      map[string]string // Sender badges
	Emotes      []IRCEmote        // Emotes in message
	MessageID   string            // Unique message ID
	ThreadID    string            // Thread ID
	Raw         string            // Raw IRC message
}

// GlobalUserState represents a GLOBALUSERSTATE message.
type GlobalUserState struct {
	UserID      string            // User's Twitch ID
	DisplayName string            // Display name
	Color       string            // Chat color
	Badges      map[string]string // Global badges
	BadgeInfo   map[string]string // Badge info
	EmoteSets   []string          // Available emote set IDs
	Raw         string            // Raw IRC message
}

// UserState represents a USERSTATE message.
type UserState struct {
	Channel     string            // Channel name
	DisplayName string            // Display name
	Color       string            // Chat color
	Badges      map[string]string // Channel-specific badges
	BadgeInfo   map[string]string // Badge info
	EmoteSets   []string          // Available emote set IDs
	IsMod       bool              // Is moderator
	IsSubscriber bool             // Is subscriber
	Raw         string            // Raw IRC message
}

// IRCMessage represents a parsed IRC message.
type IRCMessage struct {
	Raw     string            // Raw message
	Tags    map[string]string // IRCv3 tags
	Prefix  string            // Source prefix (nick!user@host)
	Command string            // IRC command
	Params  []string          // Command parameters
	Trailing string           // Trailing parameter (after :)
}
