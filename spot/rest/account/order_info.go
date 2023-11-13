package account

// Enum for order statuses
type OrderStatus string

// Values for OrderStatus
const (
	Pending  OrderStatus = "pending"
	Open     OrderStatus = "open"
	Closed   OrderStatus = "closed"
	Canceled OrderStatus = "canceled"
	Expired  OrderStatus = "expired"
)

// Description for a Order Info
type OrderInfoDescription struct {
	// Asset pair
	Pair string `json:"pair"`
	// Order direction (buy/sell)
	Type string `json:"type"`
	// Order type. Enum: "market" "limit" "stop-loss" "take-profit" "stop-loss-limit" "take-profit-limit" "settle-position"
	OrderType string `json:"ordertype"`
	// Limit or trigger price depending on order type
	Price string `json:"price"`
	// Limit price for stop/take orders
	Price2 string `json:"price2"`
	// Amount of leverage
	Leverage string `json:"leverage"`
	// Textual order description
	OrderDescription string `json:"order"`
	// Conditional close order description
	CloseOrderDescription string `json:"close,omitempty"`
}

// OrderInfo contains order data.
type OrderInfo struct {
	// Referral order transaction ID that created this order
	ReferralOrderTransactionId string `json:"refid"`
	// Optional user defined reference ID
	UserReferenceId string `json:"userref"`
	// Status of order. Enum: "pending" "open" "closed" "canceled" "expired"
	Status string `json:"status"`
	// Unix timestamp of when order was placed
	OpenTimestamp int64 `json:"opentm"`
	// Unix timestamp of order start time (or 0 if not set)
	StartTimestamp int64 `json:"starttm"`
	// Unix timestamp of order end time (or 0 if not set)
	ExpireTimestamp int64 `json:"expiretm"`
	// Order description info
	Description OrderInfoDescription `json:"descr"`
	// Volume of order
	Volume string `json:"vol"`
	// Volume executed
	VolumeExecuted string `json:"vol_exec"`
	// Total cost
	Cost string `json:"cost"`
	// Total fee
	Fee string `json:"fee"`
	// Average price
	Price string `json:"price"`
	// Stop price
	StopPrice string `json:"stopprice"`
	// Triggered limit price
	LimitPrice string `json:"limitprice"`
	// Price signal used to trigger "stop-loss" "take-profit" "stop-loss-limit" "take-profit-limit" orders.
	// Enum: last, index. last is the implied trigger if field is not set.
	Trigger string `json:"trigger"`
	// Comma delimited list of miscellaneous info
	Miscellaneous string `json:"misc"`
	// Comma delimited list of order flags
	OrderFlags string `json:"oflags"`
	// List of trade IDs related to order (if trades info requested and data available)
	Trades []string `json:"trades,omitempty"`
	// If order is closed, Unix timestamp of when order was closed
	CloseTimestamp int64 `json:"closetm,omitempty"`
	// Additional info on status if any
	Reason string `json:"reason,omitempty"`
}
