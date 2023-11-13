package account

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// QueryOrdersParameters contains Auery Orders required parameters
type QueryOrdersParameters struct {
	// List of transaction IDs to query info about (50 maximum).
	TransactionIds []string
}

// QueryOrdersOptions contains Query Orders optional parameters.
type QueryOrdersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Restrict results to given user reference id.
	UserReference *int64
}

// QueryOrdersInfoResponse contains Query Orders Info response data.
type QueryOrdersInfoResponse struct {
	common.KrakenSpotRESTResponse
	// Map where keys are transaction ID and values the requested orders
	Result map[string]OrderInfo `json:"result,omitempty"`
}
