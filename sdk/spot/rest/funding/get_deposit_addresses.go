package funding

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
)

// Data of a deposit address
type DepositAddress struct {
	// Deposit Address
	Address string `json:"address,omitempty"`
	// Expiration time as a unix timestamp (seconds). 0 if not expiring.
	Expiretm json.Number `json:"expiretm,omitempty"`
	// Whether or not address has ever been used
	New bool `json:"new"`
	// Only returned for STX, XLM, and EOS deposit addresses
	Memo string `json:"memo,omitempty"`
	// Only returned for XRP deposit addresses
	Tag string `json:"tag,omitempty"`
}

// GetDepositAddresses request parameters
type GetDepositAddressesRequestParameters struct {
	// Asset being deposited
	Asset string `json:"asset"`
	// Name of the deposit method
	Method string `json:"method"`
}

// GetDepositAddresses request options
type GetDepositAddressesRequestOptions struct {
	// Whether or not to generate a new address.
	//
	// Defaults to false.
	New bool `json:"new"`
	// Amount user wish to deposit on the address.
	//
	// This options is only required for "Bitcoin Lightning" deposit method in order to
	// craft the Lightning network invoice.
	//
	// An empty string means option will not be used.
	Amount string `json:"amount,omitempty"`
}

// Get Deposit Addresses response
type GetDepositAddressesResponse struct {
	common.KrakenSpotRESTResponse
	Result []DepositAddress `json:"result"`
}
