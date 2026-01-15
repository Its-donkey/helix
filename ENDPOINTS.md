# Twitch API Endpoints Support

This document lists all Twitch API endpoints and indicates whether they are supported by this client.

**Legend:**
- Supported: The endpoint is implemented and available
- Last Updated: When the endpoint implementation was last verified
- Twitch Note: Any special status from Twitch (deprecated, beta, new)

---

## Authentication (OIDC)

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Auth | OpenID Configuration | GET | `https://id.twitch.tv/oauth2/.well-known/openid-configuration` | 2025-12-10 | |
| Yes | Auth | Authorize | GET | `https://id.twitch.tv/oauth2/authorize` | 2025-12-10 | |
| Yes | Auth | Token | POST | `https://id.twitch.tv/oauth2/token` | 2025-12-10 | |
| Yes | Auth | Validate Token | GET | `https://id.twitch.tv/oauth2/validate` | 2025-12-10 | |
| Yes | Auth | Revoke Token | POST | `https://id.twitch.tv/oauth2/revoke` | 2025-12-10 | |
| Yes | Auth | Device Code | POST | `https://id.twitch.tv/oauth2/device` | 2025-12-10 | |
| Yes | Auth | UserInfo (OIDC) | GET | `https://id.twitch.tv/oauth2/userinfo` | 2025-12-10 | OIDC |
| Yes | Auth | JWKS | GET | `https://id.twitch.tv/oauth2/keys` | 2025-12-10 | OIDC |

## Ads

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Ads | Start Commercial | POST | `/helix/channels/commercial` | 2025-12-10 | |
| Yes | Ads | Get Ad Schedule | GET | `/helix/channels/ads` | 2025-12-10 | |
| Yes | Ads | Snooze Next Ad | POST | `/helix/channels/ads/schedule/snooze` | 2025-12-10 | |

## Analytics

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Analytics | Get Extension Analytics | GET | `/helix/analytics/extensions` | 2025-12-10 | |
| Yes | Analytics | Get Game Analytics | GET | `/helix/analytics/games` | 2025-12-10 | |

## Bits

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Bits | Get Bits Leaderboard | GET | `/helix/bits/leaderboard` | 2025-12-10 | |
| Yes | Bits | Get Cheermotes | GET | `/helix/bits/cheermotes` | 2025-12-10 | |
| Yes | Bits | Get Extension Transactions | GET | `/helix/extensions/transactions` | 2025-12-10 | |

## Channel Points

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Channel Points | Create Custom Rewards | POST | `/helix/channel_points/custom_rewards` | 2025-12-10 | |
| Yes | Channel Points | Delete Custom Reward | DELETE | `/helix/channel_points/custom_rewards` | 2025-12-10 | |
| Yes | Channel Points | Get Custom Reward | GET | `/helix/channel_points/custom_rewards` | 2025-12-10 | |
| Yes | Channel Points | Get Custom Reward Redemption | GET | `/helix/channel_points/custom_rewards/redemptions` | 2025-12-10 | |
| Yes | Channel Points | Update Custom Reward | PATCH | `/helix/channel_points/custom_rewards` | 2025-12-10 | |
| Yes | Channel Points | Update Redemption Status | PATCH | `/helix/channel_points/custom_rewards/redemptions` | 2025-12-10 | |

## Channels

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Channels | Get Channel Information | GET | `/helix/channels` | 2025-12-10 | |
| Yes | Channels | Modify Channel Information | PATCH | `/helix/channels` | 2025-12-10 | |
| Yes | Channels | Get Channel Editors | GET | `/helix/channels/editors` | 2025-12-10 | |
| Yes | Channels | Get Followed Channels | GET | `/helix/channels/followed` | 2025-12-10 | |
| Yes | Channels | Get Channel Followers | GET | `/helix/channels/followers` | 2025-12-10 | |

## Charity

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Charity | Get Charity Campaign | GET | `/helix/charity/campaigns` | 2025-12-10 | |
| Yes | Charity | Get Charity Campaign Donations | GET | `/helix/charity/donations` | 2025-12-10 | |

