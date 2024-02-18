package messages

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* UNIT TEST SUITE                                                                               */
/*************************************************************************************************/

// Unit test suite for the regex used to extract the message type out of the messages received from
// the server
type MatchingRegexUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestMatchingRegexUnitTestSuite(t *testing.T) {
	suite.Run(t, new(MatchingRegexUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test matching a pong message
func (suite *MatchingRegexUnitTestSuite) TestMatchPong() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"event": "pong",
		"reqid": 42
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	// 3 matches are expected in case of success:
	// - 0 is the original message
	// - 1 is the event type in case message is a JSON object with a event field
	// - 2 is the channel name in case message is a JSON array (publication)
	//
	// One of 1 or 2 will be empty. Users will have to check both to find the event type
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "pong", matches[1])
}

// Test matching a heartbeat message
func (suite *MatchingRegexUnitTestSuite) TestMatchHeartbeat() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"event": "heartbeat"
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "heartbeat", matches[1])
}

// Test matching a systemStatus message
func (suite *MatchingRegexUnitTestSuite) TestMatchSystemStatus() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"connectionID": 8628615390848610000,
		"event": "systemStatus",
		"status": "online",
		"version": "1.0.0"
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "systemStatus", matches[1])
}

// Test matching a subscriptionStatus message
func (suite *MatchingRegexUnitTestSuite) TestMatchSubscriptionStatus() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"channelID": 10001,
		"channelName": "ohlc-5",
		"event": "subscriptionStatus",
		"pair": "XBT/EUR",
		"reqid": 42,
		"status": "unsubscribed",
		"subscription": {
		  "interval": 5,
		  "name": "ohlc"
		}
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "subscriptionStatus", matches[1])
}

