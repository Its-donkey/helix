package helix

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewExtensionJWT(t *testing.T) {
	// Create a test secret (base64 encoded)
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))

	jwt, err := NewExtensionJWT("ext123", secret, "owner456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if jwt.ExtensionID() != "ext123" {
		t.Errorf("expected extension ID ext123, got %s", jwt.ExtensionID())
	}
	if jwt.OwnerID() != "owner456" {
		t.Errorf("expected owner ID owner456, got %s", jwt.OwnerID())
	}
}

func TestNewExtensionJWT_InvalidSecret(t *testing.T) {
	_, err := NewExtensionJWT("ext123", "not-base64!", "owner456")
	if err == nil {
		t.Error("expected error for invalid base64 secret")
	}
}

func TestExtensionJWT_CreateToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	claims := &ExtensionJWTClaims{
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: "owner456",
		Role:   ExtensionRoleExternal,
	}

	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Token should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 token parts, got %d", len(parts))
	}

	// Verify we can parse the token back
	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsed.UserID != "owner456" {
		t.Errorf("expected user ID owner456, got %s", parsed.UserID)
	}
	if parsed.Role != ExtensionRoleExternal {
		t.Errorf("expected role external, got %s", parsed.Role)
	}
}

func TestExtensionJWT_CreateEBSToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	token, err := jwt.CreateEBSToken(time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsed.Role != ExtensionRoleExternal {
		t.Errorf("expected role external, got %s", parsed.Role)
	}
	if parsed.UserID != "owner456" {
		t.Errorf("expected user ID owner456, got %s", parsed.UserID)
	}
}

func TestExtensionJWT_CreateBroadcasterToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	token, err := jwt.CreateBroadcasterToken("channel789", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsed.Role != ExtensionRoleBroadcaster {
		t.Errorf("expected role broadcaster, got %s", parsed.Role)
	}
	if parsed.ChannelID != "channel789" {
		t.Errorf("expected channel ID channel789, got %s", parsed.ChannelID)
	}
}

func TestExtensionJWT_CreatePubSubToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	listen := []string{"broadcast", "global"}
	send := []string{"broadcast"}

	token, err := jwt.CreatePubSubToken("channel789", listen, send, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := ParseExtensionJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if len(parsed.PubsubPermsListen) != 2 {
		t.Errorf("expected 2 listen permissions, got %d", len(parsed.PubsubPermsListen))
	}
	if len(parsed.PubsubPermsSend) != 1 {
		t.Errorf("expected 1 send permission, got %d", len(parsed.PubsubPermsSend))
	}
}

func TestParseExtensionJWT_InvalidFormat(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))

	tests := []struct {
		name  string
		token string
	}{
		{"no dots", "nodots"},
		{"one dot", "one.dot"},
		{"empty parts", ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseExtensionJWT(tt.token, secret)
			if err == nil {
				t.Error("expected error for invalid token format")
			}
		})
	}
}

func TestParseExtensionJWT_InvalidSignature(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("correct-secret"))
	wrongSecret := base64.StdEncoding.EncodeToString([]byte("wrong-secret"))

	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")
	token, _ := jwt.CreateEBSToken(time.Hour)

	// Try to parse with wrong secret
	_, err := ParseExtensionJWT(token, wrongSecret)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestParseExtensionJWT_Expired(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	// Create a token that's already expired
	claims := &ExtensionJWTClaims{
		Exp:    time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		UserID: "owner456",
		Role:   ExtensionRoleExternal,
	}

	token, _ := jwt.CreateToken(claims)

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expiration error, got: %v", err)
	}
}

func TestExtensionTokenProvider_GetToken(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// First call should generate a new token
	token1 := provider.GetToken()
	if token1 == nil {
		t.Fatal("expected token, got nil")
	}
	if token1.AccessToken == "" {
		t.Error("expected non-empty access token")
	}

	// Second call should return the same token (cached)
	token2 := provider.GetToken()
	if token2.AccessToken != token1.AccessToken {
		t.Error("expected same cached token")
	}
}

func TestExtensionTokenProvider_GetToken_Concurrent(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// Run concurrent GetToken calls to test thread safety
	const goroutines = 50
	done := make(chan *Token, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			token := provider.GetToken()
			done <- token
		}()
	}

	// Collect all tokens
	var tokens []*Token
	for i := 0; i < goroutines; i++ {
		token := <-done
		if token == nil {
			t.Error("expected token, got nil")
		}
		tokens = append(tokens, token)
	}

	// All tokens should be valid
	for _, token := range tokens {
		if token.AccessToken == "" {
			t.Error("expected non-empty access token")
		}
	}
}

func TestExtensionJWT_CreateToken_DefaultExpiration(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	// Create claims with Exp=0 (should use default 1 hour expiration)
	claims := &ExtensionJWTClaims{
		UserID: "user123",
		Role:   ExtensionRoleExternal,
		// Exp is 0 - should be set to default
	}

	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Verify that Exp was set
	if claims.Exp == 0 {
		t.Error("expected Exp to be set to default value")
	}
}