## Chat

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Chat | Get Chatters | GET | `/helix/chat/chatters` | 2025-12-10 | |
| Yes | Chat | Get Channel Emotes | GET | `/helix/chat/emotes` | 2025-12-10 | |
| Yes | Chat | Get Global Emotes | GET | `/helix/chat/emotes/global` | 2025-12-10 | |
| Yes | Chat | Get Emote Sets | GET | `/helix/chat/emotes/set` | 2025-12-10 | |
| Yes | Chat | Get Channel Chat Badges | GET | `/helix/chat/badges` | 2025-12-10 | |
| Yes | Chat | Get Global Chat Badges | GET | `/helix/chat/badges/global` | 2025-12-10 | |
| Yes | Chat | Get Chat Settings | GET | `/helix/chat/settings` | 2025-12-10 | |
| Yes | Chat | Get Shared Chat Session | GET | `/helix/chat/shared_chat/sessions` | 2025-12-10 | NEW |
| Yes | Chat | Get User Emotes | GET | `/helix/chat/user/emotes` | 2025-12-10 | NEW |
| Yes | Chat | Update Chat Settings | PATCH | `/helix/chat/settings` | 2025-12-10 | |
| Yes | Chat | Send Chat Announcement | POST | `/helix/chat/announcements` | 2025-12-10 | |
| Yes | Chat | Send a Shoutout | POST | `/helix/chat/shoutouts` | 2025-12-10 | |
| Yes | Chat | Send Chat Message | POST | `/helix/chat/messages` | 2025-12-10 | NEW |
| Yes | Chat | Get User Chat Color | GET | `/helix/chat/color` | 2025-12-10 | |
| Yes | Chat | Update User Chat Color | PUT | `/helix/chat/color` | 2025-12-10 | |

## Clips

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Clips | Create Clip | POST | `/helix/clips` | 2025-12-10 | |
| Yes | Clips | Create Clip From VOD | POST | `/helix/videos/clips` | 2026-01-13 | |
| Yes | Clips | Get Clips | GET | `/helix/clips` | 2025-12-10 | |
| Yes | Clips | Get Clips Download | GET | `/helix/clips/download` | 2025-12-10 | NEW |

## Conduits

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Conduits | Get Conduits | GET | `/helix/eventsub/conduits` | 2025-12-10 | |
| Yes | Conduits | Create Conduits | POST | `/helix/eventsub/conduits` | 2025-12-10 | |
| Yes | Conduits | Update Conduits | PATCH | `/helix/eventsub/conduits` | 2025-12-10 | |
| Yes | Conduits | Delete Conduit | DELETE | `/helix/eventsub/conduits` | 2025-12-10 | |
| Yes | Conduits | Get Conduit Shards | GET | `/helix/eventsub/conduits/shards` | 2025-12-10 | |
| Yes | Conduits | Update Conduit Shards | PATCH | `/helix/eventsub/conduits/shards` | 2025-12-10 | |

## Content Classification Labels

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | CCL | Get Content Classification Labels | GET | `/helix/content_classification_labels` | 2025-12-10 | |

## Entitlements

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Entitlements | Get Drops Entitlements | GET | `/helix/entitlements/drops` | 2025-12-10 | |
| Yes | Entitlements | Update Drops Entitlements | PATCH | `/helix/entitlements/drops` | 2025-12-10 | |

## EventSub

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | EventSub | Create EventSub Subscription | POST | `/helix/eventsub/subscriptions` | 2025-12-10 | |
| Yes | EventSub | Delete EventSub Subscription | DELETE | `/helix/eventsub/subscriptions` | 2025-12-10 | |
| Yes | EventSub | Get EventSub Subscriptions | GET | `/helix/eventsub/subscriptions` | 2025-12-10 | |

## Extensions

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Extensions | Get Extension Configuration Segment | GET | `/helix/extensions/configurations` | 2025-12-10 | |
| Yes | Extensions | Set Extension Configuration Segment | PUT | `/helix/extensions/configurations` | 2025-12-10 | |
| Yes | Extensions | Set Extension Required Configuration | PUT | `/helix/extensions/required_configuration` | 2025-12-10 | |
| Yes | Extensions | Send Extension PubSub Message | POST | `/helix/extensions/pubsub` | 2025-12-10 | |
| Yes | Extensions | Get Extension Live Channels | GET | `/helix/extensions/live` | 2025-12-10 | |
| Yes | Extensions | Get Extension Secrets | GET | `/helix/extensions/jwt/secrets` | 2025-12-10 | |
| Yes | Extensions | Create Extension Secret | POST | `/helix/extensions/jwt/secrets` | 2025-12-10 | |
| Yes | Extensions | Send Extension Chat Message | POST | `/helix/extensions/chat` | 2025-12-10 | |
| Yes | Extensions | Get Extensions | GET | `/helix/extensions` | 2025-12-10 | |
| Yes | Extensions | Get Released Extensions | GET | `/helix/extensions/released` | 2025-12-10 | |
| Yes | Extensions | Get Extension Bits Products | GET | `/helix/bits/extensions` | 2025-12-10 | |
| Yes | Extensions | Update Extension Bits Product | PUT | `/helix/bits/extensions` | 2025-12-10 | |

