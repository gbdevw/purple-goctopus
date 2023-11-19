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

// Unit test suite for GetSystemStatus DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetSystemStatusTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetSystemStatusTestSuite(t *testing.T) {
	suite.Run(t, new(GetSystemStatusTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetSystemStatusResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetSystemStatusResponse struct.
func (suite *GetSystemStatusTestSuite) TestGetSystemStatusResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "status": "online",
		  "timestamp": "2023-07-06T18:52:00Z"
		}
	}`
	expectedStatus := string(Online)
	expectedTimestamp := "2023-07-06T18:52:00Z"
	// Unmarshal payload into struct
	response := new(GetSystemStatusResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedStatus, response.Result.Status)
	require.Equal(suite.T(), expectedTimestamp, response.Result.Timestamp)
}
