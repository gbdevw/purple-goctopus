package trading

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Cancel Order required parameters
type CancelOrderParameters struct {
	// Open order transaction ID (txid) or user reference (userref)
	Id string
}

// CancelOrder Result
type CancelOrderResult struct {
	// Number of canceled orders
	Count int `json:"count"`
	// If set, order(s) is/are pending cancellation
	Pending bool `json:"pending"`
}

// Response for Cancel Order
type CancelOrderResponse struct {
	common.KrakenSpotRESTResponse
	Result *CancelOrderResult `json:"result,omitempty"`
}
