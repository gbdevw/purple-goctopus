package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetOpenOrdersOptions contains Get Open Orders optional parameters.
type GetOpenOrdersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Restrict results to given user reference id. A nil value means no restrictions.
	UserReference *int64
}

// Open orders
type OpenOrders struct {
	// Keys are transaction ID and values the related open order.
	Open map[string]OrderInfo `json:"open"`
}

// GetOpenOrdersResponse contains Get Open Orders response data.
type GetOpenOrdersResponse struct {
	common.KrakenSpotRESTResponse
	Result *OpenOrders `json:"result,omitempty"`
}
