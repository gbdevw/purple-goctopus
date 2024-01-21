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

// Unit test suite for QueryOrdersInfo DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type QueryOrdersInfoTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestQueryOrdersInfoTestSuite(t *testing.T) {
	suite.Run(t, new(QueryOrdersInfoTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of QueryOrdersInfo.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding QueryOrdersInfoResponse struct.
func (suite *QueryOrdersInfoTestSuite) TestQueryOrdersInfoUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "OBCMZD-JIEE7-77TH3F": {
			"refid": "None",
			"userref": 0,
			"status": "closed",
			"reason": null,
			"opentm": 1688665496.7808,
			"closetm": 1688665499.1922,
			"starttm": 0,
			"expiretm": 0,
			"descr": {
			  "pair": "XBTUSD",
			  "type": "buy",
			  "ordertype": "stop-loss-limit",
			  "price": "27500.0",
			  "price2": "0",
			  "leverage": "none",
			  "order": "buy 1.25000000 XBTUSD @ limit 27500.0",
			  "close": ""
			},
			"vol": "1.25000000",
			"vol_exec": "1.25000000",
			"cost": "27526.2",
			"fee": "26.2",
			"price": "27500.0",
			"stopprice": "0.00000",
			"limitprice": "0.00000",
			"misc": "",
			"oflags": "fciq",
			"trigger": "index",
			"trades": [
			  "TZX2WP-XSEOP-FP7WYR"
			]
		  },
		  "OMMDB2-FSB6Z-7W3HPO": {
			"refid": "None",
			"userref": 0,
			"status": "closed",
			"reason": null,
			"opentm": 1688592012.2317,
			"closetm": 1688592012.2335,
			"starttm": 0,
			"expiretm": 0,
			"descr": {
			  "pair": "XBTUSD",
			  "type": "sell",
			  "ordertype": "market",
			  "price": "0",
			  "price2": "0",
			  "leverage": "none",
			  "order": "sell 0.25000000 XBTUSD @ market",
			  "close": ""
			},
			"vol": "0.25000000",
			"vol_exec": "0.25000000",
			"cost": "7500.0",
			"fee": "7.5",
			"price": "30000.0",
			"stopprice": "0.00000",
			"limitprice": "0.00000",
			"misc": "",
			"oflags": "fcib",
			"trades": [
			  "TJUW2K-FLX2N-AR2FLU"
			]
		  }
		}
	}`
	expectedCount := 2
	expectedOrderId := "OMMDB2-FSB6Z-7W3HPO"
	expectedOrder1UsrRef := "0"
	expectedOrder2RefId := "None"
	expectedOrder2Status := string(Closed)
	expectedOrder2Reason := ""
	expectedOrder2OpenTm := "1688592012.2317"
	expectedOrder2CloseTm := "1688592012.2335"
	expectedOrder2StartTm := "0"
	expectedOrder2SExpireTm := "0"
	expectedOrder2DescrPair := "XBTUSD"
	expectedOrder2DescrType := string(Sell)
	expectedOrder2DescrOrderType := string(Market)
	expectedOrder2DescrPrice := "0"
	expectedOrder2DescrPrice2 := "0"
	expectedOrder2DescrLeverage := "none"
	expectedOrder2DescrOrder := "sell 0.25000000 XBTUSD @ market"
	expectedOrder2DescrClose := ""
	expectedOrder2Vol := "0.25000000"
	expectedOrder2ExecVol := "0.25000000"
	expectedOrder2Cost := "7500.0"
	expectedOrder2Fee := "7.5"
	expectedOrder2Price := "30000.0"
	expectedOrder2StopPrice := "0.00000"
	expectedOrder2LimitPrice := "0.00000"
	expectedOrder2Misc := ""
	expectedOrder2OFlags := "fcib"
	expectedOrder2Trades := []string{"TJUW2K-FLX2N-AR2FLU"}
	// Unmarshal payload into struct
	response := new(QueryOrdersInfoResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedOrder1UsrRef, response.Result[expectedOrderId].UserReferenceId.String())
	require.Equal(suite.T(), expectedOrder2RefId, response.Result[expectedOrderId].ReferralOrderTransactionId)
	require.Equal(suite.T(), expectedOrder2Status, response.Result[expectedOrderId].Status)
	require.Equal(suite.T(), expectedOrder2Reason, response.Result[expectedOrderId].Reason)
	require.Equal(suite.T(), expectedOrder2OpenTm, response.Result[expectedOrderId].OpenTimestamp.String())
	require.Equal(suite.T(), expectedOrder2CloseTm, response.Result[expectedOrderId].CloseTimestamp.String())
	require.Equal(suite.T(), expectedOrder2StartTm, response.Result[expectedOrderId].StartTimestamp.String())
	require.Equal(suite.T(), expectedOrder2SExpireTm, response.Result[expectedOrderId].ExpireTimestamp.String())
	require.Equal(suite.T(), expectedOrder2DescrPair, response.Result[expectedOrderId].Description.Pair)
	require.Equal(suite.T(), expectedOrder2DescrType, response.Result[expectedOrderId].Description.Type)
	require.Equal(suite.T(), expectedOrder2DescrOrderType, response.Result[expectedOrderId].Description.OrderType)
	require.Equal(suite.T(), expectedOrder2DescrPrice, response.Result[expectedOrderId].Description.Price.String())
	require.Equal(suite.T(), expectedOrder2DescrPrice2, response.Result[expectedOrderId].Description.Price2.String())
	require.Equal(suite.T(), expectedOrder2DescrLeverage, response.Result[expectedOrderId].Description.Leverage)
	require.Equal(suite.T(), expectedOrder2DescrOrder, response.Result[expectedOrderId].Description.OrderDescription)
	require.Equal(suite.T(), expectedOrder2DescrClose, response.Result[expectedOrderId].Description.CloseOrderDescription)
	require.Equal(suite.T(), expectedOrder2Vol, response.Result[expectedOrderId].Volume.String())
	require.Equal(suite.T(), expectedOrder2ExecVol, response.Result[expectedOrderId].VolumeExecuted.String())
	require.Equal(suite.T(), expectedOrder2Cost, response.Result[expectedOrderId].Cost.String())
	require.Equal(suite.T(), expectedOrder2Fee, response.Result[expectedOrderId].Fee.String())
	require.Equal(suite.T(), expectedOrder2Price, response.Result[expectedOrderId].Price.String())
	require.Equal(suite.T(), expectedOrder2StopPrice, response.Result[expectedOrderId].StopPrice.String())
	require.Equal(suite.T(), expectedOrder2LimitPrice, response.Result[expectedOrderId].LimitPrice.String())
	require.Equal(suite.T(), expectedOrder2Misc, response.Result[expectedOrderId].Miscellaneous)
	require.Equal(suite.T(), expectedOrder2OFlags, response.Result[expectedOrderId].OrderFlags)
	require.ElementsMatch(suite.T(), expectedOrder2Trades, response.Result[expectedOrderId].Trades)
}
