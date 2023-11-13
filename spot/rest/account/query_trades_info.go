package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// QueryTradesParameters contains Query Trades required parameters
type QueryTradesParameters struct {
	// List of transaction IDs to query info about (20 maximum).
	TransactionIds []string
}

// QueryTradesOptions contains Query Trades optional parameters.
type QueryTradesOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
}

// QueryTradesInfoResponse contains Query Trades Info response data.
type QueryTradesInfoResponse struct {
	common.KrakenSpotRESTResponse
	// Map where keys are transaction ID and values the requested trades.
	Result map[string]TradeInfo `json:"result,omitempty"`
}