## Games

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Games | Get Top Games | GET | `/helix/games/top` | 2025-12-10 | |
| Yes | Games | Get Games | GET | `/helix/games` | 2025-12-10 | |

## Goals

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Goals | Get Creator Goals | GET | `/helix/goals` | 2025-12-10 | |

## Guest Star

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Guest Star | Get Channel Guest Star Settings | GET | `/helix/channels/guest_star_settings` | 2025-12-10 | BETA |
| Yes | Guest Star | Update Channel Guest Star Settings | PUT | `/helix/channels/guest_star_settings` | 2025-12-10 | BETA |
| Yes | Guest Star | Get Guest Star Session | GET | `/helix/guest_star/session` | 2025-12-10 | BETA |
| Yes | Guest Star | Create Guest Star Session | POST | `/helix/guest_star/session` | 2025-12-10 | BETA |
| Yes | Guest Star | End Guest Star Session | DELETE | `/helix/guest_star/session` | 2025-12-10 | BETA |
| Yes | Guest Star | Get Guest Star Invites | GET | `/helix/guest_star/invites` | 2025-12-10 | BETA |
| Yes | Guest Star | Send Guest Star Invite | POST | `/helix/guest_star/invites` | 2025-12-10 | BETA |
| Yes | Guest Star | Delete Guest Star Invite | DELETE | `/helix/guest_star/invites` | 2025-12-10 | BETA |
| Yes | Guest Star | Assign Guest Star Slot | POST | `/helix/guest_star/slot` | 2025-12-10 | BETA |
| Yes | Guest Star | Update Guest Star Slot | PATCH | `/helix/guest_star/slot` | 2025-12-10 | BETA |
| Yes | Guest Star | Delete Guest Star Slot | DELETE | `/helix/guest_star/slot` | 2025-12-10 | BETA |
| Yes | Guest Star | Update Guest Star Slot Settings | PATCH | `/helix/guest_star/slot_settings` | 2025-12-10 | BETA |

## Hype Train

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Hype Train | Get Hype Train Events | GET | `/helix/hypetrain/events` | 2025-12-10 | DEPRECATED |
| Yes | Hype Train | Get Hype Train Status | GET | `/helix/hypetrain/status` | 2025-12-10 | |

## Moderation

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Moderation | Check AutoMod Status | POST | `/helix/moderation/enforcements/status` | 2025-12-10 | |
| Yes | Moderation | Manage Held AutoMod Messages | POST | `/helix/moderation/automod/message` | 2025-12-10 | |
| Yes | Moderation | Get AutoMod Settings | GET | `/helix/moderation/automod/settings` | 2025-12-10 | |
| Yes | Moderation | Update AutoMod Settings | PUT | `/helix/moderation/automod/settings` | 2025-12-10 | |
| Yes | Moderation | Get Banned Users | GET | `/helix/moderation/banned` | 2025-12-10 | |
| Yes | Moderation | Ban User | POST | `/helix/moderation/bans` | 2025-12-10 | |
| Yes | Moderation | Unban User | DELETE | `/helix/moderation/bans` | 2025-12-10 | |
| Yes | Moderation | Get Unban Requests | GET | `/helix/moderation/unban_requests` | 2025-12-10 | |
| Yes | Moderation | Resolve Unban Requests | PATCH | `/helix/moderation/unban_requests` | 2025-12-10 | |
| Yes | Moderation | Get Blocked Terms | GET | `/helix/moderation/blocked_terms` | 2025-12-10 | |
| Yes | Moderation | Add Blocked Term | POST | `/helix/moderation/blocked_terms` | 2025-12-10 | |
| Yes | Moderation | Remove Blocked Term | DELETE | `/helix/moderation/blocked_terms` | 2025-12-10 | |
| Yes | Moderation | Delete Chat Messages | DELETE | `/helix/moderation/chat` | 2025-12-10 | |
| Yes | Moderation | Get Moderated Channels | GET | `/helix/moderation/channels` | 2025-12-10 | |
| Yes | Moderation | Get Moderators | GET | `/helix/moderation/moderators` | 2025-12-10 | |
| Yes | Moderation | Add Channel Moderator | POST | `/helix/moderation/moderators` | 2025-12-10 | |
| Yes | Moderation | Remove Channel Moderator | DELETE | `/helix/moderation/moderators` | 2025-12-10 | |
| Yes | Moderation | Get VIPs | GET | `/helix/channels/vips` | 2025-12-10 | |
| Yes | Moderation | Add Channel VIP | POST | `/helix/channels/vips` | 2025-12-10 | |
| Yes | Moderation | Remove Channel VIP | DELETE | `/helix/channels/vips` | 2025-12-10 | |
| Yes | Moderation | Update Shield Mode Status | PUT | `/helix/moderation/shield_mode` | 2025-12-10 | |
| Yes | Moderation | Get Shield Mode Status | GET | `/helix/moderation/shield_mode` | 2025-12-10 | |
| Yes | Moderation | Warn Chat User | POST | `/helix/moderation/warnings` | 2025-12-10 | |

