---
layout: default
title: Raids API
description: Manage raids between Twitch channels.
---

## StartRaid

Start a raid from one broadcaster's channel to another.

**Requires:** `channel:manage:raids`

```go
resp, err := client.StartRaid(ctx, &helix.StartRaidParams{
    FromBroadcasterID: "12345",
    ToBroadcasterID:   "67890",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Raid started at %s, Is Mature: %v\n",
    resp.Data.CreatedAt, resp.Data.IsMature)
```

**Parameters:**
- `FromBroadcasterID` - Required. ID of the broadcaster starting the raid
- `ToBroadcasterID` - Required. ID of the broadcaster being raided

**Response Fields:**
- `CreatedAt` - Timestamp when the raid was created
- `IsMature` - Whether the target channel has mature content

**Sample Response:**
```json
{
  "data": [
    {
      "created_at": "2024-03-15T14:30:45Z",
      "is_mature": false
    }
  ]
}
```

## CancelRaid

Cancel a pending raid before it executes.

**Requires:** `channel:manage:raids`

```go
err := client.CancelRaid(ctx, "12345")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Raid cancelled successfully")
```

**Parameters:**
- `broadcasterID` (string): Required. ID of the broadcaster canceling the raid

**Sample Response:**
```
No response body. Returns 204 No Content on success.
```

