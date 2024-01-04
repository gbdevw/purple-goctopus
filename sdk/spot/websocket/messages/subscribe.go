package messages

// Request. Subscribe to a topic on a single or multiple currency pairs.
type Subscribe struct {
	// Event type. Should be 'subscribe'.
	Event string `json:"event"`
	// Optional - client originated ID reflected in response message.
	//
	// A zero value means no request ID will be provided.
	ReqId int64 `json:"reqid,omitempty"`
	// Optional - Array of currency pairs. Format of each pair is "A/B", where A and B are
	// ISO 4217-A3 for standardized assets and popular unique symbol if not standardized.
	Pairs []string `json:"pair,omitempty"`
	// Subscription details
	Subscription SuscribeDetails `json:"subscription"`
}

// Subscription details
type SuscribeDetails struct {
	// Optional - depth associated with book subscription in number of levels each side. Cf. DepthEnum for values.
	//
	// Default to a depth of 10. A zero value will trigger default behavior.
	Depth int `json:"depth,omitempty"`
	// Optional - Time interval associated with ohlc subscription in minutes. Cf. IntervalEnum for values.
	//
	// Default to 1 minute. A zero value will trigger default behavior.
	Interval int `json:"interval,omitempty"`
	// Name of the channel to subscribe to. Cf. ChannelEnum for values.
	//
	// ChannelAll will subscribe to all channels available in the target environment (eg. public | private)
	Name string `json:"name"`
	// Optional - whether to send rate-limit counter in updates (supported only for openOrders subscriptions)
	//
	// Defaut to false.
	RateCounter bool `json:"ratecounter,omitempty"`
	// Optional - whether to send historical feed data snapshot upon subscription (supported only for
	// ownTrades subscriptions)
	//
	// Default to true. A nil value means default behavior will apply.
	Snapshot *bool `json:"snapshot,omitempty"`
	// Optional - base64-encoded authentication token for private-data endpoints.
	//
	// An empty string means no token will be suppied.
	Token string `json:"token,omitempty"`
	// Optional - for ownTrades, whether to consolidate order fills by root taker trade(s).
	// If false, all order fills will show separately.
	//
	// Default to true. A nil value means default behavior will apply.
	ConsolidateTaker *bool `json:"consolidate_taker,omitempty"`
}
