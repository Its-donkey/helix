# Twitch Helix API Documentation

This documentation provides detailed guides and examples for using the Twitch Helix API wrapper.

## Table of Contents

- [Quick Start](quickstart.md) - Installation, basic usage, authentication, and error handling
- [Available Endpoints](#available-endpoints) - All API methods
- [Examples](#examples) - Working code samples

## Available Endpoints

| Category | Documentation |
|----------|---------------|
| [Ads](ads.md) | Start commercials, manage ad schedules |
| [Analytics](analytics.md) | Extension and game analytics |
| [Auth](auth.md) | OAuth flows, token management, OIDC |
| [Bits](bits.md) | Bits leaderboard, cheermotes |
| [Channel Points](channel-points.md) | Custom rewards, redemptions |
| [Channels](channels.md) | Channel info, followers, editors, VIPs |
| [Charity](charity.md) | Charity campaigns and donations |
| [Chat](chat.md) | Chatters, emotes, badges, settings, messages |
| [Clips](clips.md) | Create and get clips |
| [CCL](ccl.md) | Content classification labels |
| [Conduits](conduits.md) | EventSub conduit management |
| [Entitlements](entitlements.md) | Drops entitlements |
| [EventSub](eventsub.md) | Event subscriptions |
| [PubSub Compatibility](pubsub-compat.md) | PubSub-style API backed by EventSub |
| [Extensions](extensions.md) | Extension management |
| [Games](games.md) | Game information |
| [Goals](goals.md) | Creator goals |
| [Guest Star](guest-star.md) | Guest star sessions |
| [Hype Train](hype-train.md) | Hype train events and status |
| [Ingest](ingest.md) | Ingest server endpoints |
| [Moderation](moderation.md) | Bans, mods, AutoMod, shield mode |
| [Polls](polls.md) | Create and manage polls |
| [Predictions](predictions.md) | Create and manage predictions |
| [Raids](raids.md) | Start and cancel raids |
| [Schedule](schedule.md) | Stream schedule management |
| [Search](search.md) | Search categories and channels |
| [Streams](streams.md) | Stream info, markers |
| [Subscriptions](subscriptions.md) | Subscriber info |
| [Teams](teams.md) | Team information |
| [Users](users.md) | User info, blocks, extensions |
| [Videos](videos.md) | VODs, highlights, uploads |
| [Whispers](whispers.md) | Send whisper messages |

## Examples

See the [examples](./examples/) directory for code samples:

- [Basic Usage](./examples/basic.md) - Simple API calls and error handling
- [Chat Bot](./examples/chatbot.md) - Building a chat bot
- [API Usage](./examples/api-usage.md) - Common API patterns
- [EventSub Webhooks](./examples/eventsub-webhooks.md) - Webhook notifications
- [EventSub WebSocket](./examples/eventsub-websocket.md) - Real-time events
- [PubSub Migration](./examples/pubsub-migration.md) - Migrating from PubSub to EventSub
- [Extension JWT](./examples/extension-jwt.md) - Extension authentication
