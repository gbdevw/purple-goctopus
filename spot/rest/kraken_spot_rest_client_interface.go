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
	GetAssetInfo(ctx context.Context, opts *market.GetAssetInfoRequestOptions) (*market.GetAssetInfoResponse, error)
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
	// GetSystemStatus - Get the current system status or trading mode.
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
	// GetTradeVolume - Returns 30 day USD trading volume and resulting fee schedule for any asset pair(s) provided. Note: If an asset pair is on a maker/taker fee schedule, the taker side is given in fees and maker side in fees_maker. For pairs not on maker/taker, they will only be given in fees.
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
	// RESUME HERE
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
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
	// Please note response body will always be closed except for RetrieveDataExport.
	RequestWalletTransfer(ctx context.Context, params RequestWalletTransferParameters, secopts *SecurityOptions) (*RequestWalletTransferResponse, error)
}
