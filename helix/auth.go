package helix

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Auth URL constants
const (
	// TwitchAuthURL is the base URL for Twitch OAuth endpoints.
	TwitchAuthURL = "https://id.twitch.tv/oauth2"

	// AuthorizeEndpoint is the authorization endpoint.
	AuthorizeEndpoint = TwitchAuthURL + "/authorize"

	// TokenEndpoint is the token endpoint.
	TokenEndpoint = TwitchAuthURL + "/token"

	// ValidateEndpoint is the token validation endpoint.
	ValidateEndpoint = TwitchAuthURL + "/validate"

	// RevokeEndpoint is the token revocation endpoint.
	RevokeEndpoint = TwitchAuthURL + "/revoke"

	// DeviceEndpoint is the device authorization endpoint.
	DeviceEndpoint = TwitchAuthURL + "/device"

	// OpenIDConfigurationEndpoint is the OIDC discovery endpoint.
	OpenIDConfigurationEndpoint = TwitchAuthURL + "/.well-known/openid-configuration"

	// UserInfoEndpoint is the OIDC userinfo endpoint.
	UserInfoEndpoint = TwitchAuthURL + "/userinfo"

	// JWKSEndpoint is the JSON Web Key Set endpoint.
	JWKSEndpoint = TwitchAuthURL + "/keys"
)

// Auth errors
var (
	ErrInvalidToken         = errors.New("invalid access token")
	ErrTokenExpired         = errors.New("token has expired")
	ErrAuthorizationPending = errors.New("authorization pending")
	ErrInvalidDeviceCode    = errors.New("invalid device code")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrMissingClientID      = errors.New("client ID is required")
	ErrMissingClientSecret  = errors.New("client secret is required")
	ErrMissingRedirectURI   = errors.New("redirect URI is required")
	ErrMissingCode          = errors.New("authorization code is required")
)

// Token represents an OAuth token from Twitch.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        []string  `json:"scope,omitempty"`
	ExpiresAt    time.Time `json:"-"`
}

// IsExpired returns true if the token has expired.
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(t.ExpiresAt)
}

// Valid returns true if the token is non-empty and not expired.
func (t *Token) Valid() bool {
	return t.AccessToken != "" && !t.IsExpired()
}

// setExpiry calculates and sets the ExpiresAt field based on ExpiresIn.
func (t *Token) setExpiry() {
	if t.ExpiresIn > 0 {
		t.ExpiresAt = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	}
}

// ValidationResponse represents the response from the /validate endpoint.
type ValidationResponse struct {
	ClientID  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserID    string   `json:"user_id"`
	ExpiresIn int      `json:"expires_in"`
}

// DeviceCodeResponse represents the response from the device authorization endpoint.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
}

// AuthErrorResponse represents an error response from Twitch auth endpoints.
type AuthErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// AuthConfig holds the configuration for OAuth.
type AuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	ForceVerify  bool
	State        string
}

// AuthClient is an OAuth client for Twitch authentication.
type AuthClient struct {
	config     AuthConfig
	httpClient *http.Client
	token      *Token
	mu         sync.RWMutex
}

// NewAuthClient creates a new OAuth client with the given configuration.
func NewAuthClient(config AuthConfig) *AuthClient {
	return &AuthClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetHTTPClient sets a custom HTTP client.
func (c *AuthClient) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// SetToken sets the current token.
func (c *AuthClient) SetToken(token *Token) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

// GetToken returns the current token.
func (c *AuthClient) GetToken() *Token {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// GetAuthorizationURL returns the URL to redirect users to for authorization.
func (c *AuthClient) GetAuthorizationURL(responseType string) (string, error) {
	if c.config.ClientID == "" {
		return "", ErrMissingClientID
	}
	if c.config.RedirectURI == "" {
		return "", ErrMissingRedirectURI
	}

	params := url.Values{
		"client_id":     {c.config.ClientID},
		"redirect_uri":  {c.config.RedirectURI},
		"response_type": {responseType},
		"scope":         {strings.Join(c.config.Scopes, " ")},
	}

	if c.config.State != "" {
		params.Set("state", c.config.State)
	}
	if c.config.ForceVerify {
		params.Set("force_verify", "true")
	}

	return AuthorizeEndpoint + "?" + params.Encode(), nil
}

// GetImplicitAuthURL returns the authorization URL for the Implicit Grant flow.
func (c *AuthClient) GetImplicitAuthURL() (string, error) {
	return c.GetAuthorizationURL("token")
}

// GetCodeAuthURL returns the authorization URL for the Authorization Code Grant flow.
func (c *AuthClient) GetCodeAuthURL() (string, error) {
	return c.GetAuthorizationURL("code")
}

// ExchangeCode exchanges an authorization code for an access token.
func (c *AuthClient) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	if c.config.ClientID == "" {
		return nil, ErrMissingClientID
	}
	if c.config.ClientSecret == "" {
		return nil, ErrMissingClientSecret
	}
	if code == "" {
		return nil, ErrMissingCode
	}

	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {c.config.RedirectURI},
	}

	return c.requestToken(ctx, data)
}

