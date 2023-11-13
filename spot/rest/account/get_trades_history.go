package account

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for trade types
type TradeType string

// Values for TradeType
const (
	TradeTypeAll             TradeType = "all"
	TradeTypeAnyPosition     TradeType = "any position"
	TradeTypeClosedPosition  TradeType = "closed position"
	TradeTypeClosingPosition TradeType = "closing position"
	TradeTypeNoPosition      TradeType = "no position"
)

// GetTradesHistoryOptions contains Get Trade History optional parameters.
type GetTradesHistoryOptions struct {
	// Type of trade.
	// Defaults to "all".
	// Values: "all" "any position" "closed position" "closing position" "no position"
	Type string
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Starting unix timestamp or order tx ID of results (exclusive).
	Start *time.Time
	// Ending unix timestamp or order tx ID of results (inclusive).
	End *time.Time
	// Result offset for pagination.
	Offset *int64
}

// Trades history
type TradesHistory struct {
	// Map where each key is a transaction ID and value a trade info object
	Trades map[string]TradeInfo `json:"trades"`
	// Amount of available trades matching criteria
	Count int
}

// GetTradesHistoryResponse contains GetTradesHistory response data.
type GetTradesHistoryResponse struct {
	common.KrakenSpotRESTResponse
	Result *TradesHistory `json:"result,omitempty"`
}
