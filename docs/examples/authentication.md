---
layout: default
title: Authentication Examples
description: Complete examples for all OAuth 2.0 authentication flows supported by the Twitch API.
---

## Overview

Twitch uses OAuth 2.0 for authentication. The flow you choose depends on your application type:

| Flow | Use Case | User Interaction | Token Type |
|------|----------|------------------|------------|
| **Client Credentials** | Server-side apps, public data | None | App Access Token |
| **Authorization Code** | Web apps with backend | Browser redirect | User Access Token |
| **Device Code** | CLIs, TVs, IoT devices | Code on separate device | User Access Token |
| **Implicit Grant** | Single-page apps (SPAs) | Browser redirect | User Access Token |

**Key Concepts**:
- **App Access Tokens**: Identify your application, used for public API endpoints
- **User Access Tokens**: Identify a specific user, required for user-specific data and actions
- **Scopes**: Define what permissions your app requests (e.g., `chat:read`, `channel:manage:redemptions`)
- **Refresh Tokens**: Used to get new access tokens without re-authenticating the user

## Client Credentials Flow (App Access Token)

Use this flow for server-to-server API calls that don't require user authorization. This is the simplest flow - your server exchanges its client credentials directly for an access token.

**When to use**: Fetching public data like user profiles, live streams, or game information.

**Limitations**: Cannot access user-specific data or perform actions on behalf of a user.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Get app access token
    token, err := authClient.GetAppAccessToken(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Access Token: %s\n", token.AccessToken)
    fmt.Printf("Expires In: %d seconds\n", token.ExpiresIn)
    fmt.Printf("Token Type: %s\n", token.TokenType)

    // Create Helix client with app token
    client := helix.NewClient("your-client-id", authClient)

    // Now make API calls (limited to endpoints that accept app tokens)
    users, _ := client.GetUsers(ctx, &helix.GetUsersParams{
        Logins: []string{"twitchdev"},
    })
    fmt.Printf("User: %s\n", users.Data[0].DisplayName)
}
```

## Authorization Code Flow (User Access Token)

The most common flow for web applications. The user is redirected to Twitch to authorize your app, then redirected back with an authorization code that you exchange for tokens.

**How it works**:
1. Generate an authorization URL with requested scopes and a CSRF state token
2. Redirect user to Twitch to authorize your application
3. User approves, Twitch redirects back with an authorization code
4. Your server exchanges the code for access and refresh tokens

**Security**: Always validate the `state` parameter to prevent CSRF attacks. The state should be a cryptographically random string stored in the user's session.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURI:  "http://localhost:3000/callback",
        Scopes: []string{
            helix.ScopeUserReadEmail,
            helix.ScopeChatRead,
            helix.ScopeChatEdit,
        },
    })

    // Generate authorization URL with CSRF state
    state := "random-csrf-token"
    url := authClient.GetCodeAuthURLWithState(state)
    fmt.Printf("Open this URL in your browser:\n%s\n\n", url)

    // Handle the OAuth callback
    var authCode string
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        // Verify state to prevent CSRF
        if r.URL.Query().Get("state") != state {
            http.Error(w, "Invalid state", http.StatusBadRequest)
            return
        }

        authCode = r.URL.Query().Get("code")
        fmt.Fprintf(w, "Authorization successful! You can close this window.")

        // Exchange code for token
        token, err := authClient.ExchangeCode(ctx, authCode)
        if err != nil {
            log.Printf("Token exchange failed: %v", err)
            return
        }

        fmt.Printf("\nAccess Token: %s\n", token.AccessToken)
        fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
        fmt.Printf("Scopes: %v\n", token.Scope)
    })

    log.Fatal(http.ListenAndServe(":3000", nil))
}
```

## Device Code Flow

Designed for devices that don't have a browser or have limited input capabilities (smart TVs, game consoles, CLI tools). The user authorizes on a separate device (like their phone) by entering a code.

**How it works**:
1. Request a device code and user code from Twitch
2. Display the user code and verification URL to the user
3. User visits the URL on another device and enters the code
4. Your app polls for the token until authorization completes or times out

**User experience**: The user sees something like "Go to twitch.tv/activate and enter code: ABCD-1234"

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID: "your-client-id",
        Scopes: []string{
            helix.ScopeUserReadEmail,
            helix.ScopeChatRead,
        },
    })

    // Request device code
    deviceCode, err := authClient.GetDeviceCode(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Go to: %s\n", deviceCode.VerificationURI)
    fmt.Printf("Enter code: %s\n", deviceCode.UserCode)
    fmt.Printf("Expires in: %d seconds\n", deviceCode.ExpiresIn)

    // Poll for token (blocks until user authorizes or timeout)
    token, err := authClient.WaitForDeviceToken(ctx, deviceCode)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\nAuthorization successful!\n")
    fmt.Printf("Access Token: %s\n", token.AccessToken)
}
```

## Implicit Grant Flow

For client-side JavaScript applications (SPAs) where the client secret cannot be kept confidential. The access token is returned directly in the URL fragment.

**Important limitations**:
- No refresh tokens - user must re-authorize when token expires
- Token is exposed in browser history and potentially to JavaScript on the page
- Consider using Authorization Code Flow with PKCE for better security

**When to use**: Legacy SPAs or simple browser extensions where you can't use a backend server.

```go
package main

