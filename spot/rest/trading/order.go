package trading

// Enum for order types
type OrderTypeEnum string

// Values for OrderTypeEnum
const (
	Market          OrderTypeEnum = "market"
	Limit           OrderTypeEnum = "limit"
	StopLoss        OrderTypeEnum = "stop-loss"
	TakeProfit      OrderTypeEnum = "take-profit"
	StopLossLimit   OrderTypeEnum = "stop-loss-limit"
	TakeProfitLimit OrderTypeEnum = "take-profit-limit"
	SettlePosition  OrderTypeEnum = "settle-position"
)

// Enum for trigger types
type TriggerEnum string

// Values for TriggerEnum
const (
	Last  TriggerEnum = "last"
	Index TriggerEnum = "index"
)

// Enum for self trade prevention flags
type SelfTradePreventionFlagEnum string

// Values for SelfTradePreventionFlagEnum
const (
	STPCancelNewest SelfTradePreventionFlagEnum = "cancel-newest"
	STPCancelOldest SelfTradePreventionFlagEnum = "cancel-oldest"
	STPCancelBoth   SelfTradePreventionFlagEnum = "cancel-both"
)

// Enum for order flags
type OrderFlagEnum string

// Values for OrderFlagEnum
const (
	OFlagPost                    OrderFlagEnum = "post"
	OFlagFeeInBase               OrderFlagEnum = "fcib"
	OFlagFeeInQuote              OrderFlagEnum = "fciq"
	OFlagNoMarketPriceProtection OrderFlagEnum = "nompp"
	OFlagVolumeInQuote           OrderFlagEnum = "viqc"
)

// Enum for time in force flags
type TimeInForceEnum string

// Values for TimeInForceEnum
const (
	GoodTilCanceled   TimeInForceEnum = "GTC"
	ImmediateOrCancel TimeInForceEnum = "IOC"
	GoodTilDate       TimeInForceEnum = "GTD"
)

// Order description info
type OrderDescription struct {
	// Order description
	Order string `json:"order"`
	// Conditional close order description. Empty if not applicable
	Close string `json:"close,omitempty"`
}

// Conditional close orders are triggered by execution of the primary order in the same quantity
// and opposite direction, but once triggered are independent orders that may reduce or increase net position.
type CloseOrder struct {
	// Close order type.
	//
	// Valid types are "limit", "stop-loss", "take-profit", "stop-loss-limit", "take-profit-limit"
	OrderType string `json:"ordertype"`
	// # Description
	//
	// Price:
	//	- Limit price for limit orders
	//	- Trigger price for stop-loss, stop-loss-limit, take-profit and take-profit-limit orders
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
	// Secondary price:
	//	- Limit price for stop-loss-limit and take-profit-limit orders
	//
	// An empty value means no secondary price.
	//
	// # Note
	//
	// Either price or price2 can be preceded by +, -, or # to specify the order price as an offset
	// relative to the last traded price. + adds the amount to, and - subtracts the amount from the
	// last traded price. # will either add or subtract the amount to the last traded price,
	// depending on the direction and order type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	Price2 string `json:"price2,omitempty"`
}

// Order data
type Order struct {
	// userref is an optional user-specified integer id that can be associated with any number of orders.
	//
	// Will be ignored if a nil value is provided.
	UserReference *int64 `json:"userref,omitempty"`
	// Order type
	OrderType string `json:"ordertype"`
	// Order direction - buy/sell
	Type string `json:"type"`
	// Order quantity in terms of the base asset.
	//
	// "0" can be provided for closing margin orders to automatically fill the requisite quantity.
	Volume string `json:"volume"`
	// # Description
	//
	// Price:
	//	- Limit price for limit orders
	//	- Trigger price for stop-loss, stop-loss-limit, take-profit and take-profit-limit orders
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
	// Secondary price:
	//	- Limit price for stop-loss-limit and take-profit-limit orders
	//
	// An empty value means no secondary price.
	//
	// # Note
	//
	// Either price or price2 can be preceded by +, -, or # to specify the order price as an offset
	// relative to the last traded price. + adds the amount to, and - subtracts the amount from the
	// last traded price. # will either add or subtract the amount to the last traded price,
	// depending on the direction and order type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	Price2 string `json:"price2,omitempty"`
	// Price signal used to trigger stop and take orders.
	//
	// Default behavior if apply is "last". An empty value triggers default behavior.
	Trigger string `json:"trigger,omitempty"`
	// Amount of leverage desired expressed in a formated string "<leverage>:1".
	//
	// Will be ignored if empty.
	Leverage string `json:"leverage,omitempty"`
	// If true, order will only reduce a currently open position, not increase it or open a new position.
	ReduceOnly bool `json:"reduce_only"`
	// Self trade prevention flag.
	//
	// By default cancel-newest behavior is used. An empty value triggers default behavior.
	StpType string `json:"stp_type,omitempty"`
	// Comma delimited list of order flags.
	// Will be ignored if an empty value is provided.
	OrderFlags string `json:"oflags,omitempty"`
	// Time in force flag.
	//
	// An empty value means default Good Til Canceled behavior.
	TimeInForce string `json:"timeinforce,omitempty"`
	// Scheduled start time.
	//	- A value of 0 means now. Default behavior.
	//	- A value prefixed with + like +<n> schedules start time n seconds from now.
	//	- Other values are considered as an absolute unix timestamp for start time.
	//
	// An empty value triggers default behavior (now)
	ScheduledStartTime string `json:"starttm,omitempty"`
	// Expiration time.
	//	- 0 means no expiration (default behavior)
	//	- A value prefixed with + like +<n> schedules expiration time n seconds from now. Minimum +5 seconds.
	//	- Other values are considered as an absolute unix timestamp for exiration time.
	//
	//  An empty value triggers default behavior (now)
	ExpirationTime string `json:"expiretm,omitempty"`
	// Optional close order.
	//
	// A nil value means no close order
	Close *CloseOrder `json:"close,omitempty"`
}
