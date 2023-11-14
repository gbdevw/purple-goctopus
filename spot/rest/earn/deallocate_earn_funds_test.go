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

// Unit test suite for DeallocateFunds DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type DeallocateFundsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestDeallocateFundsTestSuite(t *testing.T) {
	suite.Run(t, new(DeallocateFundsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of DeallocateFundsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding DeallocateFundsResponse struct.
func (suite *DeallocateFundsTestSuite) TestDeallocateFundsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": true
	}`
	// Unmarshal payload into struct
	response := new(DeallocateFundsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.True(suite.T(), response.Result)
}
