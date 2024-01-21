package account

import (
	"github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
)

// Enum for LedgerInfoType
type LedgerInfoTypeEnum string

// Values for LedgerInfoTypeEnum
const (
	LedgerAll        LedgerInfoTypeEnum = "all"
	LedgerTrade      LedgerInfoTypeEnum = "trade"
	LedgerDeposit    LedgerInfoTypeEnum = "deposit"
	LedgerWithdrawal LedgerInfoTypeEnum = "withdrawal"
	LedgerTransfer   LedgerInfoTypeEnum = "transfer"
	LedgerMargin     LedgerInfoTypeEnum = "margin"
	LedgerAdjustment LedgerInfoTypeEnum = "adjustment"
	LedgerRollover   LedgerInfoTypeEnum = "rollover"
	LedgerCredit     LedgerInfoTypeEnum = "credit"
	LedgerSettled    LedgerInfoTypeEnum = "settled"
	LedgerStaking    LedgerInfoTypeEnum = "staking"
	LedgerDividend   LedgerInfoTypeEnum = "dividend"
	LedgerSale       LedgerInfoTypeEnum = "sale"
	LedgerNftRebate  LedgerInfoTypeEnum = "nft_rebate"
)

// GetLedgersInfo request options.
type GetLedgersInfoRequestOptions struct {
	// List of assets to restrict output to.
	//
	// An empty array means no filtering.
	Assets []string `json:"assets,omitempty"`
	// Asset class to restrict output to.
	//
	// Defaults to "currency". An empty string triggers the default behavior.
	AssetClass string `json:"aclass,omitempty"`
	// Type of ledger to retrieve.
	//
	// Defaults to "all". An empty string triggers the default behavior. Cf. LedgerInfoTypeEnum for values.
	Type string `json:"type,omitempty"`
	// Starting unix timestamp or ledger ID of results (exclusive).
	//
	// An empty string means no fitlering.
	Start string `json:"start,omitempty"`
	// Ending unix timestamp or ledger ID of results (inclusive)
	//
	// An empty string means no fitlering.
	End string `json:"end,omitempty"`
	// Result offset for pagination.
	//
	// A zero value means first items will be fetched.
	Offset int64 `json:"ofs,omitempty"`
	// If true, does not retrieve count of ledger entries. Request can be noticeably faster for users
	// with many ledger entries as this avoids an extra database query.
	//
	// Defaults to false.
	WithoutCount bool `json:"without_count,omitempty"`
}

// GetLedgersInfo result.
type LedgersInfoResult struct {
	// Map where each key is a ledger entry ID and value a ledger entry
	Ledgers map[string]*LedgerEntry `json:"ledger,omitempty"`
	// Amount of available ledger info matching criteria
	Count int `json:"count"`
}

// GetLedgersInfo response.
type GetLedgersInfoResponse struct {
	common.KrakenSpotRESTResponse
	Result *LedgersInfoResult `json:"result,omitempty"`
}
