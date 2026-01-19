---
layout: default
title: Chat Bot Example
description: Build a Twitch chat bot using the Helix API and EventSub WebSocket for real-time message handling.
---

## Overview

This guide demonstrates how to build a feature-rich chat bot that can:
- Receive and respond to chat messages in real-time
- Execute custom commands (e.g., `!hello`, `!dice`)
- Perform moderation actions (ban, timeout, delete messages)
- Send announcements and shoutouts
- Control chat settings (slow mode, sub-only mode, etc.)

**Architecture**: The bot uses EventSub WebSocket to receive chat events and the Helix API to send messages and perform actions. This approach provides reliable event delivery with acknowledgment.

**Alternative**: For lower latency chat handling, see the [IRC Client](Projects/Programming/Kappopher/Documents/examples/irc-client.md) examples.

## Prerequisites

Chat bots require user authentication with the following scopes:
- `chat:read` - Read chat messages
- `chat:edit` - Send chat messages
- `moderator:manage:chat_messages` - Delete messages (optional)
- `moderator:manage:banned_users` - Ban/timeout users (optional)

## Basic Chat Bot

A minimal chat bot that sends a single message to chat. This demonstrates the core setup required for any bot: authentication, client creation, and message sending.

**Note**: This example only sends messages - it doesn't receive them. For a fully interactive bot, see the EventSub example below.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    ctx := context.Background()

    // Create auth client with bot scopes
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURI:  "http://localhost:3000/callback",
        Scopes:       helix.CommonScopes.Bot,
    })

    // Get authorization URL for user to approve
    url, _ := authClient.GetCodeAuthURL()
    fmt.Printf("Authorize at: %s\n", url)

    // After user authorizes, exchange code for token
    // token, _ := authClient.ExchangeCode(ctx, "authorization-code")

    // Create Helix client
    client := helix.NewClient("your-client-id", authClient)

    // Bot configuration
    broadcasterID := "12345"  // Channel to join
    botUserID := "67890"      // Bot's user ID

    // Send a message to chat
    err := sendMessage(ctx, client, broadcasterID, botUserID, "Hello chat! Bot is online.")
    if err != nil {
        log.Fatal(err)
    }
}

