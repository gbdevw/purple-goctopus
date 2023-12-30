package account

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for RequestExportReport DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type RequestExportReportTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestRequestExportReportTestSuite(t *testing.T) {
	suite.Run(t, new(RequestExportReportTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of RequestExportReport.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding RequestExportReportResponse struct.
func (suite *RequestExportReportTestSuite) TestRequestExportReportUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "id": "TCJA"
		}
	}`
	expectedID := "TCJA"
	// Unmarshal payload into struct
	response := new(RequestExportReportResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedID, response.Result.Id)
}
