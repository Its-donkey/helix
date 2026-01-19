---
layout: default
title: Extensions API
description: Some endpoints require [Extension JWT authentication](Projects/Programming/Kappopher/Documents/extension-jwt.md) instead of standard OAuth tokens.
---

## GetExtensionConfigurationSegment

Get extension configuration segments (broadcaster, developer, or global).

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
resp, err := client.GetExtensionConfigurationSegment(ctx, &helix.GetExtensionConfigurationSegmentParams{
    ExtensionID:   "your-extension-id",
    Segment:       []string{"broadcaster", "developer", "global"},
    BroadcasterID: "12345", // Required when segment includes "broadcaster"
})
for _, segment := range resp.Data {
    fmt.Printf("Segment: %s, Content: %s\n", segment.Segment, segment.Content)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "segment": "broadcaster",
      "broadcaster_id": "12345",
      "content": "{\"setting1\": \"value1\", \"setting2\": true}",
      "version": "1.0"
    },
    {
      "segment": "developer",
      "content": "{\"debug_mode\": true, \"api_endpoint\": \"https://api.example.com\"}",
      "version": "1.0"
    },
    {
      "segment": "global",
      "content": "{\"max_users\": 1000, \"feature_flag\": \"enabled\"}",
      "version": "1.0"
    }
  ]
}
```

## SetExtensionConfigurationSegment

Set extension configuration segment data.

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
err := client.SetExtensionConfigurationSegment(ctx, &helix.SetExtensionConfigurationSegmentParams{
    ExtensionID:   "your-extension-id",
    Segment:       "broadcaster",
    BroadcasterID: "12345",
    Content:       `{"setting1": "value1", "setting2": true}`,
    Version:       "1.0",
})
```

**Sample Response:**
No response body on success (204 No Content)

## SetExtensionRequiredConfiguration

Set the required configuration for an extension.

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
err := client.SetExtensionRequiredConfiguration(ctx, &helix.SetExtensionRequiredConfigurationParams{
    ExtensionID:           "your-extension-id",
    ExtensionVersion:      "1.0.0",
    RequiredConfiguration: "config-string-v1",
})
```

**Sample Response:**
No response body on success (204 No Content)

## SendExtensionPubSubMessage

Send a PubSub message to extension users.

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
// Broadcast to all users viewing the channel
err := client.SendExtensionPubSubMessage(ctx, &helix.SendExtensionPubSubMessageParams{
    Target:        []string{"broadcast"},
    BroadcasterID: "12345",
    Message:       `{"type": "update", "data": {"score": 100}}`,
})

// Whisper to a specific user
err = client.SendExtensionPubSubMessage(ctx, &helix.SendExtensionPubSubMessageParams{
    Target:        []string{"whisper-67890"},
    BroadcasterID: "12345",
    Message:       `{"type": "private", "data": {"reward": "badge"}}`,
})

// Global broadcast
err = client.SendExtensionPubSubMessage(ctx, &helix.SendExtensionPubSubMessageParams{
    Target:            []string{"global"},
    BroadcasterID:     "12345",
    IsGlobalBroadcast: true,
    Message:           `{"type": "announcement"}`,
})
```

**Sample Response:**
No response body on success (204 No Content)

## GetExtensionLiveChannels

Get live channels that have an extension installed and activated.

```go
resp, err := client.GetExtensionLiveChannels(ctx, &helix.GetExtensionLiveChannelsParams{
    ExtensionID: "your-extension-id",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})
for _, channel := range resp.Data {
    fmt.Printf("Channel: %s playing %s\n", channel.BroadcasterName, channel.GameName)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "game_name": "Science & Technology",
      "game_id": "509670",
      "title": "TwitchDev Extension Development Workshop"
    },
    {
      "broadcaster_id": "98765432",
      "broadcaster_name": "ExampleStreamer",
      "game_name": "Just Chatting",
      "game_id": "509658",
      "title": "Testing Extensions with Chat"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6MjB9fQ"
  }
}
```

## GetExtensionSecrets

