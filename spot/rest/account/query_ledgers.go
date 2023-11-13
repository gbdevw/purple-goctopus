package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// QueryLedgersParameters contains Query Ledgers required parameters.
type QueryLedgersParameters struct {
	// List of ledger IDs to query info about (20 maximum).
	LedgerIds []string
}

// QueryLedgersOptions contains Query Ledgers optional parameters.
type QueryLedgersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
}

// QueryLedgersResponse contains Query Ledgers response data.
type QueryLedgersResponse struct {
	common.KrakenSpotRESTResponse
	// Key are ledger entry IDs and values are ledger entries
	Result map[string]LedgerEntry `json:"result,omitempty"`
}
