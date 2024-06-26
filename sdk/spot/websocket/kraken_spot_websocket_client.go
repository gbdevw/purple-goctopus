package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gbdevw/gowse/wscengine/wsadapters"
	"github.com/gbdevw/gowse/wscengine/wsclient"
	"github.com/gbdevw/purple-goctopus/sdk/noncegen"
	"github.com/gbdevw/purple-goctopus/sdk/spot/rest"
	restcommon "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/events"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// URL for Kraken spot websocket client - public endpoints - Production
	KrakenSpotWebsocketPublicProductionURL = "wss://ws.kraken.com"
	// URL for Kraken spot websocket client - public endpoints - Beta
	KrakenSpotWebsocketPublicBetaURL = "wss://beta-ws.kraken.com"
	// URL for Kraken spot websocket client - private endpoints - Production
	KrakenSpotWebsocketPrivateProductionURL = "wss://ws-auth.kraken.com"
	// URL for Kraken spot websocket client - private endpoints - Beta
	KrakenSpotWebsocketPrivateBetaURL = "wss://beta-ws-auth.kraken.com"
)

// This is the base Kraken websocket client implementation: The logic is the same for both public
// and private clients but separate clients must be built because public and private clients do
// not use the same servers and connection.
//
// Principles:
//   - Blocking writes are used to publish received messages from the websocket server
//   - For heartbeats and system status updates, overflowing messages are discarded in FIFO order.
type krakenSpotWebsocketClient struct {
	// Websocket connection adapter to use to interact with the chosen
	// underlying low-level websocket framework.
	conn wsadapters.WebsocketConnectionAdapterInterface
	// Internal nonce generator used to generate unique request IDs
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
	// Logger used to publish debug/verbose logs
	logger *log.Logger
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
	// Mutex used to protect open orders subscribe/unsubscribe methods
	openOrdersSubMu sync.Mutex
	// Mutex used to protect own trades subscribe/unsubscribe methods
	ownTradesSubMu sync.Mutex
	// Mutex used to protect pending ping request map from concurrent writes
	pendingPingMu sync.Mutex
	// Mutex used to protect pending subscribe request map from concurrent writes
	pendingSubscribeMu sync.Mutex
	// Mutex used to protect pending unsubscribe request map from concurrent writes
	pendingUnsubscribeMu sync.Mutex
	// Mutex used to protect pending addOrder request map from concurrent writes
	pendingAddOrderMu sync.Mutex
	// Mutex used to protect pending editOrder request map from concurrent writes
	pendingEditOrderMu sync.Mutex
	// Mutex used to protect pending cancelOrder request map from concurrent writes
	pendingCancelOrderMu sync.Mutex
	// Mutex used to protect pending cancelAllOrders request map from concurrent writes
	pendingCancelAllOrdersMu sync.Mutex
	// Mutex used to protect pending cancelAllOrdersAfterX request map from concurrent writes
	pendingCancelAllOrdersAfterXOrderMu sync.Mutex
	// Kraken websocket client used to get websocket token
	restClient rest.KrakenSpotRESTClientIface
	// User provided nonce generator used to generate nonces used when GetWebsocketToken is called
	cgen noncegen.NonceGenerator
	// User provided security options used when
	secopts *restcommon.SecurityOptions
	// Mutex used to protect cached websocket token
	tokenMu sync.Mutex
	// Cached websocket token
	token string
	// Cached websocket token epiration time
	tokenExpiresAt time.Time
}

// # Description
//
// Build a new krakenSpotWebsocketClient. The client must be provided to a wscengine.WebsocketEngine
// which will manage the connection with the server and execute the client's logic. Once connected, the
// client can be used to send messages to the websocket server, subscribe to data feeds, ...
//
// # Inputs
//
//   - restClient: Optional Kraken spot rest client to use to get a websocket token. Can be nil in case only public endpoints are used.
//   - clientNonceGenerator: Optional nonce generator used to get nonces used to sign requests sent with the REST Client. Can be nil in case only public endpoints are used.
//   - secopts: Optional security options (like password 2FA) to use when sending requests with the REST client.
//   - onCloseCallback: optional user defined callback which will be called when connection is closed/interrupted.
//   - onReadErrorCallback: optional user defined callback which will be called when an error occurs while reading messages from the websocket server
//   - onRestartError: optional user defined callback which will be called when the websocket engine fails to reconnect to the server.
//   - logger: Optional logger used to log debug/vebrose messages. If nil, a logger with a discard writer (noop) will be used
//   - tracerProvider: Tracer provider to use to get a tracer to instrument websocket client code. If nil, global tracer provider will be used.
//
// # Return
//
// A new krakenSpotWebsocketClient which can then be used by a wscengine.WebsocketEngine.
func newKrakenSpotWebsocketClient(
	restClient rest.KrakenSpotRESTClientIface,
	clientNonceGenerator noncegen.NonceGenerator,
	secopts *restcommon.SecurityOptions,
	onCloseCallback func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails),
	onReadErrorCallback func(ctx context.Context, restart context.CancelFunc, exit context.CancelFunc, err error),
	onRestartError func(ctx context.Context, exit context.CancelFunc, err error, retryCount int),
	logger *log.Logger,
	tracerProvider trace.TracerProvider,
) *krakenSpotWebsocketClient {
	// Create a discard logger if none is provided
	if logger == nil {
		logger = log.New(io.Discard, "", log.Default().Flags())
	}
	// Use the global tracer provider if none is provided
	if tracerProvider == nil {
		tracerProvider = otel.GetTracerProvider()
	}
	return &krakenSpotWebsocketClient{
		conn: nil,
		ngen: noncegen.NewHFNonceGenerator(),
		subscriptions: activeSubscriptions{
			heartbeat:    make(chan event.Event, 10),
			systemStatus: make(chan event.Event, 10),
			ohlcs:        make(map[messages.IntervalEnum]*ohlcSubscription),
		},
		requests: pendingRequests{
			pendingPing:                          map[int64]*pendingPing{},
			pendingSubscribe:                     map[int64]*pendingSubscribe{},
			pendingUnsubscribe:                   map[int64]*pendingUnsubscribe{},
			pendingAddOrderRequests:              map[int64]*pendingAddOrderRequest{},
			pendingEditOrderRequests:             map[int64]*pendingEditOrderRequest{},
			pendingCancelOrderRequests:           map[int64]*pendingCancelOrderRequest{},
			pendingCancelAllOrdersRequests:       map[int64]*pendingCancelAllOrdersRequest{},
			pendingCancelAllOrdersAfterXRequests: map[int64]*pendingCancelAllOrdersAfterXRequest{}},
		onCloseCallback:                     onCloseCallback,
		onReadErrorCallback:                 onReadErrorCallback,
		onRestartError:                      onRestartError,
		tracer:                              tracerProvider.Tracer(tracing.PackageName, trace.WithInstrumentationVersion(tracing.PackageVersion)),
		tickerSubMu:                         sync.Mutex{},
		ohlcSubMu:                           sync.Mutex{},
		tradeSubMu:                          sync.Mutex{},
		spreadSubMu:                         sync.Mutex{},
		bookSubMu:                           sync.Mutex{},
		openOrdersSubMu:                     sync.Mutex{},
		ownTradesSubMu:                      sync.Mutex{},
		pendingPingMu:                       sync.Mutex{},
		pendingSubscribeMu:                  sync.Mutex{},
		pendingUnsubscribeMu:                sync.Mutex{},
		pendingAddOrderMu:                   sync.Mutex{},
		pendingEditOrderMu:                  sync.Mutex{},
		pendingCancelOrderMu:                sync.Mutex{},
		pendingCancelAllOrdersMu:            sync.Mutex{},
		pendingCancelAllOrdersAfterXOrderMu: sync.Mutex{},
		logger:                              logger,
		restClient:                          restClient,
		cgen:                                clientNonceGenerator,
		secopts:                             secopts,
		tokenMu:                             sync.Mutex{},
		token:                               "", // Just to make it clear ;)
		tokenExpiresAt:                      time.Time{},
	}
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
//   - The provided context expires before pong is received (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) Ping(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "ping", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("sending ping to the server")
	// Create response channels
	errChan := make(chan error, 1)
	respChan := make(chan *messages.Pong, 1)
	// Send ping message to server
	req := &messages.Ping{
		Event: string(messages.EventTypePing),
		ReqId: client.ngen.GenerateNonce(),
	}
	// Lock pending ping request map and add request to the stack.
	client.pendingPingMu.Lock()
	client.requests.pendingPing[req.ReqId] = &pendingPing{
		resp: respChan,
		err:  errChan,
	}
	// Defer pending request map cleanup to remove it in case of failure or ensure it has been
	// removed in case of success. This is safe because pending requests ids are unique and
	// internally managed.
	defer delete(client.requests.pendingSubscribe, req.ReqId)
	// Defer unlocking pending request map.
	unlock := sync.OnceFunc(client.pendingPingMu.Unlock)
	defer unlock()
	// Marshal to JSON
	payload, err := json.Marshal(req)
	if err != nil {
		// Trace and return error -> failed to format request
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("failed to format ping request: %w", err))
	}
	// Send message to websocket server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Trace and return error -> failed to send request
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("failed to send ping request: %w", err))
	}
	// Unlock pending ping requests map so another goroutine can process the pong message and
	// fulfill the pending request. As the call is encapsulaated in a sync.Once, the deferred
	// unlock will be a noop.
	unlock()
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for pong from the server")
	select {
	case <-ctx.Done():
		// Trace and return error -> operation interrupted before completion.
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "ping", Root: fmt.Errorf("ping failed: %w", ctx.Err())})
	case err := <-errChan:
		// Trace and return error -> operation failed with an error from the server.
		return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "ping", Root: fmt.Errorf("ping failed: %w", err)})
	case <-respChan:
		// Set span status and exit
		client.logger.Println("pong received")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the tickers channel. In case of success, the websocket client will start
// publishing received events on the user's provided channel.
//
// # This client implementation uses
//
// Two types of events can be published on the channel:
//   - connection_interrupted: This event type is used when connection with the sevrer has been
//     interrupted. The event will not have any data. It only serves as a cue for the consumer
//     to allow the consumer to react when the connection with the server is interrupted.
//   - ticker: This event type is used when a message has been received from the server.
//     Published events will contain both the received data and the tracing context to continue
//     the tracing span from the source (= the websocket engine).
//
// In case when the connection with the server is lost, the websocket client will publish a
// connection_interrupted event to warn consumer about the failure.
//
// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
// when it reconnects to the server and it MUST reuse the previously provided channel to publish
// received messages.
//
// Consumers should always watch to the event type to separate messages from the connection
// failure events and react according the event type.
//
// Finally, the provided channel will be automatically closed by the client when:
//   - The user unsubscribe from the topic by using UnsubscribeTickers
//   - The websocket client definitely stops.
//
// Consumers should also watch channel closure to know when no more data will be delivered.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//   - connection_interrupted
//   - ticker
//
// # Extract data
//
// Before parsing the data, check the event type to catch rare connection_interrupted events.
//
// The event data contains the JSON payload from the server and can be parsed into a structure
// of type messages.Ticker like this:
//
//	ticker := new(messages.Ticker)
//	err := event.DataAs(ticker)
//
// The event will also contain the tracing context from OpenTelemetry. This tracing context can
// be extracted from the event to continue tracing the event processing from the source:
//
//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - pair: Pairs to subscribe to.
//   - rcv: Channel used to publish ticker messages and connection_interrupted events.
//
// # Return
//
// An error is returned when:
//
//   - There is already an active subscription.
//   - An error occurs when sending the subscription message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active subscription.
//
//   - The client MUST use the right error type as described in the "Return" section.
//
//   - A connection_interrupted event MUST be published on the channel each time the websocket
//     connection is closed.
//
//   - The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the provided channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server,
//     then the websocket client MUST resubscribe to previously subscribed channels and reuse
//     the channel that has been provided when the user subscribed to the channel.
func (client *krakenSpotWebsocketClient) SubscribeTicker(ctx context.Context, pairs []string, rcv chan event.Event) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "subscribe_ticker",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
		))
	defer span.End()
	client.logger.Println("subscribing to ticker channel", pairs)
	// Check if there is already an active subscription
	client.tickerSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.tickerSubMu.Unlock()
	if client.subscriptions.ticker != nil {
		// Trae and log error: already subscribed
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe ticker failed because there is already an active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe ticker failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for subscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error: operation interrupted before completion
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "suscribe_ticker", Root: fmt.Errorf("subscribe ticker failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "suscribe_ticker", Root: fmt.Errorf("subscribe ticker failed: %w", err)})
		}
		// Register the subscription and save the provided channel
		client.subscriptions.ticker = &tickerSubscription{
			pairs: pairs,
			pub:   rcv,
		}
		// Exit - success
		client.logger.Println("ticker channel subscribed")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the ohlc channel with the given interval. In case of success, the websocket
// client will start publishing received events on the user's provided channel.
//
// The client supports multiple subscriptions but for different interval. The client will return
// an error in case there's already a subscription for that interval.
//
// Two types of events can be published on the channel:
//   - connection_interrupted: This event type is used when connection with the sevrer has been
//     interrupted. The event will not have any data. It only serves as a cue for the consumer
//     to allow the consumer to react when the connection with the server is interrupted.
//   - ohlc: This event type is used when a message has been received from the server.
//     Published events will contain both the received data and the tracing context to continue
//     the tracing span from the source (= the websocket engine).
//
// In case when the connection with the server is lost, the websocket client will publish a
// connection_interrupted event to warn consumer about the failure.
//
// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
// when it reconnects to the server and it MUST reuse the previously provided channel to publish
// received messages.
//
// Consumers should always watch to the event type to separate messages from the connection
// failure events and react according the event type.
//
// Finally, the provided channel will be automatically closed by the client when:
//   - The user unsubscribe from the topic by using UnsubscribeOHLC
//   - The websocket client definitely stops.
//
// Consumers should also watch channel closure to know when no more data will be delivered.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//   - connection_interrupted
//   - ohlc
//
// # Extract data
//
// Before parsing the data, check the event type to catch rare connection_interrupted events.
//
// The event data contains the JSON payload from the server and can be parsed into a structure
// of type messages.OHLC like this:
//
//	ohlc := new(messages.OHLC)
//	err := event.DataAs(ohlc)
//
// The event will also contain the tracing context from OpenTelemetry. This tracing context can
// be extracted from the event to continue tracing the event processing from the source:
//
//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - pair: Pairs to subscribe to.
//   - interval: Interval for produced OHLC indicators.
//   - rcv: Channel used to publish ohlc messages and connection_interrupted events.
//
// # Return
//
// An error is returned when:
//
//   - There is already an active subscription for that interval.
//   - An error occurs when sending the subscription message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active subscription.
//
//   - The client MUST use the right error type as described in the "Return" section.
//
//   - A connection_interrupted event MUST be published on the channel each time the websocket
//     connection is closed.
//
//   - The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the provided channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server,
//     then the websocket client MUST resubscribe to previously subscribed channels and reuse
//     the channel that has been provided when the user subscribed to the channel.
func (client *krakenSpotWebsocketClient) SubscribeOHLC(ctx context.Context, pairs []string, interval messages.IntervalEnum, rcv chan event.Event) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "subscribe_ohlc",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("interval", int(interval)),
		))
	defer span.End()
	client.logger.Println("subscribing to ohlc channel", pairs, int(interval))
	// Check if there is already an active subscription
	client.ohlcSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.ohlcSubMu.Unlock()
	if client.subscriptions.ohlcs[interval] != nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe ohlc-%d failed because there is already an active subscription", int(interval)))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe ohlc-%d failed: %w", int(interval), err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for subscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "subscribe_ohlc", Root: fmt.Errorf("subscribe ohlc failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "subscribe_ohlc", Root: fmt.Errorf("subscribe ohlc failed: %w", err)})
		}
		// Register the subscription
		client.subscriptions.ohlcs[interval] = &ohlcSubscription{
			pairs:    pairs,
			pub:      rcv,
			interval: interval,
		}
		// Return publish channel
		client.logger.Println("ohlc channel subscribed")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the trades channel. In case of success, the websocket client will start
