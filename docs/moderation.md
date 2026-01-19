---
layout: default
title: Moderation API
description: Manage channel moderation including bans, moderators, blocked terms, AutoMod, and shield mode.
---

## GetBannedUsers

Get all banned and timed-out users in a channel.

**Requires:** `moderation:read`

```go
resp, err := client.GetBannedUsers(ctx, &helix.GetBannedUsersParams{
    BroadcasterID: "12345",
    UserIDs:       []string{"67890", "11111"}, // Optional, max 100
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

for _, ban := range resp.Data {
    fmt.Printf("User %s banned by %s\n", ban.UserName, ban.ModeratorName)
    fmt.Printf("Reason: %s\n", ban.Reason)
    if ban.ExpiresAt != "" {
        fmt.Printf("Expires: %s\n", ban.ExpiresAt)
    } else {
        fmt.Printf("Permanent ban\n")
    }
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "67890",
      "user_login": "spammer123",
      "user_name": "Spammer123",
      "expires_at": "2025-12-11T15:30:00Z",
      "created_at": "2025-12-11T15:00:00Z",
      "reason": "Spamming in chat",
      "moderator_id": "11111",
      "moderator_login": "coolmod",
      "moderator_name": "CoolMod"
    },
    {
      "user_id": "22222",
      "user_login": "baduser",
      "user_name": "BadUser",
      "created_at": "2025-12-10T10:00:00Z",
      "reason": "Violated community guidelines",
      "moderator_id": "11111",
      "moderator_login": "coolmod",
      "moderator_name": "CoolMod"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## BanUser

Ban a user from participating in a broadcaster's chat room, or put them in a timeout.

**Requires:** `moderator:manage:banned_users`

```go
// Permanent ban
resp, err := client.BanUser(ctx, &helix.BanUserParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Data: &helix.BanUserData{
        UserID:   "11111",
        Duration: 0, // 0 = permanent
        Reason:   "Violated community guidelines",
    },
})

// Timeout for 10 minutes (600 seconds)
resp, err = client.BanUser(ctx, &helix.BanUserParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Data: &helix.BanUserData{
        UserID:   "22222",
        Duration: 600, // 1 to 1209600 seconds (14 days max)
        Reason:   "Spamming",
    },
})

if resp.Data != nil && len(resp.Data) > 0 {
    ban := resp.Data[0]
    fmt.Printf("User banned until: %s\n", ban.EndTime)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "moderator_id": "67890",
      "user_id": "11111",
      "created_at": "2025-12-11T15:00:00Z",
      "end_time": "2025-12-11T15:10:00Z"
    }
  ]
}
```

## UnbanUser

Unban a user from participating in a broadcaster's chat room or remove a timeout.

**Requires:** `moderator:manage:banned_users`

```go
err := client.UnbanUser(ctx, &helix.UnbanUserParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    UserID:        "11111",
})

if err != nil {
    fmt.Printf("Failed to unban user: %v\n", err)
}
```

**Response:** This endpoint returns 204 No Content on success.

## GetModerators

Get all users allowed to moderate a broadcaster's chat room.

**Note:** This endpoint returns both Moderators and Lead Moderators, as Lead Moderators are also Moderators with additional privileges.

**Requires:** `moderation:read`

```go
resp, err := client.GetModerators(ctx, &helix.GetModeratorsParams{
    BroadcasterID: "12345",
    UserIDs:       []string{"67890", "11111"}, // Optional, max 100
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

for _, mod := range resp.Data {
    fmt.Printf("Moderator: %s (%s)\n", mod.UserName, mod.UserLogin)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "67890",
      "user_login": "coolmod",
      "user_name": "CoolMod"
    },
    {
      "user_id": "11111",
      "user_login": "awesomemod",
      "user_name": "AwesomeMod"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6MTB9fQ"
  }
}
```

## AddChannelModerator

Add a moderator to a broadcaster's chat room.

**Requires:** `channel:manage:moderators`

```go
err := client.AddChannelModerator(ctx, &helix.AddChannelModeratorParams{
    BroadcasterID: "12345",
    UserID:        "67890",
})

if err != nil {
    fmt.Printf("Failed to add moderator: %v\n", err)
}
```

**Response:** This endpoint returns 204 No Content on success.

## RemoveChannelModerator

Remove a moderator from a broadcaster's chat room.

**Requires:** `channel:manage:moderators`

```go
err := client.RemoveChannelModerator(ctx, &helix.RemoveChannelModeratorParams{
    BroadcasterID: "12345",
    UserID:        "67890",
})

if err != nil {
    fmt.Printf("Failed to remove moderator: %v\n", err)
}
```

**Response:** This endpoint returns 204 No Content on success.

## DeleteChatMessages

Delete one or all messages in a broadcaster's chat room.

**Requires:** `moderator:manage:chat_messages`

```go
// Delete a specific message
err := client.DeleteChatMessages(ctx, &helix.DeleteChatMessagesParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    MessageID:     "abc-123-def",
})

