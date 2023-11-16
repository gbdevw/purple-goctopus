package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetStatusOfRecentWithdrawals request options
type GetStatusOfRecentWithdrawalsRequestOptions struct {
	// Filter for specific name of withdrawal method.
	//
	// An empty string means no filter.
	Method string `json:"method"`
	// Filter for specific asset being withdrawn.
	//
	// An empty string means no filter.
	Asset string `json:"asset"`
}

// GetStatusOfRecentWithdrawals response
type GetStatusOfRecentWithdrawalsResponse struct {
	common.KrakenSpotRESTResponse
	// Recent withdrawals
	Result []TransactionDetails `json:"result"`
}
