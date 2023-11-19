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

// Unit test suite for CancelOrderBatch DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type CancelOrderBatchTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestCancelOrderBatchTestSuite(t *testing.T) {
	suite.Run(t, new(CancelOrderBatchTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of CancelOrderBatch.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding CancelOrderBatchResponse struct.
func (suite *CancelOrderBatchTestSuite) TestCancelOrderBatchUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "count": 4
		}
	}`
	expectedCount := 4
	// Unmarshal payload into struct
	response := new(CancelOrderBatchResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedCount, response.Result.Count)
}
