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

// Unit test suite for MarketData
type MarketDataUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestMarketDataUnitTestSuite(t *testing.T) {
	suite.Run(t, new(MarketDataUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example Ticker message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a Ticker message
//   - Market data can be converted to a Ticker
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonTicker() {
	// Payload to unmarshal
	payload := `[
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
		  "v": [
			"2634.11501494",
			"3591.17907851"
		  ],
		  "p": [
			"5631.44067",
			"5653.78939"
		  ],
		  "t": [
			11493,
			16267
		  ],
		  "l": [
			"5505.00000",
			"5505.00000"
		  ],
		  "h": [
			"5783.00000",
			"5783.00000"
		  ],
		  "o": [
			"5760.70000",
			"5763.40000"
		  ]
		},
		"ticker",
		"XBT/USD"
	]`
	// Expectations
	expectedChannelId := 0
	expectedPair := "XBT/USD"
	expectedOpenToday := "5760.70000"
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to Ticker
	ticker, err := target.AsTicker()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, ticker.Pair)
	require.Equal(suite.T(), expectedChannelId, ticker.ChannelId)
	require.Equal(suite.T(), expectedOpenToday, ticker.Data.GetTodayOpen().String())
}

// Test marshalling a Ticker to the same payload as the API.
//
// Payloads are different: They have the exact same structure but
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonTicker() {
	// Payload to unmarshal
	payload := `[
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
		  "v": [
			"2634.11501494",
			"3591.17907851"
		  ],
		  "p": [
			"5631.44067",
			"5653.78939"
		  ],
		  "t": [
			11493,
			16267
		  ],
		  "l": [
			"5505.00000",
			"5505.00000"
		  ],
		  "h": [
			"5783.00000",
			"5783.00000"
		  ],
		  "o": [
			"5760.70000",
			"5763.40000"
		  ]
		},
		"ticker",
		"XBT/USD"
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to Ticker
	ticker, err := target.AsTicker()
	require.NoError(suite.T(), err)
	// Marshal Ticker
	actual, err := json.Marshal(ticker)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example OHLC message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a OHLC message
//   - Market data can be converted to a OHLC
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonOHLC() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Expectations
	expectedChannelId := 42
	expectedPair := "XBT/USD"
	expectedOpen := "3586.70000"
	expectedCount := int64(2)
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to OHLC
	ohlc, err := target.AsOHLC()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, ohlc.Pair)
	require.Equal(suite.T(), expectedChannelId, ohlc.ChannelId)
	require.Equal(suite.T(), expectedOpen, ohlc.Data.Open.String())
	require.Equal(suite.T(), expectedCount, ohlc.Data.TradesCount)
}

// Test marshalling a OHLC to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonOHLC() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to OHLC
	ohlc, err := target.AsOHLC()
	require.NoError(suite.T(), err)
	// Narshal OHLC
	actual, err := json.Marshal(ohlc)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example Trade message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a Trade message
//   - Market data can be converted to a Trade
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonTrade() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Expectations
	expectedChannelId := 0
	expectedPair := "XBT/USD"
	expectedTrade0Price := "5541.20000"
	expectedCount := 2
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to Trade
	trade, err := target.AsTrade()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, trade.Pair)
	require.Equal(suite.T(), expectedChannelId, trade.ChannelId)
	require.Len(suite.T(), trade.Data, expectedCount)
	require.Equal(suite.T(), expectedTrade0Price, trade.Data[0].Price.String())
}

// Test marshalling a Trade to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonTrade() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to Trade
	trade, err := target.AsTrade()
	require.NoError(suite.T(), err)
	// Marshal Trade
	actual, err := json.Marshal(trade)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example Spread message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a Spread message
//   - Market data can be converted to a Spread
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonSpread() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Expectations
	expectedChannelId := 0
	expectedPair := "XBT/USD"
	expectedBestBidPrice := "5698.40000"
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to Spread
	spread, err := target.AsSpread()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, spread.Pair)
	require.Equal(suite.T(), expectedChannelId, spread.ChannelId)
	require.Equal(suite.T(), expectedBestBidPrice, spread.Data.BestBidPrice.String())
}

// Test marshalling a Spread to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonSpread() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to Spread
	spread, err := target.AsSpread()
	require.NoError(suite.T(), err)
	// Marshal Spread
	actual, err := json.Marshal(spread)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookSnapshot message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonBookSnapshot() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Expectations
	expectedChannelId := 0
	expectedPair := "XBT/USD"
	expectedCount := 3
	expectedBestBidPrice := "5541.20000"
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookSnapshot()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, books.Pair)
	require.Equal(suite.T(), expectedChannelId, books.ChannelId)
	require.Len(suite.T(), books.Data.Asks, expectedCount)
	require.Len(suite.T(), books.Data.Bids, expectedCount)
	require.Equal(suite.T(), expectedBestBidPrice, books.Data.Bids[0].Price.String())
}

// Test marshalling a BookSnapshot to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonBookSnapshot() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookSnapshot()
	require.NoError(suite.T(), err)
	// Marshal BookSnapshot
	actual, err := json.Marshal(books)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with both asks and bids from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonBookUpdateBothAsksAndBids() {
	// Payload to unmarshal
	payload := `[
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
		  ]
		},
		{
		  "b": [
			[
			  "5541.30000",
			  "0.00000000",
			  "1534614335.345903"
			]
		  ],
		  "c": "974942666"
		},
		"book-10",
		"XBT/USD"
	]`
	// Expectations
	expectedChannelId := 1234
	expectedPair := "XBT/USD"
	expectedChecksum := "974942666"
	expectedBidsCount := 1
	expectedAsksCount := 2
	expectedBids0Price := "5541.30000"
	expectedAsks1Price := "5542.50000"
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, books.Pair)
	require.Equal(suite.T(), expectedChannelId, books.ChannelId)
	require.Len(suite.T(), books.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), books.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedBids0Price, books.Data.Bids[0].Price.String())
	require.Equal(suite.T(), expectedAsks1Price, books.Data.Asks[1].Price.String())
	require.Equal(suite.T(), expectedChecksum, books.Data.Checksum)
}

// Test marshalling a BookUpdate with both bids and asks to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonBookUpdateBothAsksAndBids() {
	// Payload to unmarshal
	payload := `[
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
		  ]
		},
		{
		  "b": [
			[
			  "5541.30000",
			  "0.00000000",
			  "1534614335.345903"
			]
		  ],
		  "c": "974942666"
		},
		"book-10",
		"XBT/USD"
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(books)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with only asks from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonBookUpdateOnlyAsks() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Expectations
	expectedChannelId := 1234
	expectedPair := "XBT/USD"
	expectedChecksum := "974942666"
	expectedBidsCount := 0
	expectedAsksCount := 2
	expectedAsks1Price := "5542.50000"
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, books.Pair)
	require.Equal(suite.T(), expectedChannelId, books.ChannelId)
	require.Len(suite.T(), books.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), books.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedAsks1Price, books.Data.Asks[1].Price.String())
	require.Equal(suite.T(), expectedChecksum, books.Data.Checksum)
}

