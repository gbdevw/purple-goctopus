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

// Test unmarshalling an example Unsubscribe message from documentation into the corresponding struct.
func (suite *UnsubscribeUnitTestSuite) TestUnsubscribeUnmarshalJson1() {
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
	// Expectations
	expectedEvent := string(EventTypeUnsubscribe)
	expectedPairs := []string{"XBT/USD", "XBT/EUR"}
	expectedSubscriptionName := string(ChannelTicker)
	// Unmarshal payload into target struct
	target := new(Unsubscribe)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.ElementsMatch(suite.T(), target.Pairs, expectedPairs)
	require.Equal(suite.T(), expectedSubscriptionName, target.Subscription.Name)
}

// Test unmarshalling an example Unsubscribe message from documentation into the corresponding struct.
func (suite *UnsubscribeUnitTestSuite) TestUnsubscribeUnmarshalJson2() {
	// Payload to unmarshal
	payload := `{
		"event": "unsubscribe",
		"subscription": {
		  "name": "ownTrades",
		  "token": "WW91ciBhdXRoZW50aWNhdGlvbiB0b2tlbiBnb2VzIGhlcmUu"
		}
	}`
	// Expectations
	expectedEvent := string(EventTypeUnsubscribe)
	expectedSubscriptionName := string(ChannelOwnTrades)
	expectedSubscriptionToken := "WW91ciBhdXRoZW50aWNhdGlvbiB0b2tlbiBnb2VzIGhlcmUu"
	// Unmarshal payload into target struct
	target := new(Unsubscribe)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedSubscriptionName, target.Subscription.Name)
	require.Equal(suite.T(), expectedSubscriptionToken, target.Subscription.Token)
}
