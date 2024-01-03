package messages

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* UNIT TEST SUITE                                                                               */
/*************************************************************************************************/

// Unit test suite for OpenOrders
type OpenOrdersUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestOpenOrdersUnitTestSuite(t *testing.T) {
	suite.Run(t, new(OpenOrdersUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example OpenOrders message from documentation into the corresponding struct.
func (suite *OpenOrdersUnitTestSuite) TestOpenOrdersUnmarshalJson() {
	// Payload to unmarshal
	payload := `[
		[
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  }
		],
		"openOrders",
		{
		  "sequence": 234
		}
	]`
	// Expectations
	expectedChannelName := string(ChannelOpenOrders)
	expectedSeqId := int64(234)
	expectedCount := 4
	expectedOrderId := "OGTT3Y-C6I3P-XRI6HX"
	expectedVolume := "10.00345345"
	// Unmarshal payload into target struct
	target := new(OpenOrders)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedChannelName, target.ChannelName)
	require.Equal(suite.T(), expectedSeqId, target.Sequence.Sequence)
	require.Len(suite.T(), target.Orders, expectedCount)
	require.Equal(suite.T(), expectedVolume, target.Orders[0][expectedOrderId].Volume)
}

// Test marshalling an example OpenOrders message to the same paylaod as documentation.
func (suite *OpenOrdersUnitTestSuite) TestOpenOrdersMarshalJson() {
	// Payload to marshal
	payload := `[
		[
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
				"refid": "OKIVMP-5GVZN-Z2D2UA",
				"userref": 0,
				"status": "open",
				"opentm": "0.000000",
				"starttm": "0.000000",
				"expiretm": "0.000000",
				"descr": {
					"pair": "XBT/EUR",
					"type": "sell",
					"ordertype": "limit",
					"price": "34.50000",
					"price2": "0.00000",
					"leverage": "0:1",
					"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage"
				  },
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "avg_price": "34.50000",
			  "stopprice": "0.000000",
			  "limitprice": "34.50000",
			  "oflags": "fcib"
			}
		  }
		],
		"openOrders",
		{
		  "sequence": 234
		}
	]`
	// Remove whitespace
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(OpenOrders)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}