// Test matching a ticker message
func (suite *MatchingRegexUnitTestSuite) TestMatchTicker() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		0,
		{
		  "a": [
			"5525.40000",
			1,
			"1.000"
		  ],
		  "b": [
			"5525.10000",
			1,
			"1.000"
		  ],
		  "c": [
			"5525.10000",
			"0.00398963"
		  ],
		  "h": [
			"5783.00000",
			"5783.00000"
		  ],
		  "l": [
			"5505.00000",
			"5505.00000"
		  ],
		  "o": [
			"5760.70000",
			"5763.40000"
		  ],
		  "p": [
			"5631.44067",
			"5653.78939"
		  ],
		  "t": [
			11493,
			16267
		  ],
		  "v": [
			"2634.11501494",
			"3591.17907851"
		  ]
		},
		"ticker",
		"XBT/USD"
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "ticker", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a ohlc message
func (suite *MatchingRegexUnitTestSuite) TestMatchOHLC() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		42,
		[
		  "1542057314.748456",
		  "1542057360.435743",
		  "3586.70000",
		  "3586.70000",
		  "3586.60000",
		  "3586.60000",
		  "3586.68894",
		  "0.03373000",
		  2
		],
		"ohlc-5",
		"XBT/USD"
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "ohlc-5", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a trade message
func (suite *MatchingRegexUnitTestSuite) TestMatchTrade() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		0,
		[
		  [
			"5541.20000",
			"0.15850568",
			"1534614057.321597",
			"s",
			"l",
			""
		  ],
		  [
			"6060.00000",
			"0.02455000",
			"1534614057.324998",
			"b",
			"l",
			""
		  ]
		],
		"trade",
		"XBT/USD"
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "trade", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a spread message
func (suite *MatchingRegexUnitTestSuite) TestMatchSpread() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		0,
		[
		  "5698.40000",
		  "5700.00000",
		  "1542057299.545897",
		  "1.01234567",
		  "0.98765432"
		],
		"spread",
		"XBT/USD"
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "spread", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a book snapshot message
func (suite *MatchingRegexUnitTestSuite) TestMatchBookSnapshot() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		0,
		{
		  "as": [
			[
			  "5541.30000",
			  "2.50700000",
			  "1534614248.123678"
			],
			[
			  "5541.80000",
			  "0.33000000",
			  "1534614098.345543"
			],
			[
			  "5542.70000",
			  "0.64700000",
			  "1534614244.654432"
			]
		  ],
		  "bs": [
			[
			  "5541.20000",
			  "1.52900000",
			  "1534614248.765567"
			],
			[
			  "5539.90000",
			  "0.30000000",
			  "1534614241.769870"
			],
			[
			  "5539.50000",
			  "5.00000000",
			  "1534613831.243486"
			]
		  ]
		},
		"book-100",
		"XBT/USD"
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "book-100", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a book update message
func (suite *MatchingRegexUnitTestSuite) TestMatchBookUpdate() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		1234,
		{
		  "a": [
			[
			  "5541.30000",
			  "2.50700000",
			  "1534614248.456738"
			],
			[
			  "5542.50000",
			  "0.40100000",
			  "1534614248.456738"
			]
		  ],
		  "c": "974942666"
		},
		"book-10",
		"XBT/USD"
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "book-10", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a ownTrades message
func (suite *MatchingRegexUnitTestSuite) TestMatchOwnTrades() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		[
		  {
			"TDLH43-DVQXD-2KHVYY": {
			  "cost": "1000000.00000",
			  "fee": "1600.00000",
			  "margin": "0.00000",
			  "ordertxid": "TDLH43-DVQXD-2KHVYY",
			  "ordertype": "limit",
			  "pair": "XBT/EUR",
			  "postxid": "OGTT3Y-C6I3P-XRI6HX",
			  "price": "100000.00000",
			  "time": "1560516023.070651",
			  "type": "sell",
			  "vol": "1000000000.00000000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
			  "cost": "1000000.00000",
			  "fee": "600.00000",
			  "margin": "0.00000",
			  "ordertxid": "TDLH43-DVQXD-2KHVYY",
			  "ordertype": "limit",
			  "pair": "XBT/EUR",
			  "postxid": "OGTT3Y-C6I3P-XRI6HX",
			  "price": "100000.00000",
			  "time": "1560516023.070658",
			  "type": "buy",
			  "vol": "1000000000.00000000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
			  "cost": "1000000.00000",
			  "fee": "1600.00000",
			  "margin": "0.00000",
			  "ordertxid": "TDLH43-DVQXD-2KHVYY",
			  "ordertype": "limit",
			  "pair": "XBT/EUR",
			  "postxid": "OGTT3Y-C6I3P-XRI6HX",
			  "price": "100000.00000",
			  "time": "1560520332.914657",
			  "type": "sell",
			  "vol": "1000000000.00000000"
			}
		  },
		  {
			"TDLH43-DVQXD-2KHVYY": {
			  "cost": "1000000.00000",
			  "fee": "600.00000",
			  "margin": "0.00000",
			  "ordertxid": "TDLH43-DVQXD-2KHVYY",
			  "ordertype": "limit",
			  "pair": "XBT/EUR",
			  "postxid": "OGTT3Y-C6I3P-XRI6HX",
			  "price": "100000.00000",
			  "time": "1560520332.914664",
			  "type": "buy",
			  "vol": "1000000000.00000000"
			}
		  }
		],
		"ownTrades",
		{
		  "sequence": 2948
		}
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "ownTrades", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a openOrders message
func (suite *MatchingRegexUnitTestSuite) TestMatchOpenOrders() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`[
		[
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
			  "avg_price": "34.50000",
			  "cost": "0.00000",
			  "descr": {
				"close": "",
				"leverage": "0:1",
				"order": "sell 10.00345345 XBT/EUR @ limit 34.50000 with 0:1 leverage",
				"ordertype": "limit",
				"pair": "XBT/EUR",
				"price": "34.50000",
				"price2": "0.00000",
				"type": "sell"
			  },
			  "expiretm": "0.000000",
			  "fee": "0.00000",
			  "limitprice": "34.50000",
			  "misc": "",
			  "oflags": "fcib",
			  "opentm": "0.000000",
			  "refid": "OKIVMP-5GVZN-Z2D2UA",
			  "starttm": "0.000000",
			  "status": "open",
			  "stopprice": "0.000000",
			  "userref": 0,
			  "vol": "10.00345345",
			  "vol_exec": "0.00000000"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
			  "avg_price": "5334.60000",
			  "cost": "0.00000",
			  "descr": {
				"close": "",
				"leverage": "0:1",
				"order": "sell 0.00000010 XBT/EUR @ limit 5334.60000 with 0:1 leverage",
				"ordertype": "limit",
				"pair": "XBT/EUR",
				"price": "5334.60000",
				"price2": "0.00000",
				"type": "sell"
			  },
			  "expiretm": "0.000000",
			  "fee": "0.00000",
			  "limitprice": "5334.60000",
			  "misc": "",
			  "oflags": "fcib",
			  "opentm": "0.000000",
			  "refid": "OKIVMP-5GVZN-Z2D2UA",
			  "starttm": "0.000000",
			  "status": "open",
			  "stopprice": "0.000000",
			  "userref": 0,
			  "vol": "0.00000010",
			  "vol_exec": "0.00000000"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
			  "avg_price": "90.40000",
			  "cost": "0.00000",
			  "descr": {
				"close": "",
				"leverage": "0:1",
				"order": "sell 0.00001000 XBT/EUR @ limit 90.40000 with 0:1 leverage",
				"ordertype": "limit",
				"pair": "XBT/EUR",
				"price": "90.40000",
				"price2": "0.00000",
				"type": "sell"
			  },
			  "expiretm": "0.000000",
			  "fee": "0.00000",
			  "limitprice": "90.40000",
			  "misc": "",
			  "oflags": "fcib",
			  "opentm": "0.000000",
			  "refid": "OKIVMP-5GVZN-Z2D2UA",
			  "starttm": "0.000000",
			  "status": "open",
			  "stopprice": "0.000000",
			  "userref": 0,
			  "vol": "0.00001000",
			  "vol_exec": "0.00000000"
			}
		  },
		  {
			"OGTT3Y-C6I3P-XRI6HX": {
			  "avg_price": "9.00000",
			  "cost": "0.00000",
			  "descr": {
				"close": "",
				"leverage": "0:1",
				"order": "sell 0.00001000 XBT/EUR @ limit 9.00000 with 0:1 leverage",
				"ordertype": "limit",
				"pair": "XBT/EUR",
				"price": "9.00000",
				"price2": "0.00000",
				"type": "sell"
			  },
			  "expiretm": "0.000000",
			  "fee": "0.00000",
			  "limitprice": "9.00000",
			  "misc": "",
			  "oflags": "fcib",
			  "opentm": "0.000000",
			  "refid": "OKIVMP-5GVZN-Z2D2UA",
			  "starttm": "0.000000",
			  "status": "open",
			  "stopprice": "0.000000",
			  "userref": 0,
			  "vol": "0.00001000",
			  "vol_exec": "0.00000000"
			}
		  }
		],
		"openOrders",
		{
		  "sequence": 234
		}
	]`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "openOrders", matches[2]) // 3rd item is the match in case of json array
}

