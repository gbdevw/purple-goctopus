package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Data of a deposit address
type DepositAddress struct {
	// Deposit Address
	Address string `json:"address"`
	// Expiration time as a unix timestamp (seconds). 0 if not expiring.
	Expiretm int64 `json:"expiretm,string"`
	// Whether or not address has ever been used
	New bool `json:"new"`
	// Only returned for STX, XLM, and EOS deposit addresses
	Memo string `json:"memo"`
	// Only returned for XRP deposit addresses
	Tag string `json:"tag"`
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
}

// Get Deposit Addresses response
type GetDepositAddressesResponse struct {
	common.KrakenSpotRESTResponse
	Result []DepositAddress `json:"result"`
}
