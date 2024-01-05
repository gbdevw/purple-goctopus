package websocket

// CancelAllOrdersAfterX request parameters
type CancelAllOrdersAfterXRequestParameters struct {
	// Timeout specified in seconds. 0 to disable the timer.
	Timeout int `json:"timeout"`
}
