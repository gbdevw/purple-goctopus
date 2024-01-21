package messages

type EditOrderRequest struct {
	// Event type. Should be addOrder
	Event string `json:"event"`
	// Session token string
	Token string `json:"token"`
	// Original Order ID or userref.
	Id string `json:"orderid"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	//
	// A zero value means request id is not used.
	RequestId int64 `json:"reqid,omitempty"`
	// Currency pair.
	Pair string `json:"pair"`
	// Optional dependent on order type - order price
	//
	// An empty string can be used when price2 is not used.
	Price string `json:"price,omitempty"`
	// Optional dependent on order type - order secondary price
	//
	// An empty string can be used when price2 is not used.
	Price2 string `json:"price2,omitempty"`
	// Order volume in base currency
	Volume string `json:"volume,omitempty"`
	// Optional comma delimited list of order flags. C f. OrderFlagEnum for values.
	//
	// viqc = volume in quote currency (not currently available), fcib = prefer fee in base currency, fciq = prefer fee in quote currency,
	// nompp = no market price protection, post = post only order (available when ordertype = limit)
	//
	// An empty string means no order flags to provide.
	OFlags string `json:"oflags,omitempty"`
	// Optional - user reference ID for new order (should be an integer in quotes)
	//
	// An empty string means no new user reference will be defined.
	NewUserReference string `json:"newuserref,omitempty"`
	// Optional - if true, validate inputs only; do not submit order.
	Validate string `json:"validate,omitempty"`
}

// Response message for EditOrder
type EditOrderResponse struct {
	// Event type. Should be addOrderStatus
	Event string `json:"event"`
	// New order ID if successful
	TxId string `json:"txid,omitempty"`
	// Original order ID if successful
	OriginalTxId string `json:"originaltxid,omitempty"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	RequestId *int64 `json:"reqid,omitempty"`
	// Status. "ok" or "error". Cf. AddOrderStatusEnum for values.
	Status string `json:"status"`
	// New order description info (if successful)
	Description string `json:"descr,omitempty"`
	// Error message (if unsuccessful)
	Err string `json:"errorMessage,omitempty"`
}
