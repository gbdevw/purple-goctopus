package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetStatusOfRecentDeposits required parameters
type GetStatusOfRecentDepositsParameters struct {
	// Asset being deposited
	Asset string
}

// GetStatusOfRecentDeposits optional parameters
type GetStatusOfRecentDepositsOptions struct {
	// Name of the deposit method
	Method string
}

// Get Status of Recent Deposits response
type GetStatusOfRecentDepositsResponse struct {
	common.KrakenSpotRESTResponse
	// Recent deposits
	Result []TransactionDetails `json:"result"`
}
