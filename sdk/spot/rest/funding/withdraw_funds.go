package funding

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// WithdrawFunds request parameters
type WithdrawFundsRequestParameters struct {
	// Asset being withdrawn
	Asset string `json:"asset"`
	// Withdrawal address name as setup on account
	Key string `json:"key"`
	// Amount to be withdrawn
	Amount string `json:"amount"`
}

// WithdrawFunds request options
type WithdrawFundsRequestOptions struct {
	// Optional, crypto address that can be used to confirm address matches key (will return
	// Invalid withdrawal address error if different).
	//
	// An empty value means no address confirmation is required.
	Address string `json:"address,omitempty"`
	// Optional, if the processed withdrawal fee is higher than max_fee, withdrawal will fail with
	// EFunding:Max fee exceeded
	//
	// An empty value means no maximum fee set.
	MaxFee string `json:"max_fee,omitempty"`
}

// WithdrawFunds response
type WithdrawFundsResult struct {
	// Reference ID
	ReferenceID string `json:"refid"`
}

// WithdrawFunds response
type WithdrawFundsResponse struct {
	common.KrakenSpotRESTResponse
	// Results for WithdrawFunds
	Result *WithdrawFundsResult `json:"result"`
}
