---
layout: default
title: Streams API
description: Get information about active streams and manage stream markers.
---

## GetStreams

Get information about active streams.

**No authentication required**

```go
// Get streams by user IDs
resp, err := client.GetStreams(ctx, &helix.GetStreamsParams{
    UserIDs: []string{"12345", "67890"},
})

// Get streams by user logins
resp, err = client.GetStreams(ctx, &helix.GetStreamsParams{
    UserLogins: []string{"ninja", "shroud"},
})

// Get streams by game ID
resp, err = client.GetStreams(ctx, &helix.GetStreamsParams{
    GameIDs: []string{"509658"},
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

// Filter by language and type
resp, err = client.GetStreams(ctx, &helix.GetStreamsParams{
    Language: "en",
    Type:     "live",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

for _, stream := range resp.Data {
    fmt.Printf("Stream: %s playing %s - %d viewers\n",
        stream.UserName, stream.GameName, stream.ViewerCount)
    fmt.Printf("Title: %s\n", stream.Title)
    fmt.Printf("Started at: %s\n", stream.StartedAt)
}

// Paginate through results
if resp.Pagination.Cursor != "" {
    resp, err = client.GetStreams(ctx, &helix.GetStreamsParams{
        GameIDs: []string{"509658"},
        PaginationParams: &helix.PaginationParams{
            After: resp.Pagination.Cursor,
            First: 20,
        },
    })
}
```

**Parameters:**
- `UserIDs` ([]string, optional): Filter by user IDs (max 100)
- `UserLogins` ([]string, optional): Filter by user login names (max 100)
- `GameIDs` ([]string, optional): Filter by game IDs (max 100)
- `Type` (string, optional): Stream type - "all" or "live" (default: "all")
- `Language` (string, optional): Filter by broadcaster language (ISO 639-1 code)
- Pagination parameters (`First`, `Before`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "41375541868",
      "user_id": "141981764",
      "user_login": "twitchdev",
      "user_name": "TwitchDev",
      "game_id": "509658",
      "game_name": "Just Chatting",
      "type": "live",
      "title": "TwitchDev Monthly Update // May 2025",
      "viewer_count": 1234,
      "started_at": "2025-05-15T18:30:00Z",
      "language": "en",
      "thumbnail_url": "https://static-cdn.jtvnw.net/previews-ttv/live_user_twitchdev-{width}x{height}.jpg",
      "tags": ["English", "API", "Development"],
      "is_mature": false
    },
    {
      "id": "41375542987",
      "user_id": "12345678",
      "user_login": "gamergirl",
      "user_name": "GamerGirl",
      "game_id": "32982",
      "game_name": "Grand Theft Auto V",
      "type": "live",
      "title": "Late night GTA V RP - Come hang out!",
      "viewer_count": 5678,
      "started_at": "2025-05-15T16:00:00Z",
      "language": "en",
      "thumbnail_url": "https://static-cdn.jtvnw.net/previews-ttv/live_user_gamergirl-{width}x{height}.jpg",
      "tags": ["Roleplay", "English"],
      "is_mature": false
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjEwMDQ3MzA2NDo4NjQwNjU3MToxSVZCVDFKMnY5M1BYVDU1MFdRWlorTkpGTU8ifX0"
  }
}
```

**Returns (Stream):**
- `ID` (string): Stream ID
- `UserID` (string): User ID of the broadcaster
- `UserLogin` (string): User login name of the broadcaster
- `UserName` (string): User display name of the broadcaster
- `GameID` (string): Game ID being played
- `GameName` (string): Game name being played
- `Type` (string): Stream type (e.g., "live")
- `Title` (string): Stream title
- `ViewerCount` (int): Number of current viewers
- `StartedAt` (string): UTC timestamp when the stream started
- `Language` (string): Broadcaster language
- `ThumbnailURL` (string): URL template for stream thumbnail
- `Tags` ([]string): Stream tags
- `IsMature` (bool): Whether the stream is marked as mature

## GetFollowedStreams

Get information about active streams for channels that the authenticated user follows.

**Requires:** `user:read:follows`

```go
resp, err := client.GetFollowedStreams(ctx, &helix.GetFollowedStreamsParams{
    UserID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})
if err != nil {
    log.Fatal(err)
}

for _, stream := range resp.Data {
    fmt.Printf("%s is live! Playing %s with %d viewers\n",
        stream.UserName, stream.GameName, stream.ViewerCount)
}

// Paginate through more followed streams
if resp.Pagination.Cursor != "" {
    resp, err = client.GetFollowedStreams(ctx, &helix.GetFollowedStreamsParams{
        UserID: "12345",
        PaginationParams: &helix.PaginationParams{
            After: resp.Pagination.Cursor,
            First: 100,
        },
    })
}
```

**Parameters:**
- `UserID` (string): The ID of the user whose followed streams you want to get
- Pagination parameters (`First`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "42597123456",
      "user_id": "98765432",
      "user_login": "streamerninja",
      "user_name": "StreamerNinja",
      "game_id": "511224",
      "game_name": "Apex Legends",
      "type": "live",
      "title": "Ranked Grind to Predator | !socials",
      "viewer_count": 15432,
      "started_at": "2025-05-15T20:15:00Z",
      "language": "en",
      "thumbnail_url": "https://static-cdn.jtvnw.net/previews-ttv/live_user_streamerninja-{width}x{height}.jpg",
      "tags": ["English", "FPS", "Competitive"],
      "is_mature": false
    },
    {
      "id": "42597123457",
      "user_id": "11223344",
      "user_login": "cozyartist",
      "user_name": "CozyArtist",
      "game_id": "509660",
      "game_name": "Art",
      "type": "live",
      "title": "Chill painting stream - Bob Ross vibes",
      "viewer_count": 856,
      "started_at": "2025-05-15T19:00:00Z",
      "language": "en",
      "thumbnail_url": "https://static-cdn.jtvnw.net/previews-ttv/live_user_cozyartist-{width}x{height}.jpg",
      "tags": ["Art", "Chill", "English"],
      "is_mature": false
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjI1NDUzNDkyNDE6ODc2NTQzMjE6MVBWQlQxSjJ2OTNQWFQzOTBXUVl6K05KRk1PIjEyfX0"
  }
}
```

