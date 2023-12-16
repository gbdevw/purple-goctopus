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

// Unit test suite for GetLedgersInfo DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetLedgersInfoTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetLedgersInfoTestSuite(t *testing.T) {
	suite.Run(t, new(GetLedgersInfoTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetLedgersInfo.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetLedgersInfoResponse struct.
func (suite *GetLedgersInfoTestSuite) TestGetLedgersInfoUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "ledger": {
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
			},
			"LMKZCZ-Z3GVL-CXKK4H": {
			  "refid": "TBZIP2-F6QOU-TMB6FY",
			  "time": 1688444262.8888,
			  "type": "trade",
			  "subtype": "",
			  "aclass": "currency",
			  "asset": "ZUSD",
			  "amount": "0.9852",
			  "fee": "0.0010",
			  "balance": "52732.1132"
			}
		  },
		  "count": 2
		}
	}`
	expectedCount := 2
	expectedItem1LedgerId := "LMKZCZ-Z3GVL-CXKK4H"
	expectedItem1RefId := "TBZIP2-F6QOU-TMB6FY"
	expectedItem1Time := "1688444262.8888"
	expectedItem1Type := string(LedgerTrade)
	expectedItem1Subtype := ""
	expectedItem1Aclass := "currency"
	expectedItem1Asset := "ZUSD"
	expectedItem1Amount := "0.9852"
	expectedItem1Fee := "0.0010"
	expectedItem1Balance := "52732.1132"
	// Unmarshal payload into struct
	response := new(GetLedgersInfoResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Ledgers, expectedCount)
	require.Equal(suite.T(), expectedCount, response.Result.Count)
	require.Equal(suite.T(), expectedItem1RefId, response.Result.Ledgers[expectedItem1LedgerId].ReferenceId)
	require.Equal(suite.T(), expectedItem1Time, response.Result.Ledgers[expectedItem1LedgerId].Timestamp.String())
	require.Equal(suite.T(), expectedItem1Type, response.Result.Ledgers[expectedItem1LedgerId].Type)
	require.Equal(suite.T(), expectedItem1Subtype, response.Result.Ledgers[expectedItem1LedgerId].SubType)
	require.Equal(suite.T(), expectedItem1Aclass, response.Result.Ledgers[expectedItem1LedgerId].AssetClass)
	require.Equal(suite.T(), expectedItem1Asset, response.Result.Ledgers[expectedItem1LedgerId].Asset)
	require.Equal(suite.T(), expectedItem1Amount, response.Result.Ledgers[expectedItem1LedgerId].Amount.String())
	require.Equal(suite.T(), expectedItem1Fee, response.Result.Ledgers[expectedItem1LedgerId].Fee.String())
	require.Equal(suite.T(), expectedItem1Balance, response.Result.Ledgers[expectedItem1LedgerId].Balance.String())
}