// publishing received events on the user's provided channel.
//
// Two types of events can be published on the channel:
//   - connection_interrupted: This event type is used when connection with the sevrer has been
//     interrupted. The event will not have any data. It only serves as a cue for the consumer
//     to allow the consumer to react when the connection with the server is interrupted.
//   - trade: This event type is used when a message has been received from the server.
//     Published events will contain both the received data and the tracing context to continue
//     the tracing span from the source (= the websocket engine).
//
// In case when the connection with the server is lost, the websocket client will publish a
// connection_interrupted event to warn consumer about the failure.
//
// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
// when it reconnects to the server and it MUST reuse the previously provided channel to publish
// received messages.
//
// Consumers should always watch to the event type to separate messages from the connection
// failure events and react according the event type.
//
// Finally, the provided channel will be automatically closed by the client when:
//   - The user unsubscribe from the topic by using UnsubscribeTrades
//   - The websocket client definitely stops.
//
// Consumers should also watch channel closure to know when no more data will be delivered.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//   - connection_interrupted
//   - trade
//
// # Extract data
//
// Before parsing the data, check the event type to catch rare connection_interrupted events.
//
// The event data contains the JSON payload from the server and can be parsed into a structure
// of type messages.Trade like this:
//
//	trade := new(messages.Trade)
//	err := event.DataAs(trade)
//
// The event will also contain the tracing context from OpenTelemetry. This tracing context can
// be extracted from the event to continue tracing the event processing from the source:
//
//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - pair: Pairs to subscribe to.
//   - rcv: Channel used to publish trade messages and connection_interrupted events.
//
// # Return
//
// An error is returned when:
//
//   - There is already an active subscription.
//   - An error occurs when sending the subscription message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active subscription.
//
//   - The client MUST use the right error type as described in the "Return" section.
//
//   - A connection_interrupted event MUST be published on the channel each time the websocket
//     connection is closed.
//
//   - The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the provided channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server,
//     then the websocket client MUST resubscribe to previously subscribed channels and reuse
//     the channel that has been provided when the user subscribed to the channel.
func (client *krakenSpotWebsocketClient) SubscribeTrade(ctx context.Context, pairs []string, rcv chan event.Event) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "subscribe_trade",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
		))
	defer span.End()
	client.logger.Println("subscribing to trade channel", pairs)
	// Check if there is already an active subscription
	client.tradeSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.tradeSubMu.Unlock()
	if client.subscriptions.trade != nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe trade failed because there is already an active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe trade failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for subscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "subscribe_trade", Root: fmt.Errorf("subscribe trade failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "subscribe_trade", Root: fmt.Errorf("subscribe trade failed: %w", err)})
		}
		// Register the subscription
		client.subscriptions.trade = &tradeSubscription{
			pairs: pairs,
			pub:   rcv,
		}
		// Return publish channel
		client.logger.Println("trade channel subscribed")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the spreads channel. In case of success, the websocket client will start
// publishing received events on the user's provided channel.
//
// Two types of events can be published on the channel:
//   - connection_interrupted: This event type is used when connection with the sevrer has been
//     interrupted. The event will not have any data. It only serves as a cue for the consumer
//     to allow the consumer to react when the connection with the server is interrupted.
//   - spread: This event type is used when a message has been received from the server.
//     Published events will contain both the received data and the tracing context to continue
//     the tracing span from the source (= the websocket engine).
//
// In case when the connection with the server is lost, the websocket client will publish a
// connection_interrupted event to warn consumer about the failure.
//
// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
// when it reconnects to the server and it MUST reuse the previously provided channel to publish
// received messages.
//
// Consumers should always watch to the event type to separate messages from the connection
// failure events and react according the event type.
//
// Finally, the provided channel will be automatically closed by the client when:
//   - The user unsubscribe from the topic by using UnsubscribeSpreads
//   - The websocket client definitely stops.
//
// Consumers should also watch channel closure to know when no more data will be delivered.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//   - connection_interrupted
//   - spread
//
// # Extract data
//
// Before parsing the data, check the event type to catch rare connection_interrupted events.
//
// The event data contains the JSON payload from the server and can be parsed into a structure
// of type messages.Spread like this:
//
//	spread := new(messages.Spread)
//	err := event.DataAs(spread)
//
// The event will also contain the tracing context from OpenTelemetry. This tracing context can
// be extracted from the event to continue tracing the event processing from the source:
//
//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - pair: Pairs to subscribe to.
//   - rcv: Channel used to publish spread messages and connection_interrupted events.
//
// # Return
//
// An error is returned when:
//
//   - There is already an active subscription.
//   - An error occurs when sending the subscription message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active subscription.
//
//   - The client MUST use the right error type as described in the "Return" section.
//
//   - A connection_interrupted event MUST be published on the channel each time the websocket
//     connection is closed.
//
//   - The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the provided channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server,
//     then the websocket client MUST resubscribe to previously subscribed channels and reuse
//     the channel that has been provided when the user subscribed to the channel.
func (client *krakenSpotWebsocketClient) SubscribeSpread(ctx context.Context, pairs []string, rcv chan event.Event) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "subscribe_spread",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
		))
	defer span.End()
	client.logger.Println("subscribing to spread channel", pairs)
	// Check if there is already an active subscription
	client.spreadSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.spreadSubMu.Unlock()
	if client.subscriptions.spread != nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe spread failed because there is already an active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe spread failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for subscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "subscribe_spread", Root: fmt.Errorf("subscribe spread failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "subscribe_spread", Root: fmt.Errorf("subscribe spread failed: %w", err)})
		}
		// Register the subscription
		client.subscriptions.spread = &spreadSubscription{
			pairs: pairs,
			pub:   rcv,
		}
		// Return publish channel
		client.logger.Println("spread channel subscribed")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the book channel. In case of success, the websocket client will start
// publishing received events on the user's provided channel.
//
// Three types of events can be published on the channel:
//   - connection_interrupted: This event type is used when connection with the sevrer has been
//     interrupted. The event will not have any data. It only serves as a cue for the consumer
//     to allow the consumer to react when the connection with the server is interrupted.
//   - book_snapshot: This event type is used when a snapshot of the order book is received from
//     the websocket server.
//   - book_update: This event is used when an update about the order book is received from the
//     websocket server.
//
// In case when the connection with the server is lost, the websocket client will publish a
// connection_interrupted event to warn consumer about the failure.
//
// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
// when it reconnects to the server and it MUST reuse the previously provided channel to publish
// received messages.
//
// Consumers should always watch to the event type to separate messages from the connection
// failure events and react according the event type.
//
// Finally, the provided channel will be automatically closed by the client when:
//   - The user unsubscribe from the topic by using UnsubscribeBook
//   - The websocket client definitely stops.
//
// Consumers should also watch channel closure to know when no more data will be delivered.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//   - connection_interrupted
//   - book_snapshot
//   - book_update
//
// # Extract data
//
// Before parsing the data, check the event type to catch rare connection_interrupted events.
//
// The event data contains the JSON payload from the server and can be parsed into a structure
// of type messages.BookUpdate or messages.BookSnapshot depending on the event type like this:
//
//	switch(WebsocketClientEventTypeEnum(event.Type)) {
//		case events.BookSnapshot:
//			snapshot := new(messages.BookSnapshot)
//			err := event.DataAs(snapshot)
//		case events.BookUpdate:
//			update := new(messages.BookUpdate)
//			err := event.DataAs(update)
//		case events.ConnectionInterrupted:
//			panic("connection lost")
//		default:
//			panic("unknown message type", event.Type)
//	}
//
// The event will also contain the tracing context from OpenTelemetry. This tracing context can
// be extracted from the event to continue tracing the event processing from the source:
//
//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - pair: Pairs to subscribe to.
//   - rcv: Channel used to publish book_snapshot & book+update messages and
//     connection_interrupted events.
//
// # Return
//
// An error is returned when:
//
//   - There is already an active subscription.
//   - An error occurs when sending the subscription message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active subscription.
//
//   - The client MUST use the right error type as described in the "Return" section.
//
//   - A connection_interrupted event MUST be published on the channel each time the websocket
//     connection is closed.
//
//   - The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the provided channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server,
//     then the websocket client MUST resubscribe to previously subscribed channels and reuse
//     the channel that has been provided when the user subscribed to the channel.
func (client *krakenSpotWebsocketClient) SubscribeBook(ctx context.Context, pairs []string, depth messages.DepthEnum, rcv chan event.Event) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "subscribe_book",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("depth", int(depth)),
		))
	defer span.End()
	client.logger.Println("subscribing to book channel")
	// Check if there is already an active subscription
	client.bookSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book != nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe book failed because there is already an active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe book failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for subscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "subscribe_book", Root: fmt.Errorf("subscribe book failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "subscribe_book", Root: fmt.Errorf("subscribe book failed: %w", err)})
		}
		// Register the subscription
		client.subscriptions.book = &bookSubscription{
			pairs: pairs,
			pub:   rcv,
			depth: depth,
		}
		// Return publish channel
		client.logger.Println("book channel subscribed")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Unsubscribe from the ticker channel. The channel provided on subscribe will be closed by
// the websocket client.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Return
//
// An error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the unsubscribe message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - In case of success, the client MUST close the channel used to publish events.
//
//   - The client MUST use the right error type as described in the "Return" section.
func (client *krakenSpotWebsocketClient) UnsubscribeTicker(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "unsubscribe_ticker", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("unsubscribing from ticker channel")
	// Check if there is already an active subscription
	client.tickerSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.tickerSubMu.Unlock()
	if client.subscriptions.ticker == nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe ticker failed because there is no active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe ticker failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for unsubscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "unsubscribe_ticker", Root: fmt.Errorf("unsubscribe ticker failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "unsubscribe_ticker", Root: fmt.Errorf("unsubscribe ticker failed: %w", err)})
		}
		// Close the publication channel, discard the subscription and exit
		close(client.subscriptions.ticker.pub)
		client.subscriptions.ticker = nil
		client.logger.Println("unsubscribed from ticker channel")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Unsubscribe from the ohlc channel with the given interva. The channel provided on subscribe
// will be closed by the websocket client.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - interval: Used to determine which subscription must be cancelled.
//
// # Return
//
// An error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the unsubscribe message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - In case of success, the client MUST close the channel used to publish events.
//
//   - The client MUST use the right error type as described in the "Return" section.
func (client *krakenSpotWebsocketClient) UnsubscribeOHLC(ctx context.Context, interval messages.IntervalEnum) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "unsubscribe_ohlc",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.Int("interval", int(interval))))
	defer span.End()
	client.logger.Println("unsubscribing from ohlc channel", interval)
	// Check if there is already an active subscription
	client.ohlcSubMu.Lock() // Lock mutex till unsubscribe completes - this will block Subscribe
	defer client.ohlcSubMu.Unlock()
	if client.subscriptions.ohlcs[interval] == nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe ohlc failed because there is no active subscription"))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err := client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Pairs: client.subscriptions.ohlcs[interval].pairs,
			Subscription: messages.UnsuscribeDetails{
				Name:     string(messages.ChannelOHLC),
				Interval: int(interval),
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe ohlc failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for unsubscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "unsubscribe_ohlc", Root: fmt.Errorf("unsubscribe ohlc failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "unsubscribe_ohlc", Root: fmt.Errorf("unsubscribe ohlc failed: %w", err)})
		}
		// Close the publication channel, discard the subscription and exit
		close(client.subscriptions.ohlcs[interval].pub)
		delete(client.subscriptions.ohlcs, interval)
		client.logger.Println("unsubscribed from ohlc channel", interval)
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Unsubscribe from the trade channel. The channel provided on subscribe will be closed by
// the websocket client.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Return
//
// An error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the unsubscribe message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - In case of success, the client MUST close the channel used to publish events.
//
//   - The client MUST use the right error type as described in the "Return" section.
func (client *krakenSpotWebsocketClient) UnsubscribeTrade(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "unsubscribe_trade", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("unsubscribing from trade channel")
	// Check if there is already an active subscription
	client.tradeSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.tradeSubMu.Unlock()
	if client.subscriptions.trade == nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe trade failed because there is no active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe trade failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for unsubscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "unsubscribe_trade", Root: fmt.Errorf("unsubscribe trade failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "unsubscribe_trade", Root: fmt.Errorf("unsubscribe trade failed: %w", err)})
		}
		// Close the publication channel, discard the subscription and exit
		close(client.subscriptions.trade.pub)
		client.subscriptions.trade = nil
		client.logger.Println("unsubscribed from trade channel")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Unsubscribe from the spread channel. The channel provided on subscribe will be closed by
// the websocket client.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Return
//
// An error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the unsubscribe message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - In case of success, the client MUST close the channel used to publish events.
//
//   - The client MUST use the right error type as described in the "Return" section.
func (client *krakenSpotWebsocketClient) UnsubscribeSpread(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "unsubscribe_spread", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("unsubscribing from spread channel")
	// Check if there is already an active subscription
	client.spreadSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.spreadSubMu.Unlock()
	if client.subscriptions.spread == nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe spread failed because there is no active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe spread failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for unsubscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "unsubscribe_spread", Root: fmt.Errorf("unsubscribe spread failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "unsubscribe_spread", Root: fmt.Errorf("unsubscribe spread failed: %w", err)})
		}
		// close the publication channel, discard the subscription and exit
		close(client.subscriptions.spread.pub)
		client.subscriptions.spread = nil
		span.SetStatus(codes.Ok, codes.Ok.String())
		client.logger.Println("unsubscribed from spread channel")
		return nil
	}
}

// # Description
//
// Unsubscribe from the book channel. The channel provided on subscribe will be closed by
// the websocket client.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Return
//
// An error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the unsubscribe message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - In case of success, the client MUST close the channel used to publish events.
//
//   - The client MUST use the right error type as described in the "Return" section.
func (client *krakenSpotWebsocketClient) UnsubscribeBook(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "unsubscribe_book", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("unsubscribing from book channel")
	// Check if there is already an active subscription
	client.bookSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book == nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe book failed because there is no active subscription"))
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe book failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for unsubscribe response from server")
	select {
	case <-ctx.Done():
		// Trace and return error - OperationInterruptedError
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "unsubscribe_book", Root: fmt.Errorf("unsubscribe book failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error - OperationError
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "unsubscribe_book", Root: fmt.Errorf("unsubscribe book failed: %w", err)})
		}
		// Close the publication channel, discard the subscription and exit
		close(client.subscriptions.book.pub)
		client.subscriptions.book = nil
		span.SetStatus(codes.Ok, codes.Ok.String())
		client.logger.Println("unsubscribed from book channel")
		return nil
	}
}

