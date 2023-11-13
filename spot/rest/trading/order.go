package trading

// Enum for order types
type OrderType string

// Values for OrderType
const (
	Market          OrderType = "market"
	Limit           OrderType = "limit"
	StopLoss        OrderType = "stop-loss"
	TakeProfit      OrderType = "take-profit"
	StopLossLimit   OrderType = "stop-loss-limit"
	TakeProfitLimit OrderType = "take-profit-limit"
	SettlePosition  OrderType = "settle-position"
)

// Enum for trigger types
type Trigger string

// Values for Trigger
const (
	Last  Trigger = "last"
	Index Trigger = "index"
)

// Enum for self trade prevention flags
type SelfTradePreventionFlag string

// Values for SelfTradePreventionFlag
const (
	STPCancelNewest SelfTradePreventionFlag = "cancel-newest"
	STPCancelOldest SelfTradePreventionFlag = "cancel-oldest"
	STPCancelBoth   SelfTradePreventionFlag = "cancel-both"
)

// Enum for order flags
type OrderFlag string

// Values for order flags
const (
	OFlagPost                    OrderFlag = "post"
	OFlagFeeInBase               OrderFlag = "fcib"
	OFlagFeeInQuote              OrderFlag = "fciq"
	OFlagNoMarketPriceProtection OrderFlag = "nompp"
	OFlagVolumeInQuote           OrderFlag = "viqc"
)

// Enum for time in force flags
type TimeInForce string

// Values for TimeInForce
const (
	GoodTilCanceled   TimeInForce = "GTC"
	ImmediateOrCancel TimeInForce = "IOC"
	GoodTilDate       TimeInForce = "GTD"
)

// Order description info
type OrderDescription struct {
	// Order description
	Order string `json:"order"`
	// Conditional close order description. Empty if not applicable
	Close string `json:"close"`
}

// Conditional close orders are triggered by execution of the primary order in the same quantity
// and opposite direction, but once triggered are independent orders that may reduce or increase net position.
type CloseOrder struct {
	// Close order type.
	// Valid types are "limit", "stop-loss", "take-profit", "stop-loss-limit", "take-profit-limit"
	OrderType string `json:"ordertype"`
	// Price for limit orders or trigger price for stop-loss(-limit) and take-profit(-limit) orders.
	// Price can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	Price string `json:"price"`
	// Limit price for stop-loss-limit and take-profit-limit orders.
	// Price2 can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	// Price2 is ignored if an empty value is provided.
	Price2 string `json:"price2,omitempty"`
}

// Order data
type Order struct {
	// userref is an optional user-specified integer id that can be associated with any number of orders.
	// Will be ignored if a nil value is provided.
	UserReference *int64 `json:"userref,omitempty"`
	// Order type
	OrderType string `json:"ordertype"`
	// Order direction - buy/sell
	Type string `json:"type"`
	// Order quantity in terms of the base asset.
	// "0" can be provided for closing margin orders to automatically fill the requisite quantity.
	Volume string `json:"volume"`
	// Price for limit orders or trigger price for stop-loss(-limit) and take-profit(-limit) orders.
	// Price can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	// Price is ignored if an empty value is provided.
	Price string `json:"price,omitempty"`
	// Limit price for stop-loss-limit and take-profit-limit orders.
	// Price2 can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	// Price2 is ignored if an empty value is provided.
	Price2 string `json:"price2,omitempty"`
	// Price signal used to trigger stop and take orders.
	// Will be ignored if an empty value is provided.
	// Default behavior if apply is "last"
	Trigger string `json:"trigger,omitempty"`
	// Amount of leverage desired expressed in a formated string "<leverage>:1".
	// Will be ignored if empty.
	Leverage string `json:"leverage,omitempty"`
	// If true, order will only reduce a currently open position, not increase it or open a new position.
	ReduceOnly bool `json:"reduce_only"`
	// Self trade prevention flag.
	// Will be ignored if an empty value is provided.
	// By default cancel-newest behavior is used.
	StpType string `json:"stp_type,omitempty"`
	// Comma delimited list of order flags.
	// Will be ignored if an empty value is provided.
	OrderFlags string `json:"oflags,omitempty"`
	// Time in force flag.
	// Will be ignored if an empty value is provided.
	// An empty value means default Good Til Canceled behavior.
	TimeInForce string `json:"timeinforce,omitempty"`
	// Scheduled start time. Can be empty to trigger default behavior.
	// A value of 0 means now. Default behavior.
	// A value prefixed with + like +<n> schedules start time n seconds from now.
	// Other values are considered as an absolute unix timestamp for start time.
	ScheduledStartTime string `json:"starttm,omitempty"`
	// Expiration time. Can be empty to trigger default behavior
	// A value prefixed with + like +<n> schedules expiration time n seconds from now. Minimum +5 seconds.
	// Other values are considered as an absolute unix timestamp for exiration time.
	ExpirationTime string `json:"expiretm,omitempty"`
	// Close order
	// A nil value means no close order
	Close *CloseOrder `json:"close,omitempty"`
}
