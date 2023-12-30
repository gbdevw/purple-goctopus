package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// QueryLedgers request parameters.
type QueryLedgersRequestParameters struct {
	// List of ledger IDs to query info about (20 maximum).
	Id []string `json:"id"`
}

// QueryLedgers request options.
type QueryLedgersRequestOptions struct {
	// Whether or not to include trades related to position in output.
	//
	// Defaults to false.
	Trades bool `json:"trades"`
}

// QueryLedgers response.
type QueryLedgersResponse struct {
	common.KrakenSpotRESTResponse
	// Key are ledger entry IDs and values are ledger entries.
	Result map[string]*LedgerEntry `json:"result,omitempty"`
}
