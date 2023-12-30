package earn

import (
	"encoding/json"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for AutoCompound.Type
type AutoCompoundType string

// Values for AutoCompoundType
const (
	Enabled  AutoCompoundType = "enabled"
	Disabled AutoCompoundType = "disabled"
	Optional AutoCompoundType = "optional"
)

// Auto compound choices for the earn strategy.
type AutoCompound struct {
	// Autocompound type
	//
	// Cf AutoCompoundType for values
	Type string `json:"type"`
	// Default behavior when auto compound is optional.
	//
	// Use that value only if type is optional.
	Default bool `json:"default"`
}

// Enum for yield source
type YieldSourceEnum string

// Values for YieldSource
const (
	OffChain YieldSourceEnum = "off_chain"
	Staking  YieldSourceEnum = "staking"
)

// Yield generation mechanism of an earn strategy
type YieldSource struct {
	Type string `json:"type"`
}

// Enum for LockType
type LockTypeEnum string

// Values for LockTypeEnum
const (
	Instant LockTypeEnum = "instant"
	Bonded  LockTypeEnum = "bonded"
	Timed   LockTypeEnum = "timed"
	Flex    LockTypeEnum = "flex"
)

// Additional data of a bonded lock type
type BondedLockType struct {
	// Duration of the bonding period, in seconds
	BondingPeriod int64 `json:"bonding_period,omitempty"`
	// Is the bonding period length variable (true) or static (false; default)
	BondingPeriodVariable bool `json:"bonding_period_variable,omitempty"`
	// Whether rewards are earned during the bonding period (payouts occur after bonding is complete).
	BondingRewards bool `json:"bonding_rewards,omitempty"`
	// In order to remove funds, if this value is greater than 0, funds will first have to enter an
	//  exit queue and will have to wait for the exit queue period to end. Once ended, her funds
	// will then follow and respect the unbonding_period.
	//
	// If the value of the exit queue period is 0, then no waiting will have to occur and the exit
	// queue will be skipped.
	//
	// Rewards are always paid out for the exit queue
	ExitQueuePeriod int64 `json:"exit_queue_period,omitempty"`
	// Duration of the unbonding period in seconds. In order to remove funds, you must wait for
	// the unbonding period to pass after requesting removal before funds become available in
	// her spot wallet.
	UnbondingPeriod int64 `json:"unbonding_period,omitempty"`
	// Is the unbonding period length variable (true) or static (false; default)
	UnbondingPeriodVariable bool `json:"unbonding_period_variable,omitempty"`
	// Whether rewards are earned and payouts are done during the unbonding period.
	UnbondingRewards bool `json:"unbonding_rewards,omitempty"`
}

// Additional data of a timed lock type
type TimedLockType struct {
	// Funds are locked for this duration, in seconds
	Duration int64 `json:"duration,omitempty"`
}

// Lock type data. Additional data will be available in the corresponding embedded structs in case
// type is instant, timed or bonded. Other structs will remain nil.
type LockType struct {
	// Lock type.
	//
	// Cf LockTypeEnum for values.
	Type string `json:"type"`
	BondedLockType
	TimedLockType
	// At what intervals are rewards distributed and credited to the user’s ledger.
	// If 0 or absent, then the payout happens at the end of lock duration, in seconds.
	//
	// Used by both instant, bonded and timed lock types.
	PayoutFrequency int64 `json:"payout_frequency,omitempty"`
}

// Estimate for the revenues from the strategy.
//
// The estimate is based on previous revenues from the strategy.
type APREstimate struct {
	// Maximal yield percentage for one year
	High string `json:"high"`
	// Minimal yield percentage for one year
	Low string `json:"low"`
}

// Data of a single earn strategy
type EarnStrategy struct {
	// Fee applied when allocating to this strategy.
	AllocationFee json.Number `json:"allocation_fee"`
	// Reason list why user is not eligible for allocating to the strategy.
	AllocationRestrictionInfo []string `json:"allocation_restriction_info"`
	// Estimate for the revenues from the strategy.
	//
	// The estimate is based on previous revenues from the strategy.
	APREstimate *APREstimate `json:"apr_estimate,omitempty"`
	// The asset to invest for this earn strategy
	Asset string `json:"asset"`
	// Auto compound choices for the earn strategy.
	AutoCompound AutoCompound `json:"auto_compound"`
	// Is allocation available for this strategy
	CanAllocate bool `json:"can_allocate"`
	// Is deallocation available for this strategy
	CanDeallocate bool `json:"can_deallocate"`
	// Fee applied when deallocating from this strategy
	DeallocationFee json.Number `json:"deallocation_fee"`
	// The unique identifier for this strategy
	Id string `json:"id"`
	// lock_type
	LockType LockType `json:"lock_type"`
	// The maximum amount of funds that any given user may allocate to an account. Absence of value
	// means there is no limit. Zero means that all new allocations will return an error (though
	// auto-compound is unaffected).
	UserCap string `json:"user_cap,omitempty"`
	// Minimum amount (in USD) for an allocation or deallocation
	UserMinAllocation string `jsoon:"user_min_allocation,omitempty"`
	// Yield generation mechanism of this strategy
	YieldSource YieldSource `json:"yield_source"`
}

// Request options for ListEarnStrategies
type ListEarnStrategiesRequestOptions struct {
	// True to sort ascending, false (the default) for descending.
	Ascending bool `json:"ascending"`
	// Filter strategies by asset name.
	//
	// An empty value means no filter.
	Asset string `json:"asset"`
	// Empty to start at beginning/end, otherwise next page ID.
	Cursor string `json:"cursor"`
	// How many items to return per page. Note that the limit may be cap'd to lower value in the
	// application code.
	//
	// A zero value triggers the default behavior.
	Limit int `json:"limit"`
	// Filter strategies by lock type.
	//
	// An empty value means no filter.
	LockType []string `json:"lock_type"`
}

// Result for ListEarnStrategies
type ListEarnStrategiesResult struct {
	// Index to send into Cursor for next page, absence of value or an empty value means you've
	// reached the end.
	NextCursor string `json:"next_cursor,omitempty"`
	// Listed earn strategies.
	Items []EarnStrategy `json:"items"`
}

// Response for ListEarnStrategies
type ListEarnStrategiesResponse struct {
	common.KrakenSpotRESTResponse
	Result *ListEarnStrategiesResult `json:"result,omitempty"`
}
