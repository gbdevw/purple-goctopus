package account

// Enum for ledger entry types
type LedgerEntryType string

// Values for LedgerEntryType
const (
	EntryTypeTrade      LedgerEntryType = "trade"
	EntryTypeDeposit    LedgerEntryType = "deposit"
	EntryTypeWithdrawal LedgerEntryType = "withdrawal"
	EntryTypeTransfer   LedgerEntryType = "transfer"
	EntryTypeMargin     LedgerEntryType = "margin"
	EntryTypeRollover   LedgerEntryType = "rollover"
	EntryTypeSpend      LedgerEntryType = "spend"
	EntryTypeReceive    LedgerEntryType = "receive"
	EntryTypeSettled    LedgerEntryType = "settled"
	EntryTypeAdjustment LedgerEntryType = "adjustment"
)

// LedgerEntry contains ledger entry data.
type LedgerEntry struct {
	// Reference Id
	ReferenceId string `json:"refid"`
	// Unix timestamp of ledger
	Timestamp float64 `json:"time"`
	// Type of ledger entry
	Type string `json:"type"`
	// Additional info relating to the ledger entry type, where applicable
	SubType string `json:"subtype,omitempty"`
	// Asset class
	AssetClass string `json:"aclass"`
	// Asset
	Asset string `json:"asset"`
	// Transaction amount
	Amount string `json:"amount"`
	// Transaction fee
	Fee string `json:"fee"`
	// Resulting balance
	Balance string `json:"balance"`
}