Get the secrets for an extension.

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
resp, err := client.GetExtensionSecrets(ctx, "your-extension-id")
for _, secret := range resp.Data {
    fmt.Printf("Format Version: %d\n", secret.FormatVersion)
    for _, s := range secret.Secrets {
        fmt.Printf("Active at: %s, Expires at: %s\n", s.ActiveAt, s.ExpiresAt)
    }
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "format_version": 1,
      "secrets": [
        {
          "content": "Jzdmx0rGauA3xwkKJD0N4pSXMDdxb2Cs98TwJ4bPKr",
          "active_at": "2023-01-15T10:00:00Z",
          "expires_at": "2023-02-15T10:00:00Z"
        },
        {
          "content": "K8dmx1sHbuB4yxlLKE1O5qTYNEeyc3Dt09UxK5cQLs",
          "active_at": "2023-02-15T10:00:00Z",
          "expires_at": "2023-03-15T10:00:00Z"
        }
      ]
    }
  ]
}
```

## CreateExtensionSecret

Create a new secret for an extension.

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
// Create a new secret with a 300-second delay before activation
resp, err := client.CreateExtensionSecret(ctx, "your-extension-id", 300)
for _, secret := range resp.Data {
    fmt.Printf("New secret created, format version: %d\n", secret.FormatVersion)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "format_version": 1,
      "secrets": [
        {
          "content": "Jzdmx0rGauA3xwkKJD0N4pSXMDdxb2Cs98TwJ4bPKr",
          "active_at": "2023-01-15T10:00:00Z",
          "expires_at": "2023-02-15T10:00:00Z"
        },
        {
          "content": "M9enx2tIcvC5zymMLF2P6rUZOFfzd4Eu10VyL6dRMt",
          "active_at": "2023-01-15T10:05:00Z",
          "expires_at": "2023-02-15T10:05:00Z"
        }
      ]
    }
  ]
}
```

## SendExtensionChatMessage

Send a chat message from an extension.

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
err := client.SendExtensionChatMessage(ctx, &helix.SendExtensionChatMessageParams{
    BroadcasterID:    "12345",
    Text:             "Hello from the extension!",
    ExtensionID:      "your-extension-id",
    ExtensionVersion: "1.0.0",
})
```

**Sample Response:**
No response body on success (204 No Content)

## GetExtensions

Get information about an extension (any version).

**Requires:** [Extension JWT](Projects/Programming/Kappopher/Documents/extension-jwt.md)

```go
// Get latest version
resp, err := client.GetExtensions(ctx, "your-extension-id", "")

// Get specific version
resp, err = client.GetExtensions(ctx, "your-extension-id", "1.0.0")

for _, ext := range resp.Data {
    fmt.Printf("Extension: %s by %s (v%s)\n", ext.Name, ext.AuthorName, ext.Version)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "author_name": "TwitchDev",
      "bits_enabled": true,
      "can_install": true,
      "configuration_location": "hosted",
      "description": "A powerful extension for interactive streaming",
      "eula_tos_url": "https://example.com/eula",
      "has_chat_support": true,
      "icon_url": "https://static-cdn.jtvnw.net/extension-icon-256.png",
      "icon_urls": {
        "100x100": "https://static-cdn.jtvnw.net/extension-icon-100.png",
        "24x24": "https://static-cdn.jtvnw.net/extension-icon-24.png",
        "300x200": "https://static-cdn.jtvnw.net/extension-icon-300x200.png"
      },
      "id": "your-extension-id",
      "name": "Interactive Extension",
      "privacy_policy_url": "https://example.com/privacy",
      "request_identity_link": false,
      "screenshot_urls": [
        "https://static-cdn.jtvnw.net/screenshot1.png",
        "https://static-cdn.jtvnw.net/screenshot2.png"
      ],
      "state": "Released",
      "subscriptions_support_level": "optional",
      "summary": "Enhance your streams with interactive features",
      "support_email": "support@example.com",
      "version": "1.0.0",
      "viewer_summary": "Interact with your favorite streamers",
      "views": {
        "mobile": {
          "viewer_url": "https://example.com/mobile.html"
        },
        "panel": {
          "viewer_url": "https://example.com/panel.html",
          "height": 300,
          "can_link_external_content": false
        },
        "video_overlay": {
          "viewer_url": "https://example.com/overlay.html",
          "can_link_external_content": true
        },
        "component": {
          "viewer_url": "https://example.com/component.html",
          "can_link_external_content": false
        }
      },
      "allowlisted_config_urls": [
        "https://example.com/config"
      ],
      "allowlisted_panel_urls": [
        "https://example.com/panel"
      ]
    }
  ]
}
```

## GetReleasedExtensions

Get information about a released extension.

```go
resp, err := client.GetReleasedExtensions(ctx, "your-extension-id", "")
for _, ext := range resp.Data {
    fmt.Printf("Released: %s v%s\n", ext.Name, ext.Version)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "author_name": "TwitchDev",
      "bits_enabled": true,
      "can_install": true,
      "configuration_location": "hosted",
      "description": "A powerful extension for interactive streaming",
      "eula_tos_url": "https://example.com/eula",
      "has_chat_support": true,
      "icon_url": "https://static-cdn.jtvnw.net/extension-icon-256.png",
      "icon_urls": {
        "100x100": "https://static-cdn.jtvnw.net/extension-icon-100.png",
        "24x24": "https://static-cdn.jtvnw.net/extension-icon-24.png",
        "300x200": "https://static-cdn.jtvnw.net/extension-icon-300x200.png"
      },
      "id": "your-extension-id",
      "name": "Interactive Extension",
      "privacy_policy_url": "https://example.com/privacy",
      "request_identity_link": false,
      "screenshot_urls": [
        "https://static-cdn.jtvnw.net/screenshot1.png",
        "https://static-cdn.jtvnw.net/screenshot2.png"
      ],
      "state": "Released",
      "subscriptions_support_level": "optional",
      "summary": "Enhance your streams with interactive features",
      "support_email": "support@example.com",
      "version": "1.2.0",
      "viewer_summary": "Interact with your favorite streamers",
      "views": {
        "mobile": {
          "viewer_url": "https://example.com/mobile.html"
        },
        "panel": {
          "viewer_url": "https://example.com/panel.html",
          "height": 300,
          "can_link_external_content": false
        },
        "video_overlay": {
          "viewer_url": "https://example.com/overlay.html",
          "can_link_external_content": true
        },
        "component": {
          "viewer_url": "https://example.com/component.html",
          "can_link_external_content": false
        }
      },
      "allowlisted_config_urls": [
        "https://example.com/config"
      ],
      "allowlisted_panel_urls": [
        "https://example.com/panel"
      ]
    }
  ]
}
```

## GetExtensionBitsProducts

Get Bits products for an extension.

**Requires:** App access token for the extension

```go
// Get only active products
resp, err := client.GetExtensionBitsProducts(ctx, nil)

