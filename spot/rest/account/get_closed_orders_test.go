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

// Unit test suite for GetClosedOrders DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetClosedOrdersTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetClosedOrdersTestSuite(t *testing.T) {
	suite.Run(t, new(GetClosedOrdersTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetClosedOrders.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetClosedOrdersResponse struct.
func (suite *GetClosedOrdersTestSuite) TestGetClosedOrdersUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "closed": {
			"O37652-RJWRT-IMO74O": {
			  "refid": "None",
			  "userref": 1,
			  "status": "canceled",
			  "reason": "User requested",
			  "opentm": 1688148493.7708,
			  "closetm": 1688148610.0482,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTGBP",
				"type": "buy",
				"ordertype": "stop-loss-limit",
				"price": "23667.0",
				"price2": "0",
				"leverage": "none",
				"order": "buy 0.00100000 XBTGBP @ limit 23667.0",
				"close": ""
			  },
			  "vol": "0.00100000",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "price": "0.00000",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq",
			  "trigger": "index"
			},
			"O6YDQ5-LOMWU-37YKEE": {
			  "refid": "None",
			  "userref": 36493663,
			  "status": "canceled",
			  "reason": "User requested",
			  "opentm": 1688148493.7708,
			  "closetm": 1688148610.0477,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTEUR",
				"type": "buy",
				"ordertype": "take-profit-limit",
				"price": "27743.0",
				"price2": "0",
				"leverage": "none",
				"order": "buy 0.00100000 XBTEUR @ limit 27743.0",
				"close": ""
			  },
			  "vol": "0.00100000",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "price": "0.00000",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq",
			  "trigger": "index"
			}
		  },
		  "count": 2
		}
	}`
	expectedCount := 2
	expectedItem2TxId := "O6YDQ5-LOMWU-37YKEE"
	expectedItem2Refid := "None"
	expectedItem2Userref := "36493663"
	expectedItem2Status := string(Canceled)
	expectedItem2Reason := "User requested"
	expectedItem2OpenTm := "1688148493.7708"
	expectedItem2CloseTm := "1688148610.0477"
	expectedItem2StartTm := "0"
	expectedItem2ExpireTm := "0"
	expectedItem2DescrPair := "XBTEUR"
	expectedItem2DescrType := string(Buy)
	expectedItem2DescrOrdertype := string(TakeProfitLimit)
	expectedItem2DescrPrice := "27743.0"
	expectedItem2DescrPrice2 := "0"
	expectedItem2DescrLeverage := "none"
	expectedItem2DescrOrder := "buy 0.00100000 XBTEUR @ limit 27743.0"
	expectedItem2DescrClose := ""
	expectedItem2Volume := "0.00100000"
	expectedItem2VolumeExec := "0.00000000"
	expectedItem2Cost := "0.00000"
	expectedItem2Fee := "0.00000"
	expectedItem2Price := "0.00000"
	expectedItem2StopPrice := "0.00000"
	expectedItem2LimitPrice := "0.00000"
	expectedItem2Misc := ""
	expectedItem2OFlags := "fciq"
	expectedItem2Trigger := "index"
	// Unmarshal payload into struct
	response := new(GetClosedOrdersResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Closed, expectedCount)
	require.Equal(suite.T(), expectedCount, response.Result.Count)
	require.Equal(suite.T(), expectedItem2Refid, response.Result.Closed[expectedItem2TxId].ReferralOrderTransactionId)
	require.Equal(suite.T(), expectedItem2Userref, response.Result.Closed[expectedItem2TxId].UserReferenceId.String())
	require.Equal(suite.T(), expectedItem2Status, response.Result.Closed[expectedItem2TxId].Status)
	require.Equal(suite.T(), expectedItem2Reason, response.Result.Closed[expectedItem2TxId].Reason)
	require.Equal(suite.T(), expectedItem2OpenTm, response.Result.Closed[expectedItem2TxId].OpenTimestamp.String())
	require.Equal(suite.T(), expectedItem2CloseTm, response.Result.Closed[expectedItem2TxId].CloseTimestamp.String())
	require.Equal(suite.T(), expectedItem2StartTm, response.Result.Closed[expectedItem2TxId].StartTimestamp.String())
	require.Equal(suite.T(), expectedItem2ExpireTm, response.Result.Closed[expectedItem2TxId].ExpireTimestamp.String())
	require.Equal(suite.T(), expectedItem2DescrPair, response.Result.Closed[expectedItem2TxId].Description.Pair)
	require.Equal(suite.T(), expectedItem2DescrType, response.Result.Closed[expectedItem2TxId].Description.Type)
	require.Equal(suite.T(), expectedItem2DescrOrdertype, response.Result.Closed[expectedItem2TxId].Description.OrderType)
	require.Equal(suite.T(), expectedItem2DescrPrice, response.Result.Closed[expectedItem2TxId].Description.Price.String())
	require.Equal(suite.T(), expectedItem2DescrPrice2, response.Result.Closed[expectedItem2TxId].Description.Price2.String())
	require.Equal(suite.T(), expectedItem2DescrLeverage, response.Result.Closed[expectedItem2TxId].Description.Leverage)
	require.Equal(suite.T(), expectedItem2DescrOrder, response.Result.Closed[expectedItem2TxId].Description.OrderDescription)
	require.Equal(suite.T(), expectedItem2DescrClose, response.Result.Closed[expectedItem2TxId].Description.CloseOrderDescription)
	require.Equal(suite.T(), expectedItem2Volume, response.Result.Closed[expectedItem2TxId].Volume.String())
	require.Equal(suite.T(), expectedItem2VolumeExec, response.Result.Closed[expectedItem2TxId].VolumeExecuted.String())
	require.Equal(suite.T(), expectedItem2Cost, response.Result.Closed[expectedItem2TxId].Cost.String())
	require.Equal(suite.T(), expectedItem2Fee, response.Result.Closed[expectedItem2TxId].Fee.String())
	require.Equal(suite.T(), expectedItem2Price, response.Result.Closed[expectedItem2TxId].Price.String())
	require.Equal(suite.T(), expectedItem2StopPrice, response.Result.Closed[expectedItem2TxId].StopPrice.String())
	require.Equal(suite.T(), expectedItem2LimitPrice, response.Result.Closed[expectedItem2TxId].LimitPrice.String())
	require.Equal(suite.T(), expectedItem2Misc, response.Result.Closed[expectedItem2TxId].Miscellaneous)
	require.Equal(suite.T(), expectedItem2OFlags, response.Result.Closed[expectedItem2TxId].OrderFlags)
	require.Equal(suite.T(), expectedItem2Trigger, response.Result.Closed[expectedItem2TxId].Trigger)
}
