package account

import "encoding/json"

// TradeInfo contains full trade information
type TradeInfo struct {
	// Order responsible for execution of trade
	OrderTransactionId string `json:"ordertxid"`
	// Position responsible for execution of trade
	PositionId string `json:"postxid"`
	// Asset pair
	Pair string `json:"pair"`
	// Unix timestamp for the trade
	Timestamp json.Number `json:"time"`
	// Trade direction (buy/sell)
	Type string `json:"type"`
	// Order type. Cf. OrderTypeEnum for values
	OrderType string `json:"ordertype"`
	// Average price order was executed at
	Price json.Number `json:"price"`
	// Total cost of order
	Cost json.Number `json:"cost,omitempty"`
	// Total fee
	Fee json.Number `json:"fee"`
	// Volume
	Volume json.Number `json:"vol"`
	// Initial margin
	Margin json.Number `json:"margin,omitempty"`
	// Amount of leverage used in trade.
	Leverage string `json:"leverage,omitempty"`
	// Comma delimited list of miscellaneous info:
	// closing — Trade closes all or part of a position
	Miscellaneous string `json:"misc,omitempty"`
	// Position status (open/closed)
	// - Only present if trade opened a position
	PositionStatus string `json:"posstatus,omitempty"`
	// Average price of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedPrice json.Number `json:"cprice,omitempty"`
	// Total cost of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedCost json.Number `json:"ccost,omitempty"`
	// Total fee of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedFee json.Number `json:"cfee,omitempty"`
	// Total fee of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedVolume json.Number `json:"cvol,omitempty"`
	// Total margin freed in closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedMargin json.Number `json:"cmargin,omitempty"`
	// Net profit/loss of closed portion of position (quote currency, quote currency scale)
	// - Only present if trade opened a position
	ClosedNetPNL json.Number `json:"net,omitempty"`
	// List of closing trades for position (if available)
	// - Only present if trade opened a position
	ClosingTrades []string `json:"trades,omitempty"`
}
