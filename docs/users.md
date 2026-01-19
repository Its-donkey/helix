---
layout: default
title: Users API
description: Retrieve and manage Twitch user information, blocks, and extensions.
---

## GetUsers

Get information about one or more users by their user IDs or login names.

**Requires:** No authentication required for public data, `user:read:email` for email field

```go
// Get users by IDs
resp, err := client.GetUsers(ctx, &helix.GetUsersParams{
    IDs: []string{"12345", "67890"}, // Max 100
})

// Get users by login names
resp, err = client.GetUsers(ctx, &helix.GetUsersParams{
    Logins: []string{"twitchdev", "twitchapi"}, // Max 100
})

for _, user := range resp.Data {
    fmt.Printf("User: %s (%s)\n", user.DisplayName, user.Login)
    fmt.Printf("  ID: %s\n", user.ID)
    fmt.Printf("  Type: %s, Broadcaster Type: %s\n", user.Type, user.BroadcasterType)
    fmt.Printf("  Description: %s\n", user.Description)
    fmt.Printf("  Profile Image: %s\n", user.ProfileImageURL)
    fmt.Printf("  Offline Image: %s\n", user.OfflineImageURL)
    fmt.Printf("  View Count: %d (deprecated)\n", user.ViewCount)
    fmt.Printf("  Email: %s\n", user.Email) // Requires user:read:email scope
    fmt.Printf("  Created At: %s\n", user.CreatedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "141981764",
      "login": "twitchdev",
      "display_name": "TwitchDev",
      "type": "",
      "broadcaster_type": "partner",
      "description": "Supporting third-party developers building Twitch integrations from chatbots to game integrations.",
      "profile_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/8a6381c7-d0c0-4576-b179-38bd5ce1d6af-profile_image-300x300.png",
      "offline_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/3f13ab61-ec78-4fe6-8481-8682cb3b0ac2-channel_offline_image-1920x1080.png",
      "view_count": 5980557,
      "email": "not-real@email.com",
      "created_at": "2016-12-14T20:32:28Z"
    },
    {
      "id": "12826",
      "login": "twitch",
      "display_name": "Twitch",
      "type": "",
      "broadcaster_type": "partner",
      "description": "We're building the future of live entertainment.",
      "profile_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/twitch-profile_image-8a8c5be2e3b64a9a-300x300.png",
      "offline_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/twitch-channel_offline_image-404e3e605d0f61e7-1920x1080.png",
      "view_count": 124234235,
      "created_at": "2005-10-12T03:52:08Z"
    }
  ]
}
```

## GetCurrentUser

Get information about the authenticated user.

**Requires:** Valid user access token

```go
resp, err := client.GetCurrentUser(ctx)
if err != nil {
    fmt.Printf("Failed to get current user: %v\n", err)
}

user := resp.Data[0]
fmt.Printf("Current User: %s (%s)\n", user.DisplayName, user.Login)
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "141981764",
      "login": "twitchdev",
      "display_name": "TwitchDev",
      "type": "",
      "broadcaster_type": "partner",
      "description": "Supporting third-party developers building Twitch integrations from chatbots to game integrations.",
      "profile_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/8a6381c7-d0c0-4576-b179-38bd5ce1d6af-profile_image-300x300.png",
      "offline_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/3f13ab61-ec78-4fe6-8481-8682cb3b0ac2-channel_offline_image-1920x1080.png",
      "view_count": 5980557,
      "email": "not-real@email.com",
      "created_at": "2016-12-14T20:32:28Z"
    }
  ]
}
```

## UpdateUser

Update the description of the authenticated user.

**Requires:** `user:edit`

