---
layout: default
title: Search API
description: Search for categories (games) and channels on Twitch.
---

## SearchCategories

Search for categories (games) on Twitch by name.

**Requires:** No authentication required

```go
resp, err := client.SearchCategories(ctx, &helix.SearchCategoriesParams{
    Query: "League of Legends",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

for _, category := range resp.Data {
    fmt.Printf("Category: %s (ID: %s)\nBox Art: %s\n",
        category.Name, category.ID, category.BoxArtURL)
}

// Paginate through results
if resp.Pagination.Cursor != "" {
    resp, err = client.SearchCategories(ctx, &helix.SearchCategoriesParams{
        Query: "League of Legends",
        PaginationParams: &helix.PaginationParams{
            First: 20,
            After: resp.Pagination.Cursor,
        },
    })
}
```

**Parameters:**
- `Query` (string, required): Search query for categories
- Pagination parameters (`First`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "21779",
      "name": "League of Legends",
      "box_art_url": "https://static-cdn.jtvnw.net/ttv-boxart/21779-{width}x{height}.jpg"
    },
    {
      "id": "512804",
      "name": "League of Legends: Wild Rift",
      "box_art_url": "https://static-cdn.jtvnw.net/ttv-boxart/512804-{width}x{height}.jpg"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6MjB9fQ"
  }
}
```

## SearchChannels

Search for channels on Twitch by broadcaster name, login, or display name.

**Requires:** No authentication required

```go
// Search for channels
resp, err := client.SearchChannels(ctx, &helix.SearchChannelsParams{
    Query: "shroud",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

// Search for live channels only
resp, err = client.SearchChannels(ctx, &helix.SearchChannelsParams{
    Query:    "gaming",
    LiveOnly: true,
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

for _, channel := range resp.Data {
    fmt.Printf("Channel: %s (%s)\n", channel.DisplayName, channel.BroadcasterLogin)
    fmt.Printf("Game: %s (ID: %s)\n", channel.GameName, channel.GameID)
    fmt.Printf("Title: %s\n", channel.Title)
    fmt.Printf("Language: %s\n", channel.BroadcasterLanguage)
    fmt.Printf("Live: %v\n", channel.IsLive)
    if channel.IsLive {
        fmt.Printf("Started at: %s\n", channel.StartedAt)
    }
    fmt.Printf("Tags: %v\n", channel.Tags)
    fmt.Printf("Thumbnail: %s\n", channel.ThumbnailURL)
}

// Paginate through results
if resp.Pagination.Cursor != "" {
    resp, err = client.SearchChannels(ctx, &helix.SearchChannelsParams{
        Query: "shroud",
        PaginationParams: &helix.PaginationParams{
            First: 20,
            After: resp.Pagination.Cursor,
        },
    })
}
```

**Parameters:**
- `Query` (string, required): Search query for channels
- `LiveOnly` (bool, optional): Filter to only return live channels
- Pagination parameters (`First`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_language": "en",
      "broadcaster_login": "shroud",
      "display_name": "shroud",
      "game_id": "516575",
      "game_name": "VALORANT",
      "id": "37402112",
      "is_live": true,
      "tags": ["English", "Competitive", "FPS"],
      "thumbnail_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/7ed5e0c1-93f9-4f67-8c7d-aa2c0a1dd69e-profile_image-300x300.png",
      "title": "Ranked with the squad | Follow @shroud on socials",
      "started_at": "2025-12-11T15:30:00Z"
    },
    {
      "broadcaster_language": "en",
      "broadcaster_login": "shroudette",
      "display_name": "Shroudette",
      "game_id": "33214",
      "game_name": "Fortnite",
      "id": "98765432",
      "is_live": false,
      "tags": ["English", "Chill"],
      "thumbnail_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/shroudette-profile_image-300x300.png",
      "title": "Casual Fortnite vibes"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6MjB9fQ"
  }
}
```