// Delete all messages in chat
err = client.DeleteChatMessages(ctx, &helix.DeleteChatMessagesParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    // MessageID omitted to delete all messages
})

if err != nil {
    fmt.Printf("Failed to delete messages: %v\n", err)
}
```

**Response:** This endpoint returns 204 No Content on success.

## GetBlockedTerms

Get the broadcaster's list of blocked terms.

**Requires:** `moderator:read:blocked_terms`

```go
resp, err := client.GetBlockedTerms(ctx, &helix.GetBlockedTermsParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

for _, term := range resp.Data {
    fmt.Printf("Blocked term: %s (ID: %s)\n", term.Text, term.ID)
    fmt.Printf("Created at: %s\n", term.CreatedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "moderator_id": "67890",
      "id": "term-123-abc",
      "text": "badword",
      "created_at": "2025-12-10T10:00:00Z",
      "updated_at": "2025-12-10T10:00:00Z"
    },
    {
      "broadcaster_id": "12345",
      "moderator_id": "67890",
      "id": "term-456-def",
      "text": "spam",
      "created_at": "2025-12-09T14:30:00Z",
      "updated_at": "2025-12-09T14:30:00Z"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6Mn19"
  }
}
```

## AddBlockedTerm

Add a term to the broadcaster's list of blocked terms.

**Requires:** `moderator:manage:blocked_terms`

```go
resp, err := client.AddBlockedTerm(ctx, &helix.AddBlockedTermParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Text:          "badword",
})

if resp.Data != nil && len(resp.Data) > 0 {
    term := resp.Data[0]
    fmt.Printf("Added blocked term: %s (ID: %s)\n", term.Text, term.ID)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "moderator_id": "67890",
      "id": "term-789-ghi",
      "text": "badword",
      "created_at": "2025-12-11T15:00:00Z",
      "updated_at": "2025-12-11T15:00:00Z"
    }
  ]
}
```

## RemoveBlockedTerm

Remove a term from the broadcaster's list of blocked terms.

**Requires:** `moderator:manage:blocked_terms`

```go
err := client.RemoveBlockedTerm(ctx, &helix.RemoveBlockedTermParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    TermID:        "term-id-123",
})

if err != nil {
    fmt.Printf("Failed to remove blocked term: %v\n", err)
}
```

**Response:** This endpoint returns 204 No Content on success.

## GetShieldModeStatus

Get the broadcaster's Shield Mode activation status.

**Requires:** `moderator:read:shield_mode`

```go
resp, err := client.GetShieldModeStatus(ctx, &helix.GetShieldModeStatusParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
})

if resp.Data != nil && len(resp.Data) > 0 {
    status := resp.Data[0]
    fmt.Printf("Shield Mode Active: %v\n", status.IsActive)
    fmt.Printf("Last activated: %s\n", status.LastActivatedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "is_active": true,
      "moderator_id": "67890",
      "moderator_login": "coolmod",
      "moderator_name": "CoolMod",
      "last_activated_at": "2025-12-11T15:00:00Z"
    }
  ]
}
```

## UpdateShieldModeStatus

Activate or deactivate the broadcaster's Shield Mode.

**Requires:** `moderator:manage:shield_mode`

```go
// Activate Shield Mode
resp, err := client.UpdateShieldModeStatus(ctx, &helix.UpdateShieldModeStatusParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    IsActive:      true,
})

// Deactivate Shield Mode
resp, err = client.UpdateShieldModeStatus(ctx, &helix.UpdateShieldModeStatusParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    IsActive:      false,
})

if resp.Data != nil && len(resp.Data) > 0 {
    status := resp.Data[0]
    fmt.Printf("Shield Mode is now: %v\n", status.IsActive)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "is_active": false,
      "moderator_id": "67890",
      "moderator_login": "coolmod",
      "moderator_name": "CoolMod",
      "last_activated_at": "2025-12-11T15:00:00Z"
    }
  ]
}
```

## WarnChatUser

Warn a user in a broadcaster's chat room.

**Requires:** `moderator:manage:warnings`

```go
err := client.WarnChatUser(ctx, &helix.WarnChatUserParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Data: &helix.WarnChatUserData{
        UserID: "11111",
        Reason: "Please follow chat rules",
    },
})

if err != nil {
    fmt.Printf("Failed to warn user: %v\n", err)
}
```

**Response:** This endpoint returns 204 No Content on success.

## CheckAutoModStatus

Check whether AutoMod would flag one or more messages for review.

**Requires:** `moderation:read`

```go
resp, err := client.CheckAutoModStatus(ctx, &helix.CheckAutoModStatusParams{
    BroadcasterID: "12345",
    Data: []helix.AutoModMessage{
        {
            MsgID:   "msg-1",
            MsgText: "This is a test message",
        },
        {
            MsgID:   "msg-2",
            MsgText: "Another message to check",
        },
    },
})