```go
err := client.UpdateUser(ctx, &helix.UpdateUserParams{
    Description: "New profile description here!",
})
if err != nil {
    fmt.Printf("Failed to update user: %v\n", err)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "141981764",
      "login": "twitchdev",
      "display_name": "TwitchDev",
      "type": "",
      "broadcaster_type": "partner",
      "description": "New profile description here!",
      "profile_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/8a6381c7-d0c0-4576-b179-38bd5ce1d6af-profile_image-300x300.png",
      "offline_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/3f13ab61-ec78-4fe6-8481-8682cb3b0ac2-channel_offline_image-1920x1080.png",
      "view_count": 5980557,
      "created_at": "2016-12-14T20:32:28Z"
    }
  ]
}
```

## GetUserBlockList

Get a list of users that the broadcaster has blocked.

**Requires:** `user:read:blocked_users`

```go
resp, err := client.GetUserBlockList(ctx, &helix.GetUserBlockListParams{
    BroadcasterID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

for _, blockedUser := range resp.Data {
    fmt.Printf("Blocked: %s (%s)\n", blockedUser.UserName, blockedUser.UserLogin)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "135093069",
      "user_login": "sampleuser123",
      "display_name": "SampleUser123"
    },
    {
      "user_id": "182891647",
      "user_login": "spambot456",
      "display_name": "SpamBot456"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## BlockUser

Block a user from interacting with the broadcaster.

**Requires:** `user:manage:blocked_users`

```go
err := client.BlockUser(ctx, &helix.BlockUserParams{
    TargetUserID:  "67890",
    SourceContext: "chat",           // "chat" or "whisper"
    Reason:        "spam",           // "spam", "harassment", or "other"
})
if err != nil {
    fmt.Printf("Failed to block user: %v\n", err)
}
```

**Sample Response:**
```json
{}
```
Note: BlockUser returns an empty response on success (HTTP 204 No Content).

## UnblockUser

Unblock a previously blocked user.

**Requires:** `user:manage:blocked_users`

```go
err := client.UnblockUser(ctx, "67890")
if err != nil {
    fmt.Printf("Failed to unblock user: %v\n", err)
}
```

**Sample Response:**
```json
{}
```
Note: UnblockUser returns an empty response on success (HTTP 204 No Content).

## GetUserExtensions

Get a list of all extensions (active and inactive) that the broadcaster has installed.

**Requires:** `user:read:broadcast`

```go
resp, err := client.GetUserExtensions(ctx)
for _, ext := range resp.Data {
    fmt.Printf("Extension: %s (v%s)\n", ext.Name, ext.Version)
    fmt.Printf("  ID: %s\n", ext.ID)
    fmt.Printf("  Type: %s\n", ext.Type)
    fmt.Printf("  Can Activate: %t\n", ext.CanActivate)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "wi08ebtatdc7oj83wtl9uxwz807l8b",
      "version": "1.1.8",
      "name": "Streamlabs Leaderboard",
      "can_activate": true,
      "type": ["panel"]
    },
    {
      "id": "d4uvtfdr04uliagaoajjg43c7kbvdm",
      "version": "2.0.2",
      "name": "Prime Subscription Extension",
      "can_activate": true,
      "type": ["overlay"]
    },
    {
      "id": "zfh2irvx2jb4s60f02jq0ajm8vwgka",
      "version": "1.0.0",
      "name": "Viewer Participation Extension",
      "can_activate": false,
      "type": ["component", "mobile"]
    }
  ]
}
```

## GetUserActiveExtensions

Get information about the broadcaster's active extensions.

**Requires:** `user:read:broadcast` or `user:edit:broadcast`

```go
// Get active extensions for authenticated user
resp, err := client.GetUserActiveExtensions(ctx, &helix.GetUserActiveExtensionsParams{})

// Get active extensions for specific user
resp, err = client.GetUserActiveExtensions(ctx, &helix.GetUserActiveExtensionsParams{
    UserID: "12345",
})

// Access panel, overlay, and component extensions
for slot, ext := range resp.Data.Panel {
    if ext != nil {
        fmt.Printf("Panel Slot %s: %s\n", slot, ext.Name)
    }
}

