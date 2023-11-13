package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// AddOrder required parameters
type AddOrderParameters struct {
	// Asset pair related to order
	Pair string
	// Order data
	Order Order
}

// AddOrder optional parameters
type AddOrderOptions struct {
	// Validate inputs only. Do not submit order.
	Validate bool
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	// A nil value means no deadline.
	Deadline *time.Time
}

// AddOrder Result
type AddOrderResult struct {
	// Order description
	Description OrderDescription `json:"descr"`
	// Transaction IDs for order
	TransactionIDs []string `json:"txid"`
}

// Response for Add Order
type AddOrderResponse struct {
	common.KrakenSpotRESTResponse
	Result *AddOrderResult `json:"result,omitempty"`
}
