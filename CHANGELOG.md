# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [0.4.0] - 2025-12-23

### Added
- `CreateClipFromVOD` - Create clips from existing VODs (POST /videos/clips)
  - Requires `editor:manage:clips` or `channel:manage:clips` scope
  - Supports custom duration (5-60 seconds)
- IRC/TMI chat client for building chat bots
  - WebSocket-based connection to Twitch IRC
  - Message parsing for PRIVMSG, USERNOTICE, ROOMSTATE, etc.
  - Event handlers for messages, subscriptions, raids, and moderation
  - Support for sending messages, whispers, and chat commands

### Changed

### Fixed

## [0.3.1] - 2025-12-11

### Fixed
- Fixed errcheck lint issues (unchecked error returns) across all source and test files
- Fixed staticcheck lint issues (unnecessary embedded field selectors) in auth.go and eventsub.go

## [0.3.0] - 2025-12-11

### Added
- Caching layer
- Batch request support
- Webhook signature validation

## [0.2.0] - 2025-12-09

### Added
- WebSocket support for EventSub
- Retry logic with exponential backoff
- Request middleware support

## [0.1.0] - 2025-12-08

### Added

#### Authentication (`auth` package)
- Full OAuth 2.0 support for all Twitch authentication flows:
  - Implicit Grant Flow
  - Authorization Code Grant Flow
  - Client Credentials Grant Flow
  - Device Code Grant Flow
- Token management:
  - Token validation via `/validate` endpoint
  - Token refresh with automatic retry
  - Token revocation
  - Automatic token refresh background goroutine
- Comprehensive scope constants for all Twitch API scopes
- Pre-defined common scope combinations (Chat, Bot, Moderation, Channel, Broadcaster)
- Thread-safe token storage

#### Helix API Client (`helix` package)
- Full Helix API client with 100+ endpoints

**Ads**
- Start Commercial
- Get Ad Schedule
- Snooze Next Ad

**Analytics**
- Get Extension Analytics
- Get Game Analytics

**Bits**
- Get Bits Leaderboard
- Get Cheermotes

**Channel Points**
- Create/Get/Update/Delete Custom Rewards
- Get/Update Custom Reward Redemptions

**Channels**
- Get/Modify Channel Information
- Get Channel Editors
- Get Followed Channels
- Get Channel Followers
- Get/Add/Remove VIPs

**Chat**
- Get Chatters
- Get Channel/Global Emotes
- Get Emote Sets
- Get Channel/Global Chat Badges
- Get/Update Chat Settings
- Send Chat Announcement
- Send Shoutout
- Get/Update User Chat Color
- Send Chat Message

**Clips**
- Create Clip
- Get Clips

**EventSub**
- Create/Get/Delete EventSub Subscriptions

**Games**
- Get Games
- Get Top Games

**Goals**
- Get Creator Goals

**Hype Train**
- Get Hype Train Events

**Moderation**
- Get Banned Users
- Ban/Unban User
- Get/Add/Remove Moderators
- Delete Chat Messages
- Get/Add/Remove Blocked Terms
- Get/Update Shield Mode Status
- Warn Chat User

**Polls**
- Get/Create/End Polls

**Predictions**
- Get/Create/End Predictions

**Raids**
- Start/Cancel Raid

**Schedule**
- Get/Update Channel Stream Schedule
- Get Channel iCalendar
- Create/Update/Delete Schedule Segments

**Search**
- Search Categories
- Search Channels

**Streams**
- Get Streams
- Get Followed Streams
- Get Stream Key
- Create/Get Stream Markers

**Subscriptions**
- Get Broadcaster Subscriptions
- Check User Subscription

**Teams**
- Get Channel Teams
- Get Teams

**Users**
- Get Users
- Update User
- Get/Block/Unblock Users
- Get User Extensions

**Videos**
- Get/Delete Videos

**Whispers**
- Send Whisper

#### Features
- Generic response types with Go generics
- Automatic rate limit tracking from response headers
- Customizable HTTP client
- Configurable base URL for testing
- Pagination helpers

#### Testing
- Comprehensive unit tests for auth package
- Comprehensive unit tests for helix package
- Mock server support for testing

#### Documentation
- Full README with quick start guide
- Detailed documentation in `docs/` folder
- Usage examples for common scenarios
- Endpoint support matrix

### Security
- Secure token handling with mutex protection
- CSRF state parameter support
- Scope validation

---
