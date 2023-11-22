package rest

import (
	"context"
	"net/http"

	"github.com/gbdevw/purple-goctopus/spot/rest/account"
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
)

/*************************************************************************************************/
/* INTERFACE                                                                                     */
/*************************************************************************************************/

// Interface for Kraken Spot REST API client.
type KrakenSpotRESTClientIface interface {
	// # Description
	//
	// GetServerTime - Get the server time.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//
	// # Returns
	//
	//	- GetServerTimeResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetServerTime(ctx context.Context) (*market.GetServerTimeResponse, *http.Response, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//
	// # Returns
	//
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetSystemStatus(ctx context.Context) (*market.GetSystemStatusResponse, *http.Response, error)
	// # Description
	//
	// GetAssetInfo - Get information about the assets that are available for deposit, withdrawal, trading and staking.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- opts: GetAssetInfo request options. A nil value triggers all default behaviors.
	//
	// # Returns
	//
	//	- GetAssetInfoResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetAssetInfo(ctx context.Context, opts *market.GetAssetInfoRequestOptions) (*market.GetAssetInfoResponse, error)
	// # Description
	//
	// GetTradableAssetPairs - Get tradable asset pairs
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- opts: GetTradableAssetPairs request options.
	//
	// # Returns
	//
	//	- GetTradableAssetPairsResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetTradableAssetPairs(ctx context.Context, opts *market.GetTradableAssetPairsRequestOptions) (*market.GetTradableAssetPairsResponse, *http.Response, error)
	// # Description
	//
	// GetTickerInformation - Get data about today's price. Today's prices start at midnight UTC.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- opts: GetTickerInformation request options
	//
	// # Returns
	//
	//	- GetTickerInformationResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetTickerInformation(ctx context.Context, opts *market.GetTickerInformationRequestOptions) (*market.GetTickerInformationResponse, *http.Response, error)
	// # Description
	//
	// GetOHLCData - Return up to 720 OHLC data points since now or since given timestamp.
	//
	// Note: the last entry in the OHLC array is for the current, not-yet-committed frame and will always be present,
	// regardless of the value of since.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- params: GetOHLCData request parameters. Must not be nil.
	//	- opts: GetOHLCData request options.
	//
	// # Returns
	//
	//	- GetOHLCDataResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetOHLCData(ctx context.Context, params *market.GetOHLCDataRequestParameters, opts *market.GetOHLCDataRequestOptions) (*market.GetOHLCDataResponse, *http.Response, error)
	// # Description
	//
	// GetOrderBook - Get the target market order book.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- params: GetOrderBook request parameters. Must not be nil.
	//	- opts: GetOrderBook request options.
	//
	// # Returns
	//
	//	- GetOrderBookResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetOrderBook(ctx context.Context, params market.GetOrderBookRequestParameters, opts *market.GetOrderBookRequestOptions) (*market.GetOrderBookResponse, *http.Response, error)
	// # Description
	//
	// GetRecentTrades - Returns up to the last 1000 trades since now or since given timestamp.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//	- params: GetRecentTrades request parameters. Must not be nil.
	//	- opts: GetRecentTrades request options.
	//
	// # Returns
	//
	//	- GetRecentTradesResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
	//
	// # Note on error
	//
	// The error is set only when something wrong has happened either at the HTTP level (while building the request,
	// when the server is unreachable, when the API replies with a status code different from 200, ...) , when
	// an error happens while parsing the response JSON payload (in that case, error is json.UnmarshalTypeError) or
	//  when context has expired.
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
	// Please note response body will always be closed.
	GetRecentTrades(ctx context.Context, params *market.GetRecentTradesRequestParameters, opts *market.GetRecentTradesRequestOptions) (*market.GetRecentTradesResponse, *http.Response, error)
	// # Description
	//
	// GetRecentSpreads - Returns the last ~200 top-of-book spreads for a given pair as for now as as a given timestamp.
	//
	// Note: Intended for incremental updates within available dataset (does not contain all historical spreads).
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//	- params: GetRecentSpreads request parameters. Must not be nil.
	//	- opts: GetRecentSpreads request options.
	//
	// # Returns
	//	- GetRecentSpreadsResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetRecentSpreads(ctx context.Context, params *market.GetRecentSpreadsRequestParameters, opts *market.GetRecentSpreadsRequestOptions) (*market.GetRecentSpreadsResponse, *http.Response, error)
	// # Description
	//
	// GetAccountBalance - Retrieve all cash balances, net of pending withdrawals.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//	- GetAccountBalanceResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetAccountBalance(ctx context.Context, secopts *common.SecurityOptions) (*account.GetAccountBalanceResponse, *http.Response, error)

	// RESUME HERE: Add Extended Balance

	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetTradeBalance(ctx context.Context, opts *GetTradeBalanceOptions, secopts *SecurityOptions) (*GetTradeBalanceResponse, error)

	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetOpenOrders(ctx context.Context, opts *GetOpenOrdersOptions, secopts *SecurityOptions) (*GetOpenOrdersResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetClosedOrders(ctx context.Context, opts *GetClosedOrdersOptions, secopts *SecurityOptions) (*GetClosedOrdersResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	QueryOrdersInfo(ctx context.Context, params QueryOrdersParameters, opts *QueryOrdersOptions, secopts *SecurityOptions) (*QueryOrdersInfoResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetTradesHistory(ctx context.Context, opts *GetTradesHistoryOptions, secopts *SecurityOptions) (*GetTradesHistoryResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	QueryTradesInfo(ctx context.Context, params QueryTradesParameters, opts *QueryTradesOptions, secopts *SecurityOptions) (*QueryTradesInfoResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetOpenPositions(ctx context.Context, opts *GetOpenPositionsOptions, secopts *SecurityOptions) (*GetOpenPositionsResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetLedgersInfo(ctx context.Context, opts *GetLedgersInfoOptions, secopts *SecurityOptions) (*GetLedgersInfoResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	QueryLedgers(ctx context.Context, params QueryLedgersParameters, opts *QueryLedgersOptions, secopts *SecurityOptions) (*QueryLedgersResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	RequestExportReport(ctx context.Context, params RequestExportReportParameters, opts *RequestExportReportOptions, secopts *SecurityOptions) (*RequestExportReportResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetExportReportStatus(ctx context.Context, params GetExportReportStatusParameters, secopts *SecurityOptions) (*GetExportReportStatusResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	RetrieveDataExport(ctx context.Context, params RetrieveDataExportParameters, secopts *SecurityOptions) (*RetrieveDataExportResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	DeleteExportReport(ctx context.Context, params DeleteExportReportParameters, secopts *SecurityOptions) (*DeleteExportReportResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	AddOrder(ctx context.Context, params AddOrderParameters, opts *AddOrderOptions, secopts *SecurityOptions) (*AddOrderResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	AddOrderBatch(ctx context.Context, params AddOrderBatchParameters, opts *AddOrderBatchOptions, secopts *SecurityOptions) (*AddOrderBatchResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	EditOrder(ctx context.Context, params EditOrderParameters, opts *EditOrderOptions, secopts *SecurityOptions) (*EditOrderResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	CancelOrder(ctx context.Context, params CancelOrderParameters, secopts *SecurityOptions) (*CancelOrderResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	CancelAllOrders(ctx context.Context, secopts *SecurityOptions) (*CancelAllOrdersResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	CancelAllOrdersAfterX(ctx context.Context, params CancelCancelAllOrdersAfterXParameters, secopts *SecurityOptions) (*CancelAllOrdersAfterXResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	CancelOrderBatch(ctx context.Context, params CancelOrderBatchParameters, secopts *SecurityOptions) (*CancelOrderBatchResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetDepositMethods(ctx context.Context, params GetDepositMethodsParameters, secopts *SecurityOptions) (*GetDepositMethodsResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetDepositAddresses(ctx context.Context, params GetDepositAddressesParameters, opts *GetDepositAddressesOptions, secopts *SecurityOptions) (*GetDepositAddressesResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetStatusOfRecentDeposits(ctx context.Context, params GetStatusOfRecentDepositsParameters, opts *GetStatusOfRecentDepositsOptions, secopts *SecurityOptions) (*GetStatusOfRecentDepositsResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetWithdrawalInformation(ctx context.Context, params GetWithdrawalInformationParameters, secopts *SecurityOptions) (*GetWithdrawalInformationResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	WithdrawFunds(ctx context.Context, params WithdrawFundsParameters, secopts *SecurityOptions) (*WithdrawFundsResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	GetStatusOfRecentWithdrawals(ctx context.Context, params GetStatusOfRecentWithdrawalsParameters, opts *GetStatusOfRecentWithdrawalsOptions, secopts *SecurityOptions) (*GetStatusOfRecentWithdrawalsResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	RequestWithdrawalCancellation(ctx context.Context, params RequestWithdrawalCancellationParameters, secopts *SecurityOptions) (*RequestWithdrawalCancellationResponse, error)
	// # Description
	//
	// GetSystemStatus - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//
	// # Returns
	//	- GetSystemStatusResponse: The parsed response from Kraken API.
	//	- http.Response: A reference to the raw HTTP response received from Kraken API.
	//	- error: An error in case the HTTP request failed, response JSON payload could not be parsed or context has expired.
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
	// Please note response body will always be closed.
	RequestWalletTransfer(ctx context.Context, params RequestWalletTransferParameters, secopts *SecurityOptions) (*RequestWalletTransferResponse, error)
}