// GetAppAccessToken obtains an app access token using the Client Credentials flow.
func (c *AuthClient) GetAppAccessToken(ctx context.Context) (*Token, error) {
	if c.config.ClientID == "" {
		return nil, ErrMissingClientID
	}
	if c.config.ClientSecret == "" {
		return nil, ErrMissingClientSecret
	}

	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"grant_type":    {"client_credentials"},
	}

	return c.requestToken(ctx, data)
}

// GetDeviceCode initiates the Device Code flow.
func (c *AuthClient) GetDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	if c.config.ClientID == "" {
		return nil, ErrMissingClientID
	}

	data := url.Values{
		"client_id": {c.config.ClientID},
		"scopes":    {strings.Join(c.config.Scopes, " ")},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DeviceEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating device code request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing device code request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading device code response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp AuthErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("device code request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("device code request failed: %s", errResp.Message)
	}

	var dcResp DeviceCodeResponse
	if err := json.Unmarshal(body, &dcResp); err != nil {
		return nil, fmt.Errorf("parsing device code response: %w", err)
	}

	return &dcResp, nil
}

// PollDeviceToken polls for the device token after the user has authorized.
func (c *AuthClient) PollDeviceToken(ctx context.Context, deviceCode string) (*Token, error) {
	if c.config.ClientID == "" {
		return nil, ErrMissingClientID
	}

	data := url.Values{
		"client_id":   {c.config.ClientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"scopes":      {strings.Join(c.config.Scopes, " ")},
	}

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	token, err := c.requestToken(ctx, data)
	if err != nil {
		if strings.Contains(err.Error(), "authorization_pending") {
			return nil, ErrAuthorizationPending
		}
		if strings.Contains(err.Error(), "invalid device code") {
			return nil, ErrInvalidDeviceCode
		}
		return nil, err
	}

	return token, nil
}

// WaitForDeviceToken polls for the device token until it's available or the context is cancelled.
func (c *AuthClient) WaitForDeviceToken(ctx context.Context, deviceCode *DeviceCodeResponse) (*Token, error) {
	ticker := time.NewTicker(time.Duration(deviceCode.Interval) * time.Second)
	defer ticker.Stop()

	expiresAt := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(expiresAt) {
				return nil, errors.New("device code expired")
			}

			token, err := c.PollDeviceToken(ctx, deviceCode.DeviceCode)
			if err == ErrAuthorizationPending {
				continue
			}
			if err != nil {
				return nil, err
			}
			return token, nil
		}
	}
}

// RefreshToken refreshes an access token using a refresh token.
func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	if c.config.ClientID == "" {
		return nil, ErrMissingClientID
	}
	if c.config.ClientSecret == "" {
		return nil, ErrMissingClientSecret
	}
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	token, err := c.requestToken(ctx, data)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid refresh token") {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	return token, nil
}

// RefreshCurrentToken refreshes the current token if it has a refresh token.
func (c *AuthClient) RefreshCurrentToken(ctx context.Context) (*Token, error) {
	c.mu.RLock()
	if c.token == nil || c.token.RefreshToken == "" {
		c.mu.RUnlock()
		return nil, ErrInvalidRefreshToken
	}
	refreshToken := c.token.RefreshToken
	c.mu.RUnlock()

	token, err := c.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	c.SetToken(token)
	return token, nil
}

// ValidateToken validates an access token.
func (c *AuthClient) ValidateToken(ctx context.Context, accessToken string) (*ValidationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ValidateEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating validate request: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing validate request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading validate response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidToken
	}

	if resp.StatusCode != http.StatusOK {
		var errResp AuthErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("validate request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("validate request failed: %s", errResp.Message)
	}

	var valResp ValidationResponse
	if err := json.Unmarshal(body, &valResp); err != nil {
		return nil, fmt.Errorf("parsing validate response: %w", err)
	}

	return &valResp, nil
}