// # Description
//
// Get the client's built-in channel used to publish received system status updates.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//
//   - system_status
//
//     # Return
//
// The client's built-in channel used to publish received system status updates.
//
// # Implemetation and usage guidelines
//
//   - The client MUST provide the channel it will use to publish heartbeats even though the
//     cllient has not been started yet and is not connected to the server.
//
//   - The client MUST close the channel when it definitely stops.
//
//   - As the channel is automatically subscribed to, the client implementation must deal with
//     possible channel congestion by discarding messages in a FIFO or LIFO fashion. The client
//     must indicate how congestion is handled.
func (client *krakenSpotWebsocketClient) GetSystemStatusChannel() chan event.Event {
	return client.subscriptions.systemStatus
}

// # Description
//
// Get the client's built-in channel to publish received heartbeats.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//
//   - heartbeat
//
//     # Return
//
// # Implemetation and usage guidelines
//
//   - The client MUST provide the channel it will use to publish heartbeats even though the
//     cllient has not been started yet and is not connected to the server.
//
//   - The client MUST close the channel when it definitely stops.
//
//   - As the channel is automatically subscribed to, the client implementation must deal with
//     possible channel congestion by discarding messages in a FIFO or LIFO fashion. The client
//     must indicate how congestion is handled.
//
// # Return
//
// The client's built-in channel used to publish received heartbeats.
func (client *krakenSpotWebsocketClient) GetHeartbeatChannel() chan event.Event {
	return client.subscriptions.heartbeat
}

/*************************************************************************************************/
/* KRAKEN PRIVATE WEBSOCKET IMPL.                                                                */
/*************************************************************************************************/

// # Description
//
// Add a new order and wait until a AddOrderResponse response is received from the server or
// until an error or a timeout occurs.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - params: AddOrder request parameters.
//
// # Return
//
// The AddOrderResponse message from the server if any has been received. In case the response
// has its error message set, an error with the error message will also be returned.
//
// An error is returned when:
//
//   - The client failed to send the request (no specific error type).
//   - A timeout has occured before the request could be sent (no specific error type)
//   - An error message is received from the server (OperationError).
//   - A timeout or network failure occurs after sending the request to the server, while
//     waiting for the server response. In this case, a OperationInterruptedError is returned.
func (client *krakenSpotWebsocketClient) AddOrder(ctx context.Context, params AddOrderRequestParameters) (*messages.AddOrderResponse, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "add_order", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		attribute.String("order_type", params.OrderType),
		attribute.String("side", params.Type),
		attribute.String("pair", params.Pair),
		attribute.String("price", params.Price),
		attribute.String("price2", params.Price2),
		attribute.String("volume", params.Volume),
		attribute.Int("leverage", params.Leverage),
		attribute.Bool("reduce_only", params.ReduceOnly),
		attribute.String("oflags", params.OFlags),
		attribute.String("starttm", params.StartTimestamp),
		attribute.String("expiretm", params.ExpireTimestamp),
		attribute.String("deadline", params.Deadline),
		attribute.String("userref", params.UserReference),
		attribute.Bool("validate", params.Validate),
		attribute.String("close_order_type", params.CloseOrderType),
		attribute.String("close_price", params.ClosePrice),
		attribute.String("close_price2", params.ClosePrice2),
		attribute.String("time_in_force", params.TimeInForce),
	))
	defer span.End()
	client.logger.Println("sending add order request to the server", params.Pair, params.OrderType, params.Type)
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("add order failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	respChan := make(chan *messages.AddOrderResponse, 1)
	// Format request
	req := &messages.AddOrderRequest{
		Event:           string(messages.EventTypeAddOrder),
		Token:           token,
		RequestId:       client.ngen.GenerateNonce(),
		OrderType:       params.OrderType,
		Type:            params.Type,
		Pair:            params.Pair,
		Price:           params.Price,
		Price2:          params.Price2,
		Volume:          params.Volume,
		Leverage:        strconv.FormatInt(int64(params.Leverage), 10),
		ReduceOnly:      params.ReduceOnly,
		OFlags:          params.OFlags,
		StartTimestamp:  params.StartTimestamp,
		ExpireTimestamp: params.ExpireTimestamp,
		Deadline:        params.Deadline,
		UserReference:   params.UserReference,
		Validate:        strconv.FormatBool(params.Validate),
		CloseOrderType:  params.CloseOrderType,
		ClosePrice:      params.ClosePrice,
		ClosePrice2:     params.ClosePrice2,
		TimeInForce:     params.TimeInForce,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("add order failed: %w", err))
	}
	// Add pending addOrder request
	client.pendingAddOrderMu.Lock()
	client.requests.pendingAddOrderRequests[req.RequestId] = &pendingAddOrderRequest{
		resp: respChan,
		err:  errChan,
	}
	// Defer pending request cleanup
	defer delete(client.requests.pendingAddOrderRequests, req.RequestId)
	// Defer pending request map unlock in a sync.Once
	unlock := sync.OnceFunc(client.pendingAddOrderMu.Unlock)
	defer unlock()
	// Write message to the server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Trace error and exit
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("add order failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	unlock() // Unlock pending request map so another go routine can fulfil it
	client.logger.Println("waiting for a response (addOrderStatus) from the server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "add_order", Root: fmt.Errorf("add order failed: %w", ctx.Err())})
	case err := <-errChan:
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "add_order", Root: fmt.Errorf("add order failed: %w", err)})
	case resp := <-respChan:
		// Tracing: Add an event for the response
		span.AddEvent("add_order_response", trace.WithAttributes(
			attribute.String("status", resp.Status),
			attribute.String("txid", resp.TxId),
			attribute.String("error", resp.Err),
			attribute.Int64("request_id", *resp.RequestId),
		))
		// Check the response status
		if resp.Status == string(messages.Err) {
			return resp, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "add_order", Root: fmt.Errorf("add order failed: %s", resp.Err)})
		}
		// Exit - success
		span.SetStatus(codes.Ok, codes.Ok.String())
		client.logger.Println("addOrder has succeeded", resp.TxId)
		return resp, nil
	}
}

// # Description
//
// Edit an existing order and wait until a EditOrderResponse response is received from the
// server or until an error or a timeout occurs.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - params: EditOrder request parameters.
//
// # Return
//
// The EditOrderResponse message from the server if any has been received. In case the response
// has its error message set, an error with the error message will also be returned.
//
// An error is returned when:
//
//   - The client failed to send the request (no specific error type).
//   - A timeout has occured before the request could be sent (no specific error type)
//   - An error message is received from the server (OperationError).
//   - A timeout or network failure occurs after sending the request to the server, while
//     waiting for the server response. In this case, a OperationInterruptedError is returned.
func (client *krakenSpotWebsocketClient) EditOrder(ctx context.Context, params EditOrderRequestParameters) (*messages.EditOrderResponse, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "edit_order", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		attribute.String("id", params.Id),
		attribute.String("pair", params.Pair),
		attribute.String("price", params.Price),
		attribute.String("price2", params.Price2),
		attribute.String("volume", params.Volume),
		attribute.String("oflags", params.OFlags),
		attribute.String("new_userref", params.NewUserReference),
		attribute.Bool("validate", params.Validate),
	))
	defer span.End()
	client.logger.Println("sending edit order request to the server", params.Id)
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("edit order failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	respChan := make(chan *messages.EditOrderResponse, 1)
	// Format request
	req := &messages.EditOrderRequest{
		Event:            string(messages.EventTypeEditOrder),
		Token:            token,
		RequestId:        client.ngen.GenerateNonce(),
		Pair:             params.Pair,
		Price:            params.Price,
		Price2:           params.Price2,
		Volume:           params.Volume,
		OFlags:           params.OFlags,
		Validate:         strconv.FormatBool(params.Validate),
		NewUserReference: params.NewUserReference,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("edit order failed: %w", err))
	}
	// Add pending editOrder request
	client.pendingEditOrderMu.Lock()
	client.requests.pendingEditOrderRequests[req.RequestId] = &pendingEditOrderRequest{
		resp: respChan,
		err:  errChan,
	}
	// Defer map clean
	defer delete(client.requests.pendingEditOrderRequests, req.RequestId)
	// Defer unlock in a sync.Once
	unlock := sync.OnceFunc(client.pendingEditOrderMu.Unlock)
	defer unlock()
	// Write message to the server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("edit order failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	unlock() // Unlock so another goroutine can complete the request
	client.logger.Println("waiting for a response (editOrderStatus) from the server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "edit_order", Root: fmt.Errorf("edit order failed: %w", ctx.Err())})
	case err := <-errChan:
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "edit_order", Root: fmt.Errorf("edit order failed: %w", err)})
	case resp := <-respChan:
		// Tracing: Add an event for the response
		span.AddEvent("edit_order_response", trace.WithAttributes(
			attribute.String("status", resp.Status),
			attribute.String("original_txid", resp.OriginalTxId),
			attribute.String("txid", resp.TxId),
			attribute.String("description", resp.Description),
			attribute.String("error", resp.Err),
			attribute.Int64("request_id", *resp.RequestId),
		))
		// Check the response status
		if resp.Status == string(messages.Err) {
			return resp, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "edit_order", Root: fmt.Errorf("edit order failed: %s", resp.Err)})
		}
		// Exit - success
		span.SetStatus(codes.Ok, codes.Ok.String())
		client.logger.Println("editOrder has succeeded", resp.TxId)
		return resp, nil
	}
}

// # Description
//
// Cancel one or several existing orders and wait until a CancelOrderResponse response is
// received from the server or until an error or a timeout occurs.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - params: CancelOrder request parameters.
//
// # Return
//
// The CancelOrderResponse message from the server if any has been received. In case the response
// has its error message set, an error with the error message will also be returned.
//
// An error is returned when:
//
//   - The client failed to send the request (no specific error type).
//   - A timeout has occured before the request could be sent (no specific error type)
//   - An error message is received from the server (OperationError).
//   - A timeout or network failure occurs after sending the request to the server, while
//     waiting for the server response. In this case, a OperationInterruptedError is returned.
func (client *krakenSpotWebsocketClient) CancelOrder(ctx context.Context, params CancelOrderRequestParameters) (*messages.CancelOrderResponse, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "cancel_order", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		attribute.StringSlice("id", params.TxId),
	))
	defer span.End()
	client.logger.Println("sending cancel order request to the server", params.TxId)
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel order failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	respChan := make(chan *messages.CancelOrderResponse, 1)
	// Format request
	req := &messages.CancelOrderRequest{
		Event:     string(messages.EventTypeCancelOrder),
		Token:     token,
		RequestId: client.ngen.GenerateNonce(),
		TxId:      params.TxId,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel order failed: %w", err))
	}
	// Add pending cancelOrder request
	client.pendingCancelOrderMu.Lock()
	client.requests.pendingCancelOrderRequests[req.RequestId] = &pendingCancelOrderRequest{
		resp: respChan,
		err:  errChan,
	}
	// Defer map clean
	defer delete(client.requests.pendingCancelOrderRequests, req.RequestId)
	// Defer unlock in a sync.Once
	unlock := sync.OnceFunc(client.pendingCancelOrderMu.Unlock)
	defer unlock()
	// Write message to the server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Discard pending request, trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel order failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	unlock() // Unlock so another goroutine can compelte the request
	client.logger.Println("waiting for a response (cancelOrderStatus) from the server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "cancel_order", Root: fmt.Errorf("cancel order failed: %w", ctx.Err())})
	case err := <-errChan:
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "cancel_order", Root: fmt.Errorf("cancel order failed: %w", err)})
	case resp := <-respChan:
		// Tracing: Add an event for the response
		span.AddEvent("cancel_order_response", trace.WithAttributes(
			attribute.String("status", resp.Status),
			attribute.String("error", resp.Err),
			attribute.Int64("request_id", *resp.RequestId),
		))
		// Check the response status
		if resp.Status == string(messages.Err) {
			return resp, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "cancel_order", Root: fmt.Errorf("cancel order failed: %s", resp.Err)})
		}
		// Exit - success
		span.SetStatus(codes.Ok, codes.Ok.String())
		client.logger.Println("cancelOrder has succeeded")
		return resp, nil
	}
}

// # Description
//
// Cancel all orders and wait until a CancelAllOrdersResponse response is received from the
// server or until an error or a timeout occurs.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//
// # Return
//
// The CancelAllOrdersResponse message from the server if any has been received. In case the response
// has its error message set, an error with the error message will also be returned.
//
// An error is returned when:
//
//   - The client failed to send the request (no specific error type).
//   - A timeout has occured before the request could be sent (no specific error type)
//   - An error message is received from the server (OperationError).
//   - A timeout or network failure occurs after sending the request to the server, while
//     waiting for the server response. In this case, a OperationInterruptedError is returned.
func (client *krakenSpotWebsocketClient) CancellAllOrders(ctx context.Context) (*messages.CancelAllOrdersResponse, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "cancel_all_orders", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("sending cancel all orders request to the server")
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel all orders failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	respChan := make(chan *messages.CancelAllOrdersResponse, 1)
	// Format request
	req := &messages.CancelAllOrdersRequest{
		Event:     string(messages.EventTypeCancelAllOrders),
		Token:     token,
		RequestId: client.ngen.GenerateNonce(),
	}
	payload, err := json.Marshal(req)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel all orders failed: %w", err))
	}
	// Add pending cancelAllOrders request
	client.pendingCancelAllOrdersMu.Lock()
	client.requests.pendingCancelAllOrdersRequests[req.RequestId] = &pendingCancelAllOrdersRequest{
		resp: respChan,
		err:  errChan,
	}
	// Defer map clean
	defer delete(client.requests.pendingCancelAllOrdersRequests, req.RequestId)
	// Defer unlock in a sync.Once
	unlock := sync.OnceFunc(client.pendingCancelAllOrdersMu.Unlock)
	defer unlock()
	// Write message to the server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel all orders failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	unlock()
	client.logger.Println("waiting for a response (cancelAllOrdersStatus) from the server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "cancel_all_orders", Root: fmt.Errorf("cancel all orders failed: %w", ctx.Err())})
	case err := <-errChan:
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "cancel_all_orders", Root: fmt.Errorf("cancel all orders failed: %w", err)})
	case resp := <-respChan:
		// Tracing: Add an event for the response
		span.AddEvent("cancel_all_orders_response", trace.WithAttributes(
			attribute.String("status", resp.Status),
			attribute.String("error", resp.Err),
			attribute.Int64("request_id", *resp.RequestId),
		))
		// Check the response status
		if resp.Status == string(messages.Err) {
			return resp, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "cancel_all_orders", Root: fmt.Errorf("cancel all orders failed: %w", err)})
		}
		// Exit - success
		client.logger.Println("cancel all orders has succeeded")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return resp, nil
	}
}

