package messages

// Request. Unsubscribe to one or several topics.
type Unsubscribe struct {
	// Event type. Should be 'unsubscribe'
	Event string `json:"event"`
	// Optional - client originated ID reflected in response message.
	//
	// A zero value means no user reference ID will be provided.
	ReqId int64 `json:"reqid,omitempty"`
	// Optional - Array of currency pairs. Format of each pair is "A/B", where A and B are
	// ISO 4217-A3 for standardized assets and popular unique symbol if not standardized.
	Pairs []string `json:"pair,omitempty"`
	// Subscription details
	Subscription UnsuscribeDetails `json:"subscription"`
}

// Subscription details
type UnsuscribeDetails struct {
	// Optional - depth associated with book subscription in number of levels each side. Cf. DepthEnum for values.
	//
	// Default to 10. A zero value will trigger default behavior.
	Depth int `json:"depth,omitempty"`
	// Optional - Time interval associated with ohlc subscription in minutes. Cf. IntervalEnum for values.
	//
	// Default to 1 minute. A zero value will trigger default behavior.
	Interval int `json:"interval,omitempty"`
	// Name of the channel to subscribe to. Cf. ChannelEnum for values.
	Name string `json:"name"`
	// Optional - base64-encoded authentication token for private-data endpoints.
	Token string `json:"token,omitempty"`
}
