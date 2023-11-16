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

// Unit test suite for GetDepositAddresses DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetDepositAddressesTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetDepositAddressesTestSuite(t *testing.T) {
	suite.Run(t, new(GetDepositAddressesTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetDepositAddressesResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetDepositAddressesResponse struct.
func (suite *GetDepositAddressesTestSuite) TestGetDepositAddressesResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": [
		  {
			"address": "2N9fRkx5JTWXWHmXzZtvhQsufvoYRMq9ExV",
			"expiretm": "0",
			"new": true
		  },
		  {
			"address": "2NCpXUCEYr8ur9WXM1tAjZSem2w3aQeTcAo",
			"expiretm": "0",
			"new": true
		  },
		  {
			"address": "2Myd4eaAW96ojk38A2uDK4FbioCayvkEgVq",
			"expiretm": "0"
		  },
		  {
			"address": "rLHzPsX3oXdzU2qP17kHCH2G4csZv1rAJh",
			"expiretm": "0",
			"new": true,
			"tag": "1361101127"
		  },
		  {
			"address": "krakenkraken",
			"expiretm": "0",
			"memo": "4150096490"
		  }
		]
	}`
	expectedCount := 5
	expectedItem1Address := "2N9fRkx5JTWXWHmXzZtvhQsufvoYRMq9ExV"
	expectedItem2Expire := int64(0)
	expectedItem4Tag := "1361101127"
	expectedItem5Memo := "4150096490"
	// Unmarshal payload into struct
	response := new(GetDepositAddressesResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.NotEmpty(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedItem1Address, response.Result[0].Address)
	require.Equal(suite.T(), expectedItem2Expire, response.Result[1].Expiretm)
	require.Equal(suite.T(), expectedItem4Tag, response.Result[3].Tag)
	require.Equal(suite.T(), expectedItem5Memo, response.Result[4].Memo)
}
