// This package contains definitions of messages exchanged when interacting
// with Kraken spot websocket API.
package messages

// Enum for the event types supported on Kraken spot websocket API.
type EventTypeEnum string

// Values for EventTypeEnum
const (
	EventTypePing               EventTypeEnum = "ping"
	EventTypePong               EventTypeEnum = "pong"
	EventTypeHeartbeat          EventTypeEnum = "heartbeat"
	EventTypeSystemStatus       EventTypeEnum = "systemStatus"
	EventTypeSubscribe          EventTypeEnum = "subscribe"
	EventTypeUnsubscribe        EventTypeEnum = "unsubscribe"
	EventTypeSubscriptionStatus EventTypeEnum = "subscriptionStatus"
)

// Enum for the API statuses
type StatusEnum string

// Values for StatusEnum
const (
	StatusOnline      StatusEnum = "online"
	StatusMaintenance StatusEnum = "maintenance"
	StatusCancelOnly  StatusEnum = "cancel_only"
	StatusLimitOnly   StatusEnum = "limit_only"
	StatusPostOnly    StatusEnum = "post_only"
)

// Enum for the channels supported by the websocket API
type ChannelEnum string

// Values for ChannelEnum
const (
	ChannelAll        ChannelEnum = "*"
	ChannelBook       ChannelEnum = "book"
	ChannelOHLC       ChannelEnum = "ohlc"
	ChannelOpenOrders ChannelEnum = "openOrders"
	ChannelOwnTrades  ChannelEnum = "ownTrades"
	ChannelSpread     ChannelEnum = "spread"
	ChannelTicker     ChannelEnum = "ticker"
	ChannelTrade      ChannelEnum = "trade"
)

// Enum for interval used in subscription messages
type IntervalEnum int

// Values IntervalEnum
const (
	M1     IntervalEnum = 1
	M5     IntervalEnum = 5
	M15    IntervalEnum = 15
	M30    IntervalEnum = 30
	M60    IntervalEnum = 60
	M240   IntervalEnum = 240
	M1440  IntervalEnum = 1440
	M10080 IntervalEnum = 10080
	M21600 IntervalEnum = 21600
)

// Enum for depth used in subscription messages
type DepthEnum int

// Values DepthEnum
const (
	D10   DepthEnum = 10
	D25   DepthEnum = 25
	D100  DepthEnum = 100
	D500  DepthEnum = 500
	D1000 DepthEnum = 1000
)

// Enum for subscription status
type SubscriptionStatusEnum string

// Values for SubscriptionStatusEnum
const (
	Subscribed   SubscriptionStatusEnum = "subscribed"
	Unsubscribed SubscriptionStatusEnum = "unsubscribed"
	Error        SubscriptionStatusEnum = "error"
)

// Enum for trades triggering side
type TriggeringSideEnum string

// Values for TriggeringSideEnum
const (
	Buy  TriggeringSideEnum = "buy"
	Sell TriggeringSideEnum = "sell"
)

// Enum for trades triggering order type
type TriggeringOrderTypeEnum string

// Values for TriggeringOrderTypeEnum
const (
	Market TriggeringOrderTypeEnum = "market"
	Limit  TriggeringOrderTypeEnum = "limit"
)
