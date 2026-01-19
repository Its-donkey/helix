---
layout: default
title: Authentication API
description: OAuth 2.0 authentication client for Twitch, supporting multiple authorization flows and OIDC.
---

## Setup

Create an AuthClient with your application credentials:

```go
auth := helix.NewAuthClient(helix.AuthConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURI:  "http://localhost:3000/callback",
    Scopes:       []string{helix.ScopeChatRead, helix.ScopeChatEdit},
    State:        "random-state-string",
    ForceVerify:  false,
})
```

## Authorization Code Flow

The most common flow for web applications where users authorize your app.

### GetCodeAuthURL

Generate the URL to redirect users to for authorization.

```go
authURL, err := auth.GetCodeAuthURL()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Visit:", authURL)
// Redirect user to authURL
```

**Returns:** Authorization URL string

### ExchangeCode

Exchange the authorization code (from callback) for an access token.

```go
// After user is redirected back with ?code=xxx
token, err := auth.ExchangeCode(ctx, "authorization-code-from-callback")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Access Token: %s\n", token.AccessToken)
fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
fmt.Printf("Expires In: %d seconds\n", token.ExpiresIn)
```

**Parameters:**
- `code` (string): The authorization code from the callback URL

**Sample Response:**
```json
{
  "access_token": "cfabdegwdoklmawdzdo98xt2fo512y",
  "refresh_token": "b23c4a5b6789d0e1f2a3b4c5d6e7f8a9",
  "token_type": "bearer",
  "expires_in": 14400,
  "scope": ["chat:read", "chat:edit"]
}
```

## Implicit Grant Flow

For client-side applications where the token is returned directly.

### GetImplicitAuthURL

Generate the URL for implicit grant flow (token returned in URL fragment).

```go
authURL, err := auth.GetImplicitAuthURL()
if err != nil {
    log.Fatal(err)
}
// Token will be in URL fragment: #access_token=xxx&token_type=bearer
```

**Returns:** Authorization URL string

## Client Credentials Flow

For server-to-server authentication (app access tokens).

### GetAppAccessToken

Obtain an app access token for server-to-server API calls.

```go
token, err := auth.GetAppAccessToken(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("App Access Token: %s\n", token.AccessToken)
```

**Sample Response:**
```json
{
  "access_token": "jostpf5q0uzmxmkba9iyug38kjtgh",
  "token_type": "bearer",
  "expires_in": 5011271
}
```

## Device Code Flow

For devices with limited input capabilities (TVs, game consoles, CLI tools).

### GetDeviceCode

Initiate the device authorization flow.

```go
deviceCode, err := auth.GetDeviceCode(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Go to: %s\n", deviceCode.VerificationURI)
fmt.Printf("Enter code: %s\n", deviceCode.UserCode)
```

**Sample Response:**
```json
{
  "device_code": "d3f2a1b0c9e8d7f6a5b4c3d2e1f0a9b8",
  "expires_in": 1800,
  "interval": 5,
  "user_code": "ABCD-1234",
  "verification_uri": "https://www.twitch.tv/activate"
}
```

### WaitForDeviceToken

Wait for the user to authorize and retrieve the token.

```go
token, err := auth.WaitForDeviceToken(ctx, deviceCode)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Access Token: %s\n", token.AccessToken)
```

**Sample Response:**
```json
{
  "access_token": "cfabdegwdoklmawdzdo98xt2fo512y",
  "refresh_token": "b23c4a5b6789d0e1f2a3b4c5d6e7f8a9",
  "token_type": "bearer",
  "expires_in": 14400,
  "scope": ["chat:read", "chat:edit"]
}
```

### PollDeviceToken

Manually poll for device token (used internally by WaitForDeviceToken).

```go
token, err := auth.PollDeviceToken(ctx, deviceCode.DeviceCode)
if err == helix.ErrAuthorizationPending {
    // User hasn't authorized yet, wait and try again
}
```

## Token Management

### ValidateToken

Validate an access token and get information about it.

```go
validation, err := auth.ValidateToken(ctx, "access-token")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User ID: %s\n", validation.UserID)
fmt.Printf("Login: %s\n", validation.Login)
fmt.Printf("Scopes: %v\n", validation.Scopes)
fmt.Printf("Expires In: %d seconds\n", validation.ExpiresIn)
```

