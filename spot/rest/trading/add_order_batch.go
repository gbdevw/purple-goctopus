package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// AddOrderBatch required parameters
type AddOrderBatchParameters struct {
	// Asset pair related to orders
	Pair string
	// List of orders
	Orders []Order
}

// AddOrderBatch optional parameters
type AddOrderBatchOptions struct {
	// Validate inputs only. Do not submit order.
	Validate bool
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	// A nil value means no deadline.
	Deadline *time.Time
}

// AddOrderBatch Entry
type AddOrderBatchEntry struct {
	// Order description
	Description struct {
		Order string `json:"string,omitempty"`
	} `json:"descr,omitempty"`
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
