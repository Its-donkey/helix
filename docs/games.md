---
layout: default
title: Games API
description: Retrieve information about games and categories on Twitch.
---

## GetGames

Get information about one or more games by ID, name, or IGDB ID.

**Requires:** No authentication required

```go
// Get games by ID
resp, err := client.GetGames(ctx, &helix.GetGamesParams{
    IDs: []string{"509658", "32982"},
})

// Get games by name
resp, err = client.GetGames(ctx, &helix.GetGamesParams{
    Names: []string{"Just Chatting", "League of Legends"},
})

// Get games by IGDB ID
resp, err = client.GetGames(ctx, &helix.GetGamesParams{
    IGDBIDs: []string{"12345", "67890"},
})

for _, game := range resp.Data {
    fmt.Printf("Game: %s (ID: %s, IGDB ID: %s)\nBox Art: %s\n",
        game.Name, game.ID, game.IGDBId, game.BoxArtURL)
}
```

**Parameters:**
- `IDs` ([]string, optional): Game IDs (max 100)
- `Names` ([]string, optional): Game names (max 100)
- `IGDBIDs` ([]string, optional): IGDB IDs (max 100)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "509658",
      "name": "Just Chatting",
      "box_art_url": "https://static-cdn.jtvnw.net/ttv-boxart/509658-{width}x{height}.jpg",
      "igdb_id": ""
    },
    {
      "id": "32982",
      "name": "Grand Theft Auto V",
      "box_art_url": "https://static-cdn.jtvnw.net/ttv-boxart/32982_IGDB-{width}x{height}.jpg",
      "igdb_id": "1020"
    }
  ]
}
```

## GetTopGames

Get the top games being streamed on Twitch, sorted by number of current viewers.

**Requires:** No authentication required

```go
resp, err := client.GetTopGames(ctx, &helix.PaginationParams{
    First: 20,
})

for _, game := range resp.Data {
    fmt.Printf("Game: %s (ID: %s)\nBox Art: %s\n",
        game.Name, game.ID, game.BoxArtURL)
}

// Paginate through results
if resp.Pagination.Cursor != "" {
    resp, err = client.GetTopGames(ctx, &helix.PaginationParams{
        First: 20,
        After: resp.Pagination.Cursor,
    })
}
```

**Parameters:**
- Pagination parameters (`First`, `Before`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "516575",
      "name": "VALORANT",
      "box_art_url": "https://static-cdn.jtvnw.net/ttv-boxart/516575-{width}x{height}.jpg",
      "igdb_id": "128591"
    },
    {
      "id": "509658",
      "name": "Just Chatting",
      "box_art_url": "https://static-cdn.jtvnw.net/ttv-boxart/509658-{width}x{height}.jpg",
      "igdb_id": ""
    },
    {
      "id": "21779",
      "name": "League of Legends",
      "box_art_url": "https://static-cdn.jtvnw.net/ttv-boxart/21779-{width}x{height}.jpg",
      "igdb_id": "115"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6MjB9fQ"
  }
}
```

