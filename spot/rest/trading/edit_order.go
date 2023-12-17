package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for status of the edit response.
type EditResponseStatusEnum string

// Values for EditResponseStatusEnum
const (
	Ok    EditResponseStatusEnum = "ok"
	Error EditResponseStatusEnum = "err"
)

// EditOrder request parameters
type EditOrderRequestParameters struct {
	// Original Order ID or User Reference Id (userref) which is user-specified
	// integer id used with the original order. If userref is not unique and was
	// used with multiple order, edit request is denied with an error.
	Id string `json:"txid"`
	// Asset Pair
	Pair string `json:"pair"`
}

// EditOrder request options
type EditOrderRequestOptions struct {
	// New user reference id. Userref from parent order will
	// not be retained on the new order after edit.
	//
	// An empty value means data must not be changed.
	NewUserReference string `json:"userref,omitempty"`
	// Order quantity in terms of the base asset.
	//
	// An empty value means data must not be changed.
	NewVolume string `json:"volume,omitempty"`
	// # Description
	//
	// New price:
	//	- Limit price for limit orders
	//	- Trigger price for stop-loss, stop-loss-limit, take-profit and take-profit-limit orders
	//
	// An empty value means data must not be changed.
	//
	// # Note
	//
	// Either price or price2 can be preceded by +, -, or # to specify the order price as an offset
	// relative to the last traded price. + adds the amount to, and - subtracts the amount from the
	// last traded price. # will either add or subtract the amount to the last traded price,
	// depending on the direction and order type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	Price string `json:"price,omitempty"`
	// # Description
	//
	// New secondary price:
	//	- Limit price for stop-loss-limit and take-profit-limit orders
	//
	// An empty value means data must not be changed.
	//
	// # Note
	//
	// Either price or price2 can be preceded by +, -, or # to specify the order price as an offset
	// relative to the last traded price. + adds the amount to, and - subtracts the amount from the
	// last traded price. # will either add or subtract the amount to the last traded price,
	// depending on the direction and order type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	Price2 string `json:"price2,omitempty"`
	// List of order flags.
	//
	// Only these flags can be changed:
	//	- post post-only order (available when ordertype = limit).
	//
	// All the flags from the parent order are retained except post-only. post-only needs to be
	// explicitly mentioned on edit request.
	//
	// A nil value means that data must not be changed.
	OFlags []string `json:"oflags,omitempty"`
	// Validate inputs only. Do not submit order.
	Validate bool `json:"validate"`
	// Used to interpret if client wants to receive pending replace,
	// before the order is completely replaced.
	CancelResponse bool `json:"cancel_response"`
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	//
	// A zero value means no deadline.
	Deadline time.Time `json:"deadline,omitempty"`
}

// EditOrder Result
type EditOrderResult struct {
	// Order description
	Description *OrderDescription `json:"descr,omitempty"`
	// New transaction ID if order was added successfully.
	TransactionID string `json:"txid,omitempty"`
	// New user reference.
	//
	// Will be nil if no user ref was provided with request.
	NewUserReference *int64 `json:"newuserref,omitempty"`
	// Old user reference.
	//
	// Will be nil if no user ref was provided with request to create the original order.
	OldUserReference *int64 `json:"olduserref,omitempty"`
	// Number of orders canceled
	OrdersCancelled int `json:"orders_cancelled"`
	// Original transaction ID
	OriginalTransactionID string `json:"originaltxid,omitempty"`
	// Status of the order. Either "ok" or "err".
	//
	// Cf. EditResponseStatusEnum
	Status string `json:"status,omitempty"`
	// Updated volume
	Volume string `json:"volume,omitempty"`
	// Updated Price
	Price string `json:"price,omitempty"`
	// Updated Price2
	Price2 string `json:"price2,omitempty"`
	// Error message if unsuccessful
	ErrorMsg string `json:"error_message,omitempty"`
}

// EditOrder Response
type EditOrderResponse struct {
	common.KrakenSpotRESTResponse
	Result *EditOrderResult `json:"result,omitempty"`
}
