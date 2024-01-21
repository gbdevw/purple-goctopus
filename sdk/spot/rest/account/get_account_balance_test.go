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

// Unit test suite for GetAccountBalance DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetAccountBalanceTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetAccountBalanceTestSuite(t *testing.T) {
	suite.Run(t, new(GetAccountBalanceTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetAccountBalance.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetAccountBalanceResponse struct.
func (suite *GetAccountBalanceTestSuite) TestGetAccountBalanceUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "ZUSD": "171288.6158",
		  "ZEUR": "504861.8946",
		  "XXBT": "1011.1908877900",
		  "XETH": "818.5500000000",
		  "USDT": "500000.00000000",
		  "DAI": "9999.9999999999",
		  "DOT": "2.5000000000",
		  "ETH2.S": "198.3970800000",
		  "ETH2": "2.5885574330",
		  "USD.M": "1213029.2780"
		}
	  }`
	expectedCount := 10
	expectedXXBTBalance := "1011.1908877900"
	expectedDAIBalance := "9999.9999999999"
	// Unmarshal payload into struct
	response := new(GetAccountBalanceResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedXXBTBalance, response.Result["XXBT"].String())
	require.Equal(suite.T(), expectedDAIBalance, response.Result["DAI"].String())
}