**Sample Response:**
```json
{
  "client_id": "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
  "login": "twitchdev",
  "scopes": ["chat:read", "chat:edit"],
  "user_id": "141981764",
  "expires_in": 5520838
}
```

### ValidateCurrentToken

Validate the currently stored token.

```go
validation, err := auth.ValidateCurrentToken(ctx)
```

### RefreshToken

Refresh an access token using a refresh token.

```go
newToken, err := auth.RefreshToken(ctx, "refresh-token")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("New Access Token: %s\n", newToken.AccessToken)
```

**Sample Response:**
```json
{
  "access_token": "1ssjqsqfy6bads1ws7m03gras79zfr",
  "refresh_token": "eyJfMzUtNDU0OC4MWYwLTQ5MDY5ODY4NGNlMSJ9%asdfasdf=",
  "token_type": "bearer",
  "expires_in": 14400,
  "scope": ["chat:read", "chat:edit"]
}
```

### RefreshCurrentToken

Refresh the currently stored token.

```go
newToken, err := auth.RefreshCurrentToken(ctx)
```

### RevokeToken

Revoke an access token.

```go
err := auth.RevokeToken(ctx, "access-token")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Token revoked successfully")
```

**Sample Response:**

Returns HTTP 200 OK with no body on success.

### RevokeCurrentToken

Revoke the currently stored token.

```go
err := auth.RevokeCurrentToken(ctx)
```

### AutoRefresh

Start automatic token refresh in the background.

```go
cancel := auth.AutoRefresh(ctx)
defer cancel() // Stop auto-refresh when done

// Token will be automatically refreshed 5 minutes before expiry
```

## Token Helpers

### Token.IsExpired

Check if the token has expired.

```go
if token.IsExpired() {
    // Refresh the token
}
```

### Token.Valid

Check if the token is non-empty and not expired.

```go
if token.Valid() {
    // Use the token
}
```

### SetToken / GetToken

Manually manage the stored token.

```go
// Store a token
auth.SetToken(&helix.Token{
    AccessToken:  "your-token",
    RefreshToken: "your-refresh-token",
    ExpiresIn:    14400,
})

// Retrieve the stored token
token := auth.GetToken()
```

## OIDC (OpenID Connect)

Support for Twitch's OIDC implementation for identity verification.

### GetOpenIDConfiguration

Fetch the OIDC discovery document.

```go
config, err := auth.GetOpenIDConfiguration(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Issuer: %s\n", config.Issuer)
fmt.Printf("Token Endpoint: %s\n", config.TokenEndpoint)
```

**Sample Response:**
```json
{
  "issuer": "https://id.twitch.tv/oauth2",
  "authorization_endpoint": "https://id.twitch.tv/oauth2/authorize",
  "token_endpoint": "https://id.twitch.tv/oauth2/token",
  "userinfo_endpoint": "https://id.twitch.tv/oauth2/userinfo",
  "jwks_uri": "https://id.twitch.tv/oauth2/keys",
  "response_types_supported": ["code", "token", "id_token", "code id_token", "token id_token"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid"],
  "token_endpoint_auth_methods_supported": ["client_secret_post"],
  "claims_supported": ["iss", "sub", "aud", "exp", "iat", "email", "email_verified", "picture", "preferred_username", "updated_at"]
}
```

### GetOIDCAuthorizationURL

Generate an OIDC authorization URL.

```go
authURL, err := auth.GetOIDCAuthorizationURL(
    helix.ResponseTypeCodeIDToken,
    "random-nonce",
    nil, // Optional claims parameter
)
```

**Parameters:**
- `responseType` (OIDCResponseType): The response type (`ResponseTypeCode`, `ResponseTypeToken`, `ResponseTypeIDToken`, `ResponseTypeTokenIDToken`, `ResponseTypeCodeIDToken`)
- `nonce` (string): A random string to prevent replay attacks
- `claims` (map[string]interface{}): Optional claims to request

### ExchangeCodeForOIDCToken

Exchange an authorization code for an OIDC token (includes ID token).

```go
oidcToken, err := auth.ExchangeCodeForOIDCToken(ctx, "authorization-code")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Access Token: %s\n", oidcToken.AccessToken)
fmt.Printf("ID Token: %s\n", oidcToken.IDToken)
```

