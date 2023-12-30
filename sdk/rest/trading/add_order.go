package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// AddOrder request parameters
type AddOrderRequestParameters struct {
	// Asset pair related to order
	Pair string `json:"pair"`
	// Order data
	Order Order `json:"order"`
}

// AddOrder request options.
type AddOrderRequestOptions struct {
	// Validate inputs only. Do not submit order.
	Validate bool `json:"validate"`
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	//
	// A zero value means no deadline.
	Deadline time.Time `json:"deadline,omitempty"`
}

// AddOrder result
type AddOrderResult struct {
	// Order description
	Description OrderDescription `json:"descr"`
	// Transaction IDs for order
	TransactionIDs []string `json:"txid"`
}

// AddOrder response
type AddOrderResponse struct {
	common.KrakenSpotRESTResponse
	Result *AddOrderResult `json:"result,omitempty"`
}