// # Description
//
// Set, extend or unset a timer which cancels all orders when expiring and wait until a
// response is received from the server or until an error or a timeout occurs.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - params: CancellAllOrdersAfterX request parameters.
//
// # Return
//
// The CancelAllOrdersAfterXResponse message from the server if any has been received. In case
// the response has its error message set, an error with the error message is also be returned.
//
// An error is returned when:
//
//   - The client failed to send the request (no specific error type).
//   - A timeout has occured before the request could be sent (no specific error type)
//   - An error message is received from the server (OperationError).
//   - A timeout or network failure occurs after sending the request to the server, while
//     waiting for the server response. In this case, a OperationInterruptedError is returned.
func (client *krakenSpotWebsocketClient) CancellAllOrdersAfterX(ctx context.Context, params CancelAllOrdersAfterXRequestParameters) (*messages.CancelAllOrdersAfterXResponse, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "cancel_all_orders_after_x", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		attribute.Int("timeout", params.Timeout),
	))
	defer span.End()
	client.logger.Println("sending cancel all orders after x request to the server", params.Timeout)
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel all orders after x failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	respChan := make(chan *messages.CancelAllOrdersAfterXResponse, 1)
	// Format request
	req := &messages.CancelAllOrdersAfterXRequest{
		Event:     string(messages.EventTypeCancelAllOrdersAfterX),
		Token:     token,
		RequestId: client.ngen.GenerateNonce(),
		Timeout:   params.Timeout,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel all orders after x failed: %w", err))
	}
	// Add pending cancelAllOrders request
	client.pendingCancelAllOrdersAfterXOrderMu.Lock()
	client.requests.pendingCancelAllOrdersAfterXRequests[req.RequestId] = &pendingCancelAllOrdersAfterXRequest{
		resp: respChan,
		err:  errChan,
	}
	// Defer map clean
	defer delete(client.requests.pendingCancelAllOrdersAfterXRequests, req.RequestId)
	// Defer unlock in a sync.Once
	unlock := sync.OnceFunc(client.pendingCancelAllOrdersAfterXOrderMu.Unlock)
	defer unlock()
	// Write message to the server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("cancel all orders after x failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	unlock()
	client.logger.Println("waiting for a response (cancelAllOrdersAfterXStatus) from the server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "cancel_all_orders_after_x", Root: fmt.Errorf("cancel all orders after x failed: %w", ctx.Err())})
	case err := <-errChan:
		// Trace and return error
		return nil, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "cancel_all_orders_after_x", Root: fmt.Errorf("cancel all orders after x failed: %w", err)})
	case resp := <-respChan:
		// Tracing: Add an event for the response
		span.AddEvent("cancel_all_orders_after_x", trace.WithAttributes(
			attribute.String("status", resp.Status),
			attribute.String("current_time", resp.CurrentTime),
			attribute.String("trigger_time", resp.TriggerTime),
			attribute.String("error", resp.Err),
			attribute.Int64("request_id", *resp.RequestId),
		))
		// Check the response status
		if resp.Status == string(messages.Err) {
			return resp, tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "cancel_all_orders_after_x", Root: fmt.Errorf("cancel all orders after x failed: %s", resp.Err)})
		}
		// Exit - success
		client.logger.Println("cancel all orders has succeeded")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return resp, nil
	}
}

// # Description
//
// Subscribe to the ownTrades channel. In case of success, the websocket client will start
// publishing received events on the user's provided channel.
//
// Two types of events can be published on the channel:
//   - connection_interrupted: This event type is used when connection with the sevrer has been
//     interrupted. The event will not have any data. It only serves as a cue for the consumer
//     to allow the consumer to react when the connection with the server is interrupted.
//   - own_trades: This event type is used when a message has been received from the server.
//     Published events will contain both the received data and the tracing context to continue
//     the tracing span from the source (= the websocket engine).
//
// In case when the connection with the server is lost, the websocket client will publish a
// connection_interrupted event to warn consumer about the failure.
//
// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
// when it reconnects to the server and it MUST reuse the previously provided channel to publish
// received messages.
//
// Consumers should always watch to the event type to separate messages from the connection
// failure events and react according the event type.
//
// Finally, the provided channel will be automatically closed by the client when:
//   - The user unsubscribe from the topic by using UnsubscribeOwnTrades
//   - The websocket client definitely stops.
//
// Consumers should also watch channel closure to know when no more data will be delivered.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//   - connection_interrupted
//   - own_trades
//
// # Extract data
//
// Before parsing the data, check the event type to catch rare connection_interrupted events.
//
// The event data contains the JSON payload from the server and can be parsed into a structure
// of type messages.OwnTrades like this:
//
//	ownTrade := new(messages.OwnTrades)
//	err := event.DataAs(ownTrade)
//
// The event will also contain the tracing context from OpenTelemetry. This tracing context can
// be extracted from the event to continue tracing the event processing from the source:
//
//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - snapshot: If true, upon subscription, the 50 most recent user trades will be published.
//   - consolidateTaker: Whether to consolidate order fills by root taker trade(s).
//   - rcv: Channel used to publish own_trades messages and connection_interrupted events.
//
// # Return
//
// An error is returned when:
//
//   - There is already an active subscription.
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel) before subscription is completed.
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active susbscription.
//
//   - The client MUST use the right error type as described in the "Return" section.
//
//   - A connection_interrupted event MUST be published on the channel each time the websocket
//     connection is closed.
//
//   - The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the provided channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server,
//     then the websocket client MUST resubscribe to previously subscribed channels and reuse
//     the channel that has been provided when the user subscribed to the channel.
func (client *krakenSpotWebsocketClient) SubscribeOwnTrades(ctx context.Context, snapshot bool, consolidateTaker bool, rcv chan event.Event) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "subscribe_own_trades",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Bool("snapshot", snapshot),
			attribute.Bool("consolidate_taker", consolidateTaker),
		))
	defer span.End()
	client.logger.Println("subscribing to own trades channel")
	// Check if there is already an active subscription
	client.ownTradesSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.ownTradesSubMu.Unlock()
	if client.subscriptions.ownTrades != nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe own trades failed because there is already an active subscription"))
	}
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe own trades failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send subscribe message to server
	err = client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Subscription: messages.SuscribeDetails{
				Name:             string(messages.ChannelOwnTrades),
				Snapshot:         &snapshot,
				ConsolidateTaker: &consolidateTaker,
				Token:            token,
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe own trades failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for a subscribe response from the server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "subscribe_own_trades", Root: fmt.Errorf("subscribe own trades failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "subscribe_own_trades", Root: fmt.Errorf("subscribe own trades failed: %w", err)})
		}
		// Register the subscription
		client.subscriptions.ownTrades = &ownTradesSubscription{
			pub:              rcv,
			consolidateTaker: consolidateTaker,
			snapshot:         snapshot,
		}
		// Return publish channel
		client.logger.Println("subscribe own trades channel has succeeded")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the openOrders channel. In case of success, the websocket client will start
// publishing received events on the user's provided channel.
//
// Two types of events can be published on the channel:
//   - connection_interrupted: This event type is used when connection with the sevrer has been
//     interrupted. The event will not have any data. It only serves as a cue for the consumer
//     to allow the consumer to react when the connection with the server is interrupted.
//   - open_orders: This event type is used when a message has been received from the server.
//     Published events will contain both the received data and the tracing context to continue
//     the tracing span from the source (= the websocket engine).
//
// In case when the connection with the server is lost, the websocket client will publish a
// connection_interrupted event to warn consumer about the failure.
//
// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
// when it reconnects to the server and it MUST reuse the previously provided channel to publish
// received messages.
//
// Consumers should always watch to the event type to separate messages from the connection
// failure events and react according the event type.
//
// Finally, the provided channel will be automatically closed by the client when:
//   - The user unsubscribe from the topic by using UnsubscribeOpenOrders
//   - The websocket client definitely stops.
//
// Consumers should also watch channel closure to know when no more data will be delivered.
//
// # Event types
//
// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
//   - connection_interrupted
//   - open_orders
//
// # Extract data
//
// Before parsing the data, check the event type to catch rare connection_interrupted events.
//
// The event data contains the JSON payload from the server and can be parsed into a structure
// of type messages.OpenOrders like this:
//
//	openOrders := new(messages.OpenOrders)
//	err := event.DataAs(openOrders)
//
// The event will also contain the tracing context from OpenTelemetry. This tracing context can
// be extracted from the event to continue tracing the event processing from the source:
//
//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - rateCounter: If true, rate limiting information will be included in messages.
//   - rcv: Channel used to publish open_orders messages and connection_interrupted events.
//
// # Return
//
// An error is returned when:
//
//   - There is already an active subscription.
//   - An error occurs when sending the subscription message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - The client MUST return an error if there is already an active subscription.
//
//   - The client MUST use the right error type as described in the "Return" section.
//
//   - A connection_interrupted event MUST be published on the channel each time the websocket
//     connection is closed.
//
//   - The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
//
//   - The websocket client implementation CAN either use blocking writes or discard messages in
//     case the provided channel is full. It is up to the client implementation to be clear about
//     how it deals with congestion.
//
//   - If the client implementation has a mechanism to automatically reconnect to the server,
//     then the websocket client MUST resubscribe to previously subscribed channels and reuse
//     the channel that has been provided when the user subscribed to the channel.
func (client *krakenSpotWebsocketClient) SubscribeOpenOrders(ctx context.Context, rateCounter bool, rcv chan event.Event) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "subscribe_open_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Bool("rate_counter", rateCounter),
		))
	defer span.End()
	client.logger.Println("subscribing to open orders channel")
	// Check if there is already an active subscription
	client.openOrdersSubMu.Lock() // Lock mutex till subscribe completes - this will block Unsubscribe
	defer client.openOrdersSubMu.Unlock()
	if client.subscriptions.openOrders != nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe open orders failed because there is already an active subscription"))
	}
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe open orders failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send subscribe message to server
	err = client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Subscription: messages.SuscribeDetails{
				Name:        string(messages.ChannelOpenOrders),
				RateCounter: rateCounter,
				Token:       token,
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("subscribe open orders failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for a subscribe response from the server")
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "subscribe_open_orders", Root: fmt.Errorf("subscribe open orders failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "subscribe_open_orders", Root: fmt.Errorf("subscribe open orders failed: %w", err)})
		}
		// Register the subscription
		client.subscriptions.openOrders = &openOrdersSubscription{
			rateCounter: rateCounter,
			pub:         rcv,
		}
		// Return publish channel
		client.logger.Println("subscribe open orders channel has succeeded")
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Unsubscribe from the ownTrades channel. The channel provided on subscribe will bbe closed by
// the websocket client.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Return
//
// An error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the unsubscribe message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - In case of success, the client MUST close the channel used to publish events.
//
//   - The client MUST use the right error type as described in the "Return" section.
func (client *krakenSpotWebsocketClient) UnsubscribeOwnTrades(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "unsubscribe_own_trades", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("unsubscribing from own trades channel")
	// Check if there is already an active subscription
	client.ownTradesSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.ownTradesSubMu.Unlock()
	if client.subscriptions.ownTrades == nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe own trades failed because there is no active subscription"))
	}
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe own trades failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err = client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeUnsubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Subscription: messages.UnsuscribeDetails{
				Name:  string(messages.ChannelOwnTrades),
				Token: token,
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe own trades failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for a unsubscribe response from the server")
	select {
	case <-ctx.Done():
		// Trace and return error - OperationInterruptedError
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "unsubscribe_own_trades", Root: fmt.Errorf("unsubscribe own trades failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error - OperationError
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "unsubscribe_own_trades", Root: fmt.Errorf("unsubscribe own trades failed: %w", err)})
		}
		// Discard the subscription and exit
		client.logger.Println("unsubscribed from own trades channel")
		span.SetStatus(codes.Ok, codes.Ok.String())
		client.subscriptions.ownTrades = nil
		return nil
	}
}

// # Description
//
// Unsubscribe from the openOrders channel. The channel provided on subscribe will bbe closed by
// the websocket client.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Return
//
// An error is returned when:
//
//   - The channel has not been subscribed to.
//   - An error occurs when sending the unsubscribe message.
//   - The provided context expires before subscription is completed (OperationInterruptedError).
//   - An error message is received from the server (OperationError).
//
// # Implementation and usage guidelines
//
//   - In case of success, the client MUST close the channel used to publish events.
//
//   - The client MUST use the right error type as described in the "Return" section.
func (client *krakenSpotWebsocketClient) UnsubscribeOpenOrders(ctx context.Context) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "unsubscribe_open_orders", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	client.logger.Println("unsubscribing from open orders channel")
	// Check if there is already an active subscription
	client.openOrdersSubMu.Lock() // Lock mutex till subscribe completes - this will block Subscribe
	defer client.openOrdersSubMu.Unlock()
	if client.subscriptions.openOrders == nil {
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe open orders failed because there is no active subscription"))
	}
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe open orders failed: %w", err))
	}
	// Create response channels
	errChan := make(chan error, 1)
	// Send unsubscribe message to server
	err = client.sendUnsubscribeRequest(
		ctx,
		&messages.Unsubscribe{
			Event: string(messages.EventTypeUnsubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Subscription: messages.UnsuscribeDetails{
				Name:  string(messages.ChannelOpenOrders),
				Token: token,
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("unsubscribe open orders failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	client.logger.Println("waiting for a unsubscribe response from the server")
	select {
	case <-ctx.Done():
		// Trace and return error - OperationInterruptedError
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "unsubscribe_open_orders", Root: fmt.Errorf("unsubscribe open orders failed: %w", ctx.Err())})
	case err := <-errChan:
		if err != nil {
			// Trace and return error - OperationError
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "unsubscribe_open_orders", Root: fmt.Errorf("unsubscribe open orders failed: %w", err)})
		}
		// Discard the subscription and exit
		client.logger.Println("unsubscribed from open orders channel")
		span.SetStatus(codes.Ok, codes.Ok.String())
		client.subscriptions.openOrders = nil
		return nil
	}
}

