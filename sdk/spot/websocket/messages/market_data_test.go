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

// Test unmarshalling an example MarketData message from documentation into the corresponding struct.
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
