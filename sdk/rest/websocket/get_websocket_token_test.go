package websocket

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetWebsocketToken DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetWebsocketTokenTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetWebsocketTokenTestSuite(t *testing.T) {
	suite.Run(t, new(GetWebsocketTokenTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetWebsocketTokenResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetWebsocketTokenResponse struct.
func (suite *GetWebsocketTokenTestSuite) TestGetWebsocketTokenResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [ ],
		"result": {
			"token": "1Dwc4lzSwNWOAwkMdqhssNNFhs1ed606d1WcF3XfEMw",
			"expires": 900
		}
	}`
	expectedToken := "1Dwc4lzSwNWOAwkMdqhssNNFhs1ed606d1WcF3XfEMw"
	expectedExpires := int64(900)
	// Unmarshal payload into struct
	response := new(GetWebsocketTokenResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedToken, response.Result.Token)
	require.Equal(suite.T(), expectedExpires, response.Result.Expires)
}
