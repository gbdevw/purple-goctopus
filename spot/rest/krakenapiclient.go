package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/market"
	"go.opentelemetry.io/otel/trace"
)

/*****************************************************************************/
/*	API ENDPOINT PATHS													 	 */
/*****************************************************************************/

// Kraken spot REST API endpoints URL path
const (
	// Base URL
	KrakenProductionV0BaseUrl = "https://api.kraken.com/0"

	// Market Data

	serverTimePath         = "/public/Time"
	systemStatusPath       = "/public/SystemStatus"
	assetInfoUrlPath       = "/public/Assets"
	tradableAssetPairsPath = "/public/AssetPairs"
	tickerInformationPath  = "/public/Ticker"
	ohlcDataPath           = "/public/OHLC"
	orderBookPath          = "/public/Depth"
	recentTradesPath       = "/public/Trades"
	recentSpreadsPath      = "/public/Spread"

	// Account data

	getAccountBalancePath     = "/private/Balance"
	getExtendedBalance        = "/private/BalanceEx"
	getTradeBalancePath       = "/private/TradeBalance"
	getOpenOrdersPath         = "/private/OpenOrders"
	getClosedOrdersPath       = "/private/ClosedOrders"
	queryOrdersInfosPath      = "/private/QueryOrders"
	getTradesHistoryPath      = "/private/TradesHistory"
	queryTradesInfoPath       = "/private/QueryTrades"
	getOpenPositionsPath      = "/private/OpenPositions"
	getLedgersInfoPath        = "/private/Ledgers"
	queryLedgersPath          = "/private/QueryLedgers"
	getTradeVolumePath        = "/private/TradeVolume"
	requestExportReportPath   = "/private/AddExport"
	getExportReportStatusPath = "/private/ExportStatus"
	retrieveDataExportPath    = "/private/RetrieveExport"
	deleteExportReportPath    = "/private/RemoveExport"

	// Trading

	addOrderPath              = "/private/AddOrder"
	addOrderBatchPath         = "/private/AddOrderBatch"
	editOrderPath             = "/private/EditOrder"
	cancelOrderPath           = "/private/CancelOrder"
	cancelAllOrdersPath       = "/private/CancelAll"
	cancelAllOrdersAfterXPath = "/private/CancelAllOrdersAfter"
	cancelOrderBatchPath      = "/private/CancelOrderBatch"

	// User Funding

	getDepositMethodsPath             = "/private/DepositMethods"
	getDepositAddressesPath           = "/private/DepositAddresses"
	getStatusOfRecentDepositsPath     = "/private/DepositStatus"
	getWithdrawalMethodsPath          = "/private/WithdrawMethods"
	getWithdrawalAddresses            = "/private/WithdrawAddress"
	getWithdrawalInformationPath      = "/private/WithdrawInfo"
	withdrawFundsPath                 = "/private/Withdraw"
	getStatusOfRecentWithdrawalsPath  = "/private/WithdrawStatus"
	requestWithdrawalCancellationPath = "/private/WithdrawCancel"
	requestWalletTransferPath         = "/private/WalletTransfer"

	// Earn
	allocateEarnFundsPath     = "/private/Earn/Allocate"
	deallocateEarnFundsPath   = "/private/Earn/Deallocate"
	getAllocationStatusPath   = "/private/Earn/AllocateStatus"
	getDeallocationStatusPath = "/private/Earn/DeallocateStatus"
	listEarnStartegiesPath    = "/private/Earn/DeallocateStatus"
	listEarnAllocationsPath   = "/private/Earn/DeallocateStatus"
)

// Headers managed by KrakenAPIClient
const (
	// Headers

	managedHeaderContentType = "Content-Type"
	managedHeaderUserAgent   = "User-Agent"

	// Default value for User-Agent
	DefaultUserAgent = "goctopus"
)

/*****************************************************************************/
/* KRAKEN API CLIENT: MODEL & FACTORIES                                      */
/*****************************************************************************/

// KrakenSpotRESTClient is a high-level client to use Kraken spot REST API. The
// client implements KrakenSpotRESTClientIface.
type KrakenSpotRESTClient struct {
	// Base URL to use for Kraken spot REST API.
	baseURL string
	// Value for the mandatory User-Agent header.
	agent string
	// Authorizer used to authorize requests to Kraken spot REST API.
	authorizer KrakenSpotRESTClientAuthorizerIface
	// HTTP client used to perform API calls.
	client *http.Client
}

// Configuration for KrakenSpotRESTClient.
type KrakenSpotRESTClientConfiguration struct {
	// Base URL for the API.
	//
	// If an empty string is used, defaults to "https://api.kraken.com/0"
	BaseURL string
	// Value for the mandatory User-Agent.
	//
	// If an empty string is used, defaults to "goctopus"
	Agent string
	// Low level HTTP client to use to perform API calls.
	//
	// If nil, defaults to http.DefaultClient.
	Client *http.Client
}

// A factory which creates a new KrakenSpotRESTClientConfiguration with all its default values set.
func NewDefaultKrakenSpotRESTClientConfiguration() *KrakenSpotRESTClientConfiguration {
	return &KrakenSpotRESTClientConfiguration{
		BaseURL: KrakenProductionV0BaseUrl,
		Agent:   DefaultUserAgent,
		Client:  http.DefaultClient,
	}
}

// # Description
//
// A helper function which configures a KrakenSpotRESTClientAuthorizer to sign outgoing requests
// with the provide key and secret and decorate it with an instrumentation decorator that will
// use the provided tracerProvider to instrument code.
//
// # Inputs
//
//   - key: The API key used to sign requests
//   - secret: The base64 encoded API key secret provided by Kraken and used to sign requests.
//   - tracerProvider: TracerProvider to use to instrument code. If nil, global tracer provider will be used.
//
// # Returns
//
// The decorated authorizer or an error in case the provided secret cannot be base64 decoded.
func WithInstrumentedAuthorizer(key string, secret string, tracerProvider trace.TracerProvider) (KrakenSpotRESTClientAuthorizerIface, error) {
	// Create authorizer
	auth, err := NewKrakenSpotRESTClientAuthorizer(key, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorizer: %w", err)
	}
	// Decorate authorizer with an instrumentation decorator and return the decorator
	return DecorateKrakenSpotRESTClientAuthorizer(auth, tracerProvider), nil
}

