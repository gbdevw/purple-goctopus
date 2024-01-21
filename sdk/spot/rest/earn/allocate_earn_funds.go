package earn

import "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"

// Request parameters for AllocateEarnFunds
type AllocateEarnFundsRequestParameters struct {
	// The amount to allocate.
	Amount string `json:"amount"`
	// A unique identifier of the chosen earn strategy, as returned from ListEarnStrategies.
	StrategyId string `json:"strategy_id"`
}

// Response for AllocateEarnFunds
type AllocateEarnFundsResponse struct {
	common.KrakenSpotRESTResponse
	Result bool `json:"result"`
}
