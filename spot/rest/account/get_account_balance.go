package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetAccountBalanceResponse contains Get Account Balance response data.
type GetAccountBalanceResponse struct {
	common.KrakenSpotRESTResponse
	// Balance for each possessed asset
	Result map[string]string `json:"result,omitempty"`
}