// ValidateCurrentToken validates the current token.
func (c *AuthClient) ValidateCurrentToken(ctx context.Context) (*ValidationResponse, error) {
	c.mu.RLock()
	if c.token == nil {
		c.mu.RUnlock()
		return nil, ErrInvalidToken
	}
	accessToken := c.token.AccessToken
	c.mu.RUnlock()

	return c.ValidateToken(ctx, accessToken)
}

// RevokeToken revokes an access token.
func (c *AuthClient) RevokeToken(ctx context.Context, accessToken string) error {
	if c.config.ClientID == "" {
		return ErrMissingClientID
	}

	data := url.Values{
		"client_id": {c.config.ClientID},
		"token":     {accessToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RevokeEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing revoke request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading revoke response: %w", err)
	}

	var errResp AuthErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return fmt.Errorf("revoke request failed: %s", errResp.Message)
}

// RevokeCurrentToken revokes the current token.
func (c *AuthClient) RevokeCurrentToken(ctx context.Context) error {
	c.mu.RLock()
	if c.token == nil {
		c.mu.RUnlock()
		return ErrInvalidToken
	}
	accessToken := c.token.AccessToken
	c.mu.RUnlock()

	if err := c.RevokeToken(ctx, accessToken); err != nil {
		return err
	}

	c.mu.Lock()
	c.token = nil
	c.mu.Unlock()

	return nil
}

// requestToken makes a token request to the Twitch token endpoint.
func (c *AuthClient) requestToken(ctx context.Context, data url.Values) (*Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp AuthErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("token request failed: %s", errResp.Message)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	token.setExpiry()
	c.SetToken(&token)

	return &token, nil
}

// AutoRefresh starts a goroutine that automatically refreshes the token before it expires.
func (c *AuthClient) AutoRefresh(ctx context.Context) (cancel func()) {
	ctx, cancelFunc := context.WithCancel(ctx)

	go func() {
		for {
			c.mu.RLock()
			token := c.token
			c.mu.RUnlock()

			if token == nil || token.RefreshToken == "" {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Minute):
					continue
				}
			}

			refreshAt := token.ExpiresAt.Add(-5 * time.Minute)
			waitDuration := time.Until(refreshAt)

			if waitDuration <= 0 {
				if _, err := c.RefreshCurrentToken(ctx); err != nil {
					select {
					case <-ctx.Done():
						return
					case <-time.After(30 * time.Second):
						continue
					}
				}
				continue
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(waitDuration):
				if _, err := c.RefreshCurrentToken(ctx); err != nil {
					select {
					case <-ctx.Done():
						return
					case <-time.After(30 * time.Second):
						continue
					}
				}
			}
		}
	}()

	return cancelFunc
}

// ============================================================================
// OIDC Types and Functions
// ============================================================================

// OpenIDConfiguration represents the OIDC discovery document.
type OpenIDConfiguration struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSUri                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
}

// OIDCUserInfo represents the response from the OIDC UserInfo endpoint.
type OIDCUserInfo struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Picture           string `json:"picture,omitempty"`
	UpdatedAt         int64  `json:"updated_at,omitempty"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kty string `json:"kty"` // Key type (RSA)
	E   string `json:"e"`   // Exponent
	N   string `json:"n"`   // Modulus
	Kid string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm
	Use string `json:"use"` // Usage (sig)
}

// OIDCToken extends Token with OIDC-specific fields.
type OIDCToken struct {
	Token
	IDToken string `json:"id_token,omitempty"`
}

// IDTokenClaims represents the claims in an OIDC ID token.
type IDTokenClaims struct {
	Iss               string `json:"iss"`
	Sub               string `json:"sub"`
	Aud               string `json:"aud"`
	Exp               int64  `json:"exp"`
	Iat               int64  `json:"iat"`
	Nonce             string `json:"nonce,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Picture           string `json:"picture,omitempty"`
	UpdatedAt         int64  `json:"updated_at,omitempty"`
}

// OIDCResponseType represents OIDC response types.
type OIDCResponseType string

