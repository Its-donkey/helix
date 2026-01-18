package helix

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

	// Test successful device code request (using official Twitch documentation values)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceCodeResponse{
			DeviceCode:      "ike3GM8QIdYZs43KdrWPIO36LofILoCyFEzjlQ91",
			UserCode:        "ABCDEFGH",
			VerificationURI: "https://www.twitch.tv/activate?public=true&device-code=ABCDEFGH",
			ExpiresIn:       1800,
			Interval:        5,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client2 := NewAuthClient(AuthConfig{ClientID: "wbmytr93xzw8zbg0p1izqyzzc5mbiz", Scopes: []string{"channel:read:subscriptions"}})
	client2.SetHTTPClient(server.Client())
	client2.SetEndpoints("", "", "", server.URL, "", "", "")

	resp, err := client2.GetDeviceCode(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DeviceCode != "ike3GM8QIdYZs43KdrWPIO36LofILoCyFEzjlQ91" {
		t.Errorf("expected ike3GM8QIdYZs43KdrWPIO36LofILoCyFEzjlQ91, got %s", resp.DeviceCode)
	}
	if resp.UserCode != "ABCDEFGH" {
		t.Errorf("expected ABCDEFGH, got %s", resp.UserCode)
	}
}

func TestAuthClient_PollDeviceToken(t *testing.T) {
	// Test missing client ID
	client := NewAuthClient(AuthConfig{})
	_, err := client.PollDeviceToken(context.Background(), "ike3GM8QIdYZs43KdrWPIO36LofILoCyFEzjlQ91")
	if err != ErrMissingClientID {
		t.Errorf("expected ErrMissingClientID, got %v", err)
	}

	// Test successful token polling (using official Twitch documentation values)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Token{
			AccessToken:  "rfx2uswqe8l4g1mkagrvg5tv0ks3",
			RefreshToken: "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01",
			TokenType:    "bearer",
			ExpiresIn:    14124,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client2 := NewAuthClient(AuthConfig{ClientID: "wbmytr93xzw8zbg0p1izqyzzc5mbiz"})
	client2.SetHTTPClient(server.Client())
	client2.SetEndpoints(server.URL, "", "", "", "", "", "")

	token, err := client2.PollDeviceToken(context.Background(), "ike3GM8QIdYZs43KdrWPIO36LofILoCyFEzjlQ91")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "rfx2uswqe8l4g1mkagrvg5tv0ks3" {
		t.Errorf("expected rfx2uswqe8l4g1mkagrvg5tv0ks3, got %s", token.AccessToken)
	}

	// Test authorization pending
	serverPending := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "authorization_pending"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverPending.Close()

	client3 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client3.SetHTTPClient(serverPending.Client())
	client3.SetEndpoints(serverPending.URL, "", "", "", "", "", "")

	_, err = client3.PollDeviceToken(context.Background(), "device-code")
	if err != ErrAuthorizationPending {
		t.Errorf("expected ErrAuthorizationPending, got %v", err)
	}

	// Test invalid device code
	serverInvalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "invalid device code"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverInvalid.Close()

	client4 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client4.SetHTTPClient(serverInvalid.Client())
	client4.SetEndpoints(serverInvalid.URL, "", "", "", "", "", "")

	_, err = client4.PollDeviceToken(context.Background(), "device-code")
	if err != ErrInvalidDeviceCode {
		t.Errorf("expected ErrInvalidDeviceCode, got %v", err)
	}

	// Test with client secret
	serverWithSecret := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("client_secret") != "secret" {
			t.Error("expected client_secret to be set")
		}
		resp := Token{AccessToken: "token", TokenType: "bearer", ExpiresIn: 3600}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverWithSecret.Close()

	client5 := NewAuthClient(AuthConfig{ClientID: "test-client", ClientSecret: "secret"})
	client5.SetHTTPClient(serverWithSecret.Client())
	client5.SetEndpoints(serverWithSecret.URL, "", "", "", "", "", "")

	_, err = client5.PollDeviceToken(context.Background(), "device-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	// Test successful validation (using official Twitch documentation values)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "OAuth rfx2uswqe8l4g1mkagrvg5tv0ks3" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := ValidationResponse{
			ClientID:  "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
			Login:     "twitchdev",
			Scopes:    []string{"channel:read:subscriptions"},
			UserID:    "141981764",
			ExpiresIn: 5520838,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", server.URL, "", "", "", "", "")

	valResp, err := client.ValidateToken(context.Background(), "rfx2uswqe8l4g1mkagrvg5tv0ks3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valResp.Login != "twitchdev" {
		t.Errorf("expected twitchdev, got %s", valResp.Login)
	}
	if valResp.UserID != "141981764" {
		t.Errorf("expected 141981764, got %s", valResp.UserID)
	}

	// Test unauthorized
	_, err = client.ValidateToken(context.Background(), "invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}

	// Test error response with JSON
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Status: 400, Message: "Bad request"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverError.Close()

	client2 := NewAuthClient(AuthConfig{})
	client2.SetHTTPClient(serverError.Client())
	client2.SetEndpoints("", serverError.URL, "", "", "", "", "")

	_, err = client2.ValidateToken(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "Bad request") {
		t.Errorf("expected error with 'Bad request', got %v", err)
	}

	// Test error response without valid JSON
	serverBadJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("not json"))
	}))
	defer serverBadJSON.Close()

	client3 := NewAuthClient(AuthConfig{})
	client3.SetHTTPClient(serverBadJSON.Client())
	client3.SetEndpoints("", serverBadJSON.URL, "", "", "", "", "")

	_, err = client3.ValidateToken(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
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
	client2.SetEndpoints("", "", server.URL, "", "", "", "")

	err = client2.RevokeToken(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test error response with JSON
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Status: 400, Message: "Invalid token"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverError.Close()

	client3 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client3.SetHTTPClient(serverError.Client())
	client3.SetEndpoints("", "", serverError.URL, "", "", "", "")

	err = client3.RevokeToken(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "Invalid token") {
		t.Errorf("expected error with 'Invalid token', got %v", err)
	}

	// Test error response without valid JSON
	serverBadJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("not json"))
	}))
	defer serverBadJSON.Close()

	client4 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client4.SetHTTPClient(serverBadJSON.Client())
	client4.SetEndpoints("", "", serverBadJSON.URL, "", "", "", "")

	err = client4.RevokeToken(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
}

