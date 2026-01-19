---
layout: default
title: Extension JWT Authentication
description: For Twitch Extensions that require JWT authentication.
---

## Overview

Twitch Extensions use JSON Web Tokens (JWT) to securely authenticate requests between your Extension Backend Service (EBS) and Twitch's API. JWTs are signed with your extension's secret, ensuring requests are authorized.

**When you need JWT authentication**:
- Sending messages to the extension PubSub
- Accessing extension configuration segments
- Making API calls that require extension context
- Authenticating requests from your EBS to Twitch

**Key concepts**:
- **Extension Secret**: A base64-encoded secret from your Extension settings, used to sign JWTs
- **EBS Token**: Authenticates your backend service to Twitch
- **Broadcaster Token**: Performs actions on behalf of a specific broadcaster's channel

## Creating JWT Tokens

Create tokens for different authorization contexts. The JWT handler manages signing and claims automatically.

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    // Create JWT handler
    jwt, err := helix.NewExtensionJWT(
        "your-extension-id",
        "base64-encoded-secret", // From Extension settings
        "extension-owner-user-id",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create EBS (Extension Backend Service) token
    ebsToken, err := jwt.CreateEBSToken(time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("EBS Token: %s\n", ebsToken)

    // Create broadcaster token for specific channel
    broadcasterToken, err := jwt.CreateBroadcasterToken("channel-id", time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Broadcaster Token: %s\n", broadcasterToken)
}
```

## Using with Extension Client

The `ExtensionClient` automatically handles JWT token creation and attachment for extension API calls.

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    ctx := context.Background()

    // Create JWT handler
    jwt, err := helix.NewExtensionJWT(
        "your-extension-id",
        "base64-encoded-secret",
        "extension-owner-user-id",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create a client with extension JWT auth
    client := helix.NewExtensionClient("your-client-id", jwt)

    // Make extension API calls
    resp, err := client.GetExtensionConfigurationSegment(ctx, &helix.GetExtensionConfigurationSegmentParams{
        ExtensionID: "your-extension-id",
        Segment:     []string{"broadcaster"},
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, config := range resp.Data {
        fmt.Printf("Segment: %s, Content: %s\n", config.Segment, config.Content)
    }
}
```

## Adding JWT to Existing Client

```go
// Create standard auth client first
authClient := helix.NewAuthClient(helix.AuthConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
})

// Create Helix client
client := helix.NewClient("your-client-id", authClient)

// Create JWT handler
jwt, _ := helix.NewExtensionJWT(
    "your-extension-id",
    "base64-encoded-secret",
    "extension-owner-user-id",
)

// Add JWT to existing client for extension endpoints
client.SetExtensionJWT(jwt)
```

## Token Types

Different token types authorize different actions. Choose the appropriate type based on what your request needs to do.

### EBS Token

General-purpose token for your Extension Backend Service. Use this for operations that don't require specific broadcaster context.

```go
ebsToken, _ := jwt.CreateEBSToken(time.Hour)
```

### Broadcaster Token

Used when performing actions in the context of a specific broadcaster's channel. Required for channel-specific operations like sending PubSub messages to a channel or reading channel configuration.

```go
broadcasterToken, _ := jwt.CreateBroadcasterToken("channel-id", time.Hour)
```

### Custom Claims

For advanced use cases where you need fine-grained control over token permissions, including PubSub listen/send permissions and custom roles.

```go
token, _ := jwt.CreateToken(helix.ExtensionJWTClaims{
    UserID:    "user-id",
    ChannelID: "channel-id",
    Role:      "broadcaster",
    Pubsub: &helix.PubsubClaims{
        Listen: []string{"broadcast"},
        Send:   []string{"broadcast"},
    },
}, time.Hour)
```

## Extension API Endpoints
See [Extensions documentation](extensions.md) for all available extension endpoints.

