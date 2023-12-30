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

// Unit test suite for RequestWithdrawalCancellation DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type RequestWithdrawalCancellationTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestRequestWithdrawalCancellationTestSuite(t *testing.T) {
	suite.Run(t, new(RequestWithdrawalCancellationTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of RequestWithdrawalCancellationResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding RequestWithdrawalCancellationResponse struct.
func (suite *RequestWithdrawalCancellationTestSuite) TestRequestWithdrawalCancellationResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": true
	}`
	// Unmarshal payload into struct
	response := new(RequestWithdrawalCancellationResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.True(suite.T(), response.Result)
}
