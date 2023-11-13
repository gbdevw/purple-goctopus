package funding

// Enum for transaction states as described in https://github.com/globalcitizen/ifex-protocol/blob/master/draft-ifex-00.txt#L837
type TransactionState string

// Values for TransactionState
const (
	TxStateInitial TransactionState = "Initial"
	TxStatePending TransactionState = "Pending"
	TxStateSettled TransactionState = "Settled"
	TxStateSuccess TransactionState = "Success"
	TxStateFailure TransactionState = "Failure"
	TxStatePartial TransactionState = "Partial"
)

// Enum for additional properties for transaction status
type TransactionStatus string

// Values for TransactionStatus
const (
	// A return transaction initiated by Kraken
	TxStatusReturn TransactionStatus = "return"
	// Deposit is on hold pending review
	TxStatusOnHold TransactionStatus = "onhold"
	// Cancelation requested
	TxCancelPending TransactionStatus = "cancel-pending"
	// Canceled
	TxCanceled TransactionStatus = "canceled"
	// CancelDenied
	TxCancelDenied TransactionStatus = "cancel-denied"
)

// Enum for staking transaction types
type StakingTxType string

// Values for types of staking transactions
const (
	StakingBonding   StakingTxType = "bonding"
	StakingReward    StakingTxType = "reward"
	StakingUnbonding StakingTxType = "unbonding"
)

// Transaction details for a deposit or a withdrawal
type TransactionDetails struct {
	// Name of deposit method
	Method string `json:"method"`
	// Asset class
	AssetClass string `json:"aclass"`
	// Asset
	Asset string `json:"asset"`
	// Reference ID
	ReferenceID string `json:"refid"`
	// Method transaction ID
	TransactionID string `json:"txid"`
	// Method transaction information
	Info string `json:"info"`
	// Amount deposited/withdrawn
	Amount string `json:"amount"`
	// Fees paid. Can be empty
	Fee string `json:"fee"`
	// Unix timestamp when request was made
	Time int64 `json:"time"`
	// Status of deposit - IFEX financial transaction states
	Status string `json:"status"`
	// Additional status property. Can be empty
	StatusProperty string `json:"status-prop,omitempty"`
}