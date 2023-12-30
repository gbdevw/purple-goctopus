package earn

import "github.com/gbdevw/purple-goctopus/spot/rest/common"

// Amount earned using the strategy during the whole lifetime of user account.
type Reward struct {
	// Amount converted into the requested asset.
	Converted string `json:"converted"`
	// Amount in the native asset.
	Native string `json:"native"`
}

// Information about the current payout period.
type Payout struct {
	// Reward accumulated in the payout period until now.
	AccumulatedReward Reward `json:"accumulated_reward"`
	// Estimated reward from now until the payout.
	EstimatedReward Reward `json:"estimated_reward"`
	// Tentative date of the next reward payout.
	PeriodEnd string `json:"period_end"`
	// When the current payout period started. Either the date of the last payout or when it was enabled.
	PeriodStart string `json:"period_start"`
}

// Amount allocated to a state
type AllocatedAmount struct {
	Reward
	// The total number of allocations in this state for this asset.
	AllocationCount int `json:"allocation_count"`
	// Details about when each allocation will expire and move to the next state
	Allocations []AllocationDetails `json:"allocations"`
}

// Details about when each allocation will expire and move to the next state.
type AllocationDetails struct {
	Reward
	// The date and time which a request to either allocate or deallocate was received and processed.
	//
	// For a deallocation request to a strategy with an exit-queue, this will be the time the funds
	// joined the exit queue. For a deallocation request to a strategy without exit queue, this
	// will be the time the funds started unbonding.
	CreatedAt string `json:"created_at"`
	// The date/time the funds will be unbonded.
	Expires string `json:"expires"`
}

// Amounts allocated per state.
type Allocations struct {
	// Amount allocated in bonding status.
	Bonding *AllocatedAmount `json:"bonding,omitempty"`
	// Amount allocated in the exit-queue status
	ExitQueue *AllocatedAmount `json:"exit_queue,omitempty"`
	// Pending allocation amount - can be negative if the pending operation is deallocation
	Pending *Reward `json:"pending,omitempty"`
	// Amount allocated in unbonding status.
	Unbonding *AllocatedAmount `json:"unbonding,omitempty"`
	// Total amount allocated to this Earn strategy
	Total Reward `json:"total"`
}

// Amounts allocated to an earn strategy.
type EarnAllocation struct {
	// Amounts allocated to this Earn strategy
	AmountAllocated Allocations `json:"amount_allocated"`
	// The asset of the native currency of this allocation
	NativeAsset string `json:"native_asset"`
	// Information about the current payout period (if any)
	Payout *Payout `json:"payout,omitempty"`
	// Unique ID for Earn Strategy
	StrategyId string `json:"strategy_id"`
	// Amount earned using the strategy during the whole lifetime of user account
	TotalRewarded Reward `json:"total_rewarded"`
}

// Request options for ListEarnAllocations request.
type ListEarnAllocationsRequestOptions struct {
	// true to sort ascending, false (the default) for descending.
	Ascending bool `json:"ascending"`
	// A secondary currency to express the value of your allocations (the default is USD).
	//
	// An empty value triggers the default behavior.
	ConvertedAsset string `json:"converted_asset,omitempty"`
	// Omit entries for strategies that were used in the past but now they don't hold any allocation (the default is false).
	HideZeroAllocations bool `json:"hide_zero_allocations"`
}

// Result for ListEarnAllocations
type ListEarnAllocationsResult struct {
	// A secondary asset to show the value of allocations. (Eg. you also want to see the value of
	// your allocations in USD). Choose this in the request parameters.
	ConvertedAsset string `json:"converted_asset"`
	// Earn allocations by startegy.
	Items []*EarnAllocation `json:"items"`
	// The total amount allocated across all strategies, denominated in the converted_asset currency.
	TotalAllocated string `json:"total_allocated"`
	// Amount earned across all strategies during the whole lifetime of user account, denominated
	// in converted_asset currency.
	TotalRewarded string `json:"total_rewarded"`
}

// Response for ListEarnAllocations
type ListEarnAllocationsResponse struct {
	common.KrakenSpotRESTResponse
	Result *ListEarnAllocationsResult `json:"result,omitempty"`
}
