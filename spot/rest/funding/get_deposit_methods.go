package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Data of a deposit method
type DepositMethod struct {
	// Name of deposit method
	Method string `json:"method"`
	// Maximum net amount that can be deposited right now. Empty or "false" if no limit.
	Limit string `json:"limit"`
	// Amount of fees that will be paid.
	Fee string `json:"fee"`
	// Whether or not method has an address setup fee.
	AddressSetupFee string `json:"address-setup-fee"`
	// Whether new addresses can be generated for this method.
	GenAddress bool `json:"gen-address"`
	// Minimum net amount that can be deposited right now
	Minimum string `json:"minimum"`
}

// GetDepositMethods required parameters
type GetDepositMethodsParameters struct {
	// Asset being deposited
	Asset string
}

// Response returned by GetDepositMethods
type GetDepositMethodsResponse struct {
	common.KrakenSpotRESTResponse
	Result []DepositMethod `json:"result"`
}
