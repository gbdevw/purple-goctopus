package krakenapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

/*****************************************************************************/
/*	INTERFACE                                                                */
/*****************************************************************************/

// Interface for Kraken API client
type KrakenAPIClientIface interface {
	// GetServerTime Get the server time.
	GetServerTime(ctx context.Context) (*GetServerTimeResponse, error)
	// GetSystemStatus Get the current system status or trading mode.
	GetSystemStatus() (*GetSystemStatusResponse, error)
	// GetAssetInfo Get information about the assets that are available for deposit, withdrawal, trading and staking.
	GetAssetInfo(*GetAssetInfoOptions) (*GetAssetInfoResponse, error)
	// GetTradableAssetPairs Get tradable asset pairs.
	GetTradableAssetPairs(*GetTradableAssetPairsOptions) (*GetTradableAssetPairsResponse, error)
	// Get ticker information about a given list of pairs.
	// Note: Today's prices start at midnight UTC
	GetTickerInformation(*GetTickerInformationOptions) (*GetTickerInformationResponse, error)
	// GetOHLCData get Open, High, Low & Close indicators.
	// Note: the last entry in the OHLC array is for the current, not-yet-committed
	// frame and will always be present, regardless of the value of since.
	GetOHLCData(GetOHLCDataParameters, *GetOHLCDataOptions) (*GetOHLCDataResponse, error)
	// GetOrderBook Get order by for a given pair
	GetOrderBook(GetOrderBookParameters, *GetOrderBookOptions) (*GetOrderBookResponse, error)
	// GetRecentTrades Get up to the 1000 most recent trades by default
	GetRecentTrades(GetRecentTradesParameters, *GetRecentTradesOptions) (*GetRecentTradesResponse, error)
	// GetRecentSpreads Get recent spreads.
	GetRecentSpreads(GetRecentSpreadsParameters, *GetRecentSpreadsOptions) (*GetRecentSpreadsResponse, error)
	// GetAccountBalance - Retrieve all cash balances, net of pending withdrawals.
	GetAccountBalance(*SecurityOptions) (*GetAccountBalanceResponse, error)
	// GetTradeBalance - Retrieve a summary of collateral balances, margin position valuations, equity and margin level.
	GetTradeBalance(*GetTradeBalanceOptions, *SecurityOptions) (*GetTradeBalanceResponse, error)
	// GetOpenOrders - Retrieve information about currently open orders.
	GetOpenOrders(*GetOpenOrdersOptions, *SecurityOptions) (*GetOpenOrdersResponse, error)
	// GetClosedOrders -
	// Retrieve information about orders that have been closed (filled or cancelled).
	// 50 results are returned at a time, the most recent by default.
	GetClosedOrders(*GetClosedOrdersOptions, *SecurityOptions) (*GetClosedOrdersResponse, error)
	// QueryOrdersInfo - Retrieve information about specific orders.
	QueryOrdersInfo(QueryOrdersParameters, *QueryOrdersOptions, *SecurityOptions) (*QueryOrdersInfoResponse, error)
	// GetTradesHistory -
	// Retrieve information about trades/fills.
	// 50 results are returned at a time, the most recent by default.
	//
	// Unless otherwise stated, costs, fees, prices, and volumes are specified with the precision for the asset pair
	// (pair_decimals and lot_decimals), not the individual assets' precision (decimals).
	GetTradesHistory(*GetTradesHistoryOptions, *SecurityOptions) (*GetTradesHistoryResponse, error)
	// QueryTradesInfo - Retrieve information about specific trades/fills.
	QueryTradesInfo(QueryTradesParameters, *QueryTradesOptions, *SecurityOptions) (*QueryTradesInfoResponse, error)
	// GetOpenPositions - Get information about open margin positions.
	GetOpenPositions(*GetOpenPositionsOptions, *SecurityOptions) (*GetOpenPositionsResponse, error)
	// GetLedgersInfo - Retrieve information about ledger entries. 50 results are returned at a time, the most recent by default.
	GetLedgersInfo(*GetLedgersInfoOptions, *SecurityOptions) (*GetLedgersInfoResponse, error)
	// QueryLedgers - Retrieve information about specific ledger entries.
	QueryLedgers(QueryLedgersParameters, *QueryLedgersOptions, *SecurityOptions) (*QueryLedgersResponse, error)
	// RequestExportReport - Request export of trades or ledgers.
	RequestExportReport(RequestExportReportParameters, *RequestExportReportOptions, *SecurityOptions) (*RequestExportReportResponse, error)
	// GetExportReportStatus - Get status of requested data exports.
	GetExportReportStatus(GetExportReportStatusParameters, *SecurityOptions) (*GetExportReportStatusResponse, error)
	// RetrieveDataExport Get report as a zip
	RetrieveDataExport(RetrieveDataExportParameters, *SecurityOptions) (*RetrieveDataExportResponse, error)
	// DeleteExportReport - Delete exported trades/ledgers report
	DeleteExportReport(DeleteExportReportParameters, *SecurityOptions) (*DeleteExportReportResponse, error)
	// Place a new order.
	AddOrder(AddOrderParameters, *AddOrderOptions, *SecurityOptions) (*AddOrderResponse, error)
	// Send an array of orders (max: 15). Any orders rejected due to order validations, will be dropped
	// and the rest of the batch is processed. All orders in batch should be limited to a single pair.
	// The order of returned txid's in the response array is the same as the order of the order list sent in request.
	AddOrderBatch(AddOrderBatchParameters, *AddOrderBatchOptions, *SecurityOptions) (*AddOrderBatchResponse, error)
	// Edit volume and price on open orders. Uneditable orders include margin orders, triggered stop/profit orders,
	// orders with conditional close terms attached, those already cancelled or filled, and those where the executed
	// volume is greater than the newly supplied volume. post-only flag is not retained from original order after
	// successful edit. post-only needs to be explicitly set on edit request.
	EditOrder(EditOrderParameters, *EditOrderOptions, *SecurityOptions) (*EditOrderResponse, error)
	// Cancel a particular open order (or set of open orders) by txid or userref
	CancelOrder(CancelOrderParameters, *SecurityOptions) (*CancelOrderResponse, error)
	// Cancel all open orders
	CancelAllOrders(*SecurityOptions) (*CancelAllOrdersResponse, error)
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
	CancelAllOrdersAfterX(CancelCancelAllOrdersAfterXParameters, *SecurityOptions) (*CancelAllOrdersAfterXResponse, error)
	// Cancel multiple open orders by txid or userref
	CancelOrderBatch(CancelOrderBatchParameters, *SecurityOptions) (*CancelOrderBatchResponse, error)
	// Retrieve methods available for depositing a particular asset.
	GetDepositMethods(GetDepositMethodsParameters, *SecurityOptions) (*GetDepositMethodsResponse, error)
	// Retrieve (or generate a new) deposit addresses for a particular asset and method.
	GetDepositAddresses(GetDepositAddressesParameters, *GetDepositAddressesOptions, *SecurityOptions) (*GetDepositAddressesResponse, error)
	// Retrieve information about recent deposits made.
	GetStatusOfRecentDeposits(GetStatusOfRecentDepositsParameters, *GetStatusOfRecentDepositsOptions, *SecurityOptions) (*GetStatusOfRecentDepositsResponse, error)
	// Retrieve fee information about potential withdrawals for a particular asset, key and amount.
	GetWithdrawalInformation(GetWithdrawalInformationParameters, *SecurityOptions) (*GetWithdrawalInformationResponse, error)
	// Make a withdrawal request.
	WithdrawFunds(WithdrawFundsParameters, *SecurityOptions) (*WithdrawFundsResponse, error)
	// Retrieve information about recently requests withdrawals.
	GetStatusOfRecentWithdrawals(GetStatusOfRecentWithdrawalsParameters, *GetStatusOfRecentWithdrawalsOptions, *SecurityOptions) (*GetStatusOfRecentWithdrawalsResponse, error)
	// Cancel a recently requested withdrawal, if it has not already been successfully processed.
	RequestWithdrawalCancellation(RequestWithdrawalCancellationParameters, *SecurityOptions) (*RequestWithdrawalCancellationResponse, error)
	// Transfer from Kraken spot wallet to Kraken Futures holding wallet. Note that a transfer in the other direction must be requested via the Kraken Futures API endpoint.
	RequestWalletTransfer(RequestWalletTransferParameters, *SecurityOptions) (*RequestWalletTransferResponse, error)
	// StakeAsset stake an asset from spot wallet.
	StakeAsset(StakeAssetParameters, *SecurityOptions) (*StakeAssetResponse, error)
	// UnstakeAsset unstake an asset from your staking wallet.
	UnstakeAsset(UnstakeAssetParameters, *SecurityOptions) (*UnstakeAssetResponse, error)
	// ListOfStakeableAssets returns the list of assets that the user is able to stake.
	ListOfStakeableAssets(*SecurityOptions) (*ListOfStakeableAssetsResponse, error)
	// GetPendingStakingTransactions returns the list of pending staking transactions.
	GetPendingStakingTransactions(*SecurityOptions) (*GetPendingStakingTransactionsResponse, error)
	// ListOfStakingTransactions returns the list of 1000 recent staking transactions from past 90 days.
	ListOfStakingTransactions(*SecurityOptions) (*ListOfStakingTransactionsResponse, error)
}