**Returns:** Same Stream objects as GetStreams

## GetStreamKey

Get the channel stream key for a broadcaster.

**Requires:** `channel:read:stream_key`

```go
resp, err := client.GetStreamKey(ctx, "12345")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Stream key: %s\n", resp.Data.StreamKey)
```

**Parameters:**
- `broadcasterID` (string): The ID of the broadcaster

**Sample Response:**
```json
{
  "data": [
    {
      "stream_key": "live_44322889_a34ub37c8ajv98a0"
    }
  ]
}
```

**Returns (StreamKey):**
- `StreamKey` (string): The channel's stream key

## CreateStreamMarker

Create a marker in the stream at the current timestamp.

**Requires:** `channel:manage:broadcast`

```go
// Create marker without description
resp, err := client.CreateStreamMarker(ctx, &helix.CreateStreamMarkerParams{
    UserID: "12345",
})

// Create marker with description
resp, err = client.CreateStreamMarker(ctx, &helix.CreateStreamMarkerParams{
    UserID:      "12345",
    Description: "Epic boss fight begins",
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Marker created at position %d seconds\n", resp.Data.PositionSeconds)
fmt.Printf("Marker ID: %s\n", resp.Data.ID)
```

**Parameters:**
- `UserID` (string): The ID of the broadcaster
- `Description` (string, optional): Short description of the marker (max 140 characters)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "123456789",
      "created_at": "2025-05-15T21:45:32Z",
      "description": "Epic boss fight begins",
      "position_seconds": 5435
    }
  ]
}
```

**Returns (StreamMarker):**
- `ID` (string): Marker ID
- `CreatedAt` (string): UTC timestamp when the marker was created
- `Description` (string): Marker description
- `PositionSeconds` (int): Position in the stream when the marker was created (in seconds)

## GetStreamMarkers

Get a list of markers for a user's stream or a VOD.

**Requires:** `user:read:broadcast`

```go
// Get markers by user ID
resp, err := client.GetStreamMarkers(ctx, &helix.GetStreamMarkersParams{
    UserID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

// Get markers by video ID
resp, err = client.GetStreamMarkers(ctx, &helix.GetStreamMarkersParams{
    VideoID: "67890",
})

for _, video := range resp.Data {
    fmt.Printf("Video ID: %s\n", video.VideoID)
    for _, marker := range video.Markers {
        fmt.Printf("  Marker at %ds: %s (ID: %s)\n",
            marker.PositionSeconds, marker.Description, marker.ID)
    }
}

// Paginate through results
if resp.Pagination.Cursor != "" {
    resp, err = client.GetStreamMarkers(ctx, &helix.GetStreamMarkersParams{
        UserID: "12345",
        PaginationParams: &helix.PaginationParams{
            After: resp.Pagination.Cursor,
            First: 20,
        },
    })
}
```

**Parameters:**
- `UserID` (string, optional): User ID of the broadcaster (mutually exclusive with VideoID)
- `VideoID` (string, optional): VOD/video ID (mutually exclusive with UserID)
- Pagination parameters (`First`, `Before`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "141981764",
      "user_login": "twitchdev",
      "user_name": "TwitchDev",
      "videos": [
        {
          "video_id": "335921245",
          "markers": [
            {
              "id": "106b8d6243a4f883d25ad75e6cdffdc4",
              "created_at": "2025-05-15T18:45:30Z",
              "description": "Opening segment",
              "position_seconds": 120
            },
            {
              "id": "206b8d6243a4f883d25ad75e6cdffdc5",
              "created_at": "2025-05-15T19:30:15Z",
              "description": "Q&A session starts",
              "position_seconds": 2835
            },
            {
              "id": "306b8d6243a4f883d25ad75e6cdffdc6",
              "created_at": "2025-05-15T20:45:00Z",
              "description": "Demo time",
              "position_seconds": 7320
            }
          ]
        },
        {
          "video_id": "335921246",
          "markers": [
            {
              "id": "406b8d6243a4f883d25ad75e6cdffdc7",
              "created_at": "2025-05-14T16:20:45Z",
              "description": "Highlight moment",
              "position_seconds": 450
            }
          ]
        }
      ]
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjEwMDQ3MzA2NDo4NjQwNjU3MToxSVZCVDFKMnY5M1BYVDU1MFdRWlorTkpGTU8ifX0"
  }
}
```

**Returns (VideoStreamMarkers):**
- `VideoID` (string): The ID of the video
- `Markers` ([]StreamMarker): Array of stream markers for the video, each containing:
  - `ID` (string): Marker ID
  - `CreatedAt` (string): UTC timestamp when the marker was created
  - `Description` (string): Marker description
- `PositionSeconds` (int): Position in the stream (in seconds)

