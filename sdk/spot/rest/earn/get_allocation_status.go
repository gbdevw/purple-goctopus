package earn

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// Request parameters for GetAllocationStatus
type GetAllocationStatusRequestParameters struct {
	// A unique identifier of the chosen earn strategy, as returned from ListEarnStrategies.
	StrategyId string `json:"strategy_id"`
}

// Result for GetAllocationStatus
type GetAllocationStatusResult struct {
	// true if an operation is still in progress on the same strategy.
	Pending bool `json:"pending"`
}

// Response for GetAllocationStatus
type GetAllocationStatusResponse struct {
	common.KrakenSpotRESTResponse
	Result *GetAllocationStatusResult `json:"result,omitempty"`
}
