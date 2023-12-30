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

// Unit test suite for Subscribe
type SubscribeUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestSubscribeUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SubscribeUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test unmarshalling an example Subscribe message from documentation into the corresponding struct.
func (suite *SubscribeUnitTestSuite) TestSubscribeUnmarshalJson1() {
	// Payload to unmarshal
	payload := `{
		"event": "subscribe",
		"pair": [
		  "XBT/USD",
		  "XBT/EUR"
		],
		"subscription": {
		  "name": "ticker"
		}
	}`
	// Expectations
	expectedEvent := string(EventTypeSubscribe)
	expectedPairs := []string{"XBT/USD", "XBT/EUR"}
	expectedSubscriptionName := string(ChannelTicker)
	// Unmarshal payload into target struct
	target := new(Subscribe)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.ElementsMatch(suite.T(), target.Pairs, expectedPairs)
	require.Equal(suite.T(), expectedSubscriptionName, target.Subscription.Name)
}

// Test unmarshalling an example Subscribe message from documentation into the corresponding struct.
func (suite *SubscribeUnitTestSuite) TestSubscribeUnmarshalJson2() {
	// Payload to unmarshal
	payload := `{
		"event": "subscribe",
		"pair": [
		  "XBT/EUR"
		],
		"subscription": {
		  "interval": 5,
		  "name": "ohlc"
		}
	}`
	// Expectations
	expectedEvent := string(EventTypeSubscribe)
	expectedPairs := []string{"XBT/EUR"}
	expectedSubscriptionName := string(ChannelOHLC)
	expectedSubscriptionInterval := int(M5)
	// Unmarshal payload into target struct
	target := new(Subscribe)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.ElementsMatch(suite.T(), target.Pairs, expectedPairs)
	require.Equal(suite.T(), expectedSubscriptionName, target.Subscription.Name)
	require.Equal(suite.T(), expectedSubscriptionInterval, target.Subscription.Interval)
}

// Test unmarshalling an example Subscribe message from documentation into the corresponding struct.
func (suite *SubscribeUnitTestSuite) TestSubscribeUnmarshalJson3() {
	// Payload to unmarshal
	payload := `{
		"event": "subscribe",
		"subscription": {
		  "name": "ownTrades",
		  "token": "WW91ciBhdXRoZW50aWNhdGlvbiB0b2tlbiBnb2VzIGhlcmUu"
		}
	}`
	// Expectations
	expectedEvent := string(EventTypeSubscribe)
	expectedSubscriptionName := string(ChannelOwnTrades)
	expectedSubscriptionToken := "WW91ciBhdXRoZW50aWNhdGlvbiB0b2tlbiBnb2VzIGhlcmUu"
	// Unmarshal payload into target struct
	target := new(Subscribe)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedSubscriptionName, target.Subscription.Name)
	require.Equal(suite.T(), expectedSubscriptionToken, target.Subscription.Token)
}
