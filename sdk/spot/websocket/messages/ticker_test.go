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

// Unit test suite for Ticker
type TickerUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestTickerUnitTestSuite(t *testing.T) {
	suite.Run(t, new(TickerUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example Ticker message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a Ticker message
//   - Market data can be converted to a Ticker
func (suite *TickerUnitTestSuite) TestTickerUnmarshalJsonTicker() {
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
	target := new(Ticker)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Equal(suite.T(), expectedOpenToday, target.Data.GetTodayOpen().String())
}

// Test marshalling a Ticker to the same payload as the API.
//
// Payloads are different: They have the exact same structure but
func (suite *TickerUnitTestSuite) TestTickerMarshalJsonTicker() {
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
	target := new(Ticker)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal Ticker
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
