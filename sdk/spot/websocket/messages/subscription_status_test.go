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

// Test unmarshalling an example SubscriptionStatus message from documentation into the corresponding struct.
func (suite *SubscriptionStatusUnitTestSuite) TestSubscriptionStatusUnmarshalJson1() {
	// Payload to unmarshal
	payload := `{
		"channelID": 10001,
		"channelName": "ticker",
		"event": "subscriptionStatus",
		"pair": "XBT/EUR",
		"status": "subscribed",
		"subscription": {
		  "name": "ticker"
		}
	}`
	// Expectations
	expectedEvent := string(EventTypeSubscriptionStatus)
	expectedPair := "XBT/EUR"
	expectedName := string(ChannelTicker)
	expectedStatus := string(Subscribed)
	// Unmarshal payload into target struct
	target := new(SubscriptionStatus)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedName, target.ChannelName)
	require.Equal(suite.T(), expectedStatus, target.Status)
	require.Equal(suite.T(), expectedName, target.Subscription.Name)
}

// Test unmarshalling an example SubscriptionStatus message from documentation into the corresponding struct.
func (suite *SubscriptionStatusUnitTestSuite) TestSubscriptionStatusUnmarshalJson2() {
	// Payload to unmarshal
	payload := `{
		"channelID": 10001,
		"channelName": "ohlc-5",
		"event": "subscriptionStatus",
		"pair": "XBT/EUR",
		"reqid": 42,
		"status": "unsubscribed",
		"subscription": {
		  "interval": 5,
		  "name": "ohlc"
		}
	}`
	// Expectations
	expectedEvent := string(EventTypeSubscriptionStatus)
	expectedPair := "XBT/EUR"
	expectedReqId := 42
	expectedStatus := string(Unsubscribed)
	expectedChannelName := "ohlc-5"
	expectedSubscriptionName := string(ChannelOHLC)
	expectedSubscriptionInterval := int(M5)
	// Unmarshal payload into target struct
	target := new(SubscriptionStatus)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedChannelName, target.ChannelName)
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedReqId, target.ReqId)
	require.Equal(suite.T(), expectedStatus, target.Status)
	require.Equal(suite.T(), expectedSubscriptionName, target.Subscription.Name)
	require.Equal(suite.T(), expectedSubscriptionInterval, target.Subscription.Interval)
}

// Test unmarshalling an example SubscriptionStatus message from documentation into the corresponding struct.
func (suite *SubscriptionStatusUnitTestSuite) TestSubscriptionStatusUnmarshalJson3() {
	// Payload to unmarshal
	payload := `{
		"errorMessage": "Subscription depth not supported",
		"event": "subscriptionStatus",
		"pair": "XBT/USD",
		"status": "error",
		"subscription": {
		  "depth": 42,
		  "name": "book"
		}
	}`
	// Expectations
	expectedEvent := string(EventTypeSubscriptionStatus)
	expectedErrMsg := "Subscription depth not supported"
	expectedPair := "XBT/USD"
	expectedStatus := string(Error)
	expectedSubscriptionName := string(ChannelBook)
	expectedDepth := 42
	// Unmarshal payload into target struct
	target := new(SubscriptionStatus)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedErrMsg, target.Err)
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedPair, target.Pair)
	require.Equal(suite.T(), expectedStatus, target.Status)
	require.Equal(suite.T(), expectedSubscriptionName, target.Subscription.Name)
	require.Equal(suite.T(), expectedDepth, target.Subscription.Depth)
}
