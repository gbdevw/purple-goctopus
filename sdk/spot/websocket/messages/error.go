package messages

// General error message from the websocket server
type ErrorMessage struct {
	// Event type. Should be 'error'
	Event string `json:"event"`
	// Error message
	Err string `json:"errorMessage"`
	// Optional - client originated ID reflected in response message
	ReqId *int64 `json:"reqid,omitempty"`
}
