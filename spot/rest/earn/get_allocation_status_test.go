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

// Unit test suite for GetAllocationStatus DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetAllocationStatusTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetAllocationStatusTestSuite(t *testing.T) {
	suite.Run(t, new(GetAllocationStatusTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetAllocationStatusResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetAllocationStatusResponse struct.
func (suite *GetAllocationStatusTestSuite) TestGetAllocationStatusResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "pending": true
		}
	}`
	// Unmarshal payload into struct
	response := new(GetAllocationStatusResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.True(suite.T(), response.Result.Pending)
}

// Test the JSON unmarshaller of GetAllocationStatusResponse when an invalid payload is used.
//
// The test will ensure:
//   - A invalid JSON response from the API cause a json.UnmarshalTypeError to be returned by the
//     json.Unmarshal function.
func (suite *GetAllocationStatusTestSuite) TestGetAllocationStatusResponseUnmarshalJSONError() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "pending": 64
		}
	}`
	// Unmarshal payload into struct and expect a json.UnmarshalTypeError
	response := new(GetAllocationStatusResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.Error(suite.T(), err)
	target := new(json.UnmarshalTypeError)
	require.ErrorAs(suite.T(), err, &target)
}
