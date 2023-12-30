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

// Unit test suite for GetTradesHistory DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetTradesHistoryTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetTradesHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(GetTradesHistoryTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetTradesHistory.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetTradesHistoryResponse struct.
func (suite *GetTradesHistoryTestSuite) TestGetTradesHistoryUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
	      "count": 2,
		  "trades": {
			"THVRQM-33VKH-UCI7BS": {
			  "ordertxid": "OQCLML-BW3P3-BUCMWZ",
			  "postxid": "TKH2SE-M7IF5-CFI7LT",
			  "pair": "XXBTZUSD",
			  "time": 1688667796.8802,
			  "type": "buy",
			  "ordertype": "limit",
			  "price": "30010.00000",
			  "cost": "600.20000",
			  "fee": "0.00000",
			  "vol": "0.02000000",
			  "margin": "0.00000",
			  "misc": ""
			},
			"TCWJEG-FL4SZ-3FKGH6": {
			  "ordertxid": "OQCLML-BW3P3-BUCMWZ",
			  "postxid": "TKH2SE-M7IF5-CFI7LT",
			  "pair": "XXBTZUSD",
			  "time": 1688667769.6396,
			  "type": "buy",
			  "ordertype": "limit",
			  "price": "30010.00000",
			  "cost": "300.10000",
			  "fee": "0.00000",
			  "vol": "0.01000000",
			  "margin": "0.00000",
			  "misc": ""
			}
		  }
		}
	  }`
	expectedCount := 2
	expectedTrade2Id := "TCWJEG-FL4SZ-3FKGH6"
	expectedTrade2OrderTxId := "OQCLML-BW3P3-BUCMWZ"
	expectedTrade2PostxId := "TKH2SE-M7IF5-CFI7LT"
	expectedTrade2Pair := "XXBTZUSD"
	expectedTrade2Time := "1688667769.6396"
	expectedTrade2Type := string(Buy)
	expectedTrade2OrderType := string(Limit)
	expectedTrade2Price := "30010.00000"
	expectedTrade2Cost := "300.10000"
	expectedTrade2Fee := "0.00000"
	expectedTrade2Volume := "0.01000000"
	expectedTrade2Margin := "0.00000"
	expectedTrade2Misc := ""
	// Unmarshal payload into struct
	response := new(GetTradesHistoryResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Trades, expectedCount)
	require.Equal(suite.T(), expectedCount, response.Result.Count)
	require.Equal(suite.T(), expectedTrade2OrderTxId, response.Result.Trades[expectedTrade2Id].OrderTransactionId)
	require.Equal(suite.T(), expectedTrade2PostxId, response.Result.Trades[expectedTrade2Id].PositionId)
	require.Equal(suite.T(), expectedTrade2Pair, response.Result.Trades[expectedTrade2Id].Pair)
	require.Equal(suite.T(), expectedTrade2Time, response.Result.Trades[expectedTrade2Id].Timestamp.String())
	require.Equal(suite.T(), expectedTrade2Type, response.Result.Trades[expectedTrade2Id].Type)
	require.Equal(suite.T(), expectedTrade2OrderType, response.Result.Trades[expectedTrade2Id].OrderType)
	require.Equal(suite.T(), expectedTrade2Price, response.Result.Trades[expectedTrade2Id].Price.String())
	require.Equal(suite.T(), expectedTrade2Cost, response.Result.Trades[expectedTrade2Id].Cost.String())
	require.Equal(suite.T(), expectedTrade2Fee, response.Result.Trades[expectedTrade2Id].Fee.String())
	require.Equal(suite.T(), expectedTrade2Volume, response.Result.Trades[expectedTrade2Id].Volume.String())
	require.Equal(suite.T(), expectedTrade2Margin, response.Result.Trades[expectedTrade2Id].Margin.String())
	require.Equal(suite.T(), expectedTrade2Misc, response.Result.Trades[expectedTrade2Id].Miscellaneous)
}
