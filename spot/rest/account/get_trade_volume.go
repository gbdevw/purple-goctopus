package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// FeeTierInfo contains fee tier information.
type FeeTierInfo struct {
	// Current fee in percent
	Fee string `json:"fee"`
	// Minimum fee for pair if not fixed fee
	MinimumFee string `json:"min_fee"`
	// Maximum fee for pair if not fixed fee
	MaximumFee string `json:"max_fee"`
	// Next tier's fee for pair if not fixed fee, empty if at lowest fee tier
	NextFee string `json:"next_fee,omitempty"`
	// Volume level of current tier (if not fixed fee. empty if at lowest fee tier)
	TierVolume string `json:"tier_volume,omitempty"`
	// Volume level of next tier (if not fixed fee. enpty if at lowest fee tier)
	NextTierVolume string `json:"next_volume,omitempty"`
}

// GetTradeVolumeParameters contains Get Trade Volume required parameters.
type GetTradeVolumeParameters struct {
	// List of asset pairs to get fee info on
	Pairs []string
}

// GetTradeVolumeOptions contains Get Trade Volume optional parameters.
type GetTradeVolumeOptions struct {
	// Whether or not to include fee info in results
	FeeInfo bool
}

// Trade volume
type TradeVolume struct {
	// Volume currency
	Currency string `json:"currency"`
	// Current discount volume
	Volume string `json:"volume"`
	// Fee info or Taker fee if asset is submitted to maker/taker fees - each key is an asset pair
	Fees map[string]FeeTierInfo `json:"fees"`
	// Maker fee info - each key is an asset pair
	FeesMaker map[string]FeeTierInfo `json:"fees_maker"`
}

// GetTradeVolumeResponse contains Get Trade Volume response data.
type GetTradeVolumeResponse struct {
	common.KrakenSpotRESTResponse
	Result *TradeVolume `json:"result,omitempty"`
}
