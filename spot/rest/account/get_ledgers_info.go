package account

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for LedgerInfoType
type LedgerInfoType string

// Values for LedgerType
const (
	LedgerAll        LedgerInfoType = "all"
	LedgerDeposit    LedgerInfoType = "deposit"
	LedgerWithdrawal LedgerInfoType = "withdrawal"
	LedgerTrade      LedgerInfoType = "trade"
	LedgerMargin     LedgerInfoType = "margin"
	LedgerRollover   LedgerInfoType = "rollover"
	LedgerCredit     LedgerInfoType = "credit"
	LedgerTransfer   LedgerInfoType = "transfer"
	LedgerSettled    LedgerInfoType = "settled"
	LedgerStaking    LedgerInfoType = "staking"
	LedgerSale       LedgerInfoType = "sale"
)

// GetLedgersInfoOptions contains Get Ledgers Info optional parameters.
type GetLedgersInfoOptions struct {
	// List of assets to restrict output to.
	// By default, all assets are accepted.
	Assets []string
	// Asset class to restrict output to.
	// Defaults to "currency".
	AssetClass string
	// Type of ledger to retrieve.
	// Defaults to "all".
	// Values: "all" "deposit" "withdrawal" "trade" "margin" "rollover" "credit" "transfer" "settled" "staking" "sale"
	Type string
	// Starting unix timestamp or order tx ID of results (exclusive).
	Start *time.Time
	// Ending unix timestamp or order tx ID of results (inclusive).
	End *time.Time
	// Result offset for pagination.
	Offset *int64
}

// Ledgers info
type LedgersInfo struct {
	// Map where each key is a ledger entry ID and value a ledger entry
	Ledgers map[string]LedgerEntry `json:"ledger"`
	// Amount of available ledger info matching criteria
	Count int `json:"count"`
}

// GetLedgersInfoResponse contains GetLedgersInfo response data.
type GetLedgersInfoResponse struct {
	common.KrakenSpotRESTResponse
	Result *LedgersInfo `json:"result,omitempty"`
}
