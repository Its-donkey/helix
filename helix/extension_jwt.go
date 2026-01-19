package helix

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ExtensionJWTRole defines the role in an extension JWT.
type ExtensionJWTRole string

const (
	// ExtensionRoleExternal is used for EBS (Extension Backend Service) signed tokens.
	ExtensionRoleExternal ExtensionJWTRole = "external"
	// ExtensionRoleBroadcaster is used when acting as the broadcaster.
	ExtensionRoleBroadcaster ExtensionJWTRole = "broadcaster"
	// ExtensionRoleModerator is used when acting as a moderator.
	ExtensionRoleModerator ExtensionJWTRole = "moderator"
	// ExtensionRoleViewer is used when acting as a viewer.
	ExtensionRoleViewer ExtensionJWTRole = "viewer"
)

// ExtensionJWTClaims represents the claims in a Twitch Extension JWT.
type ExtensionJWTClaims struct {
	// Required claims
	Exp    int64            `json:"exp"`     // Expiration time (Unix timestamp)
	UserID string           `json:"user_id"` // Twitch user ID
	Role   ExtensionJWTRole `json:"role"`    // User's role

	// Optional claims
	ChannelID    string `json:"channel_id,omitempty"`     // Channel ID for channel-specific operations
	OpaqueUserID string `json:"opaque_user_id,omitempty"` // Opaque user identifier
	IsUnlinked   bool   `json:"is_unlinked,omitempty"`    // Whether user has not shared identity

	// PubSub specific
	PubsubPermsListen []string `json:"pubsub_perms_listen,omitempty"` // Channels to listen to
	PubsubPermsSend   []string `json:"pubsub_perms_send,omitempty"`   // Channels to send to
}

// ExtensionJWT represents an extension JWT configuration and signer.
type ExtensionJWT struct {
	extensionID string
	secret      []byte
	ownerID     string // Extension owner's user ID
}

// NewExtensionJWT creates a new extension JWT signer.
// The secret should be the base64-encoded secret from the Extension Settings page.
func NewExtensionJWT(extensionID, base64Secret, ownerID string) (*ExtensionJWT, error) {
	secret, err := base64.StdEncoding.DecodeString(base64Secret)
	if err != nil {
		return nil, fmt.Errorf("decoding secret: %w", err)
	}

	return &ExtensionJWT{
		extensionID: extensionID,
		secret:      secret,
		ownerID:     ownerID,
	}, nil
}

// ExtensionID returns the extension ID.
func (e *ExtensionJWT) ExtensionID() string {
	return e.extensionID
}

// OwnerID returns the extension owner's user ID.
func (e *ExtensionJWT) OwnerID() string {
	return e.ownerID
}

// CreateToken creates a signed JWT with the given claims.
// Returns an error if claims is nil.
func (e *ExtensionJWT) CreateToken(claims *ExtensionJWTClaims) (string, error) {
	if claims == nil {
		return "", fmt.Errorf("claims cannot be nil")
	}

	// Set default expiration if not set (1 hour)
	if claims.Exp == 0 {
		claims.Exp = time.Now().Add(time.Hour).Unix()
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshaling header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshaling claims: %w", err)
	}

	// Encode header and claims
	headerB64 := base64URLEncode(headerJSON)
	claimsB64 := base64URLEncode(claimsJSON)

	// Create signature
	message := headerB64 + "." + claimsB64
	signature := e.sign([]byte(message))
	signatureB64 := base64URLEncode(signature)

	return message + "." + signatureB64, nil
}

// CreateEBSToken creates a token for Extension Backend Service operations.
// This is the most common token type for server-to-server extension API calls.
func (e *ExtensionJWT) CreateEBSToken(expiration time.Duration) (string, error) {
	claims := &ExtensionJWTClaims{
		Exp:    time.Now().Add(expiration).Unix(),
		UserID: e.ownerID,
		Role:   ExtensionRoleExternal,
	}
	return e.CreateToken(claims)
}

