package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetTradeBalanceOptions contains Get Trade Balance optional parameters.
type GetTradeBalanceOptions struct {
	// Base asset used to determine balance.
	// Defaults to ZUSD.
	Asset string
}

// Trade balance
type TradeBalance struct {
	// Equivalent balance (combined balance of all currencies)
	EquivalentBalance string `json:"eb"`
	// Trade balance (combined balance of all equity currencies)
	TradeBalance string `json:"tb"`
	// Margin amount of open positions
	MarginAmount string `json:"m"`
	// Unrealized net profit/loss of open positions
	UnrealizedNetPNL string `json:"n"`
	// Cost basis of open positions
	CostBasis string `json:"c"`
	// Current floating valuation of open positions
	FloatingValuation string `json:"v"`
	// Equity: trade balance + unrealized net profit/loss
	Equity string `json:"e"`
	// Free margin: Equity - initial margin (maximum margin available to open new positions)
	FreeMargin string `json:"mf"`
	// Margin level: (equity / initial margin) * 100
	MarginLevel string `json:"ml"`
}

// GetTradeBalanceResponse contains GetTradeBalance response data.
type GetTradeBalanceResponse struct {
	common.KrakenSpotRESTResponse
	Result *TradeBalance `json:"result,omitempty"`
}