import (
    "fmt"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:    "your-client-id",
        RedirectURI: "http://localhost:3000/callback",
        Scopes: []string{
            helix.ScopeUserReadEmail,
        },
    })

    // Generate implicit auth URL
    // Token will be in the URL fragment after redirect
    state := "random-csrf-token"
    url := authClient.GetImplicitAuthURLWithState(state)
    fmt.Printf("Redirect user to:\n%s\n", url)

    // After redirect, extract token from URL fragment:
    // http://localhost:3000/callback#access_token=xxx&token_type=bearer&scope=user:read:email
}
```

## Token Refresh

User access tokens expire (typically after 4 hours). Use the refresh token to get a new access token without requiring the user to re-authorize.

**Important**: Refresh tokens can also expire or be revoked. Always handle refresh failures by falling back to re-authentication.

**Best practice**: Store refresh tokens securely (encrypted in database) and update them after each refresh, as Twitch may issue new refresh tokens.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Refresh an expired token
    refreshToken := "stored-refresh-token"
    newToken, err := authClient.RefreshToken(ctx, refreshToken)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("New Access Token: %s\n", newToken.AccessToken)
    fmt.Printf("New Refresh Token: %s\n", newToken.RefreshToken)

    // Store the new refresh token for future use
}
```

## Token Validation

Verify that a token is still valid and retrieve information about it. This is useful for checking token status before making API calls, or for getting the user ID associated with a token.

**Returns**: Client ID, user ID (for user tokens), login name, granted scopes, and time until expiration.

**When to use**: On app startup to check stored tokens, or periodically to ensure tokens haven't been revoked.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID: "your-client-id",
    })

    token := "access-token-to-validate"
    validation, err := authClient.ValidateToken(ctx, token)
    if err != nil {
        log.Printf("Token is invalid: %v", err)
        return
    }

    fmt.Printf("Client ID: %s\n", validation.ClientID)
    fmt.Printf("User ID: %s\n", validation.UserID)
    fmt.Printf("Login: %s\n", validation.Login)
    fmt.Printf("Scopes: %v\n", validation.Scopes)
    fmt.Printf("Expires In: %d seconds\n", validation.ExpiresIn)
}
```

## Token Revocation

Invalidate a token so it can no longer be used. This is important for security when users log out of your application.

**When to use**: User logout, account deletion, or when you detect suspicious activity.

**Note**: Revoking an access token also invalidates its associated refresh token.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID: "your-client-id",
    })

    token := "token-to-revoke"
    err := authClient.RevokeToken(ctx, token)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Token revoked successfully")
}
```

## Automatic Token Refresh

For long-running applications, you can set up automatic token refresh to ensure your tokens never expire during operation. The library will refresh tokens before they expire and notify you via a callback.

**How it works**: The auto-refresh mechanism monitors token expiration and refreshes proactively (before expiry). The callback lets you persist the new tokens.

**Use case**: Bots, dashboards, or any service that runs continuously.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Set initial token
    authClient.SetToken(&helix.TokenResponse{
        AccessToken:  "current-access-token",
        RefreshToken: "current-refresh-token",
        ExpiresIn:    3600,
    })

    // Start automatic refresh (refreshes when token is near expiry)
    stopRefresh := authClient.StartAutoRefresh(ctx, func(token *helix.TokenResponse, err error) {
        if err != nil {
            log.Printf("Auto-refresh failed: %v", err)
            return
        }
        fmt.Printf("Token auto-refreshed at %s\n", time.Now().Format(time.RFC3339))
        // Store new token in database
    })

    // Stop auto-refresh when done
    defer stopRefresh()

    // Your application runs here...
    select {}
}
```

## OpenID Connect (OIDC)

OIDC extends OAuth 2.0 to provide identity verification. When you request the `openid` scope, you can retrieve verified user information like email address and profile picture.

**When to use**: User registration/login where you need verified identity information rather than just authorization.

**Required scope**: `openid` (optionally with `user:read:email` for email access)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURI:  "http://localhost:3000/callback",
        Scopes: []string{
            helix.ScopeOpenID,
            helix.ScopeUserReadEmail,
        },
    })

    // After getting an access token with openid scope...
    // Get user info from OIDC endpoint
    userInfo, err := authClient.GetOIDCUserInfo(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Subject (User ID): %s\n", userInfo.Sub)
    fmt.Printf("Preferred Username: %s\n", userInfo.PreferredUsername)
    fmt.Printf("Email: %s\n", userInfo.Email)
    fmt.Printf("Email Verified: %v\n", userInfo.EmailVerified)
    fmt.Printf("Picture: %s\n", userInfo.Picture)
}
```

## Common Scope Combinations

The library provides pre-defined scope combinations for common use cases. These help you request the right permissions without having to look up individual scope names.

**Principle of least privilege**: Only request the scopes you actually need. Users are more likely to authorize apps that request minimal permissions.

```go
// Chat bot scopes
helix.CommonScopes.Chat // chat:read, chat:edit

// Full bot functionality
helix.CommonScopes.Bot // chat scopes + whispers, user read

// Moderation tools
helix.CommonScopes.Moderation // ban, manage messages, automod, etc.

// Channel management
helix.CommonScopes.Channel // channel:manage:*, channel:read:*

// Full broadcaster access
helix.CommonScopes.Broadcaster // all broadcaster-level scopes
```

