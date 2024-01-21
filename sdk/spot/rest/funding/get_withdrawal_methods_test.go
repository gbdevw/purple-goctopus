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

// Unit test suite for GetWithdrawalMethods DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetWithdrawalMethodsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetWithdrawalMethodsTestSuite(t *testing.T) {
	suite.Run(t, new(GetWithdrawalMethodsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetWithdrawalMethodsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetWithdrawalMethodsResponse struct.
func (suite *GetWithdrawalMethodsTestSuite) TestGetWithdrawalMethodsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": [
		  {
			"asset": "XXBT",
			"method": "Bitcoin",
			"network": "Bitcoin",
			"minimum": "0.0004"
		  },
		  {
			"asset": "XXBT",
			"method": "Bitcoin Lightning",
			"network": "Lightning",
			"minimum": "0.00001"
		  }
		]
	}`
	expectedCount := 2
	expectedItem0Asset := "XXBT"
	expectedItem0Method := "Bitcoin"
	expectedItem0Network := "Bitcoin"
	expectedItem0Min := "0.0004"
	// Unmarshal payload into struct
	response := new(GetWithdrawalMethodsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedItem0Asset, response.Result[0].Asset)
	require.Equal(suite.T(), expectedItem0Method, response.Result[0].Method)
	require.Equal(suite.T(), expectedItem0Network, response.Result[0].Network)
	require.Equal(suite.T(), expectedItem0Min, response.Result[0].Minimum.String())
}
