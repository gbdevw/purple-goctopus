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

// Unit test suite for SubscriptionStatus
type SubscriptionStatusUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestSubscriptionStatusUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionStatusUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example SubscriptionStatus message from documentation to the same payload.
func (suite *SubscriptionStatusUnitTestSuite) TestSubscriptionStatusMarshalJson1() {
	// Payload to marshal
	payload := `{
		"channelName": "ticker",
		"event": "subscriptionStatus",
		"pair": "XBT/EUR",
		"status": "subscribed",
		"subscription": {
		  "name": "ticker"
		}
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(SubscriptionStatus)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test marshalling an example SubscriptionStatus message from documentation to the same payload.
func (suite *SubscriptionStatusUnitTestSuite) TestSubscriptionStatusMarshalJson2() {
	// Payload to unmarshal
	payload := `{
		"channelName": "ohlc-5",
		"event": "subscriptionStatus",
		"reqid": 42,
		"pair": "XBT/EUR",
		"status": "unsubscribed",
		"subscription": {
		  "interval": 5,
		  "name": "ohlc"
		}
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(SubscriptionStatus)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example SubscriptionStatus message from documentation into the corresponding struct.
func (suite *SubscriptionStatusUnitTestSuite) TestSubscriptionStatusUnmarshalJson3() {
	// Payload to unmarshal
	payload := `{
		"event": "subscriptionStatus",
		"pair": "XBT/USD",
		"status": "error",
		"errorMessage": "Subscription depth not supported",
		"subscription": {
		  "depth": 42,
		  "name": "book"
		}
	}`
	// Remove whitespaces
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal payload into target struct
	target := new(SubscriptionStatus)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Compare
	require.Equal(suite.T(), payload, string(actual))
}