/*************************************************************************************************/
/* WEBSOCKET ENGINE CLIENT IMPLEMENTATION                                                        */
/*************************************************************************************************/

// # Description
//
// In case the client is reconnecting to the server, the client will attempt to resubscribe to all
// channels that have been previously subscribed. The client will attempts at most three times to
// resubscribe. THe client will not wait for resubscribe to succeed before resuming its operations.
//
// It is up to the user to monitor interruptions in stream of data and react according its own
// needs and requirements. In such a case, user can either kill/restart its application,
// unsubscribe and resubscribe to channel or shutdown and start again the wesocket client.
//
// # OnOpen Documentation
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
func (client *krakenSpotWebsocketClient) OnOpen(
	ctx context.Context,
	resp *http.Response,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	exit context.CancelFunc,
	restarting bool) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "on_open", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(
		attribute.Bool("restarting", restarting),
	))
	defer span.End()
	client.logger.Println("connection opened with the server - restarting:", restarting)
	// Store new connection
	client.conn = conn
	// Restore all active subscriptions if restarting
	if restarting {
		// Provided context is canceled by the engine after OnOpen exits. Hence, a separate context
		// with a separate timeout must be used by resubscribe goroutine otherwise they will be
		// canceled a little bit after starting
		propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
		carrier := propagation.MapCarrier{}
		propgator.Inject(ctx, carrier)
		rootctx := propgator.Extract(context.Background(), carrier)
		// Retry limit & base wait time
		base := 2.0
		limit := 3
		// Resubscribe to ticker if an active subscription is set
		client.tickerSubMu.Lock()
		defer client.tickerSubMu.Unlock()
		if client.subscriptions.ticker != nil {
			// Start a goroutine that will perform the resubscribe.
			// Goroutine will make 3 attempts then exit.
			client.logger.Println("starting process to resubscribe to ticker channel", client.subscriptions.ticker.pairs)
			go func(client *krakenSpotWebsocketClient) {
				ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
				defer cancel()
				for retry := 0; retry < limit; retry++ {
					err := client.resubscribeTicker(ctx, client.subscriptions.ticker.pairs)
					if err != nil {
						// Wait an exponential amount of time before retrying (1, 2 & 4 seconds)
						eerr := fmt.Errorf("resubscribe ticker attempt number %d failed: %w", retry+1, err)
						client.logger.Println(eerr.Error())
						time.Sleep(time.Second * time.Duration(int64(math.Pow(base, float64(retry)))))
					} else {
						// Break
						break
					}
				}
				client.logger.Println("resubscribe ticker definitly failed")
			}(client)
		}
		// Resubscribe to ohlcs if an active subscription is set
		client.ohlcSubMu.Lock()
		defer client.ohlcSubMu.Unlock()
		for interval := range client.subscriptions.ohlcs {
			osub := client.subscriptions.ohlcs[interval]
			// Start a goroutine that will perform the resubscribe.
			// Goroutine will make 3 attempts then exit.
			client.logger.Println("starting process to resubscribe to ohlc channel", osub.pairs, osub.interval)
			go func(client *krakenSpotWebsocketClient) {
				ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
				defer cancel()
				for retry := 0; retry < limit; retry++ {
					err := client.resubscribeOHLC(ctx, osub.pairs, osub.interval)
					if err != nil {
						// Wait an exponential amount of time before retrying (1, 2 & 4 seconds)
						eerr := fmt.Errorf("resubscribe ohlc attempt number %d failed: %w", retry+1, err)
						client.logger.Println(eerr.Error())
						time.Sleep(time.Second * time.Duration(int64(math.Pow(base, float64(retry)))))
					} else {
						// Break
						break
					}
				}
				client.logger.Println("resubscribe ohlc definitly failed")
			}(client)
		}
		// Resubscribe to trade if an active subscription is set
		client.tradeSubMu.Lock()
		defer client.tradeSubMu.Unlock()
		if client.subscriptions.trade != nil {
			// Start a goroutine that will perform the resubscribe.
			// Goroutine will make 3 attempts then exit.
			client.logger.Println("starting process to resubscribe to trade channel", client.subscriptions.trade.pairs)
			go func(client *krakenSpotWebsocketClient) {
				for retry := 0; retry < limit; retry++ {
					ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
					defer cancel()
					err := client.resubscribeTrade(ctx, client.subscriptions.trade.pairs)
					if err != nil {
						// Wait an exponential amount of time before retrying (1, 2 & 4 seconds)
						eerr := fmt.Errorf("resubscribe trade attempt number %d failed: %w", retry+1, err)
						client.logger.Println(eerr.Error())
						time.Sleep(time.Second * time.Duration(int64(math.Pow(base, float64(retry)))))
					} else {
						// Break
						break
					}
				}
				client.logger.Println("resubscribe trade definitly failed")
			}(client)
		}
		// Resubscribe to spread if an active subscription is set
		client.spreadSubMu.Lock()
		defer client.spreadSubMu.Unlock()
		if client.subscriptions.spread != nil {
			// Start a goroutine that will perform the resubscribe.
			// Goroutine will make 3 attempts then exit.
			client.logger.Println("starting process to resubscribe to spread channel", client.subscriptions.spread.pairs)
			go func(client *krakenSpotWebsocketClient) {
				ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
				defer cancel()
				for retry := 0; retry < limit; retry++ {
					err := client.resubscribeSpread(ctx, client.subscriptions.spread.pairs)
					if err != nil {
						// Wait an exponential amount of time before retrying (1, 2 & 4 seconds)
						eerr := fmt.Errorf("resubscribe spread attempt number %d failed: %w", retry+1, err)
						client.logger.Println(eerr.Error())
						time.Sleep(time.Second * time.Duration(int64(math.Pow(base, float64(retry)))))
					} else {
						// Break
						break
					}
				}
				client.logger.Println("resubscribe spread definitly failed")
			}(client)
		}
		// Resubscribe to book if an active subscription is set
		client.bookSubMu.Lock()
		defer client.bookSubMu.Unlock()
		if client.subscriptions.book != nil {
			// Start a goroutine that will perform the resubscribe.
			// Goroutine will make 3 attempts then exit.
			client.logger.Println("starting process to resubscribe to book channel", client.subscriptions.book.pairs, client.subscriptions.book.depth)
			go func(client *krakenSpotWebsocketClient) {
				for retry := 0; retry < limit; retry++ {
					ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
					defer cancel()
					err := client.resubscribeBook(ctx, client.subscriptions.book.pairs, client.subscriptions.book.depth)
					if err != nil {
						// Wait an exponential amount of time before retrying (1, 2 & 4 seconds)
						eerr := fmt.Errorf("resubscribe book attempt number %d failed: %w", retry+1, err)
						client.logger.Println(eerr.Error())
						time.Sleep(time.Second * time.Duration(int64(math.Pow(base, float64(retry)))))
					} else {
						// Break
						break
					}
				}
				client.logger.Println("resubscribe book definitly failed")
			}(client)
		}
		// Resubscribe to own trades if an active subscription is set
		client.ownTradesSubMu.Lock()
		defer client.ownTradesSubMu.Unlock()
		if client.subscriptions.ownTrades != nil {
			// Start a goroutine that will perform the resubscribe.
			// Goroutine will make 3 attempts then exit.
			client.logger.Println("starting process to resubscribe to own trades channel")
			go func(client *krakenSpotWebsocketClient) {
				for retry := 0; retry < limit; retry++ {
					ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
					defer cancel()
					err := client.resubscribeOwnTrades(ctx, client.subscriptions.ownTrades.snapshot, client.subscriptions.ownTrades.consolidateTaker)
					if err != nil {
						// Wait an exponential amount of time before retrying (1, 2 & 4 seconds)
						eerr := fmt.Errorf("resubscribe own trades attempt number %d failed: %w", retry+1, err)
						client.logger.Println(eerr.Error())
						time.Sleep(time.Second * time.Duration(int64(math.Pow(base, float64(retry)))))
					} else {
						// Break
						break
					}
				}
				client.logger.Println("resubscribe own trades definitly failed")
			}(client)
		}
		// Resubscribe to open orders if an active subscription is set
		client.openOrdersSubMu.Lock()
		defer client.openOrdersSubMu.Unlock()
		if client.subscriptions.openOrders != nil {
			// Start a goroutine that will perform the resubscribe.
			// Goroutine will make 3 attempts then exit.
			client.logger.Println("starting process to resubscribe to open orders channel")
			go func(client *krakenSpotWebsocketClient) {
				for retry := 0; retry < limit; retry++ {
					ctx, cancel := context.WithTimeout(rootctx, 30*time.Second)
					defer cancel()
					err := client.resubscribeOpenOrders(ctx, client.subscriptions.openOrders.rateCounter)
					if err != nil {
						// Wait an exponential amount of time before retrying (1, 2 & 4 seconds)
						eerr := fmt.Errorf("resubscribe open orders attempt number %d failed: %w", retry+1, err)
						client.logger.Println(eerr.Error())
						time.Sleep(time.Second * time.Duration(int64(math.Pow(base, float64(retry)))))
					} else {
						// Break
						break
					}
				}
				client.logger.Println("resubscribe open orders definitly failed")
			}(client)
		}
		// Do not wait for goroutines: Engine will start reading messages only after OnOpen completes
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
func (client *krakenSpotWebsocketClient) OnMessage(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "on_message",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("session_id", sessionId),
		))
	defer span.End()
	client.logger.Println("message received from the server")
	// Match the message type - 5 matches are expected
	matches := messages.MatchMessageTypeRegex.FindStringSubmatch(string(msg))
	if len(matches) != 5 {
		// Call OnReadError - Not the expected number of matches
		err := fmt.Errorf("failed to extract the message type from '%s' - not the expected number of matches %d", string(msg), len(matches))
		tracing.HandleAndTraLogError(span, client.logger, err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return
	}
	// Extract the message type from the matches. The regex will try to find the event type and the pair in case of a public
	// market event (ticker, spread, ...).
	//
	// Index 0 will contain the original message
	// Index 1 will contain the event type in case the message is a JSON object (usually request/responses)
	// Index 2 will contain the event type in case the message is a JSON Array (openOrders or ownTrades)
	// Index 3 will contain the event type in case the message is a JSON Array (public market data)
	// Index 4 will contain the pair in case the message is a public market data event like a spread.
	mType := matches[1]
	if mType == "" {
		mType = matches[2]
		if mType == "" {
			mType = matches[3]
		}
	}
	// Depending on the message type.
	splits := strings.Split(mType, "-")
	client.logger.Println("received message type: ", splits[0])
	switch splits[0] {
	// General error has been received
	case string(messages.EventTypeError):
		client.handleErrorMessage(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Trade
	case string(messages.ChannelTrade):
		client.handleTrade(ctx, conn, readMutex, restart, exit, sessionId, msgType, matches[4], msg)
	// Book
	case string(messages.ChannelBook):
		client.handleBook(ctx, conn, readMutex, restart, exit, sessionId, msgType, matches[4], msg)
	// Spread
	case string(messages.ChannelSpread):
		client.handleSpread(ctx, conn, readMutex, restart, exit, sessionId, msgType, matches[4], msg)
	// Ticker
	case string(messages.ChannelTicker):
		client.handleTicker(ctx, conn, readMutex, restart, exit, sessionId, msgType, matches[4], msg)
	// OHLC
	case string(messages.ChannelOHLC):
		// Extract interval
		if len(splits) > 0 {
			if interval, err := strconv.ParseInt(splits[1], 10, 64); err == nil {
				client.handleOHLC(ctx, conn, readMutex, restart, exit, sessionId, msgType, matches[4], msg, messages.IntervalEnum(interval))
			} else {
				err := fmt.Errorf("failed to parse interval for ohlc from '%s'", string(mType))
				tracing.HandleAndTraLogError(span, client.logger, err)
				client.OnReadError(ctx, conn, readMutex, restart, exit, err)
				return
			}
		} else {
			err := fmt.Errorf("failed to parse interval for ohlc from '%s'", string(mType))
			tracing.HandleAndTraLogError(span, client.logger, err)
			client.OnReadError(ctx, conn, readMutex, restart, exit, err)
			return
		}
	// Subscribe/Unsubscribe responses
	case string(messages.EventTypeSubscriptionStatus):
		client.handleSubscriptionStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Add order status
	case string(messages.EventTypeAddOrderStatus):
		client.handleAddOrderStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Edit order status
	case string(messages.EventTypeEditOrderStatus):
		client.handleEditOrderStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Cancel order status
	case string(messages.EventTypeCancelOrderStatus):
		client.handleCancelOrderStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Cancel all orders status
	case string(messages.EventTypeCancelAllOrderStatus):
		client.handleCancelAllOrdersStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Cancel all orders after X status
	case string(messages.EventTypeCancelAllOrderAfterXStatus):
		client.handleCancelAllOrdersAfterXStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Open orders
	case string(messages.ChannelOpenOrders):
		client.handleOpenOrders(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Owntrades
	case string(messages.ChannelOwnTrades):
		client.handleOwnTrades(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// System status
	case string(messages.EventTypeSystemStatus):
		client.handleSystemStatus(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Pong
	case string(messages.EventTypePong):
		client.handlePong(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	// Heartbeat
	case string(messages.EventTypeHeartbeat):
		client.handleHeartbeat(ctx, conn, readMutex, restart, exit, sessionId, msgType, msg)
	default:
		// Call OnReadError - Unknown message type
		eerr := fmt.Errorf("unkown or unexpected message type (%s) extracted from '%s'", mType, string(msg))
		tracing.HandleAndTraLogError(span, client.logger, eerr)
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
func (client *krakenSpotWebsocketClient) OnReadError(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	err error) {
	// Tracing: start span
	ctx, span := client.tracer.Start(ctx, "on_read_error", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	client.logger.Println("handling on read error: ", err.Error())
	// Call user callback if set
	if client.onReadErrorCallback != nil {
		client.onReadErrorCallback(ctx, restart, exit, err)
	}
}

// # Description
//
// Callback is called when the websocket connection is closed or about to be closed after a
// Stop method call or a call to the provided restart/exit functions. Callback is called once
// by the engine: the engine will not exit or restart until the callback has been completed.
//
// Callback can return an optional CloseMessageDetails which will be used to build the close
// message sent to the server if the connection needs to be closed after OnClose has finished.
// In such a case, if the returned value is nil, the engine will use 1001 "Going Away" as the
// close message.
//
// Do not close the websocket connection here if it is still open: It will be automatically
// closed by the engine with a close message.
//
// # Inputs
//
//   - ctx: Context produced from the websocket engine context and bound to OnClose lifecycle.
//   - conn: Connection to the websocket server that is closed or about to close.
//   - readMutex: A reference to engine read mutex user can lock to pause the engine.
//   - closeMessage: Websocket close message received from server or generated by the engine
//     when connection has been closed. If nil, connection might not be closed and will be
//     closed by the engine using the returned close message or the default 1001 "Going Away".
//
// # Returns
//
// A specific close message to send back to the server if connection has to be closed after
// this callback completes.
//
// # Warning
//
// Provided context will already be canceled.
func (client *krakenSpotWebsocketClient) OnClose(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	closeMessage *wsclient.CloseMessageDetails) *wsclient.CloseMessageDetails {
	// Tracing: start span
	ctx, span := client.tracer.Start(ctx, "on_close", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	client.logger.Println("handling on close")
	// Discard pending ping requests to unlock all blocked thread waiting for a response.
	client.logger.Println("discarding pending ping requests")
	client.pendingPingMu.Lock()
	defer client.pendingPingMu.Unlock()
	for reqid, req := range client.requests.pendingPing {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingPing, reqid)
		// Log
		client.logger.Println("pending ping requests discarded: ", reqid)
	}
	// Discard pending subscribe requests
	client.logger.Println("discarding pending subscribe requests")
	client.pendingSubscribeMu.Lock()
	defer client.pendingSubscribeMu.Unlock()
	for reqid, req := range client.requests.pendingSubscribe {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingSubscribe, reqid)
		// Log
		client.logger.Println("pending subscribe requests discarded: ", reqid)
	}
	// Discard pending unsubscribe requests
	client.logger.Println("discarding pending unsubscribe requests")
	client.pendingUnsubscribeMu.Lock()
	defer client.pendingUnsubscribeMu.Unlock()
	for reqid, req := range client.requests.pendingUnsubscribe {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingUnsubscribe, reqid)
		// Log
		client.logger.Println("pending unsubscribe requests discarded: ", reqid)
	}
	// Discard pending add order requests
	client.logger.Println("discarding pending add order requests")
	client.pendingAddOrderMu.Lock()
	defer client.pendingAddOrderMu.Unlock()
	for reqid, req := range client.requests.pendingAddOrderRequests {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingAddOrderRequests, reqid)
		// Log
		client.logger.Println("pending add order requests discarded: ", reqid)
	}
	// Discard pending edit order requests
	client.logger.Println("discarding pending edit order requests")
	client.pendingEditOrderMu.Lock()
	defer client.pendingEditOrderMu.Unlock()
	for reqid, req := range client.requests.pendingEditOrderRequests {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingEditOrderRequests, reqid)
		// Log
		client.logger.Println("pending edit order requests discarded: ", reqid)
	}
	// Discard pending cancel order requests
	client.logger.Println("discarding pending cancel order requests")
	client.pendingCancelOrderMu.Lock()
	defer client.pendingCancelOrderMu.Unlock()
	for reqid, req := range client.requests.pendingCancelOrderRequests {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingCancelOrderRequests, reqid)
		// Log
		client.logger.Println("pending cancel order requests discarded: ", reqid)
	}
	// Discard pending cancel all orders requests
	client.logger.Println("discarding pending cancel all orders requests")
	client.pendingCancelAllOrdersMu.Lock()
	defer client.pendingCancelAllOrdersMu.Unlock()
	for reqid, req := range client.requests.pendingCancelAllOrdersRequests {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingCancelAllOrdersRequests, reqid)
		// Log
		client.logger.Println("pending cancel all orders requests discarded: ", reqid)
	}
	// Discard pending cancel all orders requests
	client.logger.Println("discarding pending cancel all orders after x requests")
	client.pendingCancelAllOrdersAfterXOrderMu.Lock()
	defer client.pendingCancelAllOrdersAfterXOrderMu.Unlock()
	for reqid, req := range client.requests.pendingCancelAllOrdersAfterXRequests {
		// blocking write can be used as channels are managed internally and must have a capacity of 1
		req.err <- fmt.Errorf("connection has been closed")
		// Remove pending request
		delete(client.requests.pendingCancelAllOrdersAfterXRequests, reqid)
		// Log
		client.logger.Println("pending cancel all orders after x requests discarded: ", reqid)
	}
	// Send a connection interrupted event on all active subscriptions
	e := event.New()
	e.Context.SetType(string(events.ConnectionInterrupted))
	e.Context.SetID(uuid.NewString())
	e.Context.SetSource(tracing.PackageName)
	// Use blocking writes (design principle: wait 'till delivery)
	client.tickerSubMu.Lock()
	defer client.tickerSubMu.Unlock()
	if client.subscriptions.ticker != nil {
		client.logger.Println("sending a connection_interrupted event on ticker channel to warn about connection interruption")
		client.subscriptions.ticker.pub <- e
	}
	client.ohlcSubMu.Lock()
	defer client.ohlcSubMu.Unlock()
	for _, osub := range client.subscriptions.ohlcs {
		client.logger.Println("sending a connection_interrupted event on ohlc channel to warn about connection interruption", int(osub.interval))
		osub.pub <- e
	}
	client.tradeSubMu.Lock()
	defer client.tradeSubMu.Unlock()
	if client.subscriptions.trade != nil {
		client.logger.Println("sending a connection_interrupted event on trade channel to warn about connection interruption")
		client.subscriptions.trade.pub <- e
	}
	client.spreadSubMu.Lock()
	defer client.spreadSubMu.Unlock()
	if client.subscriptions.spread != nil {
		client.logger.Println("sending a connection_interrupted event on spread channel to warn about connection interruption")
		client.subscriptions.spread.pub <- e
	}
	client.bookSubMu.Lock()
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book != nil {
		client.logger.Println("sending a connection_interrupted event on book channels to warn about connection interruption")
		client.subscriptions.book.pub <- e
	}
	client.ownTradesSubMu.Lock()
	defer client.ownTradesSubMu.Unlock()
	if client.subscriptions.ownTrades != nil {
		client.logger.Println("sending a connection_interrupted event on own trades channel to warn about connection interruption")
		client.subscriptions.ownTrades.pub <- e
	}
	client.openOrdersSubMu.Lock()
	defer client.openOrdersSubMu.Unlock()
	if client.subscriptions.openOrders != nil {
		client.logger.Println("sending a connection_interrupted event on open orders channel to warn about connection interruption")
		client.subscriptions.openOrders.pub <- e
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
func (client *krakenSpotWebsocketClient) OnCloseError(
	ctx context.Context,
	err error) {
	fmt.Println("close error", err.Error())
	// Tracing: start span
	_, span := client.tracer.Start(ctx, "on_close_error",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("error", err.Error())))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	client.logger.Println("handling on close error: ", err.Error())
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
func (client *krakenSpotWebsocketClient) OnRestartError(
	ctx context.Context,
	exit context.CancelFunc,
	err error,
	retryCount int) {
	// Tracing: start span
	ctx, span := client.tracer.Start(ctx, "on_restart_error",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.Int("retry_count", retryCount),
			attribute.String("error", err.Error()),
		))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	client.logger.Println("handling on restart error: ", err.Error(), retryCount)
	// Call user callback if set
	if client.onRestartError != nil {
		client.onRestartError(ctx, exit, err, retryCount)
	}
}

/*************************************************************************************************/
/* MESSAGE HANDLERS                                                                              */
/*************************************************************************************************/

// This method contains the logic to handle a received general error message.
func (client *krakenSpotWebsocketClient) handleErrorMessage(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_error_message",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handing error message from server")
	// Parse message as error
	errMsg := new(messages.ErrorMessage)
	err := json.Unmarshal(msg, errMsg)
	if err != nil {
		// Call OnReadError - failed to parse message as error
		eerr := fmt.Errorf("failed to parse message '%s' as error message: %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Tracing: Add an event about error message
	attr := []attribute.KeyValue{
		attribute.String("error", errMsg.Err),
	}
	if errMsg.ReqId != nil {
		attr = append(attr, attribute.Int64("request_id", *errMsg.ReqId))
	}
	span.AddEvent("error_message", trace.WithAttributes(attr...))
	// If there is a joined request ID, check pending requests
	if errMsg.ReqId != nil {
		// Check pending subscribe
		client.pendingSubscribeMu.Lock()
		prSub := client.requests.pendingSubscribe[*errMsg.ReqId]
		if prSub != nil {
			// Fulfil request by publishing an error on the request error channel
			prSub.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingSubscribe, *errMsg.ReqId)
			// Unlock pending subscribe requests map & Exit
			client.pendingSubscribeMu.Unlock()
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		client.pendingSubscribeMu.Unlock()
		// Check pending addOrder
		client.pendingAddOrderMu.Lock()
		prAddOrder := client.requests.pendingAddOrderRequests[*errMsg.ReqId]
		if prAddOrder != nil {
			// Fulfil request by publishing an error on the request error channel
			prAddOrder.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingAddOrderRequests, *errMsg.ReqId)
			// Unlock pending add order requests map & Exit
			client.pendingAddOrderMu.Unlock()
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		client.pendingAddOrderMu.Unlock()
		// Check pending editOrder
		client.pendingEditOrderMu.Lock()
		prEditOrder := client.requests.pendingEditOrderRequests[*errMsg.ReqId]
		if prEditOrder != nil {
			// Fulfil request by publishing an error on the request error channel
			prEditOrder.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingEditOrderRequests, *errMsg.ReqId)
			// Unlock pending edit order requests map & Exit
			client.pendingEditOrderMu.Unlock()
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		client.pendingEditOrderMu.Unlock()
		// Check pending cancelOrder
		client.pendingCancelOrderMu.Lock()
		prCancelOrder := client.requests.pendingCancelOrderRequests[*errMsg.ReqId]
		if prCancelOrder != nil {
			// Fulfil request by publishing an error on the request error channel
			prCancelOrder.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingCancelOrderRequests, *errMsg.ReqId)
			// Unlock pending edit order requests map & Exit
			client.pendingCancelOrderMu.Unlock()
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		client.pendingCancelOrderMu.Unlock()
		// Check pending cancelAllOrders
		client.pendingCancelAllOrdersMu.Lock()
		prCancelAllOrders := client.requests.pendingCancelAllOrdersRequests[*errMsg.ReqId]
		if prCancelAllOrders != nil {
			// Fulfil request by publishing an error on the request error channel
			prCancelAllOrders.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingCancelAllOrdersRequests, *errMsg.ReqId)
			// Unlock pending edit order requests map & Exit
			client.pendingCancelAllOrdersMu.Unlock()
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		client.pendingCancelAllOrdersMu.Unlock()
		// Check pending cancelALlOrdersAfterX
		client.pendingCancelAllOrdersAfterXOrderMu.Lock()
		prCancelAllOrdersAfterX := client.requests.pendingCancelAllOrdersAfterXRequests[*errMsg.ReqId]
		if prCancelAllOrdersAfterX != nil {
			// Fulfil request by publishing an error on the request error channel
			prCancelAllOrdersAfterX.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingCancelAllOrdersAfterXRequests, *errMsg.ReqId)
			// Unlock pending edit order requests map & Exit
			client.pendingCancelAllOrdersAfterXOrderMu.Unlock()
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		client.pendingCancelAllOrdersAfterXOrderMu.Unlock()
		// Check pending unsubscribe
		client.pendingUnsubscribeMu.Lock()
		prUnsub := client.requests.pendingUnsubscribe[*errMsg.ReqId]
		if prUnsub != nil {
			// Fulfil request by publishing an error on the request error channel
			prUnsub.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingUnsubscribe, *errMsg.ReqId)
			// Unlock and exit
			client.pendingUnsubscribeMu.Unlock()
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		// Unlock pending unsubscribe requets map & Exit
		client.pendingUnsubscribeMu.Unlock()
		//  Check pending ping
		client.pendingPingMu.Lock()
		defer client.pendingPingMu.Lock()
		prPing := client.requests.pendingPing[*errMsg.ReqId]
		if prPing != nil {
			// Fulfil request by publish an error on the request error channel
			prPing.err <- fmt.Errorf("server replied with an error message: %s", errMsg.Err)
			// Discard the request
			delete(client.requests.pendingPing, *errMsg.ReqId)
			// Exit
			span.SetStatus(codes.Ok, codes.Ok.String())
			return nil
		}
		// Error no corresponding request
		eerr := fmt.Errorf("no corresponding pending request has been found for the request id %d to relay the following error: %s", *errMsg.ReqId, errMsg.Err)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Error no request ID -> As the cient force the usage of request IDs, not having one is
	// considered as an error.
	eerr := fmt.Errorf("no requests id for the following error message: %s", errMsg.Err)
	return tracing.HandleAndTraLogError(span, client.logger, eerr)
}

// This method contains the logic to handle a received heartbeat message.
func (client *krakenSpotWebsocketClient) handleHeartbeat(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_heartbeat",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling heartbeat from server")
	// Publish heartbeat - as user might not actively listen to heartbeats, manage the channel in FIFO
	// fashion by discarding oldest messages in case of congestion
	event := event.New()
	event.Context.SetType(string(events.Heartbeat))
	event.Context.SetSource(tracing.PackageName)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	select {
	case client.subscriptions.heartbeat <- event:
	default:
		// Discard oldest heartbeat & push new one
		<-client.subscriptions.heartbeat
		client.subscriptions.heartbeat <- event
	}
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received systemStatus message.
func (client *krakenSpotWebsocketClient) handleSystemStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	_, span := client.tracer.Start(ctx, "handle_system_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling system status from server")
	// Publish heartbeat - as user might not actively listen to system statuses, manage the channel
	// in FIFO fashion by discarding oldest messages in case of congestion
	event := event.New()
	event.Context.SetType(string(events.SystemStatus))
	event.Context.SetSource(tracing.PackageName)
	event.SetData("application/json", msg)
	select {
	case client.subscriptions.systemStatus <- event:
	default:
		// Discard oldest heartbeat & push new one
		<-client.subscriptions.systemStatus
		client.subscriptions.systemStatus <- event
	}
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received pong message.
func (client *krakenSpotWebsocketClient) handlePong(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_pong",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling pong from server")
	// Parse message as pong
	pong := new(messages.Pong)
	err := json.Unmarshal(msg, pong)
	if err != nil {
		// Call OnReadError - failed to parse message as pong
		eerr := fmt.Errorf("failed to parse message '%s' as pong: %w", string(msg), err)
		client.logger.Println(eerr.Error())
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Check if pong has a request ID.
	if pong.ReqId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received pong message has no request id")
		client.logger.Println(err.Error())
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Tracing: Add event for received pong
	span.AddEvent("pong", trace.WithAttributes(
		attribute.Int64("request_id", *pong.ReqId),
		attribute.String("session_id", sessionId),
	))
	// Extract pending ping request corresponding to the request ID
	client.pendingPingMu.Lock()
	defer client.pendingPingMu.Unlock()
	pr := client.requests.pendingPing[*pong.ReqId]
	if pr == nil {
		// Call OnRead error: as user defined request ids must be used. Not a corresponding
		// pending request is considered as an error
		err := fmt.Errorf("received pong has no corresponding pending ping request for id: %d", *pong.ReqId)
		client.logger.Println(err.Error())
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Fulfil pending request
	// Blocking write can be used as channel must always have a capacity of one and be internally managed
	pr.resp <- pong
	// Discard pending request now that it has been served and exit
	client.logger.Println("pong handled")
	delete(client.requests.pendingPing, *pong.ReqId)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received subscriptionStatus message as a response for
// either Subscribe or Unsubscribe.
func (client *krakenSpotWebsocketClient) handleSubscriptionStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_subscription_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling subscription status from server")
	// Parse message as SubscriptionStatus
	subs := new(messages.SubscriptionStatus)
	err := json.Unmarshal(msg, subs)
	if err != nil {
		// Call OnReadError - failed to parse message as SubscriptionStatus
		eerr := fmt.Errorf("failed to parse message '%s' as subscriptionStatus: %w", string(msg), err)
		client.logger.Println(eerr.Error())
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Check if there is a request ID.
	if subs.ReqId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received subscriptionStatus message has no request id")
		client.logger.Println(err.Error())
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
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
	span.AddEvent("subscription_status", trace.WithAttributes(attr...))
	// Extract pending subscribe request corresponding to the request ID
	client.pendingSubscribeMu.Lock()
	defer client.pendingSubscribeMu.Unlock()
	subreq := client.requests.pendingSubscribe[*subs.ReqId]
	if subreq == nil {
		// Check unsubscribe
		client.pendingUnsubscribeMu.Lock()
		defer client.pendingUnsubscribeMu.Unlock()
		unsubreq := client.requests.pendingUnsubscribe[*subs.ReqId]
		if unsubreq == nil {
			// Call OnRead error: as user defined request ids must be used. Not a corresponding
			// pending request is considered as an error
			err := fmt.Errorf("received suscriptionStatus has no corresponding pending request for id: %d", *subs.ReqId)
			client.logger.Println(err.Error())
			client.OnReadError(ctx, conn, readMutex, restart, exit, err)
			return tracing.HandleAndTraLogError(span, client.logger, err)
		}
		// Check if the message has an error message and record it if that is the case
		if subs.Status == string(messages.Err) {
			unsubreq.errPerPair[subs.Pair] = fmt.Errorf("unsubscribe for %s failed: %s", subs.Pair, subs.Err)
			tracing.HandleAndTraLogError(span, client.logger, err)
		}
		// Mark the pair as served
		unsubreq.served[subs.Pair] = true
		// Check if a response has been received for each requested pair. If that is the case fulfil the request.
		// Otherwise, do nothing and wait for more responses from the server
		fully := true
		for _, v := range unsubreq.pairs {
			// fully will remain true only if all requests have been served ;)
			fully = fully && unsubreq.served[v]
		}
		if fully {
			// Fulfil pending unsubscribe: send nil in case of success or an error with the error message if
			// unsubscribe has failed.
			err = nil
			if len(unsubreq.errPerPair) > 0 {
				// Trace error
				err = &SubscriptionError{
					Errs: unsubreq.errPerPair,
				}
				client.logger.Println(err.Error())
				tracing.HandleAndTraLogError(span, client.logger, err)
			}
			// Blocking write can be used as channel must always have a capacity of one and be internally managed
			unsubreq.err <- err
			// Discard pending request
			delete(client.requests.pendingUnsubscribe, *subs.ReqId)
		}
	} else {
		// Check if the message has an error message and record it if that is the case
		if subs.Status == string(messages.Err) {
			subreq.errPerPair[subs.Pair] = fmt.Errorf("subscribe for %s failed: %s", subs.Pair, subs.Err)
			tracing.HandleAndTraLogError(span, client.logger, err)
		}
		// Mark the pair as served
		subreq.served[subs.Pair] = true
		// Check if a response has been received for each requested pair. If that is the case fulfil the request.
		// Otherwise, do nothing and wait for more responses from the server
		fully := true
		for _, v := range subreq.pairs {
			// fully will remain true only if all requests have been served ;)
			fully = fully && subreq.served[v]
		}
		if fully {
			// Fulfil pending subscribe: send nil in case of success or an error with the error message if
			// subscribe has failed
			err = nil
			if len(subreq.errPerPair) > 0 {
				err = &SubscriptionError{
					Errs: subreq.errPerPair,
				}
				client.logger.Println(err.Error())
				tracing.HandleAndTraLogError(span, client.logger, err)
			}
			// Blocking write can be used as channel must always have a capacity of one and be internally managed
			subreq.err <- err
			// Discard pending request
			delete(client.requests.pendingSubscribe, *subs.ReqId)
		}
	}
	// Exit
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received ticker message.
func (client *krakenSpotWebsocketClient) handleTicker(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	pair string,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_ticker",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling ticker message from server")
	// Check if there is an active subscription, discard otherwise
	client.tickerSubMu.Lock()
	defer client.tickerSubMu.Unlock()
	if client.subscriptions.ticker == nil {
		err := fmt.Errorf("a ticker message has been received while there is no active subscription to ticker channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish ticker - use blocking write (block until delivery)
	event := event.New()
	event.Context.SetType(string(events.Ticker))
	event.Context.SetSource(tracing.PackageName)
	event.SetSubject(pair)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.ticker.pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received ohlc message.
func (client *krakenSpotWebsocketClient) handleOHLC(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	pair string,
	msg []byte,
	interval messages.IntervalEnum) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_ohlc",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling ohlc message from server")
	// Check if there is an active subscription, discard otherwise
	client.ohlcSubMu.Lock()
	defer client.ohlcSubMu.Unlock()
	if client.subscriptions.ohlcs == nil {
		err := fmt.Errorf("a ohlc message has been received while there is no active subscription to ohlc channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish ohlc - use blocking write (block until delivery)
	event := event.New()
	event.Context.SetType(string(events.OHLC))
	event.Context.SetSource(tracing.PackageName)
	event.SetSubject(pair)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.ohlcs[messages.IntervalEnum(interval)].pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received trade message.
func (client *krakenSpotWebsocketClient) handleTrade(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	pair string,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_trade",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling trade message from server")
	// Check if there is an active subscription, discard otherwise
	client.tradeSubMu.Lock()
	defer client.tradeSubMu.Unlock()
	if client.subscriptions.trade == nil {
		err := fmt.Errorf("a trade message has been received while there is no active subscription to trade channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish trade - use blocking write (block until delivery)
	event := event.New()
	event.Context.SetType(string(events.Trade))
	event.Context.SetSource(tracing.PackageName)
	event.SetSubject(pair)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.trade.pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received spread message.
func (client *krakenSpotWebsocketClient) handleSpread(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	pair string,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_spread",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling spread message from server")
	// Check if there is an active subscription, discard otherwise
	client.spreadSubMu.Lock()
	defer client.spreadSubMu.Unlock()
	if client.subscriptions.spread == nil {
		err := fmt.Errorf("a spread message has been received while there is no active subscription to spread channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish trade - use blocking write
	event := event.New()
	event.Context.SetType(string(events.Spread))
	event.Context.SetSource(tracing.PackageName)
	event.SetSubject(pair)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.spread.pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received book message.
func (client *krakenSpotWebsocketClient) handleBook(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	pair string,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_book",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	defer span.SetStatus(codes.Ok, codes.Ok.String())
	client.logger.Println("handling book message from server")
	// Check if it is a snapshot or an update -> an update will have a "c" field
	if strings.Contains(string(msg), `"c"`) {
		// Handle update
		return client.handleBookUpdate(ctx, conn, readMutex, restart, exit, sessionId, msgType, pair, msg)
	}
	// Hanlde snapshot
	return client.handleBookSnapshot(ctx, conn, readMutex, restart, exit, sessionId, msgType, pair, msg)
}

// This method contains the logic to handle a received book update message.
func (client *krakenSpotWebsocketClient) handleBookUpdate(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	pair string,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_book_update",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling book update message from server")
	// Check if there is an active subscription, discard otherwise
	client.bookSubMu.Lock()
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book == nil {
		err := fmt.Errorf("a book update message has been received while there is no active subscription to book channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish book update - use blocking write
	event := event.New()
	event.Context.SetType(string(events.BookUpdate))
	event.Context.SetSource(tracing.PackageName)
	event.SetSubject(pair)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.book.pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received book snapshot message.
func (client *krakenSpotWebsocketClient) handleBookSnapshot(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	pair string,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_book_snapshot",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling book snapshot message from server")
	// Check if there is an active subscription, discard otherwise
	client.bookSubMu.Lock()
	defer client.bookSubMu.Unlock()
	if client.subscriptions.book == nil {
		err := fmt.Errorf("a book snapshot message has been received while there is no active subscription to book channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish book snapshot - use blocking write (wait till delivery)
	event := event.New()
	event.Context.SetType(string(events.BookSnapshot))
	event.Context.SetSource(tracing.PackageName)
	event.SetSubject(pair)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.book.pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received own trades message.
func (client *krakenSpotWebsocketClient) handleOwnTrades(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_own_trades",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling own trades message from server")
	// Check if there is an active subscription, discard otherwise
	client.ownTradesSubMu.Lock()
	defer client.ownTradesSubMu.Unlock()
	if client.subscriptions.ownTrades == nil {
		err := fmt.Errorf("a own trades message has been received while there is no active subscription to own trades channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish own trades - use blocking write (wait till delivery)
	event := event.New()
	event.Context.SetType(string(events.OwnTrades))
	event.Context.SetSource(tracing.PackageName)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.ownTrades.pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received open orders message.
func (client *krakenSpotWebsocketClient) handleOpenOrders(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_open_orders",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling open orders message from server")
	// Check if there is an active subscription, discard otherwise
	client.openOrdersSubMu.Lock()
	defer client.openOrdersSubMu.Unlock()
	if client.subscriptions.openOrders == nil {
		err := fmt.Errorf("a open orders message has been received while there is no active subscription to open orders channel")
		client.logger.Println(err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Publish own trades - use blocking write (wait till delivery)
	event := event.New()
	event.Context.SetType(string(events.OpenOrders))
	event.Context.SetSource(tracing.PackageName)
	event.SetData("application/json", msg)
	otelObs.InjectDistributedTracingExtension(ctx, event)
	client.subscriptions.openOrders.pub <- event
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received add order status message.
func (client *krakenSpotWebsocketClient) handleAddOrderStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_add_order_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling add order status message from server")
	// Parse message as AddOrderResponse
	aos := new(messages.AddOrderResponse)
	err := json.Unmarshal(msg, aos)
	if err != nil {
		// Call OnReadError - failed to parse message as addOrderResponse
		eerr := fmt.Errorf("failed to parse message '%s' as add order response : %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Check if add order response has a request ID.
	if aos.RequestId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received add order response message has no request id")
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Tracing: Add event for received add order response
	span.AddEvent("add_order_status", trace.WithAttributes(
		attribute.String("status", aos.Status),
		attribute.String("txid", aos.TxId),
		attribute.String("description", aos.Description),
		attribute.String("error", aos.Err),
		attribute.Int64("request_id", *aos.RequestId),
		attribute.String("session_id", sessionId),
	))
	// Extract pending add order request corresponding to the request ID
	client.pendingAddOrderMu.Lock()
	defer client.pendingAddOrderMu.Unlock()
	pr := client.requests.pendingAddOrderRequests[*aos.RequestId]
	if pr == nil {
		// Call OnRead error: as user defined request ids must be used. Not having a corresponding
		// pending request is considered as an error
		err := fmt.Errorf("received add order response has no corresponding pending add order request for id: %d", *aos.RequestId)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Fulfil pending request
	// Blocking write can be used as channel must always have a capacity of one and be internally managed
	pr.resp <- aos
	// Discard pending request now that it has been served and exit
	delete(client.requests.pendingAddOrderRequests, *aos.RequestId)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received edit order status message.
func (client *krakenSpotWebsocketClient) handleEditOrderStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_edit_order_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling edit order status message from server")
	// Parse message as EditORderResponse
	eo := new(messages.EditOrderResponse)
	err := json.Unmarshal(msg, eo)
	if err != nil {
		// Call OnReadError - failed to parse message as editOrderResponse
		eerr := fmt.Errorf("failed to parse message '%s' as edit order response : %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Check if edit order response has a request ID.
	if eo.RequestId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received edit order response message has no request id")
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Tracing: Add event for received edit order response
	span.AddEvent("edit_order_status", trace.WithAttributes(
		attribute.String("status", eo.Status),
		attribute.String("txid", eo.TxId),
		attribute.String("original_txid", eo.OriginalTxId),
		attribute.String("description", eo.Description),
		attribute.String("error", eo.Err),
		attribute.Int64("request_id", *eo.RequestId),
		attribute.String("session_id", sessionId),
	))
	// Extract pending add order request corresponding to the request ID
	client.pendingEditOrderMu.Lock()
	defer client.pendingEditOrderMu.Unlock()
	pr := client.requests.pendingEditOrderRequests[*eo.RequestId]
	if pr == nil {
		// Call OnRead error: as user defined request ids must be used. Not having a corresponding
		// pending request is considered as an error
		err := fmt.Errorf("received edit order response has no corresponding pending edit order request for id: %d", *eo.RequestId)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Fulfil pending request
	// Blocking write can be used as channel must always have a capacity of one and be internally managed
	pr.resp <- eo
	// Discard pending request now that it has been served and exit
	delete(client.requests.pendingEditOrderRequests, *eo.RequestId)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received cancel order status message.
func (client *krakenSpotWebsocketClient) handleCancelOrderStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_cancel_order_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling cancel order status message from server")
	// Parse message as CancelOrderResponse
	co := new(messages.CancelOrderResponse)
	err := json.Unmarshal(msg, co)
	if err != nil {
		// Call OnReadError - failed to parse message as cancelOrderResponse
		eerr := fmt.Errorf("failed to parse message '%s' as cancel order response : %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Check if edit order response has a request ID.
	if co.RequestId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received cancel order response message has no request id")
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Tracing: Add event for received cancel order response
	span.AddEvent("cancel_order_status", trace.WithAttributes(
		attribute.String("status", co.Status),
		attribute.String("error", co.Err),
		attribute.Int64("request_id", *co.RequestId),
		attribute.String("session_id", sessionId),
	))
	// Extract pending add order request corresponding to the request ID
	client.pendingCancelOrderMu.Lock()
	defer client.pendingCancelOrderMu.Unlock()
	pr := client.requests.pendingCancelOrderRequests[*co.RequestId]
	if pr == nil {
		// Call OnRead error: as user defined request ids must be used. Not having a corresponding
		// pending request is considered as an error
		err := fmt.Errorf("received cancel order response has no corresponding pending cancel order request for id: %d", *co.RequestId)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Fulfil pending request
	// Blocking write can be used as channel must always have a capacity of one and be internally managed
	pr.resp <- co
	// Discard pending request now that it has been served and exit
	delete(client.requests.pendingCancelOrderRequests, *co.RequestId)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received cancel all orders status message.
func (client *krakenSpotWebsocketClient) handleCancelAllOrdersStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_cancel_all_orders_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling cancel all orders status message from server")
	// Parse message as CancelAllOrdersResponse
	co := new(messages.CancelAllOrdersResponse)
	err := json.Unmarshal(msg, co)
	if err != nil {
		// Call OnReadError - failed to parse message as cancelAllOrdersResponse
		eerr := fmt.Errorf("failed to parse message '%s' as cancel all orders response : %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Check if cancel all orders response has a request ID.
	if co.RequestId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received cancel all orders response message has no request id")
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Tracing: Add event for received cancel all orders response
	span.AddEvent("cancel_all_orders_status", trace.WithAttributes(
		attribute.String("status", co.Status),
		attribute.String("error", co.Err),
		attribute.Int64("request_id", *co.RequestId),
		attribute.String("session_id", sessionId),
	))
	// Extract pending cancel all orders request corresponding to the request ID
	client.pendingCancelAllOrdersMu.Lock()
	defer client.pendingCancelAllOrdersMu.Unlock()
	pr := client.requests.pendingCancelAllOrdersRequests[*co.RequestId]
	if pr == nil {
		// Call OnRead error: as user defined request ids must be used. Not having a corresponding
		// pending request is considered as an error
		err := fmt.Errorf("received cancel all orders response has no corresponding pending cancel all orders request for id: %d", *co.RequestId)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Fulfil pending request
	// Blocking write can be used as channel must always have a capacity of one and be internally managed
	pr.resp <- co
	// Discard pending request now that it has been served and exit
	delete(client.requests.pendingCancelAllOrdersRequests, *co.RequestId)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// This method contains the logic to handle a received cancel all orders status message.
func (client *krakenSpotWebsocketClient) handleCancelAllOrdersAfterXStatus(
	ctx context.Context,
	conn wsadapters.WebsocketConnectionAdapterInterface,
	readMutex *sync.Mutex,
	restart context.CancelFunc,
	exit context.CancelFunc,
	sessionId string,
	msgType wsadapters.MessageType,
	msg []byte) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "handle_cancel_all_orders_after_x_status",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("session_id", sessionId)))
	defer span.End()
	client.logger.Println("handling cancel all orders after x status message from server")
	// Parse message as CancelAllOrdersAfterXResponse
	co := new(messages.CancelAllOrdersAfterXResponse)
	err := json.Unmarshal(msg, co)
	if err != nil {
		// Call OnReadError - failed to parse message as CancelAllOrdersAfterXResponse
		eerr := fmt.Errorf("failed to parse message '%s' as cancel all orders after x response : %w", string(msg), err)
		client.OnReadError(ctx, conn, readMutex, restart, exit, eerr)
		return tracing.HandleAndTraLogError(span, client.logger, eerr)
	}
	// Check if cancel all orders after x response has a request ID.
	if co.RequestId == nil {
		// Call OnRead error: user defined request ids must be used. Not having one in responses
		// is considered as an error.
		err := fmt.Errorf("received cancel all orders after x response message has no request id")
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Tracing: Add event for received cancel all orders after x response
	span.AddEvent("cancel_all_orders_after_x_status", trace.WithAttributes(
		attribute.String("status", co.Status),
		attribute.String("current_time", co.CurrentTime),
		attribute.String("trigger_time", co.TriggerTime),
		attribute.String("error", co.Err),
		attribute.Int64("request_id", *co.RequestId),
		attribute.String("session_id", sessionId),
	))
	// Extract pending cancel all orders after x request corresponding to the request ID
	client.pendingCancelAllOrdersAfterXOrderMu.Lock()
	defer client.pendingCancelAllOrdersAfterXOrderMu.Unlock()
	pr := client.requests.pendingCancelAllOrdersAfterXRequests[*co.RequestId]
	if pr == nil {
		// Call OnRead error: as user defined request ids must be used. Not having a corresponding
		// pending request is considered as an error
		err := fmt.Errorf("received cancel all orders after x response has no corresponding pending cancel all orders after x request for id: %d", *co.RequestId)
		client.OnReadError(ctx, conn, readMutex, restart, exit, err)
		return tracing.HandleAndTraLogError(span, client.logger, err)
	}
	// Fulfil pending request
	// Blocking write can be used as channel must always have a capacity of one and be internally managed
	pr.resp <- co
	// Discard pending request now that it has been served and exit
	delete(client.requests.pendingCancelAllOrdersAfterXRequests, *co.RequestId)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

/*************************************************************************************************/
/* PRIVATE METHODS                                                                               */
/*************************************************************************************************/

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
func (client *krakenSpotWebsocketClient) sendSubscribeRequest(ctx context.Context, req *messages.Subscribe, errChan chan error) error {
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
	ctx, span := client.tracer.Start(ctx, "send_subscribe_request",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttr...))
	defer span.End()
	client.logger.Println("send subscribe request for: ", req.Subscription.Name)
	// Add pending susbcribe request to client's stack
	client.pendingSubscribeMu.Lock() // Lock to not add requests while engine is discarding pending requests
	defer client.pendingSubscribeMu.Unlock()
	client.requests.pendingSubscribe[req.ReqId] = &pendingSubscribe{
		pairs:      req.Pairs,
		served:     map[string]bool{},
		errPerPair: map[string]error{},
		err:        errChan,
	}
	// Marshal to JSON
	payload, err := json.Marshal(req)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingSubscribe, req.ReqId)
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("failed to format subscribe request: %w", err))
	}
	// Send message to websocket server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingSubscribe, req.ReqId)
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("failed to send subscribe request: %w", err))
	}
	// Set span status and exit
	client.logger.Println("subscribe request sent for: ", req.Subscription.Name)
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
func (client *krakenSpotWebsocketClient) sendUnsubscribeRequest(ctx context.Context, req *messages.Unsubscribe, errChan chan error) error {
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
	ctx, span := client.tracer.Start(ctx, "send_unsubscribe_request",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttr...))
	defer span.End()
	// Add pending unsusbcribe request to client's stack
	client.pendingUnsubscribeMu.Lock() // Lock to not add requests while engine is discarding pending requests
	defer client.pendingUnsubscribeMu.Unlock()
	client.requests.pendingUnsubscribe[req.ReqId] = &pendingUnsubscribe{
		pairs:      req.Pairs,
		served:     map[string]bool{},
		errPerPair: map[string]error{},
		err:        errChan,
	}
	client.logger.Println("send unsubscribe request for: ", req.Subscription.Name)
	// Marshal to JSON
	payload, err := json.Marshal(req)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingUnsubscribe, req.ReqId)
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("failed to format unsubscribe request: %w", err))
	}
	// Send message to websocket server
	err = client.conn.Write(ctx, wsadapters.Text, payload)
	if err != nil {
		// Remove pending request as it has failed before it even starts
		delete(client.requests.pendingUnsubscribe, req.ReqId)
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("failed to send unsubscribe request: %w", err))
	}
	// Set span status and exit
	client.logger.Println("unsubscribe request sent for: ", req.Subscription.Name)
	span.SetStatus(codes.Ok, codes.Ok.String())
	return nil
}

// # Description
//
// Resubscribe to the ticker channel.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//
// # Return
//
// An error is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel - OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) resubscribeTicker(ctx context.Context, pairs []string) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "resubscribe_ticker",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
		))
	defer span.End()
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
		fmt.Println("resubscribe failed", err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe ticker failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error - Use an operation itnerrupted error as request has been sent to the server
		fmt.Println("resubscribe failed", err.Error())
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "resubscribe_ticker", Root: fmt.Errorf("subscribe ticker failed: %w", err)})
	case err := <-errChan:
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already subscribed") {
			fmt.Println("resubscribe failed", err.Error())
			// Trace and return error - Use an operation error as the error was caused by an error emssage from the server.
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "resubscribe_ticker", Root: fmt.Errorf("subscribe ticker failed: %w", err)})
		}
		// Exit - Success
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Resubscribe to the ohlc channel.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel
//     will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//   - interval: The desired interval for OHLC indicators. Multiple subscriptions can be
//     maintained for different intervals.
//
// # Return
//
// An error is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel - OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) resubscribeOHLC(ctx context.Context, pairs []string, interval messages.IntervalEnum) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "resubscribe_ohlc",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("interval", int(interval)),
		))
	defer span.End()
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe ohlc failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "resubscribe_ohlc", Root: fmt.Errorf("resubscribe ohlc failed: %w", err)})
	case err := <-errChan:
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already subscribed") {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "resubscribe_ohlc", Root: fmt.Errorf("resubscribe ohlc failed: %w", err)})
		}
		// Exit - success
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Resubscribe to the trade channel.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//
// # Return
//
// An error is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel - OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) resubscribeTrade(ctx context.Context, pairs []string) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "resubscribe_trade",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
		))
	defer span.End()
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe trade failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "resubscribe_trade", Root: fmt.Errorf("resubscribe trade failed: %w", err)})
	case err := <-errChan:
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already subscribed") {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "resubscribe_trade", Root: fmt.Errorf("resubscribe trade failed: %w", err)})
		}
		// Exit - success
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Subscribe to the spread channel.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//
// # Return
//
// An error is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel - OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) resubscribeSpread(ctx context.Context, pairs []string) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "resubscribe_spread",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
		))
	defer span.End()
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe spread failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "resubscribe_spread", Root: fmt.Errorf("resubscribe spread failed: %w", err)})
	case err := <-errChan:
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already subscribed") {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "resubscribe_spread", Root: fmt.Errorf("resubscribe spread failed: %w", err)})
		}
		// Exit - success
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Resubscribe to the book channel.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
//   - depth: Desired book depth. Multiple subscriptions can be maintained for different depths.

// # Return
//
// An error is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel - OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) resubscribeBook(ctx context.Context, pairs []string, depth messages.DepthEnum) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "resubscribe_book",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.StringSlice("pairs", pairs),
			attribute.Int("depth", int(depth)),
		))
	defer span.End()
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
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe book failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "resubscribe_book", Root: fmt.Errorf("resubscribe book failed: %w", err)})
	case err := <-errChan:
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already subscribed") {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "resubscribe_book", Root: fmt.Errorf("resubscribe book failed: %w", err)})
		}
		// Exit - Success
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Resubscribe to the own trades channel.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - snapshot: Publish a snapshot of the recent own trades.
//   - consolidateTaker: Consolidate trades by taker.