/*****************************************************************************/
/*	COMMON TYPES                                                             */
/*****************************************************************************/

// KrakenAPIResponse wraps the Kraken API JSON response
type KrakenAPIResponse struct {
	// Errors returned with the response
	Error []string `json:"error"`
	// Result for the request
	Result interface{} `json:"result"`
}

// Container for security options to use during the API call (2FA, ...)
type SecurityOptions struct {
	// Second factor to use to sign request (authenticator app or password).
	// An empty string can be used if 2FA is not enabled.
	// Refer to https://support.kraken.com/hc/en-us/articles/360000714526-How-does-two-factor-authentication-2FA-for-API-keys-work- for additional information.
	SecondFactor string
}

/*****************************************************************************/
/*	ENUMS                                                                    */
/*****************************************************************************/

// Values for system status
const (
	Online      = "online"
	Maintenance = "maintenance"
	CancelOnly  = "cancel_only"
	PostOnly    = "post_only"
)

// Values for Asset class
const (
	// Currency asset class
	AssetClassCurrency = "currency"
)

const (
	// Get all asset pair info.
	InfoAll = "info"
	// Get leverage info
	InfoLeverage = "leverage"
	// Get fees info
	InfoFees = "fees"
	// Get margin info
	InfoMargin = "margin"
)

// Values for order statuses
const (
	OStatusPending  = "pending"
	OStatusOpen     = "open"
	OStatusClosed   = "closed"
	OStatusCanceled = "canceled"
	OStatusExpired  = "expired"
)

// Type for CloseTime
type CloseTime string

// Values for CloseTime
const (
	UseOpen  CloseTime = "open"
	UseClose CloseTime = "close"
	UseBoth  CloseTime = "both"
)

// Values for trade types
const (
	TradeTypeAll             = "all"
	TradeTypeAnyPosition     = "any position"
	TradeTypeClosedPosition  = "closed position"
	TradeTypeClosingPosition = "closing position"
	TradeTypeNoPosition      = "no position"
)

// Values for position statuses
const (
	PositionOpen   = "open"
	PositionClosed = "closed"
)

// Type for OHLC data interval
type OHLCInterval int

// Values for OHLC data interval
const (
	// 1 minute
	M1 OHLCInterval = 1
	// 5 minutes
	M5 OHLCInterval = 5
	// 15 minutes
	M15 OHLCInterval = 15
	// 30 minutes
	M30 OHLCInterval = 30
	// 60 minutes (1 hour)
	M60 OHLCInterval = 60
	// 240 minutes (4 hours)
	M240 OHLCInterval = 240
	// 1440 minutes (1 day)
	M1440 OHLCInterval = 1440
	// 10080 minutes (1 week)
	M10080 OHLCInterval = 10080
	// 21600 minutes (2 weeks)
	M21600 OHLCInterval = 21600
)

// Values for ledger entry
const (
	EntryTypeTrade      = "trade"
	EntryTypeDeposit    = "deposit"
	EntryTypeWithdrawal = "withdrawal"
	EntryTypeTransfer   = "transfer"
	EntryTypeMargin     = "margin"
	EntryTypeRollover   = "rollover"
	EntryTypeSpend      = "spend"
	EntryTypeReceive    = "receive"
	EntryTypeSettled    = "settled"
	EntryTypeAdjustment = "adjustment"
)

// Values for report types
const (
	ReportTrades  = "trades"
	ReportLedgers = "ledgers"
)

// Values for report formats
const (
	ReportFmtCSV = "CSV"
	ReportFmtTSV = "TSV"
)

// Values for order types
const (
	OTypeMarket          = "market"
	OTypeLimit           = "limit"
	OTypeStopLoss        = "stop-loss"
	OTypeTakeProfit      = "take-profit"
	OTypeStopLossLimit   = "stop-loss-limit"
	OTypeTakeProfitLimit = "take-profit-limit"
	OTypeSettlePosition  = "settle-position"
)

// Value for sides
const (
	Buy  = "buy"
	Sell = "sell"
)

// Values for triggers
const (
	TriggerLast  = "last"
	TriggerIndex = "index"
)

// Values for self trade prevention flags
const (
	StpCancelNewest = "cancel-newest"
	StpCancelOldest = "cancel-oldest"
	StpCancelBoth   = "cancel-both"
)

// Values for order flags
const (
	OFlagPost                    = "post"
	OFlagFeeInBase               = "fcib"
	OFlagFeeInQuote              = "fciq"
	OFlagNoMarketPriceProtection = "nompp"
	OFlagVolumeInQuote           = "viqc"
)

// Type for time in force flags
const (
	GoodTilCanceled   = "GTC"
	ImmediateOrCancel = "IOC"
	GoodTilDate       = "GTD"
)

// Values for ledger type
const (
	LedgerAll        = "all"
	LedgerDeposit    = "deposit"
	LedgerWithdrawal = "withdrawal"
	LedgerTrade      = "trade"
	LedgerMargin     = "margin"
	LedgerRollover   = "rollover"
	LedgerCredit     = "credit"
	LedgerTransfer   = "transfer"
	LedgerSettled    = "settled"
	LedgerStaking    = "staking"
	LedgerSale       = "sale"
)

// Values for report deletion
const (
	DeleteReport = "delete"
	CancelReport = "cancel"
)

// Transaction states as described in https://github.com/globalcitizen/ifex-protocol/blob/master/draft-ifex-00.txt#L837
const (
	TxStateInitial = "Initial"
	TxStatePending = "Pending"
	TxStateSettled = "Settled"
	TxStateSuccess = "Success"
	TxStateFailure = "Failure"
	TxStatePartial = "Partial"
)

// Additional properties for transaction status
const (
	// A return transaction initiated by Kraken
	TxStatusReturn = "return"
	// Deposit is on hold pending review
	TxStatusOnHold = "onhold"
	// Cancelation requested
	TxCancelPending = "cancel-pending"
	// Canceled
	TxCanceled = "canceled"
	// CancelDenied
	TxCancelDenied = "cancel-denied"
)

// Values for types of staking transactions
const (
	StakingBonding   = "bonding"
	StakingReward    = "reward"
	StakingUnbonding = "unbonding"
)

/*****************************************************************************/
/*	PUBLIC : MARKET DATA                                                     */
/*****************************************************************************/

// Reponse for Get Server Time
type GetServerTimeResponse struct {
	KrakenAPIResponse
	Result struct {
		// Unix timestamp
		Unixtime int64 `json:"unixtime"`
		// RFC 1123 time format
		Rfc1123 string `json:"rfc1123"`
	} `json:"result,omitempty"`
}

// Response for Get System Status
type GetSystemStatusResponse struct {
	KrakenAPIResponse
	Result struct {
		// System status
		Status string `json:"status"`
		// Current timestamp (RFC3339)
		Timestamp string `json:"timestamp"`
	} `json:"result,omitempty"`
}

// AssetInfo represents an asset information
type AssetInfo struct {
	// Asset class
	AssetClass string `json:"aclass"`
	// Alternate name
	Altname string `json:"altname"`
	// Scaling decimal places for record keeping
	Decimals int `json:"decimals"`
	// Scaling decimal places for output display
	DisplayDecimals int `json:"display_decimals"`
	// Collateral value
	CollateralValue float64 `json:"collateral_value"`
}

// GetAssetInfo options
type GetAssetInfoOptions struct {
	// List of assets to get info on.
	// Defaults to all assets.
	// A nil value triggers default behavior.
	Assets []string
	// Asset class.
	// Defaults to 'currency'.
	// An empty string triggers default behavior.
	AssetClass string
}

// GetAssetInfo response
type GetAssetInfoResponse struct {
	KrakenAPIResponse
	Result map[string]AssetInfo `json:"result,omitempty"`
}

