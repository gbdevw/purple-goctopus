package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetOpenOrders request options.
type GetOpenOrdersRequestOptions struct {
	// Whether or not to include trades related to position in output.
	//
	// Defaults to false.
	Trades bool
	// Restrict results to given user reference id.
	//
	// A nil value means no restrictions.
	UserReference *int64
}

// GetOpenOrders result
type GetOpenOrdersResult struct {
	// Keys are transaction IDs and values are the related open order.
	Open map[string]*OrderInfo `json:"open"`
}

// GetOpenOrders response
type GetOpenOrdersResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetOpenOrdersResult `json:"result,omitempty"`
}