func sendMessage(ctx context.Context, client *helix.Client, broadcasterID, senderID, message string) error {
    _, err := client.SendChatMessage(ctx, &helix.SendChatMessageParams{
        BroadcasterID: broadcasterID,
        SenderID:      senderID,
        Message:       message,
    })
    return err
}
```

## Responding to Events with EventSub

A complete chat bot with command handling using EventSub WebSocket. This approach:
- Receives chat messages in real-time via WebSocket
- Parses messages to detect commands (starting with `!`)
- Executes registered command handlers
- Sends responses back to chat

**Key components**:
- `ChatBot` struct: Manages client, configuration, and command registry
- `CommandHandler`: Function signature for command implementations
- `HandleMessage`: Parses incoming messages and routes to commands
- `RegisterCommand`: Adds new commands dynamically

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"

    "github.com/Its-donkey/helix/helix"
)

type ChatBot struct {
    client        *helix.Client
    broadcasterID string
    botUserID     string
    commands      map[string]CommandHandler
}

type CommandHandler func(ctx context.Context, event *helix.ChannelChatMessageEvent) string

func NewChatBot(client *helix.Client, broadcasterID, botUserID string) *ChatBot {
    bot := &ChatBot{
        client:        client,
        broadcasterID: broadcasterID,
        botUserID:     botUserID,
        commands:      make(map[string]CommandHandler),
    }

    // Register default commands
    bot.RegisterCommand("!hello", func(ctx context.Context, e *helix.ChannelChatMessageEvent) string {
        return fmt.Sprintf("Hello @%s!", e.ChatterUserName)
    })

    bot.RegisterCommand("!uptime", func(ctx context.Context, e *helix.ChannelChatMessageEvent) string {
        return "Stream has been live for 2 hours!" // Implement actual uptime check
    })

    bot.RegisterCommand("!commands", func(ctx context.Context, e *helix.ChannelChatMessageEvent) string {
        return "Available commands: !hello, !uptime, !commands"
    })

    return bot
}

func (b *ChatBot) RegisterCommand(name string, handler CommandHandler) {
    b.commands[strings.ToLower(name)] = handler
}

func (b *ChatBot) HandleMessage(ctx context.Context, event *helix.ChannelChatMessageEvent) {
    message := strings.TrimSpace(event.Message.Text)

    // Check if message is a command
    if !strings.HasPrefix(message, "!") {
        return
    }

    // Get command name (first word)
    parts := strings.Fields(message)
    if len(parts) == 0 {
        return
    }
    cmdName := strings.ToLower(parts[0])

    // Find and execute command
    if handler, ok := b.commands[cmdName]; ok {
        response := handler(ctx, event)
        if response != "" {
            b.SendMessage(ctx, response)
        }
    }
}

func (b *ChatBot) SendMessage(ctx context.Context, message string) error {
    _, err := b.client.SendChatMessage(ctx, &helix.SendChatMessageParams{
        BroadcasterID: b.broadcasterID,
        SenderID:      b.botUserID,
        Message:       message,
    })
    return err
}

func (b *ChatBot) SendReply(ctx context.Context, replyToID, message string) error {
    _, err := b.client.SendChatMessage(ctx, &helix.SendChatMessageParams{
        BroadcasterID:        b.broadcasterID,
        SenderID:             b.botUserID,
        Message:              message,
        ReplyParentMessageID: replyToID,
    })
    return err
}

func main() {
    ctx := context.Background()

    // Setup auth and client (see basic example)
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    _, _ = authClient.GetAppAccessToken(ctx)
    client := helix.NewClient("your-client-id", authClient)

    // Create bot
    bot := NewChatBot(client, "broadcaster-id", "bot-user-id")

    // Add custom command
    bot.RegisterCommand("!dice", func(ctx context.Context, e *helix.ChannelChatMessageEvent) string {
        return fmt.Sprintf("@%s rolled a %d!", e.ChatterUserName, 1+rand.Intn(6))
    })

    // Connect to EventSub WebSocket
    ws := helix.NewEventSubWebSocket(client)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Subscribe to chat messages
    ws.Subscribe(ctx, helix.EventSubTypeChannelChatMessage, "1",
        map[string]string{
            "broadcaster_user_id": "broadcaster-id",
            "user_id":             "bot-user-id",
        },
        func(event json.RawMessage) {
            e, err := helix.ParseWSEvent[helix.ChannelChatMessageEvent](event)
            if err != nil {
                log.Printf("Failed to parse chat message: %v", err)
                return
            }
            bot.HandleMessage(ctx, e)
        },
    )

    fmt.Println("Bot is running...")
    select {} // Keep running
}
```

## Moderation Features

Add moderation capabilities to your bot. These require the bot to have moderator privileges in the channel and appropriate scopes.

**Required scope**: `moderator:manage:banned_users` for bans/timeouts, `moderator:manage:chat_messages` for message deletion.

**Note**: The bot must be a moderator in the channel to perform these actions.