// AssetPairInfo represents asset pair information
type AssetPairInfo struct {
	// Alternate pair name
	Altname string `json:"altname"`
	// Name on Websocket API
	WsName string `json:"wsname"`
	// Asset class of base component
	AssetClassBase string `json:"aclass_base"`
	// Asset id of base component
	Base string `json:"base"`
	// Asset class of quote component
	AssetClassQuote string `json:"aclass_quote"`
	// Asset id of quote component
	Quote string `json:"quote"`
	// Scaling decimal places for pair
	PairDecimals int `json:"pair_decimals"`
	// Scaling decimal places for volume
	LotDecimals int `json:"lot_decimals"`
	// Amount to multiply lot volume by to get currency volume
	LotMultiplier int `json:"lot_multiplier"`
	// Array of leverage amounts available when buying
	LeverageBuy []int `json:"leverage_buy"`
	// Array of leverage amounts available when selling
	LeverageSell []int `json:"leverage_sell"`
	// Fee schedule array in [volume, percent fee] tuples
	Fees [][]float64 `json:"fees"`
	// Maker fee schedule array in [volume, percent fee] tuples (if on maker/taker)
	FeesMaker [][]float64 `json:"fees_maker"`
	// // Volume discount currency
	FeeVolumeCurrency string `json:"fee_volume_currency"`
	// Margin call level
	MarginCall int `json:"margin_call"`
	// Stop-out/Liquidation margin level
	MarginStop int `json:"margin_stop"`
	// Order minimum
	OrderMin string `json:"ordermin"`
}

// Options for GetTradableAssetPairs
type GetTradableAssetPairsOptions struct {
	// Pairs to get info on.
	// Defaults to all pairs
	// A nil value triggers default behavior.
	Pairs []string
	// Info to retrieve.
	// Defaults to InfoAll (info)
	// An empty string triggers default behavior.
	Info string
}

// Response for GetTradableAssetPairs
type GetTradableAssetPairsResponse struct {
	KrakenAPIResponse
	// Map each assert pair (ex: 1INCHEUR) to its info
	Result map[string]AssetPairInfo `json:"result,omitempty"`
}

// Asset Ticker Info
type AssetTickerInfo struct {
	// Ask array(<price>, <whole lot volume>, <lot volume>)
	Ask []string `json:"a"`
	// Bid array(<price>, <whole lot volume>, <lot volume>)
	Bid []string `json:"b"`
	// Last trade closed array(<price>, <lot volume>)
	Close []string `json:"c"`
	// Volume array(<today>, <last 24 hours>)
	Volume []string `json:"v"`
	// Volume weighted average price array(<today>, <last 24 hours>)
	VolumeAveragePrice []string `json:"p"`
	// Number of trades array(<today>, <last 24 hours>)
	Trades []int `json:"t"`
	// Low array(<today>, <last 24 hours>)
	Low []string `json:"l"`
	// High array(<today>, <last 24 hours>)
	High []string `json:"h"`
	// Today's opening price
	OpeningPrice string `json:"o"`
}

// GetTickerInformation Options
type GetTickerInformationOptions struct {
	// Asset pairs to get data for
	Pairs []string
}

// GetTickerInformation Response
type GetTickerInformationResponse struct {
	KrakenAPIResponse
	// Ticker data by pair
	Result map[string]AssetTickerInfo `json:"result,omitempty"`
}

// OHLC data
type OHLCData struct {
	// UTC timestamp for OHLC data
	Timestamp time.Time `json:"time"`
	// Open price
	Open string `json:"open"`
	// High
	High string `json:"high"`
	// Low
	Low string `json:"low"`
	// Close
	Close string `json:"close"`
	// Volume Weighted Average
	Avg string `json:"vwavg"`
	// Volume
	Volume string `json:"volume"`
	// Number of trades
	Count int `json:"count"`
}

// GetOHLCData required parameters
type GetOHLCDataParameters struct {
	// Asset pair to get data for.
	Pair string
}

// GetOHLCData options
type GetOHLCDataOptions struct {
	// Time frame interval in minutes
	// Default to 1.
	// A value of 0 triggers default behavior.
	Interval OHLCInterval
	// Return up to 720 OHLC data points since given timestamp.
	// By default, return the most recent OHLC data points.
	// A nil value triggers default behavior.
	Since *time.Time
}

type GetOHLCDataResult struct {
	// ID to be used as since when polling for new, committed OHLC data
	Last int64 `json:"last"`
	// OHLC Data by pairs
	OHLC map[string][]OHLCData `json:"ohlc"`
}

// GetOHLCData Response
type GetOHLCDataResponse struct {
	KrakenAPIResponse
	Result GetOHLCDataResult `json:"result,omitempty"`
}

// Order Book Entry
type OrderBookEntry struct {
	// UTC timestamp
	Timestamp time.Time `json:"time"`
	// Price level
	Price string `json:"price"`
	// Volume
	Volume string `json:"volume"`
}

// OrderBook
type OrderBook struct {
	// Buy side of the order book
	Asks []OrderBookEntry `json:"asks"`
	// Sell side of the order book
	Bids []OrderBookEntry `json:"bids"`
}

// GetOrderBook required parameters
type GetOrderBookParameters struct {
	// Asset pair to get data for.
	Pair string
}

// GetOrderBook options
type GetOrderBookOptions struct {
	// Maximum number of bid/ask entries : [1,500].
	// Defaults to 100.
	// A value of 0 triggers default behavior.
	Count int
}

// GetOrderBook Response
type GetOrderBookResponse struct {
	KrakenAPIResponse
	// Order books by asset pairs
	Result map[string]OrderBook `json:"result,omitempty"`
}

// Trade data
type Trade struct {
	// Timestamp
	Timestamp int64 `json:"time"`
	// Price for the trade
	Price string `json:"price"`
	// Volume
	Volume string `json:"volume"`
	// Side : buy/sell -> b or s
	Side string `json:"side"`
	// Type - market/limit -> m or l
	Type string `json:"type"`
	// Misc. info
	Miscellaneous string `json:"misc"`
	// Trade Id
	Id int64 `json:"id"`
}

// GetRecentTrades required parameters
type GetRecentTradesParameters struct {
	// Asset pair to get data for.
	Pair string
}

// GetRecentTrades options
type GetRecentTradesOptions struct {
	// Return up to 1000 recent trades since given timestamp.
	// By default, return the most recent trades.
	// A nil value triggers default behavior.
	Since *time.Time
}

// GetRecentTrades Result
type GetRecentTradesResult struct {
	// ID to be used as since when polling for new trade data
	Last string `json:"last"`
	// Trade data by pair
	Trades map[string][]Trade `json:"trades"`
}

// GetRecentTrades Response
type GetRecentTradesResponse struct {
	KrakenAPIResponse
	Result GetRecentTradesResult `json:"result"`
}

// Spread data
type SpreadData struct {
	// Timestamp
	Timestamp time.Time `json:"time"`
	// Best bid
	BestBid string `json:"bbid"`
	// Best ask
	BestAsk string `json:"bask"`
}

// GetRecentSpreads required parameters
type GetRecentSpreadsParameters struct {
	// Asset pair to get data for.
	Pair string
}

// GetRecentSpreads options
type GetRecentSpreadsOptions struct {
	// Return up to 1000 recent spreads since given timestamp.
	// By default, return the most recent spreads.
	// A nil value triggers default behavior.
	Since *time.Time
}

// GetRecentSpreads result
type GetRecentSpreadsResult struct {
	// ID to be used as since when polling for new spread data
	Last string `json:"last"`
	// Spreads by pair
	Spreads map[string][]SpreadData `json:"spreads"`
}

// GetRecentSpreads response
type GetRecentSpreadsResponse struct {
	KrakenAPIResponse
	Result GetRecentSpreadsResult `json:"result"`
}

/*****************************************************************************/
/*	PRIVATE : USER DATA                                                      */
/*****************************************************************************/

// GetAccountBalanceResponse contains Get Account Balance response data.
type GetAccountBalanceResponse struct {
	KrakenAPIResponse
	// Balance for each possessed asset
	Result map[string]string `json:"result"`
}

// GetTradeBalanceOptions contains Get Trade Balance optional parameters.
type GetTradeBalanceOptions struct {
	// Base asset used to determine balance.
	// Defaults to ZUSD.
	Asset string
}

// GetTradeBalance Result
type GetTradeBalanceResult struct {
	// Equivalent balance (combined balance of all currencies)
	EquivalentBalance string `json:"eb"`
	// Trade balance (combined balance of all equity currencies)
	TradeBalance string `json:"tb"`
	// Margin amount of open positions
	MarginAmount string `json:"m"`
	// Unrealized net profit/loss of open positions
	UnrealizedNetPNL string `json:"n"`
	// Cost basis of open positions
	CostBasis string `json:"c"`
	// Current floating valuation of open positions
	FloatingValuation string `json:"v"`
	// Equity: trade balance + unrealized net profit/loss
	Equity string `json:"e"`
	// Free margin: Equity - initial margin (maximum margin available to open new positions)
	FreeMargin string `json:"mf"`
	// Margin level: (equity / initial margin) * 100
	MarginLevel string `json:"ml"`
}

// GetTradeBalanceResponse contains Get Trade Balance response data.
type GetTradeBalanceResponse struct {
	KrakenAPIResponse
	Result GetTradeBalanceResult `json:"result"`
}

