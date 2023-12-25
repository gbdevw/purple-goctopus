package account

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Data of an extended balance for one asset.
type ExtendedBalance struct {
	// Total balance amount for an asset
	Balance json.Number `json:"balance"`
	// Total credit amount (only applicable if account has a credit line)
	Credit json.Number `json:"credit,omitempty"`
	// Used credit amount (only applicable if account has a credit line)
	CreditUsed json.Number `json:"credit_used.omitempty"`
	// Total held amount for an asset
	HoldTrade json.Number `json:"hold_trade"`
}

// GetExtendedBalance response
type GetExtendedBalanceResponse struct {
	common.KrakenSpotRESTResponse
	Result map[string]*ExtendedBalance `json:"result,omitempty"`
}
