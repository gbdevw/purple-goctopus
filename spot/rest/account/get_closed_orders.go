package account

import (
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// CloseTime Enum
type CloseTimeEnum string

// Values for CloseTimeEnum
const (
	UseOpen  CloseTimeEnum = "open"
	UseClose CloseTimeEnum = "close"
	UseBoth  CloseTimeEnum = "both"
)

// GetClosedOrders request options.
type GetClosedOrdersOptions struct {
	// Whether or not to include trades related to position in output.
	//
	// Defaults to false.
	Trades bool `json:"trades"`
	// Restrict results to given user reference id.
	//
	// A nil value means no user reference will be provided.
	UserReference *int64 `json:"userref"`
	// Starting unix timestamp (seconds) or order tx ID of results (exclusive).
	//
	// A zero value means no filtering based on a start date.
	Start int64 `json:"start"`
	// Ending unix timestamp (seconds) or order tx ID of results (inclusive)
	//
	// A zero value means no filtering based on a end date.
	End int64 `json:"end"`
	// Result offset for pagination
	//
	// A zero value means the first records will be fetched.
	Offset int64 `json:"ofs"`
	// Which time to use to search.
	//
	// Defaults to "both" in case an empty string is provided. Cf. CloseTimeEnum.
	Closetime string `json:"closetime"`
	// Whether or not to consolidate trades by individual taker trades.
	//
	// Defaults to false.
	ConsolidateTaker bool `json:"consolidate_taker"`
}

// GetClosedOrders results.
type GetClosedOrdersResult struct {
	// Map where keys are transaction ID and values the related closed orders.
	Closed map[string]OrderInfo `json:"closed"`
	// Amount of available order info matching criteria.
	Count int `json:"count"`
}

// GetClosedOrders response.
type GetClosedOrdersResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetClosedOrdersResult `json:"result,omitempty"`
}