// Description for a Order Info
type OrderInfoDescription struct {
	// Asset pair
	Pair string `json:"pair"`
	// Order direction (buy/sell)
	Type string `json:"type"`
	// Order type. Enum: "market" "limit" "stop-loss" "take-profit" "stop-loss-limit" "take-profit-limit" "settle-position"
	OrderType string `json:"ordertype"`
	// Limit or trigger price depending on order type
	Price string `json:"price"`
	// Limit price for stop/take orders
	Price2 string `json:"price2"`
	// Amount of leverage
	Leverage string `json:"leverage"`
	// Textual order description
	OrderDescription string `json:"order"`
	// Conditional close order description
	CloseOrderDescription string `json:"close,omitempty"`
}

// OrderInfo contains order data.
type OrderInfo struct {
	// Referral order transaction ID that created this order
	ReferralOrderTransactionId string `json:"refid"`
	// Optional user defined reference ID
	UserReferenceId string `json:"userref"`
	// Status of order. Enum: "pending" "open" "closed" "canceled" "expired"
	Status string `json:"status"`
	// Unix timestamp of when order was placed
	OpenTimestamp int64 `json:"opentm"`
	// Unix timestamp of order start time (or 0 if not set)
	StartTimestamp int64 `json:"starttm"`
	// Unix timestamp of order end time (or 0 if not set)
	ExpireTimestamp int64 `json:"expiretm"`
	// Order description info
	Description OrderInfoDescription `json:"descr"`
	// Volume of order
	Volume string `json:"vol"`
	// Volume executed
	VolumeExecuted string `json:"vol_exec"`
	// Total cost
	Cost string `json:"cost"`
	// Total fee
	Fee string `json:"fee"`
	// Average price
	Price string `json:"price"`
	// Stop price
	StopPrice string `json:"stopprice"`
	// Triggered limit price
	LimitPrice string `json:"limitprice"`
	// Price signal used to trigger "stop-loss" "take-profit" "stop-loss-limit" "take-profit-limit" orders.
	// Enum: last, index. last is the implied trigger if field is not set.
	Trigger string `json:"trigger"`
	// Comma delimited list of miscellaneous info
	Miscellaneous string `json:"misc"`
	// Comma delimited list of order flags
	OrderFlags string `json:"oflags"`
	// List of trade IDs related to order (if trades info requested and data available)
	Trades []string `json:"trades,omitempty"`
	// If order is closed, Unix timestamp of when order was closed
	CloseTimestamp int64 `json:"closetm,omitempty"`
	// Additional info on status if any
	Reason string `json:"reason,omitempty"`
}

// GetOpenOrdersOptions contains Get Open Orders optional parameters.
type GetOpenOrdersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Restrict results to given user reference id.
	UserReference *int64
}

// GetOpenOrders Result
type GetOpenOrdersResult struct {
	// Keys are transaction ID and values the related open order.
	Open map[string]OrderInfo `json:"open"`
}

// GetOpenOrdersResponse contains Get Open Orders response data.
type GetOpenOrdersResponse struct {
	KrakenAPIResponse
	Result GetOpenOrdersResult `json:"result"`
}

// GetClosedOrdersOptions contains Get Closed Orders optional parameters.
type GetClosedOrdersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Restrict results to given user reference id.
	UserReference *int64
	// Starting unix timestamp or order tx ID of results (exclusive)
	Start *time.Time
	// Ending unix timestamp or order tx ID of results (inclusive)
	End *time.Time
	// Result offset for pagination
	Offset *int64
	// Which time to use to search.
	// Defaults to "both". Values: "open" "close" "both"
	Closetime CloseTime
}

// GetClosedOrders Result
type GetClosedOrdersResult struct {
	// Map where keys are transaction ID and values the related closed order.
	Closed map[string]OrderInfo `json:"closed"`
	// Amount of available order info matching criteria.
	Count int `json:"count"`
}

// GetClosedOrdersResponse contains Get Closed Orders response data.
type GetClosedOrdersResponse struct {
	KrakenAPIResponse
	Result GetClosedOrdersResult `json:"result"`
}

// QueryOrdersParameters contains Auery Orders required parameters
type QueryOrdersParameters struct {
	// List of transaction IDs to query info about (50 maximum).
	TransactionIds []string
}

// QueryOrdersOptions contains Query Orders optional parameters.
type QueryOrdersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Restrict results to given user reference id.
	UserReference *int64
}

// QueryOrdersInfoResponse contains Query Orders Info response data.
type QueryOrdersInfoResponse struct {
	KrakenAPIResponse
	// Map where keys are transaction ID and values the requested orders
	Result map[string]OrderInfo `json:"result"`
}

// TradeInfo contains full trade information
type TradeInfo struct {
	// Order responsible for execution of trade
	OrderTransactionId string `json:"ordertxid"`
	// Asset pair
	Pair string `json:"pair"`
	// Unix timestamp for the trade
	Timestamp float64 `json:"time"`
	// Trade direction (buy/sell)
	Type string `json:"type"`
	// Order type. Enum: "market" "limit" "stop-loss" "take-profit" "stop-loss-limit" "take-profit-limit" "settle-position"
	OrderType string `json:"ordertype"`
	// Average price order was executed at
	Price string `json:"price"`
	// Total cost of order
	Cost string `json:"cost"`
	// Total fee
	Fee string `json:"fee"`
	// Volume
	Volume string `json:"vol"`
	// Initial margin
	Margin string `json:"margin"`
	// Comma delimited list of miscellaneous info:
	// closing — Trade closes all or part of a position
	Miscellaneous string `json:"misc"`
	// Position status (open/closed)
	// - Only present if trade opened a position
	PositionStatus string `json:"posstatus,omitempty"`
	// Average price of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedPrice string `json:"cprice,omitempty"`
	// Total cost of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedCost string `json:"ccost,omitempty"`
	// Total fee of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedFee string `json:"cfee,omitempty"`
	// Total fee of closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedVolume string `json:"cvol,omitempty"`
	// Total margin freed in closed portion of position (quote currency)
	// - Only present if trade opened a position
	ClosedMargin string `json:"cmargin,omitempty"`
	// Net profit/loss of closed portion of position (quote currency, quote currency scale)
	// - Only present if trade opened a position
	ClosedNetPNL string `json:"net,omitempty"`
	// List of closing trades for position (if available)
	// - Only present if trade opened a position
	ClosingTrades []string `json:"trades,omitempty"`
}

// GetTradesHistoryOptions contains Get Trade History optional parameters.
type GetTradesHistoryOptions struct {
	// Type of trade.
	// Defaults to "all".
	// Values: "all" "any position" "closed position" "closing position" "no position"
	Type string
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
	// Starting unix timestamp or order tx ID of results (exclusive).
	Start *time.Time
	// Ending unix timestamp or order tx ID of results (inclusive).
	End *time.Time
	// Result offset for pagination.
	Offset *int64
}

// GetTradesHistory Result
type GetTradesHistoryResult struct {
	// Map where each key is a transaction ID and value a trade info object
	Trades map[string]TradeInfo `json:"trades"`
	// Amount of available trades matching criteria
	Count int
}

// GetTradesHistoryResponse contains Get Trades History response data.
type GetTradesHistoryResponse struct {
	KrakenAPIResponse
	Result GetTradesHistoryResult `json:"result"`
}

// QueryTradesParameters contains Query Trades required parameters
type QueryTradesParameters struct {
	// List of transaction IDs to query info about (20 maximum).
	TransactionIds []string
}

// QueryTradesOptions contains Query Trades optional parameters.
type QueryTradesOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
}

// QueryTradesInfoResponse contains Query Trades Info response data.
type QueryTradesInfoResponse struct {
	KrakenAPIResponse
	// Map where keys are transaction ID and values the requested trades.
	Result map[string]TradeInfo `json:"result"`
}

// PositionInfo contains position information
type PositionInfo struct {
	// Order ID responsible for the position
	OrderTransactionId string `json:"ordertxid"`
	// Position status
	PositionStatus string `json:"posstatus"`
	// Asset pair
	Pair string `json:"pair"`
	// Unix timestamp of trade
	Timestamp float64 `json:"time"`
	// Direction (buy/sell) of position
	Type string `json:"type"`
	// Order type used to open position
	OrderType string `json:"ordertype"`
	// Opening cost of position (in quote currency)
	Cost string `json:"cost"`
	// Opening fee of position (in quote currency)
	Fee string `json:"fee"`
	// Position opening size (in base currency)
	Volume string `json:"vol"`
	// Quantity closed (in base currency)
	ClosedVolume string `json:"vol_closed"`
	// Initial margin consumed (in quote currency)
	Margin string `json:"margin"`
	// Current value of remaining position (if docalcs requested)
	Value string `json:"value,omitempty"`
	// Unrealised P&L of remaining position (if docalcs requested)
	Net string `json:"net,omitempty"`
	// Funding cost and term of position
	Terms string `json:"terms"`
	// Timestamp of next margin rollover fee
	RolloverTimestamp string `json:"rollovertm"`
	// Comma delimited list of add'l info
	Miscellaneous string `json:"misc"`
	// Comma delimited list of opening order flags
	OrderFlags string `json:"oflags"`
}

