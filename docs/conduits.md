---
layout: default
title: EventSub Conduits API
description: Manage EventSub conduits for receiving event notifications. Conduits allow you to manage multiple shards for high-availability event subscriptions.
---

## GetConduits

Get a list of conduits for the application.

**Requires:** App access token

```go
resp, err := client.GetConduits(ctx)
if err != nil {
    log.Fatal(err)
}
for _, conduit := range resp.Data {
    fmt.Printf("Conduit ID: %s, Shard Count: %d\n", conduit.ID, conduit.ShardCount)
}
```

**Returns:**
- `ID` (string): The conduit's ID
- `ShardCount` (int): The number of shards associated with this conduit

**Sample Response:**
```json
{
  "data": [
    {
      "id": "bfcfc791-4b4c-47f4-8f5a-7b8d0e3f1c9d",
      "shard_count": 5
    },
    {
      "id": "26b1c993-bfcf-44d9-b876-379dacafe75a",
      "shard_count": 3
    }
  ]
}
```

## CreateConduit

Create a new conduit for the application.

**Requires:** App access token

```go
resp, err := client.CreateConduit(ctx, &helix.CreateConduitParams{
    ShardCount: 5,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created conduit ID: %s with %d shards\n",
    resp.Data[0].ID, resp.Data[0].ShardCount)
```

**Parameters:**
- `ShardCount` (int): The number of shards to create for this conduit

**Returns:**
- `ID` (string): The newly created conduit's ID
- `ShardCount` (int): The number of shards associated with this conduit

**Sample Response:**
```json
{
  "data": [
    {
      "id": "bfcfc791-4b4c-47f4-8f5a-7b8d0e3f1c9d",
      "shard_count": 5
    }
  ]
}
```

## UpdateConduit

Update an existing conduit's shard count.

**Requires:** App access token

```go
resp, err := client.UpdateConduit(ctx, &helix.UpdateConduitParams{
    ID:         "existing-conduit-id",
    ShardCount: 10,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Updated conduit ID: %s, new shard count: %d\n",
    resp.Data[0].ID, resp.Data[0].ShardCount)
```

**Parameters:**
- `ID` (string): The ID of the conduit to update
- `ShardCount` (int): The new number of shards for this conduit

**Returns:**
- `ID` (string): The conduit's ID
- `ShardCount` (int): The updated number of shards

**Sample Response:**
```json
{
  "data": [
    {
      "id": "bfcfc791-4b4c-47f4-8f5a-7b8d0e3f1c9d",
      "shard_count": 10
    }
  ]
}
```

## DeleteConduit

Delete a conduit and all its shards.

**Requires:** App access token

```go
err := client.DeleteConduit(ctx, &helix.DeleteConduitParams{
    ConduitID: "conduit-to-delete",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Conduit deleted successfully")
```

**Parameters:**
- `ConduitID` (string): The ID of the conduit to delete

**Sample Response:**
```json
{}
```

## GetConduitShards

Get the shards for a conduit.

**Requires:** App access token

```go
// Get all shards for a conduit
resp, err := client.GetConduitShards(ctx, &helix.GetConduitShardsParams{
    ConduitID: "your-conduit-id",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})
if err != nil {
    log.Fatal(err)
}

// Filter by shard status
resp, err = client.GetConduitShards(ctx, &helix.GetConduitShardsParams{
    ConduitID: "your-conduit-id",
    Status:    "enabled",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

for _, shard := range resp.Data {
    fmt.Printf("Shard ID: %s, Status: %s\n", shard.ID, shard.Status)
    fmt.Printf("  Transport: %s\n", shard.Transport.Method)
}
```

**Parameters:**
- `ConduitID` (string): The ID of the conduit
- `Status` (string, optional): Filter shards by status (e.g., `enabled`, `disabled`, `webhook_callback_verification_pending`, `webhook_callback_verification_failed`, `notification_failures_exceeded`, `websocket_disconnected`, `websocket_failed_ping_pong`, `websocket_received_inbound_traffic`, `websocket_connection_unused`, `websocket_internal_error`, `websocket_network_timeout`, `websocket_network_error`)
- `PaginationParams` (optional): Pagination parameters

**Returns:**
- `ID` (string): The shard's ID
- `Status` (string): The shard's status
- `Transport` (object): Transport configuration
  - `Method` (string): The transport method (e.g., `webhook`, `websocket`)
  - `Callback` (string): The callback URL (for webhooks)
  - `SessionID` (string): The session ID (for websockets)
  - `ConnectedAt` (string): When the websocket connection was established
  - `DisconnectedAt` (string): When the websocket was disconnected

**Sample Response:**
```json
{
  "data": [
    {
      "id": "0",
      "status": "enabled",
      "transport": {
        "method": "webhook",
        "callback": "https://example.com/webhooks/callback"
      }
    },
    {
      "id": "1",
      "status": "websocket_disconnected",
      "transport": {
        "method": "websocket",
        "session_id": "AgoQHR3s6Mb4T8GFB1l3DlPfiRIGY2VsbC1h",
        "connected_at": "2023-07-15T18:42:30.073Z",
        "disconnected_at": "2023-07-15T20:15:45.321Z"
      }
    },
    {
      "id": "2",
      "status": "enabled",
      "transport": {
        "method": "websocket",
        "session_id": "AgoQO7YkKJaS1rZQhvBDP6K8xQIGY2VsbC1h",
        "connected_at": "2023-07-15T19:30:15.123Z"
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjoiMiJ9"
  }
}
```

## UpdateConduitShards

Update the shards for a conduit.

**Requires:** App access token

```go
resp, err := client.UpdateConduitShards(ctx, &helix.UpdateConduitShardsParams{
    ConduitID: "your-conduit-id",
    Shards: []helix.ConduitShardUpdate{
        {
            ID: "0",
            Transport: helix.ConduitShardTransport{
                Method:   "webhook",
                Callback: "https://example.com/webhook",
                Secret:   "your-webhook-secret",
            },
        },
        {
            ID: "1",
            Transport: helix.ConduitShardTransport{
                Method:    "websocket",
                SessionID: "your-session-id",
            },
        },
    },
})
if err != nil {
    log.Fatal(err)
}
for _, shard := range resp.Data {
    fmt.Printf("Updated shard ID: %s, Status: %s\n", shard.ID, shard.Status)
}
```

**Parameters:**
- `ConduitID` (string): The ID of the conduit
- `Shards` (array): Array of shard updates
  - `ID` (string): The shard ID
  - `Transport` (object): Transport configuration
    - `Method` (string): Transport method (`webhook` or `websocket`)
    - `Callback` (string): Webhook callback URL (required for webhook method)
    - `Secret` (string): Webhook secret (required for webhook method)
    - `SessionID` (string): WebSocket session ID (required for websocket method)

**Returns:**
- Array of updated shards with their IDs, statuses, and transport configurations
- `Errors` (array, optional): Array of errors for any shards that failed to update
  - `ID` (string): The shard ID that failed
  - `Message` (string): Error message
  - `Code` (string): Error code

**Sample Response:**
```json
{
  "data": [
    {
      "id": "0",
      "status": "enabled",
      "transport": {
        "method": "webhook",
        "callback": "https://example.com/webhook"
      }
    },
    {
      "id": "1",
      "status": "enabled",
      "transport": {
        "method": "websocket",
        "session_id": "AgoQHR3s6Mb4T8GFB1l3DlPfiRIGY2VsbC1h"
      }
    }
  ],
  "errors": [
    {
      "id": "2",
      "message": "Invalid session_id",
      "code": "invalid_parameter"
    }
  ]
}
```

