# Clips API

Create and retrieve Twitch clips from broadcasts and VODs.

## CreateClip

Create a clip from the broadcaster's stream.

**Requires:** `clips:edit` scope

```go
resp, err := client.CreateClip(ctx, &helix.CreateClipParams{
    BroadcasterID: "12345",
    HasDelay:      false, // Set to true if the broadcaster has a delay
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Clip created! ID: %s, Edit URL: %s\n",
    resp.Data.ID, resp.Data.EditURL)
```

**Parameters:**
- `BroadcasterID` (string): The ID of the broadcaster whose stream you want to clip
- `HasDelay` (bool): If true, the clip creation is delayed by a few seconds to account for the broadcaster's delay setting

**Returns:**
- `ID` (string): The clip ID
- `EditURL` (string): URL where the clip can be edited

**Sample Response:**
```json
{
  "data": [
    {
      "id": "AwkwardHelplessSalamanderSwiftRage",
      "edit_url": "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage/edit"
    }
  ]
}
```

## CreateClipFromVOD

Create a clip from an existing VOD (Video on Demand).

**Requires:** `editor:manage:clips` or `channel:manage:clips` scope

```go
// Basic usage - create a 30-second clip (default duration)
resp, err := client.CreateClipFromVOD(ctx, &helix.CreateClipFromVODParams{
    EditorID:      "11111",      // User creating the clip
    BroadcasterID: "22222",      // Channel owner
    VODID:         "1234567890", // VOD to clip from
    VODOffset:     3600,         // 1 hour into the VOD (where clip ends)
    Title:         "Epic Play!",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Clip created! ID: %s\n", resp.ID)
fmt.Printf("Edit URL: %s\n", resp.EditURL)

// With custom duration (5-60 seconds)
duration := 45.0
resp, err = client.CreateClipFromVOD(ctx, &helix.CreateClipFromVODParams{
    EditorID:      "11111",
    BroadcasterID: "22222",
    VODID:         "1234567890",
    VODOffset:     7200,         // 2 hours into the VOD
    Title:         "Amazing Moment",
    Duration:      &duration,    // 45 second clip
})
```

**Parameters:**
- `EditorID` (string, required): User ID of the editor creating the clip
- `BroadcasterID` (string, required): User ID of the channel receiving the clip
- `VODID` (string, required): ID of the VOD to create clip from
- `VODOffset` (int, required): Offset in seconds where the clip ends in the VOD
- `Title` (string, required): Title for the clip
- `Duration` (*float64, optional): Clip length in seconds (5-60, default 30, precision 0.1)

**Returns:**
- `ID` (string): The newly created clip ID
- `EditURL` (string): URL where the clip can be edited

**Sample Response:**
```json
{
  "data": [
    {
      "id": "VODClipAwesome123",
      "edit_url": "https://clips.twitch.tv/VODClipAwesome123/edit"
    }
  ]
}
```

**Example Output:**
```
Clip created! ID: VODClipAwesome123
Edit URL: https://clips.twitch.tv/VODClipAwesome123/edit
```

**Notes:**
- The `VODOffset` specifies where the clip **ends**, not where it starts
- The clip will include content from `(VODOffset - Duration)` to `VODOffset`
- Duration defaults to 30 seconds if not specified
- Duration must be between 5 and 60 seconds

## GetClips

Get one or more clips by broadcaster, game, or clip IDs.

**No authentication required**

```go
// Get clips by broadcaster ID
resp, err := client.GetClips(ctx, &helix.GetClipsParams{
    BroadcasterID: "12345",
    First:         20,
})

// Get clips by game ID
resp, err = client.GetClips(ctx, &helix.GetClipsParams{
    GameID: "67890",
    First:  10,
})

// Get specific clips by IDs (max 100)
resp, err = client.GetClips(ctx, &helix.GetClipsParams{
    IDs: []string{"clip1", "clip2", "clip3"},
})

// Get clips with date range
resp, err = client.GetClips(ctx, &helix.GetClipsParams{
    BroadcasterID: "12345",
    StartedAt:     "2023-01-01T00:00:00Z",
    EndedAt:       "2023-12-31T23:59:59Z",
    First:         50,
})

// Get featured clips
resp, err = client.GetClips(ctx, &helix.GetClipsParams{
    BroadcasterID: "12345",
    IsFeatured:    true,
})

for _, clip := range resp.Data {
    fmt.Printf("Clip: %s by %s - %d views\n",
        clip.Title, clip.CreatorName, clip.ViewCount)
}

// Paginate through results
if resp.Pagination.Cursor != "" {
    resp, err = client.GetClips(ctx, &helix.GetClipsParams{
        BroadcasterID: "12345",
        After:         resp.Pagination.Cursor,
        First:         20,
    })
}
```

