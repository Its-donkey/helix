# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in kappopher, please report it responsibly.

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Use [GitHub's private vulnerability reporting](https://github.com/Its-donkey/kappopher/security/advisories/new)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Resolution Timeline**: Depends on severity
  - Critical: 24-72 hours
  - High: 1-2 weeks
  - Medium: 2-4 weeks
  - Low: Next release cycle

### Scope

Security issues we're interested in:

- Authentication/authorization bypasses
- Token leakage or exposure
- Injection vulnerabilities (SQL, command, etc.)
- Cross-site scripting (XSS) in any web components
- Denial of service vulnerabilities
- Race conditions leading to security issues
- Cryptographic weaknesses

### Out of Scope

- Issues in dependencies (report to the upstream project)
- Social engineering attacks
- Physical attacks
- Issues requiring unlikely user interaction

## Security Best Practices

When using kappopher:

### Token Handling

```go
// DO: Use environment variables for secrets
clientID := os.Getenv("TWITCH_CLIENT_ID")
clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")

// DON'T: Hardcode credentials
clientID := "abc123" // Never do this
```

### Webhook Verification

Always verify EventSub webhook signatures:

```go
handler := helix.NewEventSubWebhookHandler(webhookSecret)
// The handler automatically verifies signatures
```

### Cache Isolation

When sharing caches across users, use context-aware cache keys:

```go
// Use CacheKeyWithContext for multi-tenant applications
key := helix.CacheKeyWithContext(baseURL, endpoint, query, helix.TokenHash(userToken))
```

### IRC Message Sanitization

The library automatically sanitizes IRC messages to prevent injection:

```go
// CR/LF characters are stripped from messages
client.Say(channel, userInput) // Safe - sanitized internally
```

## Security Features

This library includes several security features:

1. **Webhook Signature Verification**: All EventSub webhooks are verified using HMAC-SHA256
2. **Body Size Limits**: Webhook handlers limit request body size to 1MB
3. **Timestamp Validation**: Webhooks reject messages with future timestamps (with clock skew tolerance)
4. **Message Deduplication**: Prevents replay attacks on EventSub messages
5. **Token Hashing**: Cache keys use hashed tokens, never storing raw tokens
6. **IRC Injection Prevention**: CR/LF characters are sanitized from IRC messages
7. **Thread-Safe Token Storage**: All token operations use proper mutex synchronization

## Dependency Security

We use GitHub's Dependabot to monitor dependencies for known vulnerabilities. Security updates are prioritized and typically released within 48 hours of a CVE being published for a direct dependency.
