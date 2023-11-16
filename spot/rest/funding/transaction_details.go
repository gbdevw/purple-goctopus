package funding

// Enum for transaction states as described in https://github.com/globalcitizen/ifex-protocol/blob/master/draft-ifex-00.txt#L837
type TransactionStateEnum string

// Values for TransactionStateEnum
const (
	TxStateInitial TransactionStateEnum = "Initial"
	TxStatePending TransactionStateEnum = "Pending"
	TxStateSettled TransactionStateEnum = "Settled"
	TxStateSuccess TransactionStateEnum = "Success"
	TxStateFailure TransactionStateEnum = "Failure"
	TxStatePartial TransactionStateEnum = "Partial"
)

// Enum for status property.
type StatusPropertyEnum string

// Values for StatusPropertyEnum
const (
	// A return transaction initiated by Kraken
	Return StatusPropertyEnum = "return"
	// Deposit is on hold pending review
	OnHold StatusPropertyEnum = "onhold"
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
	// Additional status property. Can be empty.
	StatusProperty string `json:"status-prop,omitempty"`
	// Client sending transaction id(s) for deposits that credit with a sweeping transaction
	Originators []string `json:"originators,omitempty"`
}
