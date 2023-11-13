package spotex

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*****************************************************************************/
/*																			 */
/*	PRIVATE ENUMS															 */
/*																			 */
/*****************************************************************************/

// Public endpoint names
const (
	// Market Data endpoints - https://docs.kraken.com/rest/#tag/Market-Data
	getServerTime         = "Time"
	getSystemStatus       = "SystemStatus"
	getAssetInfo          = "Assets"
	getTradableAssetPairs = "AssetPairs"
	getTickerInformation  = "Ticker"
	getOHLCData           = "OHLC"
	getOrderBook          = "Depth"
	getRecentTrades       = "Trades"
	getRecentSpreads      = "Spread"
)

// Private endpoint names
const (
	// User Data - https://docs.kraken.com/rest/#tag/User-Data
	postGetAccountBalance     = "Balance"
	postGetTradeBalance       = "TradeBalance"
	postGetOpenOrders         = "OpenOrders"
	postGetClosedOrders       = "ClosedOrders"
	postQueryOrdersInfos      = "QueryOrders"
	postGetTradesHistory      = "TradesHistory"
	postQueryTradesInfo       = "QueryTrades"
	postGetOpenPositions      = "OpenPositions"
	postGetLedgersInfo        = "Ledgers"
	postQueryLedgers          = "QueryLedgers"
	postGetTradeVolume        = "TradeVolume"
	postRequestExportReport   = "AddExport"
	postGetExportReportStatus = "ExportStatus"
	postRetrieveDataExport    = "RetrieveExport"
	postDeleteExportReport    = "RemoveExport"

	// User trading - https://docs.kraken.com/rest/#tag/User-Trading
	postAddOrder              = "AddOrder"
	postAddOrderBatch         = "AddOrderBatch"
	postEditOrder             = "EditOrder"
	postCancelOrder           = "CancelOrder"
	postCancelAllOrders       = "CancelAll"
	postCancelAllOrdersAfterX = "CancelAllOrdersAfter"
	postCancelOrderBatch      = "CancelOrderBatch"

	// User Funding - https://docs.kraken.com/rest/#tag/User-Funding
	postGetDepositMethods             = "DepositMethods"
	postGetDepositAddresses           = "DepositAddresses"
	postGetStatusOfRecentDeposits     = "DepositStatus"
	postGetWithdrawalInformation      = "WithdrawInfo"
	postWithdrawFunds                 = "Withdraw"
	postGetStatusOfRecentWithdrawals  = "WithdrawStatus"
	postRequestWithdrawalCancellation = "WithdrawCancel"
	postRequestWalletTransfer         = "WalletTransfer"

	// User staking - https://docs.kraken.com/rest/#tag/User-Staking
	postStakeAsset                    = "Stake"
	postUnstakeAsset                  = "Unstake"
	postListOfStakeableAssets         = "Staking/Assets"
	postGetPendingStakingTransactions = "Staking/Pending"
	postListOfStakingTransactions     = "Staking/Transactions"
)

// Headers managed by KrakenAPIClient
const (
	managedHeaderContentType = "Content-Type"
	managedHeaderUserAgent   = "User-Agent"
	managedHeaderAPIKey      = "API-Key"
	managedHeaderAPISign     = "API-Sign"
)

/*****************************************************************************/
/*																			 */
/*	KRAKEN API CLIENT FACTORIES												 */
/*																			 */
/*****************************************************************************/

// Interface which defines a method to get a nonce to sign a request
type NonceGenerator interface {
	GetNextNonce() int64
}

// Built-in default generator which uses unix timestamp as nonce
type UnixtimestampBasedNonceGenerator struct {
	// Mutex to prevent concurrency when generating the nonce
	mu sync.Mutex
}

// Generate a nonce using a unix nansec timestamp
func (gen *UnixtimestampBasedNonceGenerator) GetNextNonce() int64 {
	// Lock mutex & defer unlock
	gen.mu.Lock()
	defer gen.mu.Unlock()
	// Generate and return nonce
	return time.Now().UnixNano()
}

// KrakenAPIClient represents a Kraken API Client connection
type KrakenAPIClient struct {
	// Base URL for Kraken Rest API
	baseURL string
	// Version number of the API
	version string
	// Value for the mandatory User-Agent header
	agent string
	// API Key used to sign request
	key string
	// Secret used to forge signature
	secret []byte
	// HTTP client used to perform API calls
	client *http.Client
	// Nonce generator to use to get nonces used to sign requests
	nonceGenerator NonceGenerator
	// Kraken client interface
	KrakenAPIClientIface
}

// Configurable options for Kraken REST API client
type KrakenAPIClientOptions struct {
	// Base URL for the API. Default to "https://api.kraken.com"
	BaseURL string
	// Version of the API. Default to "0"
	Version string
	// Value for the mandatory User-Agent. Default to "Kragoc"
	Agent string
	// Low level HTTP client to use to perform API calls.
	// Default to the default client provided by http package
	Client *http.Client
	// Nonce generator to use
	// By default, a nanosec unix timestamp is provided
	NonceGenerator NonceGenerator
}

