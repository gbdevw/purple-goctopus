package earn

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// Request parameters for GetDeallocationStatus
type GetDeallocationStatusRequestParameters struct {
	// A unique identifier of the chosen earn strategy, as returned from ListEarnStrategies.
	StrategyId string `json:"strategy_id"`
}

// Result for GetDeallocationStatus
type GetDeallocationStatusResult struct {
	// true if an operation is still in progress on the same strategy.
	Pending bool `json:"pending"`
}

// Response for GetDeallocationStatus
type GetDeallocationStatusResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetDeallocationStatusResult `json:"result,omitempty"`
}
