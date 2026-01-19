---
layout: default
title: Chat API
description: Manage chat rooms, emotes, badges, settings, and send messages to Twitch channels.
---

## GetChatters

Get the list of users that are connected to a broadcaster's chat session.

**Requires:** `moderator:read:chatters`

```go
resp, err := client.GetChatters(ctx, &helix.GetChattersParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

for _, chatter := range resp.Data {
    fmt.Printf("Chatter: %s (%s)\n", chatter.UserName, chatter.UserLogin)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "128393656",
      "user_login": "smittysmithers",
      "user_name": "SmittySmithers"
    },
    {
      "user_id": "129546453",
      "user_login": "twitchdev",
      "user_name": "TwitchDev"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  },
  "total": 8
}
```

## GetChannelEmotes

Get all custom emotes for a specific channel.

**No authentication required**

```go
resp, err := client.GetChannelEmotes(ctx, "12345")

for _, emote := range resp.Data {
    fmt.Printf("Emote: %s - %s\n", emote.Name, emote.ID)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "304456832",
      "name": "twitchdevPitchfork",
      "images": {
        "url_1x": "https://static-cdn.jtvnw.net/emoticons/v2/304456832/static/light/1.0",
        "url_2x": "https://static-cdn.jtvnw.net/emoticons/v2/304456832/static/light/2.0",
        "url_4x": "https://static-cdn.jtvnw.net/emoticons/v2/304456832/static/light/3.0"
      },
      "tier": "1000",
      "emote_type": "subscriptions",
      "emote_set_id": "301590448",
      "format": [
        "static"
      ],
      "scale": [
        "1.0",
        "2.0",
        "3.0"
      ],
      "theme_mode": [
        "light",
        "dark"
      ]
    }
  ],
  "template": "https://static-cdn.jtvnw.net/emoticons/v2/{{id}}/{{format}}/{{theme_mode}}/{{scale}}"
}
```

## GetGlobalEmotes

Get all global emotes available on Twitch.

**No authentication required**

```go
resp, err := client.GetGlobalEmotes(ctx)

for _, emote := range resp.Data {
    fmt.Printf("Global Emote: %s - %s\n", emote.Name, emote.ID)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "196892",
      "name": "TwitchUnity",
      "images": {
        "url_1x": "https://static-cdn.jtvnw.net/emoticons/v2/196892/static/light/1.0",
        "url_2x": "https://static-cdn.jtvnw.net/emoticons/v2/196892/static/light/2.0",
        "url_4x": "https://static-cdn.jtvnw.net/emoticons/v2/196892/static/light/3.0"
      },
      "format": [
        "static"
      ],
      "scale": [
        "1.0",
        "2.0",
        "3.0"
      ],
      "theme_mode": [
        "light",
        "dark"
      ]
    }
  ],
  "template": "https://static-cdn.jtvnw.net/emoticons/v2/{{id}}/{{format}}/{{theme_mode}}/{{scale}}"
}
```

## GetEmoteSets

Get emotes for one or more emote sets.

**No authentication required**

```go
resp, err := client.GetEmoteSets(ctx, []string{"12345", "67890"})

for _, emote := range resp.Data {
    fmt.Printf("Emote: %s (Set: %s)\n", emote.Name, emote.EmoteSetID)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "304456832",
      "name": "twitchdevPitchfork",
      "images": {
        "url_1x": "https://static-cdn.jtvnw.net/emoticons/v2/304456832/static/light/1.0",
        "url_2x": "https://static-cdn.jtvnw.net/emoticons/v2/304456832/static/light/2.0",
        "url_4x": "https://static-cdn.jtvnw.net/emoticons/v2/304456832/static/light/3.0"
      },
      "emote_type": "subscriptions",
      "emote_set_id": "301590448",
      "format": [
        "static"
      ],
      "scale": [
        "1.0",
        "2.0",
        "3.0"
      ],
      "theme_mode": [
        "light",
        "dark"
      ]
    }
  ],
  "template": "https://static-cdn.jtvnw.net/emoticons/v2/{{id}}/{{format}}/{{theme_mode}}/{{scale}}"
}
```

## GetChannelChatBadges

Get chat badges for a specific channel.

**No authentication required**

