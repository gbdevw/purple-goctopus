package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// QueryTrades request parameters.
type QueryTradesRequestParameters struct {
	// List of transaction IDs to query info about (20 maximum).
	TransactionIds []string `json:"txid"`
}

// QueryTrades request options.
type QueryTradesRequestOptions struct {
	// Whether or not to include trades related to position in output.
	//
	// Defaults to false.
	Trades bool `json:"trades"`
}

// QueryTradesInfo response.
type QueryTradesInfoResponse struct {
	common.KrakenSpotRESTResponse
	// Map where keys are transaction ID and values the requested trades.
	Result map[string]TradeInfo `json:"result,omitempty"`
}
