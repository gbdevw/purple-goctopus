package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for position statuses
type PositionStatus string

// Values for PositionStatus
const (
	PositionOpen   PositionStatus = "open"
	PositionClosed PositionStatus = "closed"
)

// PositionInfo contains position information
type PositionInfo struct {
	// Order ID responsible for the position
	OrderTransactionId string `json:"ordertxid"`
	// Position status
	PositionStatus string `json:"posstatus"`
	// Asset pair
	Pair string `json:"pair"`
	// Unix timestamp of trade
	Timestamp float64 `json:"time"`
	// Direction (buy/sell) of position
	Type string `json:"type"`
	// Order type used to open position
	OrderType string `json:"ordertype"`
	// Opening cost of position (in quote currency)
	Cost string `json:"cost"`
	// Opening fee of position (in quote currency)
	Fee string `json:"fee"`
	// Position opening size (in base currency)
	Volume string `json:"vol"`
	// Quantity closed (in base currency)
	ClosedVolume string `json:"vol_closed"`
	// Initial margin consumed (in quote currency)
	Margin string `json:"margin"`
	// Current value of remaining position (if docalcs requested)
	Value string `json:"value,omitempty"`
	// Unrealised P&L of remaining position (if docalcs requested)
	Net string `json:"net,omitempty"`
	// Funding cost and term of position
	Terms string `json:"terms"`
	// Timestamp of next margin rollover fee
	RolloverTimestamp string `json:"rollovertm"`
	// Comma delimited list of add'l info
	Miscellaneous string `json:"misc"`
	// Comma delimited list of opening order flags
	OrderFlags string `json:"oflags"`
}

// GetOpenPositionsOptions contains Get Open Positions optional parameters
type GetOpenPositionsOptions struct {
	// List of txids to limit output to
	TransactionIds []string
	// Whether to include P&L calculations.
	// Defaults to false.
	DoCalcs bool
	// Consolidate positions by market/pair.
	// Value: "market"
	// Consolidation is disabled because using market
	// changes radically the response payload and cause client to fail
	// Consolidation string
}

// GetOpenPositionsResponse contains Get Open Positions response data.
type GetOpenPositionsResponse struct {
	common.KrakenSpotRESTResponse
	// Map where each key is a transaction ID and values an open position description.
	Result map[string]PositionInfo `json:"result,omitempty"`
}
