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

	"github.com/gbdevw/purple-goctopus/spot/rest/account"
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/funding"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
	"github.com/gbdevw/purple-goctopus/spot/rest/trading"
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
	assetInfoPath          = "/public/Assets"
	tradableAssetPairsPath = "/public/AssetPairs"
	tickerInformationPath  = "/public/Ticker"
	ohlcDataPath           = "/public/OHLC"
	orderBookPath          = "/public/Depth"
	recentTradesPath       = "/public/Trades"
	recentSpreadsPath      = "/public/Spread"

	// Account Data

	getAccountBalancePath     = "/private/Balance"
	getExtendedBalancePath    = "/private/BalanceEx"
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

	// Funding

	getDepositMethodsPath             = "/private/DepositMethods"
	getDepositAddressesPath           = "/private/DepositAddresses"
	getStatusOfRecentDepositsPath     = "/private/DepositStatus"
	getWithdrawalMethodsPath          = "/private/WithdrawMethods"
	getWithdrawalAddressesPath        = "/private/WithdrawAddress"
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
	DefaultUserAgent = "Lake42-Goctopus"
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
	// If an empty string is used, defaults to "Lake42-Goctopus"
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
//     In this case, the returned http.Response will not have its body closed so the reader associated to the response body can be used.
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
			return resp, fmt.Errorf("error code received from Kraken API: %d", resp.StatusCode)
		}
		// Check mime type of response
		mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return resp, fmt.Errorf("could not decode the response Content-Type header: %w", err)
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
				return resp, fmt.Errorf("failed to read response body: %w", err)
			}
			err = json.Unmarshal(body, &receiver)
			if err != nil {
				return resp, fmt.Errorf("failed to parse JSON response: %w", err)
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
		return nil, resp, fmt.Errorf("request for GetServerTime failed: %w", err)
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
		return nil, resp, fmt.Errorf("request for GetSystemStatus failed: %w", err)
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
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, assetInfoPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetAssetInfo: %w", err)
	}
	// Send the request
	receiver := new(market.GetAssetInfoResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetAssetInfo failed: %w", err)
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
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, tradableAssetPairsPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetTradableAssetPairs: %w", err)
	}
	// Send the request
	receiver := new(market.GetTradableAssetPairsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetTradableAssetPairs failed: %w", err)
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
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, tickerInformationPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetTickerInformation: %w", err)
	}
	// Send the request
	receiver := new(market.GetTickerInformationResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetTickerInformation failed: %w", err)
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
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, ohlcDataPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetOHLCData: %w", err)
	}
	// Send the request
	receiver := new(market.GetOHLCDataResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetOHLCData failed: %w", err)
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
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, orderBookPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetOrderBook: %w", err)
	}
	// Send the request
	receiver := new(market.GetOrderBookResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetOrderBook failed: %w", err)
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
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, recentTradesPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetRecentTrades: %w", err)
	}
	// Send the request
	receiver := new(market.GetRecentTradesResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetRecentTrades failed: %w", err)
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
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, recentSpreadsPath, http.MethodGet, query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetRecentSpreads: %w", err)
	}
	// Send the request
	receiver := new(market.GetRecentSpreadsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetRecentSpreads failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

/*****************************************************************************/
/* KRAKEN API CLIENT: OPERATIONS - ACCOUNT DATA                              */
/*****************************************************************************/

// # Description
//
// GetAccountBalance - Retrieve all cash balances, net of pending withdrawals.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetAccountBalanceResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetAccountBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetAccountBalanceResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getAccountBalancePath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetAccountBalance: %w", err)
	}
	// Send the request
	receiver := new(account.GetAccountBalanceResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetAccountBalance failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetExtendedBalance - Retrieve all extended account balances, including credits and held amounts. Balance available
// for trading is calculated as: available balance = balance + credit - credit_used - hold_trade
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetExtendedBalanceResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetExtendedBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetExtendedBalanceResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getExtendedBalancePath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetExtendedBalance: %w", err)
	}
	// Send the request
	receiver := new(account.GetExtendedBalanceResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetExtendedBalance failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetTradeBalance - Retrieve a summary of collateral balances, margin position valuations, equity and margin level.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetTradeBalance request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetTradeBalanceResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetTradeBalance(ctx context.Context, nonce int64, opts *account.GetTradeBalanceRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeBalanceResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if opts.Asset != "" {
			form.Set("asset", opts.Asset)
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getTradeBalancePath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetTradeBalance: %w", err)
	}
	// Send the request
	receiver := new(account.GetTradeBalanceResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetTradeBalance failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetOpenOrders - Retrieve information about currently open orders.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetOpenOrders request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetOpenOrdersResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetOpenOrders(ctx context.Context, nonce int64, opts *account.GetOpenOrdersRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenOrdersResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if opts.Trades {
			form.Set("trades", strconv.FormatBool(opts.Trades))
		}
		if opts.UserReference != nil {
			form.Set("userref", strconv.FormatInt(*opts.UserReference, 10))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getOpenOrdersPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetOpenOrders: %w", err)
	}
	// Send the request
	receiver := new(account.GetOpenOrdersResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetOpenOrders failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetClosedOrders - Retrieve information about orders that have been closed (filled or cancelled). 50 results are returned at a time, the most recent by default.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetClosedOrders request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetClosedOrdersResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetClosedOrders(ctx context.Context, nonce int64, opts *account.GetClosedOrdersOptions, secopts *common.SecurityOptions) (*account.GetClosedOrdersResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if opts.Trades {
			form.Set("trades", strconv.FormatBool(opts.Trades))
		}
		if opts.UserReference != nil {
			form.Set("userref", strconv.FormatInt(*opts.UserReference, 10))
		}
		if opts.Start != "" {
			form.Set("start", opts.Start)
		}
		if opts.End != "" {
			form.Set("end", opts.End)
		}
		if opts.Offset != 0 {
			form.Set("ofs", strconv.FormatInt(opts.Offset, 10))
		}
		if opts.Closetime != "" {
			form.Set("closetime", string(opts.Closetime))
		}
		if opts.ConsolidateTaker {
			form.Set("consolidate_taker", strconv.FormatBool(opts.ConsolidateTaker))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getClosedOrdersPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetClosedOrders: %w", err)
	}
	// Send the request
	receiver := new(account.GetClosedOrdersResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetClosedOrders failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// QueryOrdersInfo - Retrieve information about specific orders.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: QueryOrdersInfo request parameters.
//   - opts: QueryOrdersInfo request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - QueryOrdersInfoResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) QueryOrdersInfo(ctx context.Context, nonce int64, params account.QueryOrdersInfoParameters, opts *account.QueryOrdersInfoRequestOptions, secopts *common.SecurityOptions) (*account.QueryOrdersInfoResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Add transaction ids as a comma separated string
	form.Set("txid", strings.Join(params.TxId, ","))
	// Add options
	if opts != nil {
		if opts.Trades {
			form.Set("trades", strconv.FormatBool(opts.Trades))
		}
		if opts.UserReference != nil {
			form.Set("userref", strconv.FormatInt(*opts.UserReference, 10))
		}
		// A pointer is used as the default is true so we cannot rely on Golang zero value
		if opts.ConsolidateTaker != nil {
			form.Set("consolidate_taker", strconv.FormatBool(*opts.ConsolidateTaker))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, queryOrdersInfosPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for QueryOrdersInfo: %w", err)
	}
	// Send the request
	receiver := new(account.QueryOrdersInfoResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for QueryOrdersInfo failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetTradesHistory - Retrieve information about trades/fills. 50 results are returned at a time, the most recent by default.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetTradesHistory request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetTradesHistoryResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetTradesHistory(ctx context.Context, nonce int64, opts *account.GetTradesHistoryRequestOptions, secopts *common.SecurityOptions) (*account.GetTradesHistoryResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if opts.Trades {
			form.Set("trades", strconv.FormatBool(opts.Trades))
		}
		if opts.Type != "" {
			form.Set("type", opts.Type)
		}
		if opts.Start != "" {
			form.Set("start", opts.Start)
		}
		if opts.End != "" {
			form.Set("end", opts.End)
		}
		if opts.Offset != 0 {
			form.Set("ofs", strconv.FormatInt(opts.Offset, 10))
		}
		if opts.ConsolidateTaker {
			form.Set("consolidate_taker", strconv.FormatBool(opts.ConsolidateTaker))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getTradesHistoryPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetTradesHistory: %w", err)
	}
	// Send the request
	receiver := new(account.GetTradesHistoryResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetTradesHistory failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// QueryTradesInfo - Retrieve information about specific trades/fills.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: QueryTradesInfo request parameters.
//   - opts: QueryTradesInfo request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - QueryTradesInfoResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) QueryTradesInfo(ctx context.Context, nonce int64, params account.QueryTradesRequestParameters, opts *account.QueryTradesRequestOptions, secopts *common.SecurityOptions) (*account.QueryTradesInfoResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Add transaction ids as a comma separated string
	form.Set("txid", strings.Join(params.TransactionIds, ","))
	// Add options
	if opts != nil {
		if opts.Trades {
			form.Set("trades", strconv.FormatBool(opts.Trades))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, queryTradesInfoPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for QueryTradesInfo: %w", err)
	}
	// Send the request
	receiver := new(account.QueryTradesInfoResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for QueryTradesInfo failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetOpenPositions - Get information about open margin positions.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetOpenPositions request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetOpenPositionsResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetOpenPositions(ctx context.Context, nonce int64, opts *account.GetOpenPositionsRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenPositionsRequestOptions, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if len(opts.TransactionIds) > 0 {
			form.Set("txid", strings.Join(opts.TransactionIds, ","))
		}
		if opts.DoCalcs {
			form.Set("docalcs", strconv.FormatBool(opts.DoCalcs))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getOpenPositionsPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetOpenPositions: %w", err)
	}
	// Send the request
	receiver := new(account.GetOpenPositionsRequestOptions)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetOpenPositions failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetLedgersInfo - Retrieve information about ledger entries. 50 results are returned at a time, the most recent by default.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetLedgersInfo request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetLedgersInfoResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetLedgersInfo(ctx context.Context, nonce int64, opts *account.GetLedgersInfoRequestOptions, secopts *common.SecurityOptions) (*account.GetLedgersInfoResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if opts.Assets != nil {
			form.Set("asset", strings.Join(opts.Assets, ","))
		}
		if opts.AssetClass != "" {
			form.Set("aclass", string(opts.AssetClass))
		}
		if opts.Type != "" {
			form.Set("type", string(opts.Type))
		}
		if opts.Start != "" {
			form.Set("start", opts.Start)
		}
		if opts.End != "" {
			form.Set("end", opts.End)
		}
		if opts.Offset != 0 {
			form.Set("ofs", strconv.FormatInt(opts.Offset, 10))
		}
		if opts.WithoutCount {
			form.Set("without_count", strconv.FormatBool(opts.WithoutCount))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getLedgersInfoPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetLedgersInfo: %w", err)
	}
	// Send the request
	receiver := new(account.GetLedgersInfoResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetLedgersInfo failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// QueryLedgers - Get the current system status or trading mode.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: QueryLedgers request parameters.
//   - opts: QueryLedgers request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - QueryLedgersResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) QueryLedgers(ctx context.Context, nonce int64, params account.QueryLedgersRequestParameters, opts *account.QueryLedgersOptions, secopts *common.SecurityOptions) (*account.QueryLedgersResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Add transaction ids as a comma separated string
	form.Set("id", strings.Join(params.Id, ","))
	// Add options
	if opts != nil {
		if opts.Trades {
			form.Set("trades", strconv.FormatBool(opts.Trades))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, queryLedgersPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for QueryLedgers: %w", err)
	}
	// Send the request
	receiver := new(account.QueryLedgersResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for QueryLedgers failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetTradeVolume - Returns 30 day USD trading volume and resulting fee schedule for any asset pair(s) provided.
//
// Note: If an asset pair is on a maker/taker fee schedule, the taker side is given in fees and maker side in
// fees_maker. For pairs not on maker/taker, they will only be given in fees.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetTradeVolume request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetTradeVolumeResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetTradeVolume(ctx context.Context, nonce int64, opts *account.GetTradeVolumeRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeVolumeResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if len(opts.Pairs) > 0 {
			form.Set("pair", strings.Join(opts.Pairs, ","))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getTradeVolumePath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetTradeVolume: %w", err)
	}
	// Send the request
	receiver := new(account.GetTradeVolumeResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetTradeVolume failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// RequestExportReport - Request export of trades or ledgers.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: RequestExportReport request parameters.
//   - opts: RequestExportReport request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - RequestExportReportResponse: The parsed response from Kraken API.
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
// # Note on the http.Response.
//
// A reference to the received http.Response is always returned but it may be nil if no response was received.
// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
// to extract the metadata (or any other kind of data that are not used by the API client directly).
//
// Please note response body will always be closed except for RetrieveDataExport.
func (client *KrakenSpotRESTClient) RequestExportReport(ctx context.Context, nonce int64, params account.RequestExportReportRequestParameters, opts *account.RequestExportReportRequestOptions, secopts *common.SecurityOptions) (*account.RequestExportReportResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	form.Set("report", params.Report)
	form.Set("description", params.Description)
	// Add options
	if opts != nil {
		if opts.Format != "" {
			form.Set("format", opts.Format)
		}
		if opts.Fields != nil {
			form.Set("fields", strings.Join(opts.Fields, ","))
		}
		if opts.StartTm != 0 {
			form.Set("starttm", strconv.FormatInt(opts.StartTm, 10))
		}
		if opts.EndTm != 0 {
			form.Set("endtm", strconv.FormatInt(opts.EndTm, 10))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, requestExportReportPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for RequestExportReport: %w", err)
	}
	// Send the request
	receiver := new(account.RequestExportReportResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for RequestExportReport failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetExportReportStatus - Get status of requested data exports.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: GetExportReportStatus request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetExportReportStatusResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetExportReportStatus(ctx context.Context, nonce int64, params account.GetExportReportStatusRequestParameters, secopts *common.SecurityOptions) (*account.GetExportReportStatusResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	form.Set("report", params.Report)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getExportReportStatusPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetExportReportStatus: %w", err)
	}
	// Send the request
	receiver := new(account.GetExportReportStatusResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetExportReportStatus failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// RetrieveDataExport - Retrieve a processed data export.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: RetrieveDataExport request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - RetrieveDataExportResponse: The response contains an io.Reader that is tied to the http.Response body in order to let users download data in a streamed fashion.
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
// Please note response body will not be closed in order to allow users to download the exported data in a streamed fashion.
// The io.Reader in the response is tied to the http.Response.Body.
func (client *KrakenSpotRESTClient) RetrieveDataExport(ctx context.Context, nonce int64, params account.RetrieveDataExportParameters, secopts *common.SecurityOptions) (*account.RetrieveDataExportResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	form.Set("id", params.Id)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, retrieveDataExportPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for RetrieveDataExport: %w", err)
	}
	// Send the request
	receiver := new(account.RetrieveDataExportResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, nil)
	if err != nil {
		return nil, resp, fmt.Errorf("request for RetrieveDataExport failed: %w", err)
	}
	// Assign the response body reader to the API response that will be returned
	receiver.Report = resp.Body
	// Return results
	return receiver, resp, nil
}

// # Description
//
// DeleteExportReport - Delete exported trades/ledgers report.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: DeleteExportReport request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - DeleteExportReportResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) DeleteExportReport(ctx context.Context, nonce int64, params account.DeleteExportReportRequestParameters, secopts *common.SecurityOptions) (*account.DeleteExportReportResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	form.Set("id", params.Id)
	form.Set("type", params.Type)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, deleteExportReportPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for DeleteExportReport: %w", err)
	}
	// Send the request
	receiver := new(account.DeleteExportReportResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for DeleteExportReport failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

/*****************************************************************************/
/* KRAKEN API CLIENT: OPERATIONS - TRADING                                   */
/*****************************************************************************/

// # Description
//
// AddOrder - Place a new order.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: AddOrder request parameters.
//   - opts: AddOrder request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - AddOrderResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) AddOrder(ctx context.Context, nonce int64, params trading.AddOrderRequestParameters, opts *trading.AddOrderRequestOptions, secopts *common.SecurityOptions) (*trading.AddOrderResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Set targeted asset pair
	form.Set("pair", params.Pair)
	// Add user reference if defined
	if params.Order.UserReference != nil {
		form.Set("userref", strconv.FormatInt(*params.Order.UserReference, 10))
	}
	// Set order type
	form.Set("ordertype", params.Order.OrderType)
	// Set order direction
	form.Set("type", params.Order.Type)
	// Set volume
	form.Set("volume", params.Order.Volume)
	// Set price if not empty
	if params.Order.Price != "" {
		form.Set("price", params.Order.Price)
	}
	// Set price2 if not empty
	if params.Order.Price2 != "" {
		form.Set("price2", params.Order.Price2)
	}
	// Set Trigger if value is not empty
	if params.Order.Trigger != "" {
		form.Set("trigger", params.Order.Trigger)
	}
	// Set leverage if provided value is not empty
	if params.Order.Leverage != "" {
		form.Set("leverage", params.Order.Leverage)
	}
	// Set STP flag if not empty
	if params.Order.StpType != "" {
		form.Set("stp_type", params.Order.StpType)
	}
	// Set Reduce only if set
	if params.Order.ReduceOnly {
		form.Set("reduce_only", strconv.FormatBool(params.Order.ReduceOnly))
	}
	// Set operation flags as a comma separated list if not empty
	if params.Order.OrderFlags != "" {
		form.Set("oflags", params.Order.OrderFlags)
	}
	// Set time in force if defined
	if params.Order.TimeInForce != "" {
		form.Set("timeinforce", params.Order.TimeInForce)
	}
	// Set start time if not empty
	if params.Order.ScheduledStartTime != "" {
		form.Set("starttm", params.Order.ScheduledStartTime)
	}
	// Set expire time if not empty
	if params.Order.ExpirationTime != "" {
		form.Set("expiretm", params.Order.ExpirationTime)
	}
	// Set close order if defined
	if params.Order.Close != nil {
		// Set order type
		form.Set("close[ordertype]", string(params.Order.Close.OrderType))
		// Set close order price
		form.Set("close[price]", params.Order.Close.Price)
		// Set price2 if not empty
		if params.Order.Close.Price2 != "" {
			form.Set("close[price2]", params.Order.Close.Price2)
		}
	}
	// Add options
	if opts != nil {
		// Set deadline if defined
		if !opts.Deadline.IsZero() {
			form.Set("deadline", opts.Deadline.Format(time.RFC3339))
		}
		// Set validate
		form.Set("validate", strconv.FormatBool(opts.Validate))
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, addOrderPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for AddOrder: %w", err)
	}
	// Send the request
	receiver := new(trading.AddOrderResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for AddOrder failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// AddOrderBatch - Get the current system status or trading mode.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: AddOrderBatch request parameters.
//   - opts: AddOrderBatch request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//   - AddOrderBatchResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) AddOrderBatch(ctx context.Context, nonce int64, params trading.AddOrderBatchRequestParameters, opts *trading.AddOrderBatchOptions, secopts *common.SecurityOptions) (*trading.AddOrderBatchResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Set targeted asset pair
	form.Set("pair", params.Pair)
	// Set orders
	for index, order := range params.Orders {
		// Add user reference if defined
		if order.UserReference != nil {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "userref"), strconv.FormatInt(*order.UserReference, 10))
		}

		// Set order type
		form.Set(fmt.Sprintf("orders[%d][%s]", index, "ordertype"), order.OrderType)

		// Set order direction
		form.Set(fmt.Sprintf("orders[%d][%s]", index, "type"), order.Type)

		// Set volume
		form.Set(fmt.Sprintf("orders[%d][%s]", index, "volume"), order.Volume)

		// Set price if not empty
		if order.Price != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "price"), order.Price)
		}

		// Set price2 if not empty
		if order.Price2 != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "price2"), order.Price2)
		}

		// Set Trigger if value is not empty
		if order.Trigger != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "trigger"), order.Trigger)
		}

		// Set leverage if provided value is not empty
		if order.Leverage != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "leverage"), order.Leverage)
		}

		// Set STP flag if not empty
		if order.StpType != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "stp_type"), order.StpType)
		}

		// Set Reduce only if set
		if order.ReduceOnly {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "reduce_only"), strconv.FormatBool(order.ReduceOnly))
		}

		// Set operation flags as a comma separated list if not empty
		if order.OrderFlags != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "oflags"), order.OrderFlags)
		}

		// Set time in force if defined
		if order.TimeInForce != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "timeinforce"), order.TimeInForce)
		}

		// Set start time if not empty
		if order.ScheduledStartTime != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "starttm"), order.ScheduledStartTime)
		}

		// Set expire time if not empty
		if order.ExpirationTime != "" {
			form.Set(fmt.Sprintf("orders[%d][%s]", index, "expiretm"), order.ExpirationTime)
		}
		// Set close order if defined
		if order.Close != nil {
			// Set order type
			form.Set(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "ordertype"), order.Close.OrderType)
			// Set close order price
			form.Set(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price"), order.Close.Price)
			// Set price2 if not empty
			if order.Close.Price2 != "" {
				form.Set(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price2"), order.Close.Price2)
			}
		}
	}
	// Add options
	if opts != nil {
		// Set deadline if defined
		if !opts.Deadline.IsZero() {
			form.Set("deadline", opts.Deadline.Format(time.RFC3339))
		}
		// Set validate
		form.Set("validate", strconv.FormatBool(opts.Validate))
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, addOrderBatchPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for AddOrderBatch: %w", err)
	}
	// Send the request
	receiver := new(trading.AddOrderBatchResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for AddOrderBatch failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// EditOrder - Edit volume and price on open orders. Uneditable orders include triggered
// stop/profit orders, orders with conditional close terms attached, those already cancelled
// or filled, and those where the executed volume is greater than the newly supplied volume.
// post-only flag is not retained from original order after successful edit. post-only needs
// to be explicitly set on edit request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: EditOrder request parameters.
//   - opts: EditOrder request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - EditOrderResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) EditOrder(ctx context.Context, nonce int64, params trading.EditOrderRequestParameters, opts *trading.EditOrderRequestOptions, secopts *common.SecurityOptions) (*trading.EditOrderResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Set txid
	form.Set("txid", params.Id)
	// Set targeted asset pair
	form.Set("pair", params.Pair)
	// Add options
	if opts != nil {
		// Set userref if defined
		if opts.NewUserReference != "" {
			form.Set("userref", opts.NewUserReference)
		}
		// Set volume if defined
		if opts.NewVolume != "" {
			form.Set("volume", opts.NewVolume)
		}

		// Set price if not empty
		if opts.Price != "" {
			form.Set("price", opts.Price)
		}

		// Set price2 if not empty
		if opts.Price2 != "" {
			form.Set("price2", opts.Price2)
		}

		// Set oflags if not nil
		if opts.OFlags != nil {
			// oflags is a comma separated list
			form.Set("oflags", strings.Join(opts.OFlags, ","))
		}

		// Set deadline if defined
		if !opts.Deadline.IsZero() {
			form.Set("deadline", opts.Deadline.Format(time.RFC3339))
		}

		// Set cancel_response
		form.Set("cancel_response", strconv.FormatBool(opts.CancelResponse))

		// Set validate
		form.Set("validate", strconv.FormatBool(opts.Validate))
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, editOrderPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for EditOrder: %w", err)
	}
	// Send the request
	receiver := new(trading.EditOrderResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for EditOrder failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// CancelOrder - Cancel a particular open order (or set of open orders) by txid or userref.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: CancelOrder request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - CancelOrderResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) CancelOrder(ctx context.Context, nonce int64, params trading.CancelOrderRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Set txid
	form.Set("txid", params.Id)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, cancelOrderPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for CancelOrder: %w", err)
	}
	// Send the request
	receiver := new(trading.CancelOrderResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for CancelOrder failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// CancelAllOrders - Cancel all open orders.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - CancelAllOrdersResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) CancelAllOrders(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*trading.CancelAllOrdersResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, cancelAllOrdersPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for CancelAllOrders: %w", err)
	}
	// Send the request
	receiver := new(trading.CancelAllOrdersResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for CancelAllOrders failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// CancelAllOrdersAfterX - CancelAllOrdersAfter provides a "Dead Man's Switch" mechanism to
// protect the client from network malfunction, extreme latency or unexpected matching engine
// downtime. The client can send a request with a timeout (in seconds), that will start a
// countdown timer which will cancel all client orders when the timer expires. The client has
// to keep sending new requests to push back the trigger time, or deactivate the mechanism by
// specifying a timeout of 0. If the timer expires, all orders are cancelled and then the timer
// remains disabled until the client provides a new (non-zero) timeout.
//
// The recommended use is to make a call every 15 to 30 seconds, providing a timeout of 60
// seconds. This allows the client to keep the orders in place in case of a brief disconnection
// or transient delay, while keeping them safe in case of a network breakdown. It is also
// recommended to disable the timer ahead of regularly scheduled trading engine maintenance (if
// the timer is enabled, all orders will be cancelled when the trading engine comes back from
// downtime - planned or otherwise).
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: CancelAllOrdersAfterX request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - CancelAllOrdersAfterXResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) CancelAllOrdersAfterX(ctx context.Context, nonce int64, params trading.CancelAllOrdersAfterXRequestParameters, secopts *common.SecurityOptions) (*trading.CancelAllOrdersAfterXResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Set timeout
	form.Set("timeout", strconv.FormatInt(params.Timeout, 10))
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, cancelAllOrdersAfterXPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for CancelAllOrdersAfterX: %w", err)
	}
	// Send the request
	receiver := new(trading.CancelAllOrdersAfterXResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for CancelAllOrdersAfterX failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// CancelOrderBatch - Cancel multiple open orders by txid or userref (maximum 50 total unique IDs/references)
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: CancelOrderBatch request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - CancelOrderBatchResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) CancelOrderBatch(ctx context.Context, nonce int64, params trading.CancelOrderBatchRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderBatchResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	// Set orders as a comma delimited list
	form.Set("orders", strings.Join(params.OrderIds, ","))
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, cancelOrderBatchPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for CancelOrderBatch: %w", err)
	}
	// Send the request
	receiver := new(trading.CancelOrderBatchResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for CancelOrderBatch failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

/*****************************************************************************/
/* KRAKEN API CLIENT: OPERATIONS - FUNDING                                   */
/*****************************************************************************/

// # Description
//
// GetDepositMethods - Retrieve methods available for depositing a particular asset.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: CancelOrderBatch request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetDepositMethodsResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetDepositMethods(ctx context.Context, nonce int64, params funding.GetDepositMethodsRequestParameters, secopts *common.SecurityOptions) (*funding.GetDepositMethodsResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	form.Set("asset", params.Asset)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getDepositMethodsPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetDepositMethods: %w", err)
	}
	// Send the request
	receiver := new(funding.GetDepositMethodsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetDepositMethods failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetDepositAddresses - Retrieve (or generate a new) deposit addresses for a particular asset and method.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: GetDepositAddresses request parameters.
//   - opts: GetDepositAddresses request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetDepositAddressesResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetDepositAddresses(ctx context.Context, nonce int64, params funding.GetDepositAddressesRequestParameters, opts *funding.GetDepositAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetDepositAddressesResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add parameters
	form.Set("asset", params.Asset)
	form.Set("method", params.Method)
	// Add options
	if opts != nil {
		if opts.New {
			form.Set("new", strconv.FormatBool(opts.New))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getDepositAddressesPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetDepositAddresses: %w", err)
	}
	// Send the request
	receiver := new(funding.GetDepositAddressesResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetDepositAddresses failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetStatusOfRecentDeposits - Retrieve information about recent deposits. Results are sorted
// by recency, use the cursor parameter to iterate through list of deposits (page size equal
// to value of limit) from newest to oldest.
//
// Please note pagination usage is forced as the response format is too different.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetStatusOfRecentDeposits request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetStatusOfRecentDepositsResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetStatusOfRecentDeposits(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentDepositsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentDepositsResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Force pagination usage as response is too different otherwise
	form.Set("cursor", strconv.FormatBool(true))
	// Add options
	if opts != nil {
		if opts.Asset != "" {
			form.Set("asset", opts.Asset)
		}
		if opts.Method != "" {
			form.Set("method", opts.Method)
		}
		if opts.Start != "" {
			form.Set("start", opts.Start)
		}
		if opts.End != "" {
			form.Set("end", opts.End)
		}
		if opts.Cursor != "" && opts.Cursor != strconv.FormatBool(false) {
			// Override cursor and ensure it is not deactivated
			form.Set("cursor", opts.Cursor)
		}
		if opts.Limit != 0 {
			form.Set("limit", strconv.FormatInt(opts.Limit, 10))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getStatusOfRecentDepositsPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetStatusOfRecentDeposits: %w", err)
	}
	// Send the request
	receiver := new(funding.GetStatusOfRecentDepositsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetStatusOfRecentDeposits failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetWithdrawalMethods - Retrieve a list of withdrawal addresses available for the user.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetWithdrawalMethods request options.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetWithdrawalMethodsResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetWithdrawalMethods(ctx context.Context, nonce int64, opts *funding.GetWithdrawalMethodsRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalMethodsResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if opts.Asset != "" {
			form.Set("asset", opts.Asset)
		}
		if opts.Network != "" {
			form.Set("network", opts.Network)
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getWithdrawalMethodsPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetWithdrawalMethods: %w", err)
	}
	// Send the request
	receiver := new(funding.GetWithdrawalMethodsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetWithdrawalMethods failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetWithdrawalAddresses - Retrieve a list of withdrawal addresses available for the user.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetWithdrawalAddresses request options.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetWithdrawalAddressesResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetWithdrawalAddresses(ctx context.Context, nonce int64, opts *funding.GetWithdrawalAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalAddressesResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add options
	if opts != nil {
		if opts.Asset != "" {
			form.Set("asset", opts.Asset)
		}
		if opts.Method != "" {
			form.Set("method", opts.Method)
		}
		if opts.Key != "" {
			form.Set("key", opts.Key)
		}
		if opts.Verified {
			form.Set("verified", strconv.FormatBool(opts.Verified))
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getWithdrawalAddressesPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetWithdrawalAddresses: %w", err)
	}
	// Send the request
	receiver := new(funding.GetWithdrawalAddressesResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetWithdrawalAddresses failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetWithdrawalInformation - Retrieve a list of withdrawal methods available for the user.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: GetWithdrawalInformation request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetWithdrawalInformationResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetWithdrawalInformation(ctx context.Context, nonce int64, params funding.GetWithdrawalInformationRequestParameters, secopts *common.SecurityOptions) (*funding.GetWithdrawalInformationResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add params
	form.Set("asset", params.Asset)
	form.Set("key", params.Key)
	form.Set("amount", params.Amount)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getWithdrawalInformationPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetWithdrawalInformation: %w", err)
	}
	// Send the request
	receiver := new(funding.GetWithdrawalInformationResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetWithdrawalInformation failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// WithdrawFunds - Make a withdrawal request.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: WithdrawFunds request parameters.
//   - opts: WithdrawFunds request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - WithdrawFundsResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) WithdrawFunds(ctx context.Context, nonce int64, params funding.WithdrawFundsRequestParameters, opts *funding.WithdrawFundsRequestOptions, secopts *common.SecurityOptions) (*funding.WithdrawFundsResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add params
	form.Set("asset", params.Asset)
	form.Set("key", params.Key)
	form.Set("amount", params.Amount)
	// Add options
	if opts != nil {
		if opts.Address != "" {
			form.Set("address", opts.Address)
		}
		if opts.MaxFee != "" {
			form.Set("max_fee", opts.MaxFee)
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, withdrawFundsPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for WithdrawFunds: %w", err)
	}
	// Send the request
	receiver := new(funding.WithdrawFundsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for WithdrawFunds failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// GetStatusOfRecentWithdrawals - Retrieve information about recent withdrawals. Results are
// sorted by recency, use the cursor parameter to iterate through list of withdrawals (page
// size equal to value of limit) from newest to oldest.
//
// Please note pagination is not used as documentation is unclear about the response format
// in this case.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - opts: GetStatusOfRecentWithdrawals request options. A nil value triggers all default behaviors.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - GetStatusOfRecentWithdrawalsResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) GetStatusOfRecentWithdrawals(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentWithdrawalsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentWithdrawalsResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Force deactivated pagination as documentation is unclear about the response payload.
	form.Set("cursor", strconv.FormatBool(false))
	// Add options
	if opts != nil {
		if opts.Asset != "" {
			form.Set("asset", opts.Asset)
		}
		if opts.Method != "" {
			form.Set("max_fee", opts.Method)
		}
		if opts.Start != "" {
			form.Set("start", opts.Start)
		}
		if opts.End != "" {
			form.Set("end", opts.End)
		}
	}
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, getStatusOfRecentWithdrawalsPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for GetStatusOfRecentWithdrawals: %w", err)
	}
	// Send the request
	receiver := new(funding.GetStatusOfRecentWithdrawalsResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for GetStatusOfRecentWithdrawals failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// RequestWithdrawalCancellation - Cancel a recently requested withdrawal, if it has not already been successfully processed.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: RequestWithdrawalCancellation request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - RequestWithdrawalCancellationResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) RequestWithdrawalCancellation(ctx context.Context, nonce int64, params funding.RequestWithdrawalCancellationRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWithdrawalCancellationResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add params
	form.Set("asset", params.Asset)
	form.Set("refid", params.ReferenceId)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, requestWithdrawalCancellationPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for RequestWithdrawalCancellation: %w", err)
	}
	// Send the request
	receiver := new(funding.RequestWithdrawalCancellationResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for RequestWithdrawalCancellation failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
}

// # Description
//
// RequestWalletTransfer - Transfer from a Kraken spot wallet to a Kraken Futures wallet. Note that a
// transfer in the other direction must be requested via the Kraken Futures API endpoint for
// withdrawals to Spot wallets.
//
// # Inputs
//
//   - ctx: Context used for tracing and coordination purpose.
//   - nonce: Nonce used to sign request.
//   - params: RequestWalletTransfer request parameters.
//   - secopts: Security options to use for the API call (2FA, ...)
//
// # Returns
//
//   - RequestWalletTransferResponse: The parsed response from Kraken API.
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
func (client *KrakenSpotRESTClient) RequestWalletTransfer(ctx context.Context, nonce int64, params funding.RequestWalletTransferRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWalletTransferResponse, *http.Response, error) {
	// Prepare form body.
	form := url.Values{}
	// Add nonce
	form.Set("nonce", strconv.FormatInt(nonce, 10))
	// Use 2FA if provided
	if secopts != nil {
		form.Set("otp", secopts.SecondFactor)
	}
	// Add params
	form.Set("asset", params.Asset)
	form.Set("from", params.From)
	form.Set("to", params.To)
	form.Set("amount", params.Amount)
	// Forge and authorize the request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(ctx, requestWalletTransferPath, http.MethodPost, nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to forge and authorize request for RequestWalletTransfer: %w", err)
	}
	// Send the request
	receiver := new(funding.RequestWalletTransferResponse)
	resp, err := client.doKrakenAPIRequest(ctx, req, receiver)
	if err != nil {
		return nil, resp, fmt.Errorf("request for RequestWalletTransfer failed: %w", err)
	}
	// Return results
	return receiver, resp, nil
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
