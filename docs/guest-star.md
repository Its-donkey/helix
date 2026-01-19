---
layout: default
title: Guest Star API
description: > **Note:** These are BETA endpoints and may be subject to changes.
---

## GetChannelGuestStarSettings

Get the channel settings for configuration of the Guest Star feature for a particular host.

**Requires:** `channel:read:guest_star`, `channel:manage:guest_star`, or `moderator:read:guest_star`

```go
resp, err := client.GetChannelGuestStarSettings(ctx, &helix.GetChannelGuestStarSettingsParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
})
if err != nil {
    fmt.Printf("Failed to get settings: %v\n", err)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "is_moderator_send_live_enabled": true,
      "slot_count": 6,
      "is_browser_source_audio_enabled": true,
      "group_layout": "TILED_LAYOUT",
      "browser_source_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
    }
  ]
}
```

## UpdateChannelGuestStarSettings

Update the channel settings for configuration of the Guest Star feature for a particular host.

**Requires:** `channel:manage:guest_star`

```go
err := client.UpdateChannelGuestStarSettings(ctx, &helix.UpdateChannelGuestStarSettingsParams{
    BroadcasterID:                "12345",
    IsModeratorSendLiveEnabled:   true,
    SlotCount:                    6,
    IsBrowserSourceAudioEnabled:  true,
    GroupLayout:                  "TILED_LAYOUT", // Options: TILED_LAYOUT, SCREENSHARE_LAYOUT, HORIZONTAL_LAYOUT, VERTICAL_LAYOUT
    RegenerateBrowserSources:     false,
})
if err != nil {
    fmt.Printf("Failed to update settings: %v\n", err)
}
```

**Returns:** No response body on success (HTTP 204 No Content)

## GetGuestStarSession

Get the information for the active Guest Star session for a particular channel.

**Requires:** `channel:read:guest_star`, `channel:manage:guest_star`, or `moderator:read:guest_star`

```go
resp, err := client.GetGuestStarSession(ctx, &helix.GetGuestStarSessionParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
})
if err != nil {
    fmt.Printf("Failed to get session: %v\n", err)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "v4-1234567890abcdef",
      "guests": [
        {
          "slot_id": "0",
          "is_live": true,
          "user_id": "98765",
          "user_display_name": "GuestUser",
          "user_login": "guestuser",
          "volume": 100,
          "assigned_at": "2024-01-15T10:30:00Z",
          "audio_settings": {
            "is_host_enabled": true,
            "is_self_muted": false,
            "is_available": true
          },
          "video_settings": {
            "is_host_enabled": true,
            "is_self_muted": false,
            "is_available": true
          }
        },
        {
          "slot_id": "1",
          "is_live": false,
          "user_id": "54321",
          "user_display_name": "AnotherGuest",
          "user_login": "anotherguest",
          "volume": 75,
          "assigned_at": "2024-01-15T10:35:00Z",
          "audio_settings": {
            "is_host_enabled": true,
            "is_self_muted": true,
            "is_available": true
          },
          "video_settings": {
            "is_host_enabled": false,
            "is_self_muted": false,
            "is_available": true
          }
        }
      ]
    }
  ]
}
```

## CreateGuestStarSession

Create a Guest Star session for a particular broadcaster.

**Requires:** `channel:manage:guest_star`

```go
resp, err := client.CreateGuestStarSession(ctx, "12345")
if err != nil {
    fmt.Printf("Failed to create session: %v\n", err)
}
fmt.Printf("Session created with ID: %s\n", resp.Data.SessionID)
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "v4-1234567890abcdef",
      "guests": []
    }
  ]
}
```

## EndGuestStarSession

End an active Guest Star session.

**Requires:** `channel:manage:guest_star`

```go
err := client.EndGuestStarSession(ctx, &helix.EndGuestStarSessionParams{
    BroadcasterID: "12345",
    SessionID:     "session-id-here",
})
if err != nil {
    fmt.Printf("Failed to end session: %v\n", err)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "v4-1234567890abcdef",
      "guests": []
    }
  ]
}
```

## GetGuestStarInvites

Get the list of pending invites for a particular Guest Star session.

**Requires:** `channel:read:guest_star`, `channel:manage:guest_star`, or `moderator:read:guest_star`

