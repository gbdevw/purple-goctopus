package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gbdevw/gowse/wscengine/wsadapters"
	"github.com/gbdevw/gowse/wscengine/wsclient"
	"github.com/gbdevw/purple-goctopus/sdk/noncegen"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// URL for Kraken spot websocket client - public endpoints - Production
	KrakenSpotWebsocketPublicProductionURL = "wss://ws.kraken.com"
)

type KrakenSpotPublicWebsocketClient struct {
	// Websocket connection adapter to use to interact with the chosen
	// underlying low-level websocket framework.
	conn wsadapters.WebsocketConnectionAdapterInterface
	// Nonce generator used to generate unique request IDs
	ngen noncegen.NonceGenerator
	// Subscriptions which must be maintained by the websocket client.
	subscriptions activeSubscriptions
	// Pending requests that must be served by the client.
	requests pendingRequests
	// User provided callback which extends OnClose logic. Callback will be called when connection
	// with the server is closed or lost.
	onCloseCallback func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails)
	// User provided callback which extends OnReadError logic. Callback will be called when an error
	// occurs while reading incoming messages from the server.
	onReadErrorCallback func(ctx context.Context, restart context.CancelFunc, exit context.CancelFunc, err error)
	// User provided callback which extends OnRestartError logic. Callback will be called when an
	// error occurs while the engine is reconnecting to the server.
	onRestartError func(ctx context.Context, exit context.CancelFunc, err error, retryCount int)
	// Tracer used to instrument code
	tracer trace.Tracer
	// Mutex used to protect access to pending requests while client is discarding them.
	mu sync.Mutex
	// Mutex used to protect ticker subscribe/unsubscribe methods
	tickerSubMu sync.Mutex
	// Mutex used to protect ohlc subscribe/unsubscribe methods
	ohlcSubMu sync.Mutex
	// Mutex used to protect trade subscribe/unsubscribe methods
	tradeSubMu sync.Mutex
	// Mutex used to protect spread subscribe/unsubscribe methods
	spreadSubMu sync.Mutex
	// Mutex used to protect book subscribe/unsubscribe methods
	bookSubMu sync.Mutex
}

// # Description
//
// Build a new KrakenSpotPublicWebsocketClient. The client must be provided to a wscengine.WebsocketEngine
// which will manage the connection with the server and execute the client's logic. Once connected, the
// client can be used to send messages to the websocket server, subscribe to data feeds, ...
//
// # Inputs
//
//   - onCloseCallback: optional user defined callback which will be called when connection is closed/interrupted.
//   - onReadErrorCallback: optional user defined callback which will be called when an error occurs while reading messages from the websocket server
//   - onRestartError: optional user defined callback which will be called when the websocket engine fails to reconnect to the server.
//   - tracerProvider: Tracer provider to use to get a tracer to instrument websocket client code. If nil, global tracer provider will be used.
//
// # Return
//
// A new KrakenSpotPublicWebsocketClient which can then be used by a wscengine.WebsocketEngine.
func NewKrakenSpotPublicWebsocketClient(
	onCloseCallback func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails),
	onReadErrorCallback func(ctx context.Context, restart context.CancelFunc, exit context.CancelFunc, err error),
	onRestartError func(ctx context.Context, exit context.CancelFunc, err error, retryCount int),
	tracerProvider trace.TracerProvider,
) *KrakenSpotPublicWebsocketClient {
	// Use the global tracer provider if none is provided
	if tracerProvider == nil {
		tracerProvider = otel.GetTracerProvider()
	}
	return &KrakenSpotPublicWebsocketClient{
		conn: nil,
		ngen: noncegen.NewHFNonceGenerator(),
		subscriptions: activeSubscriptions{
			ticker:       &tickerSubscription{},
			ohlcs:        &ohlcSubscription{},
			trade:        &tradeSubscription{},
			spread:       &spreadSubscription{},
			book:         &bookSubscription{},
			ownTrades:    &ownTradesSubscription{},
			openOrders:   &openOrdersSubscription{},
			heartbeat:    make(chan *messages.Heartbeat, 10),
			systemStatus: make(chan *messages.SystemStatus, 10),
		},
		requests: pendingRequests{
			pendingPing:                          map[int64]*pendingPing{},
			pendingSubscribe:                     map[int64]*pendingSubscribe{},
			pendingUnsubscribe:                   map[int64]*pendingUnsubscribe{},
			pendingAddOrderRequests:              map[int64]*pendingAddOrderRequest{},
			pendingEditOrderRequests:             map[int64]*pendingEditOrderRequest{},
			pendingCancelOrderRequests:           map[int64]*pendingCancelOrderRequest{},
			pendingCancelAllOrdersRequests:       map[int64]*pendingCancelAllOrdersRequest{},
			pendingCancelAllOrdersAfterXRequests: map[int64]*pendingCancelAllOrdersAfterXRequest{},
		},
		onCloseCallback:     onCloseCallback,
		onReadErrorCallback: onReadErrorCallback,
		onRestartError:      onRestartError,
		tracer:              tracerProvider.Tracer(tracing.PackageName, trace.WithInstrumentationVersion(tracing.PackageVersion)),
		mu:                  sync.Mutex{},
		tickerSubMu:         sync.Mutex{},
		ohlcSubMu:           sync.Mutex{},
		tradeSubMu:          sync.Mutex{},
		spreadSubMu:         sync.Mutex{},
		bookSubMu:           sync.Mutex{},
	}
}

/*************************************************************************************************/
/* WEBSOCKET ENGINE CLIENT IMPLEMENTATION                                                        */
/*************************************************************************************************/

