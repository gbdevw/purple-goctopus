package websocket

// AddOrder request parameters
type AddOrderRequestParameters struct {
	// Order type. Cf. OrderTypeEnum for values.
	OrderType string `json:"ordertype"`
	// Side, buy or sell. Cf. SideEnum for values.
	Type string `json:"type"`
	// Currency pair.
	Pair string `json:"pair"`
	// Order volume in base currency
	Volume string `json:"volume"`
	// Amount of leverage desired.
	//
	// A zero value means no leverage.
	Leverage int `json:"leverage,omitempty"`
	// If true, order will only reduce a currently open position, not increase it or open a new position.
	ReduceOnly bool `json:"reduce_only,omitempty"`
	// Optional comma delimited list of order flags. Cf. OrderFlagEnum for values.
	//
	// viqc = volume in quote currency (not currently available), fcib = prefer fee in base currency, fciq = prefer fee in quote currency,
	// nompp = no market price protection, post = post only order (available when ordertype = limit)
	//
	// An empty string means no order flags to provide.
	OFlags string `json:"oflags,omitempty"`
	// Optional - scheduled start time.
	//
	// Values can be:
	//	- 0 = now (default)
	//	- +<n> = schedule start time <n> seconds from now
	//	- <n> = unix timestamp of start time
	//
	// An empty string triggers the default behavior (startm = now)
	StartTimestamp string `json:"starttm,omitempty"`
	// Optional - expiration time.
	//
	// Values can be:
	//	- 0 = no expiration (default)
	//	- +<n> = schedule start time <n> seconds from now
	//	- <n> = unix timestamp of start time
	//
	// An empty string triggers the default behavior (no expiration)
	ExpireTimestamp string `json:"expiretm,omitempty"`
	// Optional - RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which matching engine should reject new order request,4 in presence of latency
	// or order queueing. min now + 2 seconds, max now + 60 seconds. Defaults to now + 60 seconds if not specified.
	//
	// An empty string triggers the default behavior.
	Deadline string `json:"deadline,omitempty"`
	// Optional - user reference ID (should be an integer in quotes)
	UserReference string `json:"userref,omitempty"`
	// Optional - if true, validate inputs only; do not submit order.
	//
	// Default to false.
	Validate bool `json:"validate,omitempty"`
	// Optional close order type. Cf. OrderTypeEnum
	CloseOrderType string `json:"close[ordertype],omitempty"`
	// Optional - close order price.
	ClosePrice string `json:"close[price],omitempty"`
	// Optional - close order secondary price.
	ClosePrice2 string `json:"close[price2],omitempty"`
	// Optional - time in force. Cf. TimeInForceEnum for values.
	//
	// Default to GTC (good-til-cancelled). An empty string triggers the default behavior.
	TimeInForce string `json:"timeinforce,omitempty"`
}