// Create a fully initiated KrakenAPIOptions struct with nice defaults
//
// return: KrakenAPIOptions struct with nice defaults
func DefaultOptions() *KrakenAPIClientOptions {
	return &KrakenAPIClientOptions{
		BaseURL:        "https://api.kraken.com",
		Version:        "0",
		Agent:          "Kragoc",
		Client:         http.DefaultClient,
		NonceGenerator: &UnixtimestampBasedNonceGenerator{mu: sync.Mutex{}},
	}
}

// Create a new KrakenAPI client with default options and no credentials.
//
// Only public endpoints can be used with such client. Using private endpoints
// will lead to errors because of invalid request signatures.
func NewPublicWithDefaultOptions() *KrakenAPIClient {
	return NewWithCredentialsAndOptions("", nil, DefaultOptions())
}

// Create a new KrakenAPI client with specific options and no credentials.
//
// Only public endpoints can be used with such client. Using private endpoints
// will lead to errors because of invalid request signatures.
func NewPublicWithOptions(options *KrakenAPIClientOptions) *KrakenAPIClient {
	return NewWithCredentialsAndOptions("", nil, options)
}

// Create a new Kraken API client with default options and credentials
// to use private endpoints.
func NewWithCredentialsAndDefaultOptions(key string, secret []byte) *KrakenAPIClient {
	return NewWithCredentialsAndOptions(key, secret, DefaultOptions())
}

// Create a new KrakenAPI client with specific options and credentials
// to use private endpoints.
func NewWithCredentialsAndOptions(key string, secret []byte, options *KrakenAPIClientOptions) *KrakenAPIClient {

	// Override default options with defined options provided as input
	defopts := DefaultOptions()
	if options != nil {
		if options.BaseURL != "" {
			defopts.BaseURL = options.BaseURL
		}
		if options.Version != "" {
			defopts.Version = options.Version
		}
		if options.Agent != "" {
			defopts.Agent = options.Agent
		}
		if options.Client != nil {
			defopts.Client = options.Client
		}
		if options.NonceGenerator != nil {
			defopts.NonceGenerator = options.NonceGenerator
		}
	}

	// Build client
	return &KrakenAPIClient{
		key:            key,
		secret:         secret,
		agent:          defopts.Agent,
		baseURL:        defopts.BaseURL,
		version:        defopts.Version,
		client:         defopts.Client,
		nonceGenerator: defopts.NonceGenerator,
	}
}

/*****************************************************************************/
/*                                                                           */
/*	PUBLIC ENDPOINTS - MARKET DATA                                           */
/*                                                                           */
/*****************************************************************************/

// GetServerTime Get the server's time.
func (api *KrakenAPIClient) GetServerTime() (*GetServerTimeResponse, error) {
	resp, err := api.queryPublic(getServerTime, http.MethodGet, nil, "", nil, nil, &GetServerTimeResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetServerTimeResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast response to GetServerTimeResponse. Got %T : %#v", resp, resp)
	}
	return result, nil
}

