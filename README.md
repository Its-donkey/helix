# Twitch Helix API Client for Go

[![codecov](https://codecov.io/github/Its-donkey/helix/graph/badge.svg?token=UB85N281NS)](https://codecov.io/github/Its-donkey/helix)

A comprehensive Go client library for the Twitch Helix API with full OAuth authentication support.

## Features

- **Complete OAuth Support**: All four Twitch OAuth flows (Implicit, Authorization Code, Client Credentials, Device Code)
- **Token Management**: Automatic token refresh, validation, and revocation
- **Comprehensive API Coverage**: 147+ Helix API endpoints supported
- **EventSub Webhooks**: Built-in webhook handler with signature verification and event parsing
- **EventSub WebSocket**: Real-time event streaming with automatic keepalive and reconnection
- **Extension JWT**: Full support for Twitch Extension authentication
- **Type-Safe**: Fully typed request/response structures with Go generics
- **Pagination Support**: Built-in pagination helpers
- **Rate Limiting**: Automatic rate limit tracking with retry support
- **Testable**: Mock-friendly design with customizable HTTP client

## Installation

```bash
go get github.com/Its-donkey/helix
```

## Documentation

- [Quick Start Guide](./docs/quickstart.md) - Installation, authentication, and basic usage
- [API Reference](./docs/README.md) - Full endpoint documentation
- [Examples](./docs/examples/) - Working code samples

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.
