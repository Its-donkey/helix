---
layout: default
title: Ingest Servers API
description: Get information about Twitch ingest servers for streaming.
---

> **Note:** This endpoint uses a different base URL (`ingest.twitch.tv`) than the Helix API.

## GetIngestServers

Get the list of available Twitch ingest servers for RTMP streaming.

**No authentication required**

```go
resp, err := client.GetIngestServers(ctx)
if err != nil {
    log.Fatal(err)
}
for _, ingest := range resp.Ingests {
    fmt.Printf("Server: %s (ID: %d)\n", ingest.Name, ingest.ID)
    fmt.Printf("  Availability: %.2f, Default: %v, Priority: %d\n",
        ingest.Availability, ingest.Default, ingest.Priority)
    fmt.Printf("  URL Template: %s\n", ingest.URLTemplate)
}
```

**Returns:**

`IngestServersResponse` containing:
- `Ingests` ([]IngestServer): Array of ingest server objects

Each `IngestServer` contains:
- `ID` (int): Unique identifier for the ingest server
- `Availability` (float64): Server availability metric
- `Default` (bool): Whether this is the default server
- `Name` (string): Human-readable name of the server location
- `URLTemplate` (string): RTMP URL template with `{stream_key}` placeholder
- `Priority` (int): Server priority value

**Sample Response:**
```json
{
  "ingests": [
    {
      "_id": 24,
      "availability": 1.0,
      "default": false,
      "name": "US West: San Francisco, CA",
      "url_template": "rtmp://live-sjc.twitch.tv/app/{stream_key}",
      "priority": 20
    },
    {
      "_id": 45,
      "availability": 1.0,
      "default": true,
      "name": "US East: Ashburn, VA",
      "url_template": "rtmp://live-iad05.contribute.live-video.net/app/{stream_key}",
      "priority": 10
    },
    {
      "_id": 62,
      "availability": 0.98,
      "default": false,
      "name": "EU Central: Frankfurt, Germany",
      "url_template": "rtmp://live-fra02.contribute.live-video.net/app/{stream_key}",
      "priority": 30
    },
    {
      "_id": 71,
      "availability": 0.99,
      "default": false,
      "name": "Asia Pacific: Tokyo, Japan",
      "url_template": "rtmp://live-tyo01.contribute.live-video.net/app/{stream_key}",
      "priority": 40
    },
    {
      "_id": 83,
      "availability": 1.0,
      "default": false,
      "name": "South America: Sao Paulo, Brazil",
      "url_template": "rtmp://live-gru03.contribute.live-video.net/app/{stream_key}",
      "priority": 50
    }
  ]
}
```

## Helper Methods

### GetIngestServerByName

Find a specific ingest server by its name.

```go
resp, err := client.GetIngestServers(ctx)
if err != nil {
    log.Fatal(err)
}

server := resp.GetIngestServerByName("US West: San Francisco, CA")
if server != nil {
    fmt.Printf("Found server: %s\n", server.Name)
    fmt.Printf("URL Template: %s\n", server.URLTemplate)
} else {
    fmt.Println("Server not found")
}
```

**Parameters:**
- `name` (string): The exact name of the ingest server to find

**Returns:**
- `*IngestServer`: Pointer to the matching ingest server, or `nil` if not found

### GetRTMPURL

Get the complete RTMP URL with a stream key inserted.

```go
resp, err := client.GetIngestServers(ctx)
if err != nil {
    log.Fatal(err)
}

// Get the first available server
if len(resp.Ingests) > 0 {
    server := resp.Ingests[0]
    streamKey := "your_stream_key_here"
    rtmpURL := server.GetRTMPURL(streamKey)
    fmt.Printf("RTMP URL: %s\n", rtmpURL)
    // Example output: rtmp://live-sjc.twitch.tv/app/your_stream_key_here
}

// Or use a specific server by name
server := resp.GetIngestServerByName("US West: San Francisco, CA")
if server != nil {
    rtmpURL := server.GetRTMPURL("your_stream_key_here")
    fmt.Printf("RTMP URL: %s\n", rtmpURL)
}
```

**Parameters:**
- `streamKey` (string): The broadcaster's stream key to insert into the URL template

**Returns:**
- `string`: Complete RTMP URL ready for use with streaming software