// # Return
//
// An error is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel - OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) resubscribeOwnTrades(ctx context.Context, snapshot bool, consolidateTaker bool) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "resubscribe_own_trades",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Bool("snapshot", snapshot),
			attribute.Bool("consolidate_taker", consolidateTaker),
		))
	defer span.End()
	// Create response channels
	errChan := make(chan error, 1)
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe own trades failed: %w", err))
	}
	// Send subscribe message to server
	err = client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Subscription: messages.SuscribeDetails{
				Name:             string(messages.ChannelOwnTrades),
				Snapshot:         &snapshot,
				ConsolidateTaker: &consolidateTaker,
				Token:            token,
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe own trades failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "resubscribe_own_trades", Root: fmt.Errorf("resubscribe own trades failed: %w", err)})
	case err := <-errChan:
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already subscribed") {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "resubscribe_own_trades", Root: fmt.Errorf("resubscribe own trades failed: %w", err)})
		}
		// Exit - Success
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// Resubscribe to the open orders channel.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
//   - rateCounter: Include rate limit updates in messages.

// # Return
//
// An error is returned when:
//
//   - An error occurs when sending the subscription message.
//   - The provided context expires (timeout/cancel - OperationInterruptedError).
//   - An error message is received from the server (OperationError).
func (client *krakenSpotWebsocketClient) resubscribeOpenOrders(ctx context.Context, rateCounter bool) error {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "resubscribe_open_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Bool("rate_counter", rateCounter),
		))
	defer span.End()
	// Create response channels
	errChan := make(chan error, 1)
	// Get websocket token
	token, err := client.getWebsocketToken(ctx)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe open orders failed: %w", err))
	}
	// Send subscribe message to server
	err = client.sendSubscribeRequest(
		ctx,
		&messages.Subscribe{
			Event: string(messages.EventTypeSubscribe),
			ReqId: client.ngen.GenerateNonce(),
			Subscription: messages.SuscribeDetails{
				Name:        string(messages.ChannelOpenOrders),
				RateCounter: rateCounter,
				Token:       token,
			},
		},
		errChan)
	if err != nil {
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("resubscribe open orders failed: %w", err))
	}
	// Wait for response to be published on channels or timeout
	select {
	case <-ctx.Done():
		// Trace and return error
		return tracing.HandleAndTraLogError(span, client.logger, &OperationInterruptedError{Operation: "resubscribe_open_orders", Root: fmt.Errorf("resubscribe open orders failed: %w", err)})
	case err := <-errChan:
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already subscribed") {
			// Trace and return error
			return tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "resubscribe_open_orders", Root: fmt.Errorf("resubscribe open orders failed: %w", err)})
		}
		// Exit - Success
		span.SetStatus(codes.Ok, codes.Ok.String())
		return nil
	}
}

