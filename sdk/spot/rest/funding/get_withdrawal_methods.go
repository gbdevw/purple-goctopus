package funding

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
)

// GetWithdrawalMethods request options
type GetWithdrawalMethodsRequestOptions struct {
	// Filter methods for specific asset.
	//
	// An empty string means no filtering.
	Asset string `json:"asset,omitempty"`
	// Filter methods for specific network.
	//
	// An empty string means no filtering.
	Network string `json:"network,omitempty"`
}

// Data of a withdrawal method.
type WithdrawalMethod struct {
	// Name of asset being withdrawn
	Asset string `json:"asset"`
	// Name of the withdrawal method
	Method string `json:"method"`
	// Name of the blockchain or network being withdrawn on
	Network string `json:"network"`
	// Minimum net amount that can be withdrawn right now
	Minimum json.Number `json:"minimum"`
}

// GetWithdrawalMethods response
type GetWithdrawalMethodsResponse struct {
	common.KrakenSpotRESTResponse
	Result []WithdrawalMethod `json:"result,omitempty"`
}
