---
layout: default
title: Cookbook
---

# Cookbook

Comprehensive code examples covering all kappopher features.

## Getting Started

- [Basic Usage](examples/basic.md) - Simple API calls, pagination, and error handling

## Authentication

- [Authentication](examples/authentication.md) - All OAuth 2.0 flows (client credentials, authorization code, device code, implicit, token refresh, OIDC)

## Chat & Real-Time

- [Chat Bot](examples/chatbot.md) - Building a chat bot with EventSub
- [IRC Client](examples/irc-client.md) - Low-level IRC/WebSocket chat client

## EventSub

- [EventSub Webhooks](examples/eventsub-webhooks.md) - Handle webhook notifications with signature verification
- [EventSub WebSocket](examples/eventsub-websocket.md) - Real-time event streaming without a public endpoint
- [PubSub Migration](examples/pubsub-migration.md) - PubSub-style API using EventSub

## API Usage

- [API Usage Examples](examples/api-usage.md) - Common API patterns for users, channels, streams, chat, moderation, polls, predictions, and clips
- [Channel Points](examples/channel-points.md) - Custom rewards, redemptions, real-time tracking
- [Bits & Subscriptions](examples/bits-subscriptions.md) - Bits leaderboard, cheermotes, subscriber management
- [Videos & Clips](examples/videos-clips.md) - VODs, highlights, clips, stream markers
- [Schedule & Goals](examples/schedule-goals.md) - Stream schedule management, creator goals
- [Raids & Hype Train](examples/raids-hypetrain.md) - Channel raids, hype train tracking
- [Analytics & Charity](examples/analytics-charity.md) - Extension/game analytics, charity campaigns, teams, guest star, conduits

## Advanced

- [Batch & Caching](examples/batch-caching.md) - Batch operations, caching, rate limiting, middleware
- [Extension JWT](examples/extension-jwt.md) - JWT authentication for Twitch Extensions

## Quick Reference

### By Feature

| Feature | Example |
|---------|---------|
| OAuth Authentication | [Authentication](examples/authentication.md) |
| Get Users/Channels | [Basic](examples/basic.md), [API Usage](examples/api-usage.md) |
| Send Chat Messages | [Chat Bot](examples/chatbot.md), [IRC Client](examples/irc-client.md) |
| Handle Chat Events | [IRC Client](examples/irc-client.md), [EventSub WebSocket](examples/eventsub-websocket.md) |
| Moderation | [Chat Bot](examples/chatbot.md), [API Usage](examples/api-usage.md) |
| Channel Points | [Channel Points](examples/channel-points.md) |
| Bits | [Bits & Subscriptions](examples/bits-subscriptions.md) |
| Subscriptions | [Bits & Subscriptions](examples/bits-subscriptions.md) |
| Stream Schedule | [Schedule & Goals](examples/schedule-goals.md) |
| Creator Goals | [Schedule & Goals](examples/schedule-goals.md) |
| Raids | [Raids & Hype Train](examples/raids-hypetrain.md) |
| Hype Train | [Raids & Hype Train](examples/raids-hypetrain.md) |
| Clips | [Videos & Clips](examples/videos-clips.md) |
| Videos/VODs | [Videos & Clips](examples/videos-clips.md) |
| Polls & Predictions | [API Usage](examples/api-usage.md) |
| EventSub WebSocket | [EventSub WebSocket](examples/eventsub-websocket.md) |
| EventSub Webhooks | [EventSub Webhooks](examples/eventsub-webhooks.md) |
| Batch Requests | [Batch & Caching](examples/batch-caching.md) |
| Caching | [Batch & Caching](examples/batch-caching.md) |
| Middleware | [Batch & Caching](examples/batch-caching.md) |
| Extensions | [Extension JWT](examples/extension-jwt.md) |
| Analytics | [Analytics & Charity](examples/analytics-charity.md) |
| Charity | [Analytics & Charity](examples/analytics-charity.md) |

### By Use Case

| Use Case | Start Here |
|----------|-----------|
| Build a chat bot | [Chat Bot](examples/chatbot.md) or [IRC Client](examples/irc-client.md) |
| Monitor stream events | [EventSub WebSocket](examples/eventsub-websocket.md) |
| Create a dashboard | [Batch & Caching](examples/batch-caching.md) |
| Manage channel points | [Channel Points](examples/channel-points.md) |
| Track subscriptions | [Bits & Subscriptions](examples/bits-subscriptions.md) |
| Server-side integration | [EventSub Webhooks](examples/eventsub-webhooks.md) |
| Build an extension | [Extension JWT](examples/extension-jwt.md) |
| Migrate from PubSub | [PubSub Migration](examples/pubsub-migration.md) |
