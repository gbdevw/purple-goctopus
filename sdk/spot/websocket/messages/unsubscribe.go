package messages

// Request. Subscribe to a topic on a single or multiple currency pairs.
type Unsubscribe struct {
	// Event type.
	Event string `json:"event"`
	// Optional - client originated ID reflected in response message.
	ReqId int `json:"reqid,omitempty"`
	// Optional - Array of currency pairs. Format of each pair is "A/B", where A and B are
	// ISO 4217-A3 for standardized assets and popular unique symbol if not standardized.
	Pairs []string `json:"pair,omitempty"`
	// Subscription details
	Subscription UnsuscribeDetails `json:"subscription"`
}

// Subscription details
type UnsuscribeDetails struct {
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
	// Optional - base64-encoded authentication token for private-data endpoints.
	Token string `json:"token,omitempty"`
}
