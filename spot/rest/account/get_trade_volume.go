package account

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// FeeTierInfo contains fee tier information.
type FeeTierInfo struct {
	// Current fee in percent
	Fee json.Number `json:"fee"`
	// Minimum fee for pair if not fixed fee
	MinimumFee json.Number `json:"minfee"`
	// Maximum fee for pair if not fixed fee
	MaximumFee json.Number `json:"maxfee"`
	// Next tier's fee for pair if not fixed fee, empty if at lowest fee tier
	NextFee json.Number `json:"nextfee,omitempty"`
	// Volume level of current tier (if not fixed fee. empty if at lowest fee tier)
	TierVolume json.Number `json:"tiervolume,omitempty"`
	// Volume level of next tier (if not fixed fee. enpty if at lowest fee tier)
	NextTierVolume json.Number `json:"nextvolume,omitempty"`
}

// GetTradeVolume result
type GetTradeVolumeResult struct {
	// Volume currency
	Currency string `json:"currency"`
	// Current discount volume
	Volume json.Number `json:"volume"`
	// Fee info or Taker fee if asset is submitted to maker/taker fees - each key is an asset pair
	Fees map[string]*FeeTierInfo `json:"fees"`
	// Maker fee info - each key is an asset pair
	FeesMaker map[string]*FeeTierInfo `json:"fees_maker"`
}

// GetTradeVolume request options.
type GetTradeVolumeRequestOptions struct {
	// List of asset pairs to get fee info on.
	Pairs []string
}

// GetTradeVolume response.
type GetTradeVolumeResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetTradeVolumeResult `json:"result,omitempty"`
}