for _, result := range resp.Data {
    fmt.Printf("Message %s: Approved=%v\n", result.MsgID, result.IsPermitted)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "msg_id": "msg-1",
      "is_permitted": true
    },
    {
      "msg_id": "msg-2",
      "is_permitted": false
    }
  ]
}
```

## ManageHeldAutoModMessages

Allow or deny a message that was held for review by AutoMod.

**Requires:** `moderator:manage:automod`

```go
// Allow a held message
err := client.ManageHeldAutoModMessages(ctx, &helix.ManageHeldAutoModMessagesParams{
    UserID:    "67890",
    MsgID:     "abc-123-def",
    Action:    "ALLOW",
})

// Deny a held message
err = client.ManageHeldAutoModMessages(ctx, &helix.ManageHeldAutoModMessagesParams{
    UserID:    "67890",
    MsgID:     "abc-123-def",
    Action:    "DENY",
})

if err != nil {
    fmt.Printf("Failed to manage AutoMod message: %v\n", err)
}
```

**Response:** This endpoint returns 204 No Content on success.

## GetAutoModSettings

Get the broadcaster's AutoMod settings.

**Requires:** `moderator:read:automod_settings`

```go
resp, err := client.GetAutoModSettings(ctx, &helix.GetAutoModSettingsParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
})

if resp.Data != nil && len(resp.Data) > 0 {
    settings := resp.Data[0]
    fmt.Printf("Overall Level: %d\n", settings.OverallLevel)
    fmt.Printf("Disability: %d\n", settings.Disability)
    fmt.Printf("Aggression: %d\n", settings.Aggression)
    fmt.Printf("Bullying: %d\n", settings.Bullying)
    fmt.Printf("Sexuality: %d\n", settings.SexualitySexOrGender)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "moderator_id": "67890",
      "overall_level": null,
      "disability": 2,
      "aggression": 3,
      "sexuality_sex_or_gender": 2,
      "misogyny": 3,
      "bullying": 2,
      "swearing": 1,
      "race_ethnicity_or_religion": 4,
      "sex_based_terms": 2
    }
  ]
}
```

## UpdateAutoModSettings

Update the broadcaster's AutoMod settings.

**Requires:** `moderator:manage:automod_settings`

```go
// Update with overall level
resp, err := client.UpdateAutoModSettings(ctx, &helix.UpdateAutoModSettingsParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    OverallLevel:  2, // 0-4 (0=disabled, 1-4=increasing strictness)
})

// Update with individual category levels
resp, err = client.UpdateAutoModSettings(ctx, &helix.UpdateAutoModSettingsParams{
    BroadcasterID:        "12345",
    ModeratorID:          "67890",
    Disability:           3,
    Aggression:           4,
    Bullying:             3,
    SexualitySexOrGender: 2,
    Misogyny:             3,
    RaceEthnicityReligion: 4,
    SexBasedTerms:        2,
    Swearing:             1,
})

if resp.Data != nil && len(resp.Data) > 0 {
    settings := resp.Data[0]
    fmt.Printf("AutoMod settings updated. Overall Level: %d\n", settings.OverallLevel)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "moderator_id": "67890",
      "overall_level": 2,
      "disability": 2,
      "aggression": 2,
      "sexuality_sex_or_gender": 2,
      "misogyny": 2,
      "bullying": 2,
      "swearing": 2,
      "race_ethnicity_or_religion": 2,
      "sex_based_terms": 2
    }
  ]
}
```

## GetUnbanRequests

Get a list of unban requests for a broadcaster's channel.

**Requires:** `moderator:read:unban_requests`

```go
// Get all pending unban requests
resp, err := client.GetUnbanRequests(ctx, &helix.GetUnbanRequestsParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    Status:        "pending",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

// Get unban requests for a specific user
resp, err = client.GetUnbanRequests(ctx, &helix.GetUnbanRequestsParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    UserID:        "11111",
    Status:        "approved", // pending, approved, denied, acknowledged, canceled
    PaginationParams: &helix.PaginationParams{
        First: 50,
    },
})

