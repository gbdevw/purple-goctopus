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

// Unit test suite for Trade
type TradeUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestTradeUnitTestSuite(t *testing.T) {
	suite.Run(t, new(TradeUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example Trade message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a Trade message
//   - Market data can be converted to a Trade
func (suite *TradeUnitTestSuite) TestTradeUnmarshalJsonTrade() {
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
	target := new(Trade)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Len(suite.T(), target.Data, expectedCount)
	require.Equal(suite.T(), expectedTrade0Price, target.Data[0].Price.String())
}

// Test marshalling a Trade to the same payload as the API.
//
// Payloads are different: They have the exact same structure but
func (suite *TradeUnitTestSuite) TestTradeMarshalJsonTrade() {
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
	target := new(Trade)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal Trade
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
