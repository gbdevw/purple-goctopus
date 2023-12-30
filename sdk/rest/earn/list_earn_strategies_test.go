package earn

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for ListEarnStrategies DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type ListEarnStrategiesTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestListEarnStrategiesTestSuite(t *testing.T) {
	suite.Run(t, new(ListEarnStrategiesTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of ListEarnStrategiesResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding ListEarnStrategiesResponse struct.
func (suite *ListEarnStrategiesTestSuite) TestListEarnStrategiesResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "next_cursor": "2",
		  "items": [
			{
			  "id": "ESRFUO3-Q62XD-WIOIL7",
			  "asset": "DOT",
			  "lock_type": {
				"type": "instant",
				"payout_frequency": 604800
			  },
			  "apr_estimate": {
				"low": "8.0000",
				"high": "12.0000"
			  },
			  "user_min_allocation": "0.01",
			  "allocation_fee": "0.0000",
			  "deallocation_fee": "0.0000",
			  "auto_compound": {
				"type": "enabled"
			  },
			  "yield_source": {
				"type": "staking"
			  },
			  "can_allocate": true,
			  "can_deallocate": true,
			  "allocation_restriction_info": []
			}
		  ]
		}
	}`
	expectedNextCursor := "2"
	expectedItemsCount := 1
	expectedItem1StrategyId := "ESRFUO3-Q62XD-WIOIL7"
	expectedItem1LockType := string(Instant)
	expectedItem1LockTypePayoutFreq := int64(604800)
	expectedItem1APREstimateLow := "8.0000"
	expectedItem1AutoCompoundType := string(Enabled)
	expectedItem1YieldSource := string(Staking)
	expectedCanDeallocate := true
	// Unmarshal payload into struct
	response := new(ListEarnStrategiesResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedNextCursor, response.Result.NextCursor)
	require.Len(suite.T(), response.Result.Items, expectedItemsCount)
	require.Equal(suite.T(), expectedItem1StrategyId, response.Result.Items[0].Id)
	require.Empty(suite.T(), response.Result.Items[0].AllocationRestrictionInfo)
	require.Equal(suite.T(), expectedItem1LockType, response.Result.Items[0].LockType.Type)
	require.Equal(suite.T(), expectedItem1LockTypePayoutFreq, response.Result.Items[0].LockType.PayoutFrequency)
	require.NotNil(suite.T(), response.Result.Items[0].APREstimate)
	require.Equal(suite.T(), expectedItem1APREstimateLow, response.Result.Items[0].APREstimate.Low)
	require.Equal(suite.T(), expectedItem1AutoCompoundType, response.Result.Items[0].AutoCompound.Type)
	require.Equal(suite.T(), expectedItem1YieldSource, response.Result.Items[0].YieldSource.Type)
	require.Equal(suite.T(), expectedCanDeallocate, response.Result.Items[0].CanDeallocate)
}

// Test the JSON unmarshaller of LockType for a bonded lock type.
//
// The test will ensure:
//   - A valid JSON payload can be unmarshalled into the corresponding LockType struct for
//     a bonded lock type.
func (suite *ListEarnStrategiesTestSuite) TestBondedLockTypeUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"type": "bonded",
		"bonding_period": 42,
		"bonding_rewards": true,
		"exit_queue_period": 42,
		"payout_frequency": 604800,
		"unbonding_period": 42,
		"unbonding_rewards": true
	}`
	expectedType := string(Bonded)
	expectedBondingPeriod := int64(42)
	expectedExitQueuePeriod := int64(42)
	expectedPayoutFreq := int64(604800)
	expectedUnbondingPeriod := int64(42)
	// Unmarshal payload into struct
	response := new(LockType)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedType, response.Type)
	require.Equal(suite.T(), expectedBondingPeriod, response.BondingPeriod)
	require.Equal(suite.T(), expectedExitQueuePeriod, response.ExitQueuePeriod)
	require.Equal(suite.T(), expectedPayoutFreq, response.PayoutFrequency)
	require.Equal(suite.T(), expectedUnbondingPeriod, response.UnbondingPeriod)
	require.False(suite.T(), response.BondingPeriodVariable)
	require.False(suite.T(), response.UnbondingPeriodVariable)
	require.True(suite.T(), response.UnbondingRewards)
	require.True(suite.T(), response.BondingRewards)
}
