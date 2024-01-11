package websocket

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/gbdevw/gowse/wscengine"
	"github.com/gbdevw/gowse/wscengine/wsadapters"
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
	// Create a OnClose callback that will have a spy in it to know if it has been called
	called := false
	onCloseCbk := func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails) {
		called = true
	}
	// Build websocket client withthe onclose callback set and no tracing
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

// This integration test opens a connection to the server and subscribe to the ticker channel and
// read some messages (heartbeats and tickers). Once that is done, the connection will be shutdown
// to test the subscription recovery mechanism. In case of successful recovery, a unsubscibe
// message will be sent to test channel. Finally, the connection will be shutdown again to test
// if the subscription recovery mechanism does not try to subscribe again to ticker channel.
//
// Test will ensure:
//
//   - The client can subscribe to the ticker channel
//   - The client can read ticker messages and heartbeats from the server.
//   - When connection is shutdown, the engine reconnects to the server and resubscribe to channel.
//   - A nil value is present in the ticker channel data to mark data stream interruption.
//   - The client can unsubscribe from the ticker channel
//   - When connection is shutdown again, the engine reconnects to the server and do not resubscribe to channel.
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestSubscribeTicker() {
	// Build websocket client without any callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(nil, nil, nil, nil)
	// Build server URL
	url, err := url.Parse(KrakenSpotWebsocketPublicProductionURL)
	require.NoError(suite.T(), err)
	// Build the engine that will power the client - Use default options and a gorilla based connection
	engine, err := wscengine.NewWebsocketEngine(url, gorilla.NewGorillaWebsocketConnectionAdapter(nil, nil), client, nil, nil)
	require.NoError(suite.T(), err)
	// Build a context with a timeout of 20 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Start the engine and connect
	suite.T().Log("connecting to websocket server ...")
	err = engine.Start(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("connected to websocket server!")
	// Get the builtin channel for heartbeat and systemStatus
	heartbeatChan := client.GetHeartbeatChannel()
	systemStatusChan := client.GetSystemStatusChannel()
	// Subscribe to ticker
	suite.T().Log("subscribing to ticker...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	tickerChan, err := client.SubscribeTicker(ctx, pairs, 30)
	require.NoError(suite.T(), err)
	suite.T().Log("ticker subscribed!")
	// Read a ticker
	suite.T().Log("waiting for a ticker...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case ticker := <-tickerChan:
		suite.T().Log("ticker received!")
		require.Contains(suite.T(), pairs, ticker.Pair)
		require.Equal(suite.T(), string(messages.ChannelTicker), ticker.Name)
	}
	// Read a heartbeat
	suite.T().Log("waiting for a heartbeat...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case heartbeat := <-heartbeatChan:
		suite.T().Log("heartbeat received!")
		require.Equal(suite.T(), string(messages.EventTypeHeartbeat), heartbeat.Event)
	}
	// Shutdown connection - use the client.conn directly to 'trick' the engine
	suite.T().Log("shuttting down connection...")
	client.conn.Close(ctx, wsadapters.GoingAway, "client shutdown")
	suite.T().Log("connection shutdown!")
	// Read tickers until a nil value is read (= interruption is stream of data) followed by a ticker
	// This will ensure client impl. has injected a nil value in channel when connection has been lost
	// and this will ensure subscription recovery mechanism worked well.
	gapFound := false
	done := false
	for !done {
		if !gapFound {
			suite.T().Log("waiting for a nil value in ticker data...")
		} else {
			suite.T().Log("waiting for a ticker message after recovery...")
		}
		select {
		case <-ctx.Done():
			suite.FailNow(ctx.Err().Error())
		case ticker := <-tickerChan:
			if !gapFound {
				// Check for a nil value. Discard otherwise
				if ticker == nil {
					suite.T().Log("nil value  found in ticker data!")
					gapFound = true
				}
			} else {
				// Once a gap has been found, we must read at elast one ticker to ensure
				// recovery mechanism has worked as expected
				if ticker != nil {
					suite.T().Log("ticker received after recovery!")
					done = true
				} else {
					suite.FailNow("multiple nil values read in ticker data stream")
				}
			}
		}
	}
	// Unsubscribe from ticker channel
	suite.T().Log("unsubscribing from ticker channel...")
	err = client.UnsubscribeTicker(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from ticker channel!")
	// Check the internal ticker subscription is nil
	require.Nil(suite.T(), client.subscriptions.ticker)
	// Empty previously used ticker channel and system status channel
	done = false
	for !done {
		select {
		case <-tickerChan:
		case <-systemStatusChan:
		default:
			done = true
		}
	}
	// Shutdown connection - use the client.conn directly to 'trick' the engine
	client.conn.Close(ctx, wsadapters.GoingAway, "client shutdown")
	// Read a system status
	select {
	case <-systemStatusChan:
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	}
	// Check ticker is empty
	select {
	case <-tickerChan:
		suite.FailNow("ticker channel should be empty")
	default:
	}
	// Stop engine (close the connection)
	suite.T().Log("stopping the websocket engine...")
	engine.Stop(ctx)
	suite.T().Log("websocket engine stopped!")
}
