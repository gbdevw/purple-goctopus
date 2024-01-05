package websocket

// CancelOrder request parameters
type CancelOrderRequestParameters struct {
	// Array of order IDs to be canceled. These can be user reference IDs.
	TxId []string `json:"txid"`
}
