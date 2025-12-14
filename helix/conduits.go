package helix

import (
	"context"
	"net/url"
)

// Conduit represents an EventSub conduit.
type Conduit struct {
	ID         string `json:"id"`
	ShardCount int    `json:"shard_count"`
}

// GetConduits gets the conduits for a client ID.
// Requires: App access token.
func (c *Client) GetConduits(ctx context.Context) (*Response[Conduit], error) {
	var resp Response[Conduit]
	if err := c.get(ctx, "/eventsub/conduits", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateConduitParams contains parameters for CreateConduit.
type CreateConduitParams struct {
	ShardCount int `json:"shard_count"`
}

// CreateConduit creates a new conduit.
// Requires: App access token.
func (c *Client) CreateConduit(ctx context.Context, shardCount int) (*Conduit, error) {
	params := CreateConduitParams{ShardCount: shardCount}

	var resp Response[Conduit]
	if err := c.post(ctx, "/eventsub/conduits", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UpdateConduitParams contains parameters for UpdateConduit.
type UpdateConduitParams struct {
	ID         string `json:"id"`
	ShardCount int    `json:"shard_count"`
}

// UpdateConduit updates a conduit's shard count.
// Requires: App access token.
func (c *Client) UpdateConduit(ctx context.Context, params *UpdateConduitParams) (*Conduit, error) {
	var resp Response[Conduit]
	if err := c.patch(ctx, "/eventsub/conduits", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// DeleteConduit deletes a conduit.
// Requires: App access token.
func (c *Client) DeleteConduit(ctx context.Context, conduitID string) error {
	q := url.Values{}
	q.Set("id", conduitID)

	return c.delete(ctx, "/eventsub/conduits", q, nil)
}

// ConduitShard represents a shard in a conduit.
type ConduitShard struct {
	ID        string                `json:"id"`
	Status    string                `json:"status"`
	Transport ConduitShardTransport `json:"transport"`
}

// ConduitShardTransport represents the transport for a conduit shard.
type ConduitShardTransport struct {
	Method         string `json:"method"`
	Callback       string `json:"callback,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	ConnectedAt    string `json:"connected_at,omitempty"`
	DisconnectedAt string `json:"disconnected_at,omitempty"`
}

// GetConduitShardsParams contains parameters for GetConduitShards.
type GetConduitShardsParams struct {
	ConduitID string
	Status    string // enabled, webhook_callback_verification_pending, webhook_callback_verification_failed, notification_failures_exceeded, websocket_disconnected, websocket_failed_ping_pong, websocket_received_inbound_traffic, websocket_internal_error, websocket_network_timeout, websocket_network_error, websocket_failed_to_reconnect
	*PaginationParams
}

// GetConduitShardsResponse represents the response from GetConduitShards.
type GetConduitShardsResponse struct {
	Data       []ConduitShard `json:"data"`
	Pagination *Pagination    `json:"pagination,omitempty"`
}

// GetConduitShards gets the shards for a conduit.
// Requires: App access token.
func (c *Client) GetConduitShards(ctx context.Context, params *GetConduitShardsParams) (*GetConduitShardsResponse, error) {
	q := url.Values{}
	q.Set("conduit_id", params.ConduitID)
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp GetConduitShardsResponse
	if err := c.get(ctx, "/eventsub/conduits/shards", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateConduitShardsParams contains parameters for UpdateConduitShards.
type UpdateConduitShardsParams struct {
	ConduitID string                     `json:"conduit_id"`
	Shards    []UpdateConduitShardParams `json:"shards"`
}

// UpdateConduitShardParams represents parameters for a single shard update.
type UpdateConduitShardParams struct {
	ID        string                      `json:"id"`
	Transport UpdateConduitShardTransport `json:"transport"`
}

// UpdateConduitShardTransport represents transport parameters for shard update.
type UpdateConduitShardTransport struct {
	Method    string `json:"method"`
	Callback  string `json:"callback,omitempty"`
	Secret    string `json:"secret,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// UpdateConduitShardsResponse represents the response from UpdateConduitShards.
type UpdateConduitShardsResponse struct {
	Data   []ConduitShard `json:"data"`
	Errors []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"errors,omitempty"`
}

// UpdateConduitShards updates shards for a conduit.
// Requires: App access token.
func (c *Client) UpdateConduitShards(ctx context.Context, params *UpdateConduitShardsParams) (*UpdateConduitShardsResponse, error) {
	var resp UpdateConduitShardsResponse
	if err := c.patch(ctx, "/eventsub/conduits/shards", nil, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
