---
layout: default
title: IRC Client Examples
description: Build chat bots using the low-level IRC/WebSocket client.
---

## Overview

The IRC client provides direct WebSocket connection to Twitch chat, offering:
- Lower latency than EventSub for chat messages
- Direct message sending without API rate limits
- Full IRC message parsing with tags (badges, emotes, etc.)
- Support for all TMI (Twitch Messaging Interface) events

## Basic IRC Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Create IRC client
    irc := helix.NewIRCClient(
        helix.WithIRCToken("oauth:your-oauth-token"), // Get from twitchtokengenerator.com
        helix.WithIRCNick("your_bot_username"),
        helix.WithIRCMessageHandler(func(msg *helix.IRCMessage) {
            if msg.Command == "PRIVMSG" {
                fmt.Printf("[%s] %s: %s\n", msg.Channel, msg.User, msg.Text)
            }
        }),
    )

    // Connect
    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    // Join channels
    if err := irc.Join("channel1"); err != nil {
        log.Fatal(err)
    }
    if err := irc.Join("channel2"); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to chat!")

    // Keep running
    select {}
}
```

## ChatBotClient (High-Level)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Setup auth client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Create Helix client for API calls
    helixClient := helix.NewClient("your-client-id", authClient)

    // Create ChatBotClient (combines IRC + Helix API)
    bot := helix.NewChatBotClient(helixClient,
        helix.WithChatBotMessageHandler(func(msg *helix.IRCMessage) {
            handleMessage(ctx, bot, msg)
        }),
        helix.WithChatBotAutoReconnect(true),
    )

    // Connect with user token
    if err := bot.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer bot.Close()

    // Join channel
    if err := bot.Join("channel_name"); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Bot is running!")
    select {}
}

func handleMessage(ctx context.Context, bot *helix.ChatBotClient, msg *helix.IRCMessage) {
    if msg.Command != "PRIVMSG" {
        return
    }

    // Respond to commands
    switch msg.Text {
    case "!ping":
        bot.Say(msg.Channel, "Pong!")
    case "!time":
        bot.Say(msg.Channel, fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC1123)))
    }
}
```

## IRC Message Handling

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    irc := helix.NewIRCClient(
        helix.WithIRCToken("oauth:token"),
        helix.WithIRCNick("bot_name"),
        helix.WithIRCMessageHandler(handleIRCMessage),
    )

    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    irc.Join("channel")
    select {}
}

func handleIRCMessage(msg *helix.IRCMessage) {
    switch msg.Command {
    case "PRIVMSG":
        handleChatMessage(msg)

    case "USERNOTICE":
        handleUserNotice(msg)

    case "ROOMSTATE":
        handleRoomState(msg)

    case "CLEARCHAT":
        handleClearChat(msg)

    case "CLEARMSG":
        handleClearMsg(msg)

    case "NOTICE":
        fmt.Printf("Notice: %s\n", msg.Text)

    case "RECONNECT":
        fmt.Println("Server requested reconnect")

    case "USERSTATE":
        fmt.Printf("User state updated in %s\n", msg.Channel)
    }
}

func handleChatMessage(msg *helix.IRCMessage) {
    // Access IRC tags
    displayName := msg.Tags["display-name"]
    badges := msg.Tags["badges"]
    color := msg.Tags["color"]
    emotes := msg.Tags["emotes"]
    msgID := msg.Tags["id"]

    // Check for bits
    if bits, ok := msg.Tags["bits"]; ok {
        fmt.Printf("💎 %s cheered %s bits: %s\n", displayName, bits, msg.Text)
        return
    }

    // Check user status
    isMod := msg.Tags["mod"] == "1"
    isSub := msg.Tags["subscriber"] == "1"
    isVIP := msg.Tags["vip"] == "1"

    prefix := ""
    if isMod {
        prefix = "🗡️"
    } else if isVIP {
        prefix = "💎"
    } else if isSub {
        prefix = "⭐"
    }

    fmt.Printf("%s[%s] %s: %s\n", prefix, msg.Channel, displayName, msg.Text)

    // Store message ID for potential moderation
    _ = msgID
    _ = badges
    _ = color
    _ = emotes
}

