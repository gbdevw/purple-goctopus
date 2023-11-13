package earn

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Request parameters for AllocateFunds
type AllocateFundsRequestParameters struct {
	// The amount to allocate.
	Amount string `json:"amount"`
	// A unique identifier of the chosen earn strategy, as returned from ListEarnStrategies.
	StrategyId string `json:"strategy_id"`
}

// Response for AllocateFunds
type AllocateFundsResponse struct {
	common.KrakenSpotRESTResponse
	Result bool `json:"result"`
}
