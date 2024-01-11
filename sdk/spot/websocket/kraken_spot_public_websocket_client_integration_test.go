package websocket

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/gbdevw/gowse/wscengine"
	"github.com/gbdevw/gowse/wscengine/wsadapters/gorilla"
	"github.com/gbdevw/gowse/wscengine/wsclient"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* INTEGRATION TEST SUITE                                                                        */
/*************************************************************************************************/

// Integration test suite for KrakenSpotPublicWebsocketClient
type KrakenSpotPublicWebsocketClientIntegrationTestSuite struct {
	suite.Suite
}

// Configure and run unit test suite
func TestKrakenSpotRESTClientIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if short flag is used
	if testing.Short() {
		t.SkipNow()
	}
	/// Run the test suit
	suite.Run(t, new(KrakenSpotPublicWebsocketClientIntegrationTestSuite))
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
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestConnectionOpenningAndPing() {
	// Create a OnClose callback that will hae a spy in it to know if it has been called
	called := false
	onCloseCbk := func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails) {
		called = true
	}
	// Build websocket client without any callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(onCloseCbk, nil, nil, nil)
	// Build server URL
	url, err := url.Parse(KrakenSpotWebsocketPublicProductionURL)
	require.NoError(suite.T(), err)
	// Build the engine that will power the client - Use default options and a gorilla based connection
	engine, err := wscengine.NewWebsocketEngine(url, gorilla.NewGorillaWebsocketConnectionAdapter(nil, nil), client, nil, nil)
	require.NoError(suite.T(), err)
	// Build a context with a timeout of 15 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Start the engine and connect
	suite.T().Log("connecting to websocket server ...")
	err = engine.Start(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("connected to websocket server!")
	// Get the builtin channel for system status
	systemStatusChan := client.GetSystemStatusChannel()
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
	err = client.Ping(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("pong reply received!")
	// Stop engine (close the connection)
	suite.T().Log("stopping the websocket engine...")
	engine.Stop(ctx)
	suite.T().Log("websocket engine stopped!")
	// Check OnClose callback has been called
	require.True(suite.T(), called)
}
