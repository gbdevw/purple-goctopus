package trading

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// CancelOrderBatch required parameters
type CancelOrderBatchParameters struct {
	// Open orders transaction ID (txid) or user reference (userref)
	OrderIds []string
}

// Response for Cancel Order Batch
type CancelOrderBatchResponse struct {
	common.KrakenSpotRESTResponse
	Result struct {
		// Number of canceled orders
		Count int `json:"count"`
	} `json:"result"`
}
