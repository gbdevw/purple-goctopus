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

// Test the JSON unmarshaller of DeallocateFundsResponse when an invalid payload is used.
//
// The test will ensure:
//   - A invalid JSON response from the API cause a json.UnmarshalTypeError to be returned by the
//     json.Unmarshal function.
func (suite *DeallocateFundsTestSuite) TestDeallocateFundsResponseUnmarshalJSONError() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": 64
	}`
	// Unmarshal payload into struct and expect a json.UnmarshalTypeError
	response := new(DeallocateFundsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.Error(suite.T(), err)
	target := new(json.UnmarshalTypeError)
	require.ErrorAs(suite.T(), err, &target)
}
