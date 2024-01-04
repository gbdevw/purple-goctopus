package messages

// Cancel all orders request message
type CancelAllOrdersRequest struct {
	// Event type. Should be cancelAll
	Event string `json:"event"`
	// Session token string
	Token string `json:"token"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	//
	// A zero value means request id is not used.
	RequestId int64 `json:"reqid,omitempty"`
}

// Cancel all orders response message
type CancelAllOrdersResponse struct {
	// Event type. Should be cancelAllStatus
	Event string `json:"event"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	RequestId *int64 `json:"reqid,omitempty"`
	// Number of orders cancelled.
	Count int `json:"count"`
	// Status. "ok" or "error". Cf. AddOrderStatusEnum for values.
	Status string `json:"status"`
	// Error message (if unsuccessful)
	Err string `json:"errorMessage,omitempty"`
}
