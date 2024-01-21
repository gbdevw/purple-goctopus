package websocket

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/gbdevw/gowse/wscengine"
	"github.com/gbdevw/gowse/wscengine/wsadapters/gorilla"
	"github.com/gbdevw/gowse/wscengine/wsclient"
	"go.opentelemetry.io/otel/trace"
)

// Kraken spot websocket client with access to public endpoints
type KrakenSpotPublicWebsocketClient struct {
	// Underlying base kraken spot websocket client
	*krakenSpotWebsocketClient
}

// # Description
//
// Factory which creates a KrakenSpotPublicWebsocketClient that can be provided to a websocket
// engine (wscengine.WebsocketEngine - Cf. https://github.com/gbdevw/gowse).
//
// # Inputs
//
//   - onCloseCallback: optional user defined callback which will be called when connection is closed/interrupted.
//   - onReadErrorCallback: optional user defined callback which will be called when an error occurs while reading messages from the websocket server
//   - onRestartError: optional user defined callback which will be called when the websocket engine fails to reconnect to the server.
//   - logger: Optional logger used to log debug/vebrose messages. If nil, a logger with a discard writer (noop) will be used
//   - tracerProvider: Tracer provider to use to get a tracer to instrument websocket client code. If nil, global tracer provider will be used.
//
// # Return
//
// A new KrakenSpotPublicWebsocketClient
func NewKrakenSpotPublicWebsocketClient(
	onCloseCallback func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails),
	onReadErrorCallback func(ctx context.Context, restart context.CancelFunc, exit context.CancelFunc, err error),
	onRestartError func(ctx context.Context, exit context.CancelFunc, err error, retryCount int),
	logger *log.Logger,
	tracerProvider trace.TracerProvider) *KrakenSpotPublicWebsocketClient {
	// Build & return public websocket client
	return &KrakenSpotPublicWebsocketClient{
		krakenSpotWebsocketClient: newKrakenSpotWebsocketClient(
			nil,
			nil,
			nil,
			onCloseCallback,
			onReadErrorCallback,
			onRestartError,
			logger,
			tracerProvider)}
}

// # Description
//
// Create a new KrakenSpotPublicWebsocketClient and a websocket engine to run it. The client and
// the engine will be configured with the following default options:
//
//   - Target URL for the websocket client will be the production env.: wss://ws.kraken.com
//   - The gorilla websocket framework with default settings will be used by the wesocket engine.
//   - Websocket engine settings: 4 workers, auto-reconnect enabled, 5sec exponential retry delay.
//
// The function will return the unstarted engine and the public websocket client attached to it.
//
// # Hint
//
// This functions eliminates a lot of the boiler plate code needed to create and gather components
// required to interact with Kraken websocket API:
//
//   - The websocket server URL.
//   - The public websocket client for Kraken API.
//   - The websocket connection adapter used by the engine to manage the websocket connection.
//   - The engine to run the private websocket client.
//
// # Inputs
//
//   - onCloseCallback: Optional callback called when connection is lost/stopped.
//   - onReadErrorCallback: Optional callback called when engine fails to read a message.
//   - onRestartError: Optional callback called when engine fails to reconnect to the server.
//   - logger: Optional logger used to log debug/vebrose messages. If nil, a logger with a discard writer (noop) will be used
//   - tracerProvider: Tracer provider to use to get a tracer to instrument websocket client code. If nil, global tracer provider will be used.
//
// # Returns
//
// In case of success, a ready to start websocket engine is returned along with the private websocket
// bound to the engine.
func NewDefaultEngineWithPublicWebsocketClient(
	onCloseCallback func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails),
	onReadErrorCallback func(ctx context.Context, restart context.CancelFunc, exit context.CancelFunc, err error),
	onRestartError func(ctx context.Context, exit context.CancelFunc, err error, retryCount int),
	logger *log.Logger,
	tracerProvider trace.TracerProvider,
) (*wscengine.WebsocketEngine, KrakenSpotPublicWebsocketClientInterface, error) {
	// Build websocket server URL
	url, err := url.Parse(KrakenSpotWebsocketPublicProductionURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s as a URL: %w", KrakenSpotWebsocketPublicProductionURL, err)
	}
	// Build websocket client
	wsclient := NewKrakenSpotPublicWebsocketClient(onCloseCallback, onReadErrorCallback, onRestartError, logger, tracerProvider)
	// Build engine options
	defopts := &wscengine.WebsocketEngineConfigurationOptions{
		ReaderRoutinesCount:                4,
		AutoReconnect:                      true,
		AutoReconnectRetryDelayBaseSeconds: 5,
		AutoReconnectRetryDelayMaxExponent: 3,
		OnOpenTimeoutMs:                    300000,
		StopTimeoutMs:                      300000,
	}
	// Build the engine that will power the wesocket client - Use default options and a gorilla based connection
	engine, err := wscengine.NewWebsocketEngine(url, gorilla.NewGorillaWebsocketConnectionAdapter(nil, nil), wsclient, defopts, tracerProvider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build the websocket engine: %w", err)
	}
	return engine, wsclient, nil
}