// Test matching a addOrderStatus message
func (suite *MatchingRegexUnitTestSuite) TestMatchAddOrderStatus() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"descr": "buy 0.01770000 XBTUSD @ limit 4000",
		"event": "addOrderStatus",
		"status": "ok",
		"txid": "ONPNXH-KMKMU-F4MR5V"
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "addOrderStatus", matches[1])
}

// Test matching a editOrderStatus message
func (suite *MatchingRegexUnitTestSuite) TestMatchEditOrderStatus() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"descr": "order edited price = 9000.00000000",
		"event": "editOrderStatus",
		"originaltxid": "O65KZW-J4AW3-VFS74A",
		"reqid": 3,
		"status": "ok",
		"txid": "OTI672-HJFAO-XOIPPK"
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "editOrderStatus", matches[1])
}

// Test matching a cancelOrderStatus message
func (suite *MatchingRegexUnitTestSuite) TestMatchCancelOrderStatus() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"errorMessage": "EOrder:Unknown order",
		"event": "cancelOrderStatus",
		"status": "error"
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "cancelOrderStatus", matches[1])
}

// Test matching a cancelAllStatus message
func (suite *MatchingRegexUnitTestSuite) TestMatchCancelAllStatus() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"count": 2,
		"event": "cancelAllStatus",
		"status": "ok"
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "cancelAllStatus", matches[1])
}

// Test matching a cancelAllOrdersAfterStatus message
func (suite *MatchingRegexUnitTestSuite) TestMatchCancelAllOrdersAfterStatus() {
	// Payload to match
	payload := matchesWhitespacesRegex.ReplaceAllString(`{
		"currentTime": "2020-12-21T09:37:09Z",
		"event": "cancelAllOrdersAfterStatus",
		"reqid": 1608543428051,
		"status": "ok",
		"triggerTime": "0"
	}`, "")
	matches := MatchMessageTypeRegex.FindStringSubmatch(payload)
	require.Len(suite.T(), matches, 3)
	require.Equal(suite.T(), "cancelAllOrdersAfterStatus", matches[1])
}
