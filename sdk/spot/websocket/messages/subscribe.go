package messages

// Request. Subscribe to a topic on a single or multiple currency pairs.
type Subscribe struct {
	// Event type.
	Event string `json:"event"`
	// Optional - client originated ID reflected in response message.
	ReqId int `json:"reqid,omitempty"`
	// Optional - Array of currency pairs. Format of each pair is "A/B", where A and B are
	// ISO 4217-A3 for standardized assets and popular unique symbol if not standardized.
	Pairs []string `json:"pair,omitempty"`
	// Subscription details
	Subscription Subscription `json:"subscription"`
}

// Subscription details
type Subscription struct {
	// Optional - depth associated with book subscription in number of levels each side,
	// default 10. Cf DepthEnum for values.
	//
	// A zero value will trigger default behavior.
	Depth int `json:"depth,omitempty"`
	// Optional - Time interval associated with ohlc subscription in minutes. Default 1.
	// Cf IntervalEnum for values.
	//
	// A zero value will trigger default behavior.
	Interval int `json:"interval,omitempty"`
	// Name of the channel to subscribe to. Cf. ChannelEnum for values.
	Name string `json:"name"`
	// Optional - whether to send rate-limit counter in updates (supported only for openOrders
	// subscriptions; default = false)
	RateCounter bool `json:"ratecounter,omitempty"`
	// Optional - whether to send historical feed data snapshot upon subscription (supported only
	// for ownTrades subscriptions; default = true).
	//
	// A nil value means this feature is not meant to be used and default behavior will apply.
	Snapshot *bool `json:"snapshot,omitempty"`
	// Optional - base64-encoded authentication token for private-data endpoints.
	Token string `json:"token,omitempty"`
	// Optional - for ownTrades, whether to consolidate order fills by root taker trade(s),
	// default = true. If false, all order fills will show separately.
	//
	// A nil value means this feature is not meant to be used and default behavior will apply.
	ConsolidateTaker *bool `json:"consolidate_taker,omitempty"`
}
