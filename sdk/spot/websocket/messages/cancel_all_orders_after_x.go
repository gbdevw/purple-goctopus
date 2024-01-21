package messages

// Cancel all orders after X request message
type CancelAllOrdersAfterXRequest struct {
	// Event type. Should be cancelAllOrdersAfter
	Event string `json:"event"`
	// Session token string
	Token string `json:"token"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	//
	// A zero value means request id is not used.
	RequestId int64 `json:"reqid,omitempty"`
	// Timeout specified in seconds. 0 to disable the timer.
	Timeout int `json:"timeout"`
}

// Cancel all orders after X response message
type CancelAllOrdersAfterXResponse struct {
	// Event type. Should be cancelAllOrdersAfterStatus
	Event string `json:"event"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	RequestId *int64 `json:"reqid,omitempty"`
	// Status. "ok" or "error". Cf. AddOrderStatusEnum for values.
	Status string `json:"status"`
	// Timestamp (RFC3339) reflecting when the request has been handled (second precision, rounded up)
	CurrentTime string `json:"currentTime,omitempty"`
	// Timestamp (RFC3339) reflecting the time at which all open orders will be cancelled, unless the timer is extended or disabled (second precision, rounded up)
	TriggerTime string `json:"triggerTime,omitempty"`
	// Error message (if unsuccessful)
	Err string `json:"errorMessage,omitempty"`
}