// # Description
//
// Factory for KrakenSpotRESTClient.
//
// # Inputs
//
//   - authorizer: Authorizer to use to authorize requests to Kraken spot REST API. Can be nil. See notes below on the Authorizer.
//   - cfg: KrakenSpotRESTClient configuration. A nil value means all default configuration options will be used.
//
// # Returns
//
// The factory returns a fully initiated KrakenSpotRESTClient or nil in case of error.
//
// An error will be returned if:
//   - The secret in the provided credentials cannot be base64 decoded.
//
// # Authorizer
//
// The authorizer is a separate component which manages the authorization and post-processing of outgoing HTTP requests
// to the Kraken API (or other servers, proxies, ... depending on your settings).
//
// nil can be used as value for the authorizer: In this case, the API client will skip the request authorization and send the request.
// This is useful when user only wants to use the public endpoints.
//
// The SDK provides an implementation of the authorizer which signs the outgoing HTTP request by using an API key and a base64 encoded
// secret (cf KrakenSpotRESTClientAuthorizer). Both are provided by Kraken when the user generates an API key.
//
// See https://docs.kraken.com/rest/#section/Authentication/Headers-and-Signature for details about the signature.
//
// A helper function WithInstrumentedAuthorizer is provided to configure a KrakenSpotRESTClientAuthorizer instrumented with
// the OpenTelemetry framework.
//
// More advanced use cases can require to customize the authorization logic (proxying, egress gateways, custom L7 rules, ...).
// In this case, users can implement and provide their own authorizer implementation which satisfy their requirements.
func NewKrakenSpotRESTClient(authorizer KrakenSpotRESTClientAuthorizerIface, cfg *KrakenSpotRESTClientConfiguration) *KrakenSpotRESTClient {
	// Handle configuration
	defCfg := NewDefaultKrakenSpotRESTClientConfiguration()
	if cfg != nil {
		if cfg.BaseURL != "" {
			defCfg.BaseURL = cfg.BaseURL
		}
		if cfg.Agent != "" {
			defCfg.Agent = cfg.Agent
		}
		if cfg.Client != nil {
			defCfg.Client = cfg.Client
		}
	}
	// Build and return client
	return &KrakenSpotRESTClient{
		baseURL:    defCfg.BaseURL,
		agent:      defCfg.Agent,
		authorizer: authorizer,
		client:     defCfg.Client,
	}
}

/*****************************************************************************/
/* KRAKEN API CLIENT: UTILITIES                                              */
/*****************************************************************************/

// # Description
//
// Forge and authorize a HTTP request for the Kraken spot REST API.
//
// The method will create and initialize a new http.Request with the provided context
// and data. The method will set the mandatory User-Agent header and will authorize
// the request if an authorizer is set at the client level.
//
// Data required by the authorizer are expected to be already present in the provided
// data. For example, for the provided KrakenSpotRESTClientAuthorizer, the nonce and
// the optional otp must already be present in the provided body.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - path: The URL path for the API operation to use (ex: /private/Balance)
//   - httpMethod: The http method to use for the request
//   - query: Query string parameters. Can be nil or empty if no parameters are provided.
//   - body: Reader to use to get request body data. Can be nil if no body should be set.
//
// # Returns
//
//	An http.Request ready to be sent or an error if any.
func (client *KrakenSpotRESTClient) forgeAndAuthorizeKrakenAPIRequest(
	ctx context.Context,
	path string,
	httpMethod string,
	query url.Values,
	body io.Reader,
) (*http.Request, error) {

	// Set request url
	reqURL := fmt.Sprintf("%s%s", client.baseURL, path)
	// Add query string parameters if provided to request url
	if len(query) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, query.Encode())
	}
	// Forge http request
	req, err := http.NewRequestWithContext(ctx, httpMethod, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to forge HTTP request for Kraken API: %w", err)
	}
	// Set Content-Type and User-Agent headers
	req.Header.Set(managedHeaderUserAgent, client.agent)
	// Parse form to hanlde query string parameters and form body if any
	err = req.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("failed to forge HTTP request for Kraken API: %w", err)
	}
	// If an authorizer is set, authorize the request and return results
	return client.authorizer.Authorize(ctx, req)
}

// # Description
//
// Send the provided request to Kraken spot REST API and process the response if any.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - req: Request to authorize and send.
//   - receiver: receiver that will be used to parse the JSON response for Kraken API. Can be nil if binary data are expected as response.
//
// # Returns
//
//   - The parsed JSON response from KRaken API (= receiver)
//   - A reference to the raw http.Response (with its body closed except if the response contains binary data)
//   - An error if any has occured (error at HTTP level, error when parsing response, ...)
func (client *KrakenSpotRESTClient) doKrakenAPIRequest(ctx context.Context, req *http.Request, receiver interface{}) (*http.Response, error) {
	select {
	// Abort request processing if context has expired
	case <-ctx.Done():
		return nil, fmt.Errorf("aborting request: %w", ctx.Err())
	default:
		// Send the request using the provided http client
		resp, err := client.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to process HTTP request: %w", err)
		}
		// Check status code for error status
		//
		// API documentation states that "status codes other than 200 indicate
		// that there was an issue with the request reaching our servers"
		//
		// No body will be present.
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error code received from Kraken API: %d", resp.StatusCode)
		}
		// Check mime type of response
		mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return nil, fmt.Errorf("could not decode the response Content-Type header: %w", err)
		}
		// Depending on response content type
		switch mimeType {
		case "application/octet-stream":
			// Return response with its body not closed
			return resp, nil
		case "application/zip":
			// Return response with its body not closed
			return resp, nil
		case "application/json":
			// Parse body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}
			err = json.Unmarshal(body, &receiver)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JSON response: %w", err)
			}
			// Close body and retur response
			resp.Body.Close()
			return resp, nil
		default:
			// Return error -> unsupported content type
			resp.Body.Close()
			return nil, fmt.Errorf("response Content-Type is %s but ony application/json, application/octet-stream or application/zip are expected", mimeType)
		}
	}
}

/*****************************************************************************/
/* KRAKEN API CLIENT: OPERATIONS - MARKET DATA                               */
/*****************************************************************************/

