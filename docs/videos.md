---
layout: default
title: Videos API
description: Retrieve and manage Twitch videos (VODs, highlights, and uploads).
---

## GetVideos

Get one or more videos by video ID, user ID, or game ID.

**Requires:** No authentication required

```go
// Get specific videos by IDs (max 100)
resp, err := client.GetVideos(ctx, &helix.GetVideosParams{
    IDs: []string{"video1", "video2", "video3"},
})

// Get videos by user ID
resp, err = client.GetVideos(ctx, &helix.GetVideosParams{
    UserID: "12345",
    First:  20,
})

// Get videos by game ID
resp, err = client.GetVideos(ctx, &helix.GetVideosParams{
    GameID: "67890",
    First:  10,
})

// Get videos with filters
resp, err = client.GetVideos(ctx, &helix.GetVideosParams{
    UserID:   "12345",
    Language: "en",
    Period:   "week",      // all/day/week/month
    Sort:     "views",     // time/trending/views
    Type:     "archive",   // all/archive/highlight/upload
    First:    50,
})

for _, video := range resp.Data {
    fmt.Printf("Video: %s by %s - %d views (%s)\n",
        video.Title, video.UserName, video.ViewCount, video.Duration)
}

// Paginate through results
if resp.Pagination.Cursor != "" {
    resp, err = client.GetVideos(ctx, &helix.GetVideosParams{
        UserID: "12345",
        After:  resp.Pagination.Cursor,
        First:  20,
    })
}
```

**Parameters:**
- `IDs` ([]string, optional): Get specific videos by ID (max 100)
- `UserID` (string, optional): Filter by user ID
- `GameID` (string, optional): Filter by game ID
- `Language` (string, optional): Filter by language (ISO 639-1 two-letter code)
- `Period` (string, optional): Period during which the video was created: `all`, `day`, `week`, `month`
- `Sort` (string, optional): Sort order: `time`, `trending`, `views`
- `Type` (string, optional): Video type: `all`, `archive`, `highlight`, `upload`
- Pagination parameters (`First`, `Before`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "335921245",
      "stream_id": "41375541868",
      "user_id": "141981764",
      "user_login": "twitchdev",
      "user_name": "TwitchDev",
      "title": "Twitch Developers 101",
      "description": "Welcome to Twitch development!",
      "created_at": "2018-11-14T21:30:18Z",
      "published_at": "2018-11-14T22:04:30Z",
      "url": "https://www.twitch.tv/videos/335921245",
      "thumbnail_url": "https://static-cdn.jtvnw.net/cf_vods/d2nvs31859zcd8/twitchdev/335921245/ce0f3a7f-57a3-4152-bc06-0c6610189fb3/thumb/index-0000000000-%{width}x%{height}.jpg",
      "viewable": "public",
      "view_count": 1863062,
      "language": "en",
      "type": "upload",
      "duration": "3h8m33s",
      "muted_segments": [
        {
          "duration": 30,
          "offset": 120
        }
      ]
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjoiMTUwMzQ0MTc3NjQyNDQyMjAwMCJ9"
  }
}
```

## DeleteVideos

Delete one or more videos. Videos are deleted asynchronously.

**Requires:** `channel:manage:videos`

```go
resp, err := client.DeleteVideos(ctx, []string{"video1", "video2", "video3"})
if err != nil {
    log.Fatal(err)
}
for _, videoID := range resp.Data {
    fmt.Printf("Video deleted: %s\n", videoID)
}
```

**Parameters:**
- `videoIDs` ([]string): Array of video IDs to delete

**Returns:**
- Array of deleted video IDs

