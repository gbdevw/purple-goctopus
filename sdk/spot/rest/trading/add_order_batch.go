package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
)

// AddOrderBatch request parameters
type AddOrderBatchRequestParameters struct {
	// Asset pair related to orders
	Pair string `json:"pair"`
	// List of orders
	Orders []Order `json:"orders"`
}

// AddOrderBatch optional parameters
type AddOrderBatchRequestOptions struct {
	// Validate inputs only. Do not submit order.
	Validate bool `json:"validate,omitempty"`
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	//
	// A zero value means no deadline.
	Deadline time.Time `json:"deadline,omitempty"`
}

// AddOrderBatch Entry
type AddOrderBatchEntry struct {
	// Order description
	Description *OrderDescription `json:"descr,omitempty"`
	// Transaction ID for order if added successfully
	Id string `json:"txid,omitempty"`
	// Error message for the order
	Error string `json:"error,omitempty"`
}

// AddOrderBatch Result
type AddOrderBatchResult struct {
	// Entries
	Orders []AddOrderBatchEntry `json:"orders"`
}

// Response for Add Order Batch
type AddOrderBatchResponse struct {
	common.KrakenSpotRESTResponse
	Result *AddOrderBatchResult `json:"result"`
}