**Sample Response:**
```json
{
  "access_token": "cfabdegwdoklmawdzdo98xt2fo512y",
  "refresh_token": "b23c4a5b6789d0e1f2a3b4c5d6e7f8a9",
  "token_type": "bearer",
  "expires_in": 14400,
  "scope": ["openid", "user:read:email"],
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjEifQ.eyJpc3MiOiJodHRwczovL2lkLnR3aXRjaC50di9vYXV0aDIiLCJzdWIiOiIxNDE5ODE3NjQiLCJhdWQiOiJ3Ym15dHI5M3h6dzh6YmcwcDFpenF5enpjNW1iaXoiLCJleHAiOjE2ODA0NTY3ODksImlhdCI6MTY4MDQ0MjM4OSwibm9uY2UiOiJyYW5kb20tbm9uY2UiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJ0d2l0Y2hkZXYiLCJlbWFpbCI6InR3aXRjaGRldkBleGFtcGxlLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlfQ.signature"
}
```

### GetOIDCUserInfo

Fetch user information from the OIDC UserInfo endpoint.

```go
userInfo, err := auth.GetOIDCUserInfo(ctx, "access-token")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User ID: %s\n", userInfo.Sub)
fmt.Printf("Username: %s\n", userInfo.PreferredUsername)
fmt.Printf("Email: %s\n", userInfo.Email)
```

**Sample Response:**
```json
{
  "sub": "141981764",
  "preferred_username": "TwitchDev",
  "email": "twitchdev@example.com",
  "email_verified": true,
  "picture": "https://static-cdn.jtvnw.net/jtv_user_pictures/8a6381c7-d0c0-4576-b179-38bd5ce1d6af-profile_image-300x300.png",
  "updated_at": 1680442389
}
```

### GetCurrentOIDCUserInfo

Fetch user info using the currently stored token.

```go
userInfo, err := auth.GetCurrentOIDCUserInfo(ctx)
```

### GetJWKS

Fetch the JSON Web Key Set for validating ID tokens.

```go
jwks, err := auth.GetJWKS(ctx)
if err != nil {
    log.Fatal(err)
}
key := jwks.GetKeyByID("1")
if key != nil {
    rsaKey, err := key.RSAPublicKey()
    // Use rsaKey to verify ID token signatures
}
```

**Sample Response:**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "e": "AQAB",
      "n": "6lq9MQ-q6hcxr7kOUp-tHlHtdcDsVLwVIw13iXUCvuDOeCi0VSuxCCUY6UmMjy53dX00ih2E4Y4UvlrmmurK0eG26b-HMNNAvCGsVXHU3RcRhVoHDaOwHwU72j7bpHn9XbP3Q3jebX6KIfNbei2MiR0Wyb8RZHE-aZhRYO8_-k9G2GycTpvc-2GBsP8VHLUKKfAs2B6sW3q3ymU6M0L-cFXkZ9fHkn9ejs-sqZPhMJxtBPBxoUIUQFTgv4VXTSv914f_YkNw-EjuwbgwXMvpyr06EyfImxHoxsZkFYB-qBYHtaMxTnFsZBr6fn8Ha2JqT1hoP7Z5r5wxDu3GQhKkHw",
      "kid": "1",
      "alg": "RS256",
      "use": "sig"
    }
  ]
}
```

### ParseIDToken

Parse an ID token without validating the signature.

```go
claims, err := helix.ParseIDToken(oidcToken.IDToken)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User ID: %s\n", claims.Sub)
fmt.Printf("Username: %s\n", claims.PreferredUsername)
fmt.Printf("Email: %s\n", claims.Email)
```

**ID Token Claims:**
```json
{
  "iss": "https://id.twitch.tv/oauth2",
  "sub": "141981764",
  "aud": "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
  "exp": 1680456789,
  "iat": 1680442389,
  "nonce": "random-nonce",
  "preferred_username": "TwitchDev",
  "email": "twitchdev@example.com",
  "email_verified": true,
  "picture": "https://static-cdn.jtvnw.net/jtv_user_pictures/profile_image-300x300.png",
  "updated_at": 1680442389
}
```

### ValidateIDTokenClaims

Validate the claims in an ID token.

```go
err := auth.ValidateIDTokenClaims(claims, "random-nonce")
if err != nil {
    log.Fatal("Invalid ID token:", err)
}
```

## OAuth Scopes

The library provides constants for all Twitch OAuth scopes.

### Scope Constants

```go
// Analytics
helix.ScopeAnalyticsReadExtensions  // "analytics:read:extensions"
helix.ScopeAnalyticsReadGames       // "analytics:read:games"

