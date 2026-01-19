package helix

import (
	"context"
	"net/url"
	"time"
)

// BannedUser represents a banned user.
type BannedUser struct {
	UserID         string    `json:"user_id"`
	UserLogin      string    `json:"user_login"`
	UserName       string    `json:"user_name"`
	ExpiresAt      time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	Reason         string    `json:"reason"`
	ModeratorID    string    `json:"moderator_id"`
	ModeratorLogin string    `json:"moderator_login"`
	ModeratorName  string    `json:"moderator_name"`
}

// GetBannedUsersParams contains parameters for GetBannedUsers.
type GetBannedUsersParams struct {
	BroadcasterID string
	UserIDs       []string // Filter by user IDs (max 100)
	*PaginationParams
}

// GetBannedUsers gets the list of banned users for a channel.
// Requires: moderation:read scope.
func (c *Client) GetBannedUsers(ctx context.Context, params *GetBannedUsersParams) (*Response[BannedUser], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.UserIDs {
		q.Add("user_id", id)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[BannedUser]
	if err := c.get(ctx, "/moderation/banned", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BanUserParams contains parameters for BanUser.
type BanUserParams struct {
	BroadcasterID string      `json:"-"`
	ModeratorID   string      `json:"-"`
	Data          BanUserData `json:"data"`
}

// BanUserData contains the ban data.
type BanUserData struct {
	UserID   string `json:"user_id"`
	Duration int    `json:"duration,omitempty"` // 0 = permanent, 1-1209600 seconds
	Reason   string `json:"reason,omitempty"`
}

// BanUserResponse represents the response from BanUser.
type BanUserResponse struct {
	BroadcasterID string    `json:"broadcaster_id"`
	ModeratorID   string    `json:"moderator_id"`
	UserID        string    `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	EndTime       time.Time `json:"end_time,omitempty"`
}

// BanUser bans a user from a channel.
// Requires: moderator:manage:banned_users scope.
func (c *Client) BanUser(ctx context.Context, params *BanUserParams) (*BanUserResponse, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	var resp Response[BanUserResponse]
	if err := c.post(ctx, "/moderation/bans", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UnbanUser unbans a user from a channel.
// Requires: moderator:manage:banned_users scope.
func (c *Client) UnbanUser(ctx context.Context, broadcasterID, moderatorID, userID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)
	q.Set("user_id", userID)

	return c.delete(ctx, "/moderation/bans", q, nil)
}

// Moderator represents a channel moderator.
type Moderator struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// GetModeratorsParams contains parameters for GetModerators.
type GetModeratorsParams struct {
	BroadcasterID string
	UserIDs       []string // Filter by user IDs (max 100)
	*PaginationParams
}

// GetModerators gets the list of moderators for a channel.
// Requires: moderation:read scope.
func (c *Client) GetModerators(ctx context.Context, params *GetModeratorsParams) (*Response[Moderator], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.UserIDs {
		q.Add("user_id", id)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[Moderator]
	if err := c.get(ctx, "/moderation/moderators", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddChannelModerator adds a moderator to a channel.
// Requires: channel:manage:moderators scope.
func (c *Client) AddChannelModerator(ctx context.Context, broadcasterID, userID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("user_id", userID)

	return c.post(ctx, "/moderation/moderators", q, nil, nil)
}

// RemoveChannelModerator removes a moderator from a channel.
// Requires: channel:manage:moderators scope.
func (c *Client) RemoveChannelModerator(ctx context.Context, broadcasterID, userID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("user_id", userID)

	return c.delete(ctx, "/moderation/moderators", q, nil)
}

// DeleteChatMessagesParams contains parameters for DeleteChatMessages.
type DeleteChatMessagesParams struct {
	BroadcasterID string
	ModeratorID   string
	MessageID     string // Optional: Specific message to delete. If empty, deletes all messages.
}

// DeleteChatMessages deletes chat messages.
// Requires: moderator:manage:chat_messages scope.
func (c *Client) DeleteChatMessages(ctx context.Context, params *DeleteChatMessagesParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	if params.MessageID != "" {
		q.Set("message_id", params.MessageID)
	}

	return c.delete(ctx, "/moderation/chat", q, nil)
}

// BlockedTerm represents a blocked term.
type BlockedTerm struct {
	BroadcasterID string    `json:"broadcaster_id"`
	ModeratorID   string    `json:"moderator_id"`
	ID            string    `json:"id"`
	Text          string    `json:"text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
}

// GetBlockedTermsParams contains parameters for GetBlockedTerms.
type GetBlockedTermsParams struct {
	BroadcasterID string
	ModeratorID   string
	*PaginationParams
}

// GetBlockedTerms gets the list of blocked terms for a channel.
// Requires: moderator:read:blocked_terms scope.
func (c *Client) GetBlockedTerms(ctx context.Context, params *GetBlockedTermsParams) (*Response[BlockedTerm], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[BlockedTerm]
	if err := c.get(ctx, "/moderation/blocked_terms", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddBlockedTermParams contains parameters for AddBlockedTerm.
type AddBlockedTermParams struct {
	BroadcasterID string `json:"-"`
	ModeratorID   string `json:"-"`
	Text          string `json:"text"`
}

// AddBlockedTerm adds a blocked term.
// Requires: moderator:manage:blocked_terms scope.
func (c *Client) AddBlockedTerm(ctx context.Context, params *AddBlockedTermParams) (*BlockedTerm, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	var resp Response[BlockedTerm]
	if err := c.post(ctx, "/moderation/blocked_terms", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// RemoveBlockedTerm removes a blocked term.
// Requires: moderator:manage:blocked_terms scope.
func (c *Client) RemoveBlockedTerm(ctx context.Context, broadcasterID, moderatorID, termID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)
	q.Set("id", termID)

	return c.delete(ctx, "/moderation/blocked_terms", q, nil)
}

// ShieldModeStatus represents shield mode status.
type ShieldModeStatus struct {
	IsActive        bool      `json:"is_active"`
	ModeratorID     string    `json:"moderator_id"`
	ModeratorLogin  string    `json:"moderator_login"`
	ModeratorName   string    `json:"moderator_name"`
	LastActivatedAt time.Time `json:"last_activated_at"`
}

// GetShieldModeStatus gets the shield mode status for a channel.
// Requires: moderator:read:shield_mode scope.
func (c *Client) GetShieldModeStatus(ctx context.Context, broadcasterID, moderatorID string) (*ShieldModeStatus, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)

	var resp Response[ShieldModeStatus]
	if err := c.get(ctx, "/moderation/shield_mode", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UpdateShieldModeStatusParams contains parameters for UpdateShieldModeStatus.
type UpdateShieldModeStatusParams struct {
	BroadcasterID string `json:"-"`
	ModeratorID   string `json:"-"`
	IsActive      bool   `json:"is_active"`
}

// UpdateShieldModeStatus updates the shield mode status for a channel.
// Requires: moderator:manage:shield_mode scope.
func (c *Client) UpdateShieldModeStatus(ctx context.Context, params *UpdateShieldModeStatusParams) (*ShieldModeStatus, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	var resp Response[ShieldModeStatus]
	if err := c.put(ctx, "/moderation/shield_mode", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// WarnChatUserParams contains parameters for WarnChatUser.
type WarnChatUserParams struct {
	BroadcasterID string           `json:"-"`
	ModeratorID   string           `json:"-"`
	Data          WarnChatUserData `json:"data"`
}

// WarnChatUserData contains the warning data.
type WarnChatUserData struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

// WarnChatUser warns a user in chat.
// Requires: moderator:manage:warnings scope.
func (c *Client) WarnChatUser(ctx context.Context, params *WarnChatUserParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	return c.post(ctx, "/moderation/warnings", q, params, nil)
}

// AutoModStatus represents the AutoMod status of a message.
type AutoModStatus struct {
	MsgID       string `json:"msg_id"`
	IsPermitted bool   `json:"is_permitted"`
}

// CheckAutoModStatusParams contains parameters for CheckAutoModStatus.
type CheckAutoModStatusParams struct {
	BroadcasterID string                 `json:"-"`
	Data          []AutoModStatusMessage `json:"data"`
}

// AutoModStatusMessage represents a message to check.
type AutoModStatusMessage struct {
	MsgID   string `json:"msg_id"`
	MsgText string `json:"msg_text"`
}

// CheckAutoModStatus checks if messages meet AutoMod requirements.
// Requires: moderation:read scope.
func (c *Client) CheckAutoModStatus(ctx context.Context, params *CheckAutoModStatusParams) (*Response[AutoModStatus], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)

	var resp Response[AutoModStatus]
	if err := c.post(ctx, "/moderation/enforcements/status", q, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ManageHeldAutoModMessageParams contains parameters for ManageHeldAutoModMessages.
type ManageHeldAutoModMessageParams struct {
	UserID string `json:"user_id"`
	MsgID  string `json:"msg_id"`
	Action string `json:"action"` // ALLOW or DENY
}

// ManageHeldAutoModMessages approves or denies a held AutoMod message.
// Requires: moderator:manage:automod scope.
func (c *Client) ManageHeldAutoModMessages(ctx context.Context, params *ManageHeldAutoModMessageParams) error {
	return c.post(ctx, "/moderation/automod/message", nil, params, nil)
}

// AutoModSettings represents AutoMod settings.
type AutoModSettings struct {
	BroadcasterID           string `json:"broadcaster_id"`
	ModeratorID             string `json:"moderator_id"`
	OverallLevel            *int   `json:"overall_level"`
	Disability              int    `json:"disability"`
	Aggression              int    `json:"aggression"`
	SexualitySexOrGender    int    `json:"sexuality_sex_or_gender"`
	Misogyny                int    `json:"misogyny"`
	Bullying                int    `json:"bullying"`
	Swearing                int    `json:"swearing"`
	RaceEthnicityOrReligion int    `json:"race_ethnicity_or_religion"`
	SexBasedTerms           int    `json:"sex_based_terms"`
}

// GetAutoModSettings gets the AutoMod settings for a channel.
// Requires: moderator:read:automod_settings scope.
func (c *Client) GetAutoModSettings(ctx context.Context, broadcasterID, moderatorID string) (*AutoModSettings, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)

	var resp Response[AutoModSettings]
	if err := c.get(ctx, "/moderation/automod/settings", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UpdateAutoModSettingsParams contains parameters for UpdateAutoModSettings.
type UpdateAutoModSettingsParams struct {
	BroadcasterID           string `json:"-"`
	ModeratorID             string `json:"-"`
	OverallLevel            *int   `json:"overall_level,omitempty"`
	Disability              *int   `json:"disability,omitempty"`
	Aggression              *int   `json:"aggression,omitempty"`
	SexualitySexOrGender    *int   `json:"sexuality_sex_or_gender,omitempty"`
	Misogyny                *int   `json:"misogyny,omitempty"`
	Bullying                *int   `json:"bullying,omitempty"`
	Swearing                *int   `json:"swearing,omitempty"`
	RaceEthnicityOrReligion *int   `json:"race_ethnicity_or_religion,omitempty"`
	SexBasedTerms           *int   `json:"sex_based_terms,omitempty"`
}

// UpdateAutoModSettings updates the AutoMod settings for a channel.
// Requires: moderator:manage:automod_settings scope.
func (c *Client) UpdateAutoModSettings(ctx context.Context, params *UpdateAutoModSettingsParams) (*AutoModSettings, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	var resp Response[AutoModSettings]
	if err := c.put(ctx, "/moderation/automod/settings", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UnbanRequest represents an unban request.
type UnbanRequest struct {
	ID               string `json:"id"`
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	ModeratorID      string `json:"moderator_id,omitempty"`
	ModeratorLogin   string `json:"moderator_login,omitempty"`
	ModeratorName    string `json:"moderator_name,omitempty"`
	UserID           string `json:"user_id"`
	UserLogin        string `json:"user_login"`
	UserName         string `json:"user_name"`
	Text             string `json:"text"`
	Status           string `json:"status"` // pending, approved, denied, acknowledged, canceled
	CreatedAt        string `json:"created_at"`
	ResolvedAt       string `json:"resolved_at,omitempty"`
	ResolutionText   string `json:"resolution_text,omitempty"`
}

// GetUnbanRequestsParams contains parameters for GetUnbanRequests.
type GetUnbanRequestsParams struct {
	BroadcasterID string
	ModeratorID   string
	Status        string // pending, approved, denied, acknowledged, canceled
	UserID        string // Filter by user
	*PaginationParams
}

// GetUnbanRequests gets unban requests for a channel.
// Requires: moderator:read:unban_requests scope.
func (c *Client) GetUnbanRequests(ctx context.Context, params *GetUnbanRequestsParams) (*Response[UnbanRequest], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.UserID != "" {
		q.Set("user_id", params.UserID)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[UnbanRequest]
	if err := c.get(ctx, "/moderation/unban_requests", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ResolveUnbanRequestParams contains parameters for ResolveUnbanRequest.
type ResolveUnbanRequestParams struct {
	BroadcasterID  string `json:"-"`
	ModeratorID    string `json:"-"`
	UnbanRequestID string `json:"-"`
	Status         string `json:"status"` // approved or denied
	ResolutionText string `json:"resolution_text,omitempty"`
}

// ResolveUnbanRequest resolves an unban request.
// Requires: moderator:manage:unban_requests scope.
func (c *Client) ResolveUnbanRequest(ctx context.Context, params *ResolveUnbanRequestParams) (*UnbanRequest, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	q.Set("unban_request_id", params.UnbanRequestID)
	q.Set("status", params.Status)
	if params.ResolutionText != "" {
		q.Set("resolution_text", params.ResolutionText)
	}

	var resp Response[UnbanRequest]
	if err := c.patch(ctx, "/moderation/unban_requests", q, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// ModeratedChannel represents a channel where the user is a moderator.
type ModeratedChannel struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
}

// GetModeratedChannelsParams contains parameters for GetModeratedChannels.
type GetModeratedChannelsParams struct {
	UserID string
	*PaginationParams
}

// GetModeratedChannels gets channels where the user is a moderator.
// Requires: user:read:moderated_channels scope.
func (c *Client) GetModeratedChannels(ctx context.Context, params *GetModeratedChannelsParams) (*Response[ModeratedChannel], error) {
	q := url.Values{}
	q.Set("user_id", params.UserID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[ModeratedChannel]
	if err := c.get(ctx, "/moderation/channels", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SuspiciousUserStatus represents the status of a suspicious user.
type SuspiciousUserStatus string

const (
	// SuspiciousUserStatusRestricted indicates the user is restricted from chatting.
	SuspiciousUserStatusRestricted SuspiciousUserStatus = "restricted"
	// SuspiciousUserStatusMonitored indicates the user is being monitored.
	SuspiciousUserStatusMonitored SuspiciousUserStatus = "monitored"
)

// AddSuspiciousUserStatusParams contains parameters for AddSuspiciousUserStatus.
type AddSuspiciousUserStatusParams struct {
	BroadcasterID string               `json:"-"`
	ModeratorID   string               `json:"-"`
	UserID        string               `json:"user_id"`
	Status        SuspiciousUserStatus `json:"status"`
}

// AddSuspiciousUserStatus adds a suspicious status to a chat user.
// The status can be "restricted" or "monitored".
// Requires: moderator:manage:suspicious_users scope.
func (c *Client) AddSuspiciousUserStatus(ctx context.Context, params *AddSuspiciousUserStatusParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)

	return c.post(ctx, "/moderation/suspicious_users", q, params, nil)
}

// RemoveSuspiciousUserStatusParams contains parameters for RemoveSuspiciousUserStatus.
type RemoveSuspiciousUserStatusParams struct {
	BroadcasterID string `json:"-"`
	ModeratorID   string `json:"-"`
	UserID        string `json:"-"`
}

// RemoveSuspiciousUserStatus removes a suspicious status from a chat user.
// Requires: moderator:manage:suspicious_users scope.
func (c *Client) RemoveSuspiciousUserStatus(ctx context.Context, params *RemoveSuspiciousUserStatusParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	q.Set("user_id", params.UserID)

	return c.delete(ctx, "/moderation/suspicious_users", q, nil)
}
