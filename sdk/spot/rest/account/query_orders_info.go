package account

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// QueryOrdersInfo request parameters.
type QueryOrdersInfoParameters struct {
	// List of transaction IDs to query info about (50 maximum).
	TxId []string `json:"txid"`
}

// QueryOrdersInfo request options.
type QueryOrdersInfoRequestOptions struct {
	// Whether or not to include trades related to position in output.
	//
	// Defaults to false.
	Trades bool `json:"trades"`
	// Restrict results to given user reference id.
	//
	// A nil value means no user reference will be proided.
	UserReference *int64
	// Whether or not to consolidate trades by individual taker trades.
	//
	// Default to true. A nil value triggers default behavior.
	ConsolidateTaker *bool `json:"consolidate_taker,omitempty"`
}

// QueryOrdersInfo response.
type QueryOrdersInfoResponse struct {
	common.KrakenSpotRESTResponse
	// Map where keys are transaction ID and values the requested orders
	Result map[string]*OrderInfo `json:"result,omitempty"`
}
