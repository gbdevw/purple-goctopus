package funding

import (
	"encoding/json"
	"fmt"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Data of a deposit method
type DepositMethod struct {
	// Name of deposit method
	Method string `json:"method,omitempty"`
	// Maximum net amount that can be deposited right now. Empty or "false" if no limit.
	Limit string `json:"limit,omitempty"`
	// Amount of fees that will be paid.
	Fee string `json:"fee,omitempty"`
	// Whether or not method has an address setup fee.
	AddressSetupFee string `json:"address-setup-fee,omitempty"`
	// Whether new addresses can be generated for this method.
	GenAddress bool `json:"gen-address"`
	// Minimum net amount that can be deposited right now
	Minimum string `json:"minimum,omitempty"`
}

// Data of a deposit method
type depositMethod struct {
	// Name of deposit method
	Method string `json:"method,omitempty"`
	// Maximum net amount that can be deposited right now. Empty or "false" if no limit.
	Limit interface{} `json:"limit,omitempty"`
	// Amount of fees that will be paid.
	Fee string `json:"fee,omitempty"`
	// Whether or not method has an address setup fee.
	AddressSetupFee string `json:"address-setup-fee,omitempty"`
	// Whether new addresses can be generated for this method.
	GenAddress bool `json:"gen-address"`
	// Minimum net amount that can be deposited right now
	Minimum string `json:"minimum,omitempty"`
}

func (dm *DepositMethod) UnmarshalJSON(data []byte) error {
	// Unmarshal in private struct with interface{} for limit
	tmp := new(depositMethod)
	err := json.Unmarshal(data, tmp)
	if err != nil {
		return err
	}
	// Copy fields into target expect for Limit
	dm.AddressSetupFee = tmp.AddressSetupFee
	dm.Fee = tmp.Fee
	dm.GenAddress = tmp.GenAddress
	dm.Method = tmp.Method
	dm.Minimum = tmp.Minimum
	// Leave empty if tmp.Limit is nil
	if tmp.Limit != nil {
		// Set Limit as text representation of what has been unmarshalled
		dm.Limit = fmt.Sprint(tmp.Limit)
	}
	// Exit
	return nil
}

// GetDepositMethods request parameters
type GetDepositMethodsRequestParameters struct {
	// Asset being deposited
	Asset string `json:"asset"`
}

// Response returned by GetDepositMethods
type GetDepositMethodsResponse struct {
	common.KrakenSpotRESTResponse
	Result []*DepositMethod `json:"result"`
}