func handleUserNotice(msg *helix.IRCMessage) {
    msgID := msg.Tags["msg-id"]
    displayName := msg.Tags["display-name"]

    switch msgID {
    case "sub":
        plan := msg.Tags["msg-param-sub-plan"]
        fmt.Printf("🎉 %s subscribed! (Tier %s)\n", displayName, tierName(plan))

    case "resub":
        months := msg.Tags["msg-param-cumulative-months"]
        plan := msg.Tags["msg-param-sub-plan"]
        fmt.Printf("🎉 %s resubscribed for %s months! (Tier %s)\n", displayName, months, tierName(plan))
        if msg.Text != "" {
            fmt.Printf("   Message: %s\n", msg.Text)
        }

    case "subgift":
        recipient := msg.Tags["msg-param-recipient-display-name"]
        plan := msg.Tags["msg-param-sub-plan"]
        fmt.Printf("🎁 %s gifted a sub to %s! (Tier %s)\n", displayName, recipient, tierName(plan))

    case "submysterygift":
        count := msg.Tags["msg-param-mass-gift-count"]
        fmt.Printf("🎁 %s gifted %s subs to the community!\n", displayName, count)

    case "raid":
        viewers := msg.Tags["msg-param-viewerCount"]
        fmt.Printf("🚀 %s is raiding with %s viewers!\n", displayName, viewers)

    case "announcement":
        fmt.Printf("📢 Announcement from %s: %s\n", displayName, msg.Text)
    }
}

func handleRoomState(msg *helix.IRCMessage) {
    channel := msg.Channel
    fmt.Printf("Room state for %s:\n", channel)

    if emoteOnly := msg.Tags["emote-only"]; emoteOnly == "1" {
        fmt.Println("  - Emote-only mode ON")
    }
    if followersOnly := msg.Tags["followers-only"]; followersOnly != "-1" {
        fmt.Printf("  - Followers-only: %s minutes\n", followersOnly)
    }
    if slow := msg.Tags["slow"]; slow != "0" {
        fmt.Printf("  - Slow mode: %s seconds\n", slow)
    }
    if subsOnly := msg.Tags["subs-only"]; subsOnly == "1" {
        fmt.Println("  - Subscribers-only mode ON")
    }
    if r9k := msg.Tags["r9k"]; r9k == "1" {
        fmt.Println("  - R9K mode ON")
    }
}

func handleClearChat(msg *helix.IRCMessage) {
    if targetUser := msg.Tags["target-user-id"]; targetUser != "" {
        duration := msg.Tags["ban-duration"]
        if duration == "" {
            fmt.Printf("🔨 %s was permanently banned\n", msg.Text)
        } else {
            fmt.Printf("⏰ %s was timed out for %s seconds\n", msg.Text, duration)
        }
    } else {
        fmt.Printf("🧹 Chat was cleared in %s\n", msg.Channel)
    }
}

func handleClearMsg(msg *helix.IRCMessage) {
    targetMsgID := msg.Tags["target-msg-id"]
    login := msg.Tags["login"]
    fmt.Printf("🗑️ Message from %s deleted (ID: %s)\n", login, targetMsgID)
}

func tierName(plan string) string {
    switch plan {
    case "Prime":
        return "Prime"
    case "1000":
        return "1"
    case "2000":
        return "2"
    case "3000":
        return "3"
    default:
        return plan
    }
}
```

## Sending Messages

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    irc := helix.NewIRCClient(
        helix.WithIRCToken("oauth:token"),
        helix.WithIRCNick("bot_name"),
    )

    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    channel := "channel_name"
    irc.Join(channel)

    // Send a message
    if err := irc.Say(channel, "Hello, chat!"); err != nil {
        log.Printf("Failed to send message: %v", err)
    }

    // Reply to a message
    if err := irc.Reply(channel, "message-id-here", "This is a reply!"); err != nil {
        log.Printf("Failed to send reply: %v", err)
    }

    // Send /me action
    if err := irc.Say(channel, "/me waves at chat"); err != nil {
        log.Printf("Failed to send action: %v", err)
    }
}
```