```go
resp, err := client.GetChannelChatBadges(ctx, "12345")

for _, badgeSet := range resp.Data {
    fmt.Printf("Badge Set: %s\n", badgeSet.SetID)
    for _, version := range badgeSet.Versions {
        fmt.Printf("  Version %s: %s\n", version.ID, version.Title)
    }
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "set_id": "subscriber",
      "versions": [
        {
          "id": "0",
          "image_url_1x": "https://static-cdn.jtvnw.net/badges/v1/5d9f2208-5dd8-11e7-8513-2ff4adfae661/1",
          "image_url_2x": "https://static-cdn.jtvnw.net/badges/v1/5d9f2208-5dd8-11e7-8513-2ff4adfae661/2",
          "image_url_4x": "https://static-cdn.jtvnw.net/badges/v1/5d9f2208-5dd8-11e7-8513-2ff4adfae661/3",
          "title": "Subscriber",
          "description": "Subscriber",
          "click_action": "subscribe_to_channel",
          "click_url": ""
        }
      ]
    }
  ]
}
```

## GetGlobalChatBadges

Get global chat badges available on Twitch.

**No authentication required**

```go
resp, err := client.GetGlobalChatBadges(ctx)

for _, badgeSet := range resp.Data {
    fmt.Printf("Badge Set: %s\n", badgeSet.SetID)
    for _, version := range badgeSet.Versions {
        fmt.Printf("  Version %s: %s\n", version.ID, version.Title)
    }
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "set_id": "moderator",
      "versions": [
        {
          "id": "1",
          "image_url_1x": "https://static-cdn.jtvnw.net/badges/v1/3267646d-33f0-4b17-b3df-f923a41db1d0/1",
          "image_url_2x": "https://static-cdn.jtvnw.net/badges/v1/3267646d-33f0-4b17-b3df-f923a41db1d0/2",
          "image_url_4x": "https://static-cdn.jtvnw.net/badges/v1/3267646d-33f0-4b17-b3df-f923a41db1d0/3",
          "title": "Moderator",
          "description": "Moderator",
          "click_action": "none",
          "click_url": ""
        }
      ]
    }
  ]
}
```

## Badge Constants and Helpers

The library provides constants for common badge SetIDs and helper functions for working with badges in EventSub chat events.

### Badge Constants

```go
// Channel role badges
helix.BadgeBroadcaster   // "broadcaster"
helix.BadgeModerator     // "moderator"
helix.BadgeLeadModerator // "lead_moderator"
helix.BadgeVIP           // "vip"

// Subscription badges
helix.BadgeSubscriber    // "subscriber"
helix.BadgeFounder       // "founder"
helix.BadgeSubGifter     // "sub-gifter"

// Bits badges
helix.BadgeBits          // "bits"
helix.BadgeBitsLeader    // "bits-leader"

// Twitch staff badges
helix.BadgeStaff         // "staff"
helix.BadgeAdmin         // "admin"
helix.BadgeGlobalMod     // "global_mod"

// Other badges
helix.BadgePartner       // "partner"
helix.BadgePremium       // "premium"
helix.BadgeArtist        // "artist"
helix.BadgePredictions   // "predictions"
helix.BadgeHypeTrain     // "hype-train"
```

### Lead Moderator Badge

The `lead_moderator` badge was introduced for the Lead Moderator role. Lead Moderators have additional privileges to help streamers manage their mod teams. They can choose to display either the Lead Moderator badge or the regular Moderator badge.

**Important:** If your application checks for the `moderator` badge to confirm moderator privileges, you should update your logic to check for either `moderator` OR `lead_moderator`. Use the `HasModeratorPrivileges()` helper method for this purpose.

### Badge Helper Methods

Convert EventSub badges to use helper methods:

```go
// In an EventSub chat message handler
func handleChatMessage(event *helix.ChannelChatMessageEvent) {
    // Convert badges to use helper methods
    badges := helix.ToChatEventBadges(event.Badges)

    // Check for moderator privileges (handles both moderator and lead_moderator)
    if badges.HasModeratorPrivileges() {
        fmt.Printf("%s is a moderator\n", event.ChatterUserName)
    }

    // Check for broadcaster
    if badges.HasBroadcasterPrivileges() {
        fmt.Printf("%s is the broadcaster\n", event.ChatterUserName)
    }

    // Check for VIP
    if badges.HasVIPStatus() {
        fmt.Printf("%s is a VIP\n", event.ChatterUserName)
    }

    // Check for subscriber (includes founders)
    if badges.IsSubscriber() {
        fmt.Printf("%s is a subscriber\n", event.ChatterUserName)
    }

    // Check for Twitch staff (staff, admin, global_mod)
    if badges.IsStaff() {
        fmt.Printf("%s is Twitch staff\n", event.ChatterUserName)
    }
}
```

### Available Helper Methods