for _, request := range resp.Data {
    fmt.Printf("Request ID: %s\n", request.ID)
    fmt.Printf("User: %s (%s)\n", request.UserName, request.UserLogin)
    fmt.Printf("Status: %s\n", request.Status)
    fmt.Printf("Reason: %s\n", request.Text)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "request-123-abc",
      "broadcaster_id": "12345",
      "broadcaster_login": "streamer",
      "broadcaster_name": "Streamer",
      "user_id": "11111",
      "user_login": "banneduser",
      "user_name": "BannedUser",
      "text": "I apologize for my behavior and promise to follow the rules",
      "status": "pending",
      "created_at": "2025-12-10T10:00:00Z"
    },
    {
      "id": "request-456-def",
      "broadcaster_id": "12345",
      "broadcaster_login": "streamer",
      "broadcaster_name": "Streamer",
      "moderator_id": "67890",
      "moderator_login": "coolmod",
      "moderator_name": "CoolMod",
      "user_id": "22222",
      "user_login": "anotherbanned",
      "user_name": "AnotherBanned",
      "text": "Please give me another chance",
      "status": "approved",
      "created_at": "2025-12-09T14:30:00Z",
      "resolved_at": "2025-12-10T09:00:00Z",
      "resolution_text": "User has shown understanding of community rules"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6Mn19"
  }
}
```

## ResolveUnbanRequest

Resolve an unban request by approving or denying it.

**Requires:** `moderator:manage:unban_requests`

```go
// Approve an unban request
resp, err := client.ResolveUnbanRequest(ctx, &helix.ResolveUnbanRequestParams{
    BroadcasterID:   "12345",
    ModeratorID:     "67890",
    UnbanRequestID:  "request-id-123",
    Status:          "approved",
    ResolutionText:  "User has shown understanding of community rules",
})

// Deny an unban request
resp, err = client.ResolveUnbanRequest(ctx, &helix.ResolveUnbanRequestParams{
    BroadcasterID:   "12345",
    ModeratorID:     "67890",
    UnbanRequestID:  "request-id-456",
    Status:          "denied",
    ResolutionText:  "Request denied due to severity of violation",
})

if resp.Data != nil && len(resp.Data) > 0 {
    request := resp.Data[0]
    fmt.Printf("Unban request %s: %s\n", request.ID, request.Status)
    fmt.Printf("Resolution: %s\n", request.ResolutionText)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "request-123-abc",
      "broadcaster_id": "12345",
      "broadcaster_login": "streamer",
      "broadcaster_name": "Streamer",
      "moderator_id": "67890",
      "moderator_login": "coolmod",
      "moderator_name": "CoolMod",
      "user_id": "11111",
      "user_login": "banneduser",
      "user_name": "BannedUser",
      "text": "I apologize for my behavior and promise to follow the rules",
      "status": "approved",
      "created_at": "2025-12-10T10:00:00Z",
      "resolved_at": "2025-12-11T15:00:00Z",
      "resolution_text": "User has shown understanding of community rules"
    }
  ]
}
```

## GetModeratedChannels

Get a list of channels that the specified user has moderator privileges in.

**Requires:** `user:read:moderated_channels`

```go
resp, err := client.GetModeratedChannels(ctx, &helix.GetModeratedChannelsParams{
    UserID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

for _, channel := range resp.Data {
    fmt.Printf("Moderating: %s (%s)\n", channel.BroadcasterName, channel.BroadcasterLogin)
    fmt.Printf("Broadcaster ID: %s\n", channel.BroadcasterID)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "streamer1",
      "broadcaster_name": "Streamer1"
    },
    {
      "broadcaster_id": "67890",
      "broadcaster_login": "streamer2",
      "broadcaster_name": "Streamer2"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6Mn19"
  }
}
```

## AddSuspiciousUserStatus

Add a suspicious status to a chat user. Suspicious users can be marked as "restricted" (cannot chat) or "monitored" (messages are flagged for review).

**Requires:** `moderator:manage:suspicious_users`

```go
// Mark a user as restricted (cannot send messages)
err := client.AddSuspiciousUserStatus(ctx, &helix.AddSuspiciousUserStatusParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    UserID:        "11111",
    Status:        helix.SuspiciousUserStatusRestricted,
})

// Mark a user as monitored (messages flagged for review)
err = client.AddSuspiciousUserStatus(ctx, &helix.AddSuspiciousUserStatusParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    UserID:        "22222",
    Status:        helix.SuspiciousUserStatusMonitored,
})

if err != nil {
    fmt.Printf("Failed to add suspicious user status: %v\n", err)
}
```

**Status Values:**
- `helix.SuspiciousUserStatusRestricted` ("restricted") - User cannot send messages in chat
- `helix.SuspiciousUserStatusMonitored` ("monitored") - User's messages are flagged for moderator review

**Response:** This endpoint returns 204 No Content on success.

## RemoveSuspiciousUserStatus

Remove a suspicious status from a chat user, allowing them to chat normally again.

**Requires:** `moderator:manage:suspicious_users`

```go
err := client.RemoveSuspiciousUserStatus(ctx, &helix.RemoveSuspiciousUserStatusParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    UserID:        "11111",
})

if err != nil {
    fmt.Printf("Failed to remove suspicious user status: %v\n", err)
}
```
**Response:** This endpoint returns 204 No Content on success.

