package websocket

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
)

// Container for active subscriptions that must be maintained by the websocket client.
type activeSubscriptions struct {
	// ticker subscription. Will be nil if ticker topic has never been subscribed to.
	ticker *tickerSubscription
	// OHLC subscriptions by interval. Will be nil if ohlc topic has never been subscribed to.
	ohlcs map[messages.IntervalEnum]*ohlcSubscription
	// trade subscription. Will be nil if trade topic has never been subscribed to.
	trade *tradeSubscription
	// spread subscription. Will be nil if not subscribed.
	spread *spreadSubscription
	// book subscriptions per depth. Will be nil if book topic has never been subscribed to.
	book *bookSubscription
	// ownTrades subscription. Will be nil if not subscribed.
	ownTrades *ownTradesSubscription
	// openOrders subscription. Will be nil if not subscribed.
	openOrders *openOrdersSubscription
	// Heartbeat channel
	heartbeat chan event.Event
	// SystemStatus channel
	systemStatus chan event.Event
}

// Data of a ticker subscription
type tickerSubscription struct {
	// Pairs to subscribe to
	pairs []string
	// Channel used to publish subscription's messages
	pub chan event.Event
}

// Data of a ohlc subscription
type ohlcSubscription struct {
	// Pairs to subscribe to
	pairs []string
	// Desired interval
	interval messages.IntervalEnum
	// Channel used to publish subscription's messages
	pub chan event.Event
}

// Data of a trade subscription
type tradeSubscription struct {
	// Pairs to subscribe to
	pairs []string
	// Channel used to publish subscription's messages
	pub chan event.Event
}

// Data of a spread subscription
type spreadSubscription struct {
	// Pairs to subscribe to
	pairs []string
	// Channel used to publish subscription's messages
	pub chan event.Event
}

// Data of a book subscription
type bookSubscription struct {
	// Pairs to subscribe to
	pairs []string
	// Channel used to publish bok snapshots and updates
	pub chan event.Event
	// Desired depth
	depth messages.DepthEnum
}

// Data of a ownTrades subscription
type ownTradesSubscription struct {
	// Channel used to publish subscription's messages
	pub chan event.Event
	// Desired consolidateTaker value for the subscription
	consolidateTaker bool
	// Desired snapshot value for the subscription
	snapshot bool
}

// Data of a ownTrades subscription
type openOrdersSubscription struct {
	// Channel used to publish subscription's messages
	pub chan event.Event
	// Desired ratecounter value for the subscription
	rateCounter bool
}