**Parameters:**
- `BroadcasterID` (string, optional): Filter by broadcaster ID
- `GameID` (string, optional): Filter by game ID
- `IDs` ([]string, optional): Get specific clips by ID (max 100)
- `StartedAt` (string, optional): Starting date/time for returned clips (RFC3339 format)
- `EndedAt` (string, optional): Ending date/time for returned clips (RFC3339 format)
- `IsFeatured` (bool, optional): Filter for featured clips only
- Pagination parameters (`First`, `Before`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "AwkwardHelplessSalamanderSwiftRage",
      "url": "https://clips.twitch.tv/AwkwardHelplessSalamanderSwiftRage",
      "embed_url": "https://clips.twitch.tv/embed?clip=AwkwardHelplessSalamanderSwiftRage",
      "broadcaster_id": "67955580",
      "broadcaster_name": "ChewieMelodies",
      "creator_id": "53834192",
      "creator_name": "BlackNova03",
      "video_id": "205586603",
      "game_id": "488191",
      "language": "en",
      "title": "babymetal",
      "view_count": 10,
      "created_at": "2017-11-30T22:34:18Z",
      "thumbnail_url": "https://clips-media-assets.twitch.tv/157589949-preview-480x272.jpg",
      "duration": 60,
      "vod_offset": 480,
      "is_featured": false
    },
    {
      "id": "TameIntelligentCarabeefBudStar",
      "url": "https://clips.twitch.tv/TameIntelligentCarabeefBudStar",
      "embed_url": "https://clips.twitch.tv/embed?clip=TameIntelligentCarabeefBudStar",
      "broadcaster_id": "67955580",
      "broadcaster_name": "ChewieMelodies",
      "creator_id": "53834192",
      "creator_name": "BlackNova03",
      "video_id": "205586603",
      "game_id": "488191",
      "language": "en",
      "title": "babymetal",
      "view_count": 10,
      "created_at": "2017-11-30T22:34:18Z",
      "thumbnail_url": "https://clips-media-assets.twitch.tv/157589949-preview-480x272.jpg",
      "duration": 60,
      "vod_offset": null,
      "is_featured": true
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjoiIn0"
  }
}
```

## GetClipsDownload

Get download URLs for clips. The URLs are temporary and expire after a short time.

**Requires:** `clips:edit` scope for clips the user created or for the broadcaster's clips

```go
resp, err := client.GetClipsDownload(ctx, &helix.GetClipsDownloadParams{
    ClipIDs: []string{"clip1", "clip2", "clip3"},
})
if err != nil {
    log.Fatal(err)
}
for _, clip := range resp.Data {
    fmt.Printf("Clip ID: %s, Download URL: %s (expires: %s)\n",
        clip.ID, clip.URL, clip.ExpiresAt)
}
```

**Parameters:**
- `ClipIDs` ([]string): Array of clip IDs to get download URLs for

**Returns:**
- `ID` (string): The clip ID
- `URL` (string): Temporary download URL for the clip
- `ExpiresAt` (string): When the download URL expires

**Sample Response:**
```json
{
  "data": [
    {
      "id": "AwkwardHelplessSalamanderSwiftRage",
      "url": "https://production.assets.clips.twitchcdn.net/vod-123456789-offset-123.mp4",
      "expires_at": "2023-12-15T10:30:00Z"
    },
    {
      "id": "TameIntelligentCarabeefBudStar",
      "url": "https://production.assets.clips.twitchcdn.net/vod-987654321-offset-456.mp4",
      "expires_at": "2023-12-15T10:30:00Z"
    }
  ]
}
```