// # Description
//
// GetServerTime - Get the server time.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Returns
//
//   - GetServerTimeResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError)
// or when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetServerTime(ctx context.Context) (*market.GetServerTimeResponse, *http.Response, error) {
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, serverTimePath, http.MethodGet, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetServerTime: %w", err)
	}
	// Send the request
	receiver := new(market.GetServerTimeResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetServerTime failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetSystemStatus - Get the current system status or trading mode.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//
// # Returns
//
//   - GetSystemStatusResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError)
// or when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetSystemStatus(ctx context.Context) (*market.GetSystemStatusResponse, *http.Response, error) {
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetSystemStatus: %w", err)
	}
	// Send the request
	receiver := new(market.GetSystemStatusResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetSystemStatus failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetAssetInfo - Get information about the assets that are available for deposit, withdrawal, trading and staking.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - opts: GetAssetInfo request options. A nil value triggers all default behaviors.
//
// # Returns
//
//   - GetAssetInfoResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetAssetInfo(ctx context.Context, opts *market.GetAssetInfoRequestOptions) (*market.GetAssetInfoResponse, *http.Response, error) {
	// Prepare query string params.
	query := url.Values{}
	if opts != nil {
		if len(opts.Assets) > 0 {
			query.Add("asset", strings.Join(opts.Assets, ","))
		}
		if opts.AssetClass != "" {
			query.Add("aclass", string(opts.AssetClass))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetAssetInfo: %w", err)
	}
	// Send the request
	receiver := new(market.GetAssetInfoResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetAssetInfo failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// # GetTradableAssetPairs - Get tradable asset pairs
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - opts: GetTradableAssetPairs request options. A nil value triggers all default behaviors.
//
// # Returns
//
//   - GetTradableAssetPairsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetTradableAssetPairs(ctx context.Context, opts *market.GetTradableAssetPairsRequestOptions) (*market.GetTradableAssetPairsResponse, *http.Response, error) {
	// Prepare query string params.
	query := url.Values{}
	if opts != nil {
		if len(opts.Pairs) > 0 {
			// Pairs must be provided as a comma separated string
			query.Add("pair", strings.Join(opts.Pairs, ","))
		}
		if opts.Info != "" {
			query.Add("info", opts.Info)
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetTradableAssetPairs: %w", err)
	}
	// Send the request
	receiver := new(market.GetTradableAssetPairsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetTradableAssetPairs failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetTickerInformation - Get data about today's price. Today's prices start at midnight UTC.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - opts: GetTickerInformation request options
//
// # Returns
//
//   - GetTickerInformationResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetTickerInformation(ctx context.Context, opts *market.GetTickerInformationRequestOptions) (*market.GetTickerInformationResponse, *http.Response, error) {
	// Prepare query string params.
	query := url.Values{}
	if opts != nil && len(opts.Pairs) > 0 {
		// Provide pairs as a comma separated string
		query.Add("pair", strings.Join(opts.Pairs, ","))
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetTickerInformation: %w", err)
	}
	// Send the request
	receiver := new(market.GetTickerInformationResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetTickerInformation failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetOHLCData - Return up to 720 OHLC data points since now or since given timestamp.
//
// Note: the last entry in the OHLC array is for the current, not-yet-committed frame and will always be present,
// regardless of the value of since.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - params: GetOHLCData request parameters.
//   - opts: GetOHLCData request options. A nil value triggers all default behaviors.
//
// # Returns
//
//   - GetOHLCDataResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetOHLCData(ctx context.Context, params market.GetOHLCDataRequestParameters, opts *market.GetOHLCDataRequestOptions) (*market.GetOHLCDataResponse, *http.Response, error) {
	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if opts != nil {
		if opts.Interval != 0 {
			query.Add("interval", strconv.FormatInt(int64(opts.Interval), 10))
		}
		if opts.Since != 0 {
			query.Add("since", strconv.FormatInt(opts.Since, 10))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetOHLCData: %w", err)
	}
	// Send the request
	receiver := new(market.GetOHLCDataResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetOHLCData failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetOrderBook - Get the target market order book.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - params: GetOrderBook request parameters.
//   - opts: GetOrderBook request options. A nil value triggers all default behaviors.
//
// # Returns
//
//   - GetOrderBookResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetOrderBook(ctx context.Context, params market.GetOrderBookRequestParameters, opts *market.GetOrderBookRequestOptions) (*market.GetOrderBookResponse, *http.Response, error) {
	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if opts != nil {
		if opts.Count != 0 {
			query.Add("count", strconv.FormatInt(int64(opts.Count), 10))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetOrderBook: %w", err)
	}
	// Send the request
	receiver := new(market.GetOrderBookResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetOrderBook failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetRecentTrades - Returns up to the last 1000 trades since now or since given timestamp.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose
//   - params: GetRecentTrades request parameters.
//   - opts: GetRecentTrades request options. A nil value triggers all default behaviors.
//
// # Returns
//
//   - GetRecentTradesResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
//
//	when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetRecentTrades(ctx context.Context, params market.GetRecentTradesRequestParameters, opts *market.GetRecentTradesRequestOptions) (*market.GetRecentTradesResponse, *http.Response, error) {
	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if opts != nil {
		if opts.Since != 0 {
			query.Add("since", strconv.FormatInt(opts.Since, 10))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetRecentTrades: %w", err)
	}
	// Send the request
	receiver := new(market.GetRecentTradesResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetRecentTrades failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetRecentSpreads - Returns the last ~200 top-of-book spreads for a given pair as for now as as a given timestamp.
//
// Note: Intended for incremental updates within available dataset (does not contain all historical spreads).
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose
//   - params: GetRecentSpreads request parameters.
//   - opts: GetRecentSpreads request options. A nil value triggers all default behaviors.
//
// # Returns
//
//   - GetRecentSpreadsResponse: The parsed response from Kraken API.
//   - http.Response: A reference to the raw HTTP response received from Kraken API.
//   - error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
//
// # Note on error
//
// The error is set only when something wrong has happened either at the HTTP level (while building the request,
// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
// when context has expired.
//
// An nil error does not mean everything is OK: You also have to check the response error field for specific
// errors from Kraken API.
//
// # Note on the http.Response
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) GetRecentSpreads(ctx context.Context, params market.GetRecentSpreadsRequestParameters, opts *market.GetRecentSpreadsRequestOptions) (*market.GetRecentSpreadsResponse, *http.Response, error) {
	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if opts != nil {
		if opts.Since != 0 {
			query.Add("since", strconv.FormatInt(opts.Since, 10))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, systemStatusPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetRecentSpreads: %w", err)
	}
	// Send the request
	receiver := new(market.GetRecentSpreadsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, nil, fmt.Errorf("request for GetRecentSpreads failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

/*****************************************************************************/
/*																			 */
/*	PRIVATE ENDPOINTS - USER DATA											 */
/*																			 */
/*****************************************************************************/

// GetAccountBalance - Retrieve all cash balances, net of pending withdrawals.
func (client *KrakenSpotRESTClient) GetAccountBalance(secopts *SecurityOptions) (*GetAccountBalanceResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Perform request
	resp, err := api.queryPrivate(postGetAccountBalance, http.MethodPost, nil, "", nil, nil, otp, &GetAccountBalanceResponse{})
	if err != nil {
		return nil, err
	}

	// Parse response
	balances, ok := resp.(*GetAccountBalanceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected payload received from the server. Got %T : %v", resp, resp)
	}

	// Return response
	return balances, nil
}

// GetTradeBalance - Retrieve a summary of collateral balances, margin position valuations, equity and margin level.
func (client *KrakenSpotRESTClient) GetTradeBalance(options *GetTradeBalanceOptions, secopts *SecurityOptions) (*GetTradeBalanceResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := url.Values{}
	if options != nil {
		if options.Asset != "" {
			body.Set("asset", options.Asset)
		}
	}

	// Perform request
	resp, err := api.queryPrivate(postGetTradeBalance, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetTradeBalanceResponse{})
	if err != nil {
		return nil, err
	}

	// Parse response
	respBalances, ok := resp.(*GetTradeBalanceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected payload received from the server. Got %T : %v", resp, resp)
	}

	// Return response
	return respBalances, nil
}

// GetOpenOrders - Retrieve information about currently open orders.
func (client *KrakenSpotRESTClient) GetOpenOrders(options *GetOpenOrdersOptions, secopts *SecurityOptions) (*GetOpenOrdersResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := url.Values{}
	if options != nil {
		if options.Trades {
			body.Set("trades", strconv.FormatBool(options.Trades))
		}
		if options.UserReference != nil {
			body.Set("userref", strconv.FormatInt(*options.UserReference, 10))
		}
	}

	// Perform request
	resp, err := api.queryPrivate(postGetOpenOrders, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetOpenOrdersResponse{})
	if err != nil {
		return nil, err
	}

	// Parse response
	orders, ok := resp.(*GetOpenOrdersResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected payload received from the server. Got %T : %v", resp, resp)
	}

	// Return response
	return orders, nil
}

// GetClosedOrders -
// Retrieve information about orders that have been closed (filled or cancelled).
// 50 results are returned at a time, the most recent by default.
func (client *KrakenSpotRESTClient) GetClosedOrders(options *GetClosedOrdersOptions, secopts *SecurityOptions) (*GetClosedOrdersResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := url.Values{}
	if options != nil {
		if options.Trades {
			body.Set("trades", strconv.FormatBool(options.Trades))
		}
		if options.UserReference != nil {
			body.Set("userref", strconv.FormatInt(*options.UserReference, 10))
		}
		if options.Start != nil {
			body.Set("start", strconv.FormatInt(options.Start.Unix(), 10))
		}
		if options.End != nil {
			body.Set("end", strconv.FormatInt(options.End.Unix(), 10))
		}
		if options.Offset != nil {
			body.Set("ofs", strconv.FormatInt(*options.Offset, 10))
		}
		if options.Closetime != "" {
			body.Set("closetime", string(options.Closetime))
		}
	}

	// Perform request
	resp, err := api.queryPrivate(postGetClosedOrders, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetClosedOrdersResponse{})
	if err != nil {
		return nil, err
	}

	// Parse response
	orders, ok := resp.(*GetClosedOrdersResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected payload received from the server. Got %T : %v", resp, resp)
	}

	// Return response
	return orders, nil
}

// QueryOrdersInfo - Retrieve information about specific orders.
func (client *KrakenSpotRESTClient) QueryOrdersInfo(params QueryOrdersParameters, options *QueryOrdersOptions, secopts *SecurityOptions) (*QueryOrdersInfoResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("txid", strings.Join(params.TransactionIds, ","))
	if options != nil {
		if options.Trades {
			body.Set("trades", strconv.FormatBool(options.Trades))
		}
		if options.UserReference != nil {
			body.Set("userref", strconv.FormatInt(*options.UserReference, 10))
		}
	}

	// Request
	resp, err := api.queryPrivate(postQueryOrdersInfos, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &QueryOrdersInfoResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*QueryOrdersInfoResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to QueryOrdersInfoResponse. Got %T : %v", resp, resp)
	}

	return result, nil
}

// GetTradesHistory -
// Retrieve information about trades/fills.
// 50 results are returned at a time, the most recent by default.
//
// Unless otherwise stated, costs, fees, prices, and volumes are specified with the precision for the asset pair
// (pair_decimals and lot_decimals), not the individual assets' precision (decimals).
func (client *KrakenSpotRESTClient) GetTradesHistory(options *GetTradesHistoryOptions, secopts *SecurityOptions) (*GetTradesHistoryResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Use default options if none provided
	if options == nil {
		options = &GetTradesHistoryOptions{}
	}
	// Prepare request body
	body := url.Values{}
	if options != nil {
		if options.Trades {
			body.Set("trades", strconv.FormatBool(options.Trades))
		}
		if options.Type != "" {
			body.Set("type", string(options.Type))
		}
		if options.Start != nil {
			body.Set("start", strconv.FormatInt(options.Start.Unix(), 10))
		}
		if options.End != nil {
			body.Set("end", strconv.FormatInt(options.End.Unix(), 10))
		}
		if options.Offset != nil {
			body.Set("ofs", strconv.FormatInt(*options.Offset, 10))
		}
	}

	// Request
	resp, err := api.queryPrivate(postGetTradesHistory, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetTradesHistoryResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetTradesHistoryResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to GetTradesHistoryResponse. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

// QueryTradesInfo - Retrieve information about specific trades/fills.
func (client *KrakenSpotRESTClient) QueryTradesInfo(params QueryTradesParameters, options *QueryTradesOptions, secopts *SecurityOptions) (*QueryTradesInfoResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("txid", strings.Join(params.TransactionIds, ","))
	if options != nil {
		if options.Trades {
			body.Set("trades", strconv.FormatBool(options.Trades))
		}
	}

	// Request
	resp, err := api.queryPrivate(postQueryTradesInfo, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &QueryTradesInfoResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*QueryTradesInfoResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to map of TradeInfo. Got %T : %v", resp, resp)
	}

	// Return response
	return result, nil
}

// GetOpenPositions - Get information about open margin positions.
func (client *KrakenSpotRESTClient) GetOpenPositions(options *GetOpenPositionsOptions, secopts *SecurityOptions) (*GetOpenPositionsResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	if options != nil {
		if options.TransactionIds != nil {
			body.Set("txid", strings.Join(options.TransactionIds, ","))
		}
		if options.DoCalcs {
			body.Set("docalcs", strconv.FormatBool(options.DoCalcs))
		}
		// CF DEBT.MD
		//if options.Consolidation != "" {
		//	body.Set("consolidation", options.Consolidation)
		//}
	}

	// Request
	resp, err := api.queryPrivate(postGetOpenPositions, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetOpenPositionsResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetOpenPositionsResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to GetOpenPositionsResponse. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

// GetLedgersInfo - Retrieve information about ledger entries. 50 results are returned at a time, the most recent by default.
func (client *KrakenSpotRESTClient) GetLedgersInfo(options *GetLedgersInfoOptions, secopts *SecurityOptions) (*GetLedgersInfoResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	if options != nil {
		if options.Assets != nil {
			body.Set("asset", strings.Join(options.Assets, ","))
		}
		if options.AssetClass != "" {
			body.Set("aclass", string(options.AssetClass))
		}
		if options.Type != "" {
			body.Set("type", string(options.Type))
		}
		if options.Start != nil {
			body.Set("start", strconv.FormatInt(options.Start.Unix(), 10))
		}
		if options.End != nil {
			body.Set("end", strconv.FormatInt(options.End.Unix(), 10))
		}
		if options.Offset != nil {
			body.Set("ofs", strconv.FormatInt(*options.Offset, 10))
		}
	}

	// Request
	resp, err := api.queryPrivate(postGetLedgersInfo, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetLedgersInfoResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetLedgersInfoResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to GetLedgersInfoResponse. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

// QueryLedgers - Retrieve information about specific ledger entries.
func (client *KrakenSpotRESTClient) QueryLedgers(params QueryLedgersParameters, options *QueryLedgersOptions, secopts *SecurityOptions) (*QueryLedgersResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("id", strings.Join(params.LedgerIds, ","))
	if options != nil {
		if options.Trades {
			body.Set("trades", strconv.FormatBool(options.Trades))
		}
	}

	// Request
	resp, err := api.queryPrivate(postQueryLedgers, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &QueryLedgersResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*QueryLedgersResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to map of ledger entries. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

// GetTradeVolume gets trade volume
//
// Note: If an asset pair is on a maker/taker fee schedule, the taker side is given in fees and maker
// side in fees_maker. For pairs not on maker/taker, they will only be given in fees.
func (client *KrakenSpotRESTClient) GetTradeVolume(params GetTradeVolumeParameters, options *GetTradeVolumeOptions, secopts *SecurityOptions) (*GetTradeVolumeResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("pair", strings.Join(params.Pairs, ","))
	if options != nil {
		if options.FeeInfo {
			body.Set("fee-info", strconv.FormatBool(options.FeeInfo))
		}
	}

	// Request
	resp, err := api.queryPrivate(postGetTradeVolume, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetTradeVolumeResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetTradeVolumeResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to GetTradeVolumeResponse. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

// RequestExportReport - Request export of trades or ledgers.
func (client *KrakenSpotRESTClient) RequestExportReport(params RequestExportReportParameters, options *RequestExportReportOptions, secopts *SecurityOptions) (*RequestExportReportResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("report", string(params.Report))
	body.Set("description", params.Description)
	if options != nil {
		if options.Format != "" {
			body.Set("format", string(options.Format))
		}
		if options.Fields != nil {
			body.Set("fields", strings.Join(options.Fields, ","))
		}
		if options.StartTm != nil {
			body.Set("starttm", strconv.FormatInt(options.StartTm.Unix(), 10))
		}
		if options.EndTm != nil {
			body.Set("endtm", strconv.FormatInt(options.EndTm.Unix(), 10))
		}
	}

	// Request
	resp, err := api.queryPrivate(postRequestExportReport, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &RequestExportReportResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*RequestExportReportResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to RequestExportReportResponse. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

// GetExportReportStatus - Get status of requested data exports.
func (client *KrakenSpotRESTClient) GetExportReportStatus(params GetExportReportStatusParameters, secopts *SecurityOptions) (*GetExportReportStatusResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("report", string(params.Report))

	// Request
	resp, err := api.queryPrivate(postGetExportReportStatus, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetExportReportStatusResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetExportReportStatusResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to array of ExportReportStatus. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

// RetrieveDataExport Get report as a zip
func (client *KrakenSpotRESTClient) RetrieveDataExport(params RetrieveDataExportParameters, secopts *SecurityOptions) (*RetrieveDataExportResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("id", params.Id)

	// Request
	resp, err := api.queryPrivate(postRetrieveDataExport, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, nil)
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.([]uint8)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to bytes. Got %T : %v", resp, resp)
	}

	// Return results
	return &RetrieveDataExportResponse{Report: result}, nil
}

// DeleteExportReport - Delete exported trades/ledgers report
func (client *KrakenSpotRESTClient) DeleteExportReport(params DeleteExportReportParameters, secopts *SecurityOptions) (*DeleteExportReportResponse, error) {

	// Use security options with zero values if nil is provided for secopts
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("id", params.Id)
	body.Set("type", string(params.Type))

	// Request
	resp, err := api.queryPrivate(postDeleteExportReport, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &DeleteExportReportResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*DeleteExportReportResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to DeleteExportReportResponse. Got %T : %v", resp, resp)
	}

	// Return results
	return result, nil
}

/*****************************************************************************/
/*	PRIVATE ENDPOINTS - USER TRADING										 */
/*****************************************************************************/

// AddOrder places a new order
func (client *KrakenSpotRESTClient) AddOrder(params AddOrderParameters, options *AddOrderOptions, secopts *SecurityOptions) (*AddOrderResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := make(url.Values)

	// Set targeted asset pair
	body.Set("pair", params.Pair)

	// Add user reference if defined
	if params.Order.UserReference != nil {
		body.Set("userref", strconv.FormatInt(*params.Order.UserReference, 10))
	}

	// Set order type
	body.Set("ordertype", string(params.Order.OrderType))

	// Set order direction
	body.Set("type", string(params.Order.Type))

	// Set volume
	body.Set("volume", params.Order.Volume)

	// Set price if not empty
	if params.Order.Price != "" {
		body.Set("price", params.Order.Price)
	}

	// Set price2 if not empty
	if params.Order.Price2 != "" {
		body.Set("price2", params.Order.Price2)
	}

	// Set Trigger if value is not empty
	if params.Order.Trigger != "" {
		body.Set("trigger", string(params.Order.Trigger))
	}

	// Set leverage if provided value is not empty
	if params.Order.Leverage != "" {
		body.Set("leverage", params.Order.Leverage)
	}

	// Set STP flag if not empty
	if params.Order.StpType != "" {
		body.Set("stp_type", string(params.Order.StpType))
	}

	// Set Reduce only if set
	if params.Order.ReduceOnly {
		body.Set("reduce_only", strconv.FormatBool(params.Order.ReduceOnly))
	}

	// Set operation flags as a comma separated list if not empty
	if params.Order.OrderFlags != "" {
		body.Set("oflags", params.Order.OrderFlags)
	}

	// Set time in force if defined
	if params.Order.TimeInForce != "" {
		body.Set("timeinforce", string(params.Order.TimeInForce))
	}

	// Set start time if not empty
	if params.Order.ScheduledStartTime != "" {
		body.Set("starttm", params.Order.ScheduledStartTime)
	}

	// Set expire time if not empty
	if params.Order.ExpirationTime != "" {
		body.Set("expiretm", params.Order.ExpirationTime)
	}

	// Set close order if defined
	if params.Order.Close != nil {
		// Set order type
		body.Set("close[ordertype]", string(params.Order.Close.OrderType))
		// Set close order price
		body.Set("close[price]", params.Order.Close.Price)
		// Set price2 if not empty
		if params.Order.Close.Price2 != "" {
			body.Set("close[price2]", params.Order.Close.Price2)
		}
	}

	// Set options if provided
	if options != nil {
		// Set deadline if defined
		if options.Deadline != nil {
			body.Set("deadline", options.Deadline.Format(time.RFC3339))
		}
		// Set validate
		body.Set("validate", strconv.FormatBool(options.Validate))
	}

	// Perform http request
	resp, err := api.queryPrivate(postAddOrder, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &AddOrderResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*AddOrderResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to array of AddOrderResponse. Got %T : %v", resp, resp)
	}

	// Return response
	return result, nil
}

// AddOrderBatch sends an array of orders (max: 15). Any orders rejected due to order validations,
// will be dropped and the rest of the batch is processed. All orders in batch should be limited to
// a single pair. The order of returned txid's in the response array is the same as the order of the
// order list sent in request.
func (client *KrakenSpotRESTClient) AddOrderBatch(params AddOrderBatchParameters, options *AddOrderBatchOptions, secopts *SecurityOptions) (*AddOrderBatchResponse, error) {

	// Use provided otp if any
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Check that orders is not empty
	if len(params.Orders) == 0 {
		return nil, fmt.Errorf("provided order list cannot be empty")
	}

	// Prepare request body
	body := make(url.Values)

	// Set targeted asset pair
	body.Set("pair", params.Pair)

	// Set orders
	for index, order := range params.Orders {

		// Add user reference if defined
		if order.UserReference != nil {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "userref"), strconv.FormatInt(*order.UserReference, 10))
		}

		// Set order type
		body.Set(fmt.Sprintf("orders[%d][%s]", index, "ordertype"), string(order.OrderType))

		// Set order direction
		body.Set(fmt.Sprintf("orders[%d][%s]", index, "type"), string(order.Type))

		// Set volume
		body.Set(fmt.Sprintf("orders[%d][%s]", index, "volume"), order.Volume)

		// Set price if not empty
		if order.Price != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "price"), order.Price)
		}

		// Set price2 if not empty
		if order.Price2 != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "price2"), order.Price2)
		}

		// Set Trigger if value is not empty
		if order.Trigger != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "trigger"), string(order.Trigger))
		}

		// Set leverage if provided value is not empty
		if order.Leverage != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "leverage"), order.Leverage)
		}

		// Set STP flag if not empty
		if order.StpType != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "stp_type"), string(order.StpType))
		}

		// Set Reduce only if set
		if order.ReduceOnly {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "reduce_only"), strconv.FormatBool(order.ReduceOnly))
		}

		// Set operation flags as a comma separated list if not empty
		if order.OrderFlags != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "oflags"), order.OrderFlags)
		}

		// Set time in force if defined
		if order.TimeInForce != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "timeinforce"), string(order.TimeInForce))
		}

		// Set start time if not empty
		if order.ScheduledStartTime != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "starttm"), order.ScheduledStartTime)
		}

		// Set expire time if not empty
		if order.ExpirationTime != "" {
			body.Set(fmt.Sprintf("orders[%d][%s]", index, "expiretm"), order.ExpirationTime)
		}

		// Set close order if defined
		if order.Close != nil {
			// Set order type
			body.Set(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "ordertype"), string(order.Close.OrderType))
			// Set close order price
			body.Set(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price"), order.Close.Price)
			// Set price2 if not empty
			if order.Close.Price2 != "" {
				body.Set(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price2"), order.Close.Price2)
			}
		}
	}

	// Set options if any
	if options != nil {
		// Set deadline if defined
		if options.Deadline != nil {
			body.Set("deadline", options.Deadline.Format(time.RFC3339))
		}
		// Set validate
		body.Set("validate", strconv.FormatBool(options.Validate))
	}

	// Perform request
	resp, err := api.queryPrivate(postAddOrderBatch, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &AddOrderBatchResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*AddOrderBatchResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to array of AddOrderBatchResponse. Got %T : %v", resp, resp)
	}

	// Return response
	return result, nil
}

// Edit volume and price on open orders. Uneditable orders include margin orders, triggered stop/profit orders,
// orders with conditional close terms attached, those already cancelled or filled, and those where the executed
// volume is greater than the newly supplied volume. post-only flag is not retained from original order after
// successful edit. post-only needs to be explicitly set on edit request.
func (client *KrakenSpotRESTClient) EditOrder(params EditOrderParameters, options *EditOrderOptions, secopts *SecurityOptions) (*EditOrderResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := make(url.Values)

	// Set txid
	body.Set("txid", params.Id)

	// St options if any
	if options != nil {
		// Set userref if defined
		if options.NewUserReference != "" {
			body.Set("userref", options.NewUserReference)
		}

		// Set targeted asset pair
		body.Set("pair", params.Pair)

		// Set volume if defined
		if options.NewVolume != "" {
			body.Set("volume", options.NewVolume)
		}

		// Set price if not empty
		if options.Price != "" {
			body.Set("price", options.Price)
		}

		// Set price2 if not empty
		if options.Price2 != "" {
			body.Set("price2", options.Price2)
		}

		// Set oflags if not nil
		if options.OFlags != nil {
			// oflags is a comma separated list
			body.Set("oflags", strings.Join(options.OFlags, ","))
		}

		// Set deadline if defined
		if options.Deadline != nil {
			body.Set("deadline", options.Deadline.Format(time.RFC3339))
		}

		// Set cancel_response
		body.Set("cancel_response", strconv.FormatBool(options.CancelResponse))

		// Set validate
		body.Set("validate", strconv.FormatBool(options.Validate))
	}

	// Perform request
	resp, err := api.queryPrivate(postEditOrder, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &EditOrderResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*EditOrderResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to array of EditOrderResponse. Got %T : %v", resp, resp)
	}

	// Return response
	return result, nil
}

// Cancel a particular open order (or set of open orders) by txid or userref
func (client *KrakenSpotRESTClient) CancelOrder(params CancelOrderParameters, secopts *SecurityOptions) (*CancelOrderResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("txid", params.Id)

	// Perform request
	resp, err := api.queryPrivate(postCancelOrder, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &CancelOrderResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*CancelOrderResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to array of CancelOrderResponse. Got %T : %v", resp, resp)
	}

	// Return response
	return result, nil
}

// Cancel all open orders
func (client *KrakenSpotRESTClient) CancelAllOrders(secopts *SecurityOptions) (*CancelAllOrdersResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Perform request
	resp, err := api.queryPrivate(postCancelAllOrders, http.MethodPost, nil, "", nil, nil, otp, &CancelAllOrdersResponse{})
	if err != nil {
		return nil, err
	}

	// Return result
	return resp.(*CancelAllOrdersResponse), nil
}

// CancelAllOrdersAfter provides a "Dead Man's Switch" mechanism to protect the client from network malfunction,
// extreme latency or unexpected matching engine downtime. The client can send a request with a timeout (in seconds),
// that will start a countdown timer which will cancel all client orders when the timer expires. The client has to
// keep sending new requests to push back the trigger time, or deactivate the mechanism by specifying a timeout of 0.
// If the timer expires, all orders are cancelled and then the timer remains disabled until the client provides a new
// (non-zero) timeout.
//
// The recommended use is to make a call every 15 to 30 seconds, providing a timeout of 60 seconds. This allows the
// client to keep the orders in place in case of a brief disconnection or transient delay, while keeping them safe
// in case of a network breakdown. It is also recommended to disable the timer ahead of regularly scheduled trading
// engine maintenance (if the timer is enabled, all orders will be cancelled when the trading engine comes back from
// downtime - planned or otherwise).
func (client *KrakenSpotRESTClient) CancelAllOrdersAfterX(params CancelCancelAllOrdersAfterXParameters, secopts *SecurityOptions) (*CancelAllOrdersAfterXResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("timeout", strconv.FormatInt(int64(params.Timeout), 10))

	// Perform request
	resp, err := api.queryPrivate(postCancelAllOrdersAfterX, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &CancelAllOrdersAfterXResponse{})
	if err != nil {
		return nil, err
	}

	// Return result
	return resp.(*CancelAllOrdersAfterXResponse), nil
}

// Cancel multiple open orders by txid or userref
func (client *KrakenSpotRESTClient) CancelOrderBatch(params CancelOrderBatchParameters, secopts *SecurityOptions) (*CancelOrderBatchResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Validate input
	if len(params.OrderIds) == 0 {
		return nil, fmt.Errorf("orders must not be empty")
	}

	// Prepare body
	body := make(url.Values)
	for index, value := range params.OrderIds {
		body.Set(fmt.Sprintf("orders[%d]", index), value)
	}

	// Perform request
	resp, err := api.queryPrivate(postCancelOrderBatch, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &CancelOrderBatchResponse{})
	if err != nil {
		return nil, err
	}

	// Return result
	return resp.(*CancelOrderBatchResponse), nil
}

/*****************************************************************************/
/*	PRIVATE ENDPOINTS - USER FUNDING										 */
/*****************************************************************************/

// Retrieve methods available for depositing a particular asset.
func (client *KrakenSpotRESTClient) GetDepositMethods(params GetDepositMethodsParameters, secopts *SecurityOptions) (*GetDepositMethodsResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := make(url.Values)
	body.Set("asset", params.Asset)

	// Perform request
	resp, err := api.queryPrivate(postGetDepositMethods, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetDepositMethodsResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*GetDepositMethodsResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// return response
	return response, nil
}

// Retrieve (or generate a new) deposit addresses for a particular asset and method.
func (client *KrakenSpotRESTClient) GetDepositAddresses(params GetDepositAddressesParameters, options *GetDepositAddressesOptions, secopts *SecurityOptions) (*GetDepositAddressesResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("asset", params.Asset)
	body.Set("method", params.Method)

	if options != nil {
		if options.New {
			body.Set("new", strconv.FormatBool(options.New))
		}
	}

	// Perform request
	resp, err := api.queryPrivate(postGetDepositAddresses, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetDepositAddressesResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*GetDepositAddressesResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// Return response
	return response, nil
}

// Retrieve information about recent deposits made.
func (client *KrakenSpotRESTClient) GetStatusOfRecentDeposits(params GetStatusOfRecentDepositsParameters, options *GetStatusOfRecentDepositsOptions, secopts *SecurityOptions) (*GetStatusOfRecentDepositsResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("asset", params.Asset)

	if options != nil {
		body.Set("method", options.Method)
	}
	// Perform request
	resp, err := api.queryPrivate(postGetStatusOfRecentDeposits, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetStatusOfRecentDepositsResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*GetStatusOfRecentDepositsResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// Return response
	return response, nil
}

// Retrieve fee information about potential withdrawals for a particular asset, key and amount.
func (client *KrakenSpotRESTClient) GetWithdrawalInformation(params GetWithdrawalInformationParameters, secopts *SecurityOptions) (*GetWithdrawalInformationResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("asset", params.Asset)
	body.Set("key", params.Key)
	body.Set("amount", params.Amount)

	// Perform request
	resp, err := api.queryPrivate(postGetWithdrawalInformation, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetWithdrawalInformationResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*GetWithdrawalInformationResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// Return response
	return response, nil
}

// Make a withdrawal request.
func (client *KrakenSpotRESTClient) WithdrawFunds(params WithdrawFundsParameters, secopts *SecurityOptions) (*WithdrawFundsResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("asset", params.Asset)
	body.Set("key", params.Key)
	body.Set("amount", params.Amount)

	// Perform request
	resp, err := api.queryPrivate(postWithdrawFunds, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &WithdrawFundsResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*WithdrawFundsResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// Return response
	return response, nil
}

// Retrieve information about recently requests withdrawals.
func (client *KrakenSpotRESTClient) GetStatusOfRecentWithdrawals(params GetStatusOfRecentWithdrawalsParameters, options *GetStatusOfRecentWithdrawalsOptions, secopts *SecurityOptions) (*GetStatusOfRecentWithdrawalsResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("asset", params.Asset)
	// Set options if provided
	if options != nil {
		body.Set("method", options.Method)
	}

	// Perform request
	resp, err := api.queryPrivate(postGetStatusOfRecentWithdrawals, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &GetStatusOfRecentWithdrawalsResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*GetStatusOfRecentWithdrawalsResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// Return response
	return response, nil
}

// Cancel a recently requested withdrawal, if it has not already been successfully processed.
func (client *KrakenSpotRESTClient) RequestWithdrawalCancellation(params RequestWithdrawalCancellationParameters, secopts *SecurityOptions) (*RequestWithdrawalCancellationResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("asset", params.Asset)
	body.Set("refid", params.ReferenceId)

	// Perform request
	resp, err := api.queryPrivate(postRequestWithdrawalCancellation, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &RequestWithdrawalCancellationResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*RequestWithdrawalCancellationResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// Return response
	return response, nil
}

// Transfer from Kraken spot wallet to Kraken Futures holding wallet. Note that a transfer in the other direction must be requested via the Kraken Futures API endpoint.
func (client *KrakenSpotRESTClient) RequestWalletTransfer(params RequestWalletTransferParameters, secopts *SecurityOptions) (*RequestWalletTransferResponse, error) {

	// Use 2FA if provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare body
	body := make(url.Values)
	body.Set("asset", params.Asset)
	body.Set("from", params.From)
	body.Set("to", params.To)
	body.Set("amount", params.Amount)

	// Perform request
	resp, err := api.queryPrivate(postRequestWalletTransfer, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &RequestWalletTransferResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion for response
	response, ok := resp.(*RequestWalletTransferResponse)
	if !ok {
		return nil, fmt.Errorf("could not parse response from server as expected. Got %T", resp)
	}

	// Return response
	return response, nil
}

/*****************************************************************************/
/*	PRIVATE ENDPOINTS - USER STAKING  METHODS                                */
/*****************************************************************************/

// StakeAsset stake an asset from spot wallet.
func (client *KrakenSpotRESTClient) StakeAsset(params StakeAssetParameters, secopts *SecurityOptions) (*StakeAssetResponse, error) {

	// Use empty value for otp if no second factor provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("asset", params.Asset)
	body.Set("amount", params.Amount)
	body.Set("method", params.Method)

	// Make request
	resp, err := api.queryPrivate(postStakeAsset, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &StakeAssetResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*StakeAssetResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to StakeAssetResponse. Got %T : %#v", resp, resp)
	}

	return result, nil
}

// UnstakeAsset unstake an asset from your staking wallet.
func (client *KrakenSpotRESTClient) UnstakeAsset(params UnstakeAssetParameters, secopts *SecurityOptions) (*UnstakeAssetResponse, error) {

	// Use empty value for otp if no second factor provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Prepare request body
	body := url.Values{}
	body.Set("asset", params.Asset)
	body.Set("amount", params.Amount)

	// Make request
	resp, err := api.queryPrivate(postUnstakeAsset, http.MethodPost, nil, "application/x-www-form-urlencoded", body, nil, otp, &UnstakeAssetResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*UnstakeAssetResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to UnstakeAssetResponse. Got %T : %#v", resp, resp)
	}

	return result, nil
}

// ListOfStakeableAssets returns the list of assets that the user is able to stake.
func (client *KrakenSpotRESTClient) ListOfStakeableAssets(secopts *SecurityOptions) (*ListOfStakeableAssetsResponse, error) {

	// Use empty value for otp if no second factor provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Make request
	resp, err := api.queryPrivate(postListOfStakeableAssets, http.MethodPost, nil, "", nil, nil, otp, &ListOfStakeableAssetsResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*ListOfStakeableAssetsResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to ListOfStakeableAssetsResponse. Got %T : %#v", resp, resp)
	}

	return result, nil
}

// GetPendingStakingTransactions returns the list of pending staking transactions.
func (client *KrakenSpotRESTClient) GetPendingStakingTransactions(secopts *SecurityOptions) (*GetPendingStakingTransactionsResponse, error) {

	// Use empty value for otp if no second factor provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Make request
	resp, err := api.queryPrivate(postGetPendingStakingTransactions, http.MethodPost, nil, "", nil, nil, otp, &GetPendingStakingTransactionsResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetPendingStakingTransactionsResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to GetPendingStakingTransactionsResponse. Got %T : %#v", resp, resp)
	}

	return result, nil
}

// ListOfStakingTransactions returns the list of 1000 recent staking transactions from past 90 days.
func (client *KrakenSpotRESTClient) ListOfStakingTransactions(secopts *SecurityOptions) (*ListOfStakingTransactionsResponse, error) {

	// Use empty value for otp if no second factor provided
	otp := ""
	if secopts != nil {
		otp = secopts.SecondFactor
	}

	// Make request
	resp, err := api.queryPrivate(postListOfStakingTransactions, http.MethodPost, nil, "", nil, nil, otp, &ListOfStakingTransactionsResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*ListOfStakingTransactionsResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to a list of StakingTransactionInfo. Got %T : %#v", resp, resp)
	}

	return result, nil
}
