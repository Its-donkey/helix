---
layout: default
title: IRC/Chat Client
description: The IRC client provides real-time chat functionality via Twitch's IRC (Internet Relay Chat) interface over WebSocket. This is ideal for building chat bots that need to read and send messages in Twitch channels.
---

## NewIRCClient

Create a new IRC client for connecting to Twitch chat.

```go
client := helix.NewIRCClient("bot_username", "oauth:your-token",
    helix.WithMessageHandler(func(msg *helix.ChatMessage) {
        fmt.Printf("[%s] %s: %s\n", msg.Channel, msg.User, msg.Text)
    }),
)
```

## NewIRCClientE

Create a new IRC client with error handling for invalid inputs.

```go
client, err := helix.NewIRCClientE("bot_username", "oauth:your-token",
    helix.WithAutoReconnect(true),
)
if err != nil {
    log.Fatal(err)
}
```

## Connect

Establish a connection to Twitch IRC.

```go
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## Join

Join a channel to receive messages.

```go
if err := client.Join("channel_name"); err != nil {
    log.Printf("Failed to join: %v", err)
}
```

## Part

Leave a channel.

```go
if err := client.Part("channel_name"); err != nil {
    log.Printf("Failed to part: %v", err)
}
```

## Say

Send a message to a channel.

```go
if err := client.Say("channel_name", "Hello, chat!"); err != nil {
    log.Printf("Failed to send: %v", err)
}
```

## Reply

Reply to a specific message.

```go
if err := client.Reply("channel_name", "message-id", "This is a reply!"); err != nil {
    log.Printf("Failed to reply: %v", err)
}
```

## Close

Close the IRC connection.

```go
if err := client.Close(); err != nil {
    log.Printf("Failed to close: %v", err)
}
```

## Configuration Options

### WithAutoReconnect

Enable or disable automatic reconnection.

```go
helix.WithAutoReconnect(true)
```

### WithReconnectDelay

Set the delay between reconnection attempts.

```go
helix.WithReconnectDelay(5 * time.Second)
```

### WithIRCURL

Set a custom WebSocket URL.

```go
helix.WithIRCURL("wss://custom-irc.example.com")
```

## Event Handlers

### WithMessageHandler

Handle incoming chat messages.

```go
helix.WithMessageHandler(func(msg *helix.ChatMessage) {
    fmt.Printf("%s: %s\n", msg.User, msg.Text)
})
```

### WithUserNoticeHandler

Handle user notices (subscriptions, raids, etc.).

```go
helix.WithUserNoticeHandler(func(notice *helix.UserNotice) {
    fmt.Printf("User notice: %s\n", notice.MsgID)
})
```

### WithRoomStateHandler

Handle room state changes (slow mode, emote-only, etc.).

```go
helix.WithRoomStateHandler(func(state *helix.RoomState) {
    fmt.Printf("Room state updated: %s\n", state.Channel)
})
```

### WithClearChatHandler

Handle timeout and ban events.

```go
helix.WithClearChatHandler(func(clear *helix.ClearChat) {
    fmt.Printf("User %s was timed out\n", clear.TargetUserID)
})
```

### WithClearMessageHandler

Handle deleted messages.

```go
helix.WithClearMessageHandler(func(clear *helix.ClearMessage) {
    fmt.Printf("Message deleted: %s\n", clear.TargetMsgID)
})
```

### WithWhisperHandler

Handle whisper (private) messages.

```go
helix.WithWhisperHandler(func(whisper *helix.Whisper) {
    fmt.Printf("Whisper from %s: %s\n", whisper.User, whisper.Text)
})
```

### WithJoinHandler

Handle channel join events.

```go
helix.WithJoinHandler(func(channel, user string) {
    fmt.Printf("%s joined %s\n", user, channel)
})
```

### WithPartHandler

Handle channel leave events.

```go
helix.WithPartHandler(func(channel, user string) {
    fmt.Printf("%s left %s\n", user, channel)
})
```

### WithConnectHandler

Handle successful connection events.

```go
helix.WithConnectHandler(func() {
    fmt.Println("Connected to IRC!")
})
```

### WithDisconnectHandler

Handle disconnection events.

```go
helix.WithDisconnectHandler(func() {
    fmt.Println("Disconnected from IRC")
})
```

### WithReconnectHandler

Handle reconnection events.

```go
helix.WithReconnectHandler(func() {
    fmt.Println("Reconnected!")
    // Rejoin channels
})
```

### WithIRCErrorHandler

Handle errors.

```go
helix.WithIRCErrorHandler(func(err error) {
    log.Printf("IRC error: %v", err)
})
```

### WithRawMessageHandler

Handle raw IRC messages for debugging.

```go
helix.WithRawMessageHandler(func(raw string) {
    fmt.Printf("Raw: %s\n", raw)
})
```

## See Also

- [IRC Client Examples](examples/irc-client.md) - Complete code examples
- [Chat Bot Example](examples/chatbot.md) - Building a chat bot
- [Chat API](chat.md) - Helix API chat endpoints

