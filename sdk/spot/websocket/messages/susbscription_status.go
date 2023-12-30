package messages

// Response. Subscription status response to subscribe, unsubscribe or exchange initiated unsubscribe.
type SubscriptionStatus struct {
	// Channel Name on successful subscription. For payloads 'ohlc' and 'book', respective interval or
	// depth will be added as suffix.
	ChannelName string `json:"channelName,omitempty"`
	// Event type.
	Event string `json:"event"`
	// Optional - client originated ID reflected in response message.
	ReqId int `json:"reqid,omitempty"`
	// Optional - Currency pair, applicable to public messages only.
	Pair string `json:"pair,omitempty"`
	// Status of subscription. Cf. SubscriptionStatusEnum for values.
	Status string `json:"status,omitempty"`
	// Error message if any.
	Err string `json:"errorMessage,omitempty"`
	// Subscription status details.
	Subscription *SubscriptionStatusDetails `json:"subscription,omitempty"`
}

// Subscription status details
type SubscriptionStatusDetails struct {
	// Optional - depth associated with book subscription in number of levels each side
	Depth int `json:"depth,omitempty"`
	// Optional - Time interval associated with ohlc subscription in minutes.
	Interval int `json:"interval,omitempty"`
	// Optional - max rate-limit budget. Compare to the ratecounter field in the openOrders updates to
	// check whether you are approaching the rate limit.
	MaxRateCount int `json:"maxratecount,omitempty"`
	// Name of the subscribed channel.
	Name string `json:"name"`
	// Optional - base64-encoded authentication token for private-data endpoints.
	Token string `json:"token,omitempty"`
}
