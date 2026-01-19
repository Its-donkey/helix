# Examples

Comprehensive code examples covering all kappopher features.

## Getting Started

- [Quick Start Guide](quickstart.md) - Installation, authentication, and basic setup
- [Basic Usage](basic.md) - Simple API calls, pagination, and error handling

## Authentication

- [Authentication](authentication.md) - All OAuth 2.0 flows (client credentials, authorization code, device code, implicit, token refresh, OIDC)

## Chat & Real-Time

- [Chat Bot](chatbot.md) - Building a chat bot with EventSub
- [IRC Client](Projects/Programming/Kappopher/Documents/examples/irc-client.md) - Low-level IRC/WebSocket chat client

## EventSub

- [EventSub Webhooks](eventsub-webhooks.md) - Handle webhook notifications with signature verification
- [EventSub WebSocket](eventsub-websocket.md) - Real-time event streaming without a public endpoint
- [PubSub Migration](pubsub-migration.md) - PubSub-style API using EventSub

## API Usage

- [API Usage Examples](api-usage.md) - Common API patterns for users, channels, streams, chat, moderation, polls, predictions, and clips
- [Channel Points](Projects/Programming/Kappopher/Documents/examples/channel-points.md) - Custom rewards, redemptions, real-time tracking
- [Bits & Subscriptions](bits-subscriptions.md) - Bits leaderboard, cheermotes, subscriber management
- [Videos & Clips](videos-clips.md) - VODs, highlights, clips, stream markers
- [Schedule & Goals](schedule-goals.md) - Stream schedule management, creator goals
- [Raids & Hype Train](raids-hypetrain.md) - Channel raids, hype train tracking
- [Analytics & Charity](analytics-charity.md) - Extension/game analytics, charity campaigns, teams, guest star, conduits

## Advanced

- [Batch & Caching](batch-caching.md) - Batch operations, caching, rate limiting, middleware
- [Extension JWT](Projects/Programming/Kappopher/Documents/examples/extension-jwt.md) - JWT authentication for Twitch Extensions

## Quick Reference

### By Feature

| Feature | Example File |
|---------|-------------|
| OAuth Authentication | [authentication.md](authentication.md) |
| Get Users/Channels | [basic.md](basic.md), [api-usage.md](api-usage.md) |
| Send Chat Messages | [chatbot.md](chatbot.md), [irc-client.md](Projects/Programming/Kappopher/Documents/examples/irc-client.md) |
| Handle Chat Events | [irc-client.md](Projects/Programming/Kappopher/Documents/examples/irc-client.md), [eventsub-websocket.md](eventsub-websocket.md) |
| Moderation | [chatbot.md](chatbot.md), [api-usage.md](api-usage.md) |
| Channel Points | [channel-points.md](Projects/Programming/Kappopher/Documents/examples/channel-points.md) |
| Bits | [bits-subscriptions.md](bits-subscriptions.md) |
| Subscriptions | [bits-subscriptions.md](bits-subscriptions.md) |
| Stream Schedule | [schedule-goals.md](schedule-goals.md) |
| Creator Goals | [schedule-goals.md](schedule-goals.md) |
| Raids | [raids-hypetrain.md](raids-hypetrain.md) |
| Hype Train | [raids-hypetrain.md](raids-hypetrain.md) |
| Clips | [videos-clips.md](videos-clips.md) |
| Videos/VODs | [videos-clips.md](videos-clips.md) |
| Polls & Predictions | [api-usage.md](api-usage.md) |
| EventSub WebSocket | [eventsub-websocket.md](eventsub-websocket.md) |
| EventSub Webhooks | [eventsub-webhooks.md](eventsub-webhooks.md) |
| Batch Requests | [batch-caching.md](batch-caching.md) |
| Caching | [batch-caching.md](batch-caching.md) |
| Middleware | [batch-caching.md](batch-caching.md) |
| Extensions | [extension-jwt.md](Projects/Programming/Kappopher/Documents/examples/extension-jwt.md) |
| Analytics | [analytics-charity.md](analytics-charity.md) |
| Charity | [analytics-charity.md](analytics-charity.md) |
| Teams | [analytics-charity.md](analytics-charity.md) |
| Conduits | [analytics-charity.md](analytics-charity.md) |

### By Use Case

| Use Case | Start Here |
|----------|-----------|
| Build a chat bot | [chatbot.md](chatbot.md) or [irc-client.md](Projects/Programming/Kappopher/Documents/examples/irc-client.md) |
| Monitor stream events | [eventsub-websocket.md](eventsub-websocket.md) |
| Create a dashboard | [batch-caching.md](batch-caching.md) |
| Manage channel points | [channel-points.md](Projects/Programming/Kappopher/Documents/examples/channel-points.md) |
| Track subscriptions | [bits-subscriptions.md](bits-subscriptions.md) |
| Server-side integration | [eventsub-webhooks.md](eventsub-webhooks.md) |
| Build an extension | [extension-jwt.md](Projects/Programming/Kappopher/Documents/examples/extension-jwt.md) |
| Migrate from PubSub | [pubsub-migration.md](pubsub-migration.md) |
