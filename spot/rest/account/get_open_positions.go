package account

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for position statuses
type PositionStatusEnum string

// Values for PositionStatusEnum
const (
	PositionOpen   PositionStatusEnum = "open"
	PositionClosed PositionStatusEnum = "closed"
)

// PositionInfo contains position information
type PositionInfo struct {
	// Order ID responsible for the position
	OrderTransactionId string `json:"ordertxid"`
	// Position status.
	//
	// Cf. PositionStatusEnum for values.
	PositionStatus string `json:"posstatus"`
	// Asset pair
	Pair string `json:"pair"`
	// Unix timestamp of trade
	Timestamp json.Number `json:"time"`
	// Direction (buy/sell) of position
	Type string `json:"type"`
	// Order type used to open position
	OrderType string `json:"ordertype"`
	// Opening cost of position (in quote currency)
	Cost json.Number `json:"cost"`
	// Opening fee of position (in quote currency)
	Fee json.Number `json:"fee"`
	// Position opening size (in base currency)
	Volume json.Number `json:"vol"`
	// Quantity closed (in base currency)
	ClosedVolume json.Number `json:"vol_closed"`
	// Initial margin consumed (in quote currency)
	Margin json.Number `json:"margin"`
	// Current value of remaining position (if docalcs requested)
	Value json.Number `json:"value,omitempty"`
	// Unrealised P&L of remaining position (if docalcs requested).
	//
	// A string is used because examples show values can be prefixed with a '+' that cause json.Number
	// to throw an error (so a string is used as instead).
	Net string `json:"net,omitempty"`
	// Funding cost and term of position
	Terms string `json:"terms"`
	// Timestamp of next margin rollover fee
	RolloverTimestamp json.Number `json:"rollovertm"`
	// Comma delimited list of add'l info
	Miscellaneous string `json:"misc"`
	// Comma delimited list of opening order flags
	OrderFlags string `json:"oflags"`
}

// GetOpenPositions request options.
type GetOpenPositionsRequestOptions struct {
	// List of txids to limit output to.
	//
	// An empty or nil value means no filtering.
	TransactionIds []string
	// Whether to include P&L calculations.
	//
	// Defaults to false.
	DoCalcs bool
	// Consolidate positions by market/pair.
	// Value: "market"
	// Consolidation is disabled because using market
	// changes radically the response payload and cause client to fail
	// Consolidation string
}

// GetOpenPositions response.
type GetOpenPositionsResponse struct {
	common.KrakenSpotRESTResponse
	// Map where each key is a transaction ID and values an open position description.
	Result map[string]*PositionInfo `json:"result,omitempty"`
}
