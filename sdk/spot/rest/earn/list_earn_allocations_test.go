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

// Unit test suite for ListEarnAllocations DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type ListEarnAllocationsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestListEarnAllocationsTestSuite(t *testing.T) {
	suite.Run(t, new(ListEarnAllocationsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of ListEarnAllocationsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding ListEarnAllocationsResponse struct.
func (suite *ListEarnAllocationsTestSuite) TestListEarnAllocationsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "converted_asset": "USD",
		  "total_allocated": "49.2398",
		  "total_rewarded": "0.0675",
		  "next_cursor": "2",
		  "items": [
			{
			  "strategy_id": "ESDQCOL-WTZEU-NU55QF",
			  "native_asset": "ETH",
			  "amount_allocated": {
				"bonding": {
				  "native": "0.0210000000",
				  "converted": "39.0645",
				  "allocation_count": 2,
				  "allocations": [
					{
					  "created_at": "2023-07-06T10:52:05Z",
					  "expires": "2023-08-19T02:34:05.807Z",
					  "native": "0.0010000000",
					  "converted": "1.8602"
					},
					{
					  "created_at": "2023-08-01T11:25:52Z",
					  "expires": "2023-09-06T07:55:52.648Z",
					  "native": "0.0200000000",
					  "converted": "37.2043"
					}
				  ]
				},
				"total": {
				  "native": "0.0210000000",
				  "converted": "39.0645"
				}
			  },
			  "total_rewarded": {
				"native": "0",
				"converted": "0.0000"
			  }
			}
		  ]
		}
	}`
	expectedConvertedAsset := "USD"
	expectedItemsCount := 1
	expectedItem1StrategyId := "ESDQCOL-WTZEU-NU55QF"
	expectedBondingAllocation1Expires := "2023-08-19T02:34:05.807Z"
	expectedBondingAllocation1Rewardconverted := "1.8602"
	expectedTotalRewardedNative := "0"
	// Unmarshal payload into struct
	response := new(ListEarnAllocationsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedConvertedAsset, response.Result.ConvertedAsset)
	require.Len(suite.T(), response.Result.Items, expectedItemsCount)
	require.Equal(suite.T(), expectedItem1StrategyId, response.Result.Items[0].StrategyId)
	require.Nil(suite.T(), response.Result.Items[0].Payout)
	require.Len(suite.T(), response.Result.Items[0].AmountAllocated.Bonding.Allocations, response.Result.Items[0].AmountAllocated.Bonding.AllocationCount)
	require.Equal(suite.T(), expectedBondingAllocation1Expires, response.Result.Items[0].AmountAllocated.Bonding.Allocations[0].Expires)
	require.Equal(suite.T(), expectedBondingAllocation1Rewardconverted, response.Result.Items[0].AmountAllocated.Bonding.Allocations[0].Converted)
	require.Equal(suite.T(), expectedTotalRewardedNative, response.Result.Items[0].TotalRewarded.Native)
}

// Test the JSON unmarshaller of Payout. As the examples provided by the API doc. do not show a
// Payout payload, this test will ensure it is correctly handled when using a hand crafted example.
//
// The test will ensure:
//   - A JSON payload can be unmarshalled into the corresponding Payout struct.
func (suite *ListEarnAllocationsTestSuite) TestPayoutUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"period_start": "2023-07-06T10:52:05Z",
		"period_end": "2023-08-19T02:34:05.807Z",
		"accumulated_reward": {
			"native": "0.0210000000",
			"converted": "39.0645"
		},
		"estimated_reward": {
			"native": "0.0210000000",
			"converted": "39.0645"
		}
	}`
	expectedPeriodStart := "2023-07-06T10:52:05Z"
	expectedPeriodEnd := "2023-08-19T02:34:05.807Z"
	expectedNative := "0.0210000000"
	expectedConverted := "39.0645"
	// Unmarshal payload into struct
	response := new(Payout)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedPeriodStart, response.PeriodStart)
	require.Equal(suite.T(), expectedPeriodEnd, response.PeriodEnd)
	require.Equal(suite.T(), expectedNative, response.AccumulatedReward.Native)
	require.Equal(suite.T(), expectedNative, response.EstimatedReward.Native)
	require.Equal(suite.T(), expectedConverted, response.AccumulatedReward.Converted)
	require.Equal(suite.T(), expectedConverted, response.EstimatedReward.Converted)
}
