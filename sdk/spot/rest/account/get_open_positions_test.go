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

// Unit test suite for GetOpenPositions DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetOpenPositionsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetOpenPositionsTestSuite(t *testing.T) {
	suite.Run(t, new(GetOpenPositionsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetOpenPositions.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetOpenPositionsResponse struct.
func (suite *GetOpenPositionsTestSuite) TestGetOpenPositionsUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "TF5GVO-T7ZZ2-6NBKBI": {
			"ordertxid": "OLWNFG-LLH4R-D6SFFP",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1605280097.8294,
			"type": "buy",
			"ordertype": "limit",
			"cost": "104610.52842",
			"fee": "289.06565",
			"vol": "8.82412861",
			"vol_closed": "0.20200000",
			"margin": "20922.10568",
			"value": "258797.5",
			"net": "+154186.9728",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  },
		  "T24DOR-TAFLM-ID3NYP": {
			"ordertxid": "OIVYGZ-M5EHU-ZRUQXX",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1607943827.3172,
			"type": "buy",
			"ordertype": "limit",
			"cost": "145756.76856",
			"fee": "335.24057",
			"vol": "8.00000000",
			"vol_closed": "0.00000000",
			"margin": "29151.35371",
			"value": "240124.0",
			"net": "+94367.2314",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  },
		  "TYMRFG-URRG5-2ZTQSD": {
			"ordertxid": "OF5WFH-V57DP-QANDAC",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1610448039.8374,
			"type": "buy",
			"ordertype": "limit",
			"cost": "0.00240",
			"fee": "0.00000",
			"vol": "0.00000010",
			"vol_closed": "0.00000000",
			"margin": "0.00048",
			"value": "0",
			"net": "+0.0006",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  },
		  "TAFGBN-TZNFC-7CCYIM": {
			"ordertxid": "OF5WFH-V57DP-QANDAC",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1610448039.8448,
			"type": "buy",
			"ordertype": "limit",
			"cost": "2.40000",
			"fee": "0.00264",
			"vol": "0.00010000",
			"vol_closed": "0.00000000",
			"margin": "0.48000",
			"value": "3.0",
			"net": "+0.6015",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  },
		  "T4O5L3-4VGS4-IRU2UL": {
			"ordertxid": "OF5WFH-V57DP-QANDAC",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1610448040.7722,
			"type": "buy",
			"ordertype": "limit",
			"cost": "21.59760",
			"fee": "0.02376",
			"vol": "0.00089990",
			"vol_closed": "0.00000000",
			"margin": "4.31952",
			"value": "27.0",
			"net": "+5.4133",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  }
		}
	}`
	expectedCount := 5
	expectedItem5Id := "T4O5L3-4VGS4-IRU2UL"
	expectedItem5OrderTxId := "OF5WFH-V57DP-QANDAC"
	expectedItem5PosStatus := string(PositionOpen)
	expectedItem5Pair := "XXBTZUSD"
	expectedItem5Time := "1610448040.7722"
	expectedItem5Type := string(Buy)
	expectedItem5OrderType := string(Limit)
	expectedItem5Cost := "21.59760"
	expectedItem5Fee := "0.02376"
	expectedItem5Vol := "0.00089990"
	expectedItem5VolClosed := "0.00000000"
	expectedItem5Margin := "4.31952"
	expectedItem5Value := "27.0"
	expectedItem5Net := "+5.4133"
	expectedItem5Terms := "0.0100% per 4 hours"
	expectedItem5RolloverTm := "1616672637"
	expectedItem5Misc := ""
	expectedItem5OFlags := ""
	// Unmarshal payload into struct
	response := new(GetOpenPositionsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedItem5OrderTxId, response.Result[expectedItem5Id].OrderTransactionId)
	require.Equal(suite.T(), expectedItem5PosStatus, response.Result[expectedItem5Id].PositionStatus)
	require.Equal(suite.T(), expectedItem5Pair, response.Result[expectedItem5Id].Pair)
	require.Equal(suite.T(), expectedItem5Time, response.Result[expectedItem5Id].Timestamp.String())
	require.Equal(suite.T(), expectedItem5Type, response.Result[expectedItem5Id].Type)
	require.Equal(suite.T(), expectedItem5OrderType, response.Result[expectedItem5Id].OrderType)
	require.Equal(suite.T(), expectedItem5Cost, response.Result[expectedItem5Id].Cost.String())
	require.Equal(suite.T(), expectedItem5Fee, response.Result[expectedItem5Id].Fee.String())
	require.Equal(suite.T(), expectedItem5Vol, response.Result[expectedItem5Id].Volume.String())
	require.Equal(suite.T(), expectedItem5VolClosed, response.Result[expectedItem5Id].ClosedVolume.String())
	require.Equal(suite.T(), expectedItem5Margin, response.Result[expectedItem5Id].Margin.String())
	require.Equal(suite.T(), expectedItem5Value, response.Result[expectedItem5Id].Value.String())
	require.Equal(suite.T(), expectedItem5Net, response.Result[expectedItem5Id].Net)
	require.Equal(suite.T(), expectedItem5Terms, response.Result[expectedItem5Id].Terms)
	require.Equal(suite.T(), expectedItem5RolloverTm, response.Result[expectedItem5Id].RolloverTimestamp.String())
	require.Equal(suite.T(), expectedItem5Misc, response.Result[expectedItem5Id].Miscellaneous)
	require.Equal(suite.T(), expectedItem5OFlags, response.Result[expectedItem5Id].OrderFlags)
}