// CreateBroadcasterToken creates a token for broadcaster-level operations.
func (e *ExtensionJWT) CreateBroadcasterToken(channelID string, expiration time.Duration) (string, error) {
	claims := &ExtensionJWTClaims{
		Exp:       time.Now().Add(expiration).Unix(),
		UserID:    e.ownerID,
		Role:      ExtensionRoleBroadcaster,
		ChannelID: channelID,
	}
	return e.CreateToken(claims)
}

// CreatePubSubToken creates a token for PubSub operations.
func (e *ExtensionJWT) CreatePubSubToken(channelID string, listen, send []string, expiration time.Duration) (string, error) {
	claims := &ExtensionJWTClaims{
		Exp:               time.Now().Add(expiration).Unix(),
		UserID:            e.ownerID,
		Role:              ExtensionRoleExternal,
		ChannelID:         channelID,
		PubsubPermsListen: listen,
		PubsubPermsSend:   send,
	}
	return e.CreateToken(claims)
}

// sign creates an HMAC-SHA256 signature.
func (e *ExtensionJWT) sign(message []byte) []byte {
	mac := hmac.New(sha256.New, e.secret)
	mac.Write(message)
	return mac.Sum(nil)
}

// base64URLEncode encodes bytes to base64url without padding.
func base64URLEncode(data []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(data)
	return encoded
}

// extensionTokenProvider is an adapter to provide extension JWT tokens.
type extensionTokenProvider struct {
	jwt   *ExtensionJWT
	mu    sync.RWMutex
	token *Token
}

// GetToken returns the current token, generating a new one if expired.
func (p *extensionTokenProvider) GetToken() *Token {
	// Fast path: check if token is valid with read lock
	p.mu.RLock()
	if p.token != nil && time.Now().Before(p.token.ExpiresAt) {
		token := p.token
		p.mu.RUnlock()
		return token
	}
	p.mu.RUnlock()

	// Slow path: generate new token with write lock
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if p.token != nil && time.Now().Before(p.token.ExpiresAt) {
		return p.token
	}

	// Generate a new EBS token
	tokenStr, err := p.jwt.CreateEBSToken(time.Hour)
	if err != nil {
		return nil
	}

	p.token = &Token{
		AccessToken: tokenStr,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	return p.token
}

// NewExtensionClient creates a new Helix client with extension JWT authentication.
// This client automatically adds the extension JWT to API requests.
func NewExtensionClient(clientID string, jwt *ExtensionJWT, opts ...Option) *Client {
	client := NewClient(clientID, nil, opts...)
	client.tokenProvider = &extensionTokenProvider{jwt: jwt}
	return client
}

// SetExtensionJWT updates the Client to use extension JWT authentication.
// This modifies the client to use the provided JWT for all requests.
func (c *Client) SetExtensionJWT(jwt *ExtensionJWT) {
	c.authClient = nil
	c.tokenProvider = &extensionTokenProvider{jwt: jwt}
}

// WithExtensionJWT returns an Option to configure extension JWT authentication.
func WithExtensionJWT(jwt *ExtensionJWT) Option {
	return func(c *Client) {
		c.authClient = nil
		c.tokenProvider = &extensionTokenProvider{jwt: jwt}
	}
}

// ParseExtensionJWT parses and validates an extension JWT.
// This is useful for verifying JWTs received from the Twitch extension frontend.
func ParseExtensionJWT(tokenString, base64Secret string) (*ExtensionJWTClaims, error) {
	secret, err := base64.StdEncoding.DecodeString(base64Secret)
	if err != nil {
		return nil, fmt.Errorf("decoding secret: %w", err)
	}

	// Split token into parts
	parts := splitToken(tokenString)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Verify signature
	message := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	expectedSig := mac.Sum(nil)

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decoding signature: %w", err)
	}

	if !hmac.Equal(expectedSig, sig) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding claims: %w", err)
	}

	var claims ExtensionJWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("parsing claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// splitToken splits a JWT string into its parts.
func splitToken(token string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}
