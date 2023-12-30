package funding

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// GetWithdrawalInformation request parameters
type GetWithdrawalInformationRequestParameters struct {
	// Asset being withdrawn
	Asset string `json:"asset"`
	// Withdrawal address name as setup on account
	Key string `json:"key"`
	// Anount to be withdrawn
	Amount string `json:"amount"`
}

// GetWithdrawalInformation result
type GetWithdrawalInformationResult struct {
	// Name of the withdrawal method that will be used
	Method string `json:"method"`
	// Maximum net amount that can be withdrawn right now
	Limit string `json:"limit"`
	// Net amount that will be sent, after fees
	Amount string `json:"amount"`
	// Amount of fees that will be paid
	Fee string `json:"fee"`
}

// Get Withdrawal Information response
type GetWithdrawalInformationResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetWithdrawalInformationResult `json:"result,omitempty"`
}
