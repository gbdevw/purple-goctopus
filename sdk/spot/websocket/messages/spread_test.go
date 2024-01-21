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

// Unit test suite for Spread
type SpreadUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestSpreadUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SpreadUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example Spread message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a Spread message
//   - Market data can be converted to a Spread
func (suite *SpreadUnitTestSuite) TestSpreadUnmarshalJsonSpread() {
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
	target := new(Spread)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Equal(suite.T(), expectedBestBidPrice, target.Data.BestBidPrice.String())
}

// Test marshalling a Spread to the same payload as the API.
//
// Payloads are different: They have the exact same structure but
func (suite *SpreadUnitTestSuite) TestSpreadMarshalJsonSpread() {
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
	target := new(Spread)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal Spread
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
