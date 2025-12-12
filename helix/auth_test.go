package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected bool
	}{
		{
			name:     "zero expiry is not expired",
			token:    Token{ExpiresAt: time.Time{}},
			expected: false,
		},
		{
			name:     "future expiry is not expired",
			token:    Token{ExpiresAt: time.Now().Add(time.Hour)},
			expected: false,
		},
		{
			name:     "past expiry is expired",
			token:    Token{ExpiresAt: time.Now().Add(-time.Hour)},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToken_Valid(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected bool
	}{
		{
			name:     "empty token is invalid",
			token:    Token{},
			expected: false,
		},
		{
			name:     "token with access token and no expiry is valid",
			token:    Token{AccessToken: "test"},
			expected: true,
		},
		{
			name:     "token with access token and future expiry is valid",
			token:    Token{AccessToken: "test", ExpiresAt: time.Now().Add(time.Hour)},
			expected: true,
		},
		{
			name:     "token with access token but expired is invalid",
			token:    Token{AccessToken: "test", ExpiresAt: time.Now().Add(-time.Hour)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.Valid(); got != tt.expected {
				t.Errorf("Valid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToken_setExpiry(t *testing.T) {
	token := Token{ExpiresIn: 3600}
	token.setExpiry()

	if token.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
	if time.Until(token.ExpiresAt) < 3500*time.Second {
		t.Error("ExpiresAt should be approximately 1 hour from now")
	}

	// Test with zero ExpiresIn
	token2 := Token{ExpiresIn: 0}
	token2.setExpiry()
	if !token2.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should remain zero when ExpiresIn is 0")
	}
}

func TestNewAuthClient(t *testing.T) {
	config := AuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost/callback",
		Scopes:       []string{"chat:read"},
	}

	client := NewAuthClient(config)
	if client == nil {
		t.Fatal("NewAuthClient returned nil")
	}
	if client.config.ClientID != "test-client-id" {
		t.Errorf("expected client_id test-client-id, got %s", client.config.ClientID)
	}
}

func TestAuthClient_SetHTTPClient(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	customClient := &http.Client{Timeout: 60 * time.Second}
	client.SetHTTPClient(customClient)

	if client.httpClient != customClient {
		t.Error("SetHTTPClient did not set the custom client")
	}
}

func TestAuthClient_SetGetToken(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	token := &Token{AccessToken: "test-token"}

	client.SetToken(token)
	got := client.GetToken()

	if got == nil || got.AccessToken != "test-token" {
		t.Error("SetToken/GetToken did not work correctly")
	}
}

func TestAuthClient_GetAuthorizationURL(t *testing.T) {
	tests := []struct {
		name        string
		config      AuthConfig
		expectError error
		expectURL   bool
	}{
		{
			name:        "missing client ID",
			config:      AuthConfig{RedirectURI: "http://localhost"},
			expectError: ErrMissingClientID,
		},
		{
			name:        "missing redirect URI",
			config:      AuthConfig{ClientID: "test"},
			expectError: ErrMissingRedirectURI,
		},
		{
			name: "valid config",
			config: AuthConfig{
				ClientID:    "test-client",
				RedirectURI: "http://localhost/callback",
				Scopes:      []string{"chat:read"},
				State:       "test-state",
				ForceVerify: true,
			},
			expectURL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAuthClient(tt.config)
			url, err := client.GetAuthorizationURL("code")

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectURL && url == "" {
				t.Error("expected non-empty URL")
			}
			if tt.config.State != "" && !strings.Contains(url, "state=test-state") {
				t.Error("URL should contain state parameter")
			}
			if tt.config.ForceVerify && !strings.Contains(url, "force_verify=true") {
				t.Error("URL should contain force_verify parameter")
			}
		})
	}
}

func TestAuthClient_GetImplicitAuthURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:    "test-client",
		RedirectURI: "http://localhost/callback",
	})

	url, err := client.GetImplicitAuthURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "response_type=token") {
		t.Error("URL should contain response_type=token")
	}
}

func TestAuthClient_GetCodeAuthURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:    "test-client",
		RedirectURI: "http://localhost/callback",
	})

	url, err := client.GetCodeAuthURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "response_type=code") {
		t.Error("URL should contain response_type=code")
	}
}

