package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Enum for wallet destinations
type WalletTransferDestination string

// Values for WalletTransferDestination
const (
	// Spot wallet
	Spot WalletTransferDestination = "Spot Wallet"
	// Futures wallet
	Futures WalletTransferDestination = "Futures Wallet"
)

// RequestWalletTransfer required parameters
type RequestWalletTransferParameters struct {
	// Asset being transfered
	Asset string
	// Source wallet.
	//
	// Refer to WalletTransferDestination for values.
	From string
	// Destination wallet.
	//
	// Refer to WalletTransferDestination for values.
	To string
	// Amount to be transfered
	Amount string
}

// RequestWalletTransfer result
type RequestWalletTransferResult struct {
	// Reference ID
	ReferenceID string `json:"refid"`
}

// RequestWalletTransfer response
type RequestWalletTransferResponse struct {
	common.KrakenSpotRESTResponse
	// RequestWalletTransfer result
	Result *RequestWalletTransferResult `json:"result,omitempty"`
}
