---
layout: default
title: Kappopher Documentation
---

<div style="text-align: center; margin-bottom: 2rem;">
  <img src="{{ '/assets/images/logo.png' | relative_url }}" alt="Kappopher" style="width: 180px; height: 180px;">
  <h1 style="margin-top: 1rem;">Kappopher</h1>
  <p style="font-size: 1.125rem; color: #53535F;">A comprehensive Go wrapper for the Twitch Helix API with full endpoint coverage, multiple authentication flows, and real-time event support.</p>
</div>

## Quick Links

### [Getting Started](quickstart.md)
Get up and running with Kappopher in minutes.

### [API Reference](api-reference.md)
Complete documentation for all Twitch Helix API endpoints.

### [Cookbook](cookbook.md)
Practical code examples and recipes for common use cases.

## Features

- **Full Helix API Coverage** - All endpoints implemented with typed responses
- **Multiple Auth Flows** - Client Credentials, Authorization Code, Device Code, Implicit
- **EventSub Support** - WebSocket and Webhook handlers for real-time events
- **IRC Client** - Built-in chat client for bot development
- **Batch Operations** - Concurrent requests with rate limiting
- **Caching** - Built-in response caching with configurable TTL

## Installation

```bash
go get github.com/Its-donkey/kappopher
```

## Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Create auth client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Create Helix client
    client := helix.NewClient("your-client-id", authClient)

    // Get user info
    users, err := client.GetUsers(ctx, &helix.GetUsersParams{
        Logins: []string{"twitchdev"},
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, user := range users.Data {
        fmt.Printf("User: %s (ID: %s)\n", user.DisplayName, user.ID)
    }
}
```