## Polls

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Polls | Get Polls | GET | `/helix/polls` | 2025-12-10 | |
| Yes | Polls | Create Poll | POST | `/helix/polls` | 2025-12-10 | |
| Yes | Polls | End Poll | PATCH | `/helix/polls` | 2025-12-10 | |

## Predictions

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Predictions | Get Predictions | GET | `/helix/predictions` | 2025-12-10 | |
| Yes | Predictions | Create Prediction | POST | `/helix/predictions` | 2025-12-10 | |
| Yes | Predictions | End Prediction | PATCH | `/helix/predictions` | 2025-12-10 | |

## Raids

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Raids | Start a Raid | POST | `/helix/raids` | 2025-12-10 | |
| Yes | Raids | Cancel a Raid | DELETE | `/helix/raids` | 2025-12-10 | |

## Schedule

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Schedule | Get Channel Stream Schedule | GET | `/helix/schedule` | 2025-12-10 | |
| Yes | Schedule | Get Channel iCalendar | GET | `/helix/schedule/icalendar` | 2025-12-10 | |
| Yes | Schedule | Update Channel Stream Schedule | PATCH | `/helix/schedule/settings` | 2025-12-10 | |
| Yes | Schedule | Create Channel Stream Schedule Segment | POST | `/helix/schedule/segment` | 2025-12-10 | |
| Yes | Schedule | Update Channel Stream Schedule Segment | PATCH | `/helix/schedule/segment` | 2025-12-10 | |
| Yes | Schedule | Delete Channel Stream Schedule Segment | DELETE | `/helix/schedule/segment` | 2025-12-10 | |

## Search

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Search | Search Categories | GET | `/helix/search/categories` | 2025-12-10 | |
| Yes | Search | Search Channels | GET | `/helix/search/channels` | 2025-12-10 | |

## Streams

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Streams | Get Stream Key | GET | `/helix/streams/key` | 2025-12-10 | |
| Yes | Streams | Get Streams | GET | `/helix/streams` | 2025-12-10 | |
| Yes | Streams | Get Followed Streams | GET | `/helix/streams/followed` | 2025-12-10 | |
| Yes | Streams | Create Stream Marker | POST | `/helix/streams/markers` | 2025-12-10 | |
| Yes | Streams | Get Stream Markers | GET | `/helix/streams/markers` | 2025-12-10 | |

## Subscriptions

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Subscriptions | Get Broadcaster Subscriptions | GET | `/helix/subscriptions` | 2025-12-10 | |
| Yes | Subscriptions | Check User Subscription | GET | `/helix/subscriptions/user` | 2025-12-10 | |

## Tags

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| No | Tags | Get All Stream Tags | GET | `/helix/tags/streams` | - | DEPRECATED |
| No | Tags | Get Stream Tags | GET | `/helix/streams/tags` | - | DEPRECATED |

