package account

// TradeInfo contains full trade information
type TradeInfo struct {
	// Order responsible for execution of trade
	OrderTransactionId string `json:"ordertxid"`
	// Asset pair
	Pair string `json:"pair"`
	// Unix timestamp for the trade
	Timestamp float64 `json:"time"`
	// Trade direction (buy/sell)
	Type string `json:"type"`
	// Order type. Enum: "market" "limit" "stop-loss" "take-profit" "stop-loss-limit" "take-profit-limit" "settle-position"
	OrderType string `json:"ordertype"`
	// Average price order was executed at
	Price string `json:"price"`
	// Total cost of order
	Cost string `json:"cost"`
	// Total fee
	Fee string `json:"fee"`
	// Volume
	Volume string `json:"vol"`
	// Initial margin
	Margin string `json:"margin"`
	// Comma delimited list of miscellaneous info:
	// closing — Trade closes all or part of a position
	Miscellaneous string `json:"misc"`
	// Position status (open/closed)
	// - Only present if trade opened a position
	PositionStatus string `json:"posstatus,omitempty"`
	// Average price of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedPrice string `json:"cprice,omitempty"`
	// Total cost of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedCost string `json:"ccost,omitempty"`
	// Total fee of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedFee string `json:"cfee,omitempty"`
	// Total fee of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedVolume string `json:"cvol,omitempty"`
	// Total margin freed in closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedMargin string `json:"cmargin,omitempty"`
	// Net profit/loss of closed portion of position (quote currency, quote currency scale)
	// - Only present if trade opened a position
	ClosedNetPNL string `json:"net,omitempty"`
	// List of closing trades for position (if available)
	// - Only present if trade opened a position
	ClosingTrades []string `json:"trades,omitempty"`
}