// GetSystemStatus Get the status of the system.
func (api *KrakenAPIClient) GetSystemStatus() (*GetSystemStatusResponse, error) {

	// Call API endpoint
	resp, err := api.queryPublic(getSystemStatus, http.MethodGet, nil, "", nil, nil, &GetSystemStatusResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	result, ok := resp.(*GetSystemStatusResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast response to GetSystemStatusResponse. Got %T : %#v", resp, resp)
	}
	return result, nil
}

// GetAssetInfo Return the servers available assets
func (api *KrakenAPIClient) GetAssetInfo(options *GetAssetInfoOptions) (*GetAssetInfoResponse, error) {

	// Prepare query string params.
	query := url.Values{}
	if options != nil {
		if len(options.Assets) > 0 {
			query.Add("asset", strings.Join(options.Assets, ","))
		}
		if options.AssetClass != "" {
			query.Add("aclass", string(options.AssetClass))
		}
	}

	// Perform request
	resp, err := api.queryPublic(getAssetInfo, http.MethodGet, query, "", nil, nil, &GetAssetInfoResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	data, ok := resp.(*GetAssetInfoResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to GetAssetInfoResponse. Got %T : %#v", resp, resp)
	}

	// Return response
	return data, nil
}

// GetTradableAssetPairs Return the servers available asset pairs.
func (api *KrakenAPIClient) GetTradableAssetPairs(options *GetTradableAssetPairsOptions) (*GetTradableAssetPairsResponse, error) {

	// Prepare query string params.
	query := url.Values{}
	if options != nil {
		if len(options.Pairs) > 0 {
			query.Add("pair", strings.Join(options.Pairs, ","))
		}
		if options.Info != "" {
			query.Add("info", options.Info)
		}
	}

	// Perform request
	resp, err := api.queryPublic(getTradableAssetPairs, http.MethodGet, query, "", nil, nil, &GetTradableAssetPairsResponse{})
	if err != nil {
		return nil, err
	}

	// Type assertion
	data, ok := resp.(*GetTradableAssetPairsResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response to GetTradableAssetPairsResponse. Got %T : %#v", resp, resp)
	}

	// Return response
	return data, nil
}

// GetTickerInformation Get ticker information about a given list of pairs.
// Note: Today's prices start at midnight UTC.
func (api *KrakenAPIClient) GetTickerInformation(opts *GetTickerInformationOptions) (*GetTickerInformationResponse, error) {

	// Prepare query string params.
	query := url.Values{}
	if opts != nil && len(opts.Pairs) > 0 {
		query.Add("pair", strings.Join(opts.Pairs, ","))
	}

	// Perform request
	resp, err := api.queryPublic(getTickerInformation, http.MethodGet, query, "", nil, nil, &GetTickerInformationResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	data, ok := resp.(*GetTickerInformationResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast response to GetTickerInformationResponse. Got %T", resp)
	}

	// Return result
	return data, nil
}

// GetOHLCData get Open, High, Low & Close indicators.
// Note: the last entry in the OHLC array is for the current, not-yet-committed
// frame and will always be present, regardless of the value of since.
func (api *KrakenAPIClient) GetOHLCData(params GetOHLCDataParameters, options *GetOHLCDataOptions) (*GetOHLCDataResponse, error) {

	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if options != nil {
		if options.Interval != 0 {
			query.Add("interval", strconv.FormatInt(int64(options.Interval), 10))
		}
		if options.Since != nil {
			query.Add("since", strconv.FormatInt(options.Since.Unix(), 10))
		}
	}

	// Perform request
	resp, err := api.queryPublic(getOHLCData, http.MethodGet, query, "", nil, nil, &KrakenAPIResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	r, ok := resp.(*KrakenAPIResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response. Got %T", resp)
	}

	// Process result if any
	if len(r.Error) > 0 {
		return &GetOHLCDataResponse{
			KrakenAPIResponse: KrakenAPIResponse{Error: r.Error},
			Result:            GetOHLCDataResult{},
		}, nil
	} else {
		// Result type assertion
		data, ok := r.Result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast server response result. Got %T", r.Result)
		}

		result := &GetOHLCDataResult{
			Last: 0,
			OHLC: map[string][]OHLCData{},
		}

		// Parse last field in server response
		for key, field := range data {
			// Handle last field
			if key == "last" {
				last, ok := field.(float64)
				if !ok {
					return nil, fmt.Errorf("could not parse 'last' field from response. Got %v", data)
				}
				result.Last = int64(last)
			} else {
				// Handle array of OHLC data
				candles, ok := field.([]interface{})
				if !ok {
					return nil, fmt.Errorf("could not parse candles related to %s from response. Got %T : %v", params.Pair, field, field)
				}
				// Parse data from each candle
				for _, candle := range candles {
					candle, ok := candle.([]interface{})
					if !ok {
						return nil, fmt.Errorf("could not parse candle related to %s from response. Got %T : %v", params.Pair, candle, candle)
					}
					timestamp, okt := candle[0].(float64)
					open, oko := candle[1].(string)
					high, okh := candle[2].(string)
					low, okl := candle[3].(string)
					close, okc := candle[4].(string)
					avg, oka := candle[5].(string)
					volume, okv := candle[6].(string)
					count, okn := candle[7].(float64)
					if !(okt && oko && okh && okl && okc && oka && okv && okn) {
						return nil, fmt.Errorf("could not parse data from candle related to %s. Got %T : %v", params.Pair, candle, candle)
					}
					// Add candle to response
					ohlc := OHLCData{
						Timestamp: time.Unix(int64(timestamp), 0).UTC(),
						Open:      open,
						High:      high,
						Low:       low,
						Close:     close,
						Avg:       avg,
						Volume:    volume,
						Count:     int(count),
					}
					result.OHLC[key] = append(result.OHLC[key], ohlc)
				}
			}
		}

		// Return response
		return &GetOHLCDataResponse{
			KrakenAPIResponse{Error: r.Error},
			*result,
		}, nil
	}
}

// GetOrderBook Get order by for a given pair.
func (api *KrakenAPIClient) GetOrderBook(params GetOrderBookParameters, options *GetOrderBookOptions) (*GetOrderBookResponse, error) {

	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if options != nil {
		if options.Count != 0 {
			query.Add("count", strconv.FormatInt(int64(options.Count), 10))
		}
	}

	// Perform request
	resp, err := api.queryPublic(getOrderBook, http.MethodGet, query, "", nil, nil, &KrakenAPIResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	r, ok := resp.(*KrakenAPIResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response. Got %T : %v", resp, resp)
	}

	// Process result if any
	if len(r.Error) > 0 {
		return &GetOrderBookResponse{
			KrakenAPIResponse: KrakenAPIResponse{Error: r.Error},
			Result:            map[string]OrderBook{},
		}, nil
	} else {
		// Cast result
		books, ok := r.Result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast server result. Got %T : %v", r.Result, r.Result)
		}

		// Prepare result
		result := map[string]OrderBook{}

		// Parse books
		for key, book := range books {

			book, ok := book.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("could not cast to order book. Got %T : %v", book, book)
			}

			// Prepare result data
			asks := []OrderBookEntry{}
			bids := []OrderBookEntry{}

			// Parse entries for each side
			for side, entries := range book {

				entries, ok := entries.([]interface{})
				if !ok {
					return nil, fmt.Errorf("could not cast order book entries. Got %T : %v", entries, entries)
				}

				// Parse each entry
				for _, entry := range entries {

					entry, ok := entry.([]interface{})
					if !ok {
						return nil, fmt.Errorf("could not cast order book entry. Got %T : %v", entry, entry)
					}

					price, okp := entry[0].(string)
					volume, okv := entry[1].(string)
					timestamp, okt := entry[2].(float64)
					if !(okp && okv && okt) {
						return nil, fmt.Errorf("could not cast data from order book entry. Got %T : %v", entry, entry)
					}
					elem := OrderBookEntry{
						Timestamp: time.Unix(int64(timestamp), 0).UTC(),
						Price:     price,
						Volume:    volume,
					}

					// Add entry to correct side of the order book
					switch side {
					case "asks":
						asks = append(asks, elem)
					case "bids":
						bids = append(bids, elem)
					default:
						return nil, fmt.Errorf("invalid field found in order book entries. Got %s", side)
					}
				}

				// Add parsed book to response
				result[key] = OrderBook{
					Asks: asks,
					Bids: bids,
				}
			}
		}

		// Return response
		return &GetOrderBookResponse{
			KrakenAPIResponse{Error: r.Error},
			result,
		}, nil
	}
}

// GetRecentTrades Get up to the 1000 most recent trades by default.
func (api *KrakenAPIClient) GetRecentTrades(params GetRecentTradesParameters, options *GetRecentTradesOptions) (*GetRecentTradesResponse, error) {

	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if options != nil {
		if options.Since != nil {
			query.Add("since", strconv.FormatInt(int64(options.Since.Unix()), 10))
		}
	}

	resp, err := api.queryPublic(getRecentTrades, http.MethodGet, query, "", nil, nil, &KrakenAPIResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	r, ok := resp.(*KrakenAPIResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response. Got %T : %v", resp, resp)
	}

	// Process result if any
	if len(r.Error) > 0 {
		return &GetRecentTradesResponse{
			KrakenAPIResponse: KrakenAPIResponse{Error: r.Error},
			Result:            GetRecentTradesResult{},
		}, nil
	} else {
		// Prepare result
		result := &GetRecentTradesResult{
			Last:   "0",
			Trades: map[string][]Trade{},
		}

		// Cast response result
		data, ok := r.Result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast server response result. Got %T : %v", r.Result, r.Result)
		}

		// Parse fields from map
		for key, field := range data {
			if key == "last" {
				// Parse last from response
				lastStr, ok := field.(string)
				if !ok {
					return nil, fmt.Errorf("could not cast 'last' field from response. Got %T : %v", data, data)
				}
				result.Last = lastStr
			} else {
				// Cast trades for a given pair
				trades, ok := field.([]interface{})
				if !ok {
					return nil, fmt.Errorf("could not cast trades from response. Got %T : %v", field, field)
				}
				// Unwrap trades
				for _, trade := range trades {

					trade, ok := trade.([]interface{})
					if !ok {
						return nil, fmt.Errorf("could not cast trade from trades. Got %T : %v", trade, trade)
					}
					price, okp := trade[0].(string)
					volume, okv := trade[1].(string)
					timestampF, oktm := trade[2].(float64)
					side, oks := trade[3].(string)
					typ, okt := trade[4].(string)
					misc, okm := trade[5].(string)
					id, oki := trade[6].(float64)
					if !(okp && okv && oktm && oks && okt && okm && oki) {
						return nil, fmt.Errorf("could not cast fields from trade. Got %T : %v", trade, trade)
					}

					// Add trade to response
					result.Trades[key] = append(result.Trades[key], Trade{
						// Use millisec timestamp
						Timestamp:     int64(timestampF * 1000),
						Price:         price,
						Volume:        volume,
						Side:          side,
						Type:          typ,
						Miscellaneous: misc,
						Id:            int64(id),
					})
				}
			}
		}

		// Return response
		return &GetRecentTradesResponse{
			KrakenAPIResponse: KrakenAPIResponse{Error: r.Error},
			Result:            *result,
		}, nil
	}
}

// GetRecentTrades Get recent spreads
func (api *KrakenAPIClient) GetRecentSpreads(params GetRecentSpreadsParameters, options *GetRecentSpreadsOptions) (*GetRecentSpreadsResponse, error) {

	// Prepare query string params.
	query := url.Values{}
	query.Add("pair", params.Pair)
	if options != nil {
		if options.Since != nil {
			query.Add("since", strconv.FormatInt(int64(options.Since.Unix()), 10))
		}
	}

	// Perform request
	resp, err := api.queryPublic(getRecentSpreads, http.MethodGet, query, "", nil, nil, &KrakenAPIResponse{})
	if err != nil {
		return nil, err
	}

	// Cast response
	r, ok := resp.(*KrakenAPIResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast server response. Got%T : %#v", r, r)
	}

	// Process results if any
	if len(r.Error) > 0 {
		return &GetRecentSpreadsResponse{
			KrakenAPIResponse: KrakenAPIResponse{Error: r.Error},
			Result:            GetRecentSpreadsResult{},
		}, nil
	} else {
		// Prepare result
		result := GetRecentSpreadsResult{
			Last:    "0",
			Spreads: map[string][]SpreadData{},
		}
		// Cast response result
		data, ok := r.Result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast server response result. Got %T : %v", r.Result, r.Result)
		}

		// Parse fields from response
		for key, field := range data {
			if key == "last" {
				// Parse last from response
				last, ok := data["last"].(float64)
				if !ok {
					return nil, fmt.Errorf("could not cast 'last' field from response. Got %T : %v", field, field)
				}
				result.Last = fmt.Sprintf("%d", int64(last))
			} else {
				// Parse spreads from response
				spreads, ok := field.([]interface{})
				if !ok {
					return nil, fmt.Errorf("could not cast spreads from response. Got %T : %v", field, field)
				}

				// Unwrap spreads
				for _, spread := range spreads {
					spread, ok := spread.([]interface{})
					if !ok {
						return nil, fmt.Errorf("could not cast spread. Got %T : %v", spread, spread)
					}
					timestamp, oktm := spread[0].(float64)
					bask, oka := spread[1].(string)
					bbid, okb := spread[2].(string)
					if !(oktm && oka && okb) {
						return nil, fmt.Errorf("could not cast data from spread. Got %T : %v", spread, spread)
					}

					// Add spread to response
					result.Spreads[key] = append(result.Spreads[key], SpreadData{
						Timestamp: time.Unix(int64(timestamp), 0).UTC(),
						BestBid:   bbid,
						BestAsk:   bask,
					})
				}
			}
		}

		// Return response
		return &GetRecentSpreadsResponse{
			KrakenAPIResponse: KrakenAPIResponse{Error: r.Error},
			Result:            result,
		}, nil
	}
}

/*****************************************************************************/
/*																			 */
/*	PRIVATE ENDPOINTS - USER DATA											 */
/*																			 */
/*****************************************************************************/

// GetAccountBalance - Retrieve all cash balances, net of pending withdrawals.
func (api *KrakenAPIClient) GetAccountBalance(secopts *SecurityOptions) (*GetAccountBalanceResponse, error) {

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
func (api *KrakenAPIClient) GetTradeBalance(options *GetTradeBalanceOptions, secopts *SecurityOptions) (*GetTradeBalanceResponse, error) {

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
func (api *KrakenAPIClient) GetOpenOrders(options *GetOpenOrdersOptions, secopts *SecurityOptions) (*GetOpenOrdersResponse, error) {

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
func (api *KrakenAPIClient) GetClosedOrders(options *GetClosedOrdersOptions, secopts *SecurityOptions) (*GetClosedOrdersResponse, error) {

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
func (api *KrakenAPIClient) QueryOrdersInfo(params QueryOrdersParameters, options *QueryOrdersOptions, secopts *SecurityOptions) (*QueryOrdersInfoResponse, error) {

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
func (api *KrakenAPIClient) GetTradesHistory(options *GetTradesHistoryOptions, secopts *SecurityOptions) (*GetTradesHistoryResponse, error) {

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
func (api *KrakenAPIClient) QueryTradesInfo(params QueryTradesParameters, options *QueryTradesOptions, secopts *SecurityOptions) (*QueryTradesInfoResponse, error) {

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
func (api *KrakenAPIClient) GetOpenPositions(options *GetOpenPositionsOptions, secopts *SecurityOptions) (*GetOpenPositionsResponse, error) {

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
func (api *KrakenAPIClient) GetLedgersInfo(options *GetLedgersInfoOptions, secopts *SecurityOptions) (*GetLedgersInfoResponse, error) {

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
func (api *KrakenAPIClient) QueryLedgers(params QueryLedgersParameters, options *QueryLedgersOptions, secopts *SecurityOptions) (*QueryLedgersResponse, error) {

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
func (api *KrakenAPIClient) GetTradeVolume(params GetTradeVolumeParameters, options *GetTradeVolumeOptions, secopts *SecurityOptions) (*GetTradeVolumeResponse, error) {

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
func (api *KrakenAPIClient) RequestExportReport(params RequestExportReportParameters, options *RequestExportReportOptions, secopts *SecurityOptions) (*RequestExportReportResponse, error) {

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
func (api *KrakenAPIClient) GetExportReportStatus(params GetExportReportStatusParameters, secopts *SecurityOptions) (*GetExportReportStatusResponse, error) {

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
func (api *KrakenAPIClient) RetrieveDataExport(params RetrieveDataExportParameters, secopts *SecurityOptions) (*RetrieveDataExportResponse, error) {

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
func (api *KrakenAPIClient) DeleteExportReport(params DeleteExportReportParameters, secopts *SecurityOptions) (*DeleteExportReportResponse, error) {

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
func (api *KrakenAPIClient) AddOrder(params AddOrderParameters, options *AddOrderOptions, secopts *SecurityOptions) (*AddOrderResponse, error) {

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
func (api *KrakenAPIClient) AddOrderBatch(params AddOrderBatchParameters, options *AddOrderBatchOptions, secopts *SecurityOptions) (*AddOrderBatchResponse, error) {

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
func (api *KrakenAPIClient) EditOrder(params EditOrderParameters, options *EditOrderOptions, secopts *SecurityOptions) (*EditOrderResponse, error) {

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
func (api *KrakenAPIClient) CancelOrder(params CancelOrderParameters, secopts *SecurityOptions) (*CancelOrderResponse, error) {

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
func (api *KrakenAPIClient) CancelAllOrders(secopts *SecurityOptions) (*CancelAllOrdersResponse, error) {

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
func (api *KrakenAPIClient) CancelAllOrdersAfterX(params CancelCancelAllOrdersAfterXParameters, secopts *SecurityOptions) (*CancelAllOrdersAfterXResponse, error) {

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
func (api *KrakenAPIClient) CancelOrderBatch(params CancelOrderBatchParameters, secopts *SecurityOptions) (*CancelOrderBatchResponse, error) {

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
func (api *KrakenAPIClient) GetDepositMethods(params GetDepositMethodsParameters, secopts *SecurityOptions) (*GetDepositMethodsResponse, error) {

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
func (api *KrakenAPIClient) GetDepositAddresses(params GetDepositAddressesParameters, options *GetDepositAddressesOptions, secopts *SecurityOptions) (*GetDepositAddressesResponse, error) {

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
func (api *KrakenAPIClient) GetStatusOfRecentDeposits(params GetStatusOfRecentDepositsParameters, options *GetStatusOfRecentDepositsOptions, secopts *SecurityOptions) (*GetStatusOfRecentDepositsResponse, error) {

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
func (api *KrakenAPIClient) GetWithdrawalInformation(params GetWithdrawalInformationParameters, secopts *SecurityOptions) (*GetWithdrawalInformationResponse, error) {

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
func (api *KrakenAPIClient) WithdrawFunds(params WithdrawFundsParameters, secopts *SecurityOptions) (*WithdrawFundsResponse, error) {

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
func (api *KrakenAPIClient) GetStatusOfRecentWithdrawals(params GetStatusOfRecentWithdrawalsParameters, options *GetStatusOfRecentWithdrawalsOptions, secopts *SecurityOptions) (*GetStatusOfRecentWithdrawalsResponse, error) {

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
func (api *KrakenAPIClient) RequestWithdrawalCancellation(params RequestWithdrawalCancellationParameters, secopts *SecurityOptions) (*RequestWithdrawalCancellationResponse, error) {

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
func (api *KrakenAPIClient) RequestWalletTransfer(params RequestWalletTransferParameters, secopts *SecurityOptions) (*RequestWalletTransferResponse, error) {

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
func (api *KrakenAPIClient) StakeAsset(params StakeAssetParameters, secopts *SecurityOptions) (*StakeAssetResponse, error) {

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
func (api *KrakenAPIClient) UnstakeAsset(params UnstakeAssetParameters, secopts *SecurityOptions) (*UnstakeAssetResponse, error) {

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
func (api *KrakenAPIClient) ListOfStakeableAssets(secopts *SecurityOptions) (*ListOfStakeableAssetsResponse, error) {

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
func (api *KrakenAPIClient) GetPendingStakingTransactions(secopts *SecurityOptions) (*GetPendingStakingTransactionsResponse, error) {

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
func (api *KrakenAPIClient) ListOfStakingTransactions(secopts *SecurityOptions) (*ListOfStakingTransactionsResponse, error) {

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

/*****************************************************************************/
/*	UTILITIES														 	     */
/*****************************************************************************/

// Forge a HTTP request for Kraken REST API.
//
// in: resource - Targeted resource of the API. Example /0/public/Time
//
// in: httpMethod - HTTP Method to use. Read below for further information.
//
// in: query - Values for query string. Nillable.
//
// in: contentType - Mime type for the body. Read below for further information.
//
// in: body - Body of the request. Nillable. Read below for further information.
//
// in: headers - Headers to include with the request. Nillable. Read below for further information.
//
// return: The content of result field of the API response. Must be type asserted.
//
// error: other - other unexpected errors or failed checks.
//
// NOTES ON HTTP METHOD:
// Only GET & POST are allowed because these are the only methods used by the API. Can evolve in the future.
//
// NOTES ON CONTENT TYPE:
//   - Content type is ignored if httpMethod is GET.
//   - Only application/x-www-form-urlencoded is allowed because
//     the API uses exclusively form-encoded reauest payloads.
//
// NOTES ON BODY
//   - Body is ignored if httpMethod is GET.
//   - If contentType is application/x-www-form-urlencoded, the expected type for body is url.Values.
//
// NOTES ON HEADERS
//   - Provided headers map is not modified
//   - Some headers are managed by the client and will be ignored if provided in headers.
//   - User-Agent
//   - Content-Type
//   - API-Key
//   - API-Sign
func (api *KrakenAPIClient) forgeHTTPRequest(resource string, httpMethod string, query url.Values, contentType string, body interface{}, headers map[string]string) (*http.Request, error) {

	// Forge base request URL
	reqURL := fmt.Sprintf("%s%s", api.baseURL, resource)

	// Add query parameters if provided
	if len(query) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, query.Encode())
	}

	// Forge HTTP request
	var httpRequest *http.Request
	var err error

	switch httpMethod {
	// Perform GET request to Kraken REST API
	case http.MethodGet:
		// Prepare GET request without body
		httpRequest, err = http.NewRequest(httpMethod, reqURL, nil)
		if err != nil {
			return nil, err
		}

	// Perform POST request to Kraken REST API
	case http.MethodPost:
		if body != nil {
			// If a body is set, check content type and forge request accordingly
			switch contentType {
			case "application/x-www-form-urlencoded":
				form, ok := body.(url.Values)
				if ok {
					// Prepare POST request with form-encoded payload
					httpRequest, err = http.NewRequest(httpMethod, reqURL, strings.NewReader(form.Encode()))
					if err != nil {
						return nil, err
					}
					// Set content type header in request
					httpRequest.Header.Set(managedHeaderContentType, contentType)
				} else {
					// Unexpected body type for provided content type
					return nil, fmt.Errorf("body with type url.Values is expected for %s content type. Got %T", contentType, body)
				}

			// EXTENSION: Add new content type for POST method here
			default:
				// Not supported content type
				return nil, fmt.Errorf("provided content type is not supported by the client : %s", contentType)
			}
		} else {
			// Forge POST request without body
			httpRequest, err = http.NewRequest(httpMethod, reqURL, nil)
			if err != nil {
				return nil, err
			}
		}
	// EXTENSION: Add new http methods here
	default:
		// Not supported http method
		return nil, fmt.Errorf("provided HTTP method is not supported by the client : %s", httpMethod)
	}

	// Add managed headers
	httpRequest.Header.Add(managedHeaderUserAgent, api.agent)

	// Add provided headers to request
	// Ignore managed headers
	for header, val := range headers {
		if header != managedHeaderUserAgent && header != managedHeaderContentType && header != managedHeaderAPIKey && header != managedHeaderAPISign {
			httpRequest.Header.Add(header, val)
		}
	}

	// Return forged request
	return httpRequest, nil
}

// Forge & perform a query which targets a public endpoint.
//
// in: endpoint - Name of the public endpoint to use.
//
// in: httpMethod - HTTP Method to use. Read below for further information
//
// in: query - Values for query string. Nillable
//
// in: contentType - Mime type for the body. Read below for further information
//
// in: body - Body of the request. Nillable. Read below for further information
//
// in: headers - Headers to include with the request. Nillable. User-Agent, Content-Type will be set/overriden
//
// in: typ - Expected type for data contained in result field
//
// return: The content of result field of the API response. Must be type asserted.
//
// error: HTTPError - A HTTP error status code is returned with the HTTP response
//
// error: KrakenAPIClientErrorBundle - Contain all errors from the errors field of the API response.
//
// error: others - other unexpected errors or failed checks
//
// NOTES ON HTTP METHOD:
// Only GET & POST are allowed because these are the only methods used by the API. Can evolve in the future.
//
// NOTES ON CONTENT TYPE:
//   - Content type is ignored if httpMethod is GET.
//   - Only application/x-www-form-urlencoded is allowed because
//     the API uses exclusively form-encoded reauest payloads.
//
// NOTES ON BODY
//   - Body is ignored if httpMethod is GET
//   - If contentType is application/x-www-form-urlencoded, the expected type for body is url.Values
func (api *KrakenAPIClient) queryPublic(endpoint string, httpMethod string, query url.Values, contentType string, body interface{}, headers map[string]string, typ interface{}) (interface{}, error) {

	// Forge URL resource
	resource := fmt.Sprintf("/%s/public/%s", api.version, endpoint)

	// Forge HTTP request
	req, err := api.forgeHTTPRequest(resource, httpMethod, query, contentType, body, headers)
	if err != nil {
		return nil, err
	}

	// Perform API request and return results
	return api.doKrakenAPIRequest(req, typ)
}

// Execute a query which targets a private endpoint.
//
// in: endpoint - Name of the public endpoint to use.
//
// in: httpMethod - HTTP Method to use. Read below for further information
//
// in: query - Values for query string. Nillable
//
// in: contentType - Mime type for the body. Read below for further information
//
// in: body - Body of the request. Nillable. Read below for further information
//
// in: headers - Headers to include with the request. Nillable. User-Agent, Content-Type, API-Sign, API-Key will be set/overriden
//
// in: otp - One time password for signature. An empty string can be provided if OTP is not enabled
//
// in: typ - Expected type for data contained in result field
//
// return: The content of result field of the API response. Must be type asserted.
//
// error: HTTPError - A HTTP error status code is returned with the HTTP response
//
// error: KrakenAPIClientErrorBundle - Contain all errors from the errors field of the API response.
//
// error: others - other unexpected errors or failed checks
//
// NOTES ON HTTP METHOD:
// Only GET & POST are allowed because these are the only methods used by the API. Can evolve in the future.
//
// NOTES ON CONTENT TYPE:
//   - Content type is ignored if httpMethod is GET or if body is nil.
//   - Only empty string or application/x-www-form-urlencoded are allowed because
//     the API uses exclusively form-encoded request payloads.
//
// NOTES ON BODY
//   - Body is ignored if httpMethod is GET,
//   - If contentType is application/x-www-form-urlencoded, the expected type for body is url.Values
func (api *KrakenAPIClient) queryPrivate(endpoint string, httpMethod string, query url.Values, contentType string, body interface{}, headers map[string]string, otp string, typ interface{}) (interface{}, error) {

	// Forge URL resource
	resource := fmt.Sprintf("/%s/private/%s", api.version, endpoint)

	// Forge signature in the appropriate way according provided parameters
	// Right now - API only uses POST method and form-encoded body
	// Refactor if that changes
	var signature string
	var extendedBody interface{}
	var ctype string
	switch httpMethod {
	case http.MethodPost:
		if body != nil {
			// If body is not nil, proceed according content type
			switch contentType {
			case "application/x-www-form-urlencoded":
				// Type assertion for body : expect url.Values
				form, ok := body.(url.Values)
				if !ok {
					return nil, fmt.Errorf("form-encoded body is expected to forge signature but got %T as body", body)
				}

				// Create body
				data := make(url.Values, len(form))

				// Copy body
				for key := range form {
					data.Set(key, form.Get(key))
				}

				// Add otp to body if defined
				if otp != "" {
					data.Set("otp", otp)
				}

				// Add nonce and forge signature
				data.Set("nonce", fmt.Sprintf("%d", api.nonceGenerator.GetNextNonce()))
				signature = GetKrakenSignature(resource, data, api.secret)

				// Set extended body & content type
				extendedBody = data
				ctype = contentType
			default:
				return nil, fmt.Errorf("the provided content type is not supported by the client. Got %s", contentType)
			}
		} else {
			// Body is nil, default behavior is to create a form encoded body with signature data
			data := make(url.Values, 2)

			// Add otp to body if defined
			if otp != "" {
				data.Set("otp", otp)
			}

			// Add nonce and forge signature
			data.Set("nonce", fmt.Sprintf("%d", api.nonceGenerator.GetNextNonce()))
			signature = GetKrakenSignature(resource, data, api.secret)

			// Set extendedBody with form & content type with url form encoded
			extendedBody = data
			ctype = "application/x-www-form-urlencoded"
		}
	default:
		return nil, fmt.Errorf("the provided http method is not supported by the client. Got %s", httpMethod)
	}

	// Forge request
	req, err := api.forgeHTTPRequest(resource, httpMethod, query, ctype, extendedBody, headers)
	if err != nil {
		return nil, err
	}

	// Set/Override Api-Key and API-Sign headers in request
	req.Header[managedHeaderAPIKey] = []string{api.key}
	req.Header[managedHeaderAPISign] = []string{signature}

	// Perform API call
	return api.doKrakenAPIRequest(req, typ)
}

// Execute a HTTP Request to Kraken REST API.
//
// in: req - HTTP request to enrich and perform. User-Agent header will be set/overriden
//
// in: typ - Expected type for response
//
// return: Payload contained in result field of the response. Must be type asserted.
//
// error: HTTPError if http error status has been returned
//
// error: others - other unexpected errors or failed checks
func (api *KrakenAPIClient) doKrakenAPIRequest(req *http.Request, typ interface{}) (interface{}, error) {

	// Execute request using the provided http client
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not execute HTTP request. Got %s", err.Error())
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could read HTTP response body. Got %s", err.Error())
	}

	// Check status code for error status
	// API documentation states that "status codes other than 200 indicate
	// that there was an issue with the request reaching our servers"
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code: %d - body: %v", resp.StatusCode, respBody)
	}

	// Check mime type of response
	mimeType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("could not decode Content-Type header. Got %s", err.Error())
	}

	// Depending on response content type
	switch mimeType {
	case "application/octet-stream":
		// Return raw bytes from response
		return respBody, nil
	case "application/zip":
		// Return raw bytes from response
		return respBody, nil
	case "application/json":
		// Parse body
		err = json.Unmarshal(respBody, &typ)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshall JSON response. Got %s", err)
		}

		// Return response
		return typ, nil
	default:
		// Return error -> unsupported content type
		return nil, fmt.Errorf("response Content-Type is %s but should be application/json, application/octet-stream or application/zip", mimeType)
	}
}

// Generate the signature to attach to calls to Kraken REST API private endpoints.
//
// See https://docs.kraken.com/rest/#section/Authentication/Headers-and-Signature for more information
//
// in: 	url_path - URL path of the resource called. Example : /0/private/AddOrder
//
// in: 	values - URL encoded payload. Must contain "nonce"
//
// in:	secret - The secret value related to the API Key used to sign the request.
//
// out: Signature to attach to the request
func GetKrakenSignature(url_path string, values url.Values, secret []byte) string {

	sha := sha256.New()
	sha.Write([]byte(values.Get("nonce") + values.Encode()))
	shasum := sha.Sum(nil)

	mac := hmac.New(sha512.New, secret)
	mac.Write(append([]byte(url_path), shasum...))
	macsum := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(macsum)
}
