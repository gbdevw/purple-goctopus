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

// Unit test suite for QueryTradesInfo DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type QueryTradesInfoTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestQueryTradesInfoTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTradesInfoTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of QueryTradesInfo.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding QueryTradesInfoResponse struct.
func (suite *QueryTradesInfoTestSuite) TestQueryTradesInfoUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
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
		  "TTEUX3-HDAAA-RC2RUO": {
			"ordertxid": "OH76VO-UKWAD-PSBDX6",
			"postxid": "TKH2SE-M7IF5-CFI7LT",
			"pair": "XXBTZEUR",
			"time": 1688082549.3138,
			"type": "buy",
			"ordertype": "limit",
			"price": "27732.00000",
			"cost": "0.20020",
			"fee": "0.00000",
			"vol": "0.00020000",
			"margin": "0.00000",
			"misc": ""
		  }
		}
	}`
	expectedCount := 2
	expectedTrade2Id := "THVRQM-33VKH-UCI7BS"
	expectedTrade2OrderTxId := "OQCLML-BW3P3-BUCMWZ"
	expectedTrade2PostxId := "TKH2SE-M7IF5-CFI7LT"
	expectedTrade2Pair := "XXBTZUSD"
	expectedTrade2Time := "1688667796.8802"
	expectedTrade2Type := string(Buy)
	expectedTrade2OrderType := string(Limit)
	expectedTrade2Price := "30010.00000"
	expectedTrade2Cost := "600.20000"
	expectedTrade2Fee := "0.00000"
	expectedTrade2Volume := "0.02000000"
	expectedTrade2Margin := "0.00000"
	expectedTrade2Misc := ""
	// Unmarshal payload into struct
	response := new(QueryTradesInfoResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedTrade2OrderTxId, response.Result[expectedTrade2Id].OrderTransactionId)
	require.Equal(suite.T(), expectedTrade2PostxId, response.Result[expectedTrade2Id].PositionId)
	require.Equal(suite.T(), expectedTrade2Pair, response.Result[expectedTrade2Id].Pair)
	require.Equal(suite.T(), expectedTrade2Time, response.Result[expectedTrade2Id].Timestamp.String())
	require.Equal(suite.T(), expectedTrade2Type, response.Result[expectedTrade2Id].Type)
	require.Equal(suite.T(), expectedTrade2OrderType, response.Result[expectedTrade2Id].OrderType)
	require.Equal(suite.T(), expectedTrade2Price, response.Result[expectedTrade2Id].Price.String())
	require.Equal(suite.T(), expectedTrade2Cost, response.Result[expectedTrade2Id].Cost.String())
	require.Equal(suite.T(), expectedTrade2Fee, response.Result[expectedTrade2Id].Fee.String())
	require.Equal(suite.T(), expectedTrade2Volume, response.Result[expectedTrade2Id].Volume.String())
	require.Equal(suite.T(), expectedTrade2Margin, response.Result[expectedTrade2Id].Margin.String())
	require.Equal(suite.T(), expectedTrade2Misc, response.Result[expectedTrade2Id].Miscellaneous)
}