## Teams

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Teams | Get Channel Teams | GET | `/helix/teams/channel` | 2025-12-10 | |
| Yes | Teams | Get Teams | GET | `/helix/teams` | 2025-12-10 | |

## Users

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Users | Get Users | GET | `/helix/users` | 2025-12-10 | |
| Yes | Users | Update User | PUT | `/helix/users` | 2025-12-10 | |
| Yes | Users | Get Authorization By User | GET | `/helix/users/authorization` | 2025-12-11 | NEW |
| Yes | Users | Get User Block List | GET | `/helix/users/blocks` | 2025-12-10 | |
| Yes | Users | Block User | POST | `/helix/users/blocks` | 2025-12-10 | |
| Yes | Users | Unblock User | DELETE | `/helix/users/blocks` | 2025-12-10 | |
| Yes | Users | Get User Extensions | GET | `/helix/users/extensions/list` | 2025-12-10 | |
| Yes | Users | Get User Active Extensions | GET | `/helix/users/extensions` | 2025-12-10 | |
| Yes | Users | Update User Extensions | PUT | `/helix/users/extensions` | 2025-12-10 | |

## Videos

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Videos | Get Videos | GET | `/helix/videos` | 2025-12-10 | |
| Yes | Videos | Delete Videos | DELETE | `/helix/videos` | 2025-12-10 | |

## Whispers

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Whispers | Send Whisper | POST | `/helix/whispers` | 2025-12-10 | |

## Video Broadcast (Ingest)

| Supported | Category | Endpoint Name | Method | Path | Last Updated | Twitch Note |
|:---------:|----------|---------------|--------|------|--------------|-------------|
| Yes | Ingest | Get Ingest Servers | GET | `https://ingest.twitch.tv/ingests` | 2025-12-10 | Different base URL |

---

## Summary

| Status | Count |
|--------|-------|
| Supported | 148 |
| Not Yet Supported | 2 |
| **Total Endpoints** | **150** |

### Support by Category

| Category | Supported | Total |
|----------|-----------|-------|
| Authentication | 8 | 8 |
| Ads | 3 | 3 |
| Analytics | 2 | 2 |
| Bits | 3 | 3 |
| Channel Points | 6 | 6 |
| Channels | 5 | 5 |
| Charity | 2 | 2 |
| Chat | 15 | 15 |
| Clips | 4 | 4 |
| Conduits | 6 | 6 |
| Content Classification | 1 | 1 |
| Entitlements | 2 | 2 |
| EventSub | 3 | 3 |
| Extensions | 12 | 12 |
| Games | 2 | 2 |
| Goals | 1 | 1 |
| Guest Star | 12 | 12 |
| Hype Train | 2 | 2 |
| Ingest | 1 | 1 |
| Moderation | 22 | 22 |
| Polls | 3 | 3 |
| Predictions | 3 | 3 |
| Raids | 2 | 2 |
| Schedule | 6 | 6 |
| Search | 2 | 2 |
| Streams | 5 | 5 |
| Subscriptions | 2 | 2 |
| Tags | 0 | 2 |
| Teams | 2 | 2 |
| Users | 9 | 9 |
| Videos | 2 | 2 |
| Whispers | 1 | 1 |

### Notes

- **Tags endpoints** are not implemented as they are deprecated by Twitch
- **Guest Star endpoints** are marked as BETA by Twitch and may change
- **Ingest API** uses a different base URL (`https://ingest.twitch.tv`) and does not require authentication
- **Authentication/OIDC endpoints** use a different base URL (`https://id.twitch.tv/oauth2`) and support OpenID Connect flows
- **EventSub** includes webhook handler with signature verification, challenge response, and event type definitions
- **Extension JWT** authentication is supported for extension backend services

### Real-Time Features (Non-REST)

In addition to REST API endpoints, this library supports:

| Feature | Protocol | Description |
|---------|----------|-------------|
| IRC/TMI Chat | WebSocket | Full chat client with message parsing, subs, raids, moderation |
| EventSub WebSocket | WebSocket | Real-time event streaming with auto-reconnect |
| PubSub Compatibility | WebSocket | PubSub-style API backed by EventSub (migration layer) |

The **PubSub Compatibility** layer provides familiar `Listen(topic)`/`Unlisten(topic)` semantics for developers migrating from the deprecated Twitch PubSub system (decommissioned April 2025). It internally uses EventSub WebSocket.
