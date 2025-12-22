# Kappopher

[![Go Report Card](https://goreportcard.com/badge/github.com/Its-donkey/kappopher)](https://goreportcard.com/report/github.com/Its-donkey/kappopher)
[![codecov](https://codecov.io/github/Its-donkey/helix/graph/badge.svg?token=UB85N281NS)](https://codecov.io/github/Its-donkey/helix)

A comprehensive Twitch API toolkit for Go.

## Features

- **Complete OAuth Support**: All four Twitch OAuth flows (Implicit, Authorization Code, Client Credentials, Device Code)
- **Token Management**: Automatic token refresh, validation, and revocation
- **Comprehensive API Coverage**: 147+ Helix API endpoints supported
- **EventSub Webhooks**: Built-in webhook handler with signature verification and event parsing
- **EventSub WebSocket**: Real-time event streaming with automatic keepalive and reconnection
- **Extension JWT**: Full support for Twitch Extension authentication
- **Caching Layer**: Built-in response caching with TTL support
- **Middleware System**: Chainable request/response middleware
- **Batch Operations**: Concurrent batch request processing
- **Rate Limiting**: Automatic rate limit tracking with retry support
- **Type-Safe**: Fully typed request/response structures with Go generics

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

## Comparison

| Feature | Kappopher | Other Helix wrappers |
|---------|-----------|----------------|
| EventSub WebSocket | Yes | No |
| Middleware System | Yes | No |
| Caching Layer | Yes | No |
| Batch Operations | Yes | No |
| All OAuth Flows | Yes | Partial |
| Extension JWT | Full | Basic |

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.