func TestAuthClient_RevokeCurrentToken(t *testing.T) {
	// Test with no token
	client := NewAuthClient(AuthConfig{ClientID: "client"})
	err := client.RevokeCurrentToken(context.Background())
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}

	// Test successful revoke
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client2 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client2.SetHTTPClient(server.Client())
	client2.SetEndpoints("", "", server.URL, "", "", "", "")
	client2.SetToken(&Token{AccessToken: "test-token"})

	err = client2.RevokeCurrentToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client2.GetToken() != nil {
		t.Error("token should be nil after revocation")
	}

	// Test revoke error
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "error"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverError.Close()

	client3 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client3.SetHTTPClient(serverError.Client())
	client3.SetEndpoints("", "", serverError.URL, "", "", "", "")
	client3.SetToken(&Token{AccessToken: "test-token"})

	err = client3.RevokeCurrentToken(context.Background())
	if err == nil {
		t.Error("expected error")
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
	// Using official Twitch documentation values for Authorization Code Grant Flow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		resp := Token{
			AccessToken:  "rfx2uswqe8l4g1mkagrvg5tv0ks3",
			RefreshToken: "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01",
			TokenType:    "bearer",
			ExpiresIn:    14124,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost/callback",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Test via ExchangeCode
	token, err := client.ExchangeCode(context.Background(), "auth-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "rfx2uswqe8l4g1mkagrvg5tv0ks3" {
		t.Errorf("expected rfx2uswqe8l4g1mkagrvg5tv0ks3, got %s", token.AccessToken)
	}

	// Verify token was set on client
	storedToken := client.GetToken()
	if storedToken == nil || storedToken.AccessToken != "rfx2uswqe8l4g1mkagrvg5tv0ks3" {
		t.Error("token should be stored on client")
	}
}

func TestAuthClient_requestToken_Error(t *testing.T) {
	// Test error response with JSON
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Status: 400, Message: "Invalid grant"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverError.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(serverError.Client())
	client.SetEndpoints(serverError.URL, "", "", "", "", "", "")

	_, err := client.ExchangeCode(context.Background(), "bad-code")
	if err == nil || !strings.Contains(err.Error(), "Invalid grant") {
		t.Errorf("expected error with 'Invalid grant', got %v", err)
	}

	// Test error response without valid JSON
	serverBadJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("not json"))
	}))
	defer serverBadJSON.Close()

	client2 := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client2.SetHTTPClient(serverBadJSON.Client())
	client2.SetEndpoints(serverBadJSON.URL, "", "", "", "", "", "")

	_, err = client2.ExchangeCode(context.Background(), "code")
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
}

func TestAuthClient_GetAppAccessToken_Success(t *testing.T) {
	// Using official Twitch documentation values for Client Credentials Grant Flow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("grant_type") != "client_credentials" {
			t.Errorf("expected grant_type client_credentials, got %s", r.Form.Get("grant_type"))
		}
		resp := Token{
			AccessToken: "jostpf5q0uzmxmkba9iyug38kjtgh",
			TokenType:   "bearer",
			ExpiresIn:   5011271,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	token, err := client.GetAppAccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "jostpf5q0uzmxmkba9iyug38kjtgh" {
		t.Errorf("expected jostpf5q0uzmxmkba9iyug38kjtgh, got %s", token.AccessToken)
	}
}

func TestAuthClient_RefreshToken_Success(t *testing.T) {
	// Using official Twitch documentation values for Refresh Token Flow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("expected grant_type refresh_token, got %s", r.Form.Get("grant_type"))
		}
		resp := Token{
			AccessToken:  "rfx2uswqe8l4g1mkagrvg5tv0ks3",
			RefreshToken: "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01",
			TokenType:    "bearer",
			ExpiresIn:    14124,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	token, err := client.RefreshToken(context.Background(), "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "rfx2uswqe8l4g1mkagrvg5tv0ks3" {
		t.Errorf("expected rfx2uswqe8l4g1mkagrvg5tv0ks3, got %s", token.AccessToken)
	}
}