// GetOpenPositionsOptions contains Get Open Positions optional parameters
type GetOpenPositionsOptions struct {
	// List of txids to limit output to
	TransactionIds []string
	// Whether to include P&L calculations.
	// Defaults to false.
	DoCalcs bool
	// Consolidate positions by market/pair.
	// Value: "market"
	// Consolidation is disabled because using market
	// changes radically the response payload and cause client to fail
	// Consolidation string
}

// GetOpenPositionsResponse contains Get Open Positions response data.
type GetOpenPositionsResponse struct {
	KrakenAPIResponse
	// Map where each key is a transaction ID and values an open position description.
	Result map[string]PositionInfo `json:"result"`
}

// LedgerEntry contains ledger entry data.
type LedgerEntry struct {
	// Reference Id
	ReferenceId string `json:"refid"`
	// Unix timestamp of ledger
	Timestamp float64 `json:"time"`
	// Type of ledger entry
	Type string `json:"type"`
	// Additional info relating to the ledger entry type, where applicable
	SubType string `json:"subtype,omitempty"`
	// Asset class
	AssetClass string `json:"aclass"`
	// Asset
	Asset string `json:"asset"`
	// Transaction amount
	Amount string `json:"amount"`
	// Transaction fee
	Fee string `json:"fee"`
	// Resulting balance
	Balance string `json:"balance"`
}

// GetLedgersInfoOptions contains Get Ledgers Info optional parameters.
type GetLedgersInfoOptions struct {
	// List of assets to restrict output to.
	// By default, all assets are accepted.
	Assets []string
	// Asset class to restrict output to.
	// Defaults to "currency".
	AssetClass string
	// Type of ledger to retrieve.
	// Defaults to "all".
	// Values: "all" "deposit" "withdrawal" "trade" "margin" "rollover" "credit" "transfer" "settled" "staking" "sale"
	Type string
	// Starting unix timestamp or order tx ID of results (exclusive).
	Start *time.Time
	// Ending unix timestamp or order tx ID of results (inclusive).
	End *time.Time
	// Result offset for pagination.
	Offset *int64
}

// GetLedgersInfo Result
type GetLedgersInfoResult struct {
	// Map where each key is a ledger entry ID and value a ledger entry
	Ledgers map[string]LedgerEntry `json:"ledger"`
	// Amount of available ledger info matching criteria
	Count int `json:"count"`
}

// GetLedgersInfoResponse contains Get Ledgers Info response data.
type GetLedgersInfoResponse struct {
	KrakenAPIResponse
	Result GetLedgersInfoResult `json:"result"`
}

// QueryLedgersParameters contains Query Ledgers required parameters.
type QueryLedgersParameters struct {
	// List of ledger IDs to query info about (20 maximum).
	LedgerIds []string
}

// QueryLedgersOptions contains Query Ledgers optional parameters.
type QueryLedgersOptions struct {
	// Whether or not to include trades related to position in output.
	// Defaults to false.
	Trades bool
}

// QueryLedgersResponse contains Query Ledgers response data.
type QueryLedgersResponse struct {
	KrakenAPIResponse
	// Key are ledger entry IDs and values are ledger entries
	Result map[string]LedgerEntry `json:"result"`
}

// FeeTierInfo contains fee tier information.
type FeeTierInfo struct {
	// Current fee in percent
	Fee string `json:"fee"`
	// Minimum fee for pair if not fixed fee
	MinimumFee string `json:"min_fee"`
	// Maximum fee for pair if not fixed fee
	MaximumFee string `json:"max_fee"`
	// Next tier's fee for pair if not fixed fee, empty if at lowest fee tier
	NextFee string `json:"next_fee,omitempty"`
	// Volume level of current tier (if not fixed fee. empty if at lowest fee tier)
	TierVolume string `json:"tier_volume,omitempty"`
	// Volume level of next tier (if not fixed fee. enpty if at lowest fee tier)
	NextTierVolume string `json:"next_volume,omitempty"`
}

// GetTradeVolumeParameters contains Get Trade Volume required parameters.
type GetTradeVolumeParameters struct {
	// List of asset pairs to get fee info on
	Pairs []string
}

// GetTradeVolumeOptions contains Get Trade Volume optional parameters.
type GetTradeVolumeOptions struct {
	// Whether or not to include fee info in results
	FeeInfo bool
}

// GetTradeVolume Result
type GetTradeVolumeResult struct {
	// Volume currency
	Currency string `json:"currency"`
	// Current discount volume
	Volume string `json:"volume"`
	// Fee info or Taker fee if asset is submitted to maker/taker fees - each key is an asset pair
	Fees map[string]FeeTierInfo `json:"fees"`
	// Maker fee info - each key is an asset pair
	FeesMaker map[string]FeeTierInfo `json:"fees_maker"`
}

// GetTradeVolumeResponse contains Get Trade Volume response data.
type GetTradeVolumeResponse struct {
	KrakenAPIResponse
	Result GetTradeVolumeResult `json:"result"`
}

// RequestExportReportParameters contains Request Export Report required parameters.
type RequestExportReportParameters struct {
	// Type of data to export
	// Values: "trades", "ledgers"
	Report string
	// Description for the export
	Description string
}

// RequestExportReportOptions contains Request Export Report optional parameters.
type RequestExportReportOptions struct {
	// File format to export.
	// Defaults to "CSV".
	// Values: "CSV" "TSV"
	Format string
	// List of fields to include.
	// Defaults to all
	// Values for trades: ordertxid, time, ordertype, price, cost, fee, vol, margin, misc, ledgers
	// Values for ledgers: refid, time, type, aclass, asset, amount, fee, balance
	Fields []string
	// UNIX timestamp for report start time.
	// Default 1st of the current month
	StartTm *time.Time
	// UNIX timestamp for report end time.
	// Default: now
	EndTm *time.Time
}

// RequestExportReport Result
type RequestExportReportResult struct {
	// Request ID
	Id string `json:"id"`
}

// RequestExportReportResponse contains Request Export Report response data.
type RequestExportReportResponse struct {
	KrakenAPIResponse
	Result RequestExportReportResult `json:"result"`
}

// ExportReportStatus contains export report status data.
type ExportReportStatus struct {
	// Report ID
	Id string `json:"id"`
	// Description
	Description string `json:"descr"`
	// Format
	Format string `json:"format"`
	// Report
	Report string `json:"report"`
	// Subtype
	SubType string `json:"subtype"`
	// Status of report. Enum: "Queued" "Processing" "Processed"
	Status string `json:"status"`
	// Fields
	Fields string `json:"fields"`
	// UNIX timestamp of report request
	RequestTimestamp string `json:"createdtm"`
	// UNIX timestamp report processing began
	StartTimestamp string `json:"starttm"`
	// UNIX timestamp report processing finished
	CompletedTimestamp string `json:"completedtm"`
	// UNIX timestamp of the report data start time
	DataStartTimestamp string `json:"datastarttm"`
	// UNIX timestamp of the report data end time
	DataEndTimestamp string `json:"dataendtm"`
	// Asset
	Asset string `json:"asset"`
}

// GetExportReportStatusParameters contains Get Export Report Status required parameters.
type GetExportReportStatusParameters struct {
	// Type of reports to inquire about
	// Values: "trades" "ledgers"
	Report string
}

// GetExportReportStatusResponse contains Get Export Report Status response data.
type GetExportReportStatusResponse struct {
	KrakenAPIResponse
	// Export Report Statuses
	Result []ExportReportStatus `json:"result"`
}

// RetrieveDataExportParameters contains Retrieve Data Export required parameters.
type RetrieveDataExportParameters struct {
	// Report ID to retrieve
	Id string
}

// RetrieveDataExportResponse contains Retrieve Data Export response data.
type RetrieveDataExportResponse struct {
	// Binary zip archive containing the report
	Report []byte
}

// DeleteExportReportParameters contains Delete Data Export required parameters.
type DeleteExportReportParameters struct {
	// Report ID to delete or cancel
	Id string
	// Type of deletion.
	// delete can only be used for reports that have already been processed. Use cancel for queued or processing reports.
	// Values: "delete" "cancel"
	Type string
}

// DeleteExportReport Result
type DeleteExportReportResult struct {
	// Whether deletion was successful
	Delete bool `json:"delete"`
	// Whether cancellation was successful
	Cancel bool `json:"cancel"`
}

// Delete Export Report Response
type DeleteExportReportResponse struct {
	KrakenAPIResponse
	Result DeleteExportReportResult `json:"result"`
}

/*****************************************************************************/
/*	PRIVATE : USER TRADING                                                   */
/*****************************************************************************/

