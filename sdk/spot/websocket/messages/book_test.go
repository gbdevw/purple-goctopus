package messages

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* UNIT TEST SUITE                                                                               */
/*************************************************************************************************/

// Static regex used to matches whitespaces
var matchesWhitespacesRegex = regexp.MustCompile(`\s`)

// Unit test suite for Book
type BookUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestBookUnitTestSuite(t *testing.T) {
	suite.Run(t, new(BookUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example BookSnapshot message from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *BookUnitTestSuite) TestBookUnmarshalJsonBookSnapshot() {
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
	target := new(BookSnapshot)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Len(suite.T(), target.Data.Asks, expectedCount)
	require.Len(suite.T(), target.Data.Bids, expectedCount)
	require.Equal(suite.T(), expectedBestBidPrice, target.Data.Bids[0].Price.String())
}

// Test marshalling a BookSnapshot to the same payload as the API.
func (suite *BookUnitTestSuite) TestBookMarshalJsonBookSnapshot() {
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
	target := new(BookSnapshot)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal BookSnapshot
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with both asks and bids from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *BookUnitTestSuite) TestBookUnmarshalJsonBookUpdateBothAsksAndBids() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Len(suite.T(), target.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), target.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedBids0Price, target.Data.Bids[0].Price.String())
	require.Equal(suite.T(), expectedAsks1Price, target.Data.Asks[1].Price.String())
	require.Equal(suite.T(), expectedChecksum, target.Data.Checksum)
}

// Test marshalling a BookUpdate with both bids and asks to the same payload as the API.
func (suite *BookUnitTestSuite) TestBookMarshalJsonBookUpdateBothAsksAndBids() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with only asks from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *BookUnitTestSuite) TestBookUnmarshalJsonBookUpdateOnlyAsks() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Len(suite.T(), target.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), target.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedAsks1Price, target.Data.Asks[1].Price.String())
	require.Equal(suite.T(), expectedChecksum, target.Data.Checksum)
}

// Test marshalling a BookUpdate with only asks to the same payload as the API.
func (suite *BookUnitTestSuite) TestBookMarshalJsonBookUpdateOnlyAsks() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with only bids from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *BookUnitTestSuite) TestBookUnmarshalJsonBookUpdateOnlyBids() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Len(suite.T(), target.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), target.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedBids0Price, target.Data.Bids[0].Price.String())
	require.Equal(suite.T(), expectedChecksum, target.Data.Checksum)
}

// Test marshalling a BookUpdate with only bids to the same payload as the API.
func (suite *BookUnitTestSuite) TestBookMarshalJsonBookUpdateOnlyBids() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example BookUpdate message with republish from documentation into the corresponding struct.
//
// Test will ensure:
//   - Market data custom unmarshaller can parse a BookSnapshot message
//   - Market data can be converted to a BookSnapshot
func (suite *BookUnitTestSuite) TestBookUnmarshalJsonBookUpdateWithRepublish() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check parsed data
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedChannelId, target.ChannelId)
	require.Len(suite.T(), target.Data.Asks, expectedAsksCount)
	require.Len(suite.T(), target.Data.Bids, expectedBidsCount)
	require.Equal(suite.T(), expectedAsks1Price, target.Data.Asks[1].Price.String())
	require.Equal(suite.T(), expectedRepublish, target.Data.Asks[1].UpdateType)
	require.Equal(suite.T(), expectedRepublish, target.Data.Asks[0].UpdateType)
	require.Equal(suite.T(), expectedChecksum, target.Data.Checksum)
}

// Test marshalling a BookUpdate with republish to the same payload as the API.
func (suite *BookUnitTestSuite) TestBookMarshalJsonBookUpdateWithRepublish() {
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
	target := new(BookUpdate)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal BookUpdate
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