func TestAuthClient_RefreshToken_InvalidRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "Invalid refresh token"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	_, err := client.RefreshToken(context.Background(), "bad-refresh-token")
	if err != ErrInvalidRefreshToken {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthClient_RefreshCurrentToken_Success(t *testing.T) {
	// Using official Twitch documentation values for Refresh Token Flow
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Token{
			AccessToken:  "rfx2uswqe8l4g1mkagrvg5tv0ks3",
			RefreshToken: "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01",
			TokenType:    "bearer",
			ExpiresIn:    14124,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")
	client.SetToken(&Token{AccessToken: "old-token", RefreshToken: "5b93chm6hdve3mycz05zfzatkfdenfspp1h1ar2xxdalen01"})

	token, err := client.RefreshCurrentToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "rfx2uswqe8l4g1mkagrvg5tv0ks3" {
		t.Errorf("expected rfx2uswqe8l4g1mkagrvg5tv0ks3, got %s", token.AccessToken)
	}

	// Verify token was updated on client
	storedToken := client.GetToken()
	if storedToken.AccessToken != "rfx2uswqe8l4g1mkagrvg5tv0ks3" {
		t.Error("token should be updated on client")
	}
}

func TestAuthClient_ValidateCurrentToken_Success(t *testing.T) {
	// Using official Twitch documentation values for Token Validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ValidationResponse{
			ClientID:  "wbmytr93xzw8zbg0p1izqyzzc5mbiz",
			Login:     "twitchdev",
			Scopes:    []string{"channel:read:subscriptions"},
			UserID:    "141981764",
			ExpiresIn: 5520838,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", server.URL, "", "", "", "", "")
	client.SetToken(&Token{AccessToken: "rfx2uswqe8l4g1mkagrvg5tv0ks3"})

	valResp, err := client.ValidateCurrentToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valResp.Login != "twitchdev" {
		t.Errorf("expected twitchdev, got %s", valResp.Login)
	}
}

func TestAuthClient_ValidateToken_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", server.URL, "", "", "", "", "")

	_, err := client.ValidateToken(context.Background(), "bad-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
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
	client.SetEndpoints("", "", "", "", server.URL, "", "")

	config, err := client.GetOpenIDConfiguration(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Issuer != "https://id.twitch.tv/oauth2" {
		t.Errorf("expected issuer https://id.twitch.tv/oauth2, got %s", config.Issuer)
	}
}

func TestAuthClient_GetOpenIDConfiguration_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", server.URL, "", "")

	_, err := client.GetOpenIDConfiguration(context.Background())
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
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
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	jwks, err := client.GetJWKS(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jwks.Keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(jwks.Keys))
	}
	if jwks.Keys[0].Kid != "key1" {
		t.Errorf("expected kid key1, got %s", jwks.Keys[0].Kid)
	}
}

func TestAuthClient_GetJWKS_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.GetJWKS(context.Background())
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
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
	client.SetEndpoints("", "", "", "", "", server.URL, "")

	userInfo, err := client.GetOIDCUserInfo(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userInfo.PreferredUsername != "testuser" {
		t.Errorf("expected testuser, got %s", userInfo.PreferredUsername)
	}
	if userInfo.Email != "test@example.com" {
		t.Errorf("expected test@example.com, got %s", userInfo.Email)
	}
}

func TestAuthClient_GetOIDCUserInfo_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", server.URL, "")

	_, err := client.GetOIDCUserInfo(context.Background(), "bad-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthClient_GetOIDCUserInfo_Error(t *testing.T) {
	// Test error response with JSON
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Status: 400, Message: "Bad request"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverError.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(serverError.Client())
	client.SetEndpoints("", "", "", "", "", serverError.URL, "")

	_, err := client.GetOIDCUserInfo(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "Bad request") {
		t.Errorf("expected error with 'Bad request', got %v", err)
	}

	// Test error response without valid JSON
	serverBadJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("not json"))
	}))
	defer serverBadJSON.Close()

	client2 := NewAuthClient(AuthConfig{})
	client2.SetHTTPClient(serverBadJSON.Client())
	client2.SetEndpoints("", "", "", "", "", serverBadJSON.URL, "")

	_, err = client2.GetOIDCUserInfo(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
}

func TestAuthClient_GetCurrentOIDCUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OIDCUserInfo{
			Sub:               "12345",
			PreferredUsername: "testuser",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", server.URL, "")
	client.SetToken(&Token{AccessToken: "test-token"})

	userInfo, err := client.GetCurrentOIDCUserInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userInfo.PreferredUsername != "testuser" {
		t.Errorf("expected testuser, got %s", userInfo.PreferredUsername)
	}
}

func TestAuthClient_ExchangeCodeForOIDCToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OIDCToken{
			Token: Token{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				TokenType:    "bearer",
				ExpiresIn:    3600,
			},
			IDToken: "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJ0ZXN0In0.sig",
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
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	token, err := client.ExchangeCodeForOIDCToken(context.Background(), "auth-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "access-token" {
		t.Errorf("expected access-token, got %s", token.AccessToken)
	}
	if token.IDToken == "" {
		t.Error("expected ID token to be set")
	}
}

func TestAuthClient_ExchangeCodeForOIDCToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Status: 400, Message: "Invalid code"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	_, err := client.ExchangeCodeForOIDCToken(context.Background(), "bad-code")
	if err == nil || !strings.Contains(err.Error(), "Invalid code") {
		t.Errorf("expected error with 'Invalid code', got %v", err)
	}

	// Test error without valid JSON
	serverBadJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("not json"))
	}))
	defer serverBadJSON.Close()

	client2 := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client2.SetHTTPClient(serverBadJSON.Client())
	client2.SetEndpoints(serverBadJSON.URL, "", "", "", "", "", "")

	_, err = client2.ExchangeCodeForOIDCToken(context.Background(), "code")
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
}

func TestAuthClient_GetDeviceCode_ErrorBranches(t *testing.T) {
	// Test error response with JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Status: 400, Message: "Invalid request"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", server.URL, "", "", "")

	_, err := client.GetDeviceCode(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Invalid request") {
		t.Errorf("expected error with 'Invalid request', got %v", err)
	}

	// Test error response without valid JSON
	serverBadJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("not json"))
	}))
	defer serverBadJSON.Close()

	client2 := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client2.SetHTTPClient(serverBadJSON.Client())
	client2.SetEndpoints("", "", "", serverBadJSON.URL, "", "", "")

	_, err = client2.GetDeviceCode(context.Background())
	if err == nil || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected error with status 500, got %v", err)
	}
}

func TestAuthClient_WaitForDeviceToken_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	deviceCode := &DeviceCodeResponse{
		DeviceCode: "device-code",
		Interval:   1,
		ExpiresIn:  300,
	}

	_, err := client.WaitForDeviceToken(ctx, deviceCode)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestAuthClient_WaitForDeviceToken_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 2 {
			w.WriteHeader(http.StatusBadRequest)
			resp := AuthErrorResponse{Message: "authorization_pending"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		resp := Token{
			AccessToken: "access-token",
			TokenType:   "bearer",
			ExpiresIn:   3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	deviceCode := &DeviceCodeResponse{
		DeviceCode: "device-code",
		Interval:   1, // 1 second interval
		ExpiresIn:  300,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := client.WaitForDeviceToken(ctx, deviceCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "access-token" {
		t.Errorf("expected access-token, got %s", token.AccessToken)
	}
}

func TestAuthClient_WaitForDeviceToken_Expired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "authorization_pending"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	deviceCode := &DeviceCodeResponse{
		DeviceCode: "device-code",
		Interval:   1,
		ExpiresIn:  1, // Expires in 1 second
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.WaitForDeviceToken(ctx, deviceCode)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected error with 'expired', got %v", err)
	}
}

func TestAuthClient_WaitForDeviceToken_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "some other error"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	deviceCode := &DeviceCodeResponse{
		DeviceCode: "device-code",
		Interval:   1,
		ExpiresIn:  300,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := client.WaitForDeviceToken(ctx, deviceCode)
	if err == nil || strings.Contains(err.Error(), "expired") {
		t.Errorf("expected non-expired error, got %v", err)
	}
}

func TestAuthClient_AutoRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Test with no token - should just wait
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cancelFunc := client.AutoRefresh(ctx)
	defer cancelFunc()

	time.Sleep(50 * time.Millisecond)

	// Test with token that needs refresh
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(1 * time.Second), // Expires soon
	})

	time.Sleep(100 * time.Millisecond)
}

