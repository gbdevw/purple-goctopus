package earn

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for AllocateFunds DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type AllocateFundsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestAllocateFundsTestSuite(t *testing.T) {
	suite.Run(t, new(AllocateFundsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of AllocateFundsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding AllocateFundsResponse struct.
func (suite *AllocateFundsTestSuite) TestAllocateFundsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": true
	}`
	// Unmarshal payload into struct
	response := new(AllocateEarnFundsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.True(suite.T(), response.Result)
}
