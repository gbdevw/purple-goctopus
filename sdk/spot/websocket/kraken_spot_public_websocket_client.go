package websocket

import (
	"context"
	"log"

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
