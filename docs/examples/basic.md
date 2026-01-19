---
layout: default
title: Basic Usage
description: Simple examples for getting started with the Twitch Helix API.
---

## Overview

This guide covers the fundamental operations you'll need when working with the Twitch Helix API:

- **Authentication**: Obtain an app access token using client credentials
- **User Operations**: Fetch user information by username or ID
- **Stream Data**: Check if channels are live and get stream details
- **Channel Information**: Retrieve channel metadata like title and game
- **Search**: Find channels and games/categories
- **Pagination**: Handle large result sets efficiently
- **Error Handling**: Properly handle API errors and edge cases

These examples use app access tokens (client credentials flow), which are suitable for public data that doesn't require user authorization. For user-specific data, see the [Authentication Examples](authentication.md).

## Setup

Before making any API calls, you need to:
1. Create an `AuthClient` with your Twitch application credentials
2. Obtain an access token (app token for public data, user token for private data)
3. Create a `Client` instance that handles all API requests

The `AuthClient` manages token acquisition, refresh, and storage. The `Client` uses the `AuthClient` to automatically attach authorization headers to requests.

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

    // Create auth client with your Twitch Developer Console credentials
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Get app access token - this authenticates your application
    // App tokens are used for public data that doesn't require user authorization
    token, err := authClient.GetAppAccessToken(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Token expires in: %d seconds\n", token.ExpiresIn)

    // Create Helix client - this is your main interface to the Twitch API
    // The client automatically uses the auth client for token management
    client := helix.NewClient("your-client-id", authClient)

    // Now you can make API calls
    // ...
}
```

## Get User Information

Retrieve detailed information about Twitch users. You can look up users by their login name (username) or by their numeric user ID. This is often the first API call you'll make since many other endpoints require user IDs.

**What you get**: Display name, profile image URL, account creation date, broadcaster type (affiliate/partner), and more.

```go
// Get user by login name
users, err := client.GetUsers(ctx, &helix.GetUsersParams{
    Logins: []string{"twitchdev"},
})
if err != nil {
    log.Fatal(err)
}

for _, user := range users.Data {
    fmt.Printf("User: %s\n", user.DisplayName)
    fmt.Printf("  ID: %s\n", user.ID)
    fmt.Printf("  Description: %s\n", user.Description)
    fmt.Printf("  Profile Image: %s\n", user.ProfileImageURL)
    fmt.Printf("  Created: %s\n", user.CreatedAt)
}
```

## Get Live Streams

Check if specific channels are currently live and retrieve stream metadata. This is useful for building stream directories, notifications, or checking stream status before performing stream-specific operations.

**What you get**: Stream title, game/category, viewer count, start time, thumbnail URL, and tags.

**Note**: The API only returns data for streams that are currently live. If a channel is offline, it won't appear in the results.

```go
// Get live streams for specific users
streams, err := client.GetStreams(ctx, &helix.GetStreamsParams{
    UserLogins: []string{"streamer1", "streamer2"},
})
if err != nil {
    log.Fatal(err)
}

if len(streams.Data) == 0 {
    fmt.Println("No streams are live")
} else {
    for _, stream := range streams.Data {
        fmt.Printf("%s is live!\n", stream.UserName)
        fmt.Printf("  Title: %s\n", stream.Title)
        fmt.Printf("  Game: %s\n", stream.GameName)
        fmt.Printf("  Viewers: %d\n", stream.ViewerCount)
    }
}
```

## Get Channel Information

Retrieve channel metadata that persists even when the stream is offline. Unlike stream data which only exists during a live broadcast, channel information is always available.

**What you get**: Current stream title, game/category, language, tags, and content classification labels.

**Use case**: Display channel info on a profile page, or check what game a streamer typically plays.

```go
channels, err := client.GetChannelInformation(ctx, &helix.GetChannelInformationParams{
    BroadcasterIDs: []string{"12345"},
})
if err != nil {
    log.Fatal(err)
}

