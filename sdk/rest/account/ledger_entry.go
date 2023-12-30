package account

import "encoding/json"

// Enum for ledger entry types
type LedgerEntryTypeEnum string

// Values for LedgerEntryTypeEnum
const (
	EntryTypeNone            LedgerEntryTypeEnum = "none"
	EntryTypeTrade           LedgerEntryTypeEnum = "trade"
	EntryTypeDeposit         LedgerEntryTypeEnum = "deposit"
	EntryTypeWithdrawal      LedgerEntryTypeEnum = "withdrawal"
	EntryTypeTransfer        LedgerEntryTypeEnum = "transfer"
	EntryTypeMargin          LedgerEntryTypeEnum = "margin"
	EntryTypeAdjustment      LedgerEntryTypeEnum = "adjustment"
	EntryTypeRollover        LedgerEntryTypeEnum = "rollover"
	EntryTypeSpend           LedgerEntryTypeEnum = "spend"
	EntryTypeReceive         LedgerEntryTypeEnum = "receive"
	EntryTypeSettled         LedgerEntryTypeEnum = "settled"
	EntryTypeCredit          LedgerEntryTypeEnum = "credit"
	EntryTypeStaking         LedgerEntryTypeEnum = "staking"
	EntryTypeReward          LedgerEntryTypeEnum = "reward"
	EntryTypeDividend        LedgerEntryTypeEnum = "dividend"
	EntryTypeSale            LedgerEntryTypeEnum = "sale"
	EntryTypeConverion       LedgerEntryTypeEnum = "conversion"
	EntryTypeNftTrade        LedgerEntryTypeEnum = "nfttrade"
	EntryTypeNftCreatorFee   LedgerEntryTypeEnum = "nftcreatorfee"
	EntryTypeNftRebate       LedgerEntryTypeEnum = "nftrebate"
	EntryTypeCustodyTransfer LedgerEntryTypeEnum = "custodytransfer"
)

// LedgerEntry contains ledger entry data.
type LedgerEntry struct {
	// Reference Id
	ReferenceId string `json:"refid"`
	// Unix timestamp of ledger
	Timestamp json.Number `json:"time"`
	// Type of ledger entry. Cf LedgerEntryTypeEnum for values.
	Type string `json:"type"`
	// Additional info relating to the ledger entry type, where applicable
	SubType string `json:"subtype,omitempty"`
	// Asset class
	AssetClass string `json:"aclass"`
	// Asset
	Asset string `json:"asset"`
	// Transaction amount
	Amount json.Number `json:"amount"`
	// Transaction fee
	Fee json.Number `json:"fee"`
	// Resulting balance
	Balance json.Number `json:"balance"`
}