| Method | Description |
|--------|-------------|
| `HasBadge(setID)` | Check if user has a specific badge |
| `HasAnyBadge(setIDs...)` | Check if user has any of the specified badges |
| `HasModeratorPrivileges()` | Check for `moderator` OR `lead_moderator` badge |
| `HasBroadcasterPrivileges()` | Check for `broadcaster` badge |
| `HasVIPStatus()` | Check for `vip` badge |
| `IsSubscriber()` | Check for `subscriber` OR `founder` badge |
| `IsStaff()` | Check for `staff`, `admin`, OR `global_mod` badge |
| `GetBadge(setID)` | Get a specific badge or nil if not found |

### Example: Moderator Check with Lead Moderator Support

```go
// OLD: Only checks for moderator badge (misses Lead Moderators)
func isModeratorOld(badges []helix.ChatEventBadge) bool {
    for _, badge := range badges {
        if badge.SetID == "moderator" {
            return true
        }
    }
    return false
}

// NEW: Properly checks for both moderator and lead_moderator
func isModeratorNew(badges []helix.ChatEventBadge) bool {
    return helix.ToChatEventBadges(badges).HasModeratorPrivileges()
}
```

## GetChatSettings

Get chat settings for a broadcaster's chat room.

**No authentication required**

```go
// Get chat settings
resp, err := client.GetChatSettings(ctx, "12345", "")

// Get chat settings with moderator context
resp, err = client.GetChatSettings(ctx, "12345", "67890")

settings := resp.Data[0]
fmt.Printf("Slow Mode: %v (%d seconds)\n", settings.SlowMode, settings.SlowModeWaitTime)
fmt.Printf("Follower Mode: %v (%d minutes)\n", settings.FollowerMode, settings.FollowerModeDuration)
fmt.Printf("Subscriber Mode: %v\n", settings.SubscriberMode)
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "slow_mode": true,
      "slow_mode_wait_time": 30,
      "follower_mode": true,
      "follower_mode_duration": 10,
      "subscriber_mode": false,
      "emote_mode": false,
      "unique_chat_mode": true,
      "non_moderator_chat_delay": true,
      "non_moderator_chat_delay_duration": 4
    }
  ]
}
```

## UpdateChatSettings

Update chat settings for a broadcaster's chat room.

**Requires:** `moderator:manage:chat_settings`

```go
emoteMode := true
followerMode := true
followerModeDuration := 10
nonModDelay := true
nonModDelayDuration := 2
slowMode := true
slowModeWaitTime := 30
subMode := false
uniqueChat := true

resp, err := client.UpdateChatSettings(ctx, &helix.UpdateChatSettingsParams{
    BroadcasterID:                 "12345",
    ModeratorID:                   "67890",
    EmoteMode:                     &emoteMode,
    FollowerMode:                  &followerMode,
    FollowerModeDuration:          &followerModeDuration, // 10 minutes
    NonModeratorChatDelay:         &nonModDelay,
    NonModeratorChatDelayDuration: &nonModDelayDuration, // 2 seconds
    SlowMode:                      &slowMode,
    SlowModeWaitTime:              &slowModeWaitTime, // 30 seconds
    SubscriberMode:                &subMode,
    UniqueChatMode:                &uniqueChat,
})

settings := resp.Data[0]
fmt.Printf("Chat settings updated: Slow Mode %v, Follower Mode %v\n",
    settings.SlowMode, settings.FollowerMode)
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "slow_mode": true,
      "slow_mode_wait_time": 30,
      "follower_mode": true,
      "follower_mode_duration": 10,
      "subscriber_mode": false,
      "emote_mode": true,
      "unique_chat_mode": true,
      "non_moderator_chat_delay": true,
      "non_moderator_chat_delay_duration": 2
    }
  ]
}
```

## SendChatAnnouncement

Send an announcement to a broadcaster's chat room.

**Requires:** `moderator:manage:announcements`

```go
// Send an announcement with default color
err := client.SendChatAnnouncement(ctx, &helix.SendChatAnnouncementParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Message:       "Welcome to the stream! Don't forget to follow!",
})

// Send a colored announcement
err = client.SendChatAnnouncement(ctx, &helix.SendChatAnnouncementParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Message:       "Important update coming soon!",
    Color:         "purple", // blue, green, orange, purple, or primary
})
```

**Sample Response:**
```json
204 No Content
```

## SendShoutout

Send a shoutout to another broadcaster.

**Requires:** `moderator:manage:shoutouts`

```go
err := client.SendShoutout(ctx, &helix.SendShoutoutParams{
    FromBroadcasterID: "12345",
    ToBroadcasterID:   "67890",
    ModeratorID:       "12345",
})

if err != nil {
    fmt.Printf("Failed to send shoutout: %v\n", err)
}
```