const (
	ResponseTypeCode         OIDCResponseType = "code"
	ResponseTypeToken        OIDCResponseType = "token"
	ResponseTypeIDToken      OIDCResponseType = "id_token"
	ResponseTypeTokenIDToken OIDCResponseType = "token id_token"
	ResponseTypeCodeIDToken  OIDCResponseType = "code id_token"
)

// GetOpenIDConfiguration fetches the OIDC discovery document.
func (c *AuthClient) GetOpenIDConfiguration(ctx context.Context) (*OpenIDConfiguration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, OpenIDConfigurationEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC config request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing OIDC config request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading OIDC config response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC config request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var config OpenIDConfiguration
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("parsing OIDC config response: %w", err)
	}

	return &config, nil
}

// GetOIDCAuthorizationURL returns the authorization URL for OIDC flows.
func (c *AuthClient) GetOIDCAuthorizationURL(responseType OIDCResponseType, nonce string, claims map[string]interface{}) (string, error) {
	if c.config.ClientID == "" {
		return "", ErrMissingClientID
	}
	if c.config.RedirectURI == "" {
		return "", ErrMissingRedirectURI
	}

	scopes := c.config.Scopes
	hasOpenID := false
	for _, s := range scopes {
		if s == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		scopes = append([]string{"openid"}, scopes...)
	}

	params := url.Values{
		"client_id":     {c.config.ClientID},
		"redirect_uri":  {c.config.RedirectURI},
		"response_type": {string(responseType)},
		"scope":         {strings.Join(scopes, " ")},
	}

	if c.config.State != "" {
		params.Set("state", c.config.State)
	}
	if c.config.ForceVerify {
		params.Set("force_verify", "true")
	}
	if nonce != "" {
		params.Set("nonce", nonce)
	}
	if claims != nil {
		claimsJSON, err := json.Marshal(claims)
		if err != nil {
			return "", fmt.Errorf("marshaling claims: %w", err)
		}
		params.Set("claims", string(claimsJSON))
	}

	return AuthorizeEndpoint + "?" + params.Encode(), nil
}

// ExchangeCodeForOIDCToken exchanges an authorization code for an OIDC token.
func (c *AuthClient) ExchangeCodeForOIDCToken(ctx context.Context, code string) (*OIDCToken, error) {
	if c.config.ClientID == "" {
		return nil, ErrMissingClientID
	}
	if c.config.ClientSecret == "" {
		return nil, ErrMissingClientSecret
	}
	if code == "" {
		return nil, ErrMissingCode
	}

	data := url.Values{
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {c.config.RedirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp AuthErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("token request failed: %s", errResp.Message)
	}

	var token OIDCToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	token.setExpiry()
	c.SetToken(&token.Token)

	return &token, nil
}

// GetOIDCUserInfo fetches user information from the OIDC UserInfo endpoint.
func (c *AuthClient) GetOIDCUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, UserInfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing userinfo request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading userinfo response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidToken
	}

	if resp.StatusCode != http.StatusOK {
		var errResp AuthErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("userinfo request failed: %s", errResp.Message)
	}

	var userInfo OIDCUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("parsing userinfo response: %w", err)
	}

	return &userInfo, nil
}

// GetCurrentOIDCUserInfo fetches user information using the current access token.
func (c *AuthClient) GetCurrentOIDCUserInfo(ctx context.Context) (*OIDCUserInfo, error) {
	c.mu.RLock()
	if c.token == nil {
		c.mu.RUnlock()
		return nil, ErrInvalidToken
	}
	accessToken := c.token.AccessToken
	c.mu.RUnlock()

	return c.GetOIDCUserInfo(ctx, accessToken)
}

// GetJWKS fetches the JSON Web Key Set for validating ID tokens.
func (c *AuthClient) GetJWKS(ctx context.Context) (*JWKS, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, JWKSEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating JWKS request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing JWKS request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading JWKS response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("parsing JWKS response: %w", err)
	}

	return &jwks, nil
}

// GetKeyByID returns the JWK with the specified key ID.
func (j *JWKS) GetKeyByID(kid string) *JWK {
	for i := range j.Keys {
		if j.Keys[i].Kid == kid {
			return &j.Keys[i]
		}
	}
	return nil
}

// RSAPublicKey converts the JWK to an RSA public key.
func (k *JWK) RSAPublicKey() (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", k.Kty)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decoding modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decoding exponent: %w", err)
	}
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// ParseIDToken parses an ID token without validating the signature.
func ParseIDToken(idToken string) (*IDTokenClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding ID token payload: %w", err)
	}

	var claims IDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parsing ID token claims: %w", err)
	}

	return &claims, nil
}

