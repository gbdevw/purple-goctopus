package funding

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetDepositMethods DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetDepositMethodsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetDepositMethodsTestSuite(t *testing.T) {
	suite.Run(t, new(GetDepositMethodsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetDepositMethodsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetDepositMethodsResponse struct.
func (suite *GetDepositMethodsTestSuite) TestGetDepositMethodsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": [
		  {
			"method": "Bitcoin",
			"limit": false,
			"fee": "0.0000000000",
			"gen-address": true,
			"minimum": "0.00010000"
		  },
		  {
			"method": "Bitcoin Lightning",
			"limit": false,
			"fee": "0.00000000",
			"minimum": "0.00010000"
		  }
		]
	}`
	expectedCount := 2
	expectedItem1Method := "Bitcoin"
	expectedItem1Limit := "false"
	expectedItem1Fee := "0.0000000000"
	expectedItem1GenAddress := true
	expectedItem1Minimum := "0.00010000"
	// Unmarshal payload into struct
	response := new(GetDepositMethodsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.NotEmpty(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedItem1Method, response.Result[0].Method)
	require.Equal(suite.T(), expectedItem1Limit, response.Result[0].Limit)
	require.Equal(suite.T(), expectedItem1Fee, response.Result[0].Fee)
	require.Equal(suite.T(), expectedItem1GenAddress, response.Result[0].GenAddress)
	require.Equal(suite.T(), expectedItem1Minimum, response.Result[0].Minimum)
}