**Sample Response:**
```json
204 No Content
```

## GetUserChatColor

Get the chat color for one or more users.

**No authentication required**

```go
resp, err := client.GetUserChatColor(ctx, []string{"12345", "67890"})

for _, userColor := range resp.Data {
    fmt.Printf("User %s has color: %s\n", userColor.UserID, userColor.Color)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "12345",
      "user_login": "twitchdev",
      "user_name": "TwitchDev",
      "color": "#9146FF"
    },
    {
      "user_id": "67890",
      "user_login": "smittysmithers",
      "user_name": "SmittySmithers",
      "color": "#FF0000"
    }
  ]
}
```

## UpdateUserChatColor

Update the chat color for a user.

**Requires:** `user:manage:chat_color`

```go
// Update to a named color
err := client.UpdateUserChatColor(ctx, "12345", "blue")

// Update to a custom hex color (Turbo/Prime users only)
err = client.UpdateUserChatColor(ctx, "12345", "#9146FF")
```

**Sample Response:**
```json
204 No Content
```

## SendChatMessage

Send a message to a broadcaster's chat room.

**Requires:** `user:write:chat`

```go
// Send a regular message
resp, err := client.SendChatMessage(ctx, &helix.SendChatMessageParams{
    BroadcasterID: "12345",
    SenderID:      "67890",
    Message:       "Hello, chat!",
})

// Send a reply to another message
resp, err = client.SendChatMessage(ctx, &helix.SendChatMessageParams{
    BroadcasterID:        "12345",
    SenderID:             "67890",
    Message:              "That's a great point!",
    ReplyParentMessageID: "abc-123-def",
})

if resp.IsSent {
    fmt.Printf("Message sent: %s\n", resp.MessageID)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "message_id": "abc-123-def",
      "is_sent": true
    }
  ]
}
```

Or if the message was not sent due to chat rules:

```json
{
  "data": [
    {
      "message_id": "",
      "is_sent": false,
      "drop_reason": {
        "code": "msg_rejected",
        "message": "Your message was rejected by AutoMod."
      }
    }
  ]
}
```

## GetSharedChatSession

Get information about a shared chat session.

**No authentication required**

```go
resp, err := client.GetSharedChatSession(ctx, "12345")

if resp != nil {
    fmt.Printf("Session ID: %s\n", resp.SessionID)
    fmt.Printf("Host Broadcaster: %s\n", resp.HostBroadcasterID)
    for _, participant := range resp.Participants {
        fmt.Printf("Participant: %s\n", participant.BroadcasterID)
    }
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "session_id": "d7d07c15-be07-4b55-9e73-184e1eb1a3e3",
      "host_broadcaster_id": "12345",
      "participants": [
        {
          "broadcaster_id": "12345"
        },
        {
          "broadcaster_id": "67890"
        }
      ],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

## GetUserEmotes

Get emotes available to a user based on their subscriptions.

**Requires:** `user:read:emotes`

```go
// Get all emotes for a user
resp, err := client.GetUserEmotes(ctx, &helix.GetUserEmotesParams{
    UserID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

// Get emotes for a user in a specific broadcaster's context
resp, err = client.GetUserEmotes(ctx, &helix.GetUserEmotesParams{
    UserID:        "12345",
    BroadcasterID: "67890",
    PaginationParams: &helix.PaginationParams{
        First: 50,
    },
})

for _, emote := range resp.Data {
    fmt.Printf("Emote: %s - %s (Type: %s)\n", emote.Name, emote.ID, emote.EmoteType)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "304456832",
      "name": "twitchdevPitchfork",
      "emote_type": "subscriptions",
      "emote_set_id": "301590448",
      "owner_id": "141981764",
      "format": [
        "static"
      ],
      "scale": [
        "1.0",
        "2.0",
        "3.0"
      ],
      "theme_mode": [
        "light",
        "dark"
      ]
    },
    {
      "id": "emotesv2_dc24652ada1e4c84a5e3ceebae4de709",
      "name": "PogChamp",
      "emote_type": "globals",
      "emote_set_id": "0",
      "owner_id": "",
      "format": [
        "static"
      ],
      "scale": [
        "1.0",
        "2.0",
        "3.0"
      ],
      "theme_mode": [
        "light",
        "dark"
      ]
    }
  ],
  "template": "https://static-cdn.jtvnw.net/emoticons/v2/{{id}}/{{format}}/{{theme_mode}}/{{scale}}",
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjoiMTUwMzQ0MTc3NjQyNDQyMjAwMCJ9"
  }
}
```

