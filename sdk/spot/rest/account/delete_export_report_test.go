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

// Unit test suite for DeleteExportReport DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type DeleteExportReportTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestDeleteExportReportTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteExportReportTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of DeleteExportReport.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding DeleteExportReportResponse struct.
func (suite *DeleteExportReportTestSuite) TestDeleteExportReportUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "delete": true
		}
	  }`
	// Unmarshal payload into struct
	response := new(DeleteExportReportResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.True(suite.T(), response.Result.Delete)
	require.False(suite.T(), response.Result.Cancel)
}
