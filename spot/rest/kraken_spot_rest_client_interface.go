package rest

import (
	"context"
	"net/http"

	"github.com/gbdevw/purple-goctopus/spot/rest/account"
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/earn"
	"github.com/gbdevw/purple-goctopus/spot/rest/funding"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
	"github.com/gbdevw/purple-goctopus/spot/rest/trading"
	"github.com/gbdevw/purple-goctopus/spot/rest/websocket"
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetAssetInfo(ctx context.Context, opts *market.GetAssetInfoRequestOptions) (*market.GetAssetInfoResponse, *http.Response, error)
	// # Description
	//
	// GetTradableAssetPairs - Get tradable asset pairs
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- opts: GetTradableAssetPairs request options. A nil value triggers all default behaviors.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	//	- params: GetOHLCData request parameters.
	//	- opts: GetOHLCData request options. A nil value triggers all default behaviors.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetOHLCData(ctx context.Context, params market.GetOHLCDataRequestParameters, opts *market.GetOHLCDataRequestOptions) (*market.GetOHLCDataResponse, *http.Response, error)
	// # Description
	//
	// GetOrderBook - Get the target market order book.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- params: GetOrderBook request parameters.
	//	- opts: GetOrderBook request options. A nil value triggers all default behaviors.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetOrderBook(ctx context.Context, params market.GetOrderBookRequestParameters, opts *market.GetOrderBookRequestOptions) (*market.GetOrderBookResponse, *http.Response, error)
	// # Description
	//
	// GetRecentTrades - Returns up to the last 1000 trades since now or since given timestamp.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//	- params: GetRecentTrades request parameters.
	//	- opts: GetRecentTrades request options. A nil value triggers all default behaviors.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetRecentTrades(ctx context.Context, params market.GetRecentTradesRequestParameters, opts *market.GetRecentTradesRequestOptions) (*market.GetRecentTradesResponse, *http.Response, error)
	// # Description
	//
	// GetRecentSpreads - Returns the last ~200 top-of-book spreads for a given pair as for now as as a given timestamp.
	//
	// Note: Intended for incremental updates within available dataset (does not contain all historical spreads).
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose
	//	- params: GetRecentSpreads request parameters.
	//	- opts: GetRecentSpreads request options. A nil value triggers all default behaviors.
	//
	// # Returns
	//
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetRecentSpreads(ctx context.Context, params market.GetRecentSpreadsRequestParameters, opts *market.GetRecentSpreadsRequestOptions) (*market.GetRecentSpreadsResponse, *http.Response, error)
	// # Description
	//
	// GetAccountBalance - Retrieve all cash balances, net of pending withdrawals.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetAccountBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetAccountBalanceResponse, *http.Response, error)
	// # Description
	//
	// GetExtendedBalance - Retrieve all extended account balances, including credits and held amounts. Balance available
	// for trading is calculated as: available balance = balance + credit - credit_used - hold_trade
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetExtendedBalanceResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetExtendedBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetExtendedBalanceResponse, *http.Response, error)
	// # Description
	//
	// GetTradeBalance - Retrieve a summary of collateral balances, margin position valuations, equity and margin level.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetTradeBalance request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetTradeBalanceResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetTradeBalance(ctx context.Context, nonce int64, opts *account.GetTradeBalanceRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeBalanceResponse, *http.Response, error)
	// # Description
	//
	// GetOpenOrders - Retrieve information about currently open orders.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetOpenOrders request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetOpenOrdersResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetOpenOrders(ctx context.Context, nonce int64, opts *account.GetOpenOrdersRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenOrdersResponse, *http.Response, error)
	// # Description
	//
	// GetClosedOrders - Retrieve information about orders that have been closed (filled or cancelled). 50 results are returned at a time, the most recent by default.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetClosedOrders request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetClosedOrdersResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetClosedOrders(ctx context.Context, nonce int64, opts *account.GetClosedOrdersOptions, secopts *common.SecurityOptions) (*account.GetClosedOrdersResponse, *http.Response, error)
	// # Description
	//
	// QueryOrdersInfo - Retrieve information about specific orders.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: QueryOrdersInfo request parameters.
	//	- opts: QueryOrdersInfo request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- QueryOrdersInfoResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	QueryOrdersInfo(ctx context.Context, nonce int64, params account.QueryOrdersInfoParameters, opts *account.QueryOrdersInfoRequestOptions, secopts *common.SecurityOptions) (*account.QueryOrdersInfoResponse, *http.Response, error)
	// # Description
	//
	// GetTradesHistory - Retrieve information about trades/fills. 50 results are returned at a time, the most recent by default.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetTradesHistory request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetTradesHistoryResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetTradesHistory(ctx context.Context, nonce int64, opts *account.GetTradesHistoryRequestOptions, secopts *common.SecurityOptions) (*account.GetTradesHistoryResponse, *http.Response, error)
	// # Description
	//
	// QueryTradesInfo - Retrieve information about specific trades/fills.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: QueryTradesInfo request parameters.
	//	- opts: QueryTradesInfo request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- QueryTradesInfoResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	QueryTradesInfo(ctx context.Context, nonce int64, params account.QueryTradesRequestParameters, opts *account.QueryTradesRequestOptions, secopts *common.SecurityOptions) (*account.QueryTradesInfoResponse, *http.Response, error)
	// # Description
	//
	// GetOpenPositions - Get information about open margin positions.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetOpenPositions request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetOpenPositionsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetOpenPositions(ctx context.Context, nonce int64, opts *account.GetOpenPositionsRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenPositionsRequestOptions, *http.Response, error)
	// # Description
	//
	// GetLedgersInfo - Retrieve information about ledger entries. 50 results are returned at a time, the most recent by default.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetLedgersInfo request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetLedgersInfoResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetLedgersInfo(ctx context.Context, nonce int64, opts *account.GetLedgersInfoRequestOptions, secopts *common.SecurityOptions) (*account.GetLedgersInfoResponse, *http.Response, error)
	// # Description
	//
	// QueryLedgers - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: QueryLedgers request parameters.
	//	- opts: QueryLedgers request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- QueryLedgersResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	QueryLedgers(ctx context.Context, nonce int64, params account.QueryLedgersRequestParameters, opts *account.QueryLedgersOptions, secopts *common.SecurityOptions) (*account.QueryLedgersResponse, *http.Response, error)
	// # Description
	//
	// GetTradeVolume - Returns 30 day USD trading volume and resulting fee schedule for any asset pair(s) provided.
	//
	// Note: If an asset pair is on a maker/taker fee schedule, the taker side is given in fees and maker side in
	// fees_maker. For pairs not on maker/taker, they will only be given in fees.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetTradeVolume request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetTradeVolumeResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetTradeVolume(ctx context.Context, nonce int64, opts *account.GetTradeVolumeRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeVolumeResponse, *http.Response, error)
	// # Description
	//
	// RequestExportReport - Request export of trades or ledgers.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: RequestExportReport request parameters.
	//	- opts: RequestExportReport request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- RequestExportReportResponse: The parsed response from Kraken API.
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
	// # Note on the http.Response.
	//
	// A reference to the received http.Response is always returned but it may be nil if no response was received.
	// Some endpoints of the Kraken API include tracing metadata in the response headers. The reference can be used
	// to extract the metadata (or any other kind of data that are not used by the API client directly).
	//
	// Please note response body will always be closed except for RetrieveDataExport.
	RequestExportReport(ctx context.Context, nonce int64, params account.RequestExportReportRequestParameters, opts *account.RequestExportReportRequestOptions, secopts *common.SecurityOptions) (*account.RequestExportReportResponse, *http.Response, error)
	// # Description
	//
	// GetExportReportStatus - Get status of requested data exports.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: GetExportReportStatus request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetExportReportStatusResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetExportReportStatus(ctx context.Context, nonce int64, params account.GetExportReportStatusRequestParameters, secopts *common.SecurityOptions) (*account.GetExportReportStatusResponse, *http.Response, error)
	// # Description
	//
	// RetrieveDataExport - Retrieve a processed data export.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: RetrieveDataExport request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- RetrieveDataExportResponse: The response contains an io.Reader that is tied to the http.Response body in order to let users download data in a streamed fashion.
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
	// Please note response body will not be closed in order to allow users to download the exported data in a streamed fashion.
	// The io.Reader in the response is tied to the http.Response.Body.
	RetrieveDataExport(ctx context.Context, nonce int64, params account.RetrieveDataExportParameters, secopts *common.SecurityOptions) (*account.RetrieveDataExportResponse, *http.Response, error)
	// # Description
	//
	// DeleteExportReport - Delete exported trades/ledgers report.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: DeleteExportReport request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- DeleteExportReportResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	DeleteExportReport(ctx context.Context, nonce int64, params account.DeleteExportReportRequestParameters, secopts *common.SecurityOptions) (*account.DeleteExportReportResponse, *http.Response, error)
	// # Description
	//
	// AddOrder - Place a new order.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: AddOrder request parameters.
	//	- opts: AddOrder request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- AddOrderResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	AddOrder(ctx context.Context, nonce int64, params trading.AddOrderRequestParameters, opts *trading.AddOrderRequestOptions, secopts *common.SecurityOptions) (*trading.AddOrderResponse, *http.Response, error)
	// # Description
	//
	// AddOrderBatch - Get the current system status or trading mode.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: AddOrderBatch request parameters.
	//	- opts: AddOrderBatch request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//	- AddOrderBatchResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	AddOrderBatch(ctx context.Context, nonce int64, params trading.AddOrderBatchRequestParameters, opts *trading.AddOrderBatchOptions, secopts *common.SecurityOptions) (*trading.AddOrderBatchResponse, *http.Response, error)
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
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: EditOrder request parameters.
	//	- opts: EditOrder request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- EditOrderResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	EditOrder(ctx context.Context, nonce int64, params trading.EditOrderRequestParameters, opts *trading.EditOrderRequestOptions, secopts *common.SecurityOptions) (*trading.EditOrderResponse, *http.Response, error)
	// # Description
	//
	// CancelOrder - Cancel a particular open order (or set of open orders) by txid or userref.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: CancelOrder request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- CancelOrderResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	CancelOrder(ctx context.Context, nonce int64, params trading.CancelOrderRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderResponse, *http.Response, error)
	// # Description
	//
	// CancelAllOrders - Cancel all open orders.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- CancelAllOrdersResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	CancelAllOrders(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*trading.CancelAllOrdersResponse, *http.Response, error)
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
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: CancelAllOrdersAfterX request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- CancelAllOrdersAfterXResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	CancelAllOrdersAfterX(ctx context.Context, nonce int64, params trading.CancelAllOrdersAfterXRequestParameters, secopts *common.SecurityOptions) (*trading.CancelAllOrdersAfterXResponse, *http.Response, error)
	// # Description
	//
	// CancelOrderBatch - Cancel multiple open orders by txid or userref (maximum 50 total unique IDs/references)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: CancelOrderBatch request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- CancelOrderBatchResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	CancelOrderBatch(ctx context.Context, nonce int64, params trading.CancelOrderBatchRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderBatchResponse, *http.Response, error)
	// # Description
	//
	// GetDepositMethods - Retrieve methods available for depositing a particular asset.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: CancelOrderBatch request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetDepositMethodsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetDepositMethods(ctx context.Context, nonce int64, params funding.GetDepositMethodsRequestParameters, secopts *common.SecurityOptions) (*funding.GetDepositMethodsResponse, *http.Response, error)
	// # Description
	//
	// GetDepositAddresses - Retrieve (or generate a new) deposit addresses for a particular asset and method.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: GetDepositAddresses request parameters.
	//	- opts: GetDepositAddresses request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetDepositAddressesResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetDepositAddresses(ctx context.Context, nonce int64, params funding.GetDepositAddressesRequestParameters, opts *funding.GetDepositAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetDepositAddressesResponse, *http.Response, error)
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
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetStatusOfRecentDeposits request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetStatusOfRecentDepositsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetStatusOfRecentDeposits(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentDepositsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentDepositsResponse, *http.Response, error)
	// # Description
	//
	// GetWithdrawalMethods - Retrieve a list of withdrawal addresses available for the user.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetWithdrawalMethods request options.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetWithdrawalMethodsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetWithdrawalMethods(ctx context.Context, nonce int64, opts *funding.GetWithdrawalMethodsRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalMethodsResponse, *http.Response, error)
	// # Description
	//
	// GetWithdrawalAddresses - Retrieve a list of withdrawal addresses available for the user.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetWithdrawalAddresses request options.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetWithdrawalAddressesResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetWithdrawalAddresses(ctx context.Context, nonce int64, opts *funding.GetWithdrawalAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalAddressesResponse, *http.Response, error)
	// # Description
	//
	// GetWithdrawalInformation - Retrieve a list of withdrawal methods available for the user.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: GetWithdrawalInformation request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetWithdrawalInformationResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetWithdrawalInformation(ctx context.Context, nonce int64, params funding.GetWithdrawalInformationRequestParameters, secopts *common.SecurityOptions) (*funding.GetWithdrawalInformationResponse, *http.Response, error)
	// # Description
	//
	// WithdrawFunds - Make a withdrawal request.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: WithdrawFunds request parameters.
	//	- opts: WithdrawFunds request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- WithdrawFundsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	WithdrawFunds(ctx context.Context, nonce int64, params funding.WithdrawFundsRequestParameters, opts *funding.WithdrawFundsRequestOptions, secopts *common.SecurityOptions) (*funding.WithdrawFundsResponse, *http.Response, error)
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
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: GetStatusOfRecentWithdrawals request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetStatusOfRecentWithdrawalsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetStatusOfRecentWithdrawals(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentWithdrawalsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentWithdrawalsResponse, *http.Response, error)
	// # Description
	//
	// RequestWithdrawalCancellation - Cancel a recently requested withdrawal, if it has not already been successfully processed.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: RequestWithdrawalCancellation request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- RequestWithdrawalCancellationResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	RequestWithdrawalCancellation(ctx context.Context, nonce int64, params funding.RequestWithdrawalCancellationRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWithdrawalCancellationResponse, *http.Response, error)
	// # Description
	//
	// RequestWalletTransfer - Transfer from a Kraken spot wallet to a Kraken Futures wallet. Note that a
	// transfer in the other direction must be requested via the Kraken Futures API endpoint for
	// withdrawals to Spot wallets.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: RequestWalletTransfer request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- RequestWalletTransferResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	RequestWalletTransfer(ctx context.Context, nonce int64, params funding.RequestWalletTransferRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWalletTransferResponse, *http.Response, error)
	// # Description
	//
	// AllocateEarnFunds - Allocate funds to the Strategy.
	//
	// # Usage tips
	//
	// This method is asynchronous. A couple of preflight checks are performed synchronously on
	// behalf of the method before it is dispatched further. The client is required to poll the
	// result using GetAllocationStatus.
	//
	// There can be only one (de)allocation request in progress for given user and strategy at any
	// time. While the operation is in progress:
	//
	//	- pending attribute in /Earn/Allocations response for the strategy indicates that funds are being allocated.
	//	- pending attribute in /Earn/AllocateStatus response will be true.
	//
	// Following specific errors within Earnings class can be returned by this method:
	//
	//	- Minimum allocation: EEarnings:Below min:(De)allocation operation amount less than minimum
	//	- Allocation in progress: EEarnings:Busy:Another (de)allocation for the same strategy is in progress
	//	- Service temporarily unavailable: EEarnings:Busy. Try again in a few minutes.
	//	- User tier verification: EEarnings:Permission denied:The user's tier is not high enough
	//	- Strategy not found: EGeneral:Invalid arguments:Invalid strategy ID
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: AllocateEarnFunds request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- AllocateEarnFundsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	AllocateEarnFunds(ctx context.Context, nonce int64, params earn.AllocateFundsRequestParameters, secopts *common.SecurityOptions) (*earn.AllocateFundsResponse, *http.Response, error)
	// # Description
	//
	// DeallocateEarnFunds - Deallocate funds to the Strategy.
	//
	// # Usage tips
	//
	// This method is asynchronous. A couple of preflight checks are performed synchronously on
	// behalf of the method before it is dispatched further. The client is required to poll the
	// result using GetDeallocationStatus.
	//
	// There can be only one (de)allocation request in progress for given user and strategy at any
	// time. While the operation is in progress:
	//
	//	- pending attribute in Allocations response for the strategy will hold the amount that is being deallocated (negative amount).
	//	- pending attribute in DeallocateStatus response will be true.
	//
	// Following specific errors within Earnings class can be returned by this method:
	//
	//	- Minimum allocation: EEarnings:Below min:(De)allocation operation amount less than minimum
	//	- Allocation in progress: EEarnings:Busy:Another (de)allocation for the same strategy is in progress
	//	- Service temporarily unavailable: EEarnings:Busy. Try again in a few minutes.
	//	- Strategy not found: EGeneral:Invalid arguments:Invalid strategy ID
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: DeallocateEarnFunds request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- DeallocateEarnFundsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	DeallocateEarnFunds(ctx context.Context, nonce int64, params earn.DeallocateFundsRequestParameters, secopts *common.SecurityOptions) (*earn.DeallocateFundsResponse, *http.Response, error)
	// # Description
	//
	// GetAllocationStatus - Get the status of the last allocation request.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: GetAllocationStatus request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetAllocationStatusResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetAllocationStatus(ctx context.Context, nonce int64, params earn.GetAllocationStatusRequestParameters, secopts *common.SecurityOptions) (*earn.GetAllocationStatusResponse, *http.Response, error)
	// # Description
	//
	// GetDeallocationStatus - Get the status of the last deallocation request.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- params: GetDeallocationStatus request parameters.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetDeallocationStatusResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetDeallocationStatus(ctx context.Context, nonce int64, params earn.GetDeallocationStatusRequestParameters, secopts *common.SecurityOptions) (*earn.GetDeallocationStatusResponse, *http.Response, error)
	// # Description
	//
	// ListEarnStrategies - List earn strategies along with their parameters.
	//
	// # Usage tips
	//
	// Returns only strategies that are available to the user based on geographic region.
	//
	// When the user does not meet the tier restriction, can_allocate will be false and
	// allocation_restriction_info indicates Tier as the restriction reason. Earn products
	// generally require Intermediate tier. Get your account verified to access earn.
	//
	// Paging isn't yet implemented, so it the endpoint always returns all data in the first page.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: ListEarnStrategies request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- ListEarnStrategiesResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	ListEarnStrategies(ctx context.Context, nonce int64, opts *earn.ListEarnStrategiesRequestOptions, secopts *common.SecurityOptions) (*earn.ListEarnStrategiesResponse, *http.Response, error)
	// # Description
	//
	// ListEarnAllocations - List all allocations for the user.
	//
	// # Usage tips
	//
	// By default all allocations are returned, even for strategies that have been used in the past
	// and have zero balance now. That is so that the user can see how much was earned with given
	// strategy in the past. hide_zero_allocations parameter can be used to remove zero balance
	// entries from the output. Paging hasn't been implemented for this method as we don't expect
	// the result for a particular user to be overwhelmingly large.
	//
	// All amounts in the output can be denominated in a currency of user's choice (the converted_asset parameter).
	//
	// Information about when the next reward will be paid to the client is also provided in the output.
	//
	// Allocated funds can be in up to 4 states:
	//
	//	- bonding
	//	- allocated
	//	- exit_queue (ETH only)
	//	- unbonding
	//
	// Any funds in total not in bonding/unbonding are simply allocated and earning rewards.
	// Depending on the strategy funds in the other 3 states can also be earning rewards. Consult
	// the output of /Earn/Strategies to know whether bonding/unbonding earn rewards. ETH in
	// exit_queue still earns rewards.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- opts: ListEarnAllocations request options. A nil value triggers all default behaviors.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// Note that for ETH, when the funds are in the exit_queue state, the expires time given is the
	// time when the funds will have finished unbonding, not when they go from exit queue to unbonding.
	//
	// (Un)bonding time estimate can be inaccurate right after having (de)allocated the funds. Wait
	// 1-2 minutes after (de)allocating to get an accurate result.
	//
	// # Returns
	//
	//	- ListEarnAllocationsResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	ListEarnAllocations(ctx context.Context, nonce int64, opts *earn.ListEarnAllocationsRequestOptions, secopts *common.SecurityOptions) (*earn.ListEarnAllocationsResponse, *http.Response, error)
	// # Description
	//
	// GetWebsocketToken - An authentication token must be requested via this REST API endpoint in
	// order to connect to and authenticate with our Websockets API. The token should be used
	// within 15 minutes of creation, but it does not expire once a successful Websockets
	// connection and private subscription has been made and is maintained.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- nonce: Nonce used to sign request.
	//	- secopts: Security options to use for the API call (2FA, ...)
	//
	// # Returns
	//
	//	- GetWebsocketTokenResponse: The parsed response from Kraken API.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	GetWebsocketToken(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*websocket.GetWebsocketTokenResponse, *http.Response, error)
}