// # Description
//
// Callback called when engine has (re)opened a connection to the websocket server. OnOpen is
// called once, synchronously by the engine during its (re)start phase: no messages or events
// will be processed until callback completes or a timeout occurs (default: 5 minutes).
//
// If OnOpen callback returns an error, websocket engine will:
//   - If starting: engine will close the opened connection and stop.
//   - If restarting: engine will close the opened connection and try to restart again.
//
// No other callbacks (OnReadError & OnClose) will be used in such cases.
//
// During OnOpen call, the provided exit function can be called to definitely stop the engine.
//
// # Inputs
//
//   - ctx: context produced from the websocket engine context and bound to OnOpen lifecycle.
//   - resp: The server response to the websocket handshake.
//   - conn: Websocket adapter provided during engine creation. Connection is now opened.
//   - readMutex: A reference to engine read mutex user can lock to pause the engine.
//   - exit: Function to call to definitely stop the engine (ex: when stuck in retry loop).
//   - restarting: Flag which indicates whether engine restarts (true) or is starting (false).
//
// # Returns
//
// nil in case of success or an error if an error occured during OnOpen execution.
//
// When engine is restarting, returning an error will cause engine to restart again.
//
// # Engine behavior after OnOpen completes
//
// If nil is returned and if exit function has not been called, engine will finish starting
// and create internal goroutines which will manage the websocket connection.
//
// If an error is returned, engine will close the opened connection and do the following:
//   - If engine is starting, engine will definitely stop. Calling exit will do nothing.
//   - If engine is restarting, engine will try again to restart.
//   - If engine is restarting and exit has been called, engine will definitely stop.
func (client *KrakenSpotPublicWebsocketClient) OnOpen(
	ctx context.Context,
	resp *http.Response,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	exit context.CancelFunc,
	restarting bool) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".on_open", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(
		attribute.Bool("restarting", restarting),
	))
	defer span.End()
	// Store new connection
	client.conn = conn
	// Restore all active subscriptions if restarting
	if restarting {
		if client.subscriptions.ticker != nil {
			err := client.sendSubscribeRequest(
				ctx,
				&messages.Subscribe{
					Event: string(messages.EventTypeSubscribe),
					ReqId: client.ngen.GenerateNonce(),
					Pairs: client.subscriptions.ticker.pairs,
					Subscription: messages.SuscribeDetails{
						Name: string(messages.ChannelTicker),
					},
				},
				// Error will not be watched: Subscriptions should not fail as they have been used once.
				// In case of failure, it is up to the user to monitor interruptions in the stream of
				// data and react if no data is received for a long time. User can unsubscribe and then
				// subscribe again, for example
				//
				// Use a capcity of 1 so writer will not be blocked in any case
				make(chan error, 1))
			if err != nil {
				// Return an error so engine will restart again and retry to subscribe.
				// No message will be processed until OnOpen succeed.
				eerr := fmt.Errorf("failed to resubscribe to ticker channel: %w", err)
				return tracing.HandleAndTraceError(span, eerr)
			}
		}
		if client.subscriptions.ohlcs != nil {
			err := client.sendSubscribeRequest(
				ctx,
				&messages.Subscribe{
					Event: string(messages.EventTypeSubscribe),
					ReqId: client.ngen.GenerateNonce(),
					Pairs: client.subscriptions.ohlcs.pairs,
					Subscription: messages.SuscribeDetails{
						Name:     string(messages.ChannelOHLC),
						Interval: int(client.subscriptions.ohlcs.interval),
					},
				},
				// Error will not be watched: Subscriptions should not fail as they have been used once.
				// In case of failure, it is up to the user to monitor interruptions in the stream of
				// data and react if no data is received for a long time. User can unsubscribe and then
				// subscribe again, for example
				//
				// Use a capcity of 1 so writer will not be blocked in any case
				make(chan error, 1))
			if err != nil {
				// Return an error so engine will restart again and retry to subscribe.
				// No message will be processed until OnOpen succeed.
				eerr := fmt.Errorf("failed to resubscribe to ohlc channel: %w", err)
				return tracing.HandleAndTraceError(span, eerr)
			}
		}
		if client.subscriptions.trade != nil {
			err := client.sendSubscribeRequest(
				ctx,
				&messages.Subscribe{
					Event: string(messages.EventTypeSubscribe),
					ReqId: client.ngen.GenerateNonce(),
					Pairs: client.subscriptions.trade.pairs,
					Subscription: messages.SuscribeDetails{
						Name: string(messages.ChannelTrade),
					},
				},
				// Error will not be watched: Subscriptions should not fail as they have been used once.
				// In case of failure, it is up to the user to monitor interruptions in the stream of
				// data and react if no data is received for a long time. User can unsubscribe and then
				// subscribe again, for example
				//
				// Use a capcity of 1 so writer will not be blocked in any case
				make(chan error, 1))
			if err != nil {
				// Return an error so engine will restart again and retry to subscribe.
				// No message will be processed until OnOpen succeed.
				eerr := fmt.Errorf("failed to resubscribe to trade channel: %w", err)
				return tracing.HandleAndTraceError(span, eerr)
			}
		}
		if client.subscriptions.spread != nil {
			err := client.sendSubscribeRequest(
				ctx,
				&messages.Subscribe{
					Event: string(messages.EventTypeSubscribe),
					ReqId: client.ngen.GenerateNonce(),
					Pairs: client.subscriptions.spread.pairs,
					Subscription: messages.SuscribeDetails{
						Name: string(messages.ChannelSpread),
					},
				},
				// Error will not be watched: Subscriptions should not fail as they have been used once.
				// In case of failure, it is up to the user to monitor interruptions in the stream of
				// data and react if no data is received for a long time. User can unsubscribe and then
				// subscribe again, for example
				//
				// Use a capcity of 1 so writer will not be blocked in any case
				make(chan error, 1))
			if err != nil {
				// Return an error so engine will restart again and retry to subscribe.
				eerr := fmt.Errorf("failed to resubscribe to spread channel: %w", err)
				return tracing.HandleAndTraceError(span, eerr)
			}
		}
		if client.subscriptions.book != nil {
			err := client.sendSubscribeRequest(
				ctx,
				&messages.Subscribe{
					Event: string(messages.EventTypeSubscribe),
					ReqId: client.ngen.GenerateNonce(),
					Pairs: client.subscriptions.book.pairs,
					Subscription: messages.SuscribeDetails{
						Name:  string(messages.ChannelBook),
						Depth: int(client.subscriptions.book.depth),
					},
				},
				// Error will not be watched: Subscriptions should not fail as they have been used once.
				// In case of failure, it is up to the user to monitor interruptions in the stream of
				// data and react if no data is received for a long time. User can unsubscribe and then
				// subscribe again, for example
				//
				// Use a capcity of 1 so writer will not be blocked in any case
				make(chan error, 1))
			if err != nil {
				// Return an error so engine will restart again and retry to subscribe.
				eerr := fmt.Errorf("failed to resubscribe to book channel: %w", err)
				return tracing.HandleAndTraceError(span, eerr)
			}
		}
	}
	// Return nil, will complete connection opening
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// # Description
//
// Callback called when a message is read from the server. The goroutine which has read the
// message will block until callback completes. Meanwhile, other goroutines, if any, can read
// and process other incoming messages unless read mutex is locked.
//
// # Inputs
//
//   - ctx: context produce from websocket engine context and bound to OnMessage lifecycle.
//   - conn: Websocket adapter provided during engine creation with a connection opened.
//   - readMutex: A reference to engine read mutex user can lock to pause the engine.
//   - restart: Function to call to instruct engine to stop and restart.
//   - exit: Function to call to definitely stop the engine.
//   - sessionId: Unique identifier produced by engine for each new websocket connection and
//     bound to the websocket connection lifetime.
//   - msgType: Message type returned by read function.
//   - msg: Received message as a byte array
//
// # Engine behavior on exit/restart call
//
//   - No other messages will be read if restart or exit is called.
//
//   - Engine will stop after OnMessage is completed: OnClose callback is called and then the
//     connection is closed. Depending on which function was called, the engine will restart or
//     stop for good.
//
//   - All pending messages will be discarded. The user can continue to read and send messages
//     in this callback and/or in the OnClose callback until conditions are met to stop the
//     engine and close the websocket connection.
func (client *KrakenSpotPublicWebsocketClient) OnMessage(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".on_message",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("session_id", sessionId),
		))
	defer span.End()
	// Match the message type - 3 matches are expected [<message>, <event type if msg is a JSON object>,<channel name if msg is a JSON Array>]
	matches := messages.MatchMessageTypeRegex.FindStringSubmatch(string(msg))
	if len(matches) != 3 {
		// Call OnReadError - Not the expected number of matches
		err := fmt.Errorf("failed to extract the message type from '%s'", string(msg))
		tracing.HandleAndTraceError(span, err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return
	}
	// Extract the message type from the matches. The regex will try to find the event type in a JSON Object (index 1) or
	// in a JSON Array (index 2). If the first item contains an empty string, use the second.
	//
	// The item at the index 0 contains the matched message, so the original message.
	mType := matches[1]
	if mType == "" {
		mType = matches[2]
	}
	// Depending on the message type.
	switch mType {
	// Heartbeat
	case string(messages.EventTypeHeartbeat):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleHeartbeat(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// System status
	case string(messages.EventTypeSystemStatus):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleSystemStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Pong
	case string(messages.EventTypePong):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handlePong(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Subscribe/Unsubscribe responses
	case string(messages.EventTypeSubscriptionStatus):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleSubscriptionStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	case string(messages.ChannelTicker):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleTicker(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	case string(messages.ChannelOHLC):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleOHLC(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	case string(messages.ChannelTrade):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleTrade(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	case string(messages.ChannelSpread):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleSpread(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	case string(messages.ChannelBook):
		// Do not handle again the error returned by the internal method (it is there for testing purpose)
		// Report errors in on_message only when message could not be processed at all.
		client.handleBook(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	default:
		// Call OnReadError - Unknown message type
		eerr := fmt.Errorf("unkown or unexpected message type (%s) extracted from '%s'", mType, string(msg))
		tracing.HandleAndTraceError(span, eerr)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return
	}
	// Set span status to OK and exit
	span.SetStatus(codes.Ok, codes.Ok.String())
}

// # Description
//
// This callback is called each time an error is received when reading messages from the
// websocket server that is not caused by the connection being closed.
//
// The callback is called by the engine goroutine that encountered the error. All engine
// goroutines will block until the callback is completed. This prevents other messages and
// events from being processed by the engine while the error is being handled.
//
// The engine will restart after OnReadError has finished if one of the following conditions
// is met:
// - The websocket connection is closed and the Exit function has not been called.
// - The restart function has been called.
//
// Otherwise, the engine will either continue to process messages on the same connection or
// shut down if the exit function has been called.
//
// Do not close the websocket connection manually: It will be automatically closed if necessary
// after the OnClose callback has been completed.
//
// # Inputs
//
//   - ctx: Context produced from the websocket engine context and bound to OnReadError lifecycle.
//   - conn: Connection to the websocket server.
//   - readMutex: A reference to engine read mutex user can lock to pause the engine.
//   - restart: Function to call to instruct engine to stop and restart.
//   - exit: Function to call to definitely stop the engine.
//   - err: Error returned by the websocket read operation.
//
// # Engine behavior on exit/restart call
//
//   - No other messages are read when restart or exit is called.
//
//   - Engine will stop after OnReadError: OnClose callback is called and then the connection is
//     closed. Depending on which function was called, the engine will restart or stop for good.
//
//   - All pending messages will be discarded. The user can continue to read and send messages
//     in this callback and/or in the OnClose callback until conditions are met to stop the
//     engine and close the websocket connection.
func (client *KrakenSpotPublicWebsocketClient) OnReadError(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	err error) {
	// Tracing: start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".on_read_error", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	// Call user callback if set
	if client.onReadErrorCallback != nil {
		client.onReadErrorCallback(ctx, restart, exit, err)
	}
}

// Called by the engine when the connection with the server is closed, lost or
// about to be closed because the user has called Stop.
func (client *KrakenSpotPublicWebsocketClient) OnClose(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	closeMessage *wsclient.CloseMessageDetails) *wsclient.CloseMessageDetails {
	// Tracing: start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".on_close", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	// Lock mutex before starting to discard pending requests
	client.mu.Lock()
	defer client.mu.Unlock()
	// Discard pending ping requests to unlock all blocked thread waiting for a response.
	for reqid, req := range client.requests.pendingPing {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- &OperationInterruptedError{
			Operation: "ping",
			Root:      fmt.Errorf("connection has been closed"),
		}
		// Remove pending request
		delete(client.requests.pendingPing, reqid)
	}
	// Discard pending subscribe requests
	for reqid, req := range client.requests.pendingSubscribe {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- &OperationInterruptedError{
			Operation: "subscribe",
			Root:      fmt.Errorf("connection has been closed"),
		}
		// Remove pending request
		delete(client.requests.pendingSubscribe, reqid)
	}
	// Discard pending unsubscribe requests
	for reqid, req := range client.requests.pendingUnsubscribe {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- &OperationInterruptedError{
			Operation: "subscribe",
			Root:      fmt.Errorf("connection has been closed"),
		}
		// Remove pending request
		delete(client.requests.pendingUnsubscribe, reqid)
	}
	// Send a nil value on all active subscriptions
	// Use blocking writes (design principle: wait 'till delivery)
	client.tickerSubMu.Lock()
	defer client.tickerSubMu.Unlock()
	if client.subscriptions.ticker != nil {
		client.subscriptions.ticker.pub <- nil
	}
	client.ohlcSubMu.Lock()
	defer client.ohlcSubMu.Unlock()
	if client.subscriptions.ohlcs != nil {
		client.subscriptions.ohlcs.pub <- nil
	}
	client.tradeSubMu.Lock()
	defer client.tradeSubMu.Unlock()
	if client.subscriptions.trade != nil {
		client.subscriptions.trade.pub <- nil
	}
	client.spreadSubMu.Lock()
	defer client.spreadSubMu.Unlock()
	if client.subscriptions.spread != nil {
		client.subscriptions.spread.pub <- nil
	}
	client.bookSubMu.Lock()
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book != nil {
		client.subscriptions.book.snapshots <- nil
		client.subscriptions.book.updates <- nil
	}
	// Call user callback if set
	if client.onCloseCallback != nil {
		client.onCloseCallback(ctx, closeMessage)
	}
	// Remove conn & return
	client.conn = nil
	return closeMessage
}

// # Description
//
// Callback called if an error occurred when the engine called the conn.Close method during
// the shutdown phase.
//
// # Inputs
//
//   - ctx:  Context produced OnClose context.
//   - err: Error returned by conn.Close method
func (client *KrakenSpotPublicWebsocketClient) OnCloseError(
	ctx context.Context,
	err error) {
	// Tracing: start span
	_, span := client.tracer.Start(ctx, tracing.TracesNamespace+".on_close_error",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("error", err.Error())))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	// Do nothing
}

// # Description
//
// Callback called in case an error or a timeout occured when engine tried to restart.
//
// # Inputs
//
//   - ctx:  Context used for tracing purpose. Will be Done in case a timeout has occured.
//   - exit: Function to call to stop trying to restart the engine.
//   - err: Error which has occured when restarting the engine
//   - retryCount: Number of restart retry since last time engine has successfully (re)started.
func (client *KrakenSpotPublicWebsocketClient) OnRestartError(
	ctx context.Context,
	exit context.CancelFunc,
	err error,
	retryCount int) {
	// Tracing: start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".on_restart_error",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.Int("retry_count", retryCount),
			attribute.String("error", err.Error()),
		))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	// Call user callback if set
	if client.onRestartError != nil {
		client.onRestartError(ctx, exit, err, retryCount)
	}
}

/*************************************************************************************************/
/* MESSAGE HANDLERS                                                                              */
/*************************************************************************************************/

// This method contains the logic to handle a received heartbeat message.
func (client *KrakenSpotPublicWebsocketClient) handleHeartbeat(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_heartbeat",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as heartbeat
	heartbeat := new(messages.Heartbeat)
	err := json.Unmarshal(msg, heartbeat)
	if err != nil {
		// Call OnReadError - failed to parse message as heartbeat
		eerr := fmt.Errorf("failed to parse message '%s' as heartbeat: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Publish heartbeat - as user might not actively listen to heartbeats, manage the channel in FIFO
	// fashion by discarding oldest messages in case of congestion
	select {
	case client.subscriptions.heartbeat <- heartbeat:
	default:
		// Discard oldest heartbeat & push new one
		<-client.subscriptions.heartbeat
		client.subscriptions.heartbeat <- heartbeat
	}
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received systemStatus message.
func (client *KrakenSpotPublicWebsocketClient) handleSystemStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_system_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as system status
	systemStatus := new(messages.SystemStatus)
	err := json.Unmarshal(msg, systemStatus)
	if err != nil {
		// Call OnReadError - failed to parse message as systemStatus
		eerr := fmt.Errorf("failed to parse message '%s' as systemStatus: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Tracing: add event
	span.AddEvent(tracing.TracesNamespace+".system_status", trace.WithAttributes(
		attribute.String("session_id", sessionId),
		attribute.String("status", systemStatus.Status),
		attribute.String("version", systemStatus.Version),
		attribute.Int64("connection_id", systemStatus.ConnectionId),
	))
	// Publish heartbeat - as user might not actively listen to system statuses, manage the channel
	// in FIFO fashion by discarding oldest messages in case of congestion
	select {
	case client.subscriptions.systemStatus <- systemStatus:
	default:
		// Discard oldest heartbeat & push new one
		<-client.subscriptions.systemStatus
		client.subscriptions.systemStatus <- systemStatus
	}
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received pong message.
func (client *KrakenSpotPublicWebsocketClient) handlePong(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_pong",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as pong
	pong := new(messages.Pong)
	err := json.Unmarshal(msg, pong)
	if err != nil {
		// Call OnReadError - failed to parse message as pong
		eerr := fmt.Errorf("failed to parse message '%s' as pong: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if pong has a request ID.
	if pong.ReqId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received pong message has no request id")
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraceError(span, err)
	}
	// Tracing: Add event for received pong
	span.AddEvent(tracing.TracesNamespace+".pong", trace.WithAttributes(
		attribute.Int64("request_id", *pong.ReqId),
		attribute.String("session_id", sessionId),
	))
	// Extract pending ping request corresponding to the request ID
	pr := client.requests.pendingPing[*pong.ReqId]
	if pr == nil {
		// Call OnRead error: as user defined request ids must be used. Not a corresponding
		// pending request is considered as an error
		err := fmt.Errorf("received pong has no corresponding pending ping request for id: %d", *pong.ReqId)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraceError(span, err)
	}
	// Fulfil pending request
	// Blocking write can be used as channel must always have a capacity of one and be internally managed
	pr.resp <- pong
	// Discard pending request now that it has been served and exit
	delete(client.requests.pendingPing, *pong.ReqId)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received subscriptionStatus message as a response for
// either Subscribe or Unsubscribe.
func (client *KrakenSpotPublicWebsocketClient) handleSubscriptionStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_subscription_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as SubscriptionStatus
	subs := new(messages.SubscriptionStatus)
	err := json.Unmarshal(msg, subs)
	if err != nil {
		// Call OnReadError - failed to parse message as SubscriptionStatus
		eerr := fmt.Errorf("failed to parse message '%s' as subscriptionStatus: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if there is a request ID.
	if subs.ReqId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received subscriptionStatus message has no request id")
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraceError(span, err)
	}
	// Tracing: Add event for received subscriptionStatus
	attr := []attribute.KeyValue{
		attribute.String("session_id", sessionId),
		attribute.Int64("request_id", *subs.ReqId),
		attribute.String("status", subs.Status),
		attribute.String("pair", subs.Pair),
	}
	if subs.Err == "" {
		attr = append(attr, attribute.String("error_message", subs.Err))
	} else {
		attr = append(attr, attribute.String("channel", subs.ChannelName))
	}
	if subs.Subscription != nil {
		attr = append(attr, attribute.String("topic", subs.Subscription.Name))
		switch strings.Split(subs.ChannelName, "-")[0] {
		case string(messages.ChannelOpenOrders):
			attr = append(attr, attribute.Int("max_rate_count", subs.Subscription.MaxRateCount))
		case string(messages.ChannelOHLC):
			attr = append(attr, attribute.Int("interval", subs.Subscription.Interval))
		case string(messages.ChannelBook):
			attr = append(attr, attribute.Int("depth", subs.Subscription.Depth))
		}
	}
	span.AddEvent(tracing.TracesNamespace+".subscription_status", trace.WithAttributes(attr...))
	// Extract pending subscribe request corresponding to the request ID
	subreq := client.requests.pendingSubscribe[*subs.ReqId]
	if subreq == nil {
		// Check unsubscribe
		unsubreq := client.requests.pendingUnsubscribe[*subs.ReqId]
		if unsubreq == nil {
			// Call OnRead error: as user defined request ids must be used. Not a corresponding
			// pending request is considered as an error
			err := fmt.Errorf("received suscriptionStatus has no corresponding pending request for id: %d", *subs.ReqId)
			client.OnReadError(ctx, conn, readMutex, restart, exit, err)
			return tracing.HandleAndTraceError(span, err)
		}
		// Fulfil pending unsubscribe: send nil in case of success or an error with the error message if
		// unsubscribe has failed.
		err = nil
		if subs.Status == string(messages.Err) {
			err = fmt.Errorf("unsubscribe failed: %s", subs.Err)
		}
		// Blocking write can be used as channel must always have a capacity of one and be internally managed
		unsubreq.err <- err
		// Discard pending request
		delete(client.requests.pendingUnsubscribe, *subs.ReqId)
	} else {
		// Fulfil pending subscribe: send nil in case of success or an error witht he error message if
		// unsubscribe has failed
		err = nil
		if subs.Status == string(messages.Err) {
			err = fmt.Errorf("subscribe failed: %s", subs.Err)
		}
		// Blocking write can be used as channel must always have a capacity of one and be internally managed
		subreq.err <- err
		// Discard pending request
		delete(client.requests.pendingSubscribe, *subs.ReqId)
	}
	// Exit
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received ticker message.
func (client *KrakenSpotPublicWebsocketClient) handleTicker(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_ticker",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as Ticker
	ticker := new(messages.Ticker)
	err := json.Unmarshal(msg, ticker)
	if err != nil {
		// Call OnReadError - failed to parse message as ticker
		eerr := fmt.Errorf("failed to parse message '%s' as ticker: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if there is an active subscription, discard otherwise
	if client.subscriptions.ticker == nil {
		err := fmt.Errorf("a ticker message has been received while there is no active subscription to ticker channel")
		return tracing.HandleAndTraceError(span, err)
	}
	// Publish ticker - use blocking write (block until delivery)
	client.subscriptions.ticker.pub <- ticker
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received ohlc message.
func (client *KrakenSpotPublicWebsocketClient) handleOHLC(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_ohlc",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as OHLC
	ohlc := new(messages.OHLC)
	err := json.Unmarshal(msg, ohlc)
	if err != nil {
		// Call OnReadError - failed to parse message as ohlc
		eerr := fmt.Errorf("failed to parse message '%s' as ohlc: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if there is an active subscription, discard otherwise
	if client.subscriptions.ohlcs == nil {
		err := fmt.Errorf("a ohlc message has been received while there is no active subscription to ohlc channel")
		return tracing.HandleAndTraceError(span, err)
	}
	// Publish ohlc - use blocking write (block until delivery)
	client.subscriptions.ohlcs.pub <- ohlc
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received trade message.
func (client *KrakenSpotPublicWebsocketClient) handleTrade(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_trade",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as Trade
	trade := new(messages.Trade)
	err := json.Unmarshal(msg, trade)
	if err != nil {
		// Call OnReadError - failed to parse message as trade
		eerr := fmt.Errorf("failed to parse message '%s' as trade: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if there is an active subscription, discard otherwise
	if client.subscriptions.trade == nil {
		err := fmt.Errorf("a trade message has been received while there is no active subscription to trade channel")
		return tracing.HandleAndTraceError(span, err)
	}
	// Publish trade - use blocking write (block until delivery)
	client.subscriptions.trade.pub <- trade
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received spread message.
func (client *KrakenSpotPublicWebsocketClient) handleSpread(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_spread",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as Spread
	spread := new(messages.Spread)
	err := json.Unmarshal(msg, spread)
	if err != nil {
		// Call OnReadError - failed to parse message as spread
		eerr := fmt.Errorf("failed to parse message '%s' as spread: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if there is an active subscription, discard otherwise
	if client.subscriptions.spread == nil {
		err := fmt.Errorf("a spread message has been received while there is no active subscription to spread channel")
		return tracing.HandleAndTraceError(span, err)
	}
	// Publish trade - use blocking write
	client.subscriptions.spread.pub <- spread
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received book message.
func (client *KrakenSpotPublicWebsocketClient) handleBook(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_book",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	// Check if it is a snapshot or an update -> an update will have a "c" field
	if strings.Contains(string(msg), `"c"`) {
		// Handle update
		return client.handleBookUpdate(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	}
	// Hanlde snapshot
	return client.handleBookSnapshot(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
}

// This method contains the logic to handle a received book update message.
func (client *KrakenSpotPublicWebsocketClient) handleBookUpdate(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_book_update",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as BookUpdate
	bu := new(messages.BookUpdate)
	err := json.Unmarshal(msg, bu)
	if err != nil {
		// Call OnReadError - failed to parse message as book update
		eerr := fmt.Errorf("failed to parse message '%s' as book update: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if there is an active subscription, discard otherwise
	if client.subscriptions.book == nil {
		err := fmt.Errorf("a book update message has been received while there is no active subscription to book channel")
		return tracing.HandleAndTraceError(span, err)
	}
	// Publish book update - use blocking write
	client.subscriptions.book.updates <- bu
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received book snapshot message.
func (client *KrakenSpotPublicWebsocketClient) handleBookSnapshot(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".handle_book_snapshot",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	// Parse message as BookSnapshot
	bs := new(messages.BookSnapshot)
	err := json.Unmarshal(msg, bs)
	if err != nil {
		// Call OnReadError - failed to parse message as book snapshot
		eerr := fmt.Errorf("failed to parse message '%s' as book snapshot: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraceError(span, eerr)
	}
	// Check if there is an active subscription, discard otherwise
	if client.subscriptions.book == nil {
		err := fmt.Errorf("a book snapshot message has been received while there is no active subscription to book channel")
		return tracing.HandleAndTraceError(span, err)
	}
	// Publish book snapshot - use non-blocking write
	client.subscriptions.book.snapshots <- bs
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

/*************************************************************************************************/
/* KRAKEN PUBLIC WEBSOCKET IMPL.                                                                 */
/*************************************************************************************************/

// # Description
//
// Send a ping to the websocket server and wait until a Pong response is received from the
// server or until an error or a timeout occurs.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//
// # Return
//
// Nil in case of success. Otherwise, an error is returned when:
//
//   - An error occurs when sending the message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server (OperationError).
func (client *KrakenSpotPublicWebsocketClient) Ping(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".ping", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Create response channels
	errChan := make(chan error, 1)
	respChan := make(chan *messages.Pong, 1)
	// Send ping message to server
	err := client.sendPingRequest(
		ctx,
		&messages.Ping{
			Event: string(messages.EventTypePing),
			ReqId: client.ngen.GenerateNonce(),
		},
		respChan,
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("ping failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("ping failed: %w", err))
	case <-errChan:
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("ping failed: %w", err))
	case <-respChan:
		// Set span status and exit
		span.AddEvent(tracing.TracesNamespace + ".pong_received")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the ticker channel.
//
// In case of success, a channel with the desired capacity is created and returned. The channel
// will be used to publish subscription's data.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//   - capacity: Desired channel capacity. Can be 0 (not recommended).
//
// # Return
//
// In case of success, a channel with the desired capacity will be returned. Received data will
// be published on that channel.
//
// An error (and no channel) is returned when:
//
//   - A subscription is already active.
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active susbscription.
//
//   - A nil value MUST be published on the channel ONLY when the websocket connection is closed
//     even if the client implementation has a mechanism to automatically reconnect to the
//     websocket server. This nil value will serve as a cue for the consumer to detect
//     interruptions in the stream of data and react to these interruptions.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the publish channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server AND
//     resubscribe to previously subscribed channels, then, the client implementation MUST reuse
//     the channel that has been previously created and returned to the user.
//
//   - The client MUST drop the channel if the user has used the corresponding Unsubscribe method.
//     If the user use the subscribe method again, then, a new channel MUST be created and the
//     older one MUST NOT be used anymore.
func (client *KrakenSpotPublicWebsocketClient) SubscribeTicker(ctx context.Context, pairs []string, capacity int) (chan *messages.Ticker, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".subscribe_ticker",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("capacity", capacity),
		))
	defer span.End()
	// Check if there is already an active subscription
	client.tickerSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.tickerSubMu.Unlock()
	if client.subscriptions.ticker != nil {
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ticker failed because there is already an active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send subscribe message to server
	err := client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: pairs,
			Subscription: messages.SuscribeDetails{
				Name: string(messages.ChannelTicker),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ticker failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ticker failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ticker failed: %w", err))
		}
		// Register the subscription
		client.subscriptions.ticker = &tickerSubscription{
			pairs: pairs,
			pub:   make(chan *messages.Ticker, capacity),
		}
		// Return publish channel
		return client.subscriptions.ticker.pub, nil
	}
}

// # Description
//
// Subscribe to the ohlc channel.
//
// In case of success, a channel with the desired capacity is created and returned. The channel
// will be used to publish subscription's data.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//   - interval: The desired interval for OHLC indicators. Multiple subscriptions can be
//     maintained for different intervals.
//   - capacity: Desired channel capacity. Can be 0 (not recommended).
//
// # Return
//
// In case of success, a channel with the desired capacity will be returned. Received data will
// be published on that channel.
//
// An error (and no channel) is returned when:
//
//   - A subscription is already active.
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active susbscription.
//
//   - A nil value MUST be published on the channel ONLY when the websocket connection is closed
//     even if the client implementation has a mechanism to automatically reconnect to the
//     websocket server. This nil value will serve as a cue for the consumer to detect
//     interruptions in the stream of data and react to these interruptions.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the publish channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server AND
//     resubscribe to previously subscribed channels, then, the client implementation MUST reuse
//     the channel that has been previously created and returned to the user.
//
//   - The client MUST drop the channel if the user has used the corresponding Unsubscribe method.
//     If the user use the subscribe method again, then, a new channel MUST be created and the
//     older one MUST NOT be used anymore.
func (client *KrakenSpotPublicWebsocketClient) SubscribeOHLC(ctx context.Context, pairs []string, interval messages.IntervalEnum, capacity int) (chan *messages.OHLC, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".subscribe_ohlc",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("interval", int(interval)),
			attribute.Int("capacity", capacity),
		))
	defer span.End()
	// Check if there is already an active subscription
	client.ohlcSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.ohlcSubMu.Unlock()
	if client.subscriptions.ohlcs != nil {
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ohlc failed because there is already an active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send subscribe message to server
	err := client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: pairs,
			Subscription: messages.SuscribeDetails{
				Name:     string(messages.ChannelOHLC),
				Interval: int(interval),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ohlc failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ohlc failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe ohlc failed: %w", err))
		}
		// Register the subscription
		client.subscriptions.ohlcs = &ohlcSubscription{
			pairs:    pairs,
			pub:      make(chan *messages.OHLC, capacity),
			interval: interval,
		}
		// Return publish channel
		return client.subscriptions.ohlcs.pub, nil
	}
}

// # Description
//
// Subscribe to the trade channel.
//
// In case of success, a channel with the desired capacity is created and returned. The channel
// will be used to publish subscription's data.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//   - capacity: Desired channel capacity. Can be 0 (not recommended).
//
// # Return
//
// In case of success, a channel with the desired capacity will be returned. Received data will
// be published on that channel.
//
// An error (and no channel) is returned when:
//
//   - A subscription is already active.
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active susbscription.
//
//   - A nil value MUST be published on the channel ONLY when the websocket connection is closed
//     even if the client implementation has a mechanism to automatically reconnect to the
//     websocket server. This nil value will serve as a cue for the consumer to detect
//     interruptions in the stream of data and react to these interruptions.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the publish channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server AND
//     resubscribe to previously subscribed channels, then, the client implementation MUST reuse
//     the channel that has been previously created and returned to the user.
//
//   - The client MUST drop the channel if the user has used the corresponding Unsubscribe method.
//     If the user use the subscribe method again, then, a new channel MUST be created and the
//     older one MUST NOT be used anymore.
func (client *KrakenSpotPublicWebsocketClient) SubscribeTrade(ctx context.Context, pairs []string, capacity int) (chan *messages.Trade, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".subscribe_trade",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("capacity", capacity),
		))
	defer span.End()
	// Check if there is already an active subscription
	client.tradeSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.tradeSubMu.Unlock()
	if client.subscriptions.trade != nil {
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe trade failed because there is already an active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send subscribe message to server
	err := client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: pairs,
			Subscription: messages.SuscribeDetails{
				Name: string(messages.ChannelTrade),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe trade failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe trade failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe trade failed: %w", err))
		}
		// Register the subscription
		client.subscriptions.trade = &tradeSubscription{
			pairs: pairs,
			pub:   make(chan *messages.Trade, capacity),
		}
		// Return publish channel
		return client.subscriptions.trade.pub, nil
	}
}

// # Description
//
// Subscribe to the spread channel.
//
// In case of success, a channel with the desired capacity is created and returned. The channel
// will be used to publish subscription's data.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//   - capacity: Desired channel capacity. Can be 0 (not recommended).
//
// # Return
//
// In case of success, a channel with the desired capacity will be returned. Received data will
// be published on that channel.
//
// An error (and no channel) is returned when:
//
//   - A subscription is already active.
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active susbscription.
//
//   - A nil value MUST be published on the channel ONLY when the websocket connection is closed
//     even if the client implementation has a mechanism to automatically reconnect to the
//     websocket server. This nil value will serve as a cue for the consumer to detect
//     interruptions in the stream of data and react to these interruptions.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the publish channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server AND
//     resubscribe to previously subscribed channels, then, the client implementation MUST reuse
//     the channel that has been previously created and returned to the user.
//
//   - The client MUST drop the channel if the user has used the corresponding Unsubscribe method.
//     If the user use the subscribe method again, then, a new channel MUST be created and the
//     older one MUST NOT be used anymore.
func (client *KrakenSpotPublicWebsocketClient) SubscribeSpread(ctx context.Context, pairs []string, capacity int) (chan *messages.Spread, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".subscribe_spread",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("capacity", capacity),
		))
	defer span.End()
	// Check if there is already an active subscription
	client.spreadSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.spreadSubMu.Unlock()
	if client.subscriptions.spread != nil {
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe spread failed because there is already an active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send subscribe message to server
	err := client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: pairs,
			Subscription: messages.SuscribeDetails{
				Name: string(messages.ChannelSpread),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe spread failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe spread failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe spread failed: %w", err))
		}
		// Register the subscription
		client.subscriptions.spread = &spreadSubscription{
			pairs: pairs,
			pub:   make(chan *messages.Spread, capacity),
		}
		// Return publish channel
		return client.subscriptions.spread.pub, nil
	}
}

// # Description
//
// Subscribe to the book channel.
//
// In case of success, a channel with the desired capacity is created and returned. The channel
// will be used to publish subscription's data.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//   - depth: Desired book depth. Multiple subscriptions can be maintained for different depths.
//   - capacity: Desired channel capacity. Can be 0 (not recommended).
//
// # Return
//
// In case of success, a channel with the desired capacity will be returned. Received data will
// be published on that channel.
//
// An error (and no channel) is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active susbscription.
//
//   - A nil value MUST be published on the channel ONLY when the websocket connection is closed
//     even if the client implementation has a mechanism to automatically reconnect to the
//     websocket server. This nil value will serve as a cue for the consumer to detect
//     interruptions in the stream of data and react to these interruptions.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the publish channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server AND
//     resubscribe to previously subscribed channels, then, the client implementation MUST reuse
//     the channel that has been previously created and returned to the user.
//
//   - The client MUST drop the channel if the user has used the corresponding Unsubscribe method.
//     If the user use the subscribe method again, then, a new channel MUST be created and the
//     older one MUST NOT be used anymore.
func (client *KrakenSpotPublicWebsocketClient) SubscribeBook(ctx context.Context, pairs []string, depth messages.DepthEnum, capacity int) (chan *messages.BookSnapshot, chan *messages.BookUpdate, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".subscribe_book",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("depth", int(depth)),
			attribute.Int("capacity", capacity),
		))
	defer span.End()
	// Check if there is already an active subscription
	client.bookSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book != nil {
		return nil, nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe book failed because there is already an active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send subscribe message to server
	err := client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: pairs,
			Subscription: messages.SuscribeDetails{
				Name:  string(messages.ChannelBook),
				Depth: int(depth),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return nil, nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe book failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe book failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return nil, nil, tracing.HandleAndTraceError(span, fmt.Errorf("subscribe book failed: %w", err))
		}
		// Register the subscription
		client.subscriptions.book = &bookSubscription{
			pairs:     pairs,
			updates:   make(chan *messages.BookUpdate, capacity),
			snapshots: make(chan *messages.BookSnapshot, len(pairs)*10),
			depth:     depth,
		}
		// Return publish channel
		return client.subscriptions.book.snapshots, client.subscriptions.book.updates, nil
	}
}

// # Description
//
// Unsubscribe from  the ticker topic. The previously used channel will be dropped and user
// must stop using it.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//
// # Return
//
// Nil in case of success. Otherwise, an error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server.
//
// # Implementation and usage guidelines
//
//   - The client MUST drop the channel that was used by the canceled subscription.
//
//   - The client MUST return an error if channel was not subscribed to.
func (client *KrakenSpotPublicWebsocketClient) UnsubscribeTicker(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".unsubscribe_ticker", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Check if there is already an active subscription
	client.tickerSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.tickerSubMu.Unlock()
	if client.subscriptions.ticker == nil {
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ticker failed because there is no active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err := client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeUnsubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: client.subscriptions.ticker.pairs,
			Subscription: messages.UnsuscribeDetails{
				Name: string(messages.ChannelTicker),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ticker failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ticker failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ticker failed: %w", err))
		}
		// Discard the subscription and exit
		client.subscriptions.ticker = nil
		return nil
	}
}

// # Description
//
// Unsubscribe from  the ohlc topic. The previously used channel will be dropped and user
// must stop using it.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done
//     channel will be watched for timeout/cancel signal.
//
// # Return
//
// Nil in case of success. Otherwise, an error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server.
//
// # Implementation and usage guidelines
//
//   - The client MUST drop the channel that was used by the canceled subscription.
//
//   - The client MUST return an error if channel was not subscribed to.
func (client *KrakenSpotPublicWebsocketClient) UnsubscribeOHLC(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".unsubscribe_ohlc", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Check if there is already an active subscription
	client.ohlcSubMu.Lock() // Lock mutex till unsubscribe completes - this will block Subscribe
	defer client.ohlcSubMu.Unlock()
	if client.subscriptions.ohlcs == nil {
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ohlc failed because there is no active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err := client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: client.subscriptions.ohlcs.pairs,
			Subscription: messages.UnsuscribeDetails{
				Name:     string(messages.ChannelOHLC),
				Interval: int(client.subscriptions.ohlcs.interval),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ohlc failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ohlc failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe ohlc failed: %w", err))
		}
		// Discard the subscription and exit
		client.subscriptions.ohlcs = nil
		return nil
	}
}

// # Description
//
// Unsubscribe from  the trade topic. The previously used channel will be dropped and user
// must stop using it.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//
// # Return
//
// Nil in case of success. Otherwise, an error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server.
//
// # Implementation and usage guidelines
//
//   - The client MUST drop the channel that was used by the canceled subscription.
//
//   - The client MUST return an error if channel was not subscribed to.
func (client *KrakenSpotPublicWebsocketClient) UnsubscribeTrade(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".unsubscribe_trade", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Check if there is already an active subscription
	client.tradeSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.tradeSubMu.Unlock()
	if client.subscriptions.trade == nil {
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe trade failed because there is no active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err := client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeUnsubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: client.subscriptions.trade.pairs,
			Subscription: messages.UnsuscribeDetails{
				Name: string(messages.ChannelTrade),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe trade failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe trade failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe trade failed: %w", err))
		}
		// Discard the subscription and exit
		client.subscriptions.trade = nil
		return nil
	}
}

// # Description
//
// Unsubscribe from  the spread topic. The previously used channel will be dropped and user
// must stop using it.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//
// # Return
//
// Nil in case of success. Otherwise, an error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server.
//
// # Implementation and usage guidelines
//
//   - The client MUST drop the channel that was used by the canceled subscription.
//
//   - The client MUST return an error if channel was not subscribed to.
func (client *KrakenSpotPublicWebsocketClient) UnsubscribeSpread(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".unsubscribe_spread", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Check if there is already an active subscription
	client.spreadSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.spreadSubMu.Unlock()
	if client.subscriptions.spread == nil {
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe spread failed because there is no active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err := client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeUnsubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: client.subscriptions.spread.pairs,
			Subscription: messages.UnsuscribeDetails{
				Name: string(messages.ChannelSpread),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe spread failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe spread failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe spread failed: %w", err))
		}
		// Discard the subscription and exit
		client.subscriptions.spread = nil
		return nil
	}
}

// # Description
//
// Unsubscribe from  the book topic. The previously used channel will be dropped and user
// must stop using it.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//
// # Return
//
// Nil in case of success. Otherwise, an error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the message.
//   - The provided context expires (timeout/cancel).
//   - An error message is received from the server.
//
// # Implementation and usage guidelines
//
//   - The client MUST drop the channel that was used by the canceled subscription.
//
//   - The client MUST return an error if channel was not subscribed to.
func (client *KrakenSpotPublicWebsocketClient) UnsubscribeBook(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".unsubscribe_book", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Check if there is already an active subscription
	client.bookSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book == nil {
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe book failed because there is no active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err := client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeUnsubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: client.subscriptions.book.pairs,
			Subscription: messages.UnsuscribeDetails{
				Name:  string(messages.ChannelBook),
				Depth: int(client.subscriptions.book.depth),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe book failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe book failed: %w", err))
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraceError(span, fmt.Errorf("unsubscribe book failed: %w", err))
		}
		// Discard the subscription and exit
		client.subscriptions.book = nil
		return nil
	}
}

// # Description
//
// Get the client's built-in channel to publish received system status updates.
//
// # Implemetation and usage guidelines
//
//   - As the channel is automatically subscribed to, the client implementation CAN discard messages
//     in case of congestion in the publication channel. The client implementation must be clear
//     about how it deals with congestion.
//
// # Return
//
// The client's built-in channel used to publish received system status updates.
func (client *KrakenSpotPublicWebsocketClient) GetSystemStatusChannel() chan *messages.SystemStatus {
	return client.subscriptions.systemStatus
}

// # Description
//
// Get the client's built-in channel to publish received heartbeats.
//
// # Implemetation and usage guidelines
//
//   - As the channel is automatically subscribed to, the client implementation CAN discard messages
//     in case of congestion in the publication channel. The client implementation must be clear
//     about how it deals with congestion.
//
// # Return
//
// The client's built-in channel used to publish received heartbeats.
func (client *KrakenSpotPublicWebsocketClient) GetHeartbeatChannel() chan *messages.Heartbeat {
	return client.subscriptions.heartbeat
}

/*************************************************************************************************/
/* PRIVATE METHODS                                                                               */
/*************************************************************************************************/

// # Description
//
// Send a ping request to the websocket server. The method will add a pending ping request to the
// client's pending requests stack.
//
// The method returns an error if it fails to send the request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. Will be provided to the pending request.
//   - req: Ping request to send. Must not be nil
//   - respChan: Channel used to publish pong response to requestor. Must have a capacity of 1 to not block the engine.
//   - errChan: Channel provided to the pending request. Will be used to publish the subscription results.  Must have a capacity of 1 to not block the engine.
//
// # Return
//
// An error if the request cannot be sent.
func (client *KrakenSpotPublicWebsocketClient) sendPingRequest(ctx context.Context, req *messages.Ping, respChan chan *messages.Pong, errChan chan error) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".send_ping_request",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.Int64("request_id", req.ReqId)))
	defer span.End()
	// Add pending ping request to client's stack
	client.mu.Lock() // Lock to not add requests while engine is discarding pending requests
	client.requests.pendingPing[req.ReqId] = &pendingPing{
		resp: respChan,
		err:  errChan,
	}
	client.mu.Lock() // Immediatly unlock to not block the engine while waiting for the response
	// Marshal to JSON
	payload, err := json.Marshal(req)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingPing, req.ReqId)
		return tracing.HandleAndTraceError(span, fmt.Errorf("failed to format ping request: %w", err))
	}
	// Send message to websocket server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingSubscribe, req.ReqId)
		return tracing.HandleAndTraceError(span, fmt.Errorf("failed to send ping request: %w", err))
	}
	// Set span status and exit
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// # Description
//
// Send a subscribe request to the websocket server. The method will add a pending subscribe request
// to the client's pending requests stack.
//
// The method returns an error if it fails to send the request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. Will be provided to the pending request.
//   - req: Subscribe request to send. Must not be nil
//   - errChan: Channel provided to the pending request. Will be used to publish the results.
//
// # Return
//
// An error if the request cannot be sent.
func (client *KrakenSpotPublicWebsocketClient) sendSubscribeRequest(ctx context.Context, req *messages.Subscribe, errChan chan error) error {
	// Tracing: Prepare span attributes
	reqAttr := []attribute.KeyValue{
		attribute.String("type", req.Subscription.Name),
		attribute.Int64("request_id", req.ReqId),
		attribute.StringSlice("pairs", req.Pairs),
	}
	// Tracing: Add specific attribute depending on the subscribed channel
	switch req.Subscription.Name {
	case string(messages.ChannelOwnTrades):
		if req.Subscription.Snapshot != nil && req.Subscription.ConsolidateTaker != nil {
			reqAttr = append(
				reqAttr,
				attribute.Bool("snapshot", *req.Subscription.Snapshot),
				attribute.Bool("consolidate_taker", *req.Subscription.ConsolidateTaker),
			)
		}
	case string(messages.ChannelOpenOrders):
		reqAttr = append(reqAttr, attribute.Bool("rate_counter", req.Subscription.RateCounter))
	case string(messages.ChannelOHLC):
		reqAttr = append(reqAttr, attribute.Int("interval", req.Subscription.Interval))
	case string(messages.ChannelBook):
		reqAttr = append(reqAttr, attribute.Int("depth", req.Subscription.Depth))
	}
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".send_subscribe_request",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttr...))
	defer span.End()
	// Add pending susbcribe request to client's stack
	client.mu.Lock() // Lock to not add requests while engine is discarding pending requests
	client.requests.pendingSubscribe[req.ReqId] = &pendingSubscribe{
		err: errChan,
	}
	client.mu.Lock() // Immediatly unlock to not block the engine while waiting for the response
	// Marshal to JSON
	payload, err := json.Marshal(req)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingSubscribe, req.ReqId)
		return tracing.HandleAndTraceError(span, fmt.Errorf("failed to format subscribe request: %w", err))
	}
	// Send message to websocket server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingSubscribe, req.ReqId)
		return tracing.HandleAndTraceError(span, fmt.Errorf("failed to send subscribe request: %w", err))
	}
	// Set span status and exit
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// # Description
//
// Send a unsubscribe request to the websocket server. The method will add a pending unsubscribe
// request to the client's pending requests stack.
//
// The method returns an error if it fails to send the request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. Will be provided to the pending request.
//   - req: Unsubscribe request to send. Must not be nil
//   - errChan: Channel provided to the pending request. Will be used to publish the results.
//
// # Return
//
// An error if the request cannot be sent.
func (client *KrakenSpotPublicWebsocketClient) sendUnsubscribeRequest(ctx context.Context, req *messages.Unsubscribe, errChan chan error) error {
	// Tracing: Prepare span attributes
	reqAttr := []attribute.KeyValue{
		attribute.String("type", req.Subscription.Name),
		attribute.Int64("request_id", req.ReqId),
		attribute.StringSlice("pairs", req.Pairs),
	}
	// Tracing: Add specific attribute depending on the unsubscribed channel
	switch req.Subscription.Name {
	case string(messages.ChannelOHLC):
		reqAttr = append(reqAttr, attribute.Int("interval", req.Subscription.Interval))
	case string(messages.ChannelBook):
		reqAttr = append(reqAttr, attribute.Int("depth", req.Subscription.Depth))
	}
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, tracing.TracesNamespace+".send_unsubscribe_request",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttr...))
	defer span.End()
	// Add pending unsusbcribe request to client's stack
	client.mu.Lock() // Lock to not add requests while engine is discarding pending requests
	client.requests.pendingUnsubscribe[req.ReqId] = &pendingUnsubscribe{
		err: errChan,
	}
	client.mu.Lock() // Immediatly unlock to not block the engine while waiting for the response
	// Marshal to JSON
	payload, err := json.Marshal(req)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingUnsubscribe, req.ReqId)
		return tracing.HandleAndTraceError(span, fmt.Errorf("failed to format unsubscribe request: %w", err))
	}
	// Send message to websocket server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingUnsubscribe, req.ReqId)
		return tracing.HandleAndTraceError(span, fmt.Errorf("failed to send unsubscribe request: %w", err))
	}
	// Set span status and exit
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}