for slot, ext := range resp.Data.Overlay {
    if ext != nil {
        fmt.Printf("Overlay Slot %s: %s\n", slot, ext.Name)
    }
}

for slot, ext := range resp.Data.Component {
    if ext != nil {
        fmt.Printf("Component Slot %s: %s\n", slot, ext.Name)
    }
}
```

**Sample Response:**
```json
{
  "data": {
    "panel": {
      "1": {
        "active": true,
        "id": "wi08ebtatdc7oj83wtl9uxwz807l8b",
        "version": "1.1.8",
        "name": "Streamlabs Leaderboard"
      },
      "2": {
        "active": true,
        "id": "naty2zwfp7vecaivuve8ef1hohh6bo",
        "version": "1.0.9",
        "name": "Twitch Chat Embed"
      },
      "3": {
        "active": false
      }
    },
    "overlay": {
      "1": {
        "active": true,
        "id": "d4uvtfdr04uliagaoajjg43c7kbvdm",
        "version": "2.0.2",
        "name": "Prime Subscription Extension",
        "x": 0,
        "y": 0
      }
    },
    "component": {
      "1": {
        "active": true,
        "id": "lqnf3zxk0rv0g7gq92mtmnirjz2cjj",
        "version": "0.0.1",
        "name": "Dev Component",
        "x": 0,
        "y": 0
      },
      "2": {
        "active": false
      }
    }
  }
}
```

## UpdateUserExtensions

Update the activation state, extension ID, or version of the broadcaster's active extensions.

**Requires:** `user:edit:broadcast`

```go
err := client.UpdateUserExtensions(ctx, &helix.UpdateUserExtensionsParams{
    Data: &helix.UserActiveExtensions{
        Panel: map[string]*helix.UserActiveExtension{
            "1": {
                Active: true,
                ID:     "extension_id",
                Version: "1.0.0",
            },
        },
        Overlay: map[string]*helix.UserActiveExtension{
            "1": {
                Active: true,
                ID:     "overlay_extension_id",
                Version: "1.0.0",
            },
        },
        Component: map[string]*helix.UserActiveExtension{
            "1": {
                Active: true,
                ID:     "component_extension_id",
                Version: "1.0.0",
            },
        },
    },
})
if err != nil {
    fmt.Printf("Failed to update user extensions: %v\n", err)
}
```

**Sample Response:**
```json
{
  "data": {
    "panel": {
      "1": {
        "active": true,
        "id": "extension_id",
        "version": "1.0.0",
        "name": "My Panel Extension"
      },
      "2": {
        "active": false
      },
      "3": {
        "active": false
      }
    },
    "overlay": {
      "1": {
        "active": true,
        "id": "overlay_extension_id",
        "version": "1.0.0",
        "name": "My Overlay Extension",
        "x": 0,
        "y": 0
      }
    },
    "component": {
      "1": {
        "active": true,
        "id": "component_extension_id",
        "version": "1.0.0",
        "name": "My Component Extension",
        "x": 0,
        "y": 0
      },
      "2": {
        "active": false
      }
    }
  }
}
```

## GetAuthorizationByUser

Get authorization information for a user who has authorized your app.

**Requires:** App access token

```go
resp, err := client.GetAuthorizationByUser(ctx, &helix.GetAuthorizationByUserParams{
    UserID: "12345",
})

auth := resp.Data[0]
fmt.Printf("Client ID: %s\n", auth.ClientID)
fmt.Printf("User: %s (%s)\n", auth.Login, auth.UserID)
fmt.Printf("Scopes: %v\n", auth.Scopes)
```

**Sample Response:**
```json
{
  "data": [
    {
      "client_id": "hof5gwx0su6owfnys0yan9c87zr6t",
      "user_id": "141981764",
      "login": "twitchdev",
      "scopes": [
        "channel:read:subscriptions",
        "user:read:email",
        "user:edit",
        "chat:read",
        "chat:edit"
      ]
    }
  ]
}
```

