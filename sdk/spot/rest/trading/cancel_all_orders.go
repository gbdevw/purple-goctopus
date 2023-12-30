package trading

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// CancelAllOrders result
type CancelAllOrdersResult struct {
	// Number of canceled orders
	Count int `json:"count"`
}

// CancelAllOrders response
type CancelAllOrdersResponse struct {
	common.KrakenSpotRESTResponse
	Result *CancelAllOrdersResult `json:"result,omitempty"`
}
