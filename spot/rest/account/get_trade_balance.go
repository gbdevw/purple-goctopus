package account

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// GetTradeBalance rquest options.
type GetTradeBalanceRequestOptions struct {
	// Base asset used to determine balance.
	//
	// Defaults to ZUSD. An empty value triggers the default behavior.
	Asset string
}

// Trade balance data.
type GetTradeBalanceResult struct {
	// Equivalent balance (combined balance of all currencies)
	EquivalentBalance json.Number `json:"eb,omitempty"`
	// Trade balance (combined balance of all equity currencies)
	TradeBalance json.Number `json:"tb,omitempty"`
	// Margin amount of open positions
	MarginAmount json.Number `json:"m,omitempty"`
	// Unrealized net profit/loss of open positions
	UnrealizedNetPNL json.Number `json:"n,omitempty"`
	// Cost basis of open positions
	CostBasis json.Number `json:"c,omitempty"`
	// Current floating valuation of open positions
	FloatingValuation json.Number `json:"v,omitempty"`
	// Equity: trade balance + unrealized net profit/loss
	Equity json.Number `json:"e,omitempty"`
	// Free margin: Equity - initial margin (maximum margin available to open new positions)
	FreeMargin json.Number `json:"mf,omitempty"`
	// Margin level: (equity / initial margin) * 100
	MarginLevel json.Number `json:"ml,omitempty"`
	// Value of unfilled and partially filled orders
	UnexecutedValue json.Number `json:"uv,omitempty"`
}

// GetTradeBalance response.
type GetTradeBalanceResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetTradeBalanceResult `json:"result,omitempty"`
}
