Get up and running with the Twitch Helix API wrapper.

## Installation

```bash
go get github.com/Its-donkey/helix
```

## Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    // Create auth client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Get app access token
    ctx := context.Background()
    token, err := authClient.GetAppAccessToken(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Got token: %s\n", token.AccessToken)

    // Create Helix client
    client := helix.NewClient("your-client-id", authClient)

    // Get user information
    users, err := client.GetUsers(ctx, &helix.GetUsersParams{
        Logins: []string{"twitchdev"},
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, user := range users.Data {
        fmt.Printf("User: %s (ID: %s)\n", user.DisplayName, user.ID)
    }
}
```

## Authentication

The library supports all four Twitch OAuth flows:

### 1. Client Credentials Flow (App Access Token)

Best for server-to-server requests that don't require user permission.

```go
authClient := helix.NewAuthClient(helix.AuthConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
})

token, err := authClient.GetAppAccessToken(ctx)
```

### 2. Authorization Code Flow

Best for server-based apps that can securely store client secrets.

```go
authClient := helix.NewAuthClient(helix.AuthConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURI:  "http://localhost:3000/callback",
    Scopes:       helix.CommonScopes.Bot,
    State:        "random-state-string",
})

// Step 1: Get authorization URL
url, _ := authClient.GetCodeAuthURL()
// Redirect user to url...

// Step 2: Exchange code for token (after callback)
token, err := authClient.ExchangeCode(ctx, "authorization-code")
```

### 3. Implicit Grant Flow

For client-side apps without servers.

```go
authClient := helix.NewAuthClient(helix.AuthConfig{
    ClientID:    "your-client-id",
    RedirectURI: "http://localhost:3000/callback",
    Scopes:      []string{"chat:read", "chat:edit"},
})

url, _ := authClient.GetImplicitAuthURL()
// Redirect user to url...
// Token is returned in URL fragment
```

### 4. Device Code Flow

For limited-input devices.

```go
authClient := helix.NewAuthClient(helix.AuthConfig{
    ClientID: "your-client-id",
    Scopes:   []string{"user:read:email"},
})

// Get device code
deviceCode, err := authClient.GetDeviceCode(ctx)
fmt.Printf("Go to %s and enter code: %s\n",
    deviceCode.VerificationURI, deviceCode.UserCode)

// Wait for user authorization
token, err := authClient.WaitForDeviceToken(ctx, deviceCode)
```

### Token Management

```go
// Validate token
validation, err := authClient.ValidateToken(ctx, token.AccessToken)

// Refresh token
newToken, err := authClient.RefreshToken(ctx, token.RefreshToken)

// Revoke token
err := authClient.RevokeToken(ctx, token.AccessToken)

// Auto-refresh (starts background goroutine)
cancel := authClient.AutoRefresh(ctx)
defer cancel()
```

## Creating the Helix Client

```go
client := helix.NewClient("client-id", authClient,
    helix.WithHTTPClient(customHTTPClient),  // Optional
    helix.WithBaseURL("custom-url"),         // Optional (for testing)
)
```

## Error Handling

```go
users, err := client.GetUsers(ctx, nil)
if err != nil {
    if apiErr, ok := err.(*helix.APIError); ok {
        fmt.Printf("API Error %d: %s - %s\n",
            apiErr.StatusCode, apiErr.ErrorType, apiErr.Message)
    }
    return err
}
```

## Pagination

```go
var allUsers []helix.User
cursor := ""

for {
    resp, err := client.GetUsers(ctx, &helix.GetUsersParams{
        PaginationParams: &helix.PaginationParams{
            First: 100,
            After: cursor,
        },
    })
    if err != nil {
        return err
    }

    allUsers = append(allUsers, resp.Data...)

    if resp.Pagination == nil || resp.Pagination.Cursor == "" {
        break
    }
    cursor = resp.Pagination.Cursor
}
```

## Rate Limiting

```go
remaining, reset := client.GetRateLimitInfo()
fmt.Printf("Requests remaining: %d, resets at: %s\n", remaining, reset)
```

## Next Steps

- Browse the [Available Endpoints](Projects/Programming/Kappopher/Documents/README.md#available-endpoints) to see all API methods
- Check out [complete examples](./examples/) for working code samples
- Read the [Auth documentation](auth.md) for detailed authentication info