func TestAuthClient_AutoRefresh_Cancel(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	ctx := context.Background()
	cancel := client.AutoRefresh(ctx)

	// Cancel immediately
	cancel()
}

func TestJWK_RSAPublicKey_InvalidN(t *testing.T) {
	jwk := &JWK{
		Kty: "RSA",
		N:   "!!invalid!!",
		E:   "AQAB",
	}
	_, err := jwk.RSAPublicKey()
	if err == nil || !strings.Contains(err.Error(), "decoding modulus") {
		t.Errorf("expected decoding modulus error, got %v", err)
	}
}

func TestJWK_RSAPublicKey_InvalidE(t *testing.T) {
	jwk := &JWK{
		Kty: "RSA",
		N:   "AQAB",
		E:   "!!invalid!!",
	}
	_, err := jwk.RSAPublicKey()
	if err == nil || !strings.Contains(err.Error(), "decoding exponent") {
		t.Errorf("expected decoding exponent error, got %v", err)
	}
}

func TestParseIDToken_InvalidBase64(t *testing.T) {
	// Invalid base64 in payload
	_, err := ParseIDToken("header.!!invalid!!.signature")
	if err == nil || !strings.Contains(err.Error(), "decoding") {
		t.Errorf("expected decoding error, got %v", err)
	}
}

func TestParseIDToken_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON (base64 of "not json")
	_, err := ParseIDToken("header.bm90IGpzb24.signature")
	if err == nil || !strings.Contains(err.Error(), "parsing") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_SetEndpoints_Partial(t *testing.T) {
	client := NewAuthClient(AuthConfig{})

	// Original values
	originalToken := client.tokenEndpoint
	originalValidate := client.validateEndpoint

	// Set only token endpoint
	client.SetEndpoints("http://new-token", "", "", "", "", "", "")

	if client.tokenEndpoint != "http://new-token" {
		t.Errorf("expected http://new-token, got %s", client.tokenEndpoint)
	}
	if client.validateEndpoint != originalValidate {
		t.Errorf("validate endpoint should not change, got %s", client.validateEndpoint)
	}

	// Set empty string should not change
	client.tokenEndpoint = originalToken
	client.SetEndpoints("", "", "", "", "", "", "")
	if client.tokenEndpoint != originalToken {
		t.Errorf("empty string should not change endpoint")
	}
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

// Tests for JSON parsing errors in success responses
func TestAuthClient_GetDeviceCode_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", server.URL, "", "", "")

	_, err := client.GetDeviceCode(context.Background())
	if err == nil || !strings.Contains(err.Error(), "parsing device code response") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_ValidateToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", server.URL, "", "", "", "", "")

	_, err := client.ValidateToken(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "parsing validate response") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_requestToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	_, err := client.ExchangeCode(context.Background(), "code")
	if err == nil || !strings.Contains(err.Error(), "parsing token response") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_GetOpenIDConfiguration_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", server.URL, "", "")

	_, err := client.GetOpenIDConfiguration(context.Background())
	if err == nil || !strings.Contains(err.Error(), "parsing OIDC config response") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_ExchangeCodeForOIDCToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	_, err := client.ExchangeCodeForOIDCToken(context.Background(), "code")
	if err == nil || !strings.Contains(err.Error(), "parsing token response") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_GetOIDCUserInfo_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", server.URL, "")

	_, err := client.GetOIDCUserInfo(context.Background(), "token")
	if err == nil || !strings.Contains(err.Error(), "parsing userinfo response") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_GetJWKS_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.GetJWKS(context.Background())
	if err == nil || !strings.Contains(err.Error(), "parsing JWKS response") {
		t.Errorf("expected parsing error, got %v", err)
	}
}

func TestAuthClient_AutoRefresh_ImmediateRefresh(t *testing.T) {
	var refreshCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&refreshCount, 1)
		resp := Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Set token that already needs refresh (expired)
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Already expired
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	// Wait for refresh to happen
	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt64(&refreshCount) == 0 {
		t.Error("expected at least one refresh call")
	}
}

func TestAuthClient_AutoRefresh_RefreshError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "error"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Set token that already needs refresh (expired)
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	time.Sleep(500 * time.Millisecond)
}

func TestAuthClient_AutoRefresh_WaitForRefresh(t *testing.T) {
	refreshCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshCount++
		resp := Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "bearer",
			ExpiresIn:    1, // Expires in 1 second
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Set token that expires soon
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(3 * time.Second), // 5 min threshold means refresh now
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	time.Sleep(300 * time.Millisecond)
}

// ============================================================================
// ParseIDTokenHeader Tests
// ============================================================================

