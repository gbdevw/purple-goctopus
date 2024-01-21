package messages

// Cancel order request message
type CancelOrderRequest struct {
	// Event type. Should be cancelOrder
	Event string `json:"event"`
	// Session token string
	Token string `json:"token"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	//
	// A zero value means request id is not used.
	RequestId int64 `json:"reqid,omitempty"`
	// Array of order IDs to be canceled. These can be user reference IDs.
	TxId []string `json:"txid"`
}

// Cancel order response message
type CancelOrderResponse struct {
	// Event type. Should be cancelOrderStatus
	Event string `json:"event"`
	// Optional - client originated requestID sent as acknowledgment in the message response
	RequestId *int64 `json:"reqid,omitempty"`
	// Status. "ok" or "error". Cf. AddOrderStatusEnum for values.
	Status string `json:"status"`
	// Error message (if unsuccessful)
	Err string `json:"errorMessage,omitempty"`
}
