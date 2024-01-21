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

// Unit test suite for CancelAllOrders DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type CancelAllOrdersTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestCancelAllOrdersTestSuite(t *testing.T) {
	suite.Run(t, new(CancelAllOrdersTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of CancelAllOrders.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding CancelAllOrdersResponse struct.
func (suite *CancelAllOrdersTestSuite) TestCancelAllOrdersUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "count": 4
		}
	}`
	expectedCount := 4
	// Unmarshal payload into struct
	response := new(CancelAllOrdersResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedCount, response.Result.Count)
}
