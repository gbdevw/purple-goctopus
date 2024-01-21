package trading

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// CancelOrder request parameters
type CancelOrderRequestParameters struct {
	// Open order transaction ID (txid) or user reference (userref)
	Id string `json:"id"`
}

// CancelOrder result
type CancelOrderResult struct {
	// Number of canceled orders
	Count int `json:"count"`
}

// CancelOrder response
type CancelOrderResponse struct {
	common.KrakenSpotRESTResponse
	Result *CancelOrderResult `json:"result,omitempty"`
}