// # Description
//
// This method manages the websocket token used by the private websocket client:
//   - If token is empty or if the cached token has expired, the method will fetch a new one.
//   - If there is a cached, valid token, the method returns it
//
// # Inputs
//
//   - ctx: Context used for tracing/coordination purpose
//
// # Return
//
// The token or an error if any has occured. An error will be returned when:
//
//   - The provided context has expired
//   - The request could not be sent (formatting or connection issue)
//   - The server replied with an error (OperationError)
func (client *krakenSpotWebsocketClient) getWebsocketToken(ctx context.Context) (string, error) {
	// Tracing: Start span
	ctx, span := client.tracer.Start(ctx, "get_websocket_token", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Acquire token mutex
	client.tokenMu.Lock()
	defer client.tokenMu.Unlock()
	// Check if a token is cached and is still valid
	now := time.Now()
	if client.token == "" || client.tokenExpiresAt.Compare(now) >= 0 {
		// Acquire a new token
		client.logger.Println("requesting new websocket token")
		resp, _, err := client.restClient.GetWebsocketToken(ctx, client.cgen.GenerateNonce(), client.secopts)
		if err != nil {
			// Trace and return error
			return "", tracing.HandleAndTraLogError(span, client.logger, fmt.Errorf("get websocket token failed: %w", err))
		}
		if len(resp.Error) > 0 || resp.Result == nil {
			// Trace and return error
			return "", tracing.HandleAndTraLogError(span, client.logger, &OperationError{Operation: "get_websocket_token", Root: fmt.Errorf("get websocket token failed: %v", resp.Error)})
		}
		// Cache token & set expire (substract 5 seconds to be sure to refresh the token before it really expire)
		client.token = resp.Result.Token
		client.tokenExpiresAt = now.Add(time.Duration(resp.Result.Expires-5) * time.Second)
		client.logger.Println("websocket token refreshed")
	}
	// Return cached token
	span.SetStatus(codes.Ok, codes.Ok.String())
	return client.token, nil
}