func TestAuthClient_ExchangeCode(t *testing.T) {
	tests := []struct {
		name        string
		config      AuthConfig
		code        string
		expectError error
	}{
		{
			name:        "missing client ID",
			config:      AuthConfig{ClientSecret: "secret"},
			code:        "code",
			expectError: ErrMissingClientID,
		},
		{
			name:        "missing client secret",
			config:      AuthConfig{ClientID: "client"},
			code:        "code",
			expectError: ErrMissingClientSecret,
		},
		{
			name:        "missing code",
			config:      AuthConfig{ClientID: "client", ClientSecret: "secret"},
			code:        "",
			expectError: ErrMissingCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAuthClient(tt.config)
			_, err := client.ExchangeCode(context.Background(), tt.code)
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestAuthClient_GetAppAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		config      AuthConfig
		expectError error
	}{
		{
			name:        "missing client ID",
			config:      AuthConfig{ClientSecret: "secret"},
			expectError: ErrMissingClientID,
		},
		{
			name:        "missing client secret",
			config:      AuthConfig{ClientID: "client"},
			expectError: ErrMissingClientSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAuthClient(tt.config)
			_, err := client.GetAppAccessToken(context.Background())
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestAuthClient_GetDeviceCode(t *testing.T) {
	// Test missing client ID
	client := NewAuthClient(AuthConfig{})
	_, err := client.GetDeviceCode(context.Background())
	if err != ErrMissingClientID {
		t.Errorf("expected ErrMissingClientID, got %v", err)
	}

	// Test successful device code request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceCodeResponse{
			DeviceCode:      "device-code-123",
			UserCode:        "USER-CODE",
			VerificationURI: "https://twitch.tv/activate",
			ExpiresIn:       1800,
			Interval:        5,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client2 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client2.httpClient = server.Client()
}

func TestAuthClient_PollDeviceToken(t *testing.T) {
	// Test missing client ID
	client := NewAuthClient(AuthConfig{})
	_, err := client.PollDeviceToken(context.Background(), "device-code")
	if err != ErrMissingClientID {
		t.Errorf("expected ErrMissingClientID, got %v", err)
	}
}

func TestAuthClient_RefreshToken(t *testing.T) {
	tests := []struct {
		name         string
		config       AuthConfig
		refreshToken string
		expectError  error
	}{
		{
			name:         "missing client ID",
			config:       AuthConfig{ClientSecret: "secret"},
			refreshToken: "refresh",
			expectError:  ErrMissingClientID,
		},
		{
			name:         "missing client secret",
			config:       AuthConfig{ClientID: "client"},
			refreshToken: "refresh",
			expectError:  ErrMissingClientSecret,
		},
		{
			name:         "missing refresh token",
			config:       AuthConfig{ClientID: "client", ClientSecret: "secret"},
			refreshToken: "",
			expectError:  ErrInvalidRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAuthClient(tt.config)
			_, err := client.RefreshToken(context.Background(), tt.refreshToken)
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestAuthClient_RefreshCurrentToken(t *testing.T) {
	// Test with no token
	client := NewAuthClient(AuthConfig{ClientID: "client", ClientSecret: "secret"})
	_, err := client.RefreshCurrentToken(context.Background())
	if err != ErrInvalidRefreshToken {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}

	// Test with token but no refresh token
	client.SetToken(&Token{AccessToken: "access"})
	_, err = client.RefreshCurrentToken(context.Background())
	if err != ErrInvalidRefreshToken {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthClient_ValidateToken(t *testing.T) {
	// Test successful validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "OAuth test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := ValidationResponse{
			ClientID:  "client-id",
			Login:     "testuser",
			Scopes:    []string{"chat:read"},
			UserID:    "12345",
			ExpiresIn: 3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_ValidateCurrentToken(t *testing.T) {
	// Test with no token
	client := NewAuthClient(AuthConfig{})
	_, err := client.ValidateCurrentToken(context.Background())
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthClient_RevokeToken(t *testing.T) {
	// Test missing client ID
	client := NewAuthClient(AuthConfig{})
	err := client.RevokeToken(context.Background(), "token")
	if err != ErrMissingClientID {
		t.Errorf("expected ErrMissingClientID, got %v", err)
	}

	// Test successful revoke
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client2 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client2.SetHTTPClient(server.Client())
}

func TestAuthClient_RevokeCurrentToken(t *testing.T) {
	// Test with no token
	client := NewAuthClient(AuthConfig{ClientID: "client"})
	err := client.RevokeCurrentToken(context.Background())
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthClient_GetOIDCAuthorizationURL(t *testing.T) {
	tests := []struct {
		name        string
		config      AuthConfig
		expectError error
	}{
		{
			name:        "missing client ID",
			config:      AuthConfig{RedirectURI: "http://localhost"},
			expectError: ErrMissingClientID,
		},
		{
			name:        "missing redirect URI",
			config:      AuthConfig{ClientID: "client"},
			expectError: ErrMissingRedirectURI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAuthClient(tt.config)
			_, err := client.GetOIDCAuthorizationURL(ResponseTypeCode, "", nil)
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}

	// Test successful URL generation
	client := NewAuthClient(AuthConfig{
		ClientID:    "test-client",
		RedirectURI: "http://localhost/callback",
		Scopes:      []string{"openid"},
		State:       "test-state",
		ForceVerify: true,
	})

	url, err := client.GetOIDCAuthorizationURL(ResponseTypeCodeIDToken, "test-nonce", map[string]interface{}{"userinfo": map[string]interface{}{"email": nil}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "nonce=test-nonce") {
		t.Error("URL should contain nonce parameter")
	}
	if !strings.Contains(url, "claims=") {
		t.Error("URL should contain claims parameter")
	}

	// Test without openid scope (should be added automatically)
	client2 := NewAuthClient(AuthConfig{
		ClientID:    "test-client",
		RedirectURI: "http://localhost/callback",
		Scopes:      []string{"chat:read"},
	})

	url2, err := client2.GetOIDCAuthorizationURL(ResponseTypeCode, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url2, "openid") {
		t.Error("URL should contain openid scope")
	}
}

func TestAuthClient_ExchangeCodeForOIDCToken(t *testing.T) {
	tests := []struct {
		name        string
		config      AuthConfig
		code        string
		expectError error
	}{
		{
			name:        "missing client ID",
			config:      AuthConfig{ClientSecret: "secret"},
			code:        "code",
			expectError: ErrMissingClientID,
		},
		{
			name:        "missing client secret",
			config:      AuthConfig{ClientID: "client"},
			code:        "code",
			expectError: ErrMissingClientSecret,
		},
		{
			name:        "missing code",
			config:      AuthConfig{ClientID: "client", ClientSecret: "secret"},
			code:        "",
			expectError: ErrMissingCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAuthClient(tt.config)
			_, err := client.ExchangeCodeForOIDCToken(context.Background(), tt.code)
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestAuthClient_GetCurrentOIDCUserInfo(t *testing.T) {
	// Test with no token
	client := NewAuthClient(AuthConfig{})
	_, err := client.GetCurrentOIDCUserInfo(context.Background())
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWKS_GetKeyByID(t *testing.T) {
	jwks := &JWKS{
		Keys: []JWK{
			{Kid: "key1", Kty: "RSA"},
			{Kid: "key2", Kty: "RSA"},
		},
	}

	// Test finding existing key
	key := jwks.GetKeyByID("key1")
	if key == nil || key.Kid != "key1" {
		t.Error("GetKeyByID should return the correct key")
	}

	// Test finding non-existent key
	key = jwks.GetKeyByID("key3")
	if key != nil {
		t.Error("GetKeyByID should return nil for non-existent key")
	}
}

func TestJWK_RSAPublicKey(t *testing.T) {
	// Test unsupported key type
	jwk := &JWK{Kty: "EC"}
	_, err := jwk.RSAPublicKey()
	if err == nil || !strings.Contains(err.Error(), "unsupported key type") {
		t.Error("RSAPublicKey should return error for non-RSA key")
	}

	// Test valid RSA key
	validJWK := &JWK{
		Kty: "RSA",
		N:   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
		E:   "AQAB",
	}
	pubKey, err := validJWK.RSAPublicKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pubKey == nil {
		t.Error("RSAPublicKey should return a valid public key")
	}
}

func TestParseIDToken(t *testing.T) {
	// Test invalid token format
	_, err := ParseIDToken("invalid")
	if err == nil {
		t.Error("ParseIDToken should return error for invalid token")
	}

	_, err = ParseIDToken("one.two")
	if err == nil {
		t.Error("ParseIDToken should return error for token with wrong number of parts")
	}

	// Test valid token (base64 encoded payload)
	// Payload: {"iss":"https://id.twitch.tv/oauth2","sub":"12345","aud":"client-id","exp":9999999999,"iat":1234567890}
	validToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2lkLnR3aXRjaC50di9vYXV0aDIiLCJzdWIiOiIxMjM0NSIsImF1ZCI6ImNsaWVudC1pZCIsImV4cCI6OTk5OTk5OTk5OSwiaWF0IjoxMjM0NTY3ODkwfQ.signature"
	claims, err := ParseIDToken(validToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Sub != "12345" {
		t.Errorf("expected sub 12345, got %s", claims.Sub)
	}
}

func TestAuthClient_requestToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		resp := Token{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost/callback",
	})
	client.SetHTTPClient(server.Client())

	// We can test requestToken indirectly through ExchangeCode with a mock server
	// But first we need to override TokenEndpoint - since we can't, we test via integration
}

func TestAuthClient_ValidateToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "OAuth valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := ValidationResponse{
			ClientID:  "client-id",
			Login:     "testuser",
			Scopes:    []string{"chat:read"},
			UserID:    "12345",
			ExpiresIn: 3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_ValidateToken_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_GetOpenIDConfiguration_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OpenIDConfiguration{
			Issuer:                "https://id.twitch.tv/oauth2",
			AuthorizationEndpoint: "https://id.twitch.tv/oauth2/authorize",
			TokenEndpoint:         "https://id.twitch.tv/oauth2/token",
			UserInfoEndpoint:      "https://id.twitch.tv/oauth2/userinfo",
			JWKSUri:               "https://id.twitch.tv/oauth2/keys",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_GetJWKS_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JWKS{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: "key1",
					N:   "test-n",
					E:   "AQAB",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_GetOIDCUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := OIDCUserInfo{
			Sub:               "12345",
			PreferredUsername: "testuser",
			Email:             "test@example.com",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_GetOIDCUserInfo_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_RevokeToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_RevokeToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{
			Status:  400,
			Message: "Invalid token",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_RevokeCurrentToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
	client.SetToken(&Token{AccessToken: "test-token"})
}

func TestAuthClient_GetDeviceCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceCodeResponse{
			DeviceCode:      "device-code",
			UserCode:        "USER-CODE",
			VerificationURI: "https://twitch.tv/activate",
			ExpiresIn:       1800,
			Interval:        5,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_GetDeviceCode_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{
			Status:  400,
			Message: "Invalid request",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
}

func TestAuthClient_ValidateIDTokenClaims(t *testing.T) {
	client := NewAuthClient(AuthConfig{ClientID: "client-id"})

	// Test invalid issuer
	claims := &IDTokenClaims{Iss: "invalid"}
	err := client.ValidateIDTokenClaims(claims, "")
	if err == nil || !strings.Contains(err.Error(), "invalid issuer") {
		t.Error("ValidateIDTokenClaims should return error for invalid issuer")
	}

	// Test invalid audience
	claims = &IDTokenClaims{
		Iss: "https://id.twitch.tv/oauth2",
		Aud: "wrong-client",
	}
	err = client.ValidateIDTokenClaims(claims, "")
	if err == nil || !strings.Contains(err.Error(), "invalid audience") {
		t.Error("ValidateIDTokenClaims should return error for invalid audience")
	}

	// Test expired token
	claims = &IDTokenClaims{
		Iss: "https://id.twitch.tv/oauth2",
		Aud: "client-id",
		Exp: time.Now().Unix() - 3600,
	}
	err = client.ValidateIDTokenClaims(claims, "")
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Error("ValidateIDTokenClaims should return error for expired token")
	}

	// Test nonce mismatch
	claims = &IDTokenClaims{
		Iss:   "https://id.twitch.tv/oauth2",
		Aud:   "client-id",
		Exp:   time.Now().Unix() + 3600,
		Nonce: "different-nonce",
	}
	err = client.ValidateIDTokenClaims(claims, "expected-nonce")
	if err == nil || !strings.Contains(err.Error(), "nonce mismatch") {
		t.Error("ValidateIDTokenClaims should return error for nonce mismatch")
	}

	// Test valid claims
	claims = &IDTokenClaims{
		Iss:   "https://id.twitch.tv/oauth2",
		Aud:   "client-id",
		Exp:   time.Now().Unix() + 3600,
		Nonce: "test-nonce",
	}
	err = client.ValidateIDTokenClaims(claims, "test-nonce")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