func TestParseIDTokenHeader_Valid(t *testing.T) {
	// Create a valid JWT header: {"alg":"RS256","typ":"JWT","kid":"key1"}
	headerJSON := `{"alg":"RS256","typ":"JWT","kid":"key1"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	token := headerB64 + ".payload.signature"

	header, err := ParseIDTokenHeader(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if header.Alg != "RS256" {
		t.Errorf("expected alg RS256, got %s", header.Alg)
	}
	if header.Typ != "JWT" {
		t.Errorf("expected typ JWT, got %s", header.Typ)
	}
	if header.Kid != "key1" {
		t.Errorf("expected kid key1, got %s", header.Kid)
	}
}

func TestParseIDTokenHeader_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"too few parts", "header.payload"},
		{"too many parts", "a.b.c.d"},
		{"single part", "justonepart"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseIDTokenHeader(tt.token)
			if err == nil {
				t.Error("expected error for invalid format")
			}
			if !strings.Contains(err.Error(), "invalid ID token format") {
				t.Errorf("expected 'invalid ID token format' error, got: %v", err)
			}
		})
	}
}

func TestParseIDTokenHeader_InvalidBase64(t *testing.T) {
	// Invalid base64 in header
	_, err := ParseIDTokenHeader("!!invalid!!.payload.signature")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
	if !strings.Contains(err.Error(), "decoding ID token header") {
		t.Errorf("expected decoding error, got: %v", err)
	}
}

func TestParseIDTokenHeader_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON (base64 of "not json")
	invalidHeader := base64.RawURLEncoding.EncodeToString([]byte("not json"))
	_, err := ParseIDTokenHeader(invalidHeader + ".payload.signature")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parsing ID token header") {
		t.Errorf("expected parsing error, got: %v", err)
	}
}

// ============================================================================
// VerifyIDTokenSignature Tests
// ============================================================================

// Helper function to create a test RSA key pair and signed JWT
func createTestJWT(t *testing.T, claims map[string]interface{}, kid string) (string, *rsa.PrivateKey) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	// Create header
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": kid,
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Create payload
	payloadJSON, _ := json.Marshal(claims)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Sign
	signedContent := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signedContent))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signedContent + "." + signatureB64, privateKey
}

// Helper function to create JWKS from private key
func createTestJWKS(privateKey *rsa.PrivateKey, kid string) *JWKS {
	pubKey := &privateKey.PublicKey
	return &JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Kid: kid,
				Alg: "RS256",
				Use: "sig",
				N:   base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}), // 65537 = 0x010001
			},
		},
	}
}

func TestVerifyIDTokenSignature_Valid(t *testing.T) {
	claims := map[string]interface{}{
		"iss": "https://id.twitch.tv/oauth2",
		"sub": "12345",
		"aud": "client-id",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key-1")
	jwks := createTestJWKS(privateKey, "test-key-1")

	err := VerifyIDTokenSignature(jwt, jwks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyIDTokenSignature_InvalidFormat(t *testing.T) {
	jwks := &JWKS{Keys: []JWK{}}
	err := VerifyIDTokenSignature("invalid", jwks)
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid ID token format") {
		t.Errorf("expected format error, got: %v", err)
	}
}

func TestVerifyIDTokenSignature_UnsupportedAlgorithm(t *testing.T) {
	// Create a JWT with HS256 algorithm
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
		"kid": "key1",
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"123"}`))
	token := headerB64 + "." + payloadB64 + ".signature"

	jwks := &JWKS{Keys: []JWK{}}
	err := VerifyIDTokenSignature(token, jwks)
	if err == nil {
		t.Error("expected error for unsupported algorithm")
	}
	if !strings.Contains(err.Error(), "unsupported signing algorithm") {
		t.Errorf("expected algorithm error, got: %v", err)
	}
}

func TestVerifyIDTokenSignature_KeyNotFound(t *testing.T) {
	claims := map[string]interface{}{"sub": "123"}
	jwt, _ := createTestJWT(t, claims, "unknown-key")

	// JWKS with different key ID
	jwks := &JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Kid: "different-key",
				N:   "AQAB",
				E:   "AQAB",
			},
		},
	}

	err := VerifyIDTokenSignature(jwt, jwks)
	if err == nil {
		t.Error("expected error for key not found")
	}
	if !strings.Contains(err.Error(), "key not found in JWKS") {
		t.Errorf("expected key not found error, got: %v", err)
	}
}

func TestVerifyIDTokenSignature_InvalidSignature(t *testing.T) {
	claims := map[string]interface{}{"sub": "123"}
	jwt, privateKey := createTestJWT(t, claims, "key1")
	jwks := createTestJWKS(privateKey, "key1")

	// Tamper with the signature
	parts := strings.Split(jwt, ".")
	parts[2] = "invalid_signature_AAAA"
	tamperedJWT := strings.Join(parts, ".")

	err := VerifyIDTokenSignature(tamperedJWT, jwks)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
	if !strings.Contains(err.Error(), "signature verification failed") && !strings.Contains(err.Error(), "decoding signature") {
		t.Errorf("expected signature error, got: %v", err)
	}
}

func TestVerifyIDTokenSignature_InvalidRSAKey(t *testing.T) {
	claims := map[string]interface{}{"sub": "123"}
	jwt, _ := createTestJWT(t, claims, "key1")

	// JWKS with invalid RSA key (non-RSA type)
	jwks := &JWKS{
		Keys: []JWK{
			{
				Kty: "EC", // Wrong key type
				Kid: "key1",
				N:   "AQAB",
				E:   "AQAB",
			},
		},
	}

	err := VerifyIDTokenSignature(jwt, jwks)
	if err == nil {
		t.Error("expected error for invalid RSA key")
	}
	if !strings.Contains(err.Error(), "converting JWK to RSA key") {
		t.Errorf("expected RSA conversion error, got: %v", err)
	}
}

func TestVerifyIDTokenSignature_InvalidSignatureBase64(t *testing.T) {
	claims := map[string]interface{}{"sub": "123"}
	jwt, privateKey := createTestJWT(t, claims, "key1")
	jwks := createTestJWKS(privateKey, "key1")

	// Replace signature with invalid base64
	parts := strings.Split(jwt, ".")
	parts[2] = "!!invalid-base64!!"
	tamperedJWT := strings.Join(parts, ".")

	err := VerifyIDTokenSignature(tamperedJWT, jwks)
	if err == nil {
		t.Error("expected error for invalid base64 signature")
	}
	if !strings.Contains(err.Error(), "decoding signature") {
		t.Errorf("expected decoding error, got: %v", err)
	}
}

// ============================================================================
// VerifyAndParseIDToken Tests
// ============================================================================

func TestVerifyAndParseIDToken_Success(t *testing.T) {
	claims := map[string]interface{}{
		"iss":                "https://id.twitch.tv/oauth2",
		"sub":                "12345",
		"aud":                "client-id",
		"exp":                float64(time.Now().Add(time.Hour).Unix()),
		"iat":                float64(time.Now().Unix()),
		"preferred_username": "testuser",
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key")
	jwks := createTestJWKS(privateKey, "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	parsedClaims, err := client.VerifyAndParseIDToken(context.Background(), jwt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedClaims.Sub != "12345" {
		t.Errorf("expected sub 12345, got %s", parsedClaims.Sub)
	}
	if parsedClaims.PreferredUsername != "testuser" {
		t.Errorf("expected preferred_username testuser, got %s", parsedClaims.PreferredUsername)
	}
}

func TestVerifyAndParseIDToken_JWKSFetchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.VerifyAndParseIDToken(context.Background(), "header.payload.sig")
	if err == nil {
		t.Error("expected error for JWKS fetch failure")
	}
	if !strings.Contains(err.Error(), "fetching JWKS") {
		t.Errorf("expected JWKS fetch error, got: %v", err)
	}
}

