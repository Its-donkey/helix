package helix

import (
	"context"
	"net/url"
	"strconv"
)

// ExtensionConfigurationSegment represents a configuration segment.
type ExtensionConfigurationSegment struct {
	Segment       string `json:"segment"` // broadcaster, developer, global
	BroadcasterID string `json:"broadcaster_id,omitempty"`
	Content       string `json:"content"`
	Version       string `json:"version"`
}

// GetExtensionConfigurationSegmentParams contains parameters for GetExtensionConfigurationSegment.
type GetExtensionConfigurationSegmentParams struct {
	ExtensionID   string
	Segment       []string // broadcaster, developer, global
	BroadcasterID string   // Required if segment includes "broadcaster"
}

// GetExtensionConfigurationSegment gets extension configuration segment data.
// Requires: JWT created by extension.
func (c *Client) GetExtensionConfigurationSegment(ctx context.Context, params *GetExtensionConfigurationSegmentParams) (*Response[ExtensionConfigurationSegment], error) {
	q := url.Values{}
	q.Set("extension_id", params.ExtensionID)
	for _, seg := range params.Segment {
		q.Add("segment", seg)
	}
	if params.BroadcasterID != "" {
		q.Set("broadcaster_id", params.BroadcasterID)
	}

	var resp Response[ExtensionConfigurationSegment]
	if err := c.get(ctx, "/extensions/configurations", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetExtensionConfigurationSegmentParams contains parameters for SetExtensionConfigurationSegment.
type SetExtensionConfigurationSegmentParams struct {
	ExtensionID   string `json:"extension_id"`
	Segment       string `json:"segment"` // broadcaster, developer, global
	BroadcasterID string `json:"broadcaster_id,omitempty"`
	Content       string `json:"content,omitempty"`
	Version       string `json:"version,omitempty"`
}

// SetExtensionConfigurationSegment sets extension configuration segment data.
// Requires: JWT created by extension.
func (c *Client) SetExtensionConfigurationSegment(ctx context.Context, params *SetExtensionConfigurationSegmentParams) error {
	return c.put(ctx, "/extensions/configurations", nil, params, nil)
}

// SetExtensionRequiredConfigurationParams contains parameters for SetExtensionRequiredConfiguration.
type SetExtensionRequiredConfigurationParams struct {
	BroadcasterID         string `json:"-"`
	ExtensionID           string `json:"extension_id"`
	ExtensionVersion      string `json:"extension_version"`
	RequiredConfiguration string `json:"required_configuration"`
}

// SetExtensionRequiredConfiguration sets the required configuration for an extension.
// Requires: JWT created by extension.
func (c *Client) SetExtensionRequiredConfiguration(ctx context.Context, params *SetExtensionRequiredConfigurationParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)

	return c.put(ctx, "/extensions/required_configuration", q, params, nil)
}

// SendExtensionPubSubMessageParams contains parameters for SendExtensionPubSubMessage.
type SendExtensionPubSubMessageParams struct {
	Target            []string `json:"target"` // broadcast, global, whisper-<user_id>
	BroadcasterID     string   `json:"broadcaster_id"`
	IsGlobalBroadcast bool     `json:"is_global_broadcast,omitempty"`
	Message           string   `json:"message"`
}

// SendExtensionPubSubMessage sends a PubSub message for an extension.
// Requires: JWT created by extension.
func (c *Client) SendExtensionPubSubMessage(ctx context.Context, params *SendExtensionPubSubMessageParams) error {
	return c.post(ctx, "/extensions/pubsub", nil, params, nil)
}

// ExtensionLiveChannel represents a live channel using an extension.
type ExtensionLiveChannel struct {
	BroadcasterID   string `json:"broadcaster_id"`
	BroadcasterName string `json:"broadcaster_name"`
	GameName        string `json:"game_name"`
	GameID          string `json:"game_id"`
	Title           string `json:"title"`
}

// GetExtensionLiveChannelsParams contains parameters for GetExtensionLiveChannels.
type GetExtensionLiveChannelsParams struct {
	ExtensionID string
	*PaginationParams
}

// GetExtensionLiveChannels gets live channels that have an extension installed and activated.
func (c *Client) GetExtensionLiveChannels(ctx context.Context, params *GetExtensionLiveChannelsParams) (*Response[ExtensionLiveChannel], error) {
	q := url.Values{}
	q.Set("extension_id", params.ExtensionID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[ExtensionLiveChannel]
	if err := c.get(ctx, "/extensions/live", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExtensionSecret represents an extension secret.
type ExtensionSecret struct {
	FormatVersion int                   `json:"format_version"`
	Secrets       []ExtensionSecretData `json:"secrets"`
}

// ExtensionSecretData represents the secret data.
type ExtensionSecretData struct {
	Content   string `json:"content"`
	ActiveAt  string `json:"active_at"`
	ExpiresAt string `json:"expires_at"`
}

// GetExtensionSecrets gets the secrets for an extension.
// Requires: JWT created by extension.
func (c *Client) GetExtensionSecrets(ctx context.Context, extensionID string) (*Response[ExtensionSecret], error) {
	q := url.Values{}
	q.Set("extension_id", extensionID)

	var resp Response[ExtensionSecret]
	if err := c.get(ctx, "/extensions/jwt/secrets", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateExtensionSecret creates a new secret for an extension.
// Requires: JWT created by extension.
func (c *Client) CreateExtensionSecret(ctx context.Context, extensionID string, delay int) (*Response[ExtensionSecret], error) {
	q := url.Values{}
	q.Set("extension_id", extensionID)
	if delay > 0 {
		q.Set("delay", strconv.Itoa(delay))
	}

	var resp Response[ExtensionSecret]
	if err := c.post(ctx, "/extensions/jwt/secrets", q, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendExtensionChatMessageParams contains parameters for SendExtensionChatMessage.
type SendExtensionChatMessageParams struct {
	BroadcasterID    string `json:"broadcaster_id"`
	Text             string `json:"text"`
	ExtensionID      string `json:"extension_id"`
	ExtensionVersion string `json:"extension_version"`
}

// SendExtensionChatMessage sends a chat message from an extension.
// Requires: JWT created by extension.
func (c *Client) SendExtensionChatMessage(ctx context.Context, params *SendExtensionChatMessageParams) error {
	return c.post(ctx, "/extensions/chat", nil, params, nil)
}

// Extension represents a Twitch extension.
type Extension struct {
	AuthorName                string            `json:"author_name"`
	BitsEnabled               bool              `json:"bits_enabled"`
	CanInstall                bool              `json:"can_install"`
	ConfigurationLocation     string            `json:"configuration_location"`
	Description               string            `json:"description"`
	EULAToSURL                string            `json:"eula_tos_url"`
	HasChatSupport            bool              `json:"has_chat_support"`
	IconURL                   string            `json:"icon_url"`
	IconURLs                  map[string]string `json:"icon_urls"`
	ID                        string            `json:"id"`
	Name                      string            `json:"name"`
	PrivacyPolicyURL          string            `json:"privacy_policy_url"`
	RequestIdentityLink       bool              `json:"request_identity_link"`
	ScreenshotURLs            []string          `json:"screenshot_urls"`
	State                     string            `json:"state"`
	SubscriptionsSupportLevel string            `json:"subscriptions_support_level"`
	Summary                   string            `json:"summary"`
	SupportEmail              string            `json:"support_email"`
	Version                   string            `json:"version"`
	ViewerSummary             string            `json:"viewer_summary"`
	Views                     ExtensionViews    `json:"views"`
	AllowlistedConfigURLs     []string          `json:"allowlisted_config_urls"`
	AllowlistedPanelURLs      []string          `json:"allowlisted_panel_urls"`
}

// ExtensionViews represents the views configuration for an extension.
type ExtensionViews struct {
	Mobile       ExtensionView `json:"mobile,omitempty"`
	Panel        ExtensionView `json:"panel,omitempty"`
	VideoOverlay ExtensionView `json:"video_overlay,omitempty"`
	Component    ExtensionView `json:"component,omitempty"`
}

// ExtensionView represents a single view configuration.
type ExtensionView struct {
	ViewerURL              string `json:"viewer_url"`
	Height                 int    `json:"height,omitempty"`
	CanLinkExternalContent bool   `json:"can_link_external_content,omitempty"`
}

// GetExtensions gets information about extensions.
// Requires: JWT created by extension.
func (c *Client) GetExtensions(ctx context.Context, extensionID, extensionVersion string) (*Response[Extension], error) {
	q := url.Values{}
	q.Set("extension_id", extensionID)
	if extensionVersion != "" {
		q.Set("extension_version", extensionVersion)
	}

	var resp Response[Extension]
	if err := c.get(ctx, "/extensions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetReleasedExtensions gets information about a released extension.
func (c *Client) GetReleasedExtensions(ctx context.Context, extensionID, extensionVersion string) (*Response[Extension], error) {
	q := url.Values{}
	q.Set("extension_id", extensionID)
	if extensionVersion != "" {
		q.Set("extension_version", extensionVersion)
	}

	var resp Response[Extension]
	if err := c.get(ctx, "/extensions/released", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExtensionBitsProduct represents an extension Bits product.
type ExtensionBitsProduct struct {
	SKU           string            `json:"sku"`
	Cost          ExtensionBitsCost `json:"cost"`
	InDevelopment bool              `json:"in_development"`
	DisplayName   string            `json:"display_name"`
	Expiration    string            `json:"expiration,omitempty"`
	IsBroadcast   bool              `json:"is_broadcast"`
}

// ExtensionBitsCost represents the cost of an extension Bits product.
type ExtensionBitsCost struct {
	Amount int    `json:"amount"`
	Type   string `json:"type"` // bits
}

// GetExtensionBitsProductsParams contains parameters for GetExtensionBitsProducts.
type GetExtensionBitsProductsParams struct {
	ShouldIncludeAll bool // Include disabled/expired products
}

// GetExtensionBitsProducts gets Bits products for an extension.
// Requires: App access token for the extension.
func (c *Client) GetExtensionBitsProducts(ctx context.Context, params *GetExtensionBitsProductsParams) (*Response[ExtensionBitsProduct], error) {
	q := url.Values{}
	if params != nil && params.ShouldIncludeAll {
		q.Set("should_include_all", "true")
	}

	var resp Response[ExtensionBitsProduct]
	if err := c.get(ctx, "/bits/extensions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateExtensionBitsProductParams contains parameters for UpdateExtensionBitsProduct.
type UpdateExtensionBitsProductParams struct {
	SKU           string            `json:"sku"`
	Cost          ExtensionBitsCost `json:"cost"`
	DisplayName   string            `json:"display_name"`
	InDevelopment bool              `json:"in_development,omitempty"`
	Expiration    string            `json:"expiration,omitempty"`
	IsBroadcast   bool              `json:"is_broadcast,omitempty"`
}

// UpdateExtensionBitsProduct updates an extension Bits product.
// Requires: App access token for the extension.
func (c *Client) UpdateExtensionBitsProduct(ctx context.Context, params *UpdateExtensionBitsProductParams) (*Response[ExtensionBitsProduct], error) {
	var resp Response[ExtensionBitsProduct]
	if err := c.put(ctx, "/bits/extensions", nil, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExtensionTransaction represents an extension transaction.
type ExtensionTransaction struct {
	ID               string                      `json:"id"`
	Timestamp        string                      `json:"timestamp"`
	BroadcasterID    string                      `json:"broadcaster_id"`
	BroadcasterLogin string                      `json:"broadcaster_login"`
	BroadcasterName  string                      `json:"broadcaster_name"`
	UserID           string                      `json:"user_id"`
	UserLogin        string                      `json:"user_login"`
	UserName         string                      `json:"user_name"`
	ProductType      string                      `json:"product_type"`
	ProductData      ExtensionTransactionProduct `json:"product_data"`
}

// ExtensionTransactionProduct represents the product in a transaction.
type ExtensionTransactionProduct struct {
	SKU           string            `json:"sku"`
	Cost          ExtensionBitsCost `json:"cost"`
	DisplayName   string            `json:"display_name"`
	InDevelopment bool              `json:"in_development"`
}

// GetExtensionTransactionsParams contains parameters for GetExtensionTransactions.
type GetExtensionTransactionsParams struct {
	ExtensionID string
	IDs         []string // Transaction IDs (max 100)
	*PaginationParams
}

// GetExtensionTransactions gets extension transactions.
// Requires: App access token for the extension.
func (c *Client) GetExtensionTransactions(ctx context.Context, params *GetExtensionTransactionsParams) (*Response[ExtensionTransaction], error) {
	q := url.Values{}
	q.Set("extension_id", params.ExtensionID)
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[ExtensionTransaction]
	if err := c.get(ctx, "/extensions/transactions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