// ValidateIDTokenClaims validates the claims in an ID token.
func (c *AuthClient) ValidateIDTokenClaims(claims *IDTokenClaims, nonce string) error {
	if claims.Iss != "https://id.twitch.tv/oauth2" {
		return fmt.Errorf("invalid issuer: %s", claims.Iss)
	}

	if claims.Aud != c.config.ClientID {
		return fmt.Errorf("invalid audience: %s", claims.Aud)
	}

	if time.Now().Unix() > claims.Exp {
		return fmt.Errorf("ID token has expired")
	}

	if nonce != "" && claims.Nonce != nonce {
		return fmt.Errorf("nonce mismatch")
	}

	return nil
}

// ============================================================================
// Scope Constants
// ============================================================================

// Scope constants for Twitch API permissions.
const (
	// Analytics scopes
	ScopeAnalyticsReadExtensions = "analytics:read:extensions"
	ScopeAnalyticsReadGames      = "analytics:read:games"

	// Bits scopes
	ScopeBitsRead = "bits:read"

	// Channel scopes
	ScopeChannelBot               = "channel:bot"
	ScopeChannelEditCommercial    = "channel:edit:commercial"
	ScopeChannelManageAds         = "channel:manage:ads"
	ScopeChannelManageBroadcast   = "channel:manage:broadcast"
	ScopeChannelManageClips       = "channel:manage:clips"
	ScopeChannelManageExtensions  = "channel:manage:extensions"
	ScopeChannelManageModerators  = "channel:manage:moderators"
	ScopeChannelManagePolls       = "channel:manage:polls"
	ScopeChannelManagePredictions = "channel:manage:predictions"
	ScopeChannelManageRaids       = "channel:manage:raids"
	ScopeChannelManageRedemptions = "channel:manage:redemptions"
	ScopeChannelManageSchedule    = "channel:manage:schedule"
	ScopeChannelManageVideos      = "channel:manage:videos"
	ScopeChannelManageVIPs        = "channel:manage:vips"
	ScopeChannelManageGuestStar   = "channel:manage:guest_star"
	ScopeChannelModerate          = "channel:moderate"
	ScopeChannelReadAds           = "channel:read:ads"
	ScopeChannelReadCharity       = "channel:read:charity"
	ScopeChannelReadEditors       = "channel:read:editors"
	ScopeChannelReadGoals         = "channel:read:goals"
	ScopeChannelReadGuestStar     = "channel:read:guest_star"
	ScopeChannelReadHypeTrain     = "channel:read:hype_train"
	ScopeChannelReadPolls         = "channel:read:polls"
	ScopeChannelReadPredictions   = "channel:read:predictions"
	ScopeChannelReadRedemptions   = "channel:read:redemptions"
	ScopeChannelReadStreamKey     = "channel:read:stream_key"
	ScopeChannelReadSubscriptions = "channel:read:subscriptions"
	ScopeChannelReadVIPs          = "channel:read:vips"

	// Chat scopes
	ScopeChatEdit = "chat:edit"
	ScopeChatRead = "chat:read"

	// Clips scopes
	ScopeClipsEdit = "clips:edit"

	// Editor scopes
	ScopeEditorManageClips = "editor:manage:clips"

	// Moderation scopes
	ScopeModerationRead                 = "moderation:read"
	ScopeModeratorManageAnnouncements   = "moderator:manage:announcements"
	ScopeModeratorManageAutomod         = "moderator:manage:automod"
	ScopeModeratorManageAutomodSettings = "moderator:manage:automod_settings"
	ScopeModeratorManageBannedUsers     = "moderator:manage:banned_users"
	ScopeModeratorManageBlockedTerms    = "moderator:manage:blocked_terms"
	ScopeModeratorManageChatMessages    = "moderator:manage:chat_messages"
	ScopeModeratorManageChatSettings    = "moderator:manage:chat_settings"
	ScopeModeratorManageGuestStar       = "moderator:manage:guest_star"
	ScopeModeratorManageShieldMode      = "moderator:manage:shield_mode"
	ScopeModeratorManageShoutouts       = "moderator:manage:shoutouts"
	ScopeModeratorManageWarnings        = "moderator:manage:warnings"
	ScopeModeratorManageUnbanRequests   = "moderator:manage:unban_requests"
	ScopeModeratorReadAutomodSettings   = "moderator:read:automod_settings"
	ScopeModeratorReadBannedUsers       = "moderator:read:banned_users"
	ScopeModeratorReadBlockedTerms      = "moderator:read:blocked_terms"
	ScopeModeratorReadChatMessages      = "moderator:read:chat_messages"
	ScopeModeratorReadChatSettings      = "moderator:read:chat_settings"
	ScopeModeratorReadChatters          = "moderator:read:chatters"
	ScopeModeratorReadFollowers         = "moderator:read:followers"
	ScopeModeratorReadGuestStar         = "moderator:read:guest_star"
	ScopeModeratorReadShieldMode        = "moderator:read:shield_mode"
	ScopeModeratorReadShoutouts         = "moderator:read:shoutouts"
	ScopeModeratorReadSuspiciousUsers   = "moderator:read:suspicious_users"
	ScopeModeratorReadUnbanRequests     = "moderator:read:unban_requests"
	ScopeModeratorReadVIPs              = "moderator:read:vips"
	ScopeModeratorReadWarnings          = "moderator:read:warnings"

	// User scopes
	ScopeUserBot                   = "user:bot"
	ScopeUserEdit                  = "user:edit"
	ScopeUserEditBroadcast         = "user:edit:broadcast"
	ScopeUserManageBlockedUsers    = "user:manage:blocked_users"
	ScopeUserManageChatColor       = "user:manage:chat_color"
	ScopeUserManageWhispers        = "user:manage:whispers"
	ScopeUserReadBlockedUsers      = "user:read:blocked_users"
	ScopeUserReadBroadcast         = "user:read:broadcast"
	ScopeUserReadChat              = "user:read:chat"
	ScopeUserReadEmail             = "user:read:email"
	ScopeUserReadEmotes            = "user:read:emotes"
	ScopeUserReadFollows           = "user:read:follows"
	ScopeUserReadModeratedChannels = "user:read:moderated_channels"
	ScopeUserReadSubscriptions     = "user:read:subscriptions"
	ScopeUserReadWhispers          = "user:read:whispers"
	ScopeUserWriteChat             = "user:write:chat"

	// Whispers scope
	ScopeWhispersRead = "whispers:read"
)

