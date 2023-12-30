package account

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetExtendedBalance DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetExtendedBalanceTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetExtendedBalanceTestSuite(t *testing.T) {
	suite.Run(t, new(GetExtendedBalanceTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetExtendedBalance.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetExtendedBalanceResponse struct.
func (suite *GetExtendedBalanceTestSuite) TestGetExtendedBalanceUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "ZUSD": {
			"balance": 25435.21,
			"hold_trade": 8249.76
		  },
		  "XXBT": {
			"balance": 1.2435,
			"hold_trade": 0.8423
		  }
		}
	}`
	expectedCount := 2
	expectedZUSDBalance := "25435.21"
	expectedXXBTHoldTrade := "0.8423"
	// Unmarshal payload into struct
	response := new(GetExtendedBalanceResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedZUSDBalance, response.Result["ZUSD"].Balance.String())
	require.Equal(suite.T(), expectedXXBTHoldTrade, response.Result["XXBT"].HoldTrade.String())
	require.Empty(suite.T(), response.Result["XXBT"].Credit)
	require.Empty(suite.T(), response.Result["XXBT"].CreditUsed)
}