// Bits
helix.ScopeBitsRead                 // "bits:read"

// Channel
helix.ScopeChannelBot               // "channel:bot"
helix.ScopeChannelEditCommercial    // "channel:edit:commercial"
helix.ScopeChannelManageAds         // "channel:manage:ads"
helix.ScopeChannelManageBroadcast   // "channel:manage:broadcast"
helix.ScopeChannelManageModerators  // "channel:manage:moderators"
helix.ScopeChannelManagePolls       // "channel:manage:polls"
helix.ScopeChannelManagePredictions // "channel:manage:predictions"
helix.ScopeChannelManageRaids       // "channel:manage:raids"
helix.ScopeChannelManageRedemptions // "channel:manage:redemptions"
helix.ScopeChannelManageSchedule    // "channel:manage:schedule"
helix.ScopeChannelManageVideos      // "channel:manage:videos"
helix.ScopeChannelManageVIPs        // "channel:manage:vips"
helix.ScopeChannelReadSubscriptions // "channel:read:subscriptions"
// ... and many more

// Chat
helix.ScopeChatRead                 // "chat:read"
helix.ScopeChatEdit                 // "chat:edit"

// Moderation
helix.ScopeModerationRead           // "moderation:read"
helix.ScopeModeratorManageBannedUsers    // "moderator:manage:banned_users"
helix.ScopeModeratorManageChatMessages   // "moderator:manage:chat_messages"
// ... and many more

// User
helix.ScopeUserReadEmail            // "user:read:email"
helix.ScopeUserManageWhispers       // "user:manage:whispers"
helix.ScopeUserWriteChat            // "user:write:chat"
// ... and many more
```

### Common Scope Combinations

Pre-defined scope combinations for common use cases:

```go
// Chat bot scopes
auth := helix.NewAuthClient(helix.AuthConfig{
    Scopes: helix.CommonScopes.Chat,
    // Includes: chat:read, chat:edit, user:write:chat, user:read:chat
})

// Moderation scopes
auth := helix.NewAuthClient(helix.AuthConfig{
    Scopes: helix.CommonScopes.Moderation,
    // Includes: moderation:read, moderator:manage:banned_users, etc.
})

// Bot scopes (chat + bot permissions)
auth := helix.NewAuthClient(helix.AuthConfig{
    Scopes: helix.CommonScopes.Bot,
    // Includes: chat:read, chat:edit, channel:bot, user:bot, etc.
})

// Broadcaster scopes (comprehensive channel management)
auth := helix.NewAuthClient(helix.AuthConfig{
    Scopes: helix.CommonScopes.Broadcaster,
    // Includes: channel management, moderation, polls, predictions, etc.
})

// Analytics scopes
auth := helix.NewAuthClient(helix.AuthConfig{
    Scopes: helix.CommonScopes.Analytics,
    // Includes: analytics:read:extensions, analytics:read:games
})
```

## Error Handling

The library provides typed errors for common authentication failures:

```go
var (
    helix.ErrInvalidToken         // Invalid access token
    helix.ErrTokenExpired         // Token has expired
    helix.ErrAuthorizationPending // Device code flow: user hasn't authorized yet
    helix.ErrInvalidDeviceCode    // Invalid device code
    helix.ErrInvalidRefreshToken  // Invalid refresh token
    helix.ErrMissingClientID      // Client ID is required
    helix.ErrMissingClientSecret  // Client secret is required
    helix.ErrMissingRedirectURI   // Redirect URI is required
    helix.ErrMissingCode          // Authorization code is required
)
```

Example error handling:

```go
token, err := auth.RefreshToken(ctx, refreshToken)
if err != nil {
    switch err {
    case helix.ErrInvalidRefreshToken:
        // Refresh token is invalid, need to re-authenticate
    case helix.ErrMissingClientSecret:
        // Configuration error
    default:
        // Other error
    }
}
```

