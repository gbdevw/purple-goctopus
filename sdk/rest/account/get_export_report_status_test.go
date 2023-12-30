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

// Unit test suite for GetExportReportStatus DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetExportReportStatusTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetExportReportStatusTestSuite(t *testing.T) {
	suite.Run(t, new(GetExportReportStatusTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetExportReportStatus.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetExportReportStatusResponse struct.
func (suite *GetExportReportStatusTestSuite) TestGetExportReportStatusUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": [
		  {
			"id": "VSKC",
			"descr": "my_trades_1",
			"format": "CSV",
			"report": "trades",
			"subtype": "all",
			"status": "Processed",
			"flags": "0",
			"fields": "all",
			"createdtm": "1688669085",
			"expiretm": "1688878685",
			"starttm": "1688669093",
			"completedtm": "1688669093",
			"datastarttm": "1683556800",
			"dataendtm": "1688669085",
			"aclass": "forex",
			"asset": "all"
		  },
		  {
			"id": "TCJA",
			"descr": "my_trades_1",
			"format": "CSV",
			"report": "trades",
			"subtype": "all",
			"status": "Processed",
			"flags": "0",
			"fields": "all",
			"createdtm": "1688363637",
			"expiretm": "1688573237",
			"starttm": "1688363664",
			"completedtm": "1688363664",
			"datastarttm": "1683235200",
			"dataendtm": "1688363637",
			"aclass": "forex",
			"asset": "all"
		  }
		]
	}`
	expectedCount := 2
	expectedItem2ID := "TCJA"
	expectedItem2Descr := "my_trades_1"
	expectedItem2Format := "CSV"
	expectedItem2Report := string(ReportTrades)
	expectedItem2Subtype := "all"
	expectedItem2Status := string(Processed)
	expectedItem2Fields := "all"
	expectedItem2CreatedTm := "1688363637"
	expectedItem2StartTm := "1688363664"
	expectedItem2CompletedTm := "1688363664"
	expectedItem2DataStartTm := "1683235200"
	expectedItem2DataEndTm := "1688363637"
	expectedItem2Asset := "all"
	// Unmarshal payload into struct
	response := new(GetExportReportStatusResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedItem2ID, response.Result[1].Id)
	require.Equal(suite.T(), expectedItem2Descr, response.Result[1].Description)
	require.Equal(suite.T(), expectedItem2Format, response.Result[1].Format)
	require.Equal(suite.T(), expectedItem2Report, response.Result[1].Report)
	require.Equal(suite.T(), expectedItem2Subtype, response.Result[1].SubType)
	require.Equal(suite.T(), expectedItem2Status, response.Result[1].Status)
	require.Equal(suite.T(), expectedItem2Fields, response.Result[1].Fields)
	require.Equal(suite.T(), expectedItem2CreatedTm, response.Result[1].CreatedTimestamp.String())
	require.Equal(suite.T(), expectedItem2StartTm, response.Result[1].StartTimestamp.String())
	require.Equal(suite.T(), expectedItem2CompletedTm, response.Result[1].CompletedTimestamp.String())
	require.Equal(suite.T(), expectedItem2DataStartTm, response.Result[1].DataStartTimestamp.String())
	require.Equal(suite.T(), expectedItem2DataEndTm, response.Result[1].DataEndTimestamp.String())
	require.Equal(suite.T(), expectedItem2Asset, response.Result[1].Asset)
}
