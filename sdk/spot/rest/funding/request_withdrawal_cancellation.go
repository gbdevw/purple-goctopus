package funding

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// RequestWithdrawalCancellation request parameters
type RequestWithdrawalCancellationRequestParameters struct {
	// Asset being withdrawn
	Asset string `json:"asset"`
	// Withdrawal reference ID
	ReferenceId string `json:"refid"`
}

// RequestWithdrawalCancellation response
type RequestWithdrawalCancellationResponse struct {
	common.KrakenSpotRESTResponse
	// Result for RequestWithdrawalCancellation
	Result bool `json:"result"`
}
