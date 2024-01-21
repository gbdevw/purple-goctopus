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

// Unit test suite for GetServerTime DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetServerTimeTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetServerTimeTestSuite(t *testing.T) {
	suite.Run(t, new(GetServerTimeTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetServerTimeResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetServerTimeResponse struct.
func (suite *GetServerTimeTestSuite) TestGetServerTimeResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "unixtime": 1688669448,
		  "rfc1123": "Thu, 06 Jul 23 18:50:48 +0000"
		}
	  }`
	expectedUnixTime := int64(1688669448)
	expectedRfc1123 := "Thu, 06 Jul 23 18:50:48 +0000"
	// Unmarshal payload into struct
	response := new(GetServerTimeResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedUnixTime, response.Result.Unixtime)
	require.Equal(suite.T(), expectedRfc1123, response.Result.Rfc1123)
}