```go
// Timeout a user
func (b *ChatBot) TimeoutUser(ctx context.Context, userID string, duration int, reason string) error {
    return b.client.BanUser(ctx, &helix.BanUserParams{
        BroadcasterID: b.broadcasterID,
        ModeratorID:   b.botUserID,
        Data: helix.BanUserData{
            UserID:   userID,
            Duration: duration,
            Reason:   reason,
        },
    })
}

// Ban a user permanently
func (b *ChatBot) BanUser(ctx context.Context, userID, reason string) error {
    return b.client.BanUser(ctx, &helix.BanUserParams{
        BroadcasterID: b.broadcasterID,
        ModeratorID:   b.botUserID,
        Data: helix.BanUserData{
            UserID: userID,
            Reason: reason,
        },
    })
}

// Unban a user
func (b *ChatBot) UnbanUser(ctx context.Context, userID string) error {
    return b.client.UnbanUser(ctx, b.broadcasterID, b.botUserID, userID)
}

// Delete a message
func (b *ChatBot) DeleteMessage(ctx context.Context, messageID string) error {
    return b.client.DeleteChatMessages(ctx, &helix.DeleteChatMessagesParams{
        BroadcasterID: b.broadcasterID,
        ModeratorID:   b.botUserID,
        MessageID:     messageID,
    })
}

// Clear chat
func (b *ChatBot) ClearChat(ctx context.Context) error {
    return b.client.DeleteChatMessages(ctx, &helix.DeleteChatMessagesParams{
        BroadcasterID: b.broadcasterID,
        ModeratorID:   b.botUserID,
    })
}
```

## Announcements

Send highlighted announcements that stand out in chat. Announcements appear with a colored background and are useful for important messages.

**Required scope**: `moderator:manage:announcements`

**Colors available**: `"blue"`, `"green"`, `"orange"`, `"purple"`, or `"primary"` (channel accent color)

```go
// Send an announcement
func (b *ChatBot) Announce(ctx context.Context, message, color string) error {
    return b.client.SendChatAnnouncement(ctx, &helix.SendChatAnnouncementParams{
        BroadcasterID: b.broadcasterID,
        ModeratorID:   b.botUserID,
        Message:       message,
        Color:         color, // "blue", "green", "orange", "purple", or "primary"
    })
}
```

## Shoutouts

Send a shoutout to promote another streamer. This displays a card in chat with information about the target channel.

**Required scope**: `moderator:manage:shoutouts`

**Cooldown**: There's a 2-minute cooldown between shoutouts to the same user.

```go
// Give a shoutout to another streamer
func (b *ChatBot) Shoutout(ctx context.Context, targetUserID string) error {
    return b.client.SendShoutout(ctx, &helix.SendShoutoutParams{
        FromBroadcasterID: b.broadcasterID,
        ToBroadcasterID:   targetUserID,
        ModeratorID:       b.botUserID,
    })
}
```

## Chat Settings

Control chat modes programmatically. Useful for automating chat management during raids, high-activity moments, or scheduled events.

**Required scope**: `moderator:manage:chat_settings`

**Available modes**:
- **Slow mode**: Limit how often users can send messages
- **Subscriber-only mode**: Only subscribers can chat
- **Emote-only mode**: Only emotes allowed in messages
- **Follower-only mode**: Only followers can chat (with optional minimum follow time)

```go
// Enable slow mode
func (b *ChatBot) EnableSlowMode(ctx context.Context, seconds int) error {
    return b.client.UpdateChatSettings(ctx, &helix.UpdateChatSettingsParams{
        BroadcasterID:    b.broadcasterID,
        ModeratorID:      b.botUserID,
        SlowMode:         boolPtr(true),
        SlowModeWaitTime: &seconds,
    })
}

// Enable subscriber-only mode
func (b *ChatBot) EnableSubOnlyMode(ctx context.Context) error {
    return b.client.UpdateChatSettings(ctx, &helix.UpdateChatSettingsParams{
        BroadcasterID:      b.broadcasterID,
        ModeratorID:        b.botUserID,
        SubscriberMode:     boolPtr(true),
    })
}

// Enable emote-only mode
func (b *ChatBot) EnableEmoteOnlyMode(ctx context.Context) error {
    return b.client.UpdateChatSettings(ctx, &helix.UpdateChatSettingsParams{
        BroadcasterID: b.broadcasterID,
        ModeratorID:   b.botUserID,
        EmoteMode:     boolPtr(true),
    })
}

func boolPtr(b bool) *bool { return &b }
```