// Order description info
type OrderDescription struct {
	// Order description
	Order string `json:"order"`
	// Conditional close order description. Empty if not applicable
	Close string `json:"close"`
}

// Conditional close orders are triggered by execution of the primary order in the same quantity
// and opposite direction, but once triggered are independent orders that may reduce or increase net position.
type CloseOrder struct {
	// Close order type.
	// Valid types are "limit", "stop-loss", "take-profit", "stop-loss-limit", "take-profit-limit"
	OrderType string `json:"ordertype"`
	// Price for limit orders or trigger price for stop-loss(-limit) and take-profit(-limit) orders.
	// Price can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	Price string `json:"price"`
	// Limit price for stop-loss-limit and take-profit-limit orders.
	// Price2 can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	// Price2 is ignored if an empty value is provided.
	Price2 string `json:"price2,omitempty"`
}

// Order data
type Order struct {
	// userref is an optional user-specified integer id that can be associated with any number of orders.
	// Will be ignored if a nil value is provided.
	UserReference *int64 `json:"userref,omitempty"`
	// Order type
	OrderType string `json:"ordertype"`
	// Order direction - buy/sell
	Type string `json:"type"`
	// Order quantity in terms of the base asset.
	// "0" can be provided for closing margin orders to automatically fill the requisite quantity.
	Volume string `json:"volume"`
	// Price for limit orders or trigger price for stop-loss(-limit) and take-profit(-limit) orders.
	// Price can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	// Price is ignored if an empty value is provided.
	Price string `json:"price,omitempty"`
	// Limit price for stop-loss-limit and take-profit-limit orders.
	// Price2 can be preceded by +, -, or # to specify the order price as an offset relative to the last traded price.
	// + adds the amount to, and - subtracts the amount from the last traded price.
	// # will either add or subtract the amount to the last traded price depending on the direction and order type used.
	// Relative prices can be suffixed with a % to signify the relative amount as a percentage.
	// Price2 is ignored if an empty value is provided.
	Price2 string `json:"price2,omitempty"`
	// Price signal used to trigger stop and take orders.
	// Will be ignored if an empty value is provided.
	// Default behavior if apply is "last"
	Trigger string `json:"trigger,omitempty"`
	// Amount of leverage desired expressed in a formated string "<leverage>:1".
	// Will be ignored if empty.
	Leverage string `json:"leverage,omitempty"`
	// If true, order will only reduce a currently open position, not increase it or open a new position.
	ReduceOnly bool `json:"reduce_only"`
	// Self trade prevention flag.
	// Will be ignored if an empty value is provided.
	// By default cancel-newest behavior is used.
	StpType string `json:"stp_type,omitempty"`
	// Comma delimited list of order flags.
	// Will be ignored if an empty value is provided.
	OrderFlags string `json:"oflags,omitempty"`
	// Time in force flag.
	// Will be ignored if an empty value is provided.
	// An empty value means default Good Til Canceled behavior.
	TimeInForce string `json:"timeinforce,omitempty"`
	// Scheduled start time. Can be empty to trigger default behavior.
	// A value of 0 means now. Default behavior.
	// A value prefixed with + like +<n> schedules start time n seconds from now.
	// Other values are considered as an absolute unix timestamp for start time.
	ScheduledStartTime string `json:"starttm,omitempty"`
	// Expiration time. Can be empty to trigger default behavior
	// A value prefixed with + like +<n> schedules expiration time n seconds from now. Minimum +5 seconds.
	// Other values are considered as an absolute unix timestamp for exiration time.
	ExpirationTime string `json:"expiretm,omitempty"`
	// Close order
	// A nil value means no close order
	Close *CloseOrder `json:"close,omitempty"`
}

// AddOrder required parameters
type AddOrderParameters struct {
	// Asset pair related to order
	Pair string
	// Order data
	Order Order
}

// AddOrder optional parameters
type AddOrderOptions struct {
	// Validate inputs only. Do not submit order.
	Validate bool
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	// A nil value means no deadline.
	Deadline *time.Time
}

// AddOrderBatch required parameters
type AddOrderBatchParameters struct {
	// Asset pair related to orders
	Pair string
	// List of orders
	Orders []Order
}

// AddOrderBatch optional parameters
type AddOrderBatchOptions struct {
	// Validate inputs only. Do not submit order.
	Validate bool
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	// A nil value means no deadline.
	Deadline *time.Time
}

// AddOrder Result
type AddOrderResult struct {
	// Order description
	Description OrderDescription `json:"descr"`
	// Transaction IDs for order
	TransactionIDs []string `json:"txid"`
}

// Response for Add Order
type AddOrderResponse struct {
	KrakenAPIResponse
	Result AddOrderResult `json:"result"`
}

// AddOrderBatch Entry
type AddOrderBatchEntry struct {
	AddOrderResult
	// Error messages for the order
	Error []string `json:"error"`
}

// AddOrderBatch Result
type AddOrderBatchResult struct {
	// Entries
	Orders []AddOrderBatchEntry `json:"orders"`
}

// Custom parser for AddOrderBatchResult
func (r *AddOrderBatchResult) UnmarshalJSON(data []byte) error {

	// Skip if empty
	if data != nil || string(data) != "" {

		// Temp structure used for lazy evaluation
		type Temp struct {
			Orders []struct {
				// Order description
				Description OrderDescription `json:"descr"`
				// Transaction IDs for order
				TransactionIDs interface{} `json:"txid"`
				// Error message for the order
				Error interface{} `json:"error"`
			} `json:"orders"`
		}

		// Unmarshal in container structure
		container := &Temp{}
		err := json.Unmarshal(data, container)
		if err != nil {
			return err
		}

		// Fill data with what has been parsed
		for _, value := range container.Orders {

			// Create entry to append to result
			entry := &AddOrderBatchEntry{
				AddOrderResult: AddOrderResult{
					Description: OrderDescription{
						Order: value.Description.Order,
						Close: value.Description.Close,
					},
					TransactionIDs: []string{},
				},
				Error: []string{},
			}

			// Process transaction IDs if any
			if value.TransactionIDs != nil {
				singleTxId, ok := value.TransactionIDs.(string)
				if ok {
					// Only one transaction ID as string. Add it
					entry.TransactionIDs = append(entry.TransactionIDs, singleTxId)
				} else {
					mulTxIds, ok := value.TransactionIDs.([]interface{})
					if ok {
						for _, id := range mulTxIds {
							// Type assert id
							i, ok := id.(string)
							if !ok {
								return fmt.Errorf("one transaction ID is not a string. Got %v", i)
							}
							entry.TransactionIDs = append(entry.TransactionIDs, i)
						}
					} else {
						// Error
						return fmt.Errorf("unexpected order transaction IDs value. Got %v", value.TransactionIDs)
					}
				}
			}

			// Process errors if any
			if value.Error != nil {
				singleErr, ok := value.Error.(string)
				if ok {
					// Only one error as string. Add it
					entry.Error = append(entry.Error, singleErr)
				} else {
					mulErrs, ok := value.Error.([]interface{})
					if ok {
						for _, e := range mulErrs {
							// Type assert error
							i, ok := e.(string)
							if !ok {
								return fmt.Errorf("one transaction ID is not a string. Got %v", i)
							}
							entry.TransactionIDs = append(entry.TransactionIDs, i)
						}
					} else {
						// Error
						return fmt.Errorf("unexpected order transaction errors value. Got %v", value.Error)
					}
				}
			}

			// Add entry to result
			r.Orders = append(r.Orders, *entry)
		}
	}
	return nil
}

// Response for Add Order Batch
type AddOrderBatchResponse struct {
	KrakenAPIResponse
	Result AddOrderBatchResult `json:"result"`
}

// EditOrder required parameters
type EditOrderParameters struct {
	// Original Order ID or User Reference Id (userref) which is user-specified
	// integer id used with the original order. If userref is not unique and was
	// used with multiple order, edit request is denied with an error.
	Id string
	// Asset Pair
	Pair string
}

// EditOrder optional parameters
type EditOrderOptions struct {
	// New user reference id. Userref from parent order will
	// not be retained on the new order after edit.
	// An empty value means that data must not be changed.
	NewUserReference string
	// Order quantity in terms of the base asset.
	// A nil value means that data must not be changed.
	NewVolume string
	// New limit price or trigger price. Either price or price2
	// can be preceded by +, -, or # to specify the order price
	// as an offset relative to the last traded price. + adds
	// the amount to, and - subtracts the amount from the last
	// traded price. # will either add or subtract the amount to
	// the last traded price, depending on the direction and order
	// type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	// An empty value means that data must not be changed.
	Price string
	// New limit price for stop/take-limit order. Either price or price2
	// can be preceded by +, -, or # to specify the order price
	// as an offset relative to the last traded price. + adds
	// the amount to, and - subtracts the amount from the last
	// traded price. # will either add or subtract the amount to
	// the last traded price, depending on the direction and order
	// type used. Relative prices can be suffixed with a % to
	// signify the relative amount as a percentage.
	// An empty value means that data must not be changed.
	Price2 string
	// List of order flags. Only these flags can be
	// changed: - post post-only order (available when ordertype =
	// limit). All the flags from the parent order are retained except
	// post-only. post-only needs to be explicitly mentioned on edit request.
	// A nil value means that data must not be changed.
	OFlags []string
	// Validate inputs only. Do not submit order.
	Validate bool
	// Used to interpret if client wants to receive pending replace,
	// before the order is completely replaced
	CancelResponse bool
	// RFC3339 timestamp (e.g. 2021-04-01T00:18:45Z) after which the matching
	// engine should reject  the new order request, in presence of latency or
	// order queueing. min now() + 2 seconds, max now() + 60 seconds.
	// A nil value means no deadline.
	Deadline *time.Time
}

