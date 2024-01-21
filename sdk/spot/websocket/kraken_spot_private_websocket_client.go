package websocket

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gbdevw/gowse/wscengine"
	"github.com/gbdevw/gowse/wscengine/wsadapters/gorilla"
	"github.com/gbdevw/gowse/wscengine/wsclient"
	"github.com/gbdevw/purple-goctopus/sdk/noncegen"
	"github.com/gbdevw/purple-goctopus/sdk/spot/rest"
	restcommon "github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
	"github.com/hashicorp/go-retryablehttp"
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

// # Description
//
// Create a new KrakenSpotPrivateWebsocketClient and a websocket engine to run it. The client and
// the engine will be configured with the following default options:
//
//   - Target URL for the websocket client will be the production env.: wss://ws-auth.kraken.com
//   - Base URL for the REST client will be the production env. : https://api.kraken.com/0
//   - A HFNonceGenerator will be used as nonce generator for the Get Websocket Token requests
//   - a retryablehhttpclient will be used as HTTP client (3 retries, 1sec retry delay).
//   - InstrumentKrakenSpotRESTClientAuthorizer will be used to sign Get Websocket Token requests.
//   - The gorilla websocket framework with default settings will be used by the wesocket engine.
//   - Websocket engine settings: 4 workers, auto-reconnect enabled, 5sec exponential retry delay.
//
// The function will return the unstarted engine and the private websocket client attached to it.
//
// # Hint
//
// This functions eliminates a lot of the boiler plate code needed to create and gather components
// required to interact with Kraken websocket API:
//
//   - The websocket server URL.
//   - The base http.Client used to send HTTP requests.
//   - The REST client authorizer for Kraken API.
//   - The REST client for Kraken API.
//   - The nonce generator used to send requests with the REST client.
//   - The private websocket client for Kraken API.
//   - The websocket connection adapter used by the engine to manage the websocket connection.
//   - The engine to run the private websocket client.
//
// # Inputs
//
//   - key: API key used to authorize requests to the REST API (Get Websocket Token)
//   - b64secret: API secret provided as a base64 encoded bytestring.
//   - secopts: Optional security options to use when sending Get Websocket Token requests.
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
func NewDefaultEngineWithPrivateWebsocketClient(
	key string,
	b64secret string,
	secopts *restcommon.SecurityOptions,
	onCloseCallback func(ctx context.Context, closeMessage *wsclient.CloseMessageDetails),
	onReadErrorCallback func(ctx context.Context, restart context.CancelFunc, exit context.CancelFunc, err error),
	onRestartError func(ctx context.Context, exit context.CancelFunc, err error, retryCount int),
	logger *log.Logger,
	tracerProvider trace.TracerProvider,
) (*wscengine.WebsocketEngine, KrakenSpotPrivateWebsocketClientInterface, error) {
	// Build websocket server URL
	url, err := url.Parse(KrakenSpotWebsocketPrivateProductionURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s as a URL: %w", KrakenSpotWebsocketPrivateProductionURL, err)
	}
	// Create instrumented authorizer
	auth, err := rest.NewKrakenSpotRESTClientAuthorizer(key, b64secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build REST client's authorizer: %w", err)
	}
	authorizer := rest.InstrumentKrakenSpotRESTClientAuthorizer(auth, tracerProvider)
	// Build and configure a retryable http client
	httpclient := retryablehttp.NewClient()
	httpclient.RetryWaitMax = 1 * time.Second
	httpclient.RetryWaitMin = 1 * time.Second
	httpclient.RetryMax = 3
	httpclient.Logger = logger
	// Create an instrumented Kraken spot REST API.
	//	- REST client will target production environment.
	//	- REST client will use the retryable http client as underlying HTTP client.
	//	- REST client will use global tracer provider.
	restClient := rest.InstrumentKrakenSpotRESTClient(
		rest.NewKrakenSpotRESTClient(
			authorizer,
			&rest.KrakenSpotRESTClientConfiguration{
				BaseURL: rest.KrakenProductionV0BaseUrl,
				Client:  httpclient.StandardClient(),
			}),
		tracerProvider)
	// Create a HFNonceGenerator
	cngen := noncegen.NewHFNonceGenerator()
	// Build websocket client
	wsclient, err := NewKrakenSpotPrivateWebsocketClient(restClient, cngen, secopts, onCloseCallback, onReadErrorCallback, onRestartError, logger, tracerProvider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build the private websocket client: %w", err)
	}
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
