package funding

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// Enum for wallet destinations
type WalletTransferDestination string

// Values for WalletTransferDestination
const (
	// Spot wallet
	Spot WalletTransferDestination = "Spot Wallet"
	// Futures wallet
	Futures WalletTransferDestination = "Futures Wallet"
)

// RequestWalletTransfer request parameters
type RequestWalletTransferRequestParameters struct {
	// Asset being transfered
	Asset string `json:"asset"`
	// Source wallet.
	//
	// Refer to WalletTransferDestination for values.
	From string `json:"from"`
	// Destination wallet.
	//
	// Refer to WalletTransferDestination for values.
	To string `json:"to"`
	// Amount to be transfered
	Amount string `json:"amount"`
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