// EditOrder Result
type EditOrderResult struct {
	// Order description
	Description OrderDescription `json:"descr"`
	// New transaction ID
	TransactionID string `json:"txid"`
	// New user reference.
	// Will be nil if no user ref was provided with request.
	NewUserReference *int64 `json:"newuserref"`
	// Old user reference
	// Will be nil if no user ref was provided with request to create the original order.
	OldUserReference *int64 `json:"olduserref"`
	// Number of orders canceled
	OrdersCancelled int `json:"orders_cancelled"`
	// Original transaction ID
	OriginalTransactionID string `json:"originaltxid"`
	// Status of the order. Either "ok" or "err"
	Status string `json:"status"`
	// Updated volume
	Volume string `json:"volume"`
	// Updated Price
	Price string `json:"price"`
	// Updated Price2
	Price2 string `json:"price2"`
	// Error message if unsuccessful
	ErrorMsg string `json:"error_message"`
}

// EditOrder Response
type EditOrderResponse struct {
	KrakenAPIResponse
	Result EditOrderResult `json:"result"`
}

// Cancel Order required parameters
type CancelOrderParameters struct {
	// Open order transaction ID (txid) or user reference (userref)
	Id string
}

// CancelOrder Result
type CancelOrderResult struct {
	// Number of canceled orders
	Count int `json:"count"`
	// If set, order(s) is/are pending cancellation
	Pending bool `json:"pending"`
}

// Response for Cancel Order
type CancelOrderResponse struct {
	KrakenAPIResponse
	Result CancelOrderResult `json:"result"`
}

// Response for Cancel All Orders
type CancelAllOrdersResponse struct {
	KrakenAPIResponse
	Result struct {
		// Number of canceled orders
		Count int `json:"count"`
	} `json:"result"`
}

// CancelAllOrdersAfterXParameters
type CancelCancelAllOrdersAfterXParameters struct {
	// Duration (in seconds) to set/extend the timer by
	Timeout int64
}

// Response for Cancel All Orders After X
type CancelAllOrdersAfterXResponse struct {
	KrakenAPIResponse
	Result struct {
		// Timestamp (RFC3339 format) at which the request was received
		CurrentTime time.Time `json:"currentTime"`
		// Timestamp (RFC3339 format) after which all orders will be cancelled, unless the timer is extended or disabled
		TriggerTime time.Time `json:"triggerTime"`
	} `json:"result"`
}

// CancelOrderBatch required parameters
type CancelOrderBatchParameters struct {
	// Open orders transaction ID (txid) or user reference (userref)
	OrderIds []string
}

// Response for Cancel Order Batch
type CancelOrderBatchResponse struct {
	KrakenAPIResponse
	Result struct {
		// Number of canceled orders
		Count int `json:"count"`
	} `json:"result"`
}

/*****************************************************************************/
/*	PRIVATE : USER FUNDING RESPONSES                                         */
/*****************************************************************************/

// Data of a deposit method
type DepositMethod struct {
	// Name of deposit method
	Method string `json:"method"`
	// Maximum net amount that can be deposited right now, or empty or "false" if no limit
	Limit string `json:"limit"`
	// Amount of fees that will be paid, can be empty if no fee
	Fee string `json:"fee"`
	// Whether or not method has an address setup fee. Can be empty
	AddressSetupFee string `json:"address-setup-fee"`
	// Whether new addresses can be generated for this method, Can be empty
	GenAddress string `json:"gen-address"`
}

// Custom JSON parser for DepositMethod
func (d *DepositMethod) UnmarshalJSON(data []byte) error {

	// Temp struct with interface for Limit for lazy evaluation
	type Temp struct {
		// Name of deposit method
		Method string `json:"method"`
		// Maximum net amount that can be deposited right now, or empty or "false" if no limit
		Limit interface{} `json:"limit"`
		// Amount of fees that will be paid, can be empty if no fee
		Fee string `json:"fee"`
		// Whether or not method has an address setup fee. Can be empty
		AddressSetupFee string `json:"address-setup-fee"`
		// Whether new addresses can be generated for this method, Can be empty
		GenAddress interface{} `json:"gen-address"`
	}

	container := &Temp{}

	// Unmarshall in safe container
	err := json.Unmarshal(data, container)
	if err != nil {
		return err
	}

	// Populate result
	d.AddressSetupFee = container.AddressSetupFee
	d.Fee = container.Fee
	if container.GenAddress != nil {
		d.GenAddress = fmt.Sprintf("%v", container.GenAddress)
	}
	if container.Limit != nil {
		d.Limit = fmt.Sprintf("%v", container.Limit)
	}
	d.Method = container.Method

	return nil
}

// GetDepositMethods required parameters
type GetDepositMethodsParameters struct {
	// Asset being deposited
	Asset string
}

// Response returned by Get Deposit Methods
type GetDepositMethodsResponse struct {
	KrakenAPIResponse
	Result []DepositMethod `json:"result"`
}

// Data of a deposit address
type DepositAddress struct {
	// Deposit Address
	Address string `json:"address"`
	// Expiration time in unix timestamp, or 0 if not expiring
	Expiretm int64 `json:"expiretm,string"`
	// Whether or not address has ever been used
	New bool `json:"new"`
}

// GetDepositAddresses required parameters
type GetDepositAddressesParameters struct {
	// Asset being deposited
	Asset string
	// Name of the deposit method
	Method string
}

// GetDepositAddresses optional parameters
type GetDepositAddressesOptions struct {
	// Whether or not to generate a new address.
	// Defaults to false.
	New bool
}

// Get Deposit Addresses response
type GetDepositAddressesResponse struct {
	KrakenAPIResponse
	Result []DepositAddress `json:"result"`
}

// Transaction details for a deposit or a withdrawal
type TransactionDetails struct {
	// Name of deposit method
	Method string `json:"method"`
	// Asset class
	AssetClass string `json:"aclass"`
	// Asset
	Asset string `json:"asset"`
	// Reference ID
	ReferenceID string `json:"refid"`
	// Method transaction ID
	TransactionID string `json:"txid"`
	// Method transaction information
	Info string `json:"info"`
	// Amount deposited/withdrawn
	Amount string `json:"amount"`
	// Fees paid. Can be empty
	Fee string `json:"fee"`
	// Unix timestamp when request was made
	Time int64 `json:"time"`
	// Status of deposit - IFEX financial transaction states
	Status string `json:"status"`
	// Additional status property. Can be empty
	StatusProperty string `json:"status-prop,omitempty"`
}

// GetStatusOfRecentDeposits required parameters
type GetStatusOfRecentDepositsParameters struct {
	// Asset being deposited
	Asset string
}

// GetStatusOfRecentDeposits optional parameters
type GetStatusOfRecentDepositsOptions struct {
	// Name of the deposit method
	Method string
}

// Get Status of Recent Deposits response
type GetStatusOfRecentDepositsResponse struct {
	// Recent deposits
	Result []TransactionDetails `json:"result"`
}

// GetWithdrawalInformationrequired parameters
type GetWithdrawalInformationParameters struct {
	// Asset being withdrawn
	Asset string
	// Withdrawal address name as setup on account
	Key string
	// Anount to be withdrawn
	Amount string
}

// Get Withdrawal Information response
type GetWithdrawalInformationResponse struct {
	KrakenAPIResponse
	Result GetWithdrawalInformationResult `json:"result"`
}

// GetWithdrawalInformation Result
type GetWithdrawalInformationResult struct {
	// Name of the withdrawal method that will be used
	Method string `json:"method"`
	// Maximum net amount that can be withdrawn right now
	Limit string `json:"limit"`
	// Net amount that will be sent, after fees
	Amount string `json:"amount"`
	// Amount of fees that will be paid
	Fee string `json:"fee"`
}

// WithdrawFunds required parameters
type WithdrawFundsParameters struct {
	// Asset being withdrawn
	Asset string
	// Withdrawal address name as setup on account
	Key string
	// Anount to be withdrawn
	Amount string
}

