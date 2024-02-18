package websocket

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gbdevw/gowse/wscengine"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/events"
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
	// Websocket engine
	engine *wscengine.WebsocketEngine
	// Public websocket client
	wsclient KrakenSpotPublicWebsocketClientInterface
}

// Configure and run unit test suite
func TestKrakenSpotPublicWebsocketClientIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if short flag is used
	if testing.Short() {
		t.SkipNow()
	}
	// Build the client and the engine
	engine, wsclient, err := NewDefaultEngineWithPublicWebsocketClient(nil, nil, nil, log.Default(), nil)
	require.NoError(t, err)
	// Run the test suit
	suite.Run(t, &KrakenSpotPublicWebsocketClientIntegrationTestSuite{
		Suite:    suite.Suite{},
		engine:   engine,
		wsclient: wsclient,
	})
}

// Start the websocket engine and connect to the server before each test
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) BeforeTest(suiteName, testName string) {
	// STart the websocket engine and connect to the server
	err := suite.engine.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

// Stop the websocket engine and disconnect from the server after each test
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) AfterTest(suiteName, testName string) {
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
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestConnectionOpenningAndPing() {
	// Build a context with a timeout of 15 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Get the builtin channel for system status
	systemStatusChan := suite.wsclient.GetSystemStatusChannel()
	// Read a system status
	suite.T().Log("waiting for a system status message...")
	select {
	case <-ctx.Done():
		// Fail -> timeout
		suite.FailNow(ctx.Err().Error())
	case event := <-systemStatusChan:
		// Parse event data
		require.Equal(suite.T(), string(events.SystemStatus), event.Type())
		syss := new(messages.SystemStatus)
		err := event.DataAs(syss)
		require.NoError(suite.T(), err)
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

// This integration test opens a connection to the server, subscribes to the ticker channel and
// reads some messages (heartbeats and tickers). Once that is done, a unsubscribe message will be
// sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the ticker channel
//   - The client can read ticker messages and heartbeats from the server.
//   - The client can unsubscribe from the ticker channel
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestSubscribeTicker() {
	// Build a context with a timeout of 15 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Get the builtin channel for heartbeat and systemStatus
	heartbeatChan := suite.wsclient.GetHeartbeatChannel()
	// Subscribe to ticker
	suite.T().Log("subscribing to ticker...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	tickerChan := make(chan event.Event, 30)
	err := suite.wsclient.SubscribeTicker(ctx, pairs, tickerChan)
	require.NoError(suite.T(), err)
	suite.T().Log("ticker subscribed!")
	// Read a ticker
	suite.T().Log("waiting for a ticker...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case event := <-tickerChan:
		require.Equal(suite.T(), string(events.Ticker), event.Type())
		ticker := new(messages.Ticker)
		err = event.DataAs(ticker)
		require.NoError(suite.T(), err)
		suite.T().Log("ticker received!", *ticker)
		require.Contains(suite.T(), pairs, ticker.Pair)
		require.Equal(suite.T(), string(messages.ChannelTicker), ticker.Name)
	}
	// Read a heartbeat
	suite.T().Log("waiting for a heartbeat...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case event := <-heartbeatChan:
		require.Equal(suite.T(), string(events.Heartbeat), event.Type())
		heartbeat := new(messages.Heartbeat)
		err := event.DataAs(heartbeat)
		require.NoError(suite.T(), err)
		suite.T().Log("heartbeat received!")
		require.Equal(suite.T(), string(messages.EventTypeHeartbeat), heartbeat.Event)
	}
	// Unsubscribe from ticker channel
	suite.T().Log("unsubscribing from ticker channel...")
	err = suite.wsclient.UnsubscribeTicker(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from ticker channel!")
}

// This integration test opens a connection to the server, subscribes to the ohlc channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the ohlc channel
//   - The client can read ohlc messages from the server.
//   - The client can unsubscribe from the ohlc channel
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestSubscribeOHLC() {
	// Build a context with a timeout of 20 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Subscribe to ohlc
	suite.T().Log("subscribing to ohlc...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	ohlcChan := make(chan event.Event, 30)
	err := suite.wsclient.SubscribeOHLC(ctx, pairs, messages.M15, ohlcChan)
	require.NoError(suite.T(), err)
	suite.T().Log("ohlc subscribed!")
	// Read a ohlc
	suite.T().Log("waiting for a OHLC...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case event := <-ohlcChan:
		require.Equal(suite.T(), string(events.OHLC), event.Type())
		ohlc := new(messages.OHLC)
		err = event.DataAs(ohlc)
		require.NoError(suite.T(), err)
		suite.T().Log("ohlc received!", *ohlc)
		require.Contains(suite.T(), pairs, ohlc.Pair)
		require.Contains(suite.T(), ohlc.Name, string(messages.ChannelOHLC))
	}
	// Unsubscribe from ohlc channel
	suite.T().Log("unsubscribing from ohlc channel...")
	err = suite.wsclient.UnsubscribeOHLC(ctx, messages.M15)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from ohlc channel!")
}

// This integration test opens a connection to the server, subscribes to the trade channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the trade channel
//   - The client can read trade messages from the server.
//   - The client can unsubscribe from the trade channel
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestSubscribeTrade() {
	// Build a context with a timeout of 20 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Subscribe to trade
	suite.T().Log("subscribing to trade...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	tradeChan := make(chan event.Event, 30)
	err := suite.wsclient.SubscribeTrade(ctx, pairs, tradeChan)
	require.NoError(suite.T(), err)
	suite.T().Log("trade subscribed!")
	// Read a trade
	suite.T().Log("waiting for a trade...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case event := <-tradeChan:
		require.Equal(suite.T(), string(events.Trade), event.Type())
		trade := new(messages.Trade)
		err := event.DataAs(trade)
		require.NoError(suite.T(), err)
		suite.T().Log("trade received!", *trade)
		require.Contains(suite.T(), pairs, trade.Pair)
		require.Contains(suite.T(), trade.Name, string(messages.ChannelTrade))
	}
	// Unsubscribe from trade channel
	suite.T().Log("unsubscribing from trade channel...")
	err = suite.wsclient.UnsubscribeTrade(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from trade channel!")
	// Empty trade channel and catch channel closure
	empty := false
	for !empty {
		_, more := <-tradeChan
		empty = !more
	}
}

// This integration test opens a connection to the server, subscribes to the spread channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the spread channel
//   - The client can read spread messages from the server.
//   - The client can unsubscribe from the spread channel
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestSubscribeSpread() {
	// Build a context with a timeout of 20 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Subscribe to spread
	suite.T().Log("subscribing to spread...")
	pairs := []string{"XBT/USD", "XBT/EUR"}
	spreadChan := make(chan event.Event, 30)
	err := suite.wsclient.SubscribeSpread(ctx, pairs, spreadChan)
	require.NoError(suite.T(), err)
	suite.T().Log("spread subscribed!")
	// Read a spread
	suite.T().Log("waiting for a spread...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case event := <-spreadChan:
		require.Equal(suite.T(), string(events.Spread), event.Type())
		spread := new(messages.Spread)
		err = event.DataAs(spread)
		require.NoError(suite.T(), err)
		suite.T().Log("spread received!", *spread)
		require.Contains(suite.T(), pairs, spread.Pair)
		require.Contains(suite.T(), spread.Name, string(messages.ChannelSpread))
	}
	// Unsubscribe from spread channel
	suite.T().Log("unsubscribing from spread channel...")
	err = suite.wsclient.UnsubscribeSpread(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from spread channel!")
}

// This integration test opens a connection to the server, subscribes to the book channel and
// reads some messages. Once that is done, a unsubscribe message will be sent to the server.
//
// Test will ensure:
//
//   - The client can subscribe to the book channel
//   - The client can read book messages from the server.
//   - The client can unsubscribe from the book channel
func (suite *KrakenSpotPublicWebsocketClientIntegrationTestSuite) TestSubscribeBook() {
	// Build a context with a timeout of 20 seconds for the test
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Subscribe to book
	suite.T().Log("subscribing to book...")
	pairs := []string{"XBT/USD"}
	bookChan := make(chan event.Event, 30)
	err := suite.wsclient.SubscribeBook(ctx, pairs, messages.D10, bookChan)
	require.NoError(suite.T(), err)
	suite.T().Log("book subscribed!")
	// Read a book snapshot
	suite.T().Log("waiting for a book snapshot...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case event := <-bookChan:
		require.Equal(suite.T(), string(events.BookSnapshot), event.Type())
		snapshot := new(messages.BookSnapshot)
		err = event.DataAs(snapshot)
		require.NoError(suite.T(), err)
		suite.T().Log("book snapshot received!", *snapshot)
		require.Contains(suite.T(), pairs, snapshot.Pair)
		require.Contains(suite.T(), snapshot.Name, string(messages.ChannelBook))
	}
	// Read a book update
	suite.T().Log("waiting for a book update...")
	select {
	case <-ctx.Done():
		suite.FailNow(ctx.Err().Error())
	case event := <-bookChan:
		require.Equal(suite.T(), string(events.BookUpdate), event.Type())
		update := new(messages.BookUpdate)
		err = event.DataAs(update)
		require.NoError(suite.T(), err)
		suite.T().Log("book update received!", *update)
		require.Contains(suite.T(), pairs, update.Pair)
		require.Contains(suite.T(), update.Name, string(messages.ChannelBook))
	}
	// Unsubscribe from book channel
	suite.T().Log("unsubscribing from book channel...")
	err = suite.wsclient.UnsubscribeBook(ctx)
	require.NoError(suite.T(), err)
	suite.T().Log("unsubscribed from book channel!")
}
