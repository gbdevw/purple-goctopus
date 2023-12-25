package account

import (
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for trade types
type TradeTypeEnum string

// Values for TradeTypeEnum
const (
	TradeTypeAll             TradeTypeEnum = "all"
	TradeTypeAnyPosition     TradeTypeEnum = "any position"
	TradeTypeClosedPosition  TradeTypeEnum = "closed position"
	TradeTypeClosingPosition TradeTypeEnum = "closing position"
	TradeTypeNoPosition      TradeTypeEnum = "no position"
)

// GetTradesHistory request options.
type GetTradesHistoryRequestOptions struct {
	// Type of trade. Cf TradeTypeEnum for values.
	//
	// Defaults to "all". An empty string triggers the default behavior.
	Type string `json:"type,omitempty"`
	// Whether or not to include trades related to position in output.
	//
	// Defaults to false.
	Trades bool `json:"trades"`
	// Starting unix timestamp or trade tx ID of results (exclusive)
	//
	// An empty string means no filtering.
	Start string `json:"start,omitempty"`
	// Ending unix timestamp or order tx ID of results (inclusive).
	//
	// An empty string means no filtering.
	End string `json:"end,omitempty"`
	// Result offset for pagination.
	//
	// A zero values means first items will be fetched.
	Offset int64 `json:"ofs,omitempty"`
	// Whether or not to consolidate trades by individual taker trades.
	//
	// Defaults to false.
	ConsolidateTaker bool `json:"consolidate_taker"`
}

// GetTradesHistory results.
type GetTradesHistoryResult struct {
	// Map where each key is a transaction ID and value a trade info object.
	Trades map[string]*TradeInfo `json:"trades,omitempty"`
	// Amount of available trades matching criteria.
	Count int `json:"count"`
}

// GetTradesHistory response.
type GetTradesHistoryResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetTradesHistoryResult `json:"result,omitempty"`
}