// Withdraw Funds responses
type WithdrawFundsResponse struct {
	KrakenAPIResponse
	Result struct {
		// Reference ID
		ReferenceID string `json:"refid"`
	} `json:"result"`
}

// GetStatusOfRecentWithdrawals required parameters
type GetStatusOfRecentWithdrawalsParameters struct {
	// Asset being deposited
	Asset string
}

// GetStatusOfRecentWithdrawals optional parameters
type GetStatusOfRecentWithdrawalsOptions struct {
	// Name of the deposit method
	Method string
}

// Get Status of Recent Withdrawals response
type GetStatusOfRecentWithdrawalsResponse struct {
	KrakenAPIResponse
	// Recent withdrawals
	Result []TransactionDetails `json:"result"`
}

// RequestWithdrawalCancellation required parameters
type RequestWithdrawalCancellationParameters struct {
	// Asset being withdrawn
	Asset string
	// Withdrawal reference ID
	ReferenceId string
}

// Request Withdrawal Cancellation Response
type RequestWithdrawalCancellationResponse struct {
	KrakenAPIResponse
	Result bool `json:"result"`
}

// RequestWalletTransfer required parameters
type RequestWalletTransferParameters struct {
	// Asset being transfered
	Asset string
	// Source wallet
	From string
	// Destination wallet
	To string
	// Amount to be transfered
	Amount string
}

// Request Wallet Transfer response
type RequestWalletTransferResponse struct {
	KrakenAPIResponse
	Result struct {
		// Reference ID
		ReferenceID string `json:"refid"`
	} `json:"result"`
}

/*****************************************************************************/
/*	PRIVATE : USER STAKING RESPONSES                                         */
/*****************************************************************************/

// Stake Asset parameters
type StakeAssetParameters struct {
	// Asset to stake (asset ID or altname)
	Asset string
	// Amount of the asset to stake
	Amount string
	// Name of the staking option to use (refer to the Staking Assets endpoint for the correct method names for each asset)
	Method string
}

// Stake Asset Response
type StakeAssetResponse struct {
	KrakenAPIResponse
	Result struct {
		// Reference ID
		ReferenceID string `json:"refid"`
	} `json:"result"`
}

// Unstake Asset parameters
type UnstakeAssetParameters struct {
	// Asset to stake (asset ID or altname)
	Asset string
	// Amount of the asset to stake
	Amount string
}

// Unstake Asset Response
type UnstakeAssetResponse struct {
	KrakenAPIResponse
	Result struct {
		// Reference ID
		ReferenceID string `json:"refid"`
	} `json:"result"`
}

// Staking Asset Minimal Amount
type StakingAssetMinAmount struct {
	Unstaking string `json:"unstaking"`
	Staking   string `json:"staking"`
}

// UnmarshalJSON is a custom JSON unmarshaller for StakingAssetMinAmount. This
// is necessary because server response defines for many fields default values
// that are different from Go zero values.
func (s *StakingAssetMinAmount) UnmarshalJSON(data []byte) error {

	// Unmarshal in map to see which fields are missing
	var v map[string]string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	// Set min amount with default values
	s.Unstaking = "0"
	s.Staking = "0"
	if v["staking"] != "" {
		s.Staking = v["staking"]
	}
	if v["unstaking"] != "" {
		s.Unstaking = v["unstaking"]
	}

	return nil
}

// StakingAssetLockPeriod describe a locking period for a stacking operation.
type StakingAssetLockPeriod struct {
	// Days the funds are locked.
	Days float64 `json:"days"`
	// Percentage of the funds that are locked (0 - 100)
	Percentage float64 `json:"percentage"`
}

// StakingAssetLockup describes asset lockup periods for staking operations
type StakingAssetLockup struct {
	// Optional lockup period for unstaking operation.
	Unstaking *StakingAssetLockPeriod `json:"unstaking,omitempty"`
	// Optional lockup period for staking operation.
	Staking *StakingAssetLockPeriod `json:"staking,omitempty"`
	// Optional lockup period.
	Lockup *StakingAssetLockPeriod `json:"lockup,omitempty"`
}

// StakingAssetReward describes the rewards earned while staking
type StakingAssetReward struct {
	// Reward earned while staking
	Reward string `json:"reward"`
	// Reward type. Value : percentage
	Type string `json:"type"`
}

// Staking Asset Information
type StakingAssetInformation struct {
	// Asset code
	Asset string `json:"asset"`
	// Staking asset code
	StakingAsset string `json:"staking_asset"`
	// Unique ID of the staking option (used in Stake/Unstake operations)
	Method string `json:"method"`
	// Whether the staking operation is on-chain or not.
	OnChain bool `json:"on_chain"`
	// Whether the user will be able to stake this asset.
	CanStake bool `json:"can_stake"`
	// Whether the user will be able to unstake this asset.
	CanUnstake bool `json:"can_unstake"`
	// Optional minimium amount for staking/unstaking.
	MinAmount *StakingAssetMinAmount `json:"minimum_amount"`
	// Optional field which describes the locking periods and percentages for staking operations.
	Lock *StakingAssetLockup `json:"lock"`
	// Enabled for user
	EnabledForUser bool `json:"enabled_for_user"`
	// Disabled - default true
	Disabled bool `json:"disabled"`
	// Describes the rewards earned while staking
	Rewards StakingAssetReward `json:"rewards"`
}

// UnmarshalJSON is a custom JSON unmarshaller for StakingAssetInformation. This
// is necessary because server response defines for many fields default values
// that are different from Go zero values.
func (s *StakingAssetInformation) UnmarshalJSON(data []byte) error {

	// Use a different structure to handle optional bools with
	// true as default values.
	var v struct {
		// Asset code
		Asset string `json:"asset"`
		// Staking asset code
		StakingAsset string `json:"staking_asset"`
		// Unique ID of the staking option (used in Stake/Unstake operations)
		Method string `json:"method"`
		// Whether the staking operation is on-chain or not.
		OnChain *bool `json:"on_chain"`
		// Whether the user will be able to stake this asset.
		CanStake *bool `json:"can_stake"`
		// Whether the user will be able to unstake this asset.
		CanUnstake *bool `json:"can_unstake"`
		// Optional minimium amount for staking/unstaking.
		MinAmount *StakingAssetMinAmount `json:"minimum_amount"`
		// Optional field which describes the locking periods and percentages for staking operations.
		Lock *StakingAssetLockup `json:"lock"`
		// Enabled for user
		EnabledForUser *bool `json:"enabled_for_user"`
		// Disabled - default true
		Disabled *bool `json:"disabled"`
		// Describes the rewards earned while staking
		Rewards StakingAssetReward `json:"rewards"`
	}

	// Unmarshal data using the intermediate struct
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	// Set values in destination with good defaults
	s.Asset = v.Asset
	s.StakingAsset = v.StakingAsset
	s.Method = v.Method
	if v.OnChain != nil {
		s.OnChain = *v.OnChain
	} else {
		s.OnChain = true
	}
	if v.CanStake != nil {
		s.CanStake = *v.CanStake
	} else {
		s.CanStake = true
	}
	if v.CanUnstake != nil {
		s.CanUnstake = *v.CanUnstake
	} else {
		s.CanUnstake = true
	}
	if v.EnabledForUser != nil {
		s.EnabledForUser = *v.EnabledForUser
	} else {
		s.EnabledForUser = true
	}
	if v.Disabled != nil {
		s.Disabled = *v.Disabled
	}
	s.MinAmount = v.MinAmount
	s.Lock = v.Lock
	s.Rewards = v.Rewards

	return nil
}

// List Of Stakeable Assets Response
type ListOfStakeableAssetsResponse struct {
	// List of satekeable assets
	Result []StakingAssetInformation `json:"result"`
}

// Staking Transaction Info
type StakingTransactionInfo struct {
	// Reference ID for transaction
	ReferenceId string `json:"refid"`
	// Asset code
	Asset string `json:"asset"`
	// Asset class
	AssetClass string `json:"aclass"`
	// Type of transaction. Enum: "bonding" "reward" "unbonding"
	Type string `json:"type"`
	// Method
	Method string `json:"method"`
	// Amount
	Amount string `json:"amount"`
	// Fee
	Fee string `json:"fee"`
	// Timestamp
	Timestamp int64 `json:"time"`
	// Status
	Status string `json:"status"`
	// Optional unix timestamp from the start of bond period of bonding transactions.
	BondStart *int64 `json:"bond_start"`
	// Optional unix timestamp from the end of bond period of bonding transactions.
	BondEnd *int64 `json:"bond_end"`
}

// Get Pending Staking Transactions Response
type GetPendingStakingTransactionsResponse struct {
	// Pending staking transactions.
	Result []StakingTransactionInfo `json:"result"`
}

// List Of Staking Transactions Response
type ListOfStakingTransactionsResponse struct {
	// Staking transactions.
	Result []StakingTransactionInfo `json:"result"`
}
