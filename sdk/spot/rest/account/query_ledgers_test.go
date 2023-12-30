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

// Unit test suite for QueryLedgers DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type QueryLedgersTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestQueryLedgersTestSuite(t *testing.T) {
	suite.Run(t, new(QueryLedgersTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of QueryLedgers.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding QueryLedgersResponse struct.
func (suite *QueryLedgersTestSuite) TestQueryLedgersUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "L4UESK-KG3EQ-UFO4T5": {
			"refid": "TJKLXF-PGMUI-4NTLXU",
			"time": 1688464484.1787,
			"type": "trade",
			"subtype": "",
			"aclass": "currency",
			"asset": "ZGBP",
			"amount": "-24.5000",
			"fee": "0.0490",
			"balance": "459567.9171"
		  }
		}
	}`
	expectedCount := 1
	expectedLedgerId := "L4UESK-KG3EQ-UFO4T5"
	expectedRefId := "TJKLXF-PGMUI-4NTLXU"
	expectedTime := "1688464484.1787"
	expectedType := string(EntryTypeTrade)
	expectedSubtype := ""
	expectedAclass := "currency"
	expectedAsset := "ZGBP"
	expectedAmount := "-24.5000"
	expectedFee := "0.0490"
	expectedBalance := "459567.9171"
	// Unmarshal payload into struct
	response := new(QueryLedgersResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedRefId, response.Result[expectedLedgerId].ReferenceId)
	require.Equal(suite.T(), expectedTime, response.Result[expectedLedgerId].Timestamp.String())
	require.Equal(suite.T(), expectedType, response.Result[expectedLedgerId].Type)
	require.Equal(suite.T(), expectedSubtype, response.Result[expectedLedgerId].SubType)
	require.Equal(suite.T(), expectedAclass, response.Result[expectedLedgerId].AssetClass)
	require.Equal(suite.T(), expectedAsset, response.Result[expectedLedgerId].Asset)
	require.Equal(suite.T(), expectedAmount, response.Result[expectedLedgerId].Amount.String())
	require.Equal(suite.T(), expectedFee, response.Result[expectedLedgerId].Fee.String())
	require.Equal(suite.T(), expectedBalance, response.Result[expectedLedgerId].Balance.String())
}
