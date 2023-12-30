package market

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetAssetInfo DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetAssetInfoTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetAssetInfoTestSuite(t *testing.T) {
	suite.Run(t, new(GetAssetInfoTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetAssetInfoResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetAssetInfoResponse struct.
func (suite *GetAssetInfoTestSuite) TestGetAssetInfoResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBT": {
			"aclass": "currency",
			"altname": "XBT",
			"decimals": 10,
			"display_decimals": 5,
			"collateral_value": 1,
			"status": "enabled"
		  },
		  "ZEUR": {
			"aclass": "currency",
			"altname": "EUR",
			"decimals": 4,
			"display_decimals": 2,
			"collateral_value": 1,
			"status": "enabled"
		  },
		  "ZUSD": {
			"aclass": "currency",
			"altname": "USD",
			"decimals": 4,
			"display_decimals": 2,
			"collateral_value": 1,
			"status": "enabled"
		  }
		}
	}`
	expectedResultCount := 3
	expectedZUSDAltname := "USD"
	// Unmarshal payload into struct
	response := new(GetAssetInfoResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedResultCount)
	require.Equal(suite.T(), expectedZUSDAltname, response.Result["ZUSD"].Altname)
}
