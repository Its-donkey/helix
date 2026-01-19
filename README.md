<p align="center">
  <img src="docs/assets/images/logo.png" alt="Kappopher" width="180">
</p>

<h1 align="center">Kappopher</h1>

<p align="center">
  <a href="https://pkg.go.dev/github.com/Its-donkey/kappopher"><img src="https://pkg.go.dev/badge/github.com/Its-donkey/kappopher.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/Its-donkey/kappopher"><img src="https://goreportcard.com/badge/github.com/Its-donkey/kappopher" alt="Go Report Card"></a>
  <a href="https://github.com/Its-donkey/kappopher/actions/workflows/test.yml"><img src="https://github.com/Its-donkey/kappopher/actions/workflows/test.yml/badge.svg" alt="Tests"></a>
  <a href="https://github.com/Its-donkey/kappopher/actions/workflows/test.yml"><img src="https://raw.githubusercontent.com/Its-donkey/kappopher/main/.github/coverage.svg" alt="Coverage"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/Its-donkey/kappopher"><img src="https://img.shields.io/github/go-mod/go-version/Its-donkey/kappopher" alt="Go Version"></a>
</p>

A comprehensive Twitch API toolkit for Go.

## Features

- **Complete OAuth Support**: All four Twitch OAuth flows (Implicit, Authorization Code, Client Credentials, Device Code)
- **Token Management**: Automatic token refresh, validation, and revocation
- **Comprehensive API Coverage**: 147+ Helix API endpoints supported
- **IRC/TMI Chat**: Full IRC chat client for building chat bots with message parsing, subs, raids, and moderation events
- **EventSub Webhooks**: Built-in webhook handler with signature verification and event parsing
- **EventSub WebSocket**: Real-time event streaming with automatic keepalive and reconnection
- **PubSub Compatibility**: Migration layer providing familiar PubSub-style API backed by EventSub
- **Extension JWT**: Full support for Twitch Extension authentication
- **Caching Layer**: Built-in response caching with TTL support
- **Middleware System**: Chainable request/response middleware
- **Batch Operations**: Concurrent batch request processing
- **Rate Limiting**: Automatic rate limit tracking with retry support
- **Type-Safe**: Fully typed request/response structures with Go generics

**[View Full Documentation and a complete cookbook of examples](https://its-donkey.github.io/kappopher/)**

## Installation

```bash
go get github.com/Its-donkey/kappopher
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    // Create auth client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Get app access token
    token, _ := authClient.GetAppAccessToken(context.Background())
    authClient.SetToken(token)

    // Create API client
    client := helix.NewClient("your-client-id", authClient)

    // Get user info
    resp, _ := client.GetUsers(context.Background(), &helix.GetUsersParams{
        Logins: []string{"shroud"},
    })

    fmt.Printf("User: %s (ID: %s)\n", resp.Data[0].DisplayName, resp.Data[0].ID)
}
```

## Documentation

- [Quick Start Guide](./docs/quickstart.md) - Installation, authentication, and basic usage
- [API Reference](./docs/README.md) - Full endpoint documentation

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.
