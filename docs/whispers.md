---
layout: default
title: Whispers API
description: Send private whisper messages between Twitch users.
---

## SendWhisper

Send a whisper message from one user to another.

**Requires:** `user:manage:whispers`

```go
err := client.SendWhisper(ctx, &helix.SendWhisperParams{
    FromUserID: "12345",
    ToUserID:   "67890",
    Message:    "Hey! Thanks for watching the stream!",
})

if err != nil {
    fmt.Printf("Failed to send whisper: %v\n", err)
}
```

**Parameters:**
- `FromUserID` (string, required): The ID of the user sending the whisper
- `ToUserID` (string, required): The ID of the user to receive the whisper
- `Message` (string, required): The whisper message to send (max 500 characters)

**Sample Response:**
Returns `204 No Content` on success with an empty response body.

