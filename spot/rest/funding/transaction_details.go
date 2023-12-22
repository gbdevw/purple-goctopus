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
