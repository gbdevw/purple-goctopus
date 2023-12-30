package funding

import (
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// GetWithdrawalAddresses request options
type GetWithdrawalAddressesRequestOptions struct {
	// Filter addresses for specific asset.
	//
	// An empty string means no filtering.
	Asset string `json:"asset,omitempty"`
	// Filter Addresses for specific method.
	//
	// An empty string means no filtering.
	Method string `json:"method,omitempty"`
	// Find address for by withdrawal key name, as set up on your account.
	//
	// An empty string means no filtering.
	Key string `json:"key,omitempty"`
	// Filter by verification status of the withdrawal address. Withdrawal addresses successfully
	// completing email confirmation will have a verification status of true.
	//
	// Defaults to false.
	Verified bool `json:"verified,omitempty"`
}

// Data of a withdrawal address.
type WithdrawalAddress struct {
	// Withdrawal address.
	Address string `json:"address"`
	// Name of asset being withdrawn.
	Asset string `json:"asset"`
	// Name of the withdrawal method.
	Method string `json:"method"`
	// Withdrawal key name, as set up on your account.
	Key string `json:"key"`
	// Memo for withdrawal address, as set up on your account, if applicable.
	Memo string `json:"memo,omitempty"`
	// Verification status of withdrawal address.
	Verified bool `json:"verified"`
}

// GetWithdrawalAddresses response
type GetWithdrawalAddressesResponse struct {
	common.KrakenSpotRESTResponse
	Result []WithdrawalAddress `json:"result,omitempty"`
}
