package account

import "encoding/json"

// Enum for sides
type SideEnum string

// Value for Side
const (
	Buy  SideEnum = "buy"
	Sell SideEnum = "sell"
)

// Enum for order statuses
type OrderStatusEnum string

// Values for OrderStatus
const (
	Pending  OrderStatusEnum = "pending"
	Open     OrderStatusEnum = "open"
	Closed   OrderStatusEnum = "closed"
	Canceled OrderStatusEnum = "canceled"
	Expired  OrderStatusEnum = "expired"
)

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

// Description for a Order Info
type OrderInfoDescription struct {
	// Asset pair
	Pair string `json:"pair,omitempty"`
	// Order direction (buy/sell). Cf. SideEnum.
	Type string `json:"type,omitempty"`
	// Order type. Cf. OrderTypeEnum
	OrderType string `json:"ordertype,omitempty"`
	// Limit or trigger price depending on order type
	Price json.Number `json:"price,omitempty"`
	// Limit price for stop/take orders
	Price2 json.Number `json:"price2,omitempty"`
	// Amount of leverage
	Leverage string `json:"leverage,omitempty"`
	// Textual order description
	OrderDescription string `json:"order,omitempty"`
	// Conditional close order description
	CloseOrderDescription string `json:"close,omitempty"`
}

// OrderInfo contains order data.
type OrderInfo struct {
	// Referral order transaction ID that created this order
	ReferralOrderTransactionId string `json:"refid,omitempty"`
	// Optional user defined reference ID
	UserReferenceId json.Number `json:"userref,omitempty"`
	// Status of order. Cf. OrderStatusEnum
	Status string `json:"status"`
	// Unix timestamp of when order was placed.
	//
	// Unix seconds timestamp with nanoseconds as decimal part (ex: 1688666559.8974)
	OpenTimestamp json.Number `json:"opentm,omitempty"`
	// Unix timestamp of order start time (or 0 if not set)
	StartTimestamp json.Number `json:"starttm,omitempty"`
	// Unix timestamp of order end time (or 0 if not set)
	ExpireTimestamp json.Number `json:"expiretm,omitempty"`
	// Order description info
	Description OrderInfoDescription `json:"descr"`
	// Volume of order (base currency)
	Volume json.Number `json:"vol,omitempty"`
	// Volume executed (base currency)
	VolumeExecuted json.Number `json:"vol_exec,omitempty"`
	// Total cost (quote currency unless)
	Cost json.Number `json:"cost,omitempty"`
	// Total fee  (quote currency)
	Fee json.Number `json:"fee,omitempty"`
	// Average price  (quote currency)
	Price json.Number `json:"price,omitempty"`
	// Stop price  (quote currency)
	StopPrice json.Number `json:"stopprice,omitempty"`
	// Triggered limit price  (quote currency, when limit based order type triggered)
	LimitPrice json.Number `json:"limitprice,omitempty"`
	// Price signal used to trigger "stop-loss" "take-profit" "stop-loss-limit" "take-profit-limit" orders.
	//
	// Cf. TriggerEnum. 'last' is the implied trigger if field is not set.
	Trigger string `json:"trigger,omitempty"`
	// Comma delimited list of miscellaneous info
	Miscellaneous string `json:"misc,omitempty"`
	// Comma delimited list of order flags
	OrderFlags string `json:"oflags,omitempty"`
	// List of trade IDs related to order (if trades info requested and data available)
	Trades []string `json:"trades,omitempty"`
	// If order is closed, Unix timestamp of when order was closed
	CloseTimestamp json.Number `json:"closetm,omitempty"`
	// Additional info on status if any
	Reason string `json:"reason,omitempty"`
}
