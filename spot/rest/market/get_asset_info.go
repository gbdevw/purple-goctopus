package market

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for asset statuses
type AssetStatus string

// Values for AssetStatus
const (
	Enabled                    AssetStatus = "enabled"
	DepositOnly                AssetStatus = "deposit_only"
	WithdrawalOnly             AssetStatus = "withdrawal_only"
	FundingTemporarilyDisabled AssetStatus = "funding_temporarily_disabled"
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
	// Asset status
	Status AssetStatus `json:"status"`
}

// GetAssetInfo options
type GetAssetInfoOptions struct {
	// List of assets to get info on.
	// Defaults to all assets.
	// A nil value triggers default behavior.
	Assets []string
	// Asset class.
	// Defaults to 'currency'.
	// An empty string triggers default behavior.
	AssetClass string
}

// GetAssetInfo response
type GetAssetInfoResponse struct {
	common.KrakenSpotRESTResponse
	Result map[string]AssetInfo `json:"result,omitempty"`
}