// Get all products including disabled/expired
resp, err = client.GetExtensionBitsProducts(ctx, &helix.GetExtensionBitsProductsParams{
    ShouldIncludeAll: true,
})

for _, product := range resp.Data {
    fmt.Printf("%s (%s): %d bits\n", product.DisplayName, product.SKU, product.Cost.Amount)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "sku": "power-up-001",
      "cost": {
        "amount": 100,
        "type": "bits"
      },
      "in_development": false,
      "display_name": "Super Power-Up",
      "expiration": "",
      "is_broadcast": true
    },
    {
      "sku": "special-badge",
      "cost": {
        "amount": 500,
        "type": "bits"
      },
      "in_development": false,
      "display_name": "Exclusive Badge",
      "expiration": "",
      "is_broadcast": false
    }
  ]
}
```

## UpdateExtensionBitsProduct

Create or update an extension Bits product.

**Requires:** App access token for the extension

```go
resp, err := client.UpdateExtensionBitsProduct(ctx, &helix.UpdateExtensionBitsProductParams{
    SKU:         "power-up-001",
    DisplayName: "Super Power-Up",
    Cost: helix.ExtensionBitsCost{
        Amount: 100,
        Type:   "bits",
    },
    InDevelopment: false,
    IsBroadcast:   true,
})
```

**Sample Response:**
```json
{
  "data": [
    {
      "sku": "power-up-001",
      "cost": {
        "amount": 100,
        "type": "bits"
      },
      "in_development": false,
      "display_name": "Super Power-Up",
      "expiration": "",
      "is_broadcast": true
    }
  ]
}
```

## GetExtensionTransactions

Get extension Bits transactions.

**Requires:** App access token for the extension

```go
// Get all recent transactions
resp, err := client.GetExtensionTransactions(ctx, &helix.GetExtensionTransactionsParams{
    ExtensionID: "your-extension-id",
    PaginationParams: &helix.PaginationParams{
        First: 50,
    },
})

// Get specific transactions by ID
resp, err = client.GetExtensionTransactions(ctx, &helix.GetExtensionTransactionsParams{
    ExtensionID: "your-extension-id",
    IDs:         []string{"transaction-id-1", "transaction-id-2"},
})

for _, tx := range resp.Data {
    fmt.Printf("%s bought %s for %d bits\n",
        tx.UserName, tx.ProductData.DisplayName, tx.ProductData.Cost.Amount)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "74c52265-e214-48a6-91b9-23b6014e8041",
      "timestamp": "2023-01-15T10:30:00Z",
      "broadcaster_id": "141981764",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "user_id": "98765432",
      "user_login": "vieweruser",
      "user_name": "ViewerUser",
      "product_type": "BITS_IN_EXTENSION",
      "product_data": {
        "sku": "power-up-001",
        "cost": {
          "amount": 100,
          "type": "bits"
        },
        "display_name": "Super Power-Up",
        "in_development": false
      }
    },
    {
      "id": "85d63376-f325-59b7-a2c0-34c7125f9f52",
      "timestamp": "2023-01-15T11:45:00Z",
      "broadcaster_id": "141981764",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "user_id": "87654321",
      "user_login": "anotherviewer",
      "user_name": "AnotherViewer",
      "product_type": "BITS_IN_EXTENSION",
      "product_data": {
        "sku": "special-badge",
        "cost": {
          "amount": 500,
          "type": "bits"
        },
        "display_name": "Exclusive Badge",
        "in_development": false
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6MjB9fQ"
  }
}
```