// CommonScopes provides commonly used scope combinations.
var CommonScopes = struct {
	Chat        []string
	Moderation  []string
	Channel     []string
	Bot         []string
	Analytics   []string
	Broadcaster []string
}{
	Chat: []string{
		ScopeChatRead,
		ScopeChatEdit,
		ScopeUserWriteChat,
		ScopeUserReadChat,
	},
	Moderation: []string{
		ScopeModerationRead,
		ScopeModeratorManageBannedUsers,
		ScopeModeratorManageChatMessages,
		ScopeModeratorReadChatters,
		ScopeModeratorManageAnnouncements,
	},
	Channel: []string{
		ScopeChannelManageBroadcast,
		ScopeChannelReadEditors,
		ScopeChannelReadSubscriptions,
		ScopeChannelManagePolls,
		ScopeChannelManagePredictions,
	},
	Bot: []string{
		ScopeChatRead,
		ScopeChatEdit,
		ScopeChannelBot,
		ScopeUserBot,
		ScopeUserWriteChat,
		ScopeUserReadChat,
		ScopeModeratorReadChatters,
	},
	Analytics: []string{
		ScopeAnalyticsReadExtensions,
		ScopeAnalyticsReadGames,
	},
	Broadcaster: []string{
		ScopeChannelManageBroadcast,
		ScopeChannelManagePolls,
		ScopeChannelManagePredictions,
		ScopeChannelManageRaids,
		ScopeChannelManageSchedule,
		ScopeChannelManageVideos,
		ScopeChannelReadEditors,
		ScopeChannelReadGoals,
		ScopeChannelReadHypeTrain,
		ScopeChannelReadPolls,
		ScopeChannelReadPredictions,
		ScopeChannelReadSubscriptions,
		ScopeChannelEditCommercial,
		ScopeModerationRead,
		ScopeModeratorManageBannedUsers,
		ScopeModeratorManageChatMessages,
		ScopeModeratorManageChatSettings,
		ScopeModeratorManageAnnouncements,
		ScopeModeratorReadChatters,
		ScopeClipsEdit,
	},
}