for _, channel := range channels.Data {
    fmt.Printf("Channel: %s\n", channel.BroadcasterName)
    fmt.Printf("  Title: %s\n", channel.Title)
    fmt.Printf("  Game: %s\n", channel.GameName)
    fmt.Printf("  Language: %s\n", channel.BroadcasterLanguage)
}
```

## Search for Channels

Search for channels by keyword. Results include both live and offline channels, with live channels prioritized. The `IsLive` field indicates current broadcast status.

**Use case**: Build a search feature for finding streamers, or discover channels in specific categories.

```go
results, err := client.SearchChannels(ctx, &helix.SearchChannelsParams{
    Query: "programming",
    First: 10,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d channels:\n", len(results.Data))
for _, channel := range results.Data {
    status := "offline"
    if channel.IsLive {
        status = "LIVE"
    }
    fmt.Printf("  [%s] %s - %s\n", status, channel.DisplayName, channel.Title)
}
```

## Search for Games/Categories

Search for games and categories by name. This returns the game ID which is required for many other API calls (filtering streams by game, getting game analytics, etc.).

**Use case**: Autocomplete for game selection, or find the correct game ID for filtering.

```go
games, err := client.SearchCategories(ctx, &helix.SearchCategoriesParams{
    Query: "minecraft",
    First: 5,
})
if err != nil {
    log.Fatal(err)
}

for _, game := range games.Data {
    fmt.Printf("Game: %s (ID: %s)\n", game.Name, game.ID)
}
```

## Get Top Games

Retrieve the most popular games/categories on Twitch right now, ranked by current viewer count. This data updates in real-time.

**Use case**: Build a "Browse" page showing trending categories, or display popular games for discovery.

```go
topGames, err := client.GetTopGames(ctx, &helix.GetTopGamesParams{
    First: 10,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println("Top 10 Games:")
for i, game := range topGames.Data {
    fmt.Printf("  %d. %s\n", i+1, game.Name)
}
```

## Pagination Example

Most Twitch API endpoints return paginated results. The `First` parameter controls how many results per page (max 100 for most endpoints), and the `Pagination.Cursor` in the response tells you where to continue.

**How it works**:
1. Make initial request with `First` set to your page size
2. Check if `Pagination.Cursor` exists in the response
3. If it exists, make another request with `After` set to the cursor
4. Repeat until no cursor is returned

**Note**: For large datasets (like all followers of a popular channel), consider implementing rate limiting and storing results incrementally to avoid memory issues.

```go
var allFollowers []helix.ChannelFollower
cursor := ""

for {
    resp, err := client.GetChannelFollowers(ctx, &helix.GetChannelFollowersParams{
        BroadcasterID: "12345",
        First:         100,
        After:         cursor,
    })
    if err != nil {
        log.Fatal(err)
    }

    allFollowers = append(allFollowers, resp.Data...)

    // Check if there are more pages
    if resp.Pagination == nil || resp.Pagination.Cursor == "" {
        break
    }
    cursor = resp.Pagination.Cursor
}

fmt.Printf("Total followers: %d\n", len(allFollowers))
```

## Error Handling

The library returns `*helix.APIError` for HTTP errors from the Twitch API. This allows you to inspect the status code and error message to handle different error conditions appropriately.

**Common error codes**:
- `401 Unauthorized`: Token expired or invalid - refresh or re-authenticate
- `404 Not Found`: Resource doesn't exist (but note: `GetUsers` returns empty data, not 404, for non-existent users)
- `429 Too Many Requests`: Rate limited - wait before retrying
- `500+`: Server errors - retry with exponential backoff

**Important**: Some endpoints return empty results instead of errors for "not found" cases. Always check both the error and the result data length.

```go
users, err := client.GetUsers(ctx, &helix.GetUsersParams{
    Logins: []string{"nonexistent_user_12345"},
})
if err != nil {
    if apiErr, ok := err.(*helix.APIError); ok {
        fmt.Printf("API Error: %d - %s\n", apiErr.StatusCode, apiErr.Message)
        switch apiErr.StatusCode {
        case 401:
            fmt.Println("Token expired, need to refresh")
        case 429:
            fmt.Println("Rate limited, slow down requests")
        case 404:
            fmt.Println("Resource not found")
        }
    } else {
        fmt.Printf("Network error: %v\n", err)
    }
    return
}

if len(users.Data) == 0 {
    fmt.Println("User not found")
}
```

