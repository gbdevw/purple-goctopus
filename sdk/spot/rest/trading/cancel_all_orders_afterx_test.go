package trading

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for CancelAllOrdersAfterX DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type CancelAllOrdersAfterXTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestCancelAllOrdersAfterXTestSuite(t *testing.T) {
	suite.Run(t, new(CancelAllOrdersAfterXTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of CancelAllOrdersAfterX.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding CancelAllOrdersAfterXResponse struct.
func (suite *CancelAllOrdersAfterXTestSuite) TestCancelAllOrdersAfterXUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "currentTime": "2023-03-24T17:41:56Z",
		  "triggerTime": "2023-03-24T17:42:56Z"
		}
	}`
	expectedCurrentTime := "2023-03-24T17:41:56Z"
	expectedTriggerTime := "2023-03-24T17:42:56Z"
	// Unmarshal payload into struct
	response := new(CancelAllOrdersAfterXResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedCurrentTime, response.Result.CurrentTime.Format(time.RFC3339))
	require.Equal(suite.T(), expectedTriggerTime, response.Result.TriggerTime.Format(time.RFC3339))
}
