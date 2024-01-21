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

// Unit test suite for Unsubscribe
type UnsubscribeUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestUnsubscribeUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnsubscribeUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example Unsubscribe message from documentation into the same payload.
func (suite *UnsubscribeUnitTestSuite) TestUnsubscribeMarshalJson1() {
	// Payload to unmarshal
	payload := `{
		"event": "unsubscribe",
		"pair": [
		  "XBT/EUR",
		  "XBT/USD"
		],
		"subscription": {
		  "name": "ticker"
		}
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(Unsubscribe)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test marshalling an example Unsubscribe message from documentation into the same payload.
func (suite *UnsubscribeUnitTestSuite) TestUnsubscribeMarshalJson2() {
	// Payload to unmarshal
	payload := `{
		"event": "unsubscribe",
		"subscription": {
		  "name": "ownTrades",
		  "token": "WW91ciBhdXRoZW50aWNhdGlvbiB0b2tlbiBnb2VzIGhlcmUu"
		}
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(Unsubscribe)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
