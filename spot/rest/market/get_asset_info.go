package market

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for asset statuses
type AssetStatusEnum string

// Values for AssetStatus
const (
	Enabled                    AssetStatusEnum = "enabled"
	DepositOnly                AssetStatusEnum = "deposit_only"
	WithdrawalOnly             AssetStatusEnum = "withdrawal_only"
	FundingTemporarilyDisabled AssetStatusEnum = "funding_temporarily_disabled"
)

// AssetInfo represents an asset information
type AssetInfo struct {
	// Asset class
	AssetClass string `json:"aclass"`
	// Alternative name
	Altname string `json:"altname"`
	// Scaling decimal places for record keeping
	Decimals int `json:"decimals"`
	// Scaling decimal places for output display
	DisplayDecimals int `json:"display_decimals"`
	// Collateral value
	CollateralValue float64 `json:"collateral_value"`
	// Asset status.
	//
	// Cf AssetStatusEnum for values.
	Status string `json:"status"`
}

// GetAssetInfo request options
type GetAssetInfoRequestOptions struct {
	// List of assets to get info on.
	//
	// Defaults to all assets. An empty or nil value triggers default behavior.
	Assets []string `json:"assets,omitempty"`
	// Asset class.
	//
	// Defaults to 'currency'. An empty string triggers default behavior.
	AssetClass string `json:"asset_class,omitempty"`
}

// GetAssetInfo response
type GetAssetInfoResponse struct {
	common.KrakenSpotRESTResponse
	Result map[string]*AssetInfo `json:"result,omitempty"`
}
