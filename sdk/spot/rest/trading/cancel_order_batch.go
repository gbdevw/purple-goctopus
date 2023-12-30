package trading

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// CancelOrderBatch request parameters
type CancelOrderBatchRequestParameters struct {
	// Open orders transaction ID (txid) or user reference (userref)
	OrderIds []string `json:"orders"`
}

// CancelOrderBatch result
type CancelOrderBatchResult struct {
	// Number of canceled orders
	Count int `json:"count"`
}

// CancelOrderBatch response
type CancelOrderBatchResponse struct {
	common.KrakenSpotRESTResponse
	Result *CancelOrderBatchResult `json:"result,omitempty"`
}
