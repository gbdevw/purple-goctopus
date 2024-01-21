package websocket

import (
	"context"
	"log"
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

// Integration test suite for krakenSpotWebsocketClient
type krakenSpotWebsocketClientIntegrationTestSuite struct {
	suite.Suite
}

// Configure and run unit test suite
func TestKrakenSpotRESTClientIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if short flag is used
	if testing.Short() {
		t.SkipNow()
	}
	/// Run the test suit
	suite.Run(t, new(krakenSpotWebsocketClientIntegrationTestSuite))
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
func (suite *krakenSpotWebsocketClientIntegrationTestSuite) TestConnectionOpenningAndPing() {
	// Create a OnClose callback that will have a spy in it to know if it has been called
	called := false
	onCloseCbk := func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails) {
		called = true
	}
	// Build websocket client withthe onclose callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(onCloseCbk, nil, nil, log.Default(), nil)
	// Build server URL
	url, err := url.Parse(KrakenSpotWebsocketPublicProductionURL)
	require.NoError(suite.T(), err)
	// Build the engine that will power the client - Use default options and a gorilla based connection
	engine, err := wscengine.NewWebsocketEngine(url, gorilla.NewGorillaWebsocketConnectionAdapter(nil, nil), client, nil, nil)
	require.NoError(suite.T(), err)
	// Build a context with a timeout of 15 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

// This integration test opens a connection to the server, subscribes to the ticker channel and
// reads some messages (heartbeats and tickers). Once that is done, a unsubscribe message will be
// sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the ticker channel
//   - The client can read ticker messages and heartbeats from the server.
//   - The client can unsubscribe from the ticker channel
func (suite *krakenSpotWebsocketClientIntegrationTestSuite) TestSubscribeTicker() {
	// Build websocket client without any callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(nil, nil, nil, log.Default(), nil)
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
		suite.T().Log("ticker received!", *ticker)
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
	// Unsubscribe from ticker channel
	suite.T().Log("unsubscribing from ticker channel...")
	err = client.UnsubscribeTicker(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from ticker channel!")
	// Check the internal ticker subscription is nil
	require.Nil(suite.T(), client.subscriptions.ticker)
	// Empty ticker channel
	empty := false
	for !empty {
		select {
		case <-tickerChan:
		default:
			empty = true
		}
	}
	// Check unusubscribed is OK: try tor read ticker messages for 5 seconds
	// If no messages are received, this mean everything is OK
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	select {
	case <-tickerChan:
		suite.FailNow("no ticker message should be received after unsubscribe")
	case <-ctx.Done():
	}
	// Stop engine (close the connection)
	suite.T().Log("stopping the websocket engine...")
	engine.Stop(ctx)
	suite.T().Log("websocket engine stopped!")
}

// This integration test opens a connection to the server, subscribes to the ohlc channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the ohlc channel
//   - The client can read ohlc messages from the server.
//   - The client can unsubscribe from the ohlc channel
func (suite *krakenSpotWebsocketClientIntegrationTestSuite) TestSubscribeOHLC() {
	// Build websocket client without any callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(nil, nil, nil, log.Default(), nil)
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
	// Subscribe to ohlc
	suite.T().Log("subscribing to ohlc...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	ohlcChan, err := client.SubscribeOHLC(ctx, pairs, messages.M15, 30)
	require.NoError(suite.T(), err)
	suite.T().Log("ohlc subscribed!")
	// Read a ohlc
	suite.T().Log("waiting for a OHLC...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case ohlc := <-ohlcChan:
		suite.T().Log("ohlc received!", *ohlc)
		require.Contains(suite.T(), pairs, ohlc.Pair)
		require.Contains(suite.T(), ohlc.Name, string(messages.ChannelOHLC))
	}
	// Unsubscribe from ohlc channel
	suite.T().Log("unsubscribing from ohlc channel...")
	err = client.UnsubscribeOHLC(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from ohlc channel!")
	// Check the internal ohlc subscription is nil
	require.Nil(suite.T(), client.subscriptions.ohlcs)
	// Stop engine (close the connection)
	suite.T().Log("stopping the websocket engine...")
	engine.Stop(ctx)
	suite.T().Log("websocket engine stopped!")
}

// This integration test opens a connection to the server, subscribes to the trade channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the trade channel
//   - The client can read trade messages from the server.
//   - The client can unsubscribe from the trade channel
func (suite *krakenSpotWebsocketClientIntegrationTestSuite) TestSubscribeTrade() {
	// Build websocket client without any callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(nil, nil, nil, log.Default(), nil)
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
	// Subscribe to trade
	suite.T().Log("subscribing to trade...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	tradeChan, err := client.SubscribeTrade(ctx, pairs, 30)
	require.NoError(suite.T(), err)
	suite.T().Log("trade subscribed!")
	// Read a trade
	suite.T().Log("waiting for a trade...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case trade := <-tradeChan:
		suite.T().Log("trade received!", *trade)
		require.Contains(suite.T(), pairs, trade.Pair)
		require.Contains(suite.T(), trade.Name, string(messages.ChannelTrade))
	}
	// Unsubscribe from trade channel
	suite.T().Log("unsubscribing from trade channel...")
	err = client.UnsubscribeTrade(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from trade channel!")
	// Check the internal trade subscription is nil
	require.Nil(suite.T(), client.subscriptions.trade)
	// Stop engine (close the connection)
	suite.T().Log("stopping the websocket engine...")
	engine.Stop(ctx)
	suite.T().Log("websocket engine stopped!")
}

// This integration test opens a connection to the server, subscribes to the spread channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the spread channel
//   - The client can read spread messages from the server.
//   - The client can unsubscribe from the spread channel
func (suite *krakenSpotWebsocketClientIntegrationTestSuite) TestSubscribeSpread() {
	// Build websocket client without any callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(nil, nil, nil, log.Default(), nil)
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
	// Subscribe to spread
	suite.T().Log("subscribing to spread...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	spreadChan, err := client.SubscribeSpread(ctx, pairs, 30)
	require.NoError(suite.T(), err)
	suite.T().Log("spread subscribed!")
	// Read a spread
	suite.T().Log("waiting for a spread...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case spread := <-spreadChan:
		suite.T().Log("spread received!", *spread)
		require.Contains(suite.T(), pairs, spread.Pair)
		require.Contains(suite.T(), spread.Name, string(messages.ChannelSpread))
	}
	// Unsubscribe from spread channel
	suite.T().Log("unsubscribing from spread channel...")
	err = client.UnsubscribeSpread(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from spread channel!")
	// Check the internal spread subscription is nil
	require.Nil(suite.T(), client.subscriptions.spread)
	// Stop engine (close the connection)
	suite.T().Log("stopping the websocket engine...")
	engine.Stop(ctx)
	suite.T().Log("websocket engine stopped!")
}

// This integration test opens a connection to the server, subscribes to the book channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the book channel
//   - The client can read book messages from the server.
//   - The client can unsubscribe from the book channel
func (suite *krakenSpotWebsocketClientIntegrationTestSuite) TestSubscribeBook() {
	// Build websocket client without any callback set and no tracing
	client := NewKrakenSpotPublicWebsocketClient(nil, nil, nil, log.Default(), nil)
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
	// Subscribe to book
	suite.T().Log("subscribing to book...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	bookSnapshotChan, bookUpdatesChan, err := client.SubscribeBook(ctx, pairs, messages.D10, 30)
	require.NoError(suite.T(), err)
	suite.T().Log("book subscribed!")
	// Read a book snapshot
	suite.T().Log("waiting for a book snapshot...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case snapshot := <-bookSnapshotChan:
		suite.T().Log("book snapshot received!", *snapshot)
		require.Contains(suite.T(), pairs, snapshot.Pair)
		require.Contains(suite.T(), snapshot.Name, string(messages.ChannelBook))
	}
	// Read a book update
	suite.T().Log("waiting for a book update...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case update := <-bookUpdatesChan:
		suite.T().Log("book update received!", *update)
		require.Contains(suite.T(), pairs, update.Pair)
		require.Contains(suite.T(), update.Name, string(messages.ChannelBook))
	}
	// Unsubscribe from book channel
	suite.T().Log("unsubscribing from book channel...")
	err = client.UnsubscribeBook(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from book channel!")
	// Check the internal book subscription is nil
	require.Nil(suite.T(), client.subscriptions.book)
	// Stop engine (close the connection)
	suite.T().Log("stopping the websocket engine...")
	engine.Stop(ctx)
	suite.T().Log("websocket engine stopped!")
}
