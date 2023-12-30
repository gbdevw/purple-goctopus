package earn

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Request parameters for DeallocateEarnFunds
type DeallocateEarnFundsRequestParameters struct {
	// The amount to deallocate.
	Amount string `json:"amount"`
	// A unique identifier of the chosen earn strategy, as returned from ListEarnStrategies.
	StrategyId string `json:"strategy_id"`
}

// Response for DeallocateEarnFunds
type DeallocateEarnFundsResponse struct {
	common.KrakenSpotRESTResponse
	Result bool `json:"result"`
}