```go
resp, err := client.GetGuestStarInvites(ctx, &helix.GetGuestStarInvitesParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    SessionID:     "session-id-here",
})
for _, invite := range resp.Data {
    fmt.Printf("Invite for user: %s\n", invite.UserID)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "98765",
      "invited_at": "2024-01-15T10:25:00Z",
      "status": "INVITED",
      "is_video_enabled": true,
      "is_audio_enabled": true,
      "is_video_available": true,
      "is_audio_available": true
    },
    {
      "user_id": "54321",
      "invited_at": "2024-01-15T10:28:00Z",
      "status": "ACCEPTED",
      "is_video_enabled": true,
      "is_audio_enabled": true,
      "is_video_available": true,
      "is_audio_available": true
    }
  ]
}
```

## SendGuestStarInvite

Send a Guest Star invite to a user.

**Requires:** `channel:manage:guest_star` or `moderator:manage:guest_star`

```go
err := client.SendGuestStarInvite(ctx, &helix.SendGuestStarInviteParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    SessionID:     "session-id-here",
    GuestID:       "guest-user-id",
})
if err != nil {
    fmt.Printf("Failed to send invite: %v\n", err)
}
```

**Returns:** No response body on success (HTTP 204 No Content)

## DeleteGuestStarInvite

Revoke a Guest Star invite for a user.

**Requires:** `channel:manage:guest_star` or `moderator:manage:guest_star`

```go
err := client.DeleteGuestStarInvite(ctx, &helix.DeleteGuestStarInviteParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    SessionID:     "session-id-here",
    GuestID:       "guest-user-id",
})
if err != nil {
    fmt.Printf("Failed to delete invite: %v\n", err)
}
```

**Returns:** No response body on success (HTTP 204 No Content)

## AssignGuestStarSlot

Assign a user to a slot in a Guest Star session.

**Requires:** `channel:manage:guest_star` or `moderator:manage:guest_star`

```go
err := client.AssignGuestStarSlot(ctx, &helix.AssignGuestStarSlotParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    SessionID:     "session-id-here",
    GuestID:       "guest-user-id",
    SlotID:        "1",
})
if err != nil {
    fmt.Printf("Failed to assign slot: %v\n", err)
}
```

**Returns:** No response body on success (HTTP 204 No Content)

## UpdateGuestStarSlot

Move a user between slots within the active Guest Star session.

**Requires:** `channel:manage:guest_star` or `moderator:manage:guest_star`

```go
err := client.UpdateGuestStarSlot(ctx, &helix.UpdateGuestStarSlotParams{
    BroadcasterID:      "12345",
    ModeratorID:        "67890",
    SessionID:          "session-id-here",
    SourceSlotID:       "1",
    DestinationSlotID:  "2",
})
if err != nil {
    fmt.Printf("Failed to update slot: %v\n", err)
}
```

**Returns:** No response body on success (HTTP 204 No Content)

## DeleteGuestStarSlot

Remove a user from a slot in the active Guest Star session.

**Requires:** `channel:manage:guest_star` or `moderator:manage:guest_star`

```go
err := client.DeleteGuestStarSlot(ctx, &helix.DeleteGuestStarSlotParams{
    BroadcasterID: "12345",
    ModeratorID:   "67890",
    SessionID:     "session-id-here",
    GuestID:       "guest-user-id",
    SlotID:        "1",
})
if err != nil {
    fmt.Printf("Failed to delete slot: %v\n", err)
}
```

**Returns:** No response body on success (HTTP 204 No Content)

## UpdateGuestStarSlotSettings

Update settings for a particular guest's slot in an active Guest Star session.

**Requires:** `channel:manage:guest_star` or `moderator:manage:guest_star`

```go
err := client.UpdateGuestStarSlotSettings(ctx, &helix.UpdateGuestStarSlotSettingsParams{
    BroadcasterID:  "12345",
    ModeratorID:    "67890",
    SessionID:      "session-id-here",
    SlotID:         "1",
    IsAudioEnabled: true,
    IsVideoEnabled: true,
    IsLive:         true,
    Volume:         100,
})
if err != nil {
    fmt.Printf("Failed to update slot settings: %v\n", err)
}
```
**Returns:** No response body on success (HTTP 204 No Content)

