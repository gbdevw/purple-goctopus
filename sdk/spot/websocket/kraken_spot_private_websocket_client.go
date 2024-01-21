package websocket

import (
	"context"
	"fmt"
	"log"

	"github.com/gbdevw/gowse/wscengine/wsclient"
	"github.com/gbdevw/purple-goctopus/sdk/noncegen"
	"github.com/gbdevw/purple-goctopus/sdk/spot/rest"
	restcommon "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
	"go.opentelemetry.io/otel/trace"
)

// Kraken spot websocket client with access to private endpoints
type KrakenSpotPrivateWebsocketClient struct {
	// Underlying base kraken spot websocket client
	*krakenSpotWebsocketClient
}

// # Description
//
// Factory which creates a KrakenSpotPrivateWebsocketClient that can be provided to a websocket
// engine (wscengine.WebsocketEngine - Cf. https://github.com/gbdevw/gowse).
//
// # Inputs
//
//   - restClient: Optional Kraken spot rest client to use to get a websocket token. Must not be nil.
//   - clientNonceGenerator: Optional nonce generator used to get nonces used to sign requests sent with the REST Client. Must not be nil.
//   - secopts: Optional security options (like password 2FA) to use when sending requests with the REST client. Can be nil if 2FA is not used.
//   - onCloseCallback: optional user defined callback which will be called when connection is closed/interrupted.
//   - onReadErrorCallback: optional user defined callback which will be called when an error occurs while reading messages from the websocket server
//   - onRestartError: optional user defined callback which will be called when the websocket engine fails to reconnect to the server.
//   - logger: Optional logger used to log debug/vebrose messages. If nil, a logger with a discard writer (noop) will be used
//   - tracerProvider: Tracer provider to use to get a tracer to instrument websocket client code. If nil, global tracer provider will be used.
//
// # Return
//
// A new KrakenSpotPrivateWebsocketClient
func NewKrakenSpotPrivateWebsocketClient(
	restClient rest.KrakenSpotRESTClientIface,
	clientNonceGenerator noncegen.NonceGenerator,
	secopts *restcommon.SecurityOptions,
	onCloseCallback func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails),
	onReadErrorCallback func(ctx context.Context, restart context.CancelFunc, exit context.CancelFunc, err error),
	onRestartError func(ctx context.Context, exit context.CancelFunc, err error, retryCount int),
	logger *log.Logger,
	tracerProvider trace.TracerProvider) (*KrakenSpotPrivateWebsocketClient, error) {
	// Check inputs
	if restClient == nil || clientNonceGenerator == nil {
		return nil, fmt.Errorf("rest client and nonce generator cannot be nil")
	}
	// Build & return public websocket client
	return &KrakenSpotPrivateWebsocketClient{
		krakenSpotWebsocketClient: newKrakenSpotWebsocketClient(
			restClient,
			clientNonceGenerator,
			secopts,
			onCloseCallback,
			onReadErrorCallback,
			onRestartError,
			logger,
			tracerProvider)}, nil
}
