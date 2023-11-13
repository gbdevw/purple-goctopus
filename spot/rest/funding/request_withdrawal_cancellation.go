package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// RequestWithdrawalCancellation required parameters
type RequestWithdrawalCancellationParameters struct {
	// Asset being withdrawn
	Asset string
	// Withdrawal reference ID
	ReferenceId string
}

// RequestWithdrawalCancellation response
type RequestWithdrawalCancellationResponse struct {
	common.KrakenSpotRESTResponse
	// Result for RequestWithdrawalCancellation
	Result bool `json:"result"`
}
