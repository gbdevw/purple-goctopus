package websocket

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gbdevw/gowse/wscengine"
	restcommon "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* INTEGRATION TEST SUITE                                                                        */
/*************************************************************************************************/

// Integration test suite for KrakenSpotPrivateWebsocketClient
type KrakenSpotPrivateWebsocketClientIntegrationTestSuite struct {
	suite.Suite
	// Websocket engine
	engine *wscengine.WebsocketEngine
	// Private websocket client
	wsclient KrakenSpotPrivateWebsocketClientInterface
}

// Configure and run unit test suite
func TestKrakenSpotPrivateWebsocketClientIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if short flag is used
	if testing.Short() {
		t.SkipNow()
	}
	// Load credentials for Kraken API
	key := os.Getenv("KRAKEN_API_KEY")
	b64Secret := os.Getenv("KRAKEN_API_SECRET")
	otp := os.Getenv("KRAKEN_API_OTP") // Optional
	// Create secopts from otp
	var secopts *restcommon.SecurityOptions = nil
	if otp != "" {
		secopts = &restcommon.SecurityOptions{
			SecondFactor: otp,
		}
	}
	// Build the websocket engine & private client with nice defaults
	engine, wsclient, err := NewDefaultEngineWithPrivateWebsocketClient(key, b64Secret, secopts, nil, nil, nil, log.Default(), nil)
	require.NoError(t, err)
	require.NotNil(t, wsclient)
	require.NotNil(t, engine)
	// Run the test suit
	suite.Run(t, &KrakenSpotPrivateWebsocketClientIntegrationTestSuite{
		Suite:    suite.Suite{},
		engine:   engine,
		wsclient: wsclient,
	})
}

// Start the websocket engine and connect to the server before each test
func (suite *KrakenSpotPrivateWebsocketClientIntegrationTestSuite) BeforeTest(suiteName, testName string) {
	// STart the websocket engine and connect to the server
	err := suite.engine.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

// Stop the websocket engine and disconnect from the server after each test
func (suite *KrakenSpotPrivateWebsocketClientIntegrationTestSuite) AfterTest(suiteName, testName string) {
	// STart the websocket engine and connect to the server
	err := suite.engine.Stop(context.Background())
	if err != nil {
		panic(err)
	}
}

/*************************************************************************************************/
/* INTEGRATION TESTS                                                                             */
/*************************************************************************************************/

// This integration test opens a connection to the server and send a Ping request.
//
// Test will ensure:
//   - The client can open a connection to the websocket server
//   - The client can read the initial status message from the server
//   - The client can send a Ping to the server ad read its response
//   - The client OnCloseCallback is called when connection is shutdown from client side
func (suite *KrakenSpotPrivateWebsocketClientIntegrationTestSuite) TestConnectionOpenningAndPing() {
	// Build a context with a timeout of 15 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// Get the builtin channel for system status
	systemStatusChan := suite.wsclient.GetSystemStatusChannel()
	// Read a system status
	suite.T().Log("waiting for a system status message...")
	select {
	case <-ctx.Done():
		// Fail -> timeout
		suite.FailNow(ctx.Err().Error())
	case syss := <-systemStatusChan:
		// Check received messages
		suite.T().Log("system status message received!")
		require.NotEmpty(suite.T(), syss.Status)
		require.NotEmpty(suite.T(), syss.Version)
		require.NotEmpty(suite.T(), syss.ConnectionId)
		require.Equal(suite.T(), string(messages.EventTypeSystemStatus), syss.Event)
	}
	// Send a Ping
	suite.T().Log("sending a ping message...")
	err := suite.wsclient.Ping(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("pong reply received!")
}

// This integration test opens a connection to the server and send a AddOrder request for validation
// to the server.
//
// Test will ensure:
//   - The client can open a connection to the websocket server
//   - The client can send a valid AddOrder request and process the response
func (suite *KrakenSpotPrivateWebsocketClientIntegrationTestSuite) TestAddOrder() {
	// Build a context with a timeout of 15 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// Prepare order
	order := AddOrderRequestParameters{
		OrderType:     string(messages.Market),
		Type:          string(messages.Buy),
		Pair:          "XBT/USD",
		Volume:        "0.0002",
		UserReference: "42",
		Validate:      true,
	}
	// Send a addOrder (validate = true)
	suite.T().Log("sending a addOrder message...")
	resp, err := suite.wsclient.AddOrder(ctx, order)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
	suite.T().Log("addOrder response received!")
	// Check response
	require.Equal(suite.T(), string(messages.Ok), resp.Status)
	require.Empty(suite.T(), resp.Err)
	require.NotEmpty(suite.T(), resp.Description)
}
