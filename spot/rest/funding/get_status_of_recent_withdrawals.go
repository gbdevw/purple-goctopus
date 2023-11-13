package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetStatusOfRecentWithdrawals optional parameters
type GetStatusOfRecentWithdrawalsOptions struct {
	// Filter for specific name of withdrawal method.
	//
	// Defaults to all methods (no filtering). An empty string triggers the default behavior.
	Method string
	// Filter for specific asset being withdrawn.
	//
	// Defaults to all assets (no filtering). An empty string triggers the default behavior.
	Asset string
}

// GetStatusOfRecentWithdrawals response
type GetStatusOfRecentWithdrawalsResponse struct {
	common.KrakenSpotRESTResponse
	// Recent withdrawals
	Result []TransactionDetails `json:"result"`
}
