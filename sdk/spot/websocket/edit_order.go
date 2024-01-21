package websocket

// EditOrder request parameters
//
// At least one of the optional edittable data must be set.
type EditOrderRequestParameters struct {
	// Original Order ID or userref.
	Id string `json:"orderid"`
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
	//
	// Default to false.
	Validate bool `json:"validate,omitempty"`
}
