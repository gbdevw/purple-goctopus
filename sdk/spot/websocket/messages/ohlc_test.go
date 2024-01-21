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

// Unit test suite for OHLC
type OHLCUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestOHLCUnitTestSuite(t *testing.T) {
	suite.Run(t, new(OHLCUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example OHLC message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a OHLC message
//   - Market data can be converted to a OHLC
func (suite *OHLCUnitTestSuite) TestOHLCUnmarshalJsonOHLC() {
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
	target := new(OHLC)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Equal(suite.T(), expectedOpen, target.Data.Open.String())
	require.Equal(suite.T(), expectedCount, target.Data.TradesCount)
}

// Test marshalling a OHLC to the same payload as the API.
//
// Payloads are different: They have the exact same structure but
func (suite *OHLCUnitTestSuite) TestOHLCMarshalJsonOHLC() {
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
	target := new(OHLC)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal OHLC
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
