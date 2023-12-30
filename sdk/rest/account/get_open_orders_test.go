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

// Unit test suite for GetOpenOrders DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetOpenOrdersTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetOpenOrdersTestSuite(t *testing.T) {
	suite.Run(t, new(GetOpenOrdersTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetOpenOrders.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetOpenOrdersResponse struct.
func (suite *GetOpenOrdersTestSuite) TestGetOpenOrdersUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "open": {
			"OQCLML-BW3P3-BUCMWZ": {
			  "refid": "None",
			  "userref": 0,
			  "status": "open",
			  "opentm": 1688666559.8974,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTUSD",
				"type": "buy",
				"ordertype": "limit",
				"price": "30010.0",
				"price2": "0",
				"leverage": "none",
				"order": "buy 1.25000000 XBTUSD @ limit 30010.0",
				"close": ""
			  },
			  "vol": "1.25000000",
			  "vol_exec": "0.37500000",
			  "cost": "11253.7",
			  "fee": "0.00000",
			  "price": "30010.0",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq",
			  "trades": [
				"TCCCTY-WE2O6-P3NB37"
			  ]
			},
			"OB5VMB-B4U2U-DK2WRW": {
			  "refid": "None",
			  "userref": 45326,
			  "status": "open",
			  "opentm": 1688665899.5699,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTUSD",
				"type": "buy",
				"ordertype": "limit",
				"price": "14500.0",
				"price2": "0",
				"leverage": "5:1",
				"order": "buy 0.27500000 XBTUSD @ limit 14500.0 with 5:1 leverage",
				"close": ""
			  },
			  "vol": "0.27500000",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "price": "0.00000",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq"
			}
		  }
		}
	}`
	expectedCount := 2
	expectedOrder1UsrRef := "0"
	expectedOrder2RefId := "None"
	expectedOrder2RUsrRef := "45326"
	expectedOrder2Status := string(Open)
	expectedOrder2OpenTm := "1688665899.5699"
	expectedOrder2StartTm := "0"
	expectedOrder2SExpireTm := "0"
	expectedOrder2DescrPair := "XBTUSD"
	expectedOrder2DescrType := string(Buy)
	expectedOrder2DescrOrderType := string(Limit)
	expectedOrder2DescrPrice := "14500.0"
	expectedOrder2DescrPrice2 := "0"
	expectedOrder2DescrLeverage := "5:1"
	expectedOrder2DescrOrder := "buy 0.27500000 XBTUSD @ limit 14500.0 with 5:1 leverage"
	expectedOrder2DescrClose := ""
	expectedOrder2Vol := "0.27500000"
	expectedOrder2ExecVol := "0.00000000"
	expectedOrder2Cost := "0.00000"
	expectedOrder2Fee := "0.00000"
	expectedOrder2Price := "0.00000"
	expectedOrder2StopPrice := "0.00000"
	expectedOrder2LimitPrice := "0.00000"
	expectedOrder2Misc := ""
	expectedOrder2OFlags := "fciq"
	// Unmarshal payload into struct
	response := new(GetOpenOrdersResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Open, expectedCount)
	require.Equal(suite.T(), expectedOrder1UsrRef, response.Result.Open["OQCLML-BW3P3-BUCMWZ"].UserReferenceId.String())
	require.Equal(suite.T(), expectedOrder2RefId, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].ReferralOrderTransactionId)
	require.Equal(suite.T(), expectedOrder2RUsrRef, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].UserReferenceId.String())
	require.Equal(suite.T(), expectedOrder2Status, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Status)
	require.Equal(suite.T(), expectedOrder2OpenTm, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].OpenTimestamp.String())
	require.Equal(suite.T(), expectedOrder2StartTm, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].StartTimestamp.String())
	require.Equal(suite.T(), expectedOrder2SExpireTm, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].ExpireTimestamp.String())
	require.Equal(suite.T(), expectedOrder2DescrPair, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.Pair)
	require.Equal(suite.T(), expectedOrder2DescrType, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.Type)
	require.Equal(suite.T(), expectedOrder2DescrOrderType, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.OrderType)
	require.Equal(suite.T(), expectedOrder2DescrPrice, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.Price.String())
	require.Equal(suite.T(), expectedOrder2DescrPrice2, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.Price2.String())
	require.Equal(suite.T(), expectedOrder2DescrLeverage, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.Leverage)
	require.Equal(suite.T(), expectedOrder2DescrOrder, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.OrderDescription)
	require.Equal(suite.T(), expectedOrder2DescrClose, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Description.CloseOrderDescription)
	require.Equal(suite.T(), expectedOrder2Vol, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Volume.String())
	require.Equal(suite.T(), expectedOrder2ExecVol, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].VolumeExecuted.String())
	require.Equal(suite.T(), expectedOrder2Cost, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Cost.String())
	require.Equal(suite.T(), expectedOrder2Fee, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Fee.String())
	require.Equal(suite.T(), expectedOrder2Price, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Price.String())
	require.Equal(suite.T(), expectedOrder2StopPrice, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].StopPrice.String())
	require.Equal(suite.T(), expectedOrder2LimitPrice, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].LimitPrice.String())
	require.Equal(suite.T(), expectedOrder2Misc, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Miscellaneous)
	require.Equal(suite.T(), expectedOrder2OFlags, response.Result.Open["OB5VMB-B4U2U-DK2WRW"].OrderFlags)
	require.Empty(suite.T(), response.Result.Open["OB5VMB-B4U2U-DK2WRW"].Reason)
	require.Empty(suite.T(), response.Result.Open["OB5VMB-B4U2U-DK2WRW"].CloseTimestamp)
}
