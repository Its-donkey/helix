---
layout: default
title: Extension JWT Authentication
description: Certain Twitch Extension API endpoints require a special **Extension JWT** (JSON Web Token) instead of a regular OAuth access token.
---

## How It Works

1. Your extension has a **secret** (obtained from the Twitch Developer Console or via `GetExtensionSecrets`)
2. Your Extension Backend Service (EBS) creates and signs a JWT using that secret
3. The JWT is passed as the Bearer token in API requests

## Creating an Extension JWT

```go
import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

func CreateExtensionJWT(extensionSecret, userID, channelID string) (string, error) {
    claims := jwt.MapClaims{
        "exp":        time.Now().Add(time.Minute * 3).Unix(),
        "user_id":    userID,      // Your extension's owner user ID
        "role":       "external",  // "external" for EBS calls
        "channel_id": channelID,   // Target channel (if applicable)
        "pubsub_perms": map[string][]string{
            "send": {"broadcast", "global"},
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(extensionSecret))
}
```

## JWT Claims

| Claim | Description |
|-------|-------------|
| `exp` | Expiration time (Unix timestamp). Max 3 minutes from now. |
| `user_id` | The user ID making the request (typically extension owner). |
| `role` | Set to `"external"` for EBS calls. |
| `channel_id` | The target channel ID (required for channel-specific operations). |
| `pubsub_perms` | PubSub permissions object (required for PubSub messages). |

## Auth Type Comparison

| Auth Type | Used For |
|-----------|----------|
| Extension JWT | Extension-specific endpoints (config, PubSub, secrets, chat) |
| App Access Token | Public extension data (Bits products, transactions) |
| User Access Token | User-authorized actions |

## Example Usage

```go
// Create JWT
secret := "your-extension-secret"
token, err := CreateExtensionJWT(secret, "12345", "67890")
if err != nil {
    log.Fatal(err)
}

// Use with client
client := helix.NewClient("client-id", token)
resp, err := client.GetExtensionConfigurationSegment(ctx, &helix.GetExtensionConfigurationSegmentParams{
    ExtensionID:   "your-extension-id",
    Segment:       []string{"broadcaster"},
    BroadcasterID: "67890",
})
```

