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

// Unit test suite for GetWithdrawalAddresses DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetWithdrawalAddressesTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetWithdrawalAddressesTestSuite(t *testing.T) {
	suite.Run(t, new(GetWithdrawalAddressesTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetWithdrawalAddressesResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetWithdrawalAddressesResponse struct.
func (suite *GetWithdrawalAddressesTestSuite) TestGetWithdrawalAddressesResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": [
		  {
			"address": "bc1qxdsh4sdd29h6ldehz0se5c61asq8cgwyjf2y3z",
			"asset": "XBT",
			"method": "Bitcoin",
			"key": "btc-wallet-1",
			"verified": true
		  }
		]
	  }`
	expectedCount := 1
	expectedAddress := "bc1qxdsh4sdd29h6ldehz0se5c61asq8cgwyjf2y3z"
	expectedItem0Asset := "XBT"
	expectedItem0Method := "Bitcoin"
	expectedItem0Key := "btc-wallet-1"
	// Unmarshal payload into struct
	response := new(GetWithdrawalAddressesResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedItem0Asset, response.Result[0].Asset)
	require.Equal(suite.T(), expectedItem0Method, response.Result[0].Method)
	require.Equal(suite.T(), expectedAddress, response.Result[0].Address)
	require.Equal(suite.T(), expectedItem0Key, response.Result[0].Key)
}
