package market

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for the status of the asset pair. Possible values: online, cancel_only, post_only, limit_only, reduce_only.
type PairStatus string

// Values for PairStatus
const (
	PairOnline     PairStatus = "online"
	PairCancelOnly PairStatus = "cancel_only"
	PairPostOnly   PairStatus = "post_only"
	PairLimitOnly  PairStatus = "limit_only"
	PairReduceOnly PairStatus = "reduce_only"
)

// Enum for pair info type
type PairInfo string

// Values for PairInfo
const (
	// Get all asset pair info.
	InfoAll PairInfo = "info"
	// Get leverage info
	InfoLeverage PairInfo = "leverage"
	// Get fees info
	InfoFees PairInfo = "fees"
	// Get margin info
	InfoMargin PairInfo = "margin"
)

// AssetPairInfo represents asset pair information
type AssetPairInfo struct {
	// Alternative pair name
	AlternativeName string `json:"altname"`
	// Name on Websocket API (if available)
	WebsocketName string `json:"wsname,omitempty"`
	// Asset class of base component
	AssetClassBase string `json:"aclass_base"`
	// Asset id of base currency/asset
	Base string `json:"base"`
	// Asset class of quote currency/asset
	AssetClassQuote string `json:"aclass_quote"`
	// Asset id of quote currency/asset
	Quote string `json:"quote"`
	// Scaling decimal places for cost
	CostDecimals int `json:"cost_decimals"`
	// Scaling decimal places for pair
	PairDecimals int `json:"pair_decimals"`
	// Scaling decimal places for volume
	LotDecimals int `json:"lot_decimals"`
	// Amount to multiply lot volume by to get currency volume
	LotMultiplier int `json:"lot_multiplier"`
	// Array of leverage amounts available when buying
	LeverageBuy []int `json:"leverage_buy"`
	// Array of leverage amounts available when selling
	LeverageSell []int `json:"leverage_sell"`
	// Fee schedule array in [volume, percent fee] tuples
	Fees [][]float64 `json:"fees"`
	// Maker fee schedule array in [volume, percent fee] tuples (if on maker/taker)
	FeesMaker [][]float64 `json:"fees_maker"`
	// // Volume discount currency
	FeeVolumeCurrency string `json:"fee_volume_currency"`
	// Margin call level
	MarginCall int `json:"margin_call"`
	// Stop-out/Liquidation margin level
	MarginStop int `json:"margin_stop"`
	// Order minimum
	OrderMin string `json:"ordermin"`
	// Minimum order cost (in terms of quote currency)
	CostMin string `json:"costmin"`
	// Minimum increment between valid price levels
	TickSize string `json:"tick_size"`
	// Status of asset. Possible values: online, cancel_only, post_only, limit_only, reduce_only.
	Status PairStatus `json:"status"`
	// Maximum long margin position size (in terms of base currency)
	LongPositionLimit int `json:"long_position_limit"`
	// Maximum short margin position size (in terms of base currency)
	ShortPositionLimit int `json:"short_position_limit"`
}

// GetTradableAssetPairs request options
type GetTradableAssetPairsRequestOptions struct {
	// Pairs to get info on.
	//
	// Defaults to all pairs. An empty string triggers default behavior.
	Pairs []string `json:"pairs,omitempty"`
	// Data to retrieve. Cf PairInfo for values.
	//
	// Defaults to InfoAll (info). An empty string triggers default behavior.
	Info string `json:"info,omitempty"`
}

// GetTradableAssetPairs response
type GetTradableAssetPairsResponse struct {
	common.KrakenSpotRESTResponse
	// Map each assert pair (ex: 1INCHEUR) to its info
	Result map[string]*AssetPairInfo `json:"result,omitempty"`
}
