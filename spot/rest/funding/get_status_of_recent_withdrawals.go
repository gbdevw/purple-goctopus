package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetStatusOfRecentWithdrawals request options
type GetStatusOfRecentWithdrawalsRequestOptions struct {
	// Filter for specific name of withdrawal method.
	//
	// An empty string means no filter.
	Method string `json:"method,omitempty"`
	// Filter for specific asset being withdrawn.
	//
	// An empty string means no filter.
	Asset string `json:"asset,omitempty"`
	// Start timestamp, deposits created strictly before will not be included in the response.
	//
	// An empty value means no filter.
	Start string `json:"start,omitempty"`
	// End timestamp, deposits created stricly after will be not be included in the response.
	//
	// An empty value means no filter.
	End string `json:"end,omitempty"`

	// Cursor and limits are disabled as documentation is unclear about what the payload looks like in case they are used.

	// Cursor for next page of results (string)
	//
	// An empty value means first page/request.
	//Cursor string `json:"cursor,omitempty"`
	// Number of results to include per page.
	//
	// Default to 25. A zero value will trigger the default behavior.
	//Limit int64 `json:"limit,omitempty"`
}

// Transaction details for a withdrawal
type Withdrawal struct {
	// Name of deposit method
	Method string `json:"method,omitempty"`
	// Network name based on the funding method used
	Network string `json:"network,omitempty"`
	// Asset class
	AssetClass string `json:"aclass,omitempty"`
	// Asset
	Asset string `json:"asset,omitempty"`
	// Reference ID
	ReferenceID string `json:"refid,omitempty"`
	// Method transaction ID
	TransactionID string `json:"txid,omitempty"`
	// Method transaction information
	Info string `json:"info,omitempty"`
	// Amount deposited/withdrawn
	Amount string `json:"amount"`
	// Fees paid. Can be empty
	Fee string `json:"fee,omitempty"`
	// Unix timestamp when request was made
	Time int64 `json:"time"`
	// Status of deposit - IFEX financial transaction states
	Status string `json:"status"`
	// Additional status property. Can be empty.
	StatusProperty string `json:"status-prop,omitempty"`
	// Withdrawal key name, as set up on your account
	Key string `json:"key,omitempty"`
}

// GetStatusOfRecentWithdrawals response
type GetStatusOfRecentWithdrawalsResponse struct {
	common.KrakenSpotRESTResponse
	// Recent withdrawals
	Result []*Withdrawal `json:"result"`
}
