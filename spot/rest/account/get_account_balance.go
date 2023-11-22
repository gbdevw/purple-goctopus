package account

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// GetAccountBalance response.
type GetAccountBalanceResponse struct {
	common.KrakenSpotRESTResponse
	// Balances for each possessed asset
	Result map[string]json.Number `json:"result,omitempty"`
}
