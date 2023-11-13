package account

import (
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Type for CloseTime
type CloseTime string

// Values for CloseTime
const (
	UseOpen  CloseTime = "open"
	UseClose CloseTime = "close"
	UseBoth  CloseTime = "both"
)

// GetClosedOrdersOptions contains Get Closed Orders optional parameters.
type GetClosedOrdersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Restrict results to given user reference id.
	UserReference *int64
	// Starting unix timestamp or order tx ID of results (exclusive)
	Start *time.Time
	// Ending unix timestamp or order tx ID of results (inclusive)
	End *time.Time
	// Result offset for pagination
	Offset *int64
	// Which time to use to search.
	// Defaults to "both". Values: "open" "close" "both"
	Closetime CloseTime
}

// Closed orders
type ClosedOrders struct {
	// Map where keys are transaction ID and values the related closed order.
	Closed map[string]OrderInfo `json:"closed"`
	// Amount of available order info matching criteria.
	Count int `json:"count"`
}

// GetClosedOrdersResponse contains Get Closed Orders response data.
type GetClosedOrdersResponse struct {
	common.KrakenSpotRESTResponse
	Result *ClosedOrders `json:"result,omitempty"`
}
