package trading

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// EditOrder required parameters
type EditOrderParameters struct {
	// Original Order ID or User Reference Id (userref) which is user-specified
	// integer id used with the original order. If userref is not unique and was
	// used with multiple order, edit request is denied with an error.
	Id string
	// Asset Pair
	Pair string
}

// EditOrder optional parameters
type EditOrderOptions struct {
	// New user reference id. Userref from parent order will
	// not be retained on the new order after edit.
	// An empty value means that data must not be changed.
	NewUserReference string
	// Order quantity in terms of the base asset.
	// A nil value means that data must not be changed.
	NewVolume string
	// New limit price or trigger price. Either price or price2
	// can be preceded by +, -, or # to specify the order price
	// as an offset relative to the last traded price. + adds
	// the amount to, and - subtracts the amount from the last
	// traded price. # will either add or subtract the amount to
	// the last traded price, depending on the direction and order
	// type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	// An empty value means that data must not be changed.
	Price string
	// New limit price for stop/take-limit order. Either price or price2
	// can be preceded by +, -, or # to specify the order price
	// as an offset relative to the last traded price. + adds
	// the amount to, and - subtracts the amount from the last
	// traded price. # will either add or subtract the amount to
	// the last traded price, depending on the direction and order
	// type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	// An empty value means that data must not be changed.
	Price2 string
	// List of order flags. Only these flags can be
	// changed: - post post-only order (available when ordertype =
	// limit). All the flags from the parent order are retained except
	// post-only. post-only needs to be explicitly mentioned on edit request.
	// A nil value means that data must not be changed.
	OFlags []string
	// Validate inputs only. Do not submit order.
	Validate bool
	// Used to interpret if client wants to receive pending replace,
	// before the order is completely replaced
	CancelResponse bool
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	// A nil value means no deadline.
	Deadline *time.Time
}

// EditOrder Result
type EditOrderResult struct {
	// Order description
	Description OrderDescription `json:"descr"`
	// New transaction ID
	TransactionID string `json:"txid"`
	// New user reference.
	// Will be nil if no user ref was provided with request.
	NewUserReference *int64 `json:"newuserref"`
	// Old user reference
	// Will be nil if no user ref was provided with request to create the original order.
	OldUserReference *int64 `json:"olduserref"`
	// Number of orders canceled
	OrdersCancelled int `json:"orders_cancelled"`
	// Original transaction ID
	OriginalTransactionID string `json:"originaltxid"`
	// Status of the order. Either "ok" or "err"
	Status string `json:"status"`
	// Updated volume
	Volume string `json:"volume"`
	// Updated Price
	Price string `json:"price"`
	// Updated Price2
	Price2 string `json:"price2"`
	// Error message if unsuccessful
	ErrorMsg string `json:"error_message"`
}

// EditOrder Response
type EditOrderResponse struct {
	common.KrakenSpotRESTResponse
	Result *EditOrderResult `json:"result,omitempty"`
}