// Test marshalling a BookUpdate with only asks to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonBookUpdateOnlyAsks() {
	// Payload to unmarshal
	payload := `[
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
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(books)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with only bids from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonBookUpdateOnlyBids() {
	// Payload to unmarshal
	payload := `[
		1234,
		{
		  "b": [
			[
			  "5541.30000",
			  "0.00000000",
			  "1534614335.345903"
			]
		  ],
		  "c": "974942666"
		},
		"book-10",
		"XBT/USD"
	]`
	// Expectations
	expectedChannelId := 1234
	expectedPair := "XBT/USD"
	expectedChecksum := "974942666"
	expectedBidsCount := 1
	expectedAsksCount := 0
	expectedBids0Price := "5541.30000"
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, books.Pair)
	require.Equal(suite.T(), expectedChannelId, books.ChannelId)
	require.Len(suite.T(), books.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), books.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedBids0Price, books.Data.Bids[0].Price.String())
	require.Equal(suite.T(), expectedChecksum, books.Data.Checksum)
}

// Test marshalling a BookUpdate with only bids to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonBookUpdateOnlyBids() {
	// Payload to unmarshal
	payload := `[
		1234,
		{
		  "b": [
			[
			  "5541.30000",
			  "0.00000000",
			  "1534614335.345903"
			]
		  ],
		  "c": "974942666"
		},
		"book-10",
		"XBT/USD"
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(books)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with republish from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *MarketDataUnitTestSuite) TestMarketDataUnmarshalJsonBookUpdateWithRepublish() {
	// Payload to unmarshal
	payload := `[
		1234,
		{
		  "a": [
			[
			  "5541.30000",
			  "2.50700000",
			  "1534614248.456738",
			  "r"
			],
			[
			  "5542.50000",
			  "0.40100000",
			  "1534614248.456738",
			  "r"
			]
		  ],
		  "c": "974942666"
		},
		"book-25",
		"XBT/USD"
	]`
	// Expectations
	expectedChannelId := 1234
	expectedPair := "XBT/USD"
	expectedChecksum := "974942666"
	expectedBidsCount := 0
	expectedAsksCount := 2
	expectedAsks1Price := "5542.50000"
	expectedRepublish := "r"
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, books.Pair)
	require.Equal(suite.T(), expectedChannelId, books.ChannelId)
	require.Len(suite.T(), books.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), books.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedAsks1Price, books.Data.Asks[1].Price.String())
	require.Equal(suite.T(), expectedRepublish, books.Data.Asks[1].UpdateType)
	require.Equal(suite.T(), expectedRepublish, books.Data.Asks[0].UpdateType)
	require.Equal(suite.T(), expectedChecksum, books.Data.Checksum)
}

// Test marshalling a BookUpdate with republish to the same payload as the API.
func (suite *MarketDataUnitTestSuite) TestMarketDataMarshalJsonBookUpdateWithRepublish() {
	// Payload to unmarshal
	payload := `[
		1234,
		{
		  "a": [
			[
			  "5541.30000",
			  "2.50700000",
			  "1534614248.456738",
			  "r"
			],
			[
			  "5542.50000",
			  "0.40100000",
			  "1534614248.456738",
			  "r"
			]
		  ],
		  "c": "974942666"
		},
		"book-25",
		"XBT/USD"
	]`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(MarketData)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Convert to BookSnapshot
	books, err := target.AsBookUpdate()
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(books)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