func TestVerifyAndParseIDToken_SignatureVerificationError(t *testing.T) {
	// Create a valid-looking JWT but with wrong signature
	claims := map[string]interface{}{"sub": "123"}
	jwt, _ := createTestJWT(t, claims, "key1")

	// JWKS with different key (signature won't match)
	differentKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwks := createTestJWKS(differentKey, "key1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.VerifyAndParseIDToken(context.Background(), jwt)
	if err == nil {
		t.Error("expected error for signature verification failure")
	}
}

func TestVerifyAndParseIDToken_ParseError(t *testing.T) {
	// Create a JWT with valid signature but invalid payload JSON
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwks := createTestJWKS(privateKey, "key1")

	// Manually create a JWT with invalid payload
	header := map[string]string{"alg": "RS256", "typ": "JWT", "kid": "key1"}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	// Invalid JSON payload
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte("not valid json"))

	signedContent := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signedContent))
	signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)
	jwt := signedContent + "." + signatureB64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.VerifyAndParseIDToken(context.Background(), jwt)
	if err == nil {
		t.Error("expected error for parse failure")
	}
}

// ============================================================================
// ValidateIDToken Tests
// ============================================================================

func TestValidateIDToken_Success(t *testing.T) {
	claims := map[string]interface{}{
		"iss":   "https://id.twitch.tv/oauth2",
		"sub":   "12345",
		"aud":   "client-id",
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"iat":   float64(time.Now().Unix()),
		"nonce": "test-nonce",
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key")
	jwks := createTestJWKS(privateKey, "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client-id"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	parsedClaims, err := client.ValidateIDToken(context.Background(), jwt, "test-nonce")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedClaims.Sub != "12345" {
		t.Errorf("expected sub 12345, got %s", parsedClaims.Sub)
	}
}

func TestValidateIDToken_VerificationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client-id"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.ValidateIDToken(context.Background(), "header.payload.sig", "")
	if err == nil {
		t.Error("expected error for verification failure")
	}
}

func TestValidateIDToken_ClaimsValidationError(t *testing.T) {
	// Create a valid signed JWT but with invalid claims (wrong issuer)
	claims := map[string]interface{}{
		"iss": "https://wrong.issuer.com",
		"sub": "12345",
		"aud": "client-id",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"iat": float64(time.Now().Unix()),
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key")
	jwks := createTestJWKS(privateKey, "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client-id"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.ValidateIDToken(context.Background(), jwt, "")
	if err == nil {
		t.Error("expected error for invalid claims")
	}
	if !strings.Contains(err.Error(), "invalid issuer") {
		t.Errorf("expected issuer error, got: %v", err)
	}
}

func TestValidateIDToken_WithNonceValidation(t *testing.T) {
	claims := map[string]interface{}{
		"iss":   "https://id.twitch.tv/oauth2",
		"sub":   "12345",
		"aud":   "client-id",
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"iat":   float64(time.Now().Unix()),
		"nonce": "wrong-nonce",
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key")
	jwks := createTestJWKS(privateKey, "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client-id"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.ValidateIDToken(context.Background(), jwt, "expected-nonce")
	if err == nil {
		t.Error("expected error for nonce mismatch")
	}
	if !strings.Contains(err.Error(), "nonce mismatch") {
		t.Errorf("expected nonce error, got: %v", err)
	}
}

// ============================================================================
// AutoRefresh Additional Coverage Tests
// ============================================================================

func TestAuthClient_AutoRefresh_NoToken(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	// No token set

	ctx, ctxCancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	// Should just wait without panicking
	time.Sleep(100 * time.Millisecond)
}

func TestAuthClient_AutoRefresh_NoRefreshToken(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	// Token with no refresh token
	client.SetToken(&Token{
		AccessToken: "access-token",
		// No refresh token
		ExpiresAt: time.Now().Add(time.Hour),
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	time.Sleep(100 * time.Millisecond)
}

func TestAuthClient_AutoRefresh_ZeroExpiresAt(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	// Token with refresh token but zero ExpiresAt
	client.SetToken(&Token{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		// ExpiresAt is zero
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	time.Sleep(100 * time.Millisecond)
}

func TestAuthClient_AutoRefresh_ScheduledRefresh(t *testing.T) {
	var refreshCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&refreshCount, 1)
		resp := Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Token that expires soon (within 5 min threshold)
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(2 * time.Minute), // Less than 5 min, should trigger refresh
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt64(&refreshCount) == 0 {
		t.Error("expected refresh to be triggered for token expiring soon")
	}
}

func TestAuthClient_AutoRefresh_ScheduledRefreshError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "error"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Token that expires soon
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(2 * time.Minute),
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	// Should handle error gracefully and retry
	time.Sleep(300 * time.Millisecond)
}

func TestAuthClient_AutoRefresh_ContextCancelDuringWait(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	// Token with future expiry (will wait)
	client.SetToken(&Token{
		AccessToken:  "token",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	})

	ctx, ctxCancel := context.WithCancel(context.Background())
	cancel := client.AutoRefresh(ctx)

	// Cancel after a short time
	go func() {
		time.Sleep(50 * time.Millisecond)
		ctxCancel()
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
}

// ============================================================================
// Additional edge case tests
// ============================================================================

func TestVerifyIDTokenSignature_WrongKeySignature(t *testing.T) {
	// Create a JWT signed with one key
	claims := map[string]interface{}{"sub": "123"}
	jwt, _ := createTestJWT(t, claims, "key1")

	// Create JWKS with a different key but same kid
	differentKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwks := createTestJWKS(differentKey, "key1")

	err := VerifyIDTokenSignature(jwt, jwks)
	if err == nil {
		t.Error("expected signature verification to fail")
	}
	if !strings.Contains(err.Error(), "signature verification failed") {
		t.Errorf("expected signature verification error, got: %v", err)
	}
}

func TestValidateIDToken_ExpiredToken(t *testing.T) {
	claims := map[string]interface{}{
		"iss": "https://id.twitch.tv/oauth2",
		"sub": "12345",
		"aud": "client-id",
		"exp": float64(time.Now().Add(-time.Hour).Unix()), // Expired
		"iat": float64(time.Now().Add(-2 * time.Hour).Unix()),
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key")
	jwks := createTestJWKS(privateKey, "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client-id"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.ValidateIDToken(context.Background(), jwt, "")
	if err == nil {
		t.Error("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expired error, got: %v", err)
	}
}

func TestValidateIDToken_WrongAudience(t *testing.T) {
	claims := map[string]interface{}{
		"iss": "https://id.twitch.tv/oauth2",
		"sub": "12345",
		"aud": "wrong-client-id",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"iat": float64(time.Now().Unix()),
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key")
	jwks := createTestJWKS(privateKey, "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client-id"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	_, err := client.ValidateIDToken(context.Background(), jwt, "")
	if err == nil {
		t.Error("expected error for wrong audience")
	}
	if !strings.Contains(err.Error(), "invalid audience") {
		t.Errorf("expected audience error, got: %v", err)
	}
}

func TestValidateIDToken_EmptyNonceNotChecked(t *testing.T) {
	claims := map[string]interface{}{
		"iss":   "https://id.twitch.tv/oauth2",
		"sub":   "12345",
		"aud":   "client-id",
		"exp":   float64(time.Now().Add(time.Hour).Unix()),
		"iat":   float64(time.Now().Unix()),
		"nonce": "some-nonce",
	}

	jwt, privateKey := createTestJWT(t, claims, "test-key")
	jwks := createTestJWKS(privateKey, "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{ClientID: "client-id"})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints("", "", "", "", "", "", server.URL)

	// Empty nonce should not be checked
	parsedClaims, err := client.ValidateIDToken(context.Background(), jwt, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedClaims.Nonce != "some-nonce" {
		t.Errorf("expected nonce some-nonce, got %s", parsedClaims.Nonce)
	}
}

// Test error for parsing header due to invalid header in VerifyIDTokenSignature
func TestVerifyIDTokenSignature_InvalidHeaderJSON(t *testing.T) {
	// Create a token with invalid JSON in header (valid base64 of invalid JSON)
	invalidHeader := base64.RawURLEncoding.EncodeToString([]byte("not valid json"))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"123"}`))
	token := fmt.Sprintf("%s.%s.signature", invalidHeader, payload)

	jwks := &JWKS{Keys: []JWK{}}
	err := VerifyIDTokenSignature(token, jwks)
	if err == nil {
		t.Error("expected error for invalid header JSON")
	}
	if !strings.Contains(err.Error(), "parsing token header") {
		t.Errorf("expected parsing header error, got: %v", err)
	}
}

// ============================================================================
// Additional coverage tests for remaining edge cases
// ============================================================================

// Failing HTTP client for testing network errors
type failingTransport struct{}

func (f *failingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("network error")
}

func TestAuthClient_GetDeviceCode_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	_, err := client.GetDeviceCode(context.Background())
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing device code request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

func TestAuthClient_ValidateToken_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	_, err := client.ValidateToken(context.Background(), "token")
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing validate request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

func TestAuthClient_RevokeToken_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	err := client.RevokeToken(context.Background(), "token")
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing revoke request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

func TestAuthClient_requestToken_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	_, err := client.ExchangeCode(context.Background(), "code")
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing token request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

func TestAuthClient_GetOpenIDConfiguration_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	_, err := client.GetOpenIDConfiguration(context.Background())
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing OIDC config request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

func TestAuthClient_GetJWKS_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	_, err := client.GetJWKS(context.Background())
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing JWKS request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

func TestAuthClient_GetOIDCUserInfo_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	_, err := client.GetOIDCUserInfo(context.Background(), "token")
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing userinfo request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

func TestAuthClient_ExchangeCodeForOIDCToken_NetworkError(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(&http.Client{Transport: &failingTransport{}})

	_, err := client.ExchangeCodeForOIDCToken(context.Background(), "code")
	if err == nil {
		t.Error("expected network error")
	}
	if !strings.Contains(err.Error(), "executing token request") {
		t.Errorf("expected executing error, got: %v", err)
	}
}

// Test reading body error using a response that errors on Read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (e *errorReader) Close() error {
	return nil
}

// bodyErrorTransport returns a response with an error-producing body
type bodyErrorTransport struct {
	statusCode int
}

func (t *bodyErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: t.statusCode,
		Body:       &errorReader{},
		Header:     make(http.Header),
	}, nil
}

func TestAuthClient_GetDeviceCode_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 200}})
	client.SetEndpoints("", "", "", "http://test", "", "", "")

	_, err := client.GetDeviceCode(context.Background())
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading device code response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestAuthClient_ValidateToken_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 200}})
	client.SetEndpoints("", "http://test", "", "", "", "", "")

	_, err := client.ValidateToken(context.Background(), "token")
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading validate response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestAuthClient_RevokeToken_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 400}}) // Non-200 to trigger body read
	client.SetEndpoints("", "", "http://test", "", "", "", "")

	err := client.RevokeToken(context.Background(), "token")
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading revoke response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestAuthClient_requestToken_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 200}})
	client.SetEndpoints("http://test", "", "", "", "", "", "")

	_, err := client.ExchangeCode(context.Background(), "code")
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading token response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestAuthClient_GetOpenIDConfiguration_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 200}})
	client.SetEndpoints("", "", "", "", "http://test", "", "")

	_, err := client.GetOpenIDConfiguration(context.Background())
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading OIDC config response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestAuthClient_GetJWKS_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 200}})
	client.SetEndpoints("", "", "", "", "", "", "http://test")

	_, err := client.GetJWKS(context.Background())
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading JWKS response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestAuthClient_GetOIDCUserInfo_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 200}})
	client.SetEndpoints("", "", "", "", "", "http://test", "")

	_, err := client.GetOIDCUserInfo(context.Background(), "token")
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading userinfo response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestAuthClient_ExchangeCodeForOIDCToken_ReadBodyError(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(&http.Client{Transport: &bodyErrorTransport{statusCode: 200}})
	client.SetEndpoints("http://test", "", "", "", "", "", "")

	_, err := client.ExchangeCodeForOIDCToken(context.Background(), "code")
	if err == nil {
		t.Error("expected read error")
	}
	if !strings.Contains(err.Error(), "reading token response") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

// ============================================================================
// AutoRefresh exhaustive coverage tests
// ============================================================================

func TestAuthClient_AutoRefresh_ContextCancelInNoTokenWait(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	// No token set - will enter the "no token" wait path

	ctx, ctxCancel := context.WithCancel(context.Background())

	cancel := client.AutoRefresh(ctx)

	// Give goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel context during the wait
	ctxCancel()

	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestAuthClient_AutoRefresh_ContextCancelInZeroExpiryWait(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	// Token with refresh token but zero ExpiresAt - will enter the "zero expiry" wait path
	client.SetToken(&Token{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		// ExpiresAt is zero
	})

	ctx, ctxCancel := context.WithCancel(context.Background())

	cancel := client.AutoRefresh(ctx)

	// Give goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel context during the wait
	ctxCancel()

	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestAuthClient_AutoRefresh_ContextCancelDuringImmediateRefreshError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "error"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Token that needs immediate refresh (expired)
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
	})

	ctx, ctxCancel := context.WithCancel(context.Background())

	cancel := client.AutoRefresh(ctx)

	// Give goroutine time to hit the error and enter the retry wait
	time.Sleep(50 * time.Millisecond)

	// Cancel during the error retry wait
	ctxCancel()

	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestAuthClient_AutoRefresh_ContextCancelDuringScheduledRefreshWait(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	// Token with future expiry (will wait for scheduled refresh)
	client.SetToken(&Token{
		AccessToken:  "token",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(10 * time.Minute), // More than 5 min threshold
	})

	ctx, ctxCancel := context.WithCancel(context.Background())
	cancel := client.AutoRefresh(ctx)

	// Give goroutine time to enter the wait
	time.Sleep(10 * time.Millisecond)

	// Cancel during the scheduled wait
	ctxCancel()

	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestAuthClient_AutoRefresh_ContextCancelDuringScheduledRefreshError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest)
		resp := AuthErrorResponse{Message: "error"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Token that expires very soon (within 5 min threshold)
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(1 * time.Minute),
	})

	ctx, ctxCancel := context.WithCancel(context.Background())

	cancel := client.AutoRefresh(ctx)

	// Wait for refresh to fail and enter error retry wait
	time.Sleep(100 * time.Millisecond)

	// Cancel during the error retry wait
	ctxCancel()

	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestAuthClient_AutoRefresh_SuccessfulRefreshAfterWait(t *testing.T) {
	var refreshCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&refreshCount, 1)
		resp := Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetHTTPClient(server.Client())
	client.SetEndpoints(server.URL, "", "", "", "", "", "")

	// Token that expires in 4 minutes (less than 5 min threshold, triggers refresh after short wait)
	client.SetToken(&Token{
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(4 * time.Minute),
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer ctxCancel()

	cancel := client.AutoRefresh(ctx)
	defer cancel()

	// Wait for refresh to happen
	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt64(&refreshCount) == 0 {
		t.Error("expected at least one refresh call")
	}
}

// Test GetOIDCAuthorizationURL with claims JSON marshaling error (using invalid value)
func TestAuthClient_GetOIDCAuthorizationURL_ClaimsMarshalError(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:    "test-client",
		RedirectURI: "http://localhost/callback",
	})

	// Create claims with a channel value which can't be marshaled to JSON
	invalidClaims := map[string]interface{}{
		"test": make(chan int),
	}

	_, err := client.GetOIDCAuthorizationURL(ResponseTypeCode, "", invalidClaims)
	if err == nil {
		t.Error("expected marshaling error for invalid claims")
	}
	if !strings.Contains(err.Error(), "marshaling claims") {
		t.Errorf("expected marshaling error, got: %v", err)
	}
}

// ============================================================================
// Tests for http.NewRequestWithContext error paths using invalid URLs
// ============================================================================

func TestAuthClient_GetDeviceCode_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	// Set an invalid URL that will fail URL parsing
	client.SetEndpoints("", "", "", "://invalid", "", "", "")

	_, err := client.GetDeviceCode(context.Background())
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating device code request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}

func TestAuthClient_ValidateToken_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetEndpoints("", "://invalid", "", "", "", "", "")

	_, err := client.ValidateToken(context.Background(), "token")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating validate request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}

func TestAuthClient_RevokeToken_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{ClientID: "test-client"})
	client.SetEndpoints("", "", "://invalid", "", "", "", "")

	err := client.RevokeToken(context.Background(), "token")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating revoke request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}

func TestAuthClient_requestToken_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetEndpoints("://invalid", "", "", "", "", "", "")

	_, err := client.ExchangeCode(context.Background(), "code")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating token request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}

func TestAuthClient_GetOpenIDConfiguration_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetEndpoints("", "", "", "", "://invalid", "", "")

	_, err := client.GetOpenIDConfiguration(context.Background())
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating OIDC config request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}

func TestAuthClient_GetJWKS_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetEndpoints("", "", "", "", "", "", "://invalid")

	_, err := client.GetJWKS(context.Background())
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating JWKS request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}

func TestAuthClient_GetOIDCUserInfo_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{})
	client.SetEndpoints("", "", "", "", "", "://invalid", "")

	_, err := client.GetOIDCUserInfo(context.Background(), "token")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating userinfo request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}

func TestAuthClient_ExchangeCodeForOIDCToken_InvalidURL(t *testing.T) {
	client := NewAuthClient(AuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	client.SetEndpoints("://invalid", "", "", "", "", "", "")

	_, err := client.ExchangeCodeForOIDCToken(context.Background(), "code")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "creating token request") {
		t.Errorf("expected creating request error, got: %v", err)
	}
}