## Auto-Reconnect

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    irc := helix.NewIRCClient(
        helix.WithIRCToken("oauth:token"),
        helix.WithIRCNick("bot_name"),
        helix.WithIRCAutoReconnect(true),
        helix.WithIRCReconnectHandler(func() {
            fmt.Println("Reconnected! Rejoining channels...")
            // Rejoin channels after reconnect
            irc.Join("channel1")
            irc.Join("channel2")
        }),
        helix.WithIRCErrorHandler(func(err error) {
            log.Printf("IRC error: %v", err)
        }),
    )

    if err := irc.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer irc.Close()

    irc.Join("channel1")
    irc.Join("channel2")

    fmt.Println("Bot running with auto-reconnect...")
    select {}
}
```

## Complete Chat Bot Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "strings"
    "sync"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

type ChatBot struct {
    irc      *helix.IRCClient
    commands map[string]CommandFunc
    cooldown map[string]time.Time
    cdMu     sync.Mutex
}

type CommandFunc func(msg *helix.IRCMessage, args []string) string

func NewChatBot(token, nick string) *ChatBot {
    bot := &ChatBot{
        commands: make(map[string]CommandFunc),
        cooldown: make(map[string]time.Time),
    }

    bot.irc = helix.NewIRCClient(
        helix.WithIRCToken(token),
        helix.WithIRCNick(nick),
        helix.WithIRCAutoReconnect(true),
        helix.WithIRCMessageHandler(bot.handleMessage),
    )

    // Register commands
    bot.commands["!ping"] = func(msg *helix.IRCMessage, args []string) string {
        return "Pong!"
    }

    bot.commands["!dice"] = func(msg *helix.IRCMessage, args []string) string {
        return fmt.Sprintf("@%s rolled a %d!", msg.Tags["display-name"], rand.Intn(6)+1)
    }

    bot.commands["!uptime"] = func(msg *helix.IRCMessage, args []string) string {
        // Implement actual uptime check
        return "Stream has been live for 2 hours!"
    }

    bot.commands["!commands"] = func(msg *helix.IRCMessage, args []string) string {
        cmds := make([]string, 0, len(bot.commands))
        for cmd := range bot.commands {
            cmds = append(cmds, cmd)
        }
        return "Commands: " + strings.Join(cmds, ", ")
    }

    bot.commands["!8ball"] = func(msg *helix.IRCMessage, args []string) string {
        answers := []string{
            "Yes!", "No!", "Maybe...", "Ask again later",
            "Definitely!", "I don't think so", "It is certain",
        }
        return fmt.Sprintf("🎱 %s", answers[rand.Intn(len(answers))])
    }

    return bot
}

func (b *ChatBot) handleMessage(msg *helix.IRCMessage) {
    if msg.Command != "PRIVMSG" {
        return
    }

    text := strings.TrimSpace(msg.Text)
    if !strings.HasPrefix(text, "!") {
        return
    }

    parts := strings.Fields(text)
    if len(parts) == 0 {
        return
    }

    cmdName := strings.ToLower(parts[0])
    args := parts[1:]

    // Check cooldown (5 seconds per user per command)
    cdKey := fmt.Sprintf("%s:%s:%s", msg.Channel, msg.User, cmdName)
    b.cdMu.Lock()
    if lastUse, ok := b.cooldown[cdKey]; ok && time.Since(lastUse) < 5*time.Second {
        b.cdMu.Unlock()
        return
    }
    b.cooldown[cdKey] = time.Now()
    b.cdMu.Unlock()

    // Execute command
    if handler, ok := b.commands[cmdName]; ok {
        response := handler(msg, args)
        if response != "" {
            b.irc.Say(msg.Channel, response)
        }
    }
}

func (b *ChatBot) Connect(ctx context.Context) error {
    return b.irc.Connect(ctx)
}

func (b *ChatBot) Join(channel string) error {
    return b.irc.Join(channel)
}

func (b *ChatBot) Close() error {
    return b.irc.Close()
}

func main() {
    ctx := context.Background()

    bot := NewChatBot("oauth:your-token", "your_bot_name")

    if err := bot.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer bot.Close()

    // Join channels
    channels := []string{"channel1", "channel2"}
    for _, ch := range channels {
        if err := bot.Join(ch); err != nil {
            log.Printf("Failed to join %s: %v", ch, err)
        }
    }

    fmt.Println("Bot is running!")
    select {}
}
```

## IRC vs EventSub Comparison

| Feature | IRC Client | EventSub |
|---------|-----------|----------|
| Chat Messages | ✅ Real-time | ✅ Real-time |
| Latency | Lower (~50ms) | Higher (~200ms) |
| Message Sending | Direct | Via API |
| Rate Limits | IRC limits (20/30s) | API limits |
| Badges/Emotes | In tags | In event data |
| Subscriptions | USERNOTICE | EventSub events |
| Bits | In message tags | EventSub events |
| Raids | USERNOTICE | EventSub events |
| Moderation | CLEARCHAT/CLEARMSG | EventSub events |
| Connection | WebSocket to TMI | WebSocket to EventSub |

**Use IRC when:**
- You need lowest latency chat
- Building a chat-focused bot
- You need to send many messages

**Use EventSub when:**
- You need non-chat events (follows, subs, etc.)
- Building a dashboard/overlay
- You need reliable delivery with acknowledgment

