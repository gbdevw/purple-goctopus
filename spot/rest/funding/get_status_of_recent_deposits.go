package funding

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// GetStatusOfRecentDeposits request options
type GetStatusOfRecentDepositsRequestOptions struct {
	// Filter for specific asset being deposited.
	//
	// An empty value means no filter.
	Asset string `json:"asset,omitempty"`
	// Filter for specific name of deposit method.
	//
	// An empty value means no filter.
	Method string `json:"method,omitempty"`
	// Start timestamp, deposits created strictly before will not be included in the response.
	//
	// An empty value means no filter.
	Start string `json:"start,omitempty"`
	// End timestamp, deposits created stricly after will be not be included in the response.
	//
	// An empty value means no filter.
	End string `json:"end,omitempty"`
	// Cursor for next page of results (string)
	//
	// An empty value means first page/request.
	Cursor string `json:"cursor,omitempty"`
	// Number of results to include per page.
	//
	// Default to 25. A zero value will trigger the default behavior.
	Limit int64 `json:"limit,omitempty"`
}

// GetStatusOfRecentDeposits result
type GetStatusOfRecentDepositsResult struct {
	// Provides next input to use for cursor in pagination.
	NextCursor string `json:"next_cursor,omitempty"`
	// Listed deposits
	Deposits []TransactionDetails `json:"deposit"`
}

// GetStatusOfRecentDeposits response
type GetStatusOfRecentDepositsResponse struct {
	common.KrakenSpotRESTResponse
	// Recent deposits
	Result *GetStatusOfRecentDepositsResult `json:"result,omitempty"`
}