func TestExtensionTokenProvider_GetToken_DoubleCheck(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// Pre-populate with a valid token
	provider.mu.Lock()
	provider.token = &Token{
		AccessToken: "pre-existing-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	provider.mu.Unlock()

	// GetToken should return the cached token (hitting the fast path)
	token := provider.GetToken()
	if token == nil {
		t.Fatal("expected token, got nil")
	}
	if token.AccessToken != "pre-existing-token" {
		t.Error("expected cached token to be returned")
	}
}

func TestExtensionTokenProvider_GetToken_ExpiredThenRefreshed(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	provider := &extensionTokenProvider{jwt: jwt}

	// Pre-populate with an expired token
	provider.mu.Lock()
	provider.token = &Token{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-time.Hour), // Already expired
	}
	provider.mu.Unlock()

	// GetToken should generate a new token
	token := provider.GetToken()
	if token == nil {
		t.Fatal("expected token, got nil")
	}
	if token.AccessToken == "expired-token" {
		t.Error("expected new token, got expired token")
	}
}

func TestNewExtensionClient(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	client := NewExtensionClient("client-id", jwt)
	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.tokenProvider == nil {
		t.Error("expected token provider to be set")
	}
}

func TestWithExtensionJWT(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("client-id", authClient, WithExtensionJWT(jwt))

	if client.tokenProvider == nil {
		t.Error("expected token provider to be set")
	}
	if client.authClient != nil {
		t.Error("expected authClient to be nil when using extension JWT")
	}
}

func TestClient_SetExtensionJWT(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("client-id", authClient)

	// Initially should have authClient set
	if client.authClient == nil {
		t.Error("expected authClient to be set initially")
	}

	// After setting extension JWT, authClient should be nil
	client.SetExtensionJWT(jwt)
	if client.authClient != nil {
		t.Error("expected authClient to be nil after SetExtensionJWT")
	}
	if client.tokenProvider == nil {
		t.Error("expected tokenProvider to be set after SetExtensionJWT")
	}
}

func TestJWTClaimsStructure(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")

	claims := &ExtensionJWTClaims{
		Exp:               time.Now().Add(time.Hour).Unix(),
		UserID:            "user123",
		Role:              ExtensionRoleBroadcaster,
		ChannelID:         "channel456",
		OpaqueUserID:      "U123",
		IsUnlinked:        true,
		PubsubPermsListen: []string{"broadcast"},
		PubsubPermsSend:   []string{"broadcast", "whisper-user123"},
	}

	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Decode the claims part manually to verify structure
	parts := strings.Split(token, ".")
	claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])

	var decoded map[string]interface{}
	if err := json.Unmarshal(claimsJSON, &decoded); err != nil {
		t.Fatalf("failed to decode claims: %v", err)
	}

	// Verify key fields are present
	if decoded["user_id"] != "user123" {
		t.Errorf("expected user_id user123, got %v", decoded["user_id"])
	}
	if decoded["role"] != string(ExtensionRoleBroadcaster) {
		t.Errorf("expected role broadcaster, got %v", decoded["role"])
	}
	if decoded["channel_id"] != "channel456" {
		t.Errorf("expected channel_id channel456, got %v", decoded["channel_id"])
	}
	if decoded["is_unlinked"] != true {
		t.Errorf("expected is_unlinked true, got %v", decoded["is_unlinked"])
	}
}

func TestParseExtensionJWT_InvalidClaimsBase64(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))

	// Create a token with invalid base64 in the claims part
	// Use valid header but invalid claims that will fail base64 decoding
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	invalidClaims := "!!!invalid-base64!!!"
	sig := base64.RawURLEncoding.EncodeToString([]byte("fake-sig"))

	token := header + "." + invalidClaims + "." + sig

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for invalid claims base64")
	}
}

func TestParseExtensionJWT_InvalidClaimsJSON(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret-key-1234567890"))

	// Create a token with valid base64 but invalid JSON in claims
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	invalidJSON := base64.RawURLEncoding.EncodeToString([]byte(`{not valid json`))

	// Create a valid signature for these parts
	jwt, _ := NewExtensionJWT("ext123", secret, "owner456")
	message := header + "." + invalidJSON
	sig := base64.RawURLEncoding.EncodeToString(jwt.sign([]byte(message)))

	token := message + "." + sig

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for invalid claims JSON")
	}
	if !strings.Contains(err.Error(), "parsing claims") {
		t.Errorf("expected 'parsing claims' error, got: %v", err)
	}
}

func TestParseExtensionJWT_InvalidSecretBase64(t *testing.T) {
	_, err := ParseExtensionJWT("header.claims.sig", "!!!invalid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid secret base64")
	}
	if !strings.Contains(err.Error(), "decoding secret") {
		t.Errorf("expected 'decoding secret' error, got: %v", err)
	}
}

func TestParseExtensionJWT_InvalidSignatureBase64(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))

	// Create a token with invalid base64 in the signature part
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"exp":9999999999,"user_id":"test","role":"external"}`))
	invalidSig := "!!!invalid-base64!!!"

	token := header + "." + claims + "." + invalidSig

	_, err := ParseExtensionJWT(token, secret)
	if err == nil {
		t.Error("expected error for invalid signature base64")
	}
	if !strings.Contains(err.Error(), "decoding signature") {
		t.Errorf("expected 'decoding signature' error, got: %v", err)
	}
}
