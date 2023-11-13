package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

type WithdrawFundsResult struct {
	// Reference ID
	ReferenceID string `json:"refid"`
}

// WithdrawFunds required parameters
type WithdrawFundsParameters struct {
	// Asset being withdrawn
	Asset string
	// Withdrawal address name as setup on account
	Key string
	// Anount to be withdrawn
	Amount string
}

// WithdrawFunds response
type WithdrawFundsResponse struct {
	common.KrakenSpotRESTResponse
	// Results for WithdrawFunds
	Result *WithdrawFundsResult `json:"result"`
}
