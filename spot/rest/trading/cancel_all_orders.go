package trading

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Response for Cancel All Orders
type CancelAllOrdersResponse struct {
	common.KrakenSpotRESTResponse
	Result struct {
		// Number of canceled orders
		Count int `json:"count"`
	} `json:"result,omitempty"`
}
