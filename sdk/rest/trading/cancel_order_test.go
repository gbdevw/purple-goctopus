package trading

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for CancelOrder DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type CancelOrderTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestCancelOrderTestSuite(t *testing.T) {
	suite.Run(t, new(CancelOrderTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of CancelOrder.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding CancelOrderResponse struct.
func (suite *CancelOrderTestSuite) TestCancelOrderUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "count": 1
		}
	}`
	expectedCount := 1
	// Unmarshal payload into struct
	response := new(CancelOrderResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedCount, response.Result.Count)
}
