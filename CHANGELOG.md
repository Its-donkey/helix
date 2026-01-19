# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [1.0.2] - 2026-01-19 ([#45](https://github.com/Its-donkey/kappopher/pull/45))

### Added
- High-level EventSub WebSocket handlers: `WithEventSubRevocationHandler`, `WithEventSubReconnectHandler`, `WithEventSubErrorHandler`
- Automatic reconnection handling for EventSub WebSocket when Twitch sends reconnect message
- `CacheKeyWithContext()` function for cache key generation with base URL and token isolation
- `TokenHash()` function to generate hashed tokens for cache key isolation
- Body size limit (1MB) for EventSub webhook handlers to prevent memory exhaustion attacks
- Future timestamp rejection in EventSub webhook handlers (with 1-minute clock skew tolerance)
- `InvalidateCacheWithContext()` method for context-aware cache invalidation with token isolation

### Changed
- `MemoryCache.Get()` and `Set()` now copy byte slices to prevent mutation of cached data
- `MessageDeduplicator` now properly enforces `maxSize` by evicting oldest entries when at capacity
- `MessageDeduplicator` now treats `maxSize <= 0` as unlimited (previously broke deduplication)
- `SetExtensionRequiredConfigurationParams` now includes `ExtensionID` field as required by Twitch API
- Removed duplicate `irc/` package (functionality consolidated in `helix/` IRC implementation)

### Fixed
- IRC `Close()` now uses `sync.Once` to prevent panic from closing `stopChan` multiple times
- IRC `Close()` now waits for `readLoop()` to finish before returning
- IRC `Close()` now properly closes in-progress connections (not just fully connected ones)
- IRC `Connect()` is now serialized - concurrent calls return `ErrAlreadyConnected`
- IRC `waitForAuth()` now respects context deadline with a 30-second default timeout
- IRC handler panics are now recovered to prevent crashing the connection loop
- IRC messages now sanitize CR/LF characters to prevent injection attacks
- EventSub WebSocket `waitForWelcome()` now respects context deadline (uses whichever is sooner between context deadline and 10-second default)
- EventSub WebSocket `Close()` now properly handles in-progress connections
- EventSub WebSocket `Connect()` now closes existing connection before reconnecting
- `ChatBotClient.Connect()` now returns an error if no authentication token is available instead of panicking
- `extensionTokenProvider.GetToken()` is now thread-safe with proper mutex synchronization
- `PubSubClient.Connect()` no longer holds the lock when calling the `onConnect` handler, preventing deadlocks
- `PubSubClient.Close()` now returns combined errors using `errors.Join`
- `PubSubClient` goroutines are now tracked with `sync.WaitGroup` to prevent leaks
- `NewPubSubClient()` now validates that `helixClient` is not nil
- `NewEventSubWebSocket()` now validates that `helixClient` is not nil
- `WaitForDeviceToken()` now validates that `deviceCode` is not nil and `Interval` > 0
- `CreateToken()` now validates that `claims` is not nil
- `PubSubClient.Listen()` now handles empty response from `CreateEventSubSubscription` instead of panicking
- `PubSubClient.handleReconnect()` now copies `ws` under lock to prevent race conditions
- `EventSubWebSocket.handleReconnect()` now copies `ws` under lock to prevent race conditions
- IRC `Join`/`Part`/`Say`/`Reply` now sanitize channel names to prevent IRC command injection

## [1.0.1] - 2026-01-18 ([#37](https://github.com/Its-donkey/kappopher/pull/37))

### Added
- Separate `docs/hype-train.md` documentation for Hype Train API
- Documentation for `GetHypeTrainStatus` endpoint (previously undocumented)

### Changed
- Separated Hype Train API into its own file (`helix/hype_train.go`) to match Twitch API structure
- Updated Goals and Hype Train tests to use official Twitch API response samples
- Updated `docs/goals.md` to focus on Goals API only
- Updated `docs/README.md` to list Goals and Hype Train as separate entries

### Fixed
- WebSocket `Close()` and `Reconnect()` now set an immediate read deadline before closing to ensure `ReadMessage` unblocks quickly
- WebSocket error handler no longer reports expected close errors (connection closed, timeout during shutdown)
- IRC `Reconnect()` now properly triggers auto-reconnect when Twitch sends a RECONNECT command
- IRC reconnect loop now checks if auto-reconnect was disabled during the delay period

## [1.0.0] - 2026-01-15 ([#34](https://github.com/Its-donkey/kappopher/pull/34))

### Added
- Hype Train EventSub v2 support (v1 is deprecated by Twitch):
  - `Type` field indicating hype train type (`regular`, `golden_kappa`, `shared`)
  - `IsSharedTrain` flag for shared hype trains
  - `SharedTrainParticipants` list for multi-channel shared trains
  - `AllTimeHighLevel` and `AllTimeHighTotal` for channel records
- Automatic v1→v2 field conversion during JSON unmarshaling to ease migration from deprecated v1

### Changed
- Default Hype Train EventSub version changed from v1 to v2

### Deprecated
- `IsGoldenKappaTrain` field - use `Type == "golden_kappa"` instead
- `EventSubVersionHypeTrainV1` constant - v1 is deprecated by Twitch

### Fixed

## [0.4.0] - 2025-12-23 ([#30](https://github.com/Its-donkey/kappopher/pull/30))

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
