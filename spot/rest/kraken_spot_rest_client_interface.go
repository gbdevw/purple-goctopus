package spotex

import (
	"context"
	"encoding/json"
)

/*************************************************************************************************/
/* INTERFACE                                                                                     */
/*************************************************************************************************/

// Interface for Kraken Spot REST client
type KrakenSpotRESTClientIface interface {
	// GetServerTime Get the server time.
	GetServerTime(ctx context.Context) (*models.GetServerTimeResponse, error)
	// GetSystemStatus Get the current system status or trading mode.
	GetSystemStatus(ctx context.Context) (*GetSystemStatusResponse, error)
	// GetAssetInfo Get information about the assets that are available for deposit, withdrawal, trading and staking.
	GetAssetInfo(ctx context.Context, opts *GetAssetInfoOptions) (*GetAssetInfoResponse, error)
	// GetTradableAssetPairs Get tradable asset pairs.
	GetTradableAssetPairs(ctx context.Context, opts *GetTradableAssetPairsOptions) (*GetTradableAssetPairsResponse, error)
	// Get ticker information about a given list of pairs.
	// Note: Today's prices start at midnight UTC
	GetTickerInformation(ctx context.Context, opts *GetTickerInformationOptions) (*GetTickerInformationResponse, error)
	// GetOHLCData get Open, High, Low & Close indicators.
	// Note: the last entry in the OHLC array is for the current, not-yet-committed
	// frame and will always be present, regardless of the value of since.
	GetOHLCData(ctx context.Context, params GetOHLCDataParameters, opts *GetOHLCDataOptions) (*GetOHLCDataResponse, error)
	// GetOrderBook Get order by for a given pair
	GetOrderBook(ctx context.Context, params GetOrderBookParameters, opts *GetOrderBookOptions) (*GetOrderBookResponse, error)
	// GetRecentTrades Get up to the 1000 most recent trades by default
	GetRecentTrades(ctx context.Context, params GetRecentTradesParameters, opts *GetRecentTradesOptions) (*GetRecentTradesResponse, error)
	// GetRecentSpreads Get recent spreads.
	GetRecentSpreads(ctx context.Context, params GetRecentSpreadsParameters, opts *GetRecentSpreadsOptions) (*GetRecentSpreadsResponse, error)
	// GetAccountBalance - Retrieve all cash balances, net of pending withdrawals.
	GetAccountBalance(ctx context.Context, secopts *SecurityOptions) (*GetAccountBalanceResponse, error)
	// GetTradeBalance - Retrieve a summary of collateral balances, margin position valuations, equity and margin level.
	GetTradeBalance(ctx context.Context, opts *GetTradeBalanceOptions, secopts *SecurityOptions) (*GetTradeBalanceResponse, error)
	// GetOpenOrders - Retrieve information about currently open orders.
	GetOpenOrders(ctx context.Context, opts *GetOpenOrdersOptions, secopts *SecurityOptions) (*GetOpenOrdersResponse, error)
	// GetClosedOrders -
	// Retrieve information about orders that have been closed (filled or cancelled).
	// 50 results are returned at a time, the most recent by default.
	GetClosedOrders(ctx context.Context, opts *GetClosedOrdersOptions, secopts *SecurityOptions) (*GetClosedOrdersResponse, error)
	// QueryOrdersInfo - Retrieve information about specific orders.
	QueryOrdersInfo(ctx context.Context, params QueryOrdersParameters, opts *QueryOrdersOptions, secopts *SecurityOptions) (*QueryOrdersInfoResponse, error)
	// GetTradesHistory -
	// Retrieve information about trades/fills.
	// 50 results are returned at a time, the most recent by default.
	//
	// Unless otherwise stated, costs, fees, prices, and volumes are specified with the precision for the asset pair
	// (pair_decimals and lot_decimals), not the individual assets' precision (decimals).
	GetTradesHistory(ctx context.Context, opts *GetTradesHistoryOptions, secopts *SecurityOptions) (*GetTradesHistoryResponse, error)
	// QueryTradesInfo - Retrieve information about specific trades/fills.
	QueryTradesInfo(ctx context.Context, params QueryTradesParameters, opts *QueryTradesOptions, secopts *SecurityOptions) (*QueryTradesInfoResponse, error)
	// GetOpenPositions - Get information about open margin positions.
	GetOpenPositions(ctx context.Context, opts *GetOpenPositionsOptions, secopts *SecurityOptions) (*GetOpenPositionsResponse, error)
	// GetLedgersInfo - Retrieve information about ledger entries. 50 results are returned at a time, the most recent by default.
	GetLedgersInfo(ctx context.Context, opts *GetLedgersInfoOptions, secopts *SecurityOptions) (*GetLedgersInfoResponse, error)
	// QueryLedgers - Retrieve information about specific ledger entries.
	QueryLedgers(ctx context.Context, params QueryLedgersParameters, opts *QueryLedgersOptions, secopts *SecurityOptions) (*QueryLedgersResponse, error)
	// RequestExportReport - Request export of trades or ledgers.
	RequestExportReport(ctx context.Context, params RequestExportReportParameters, opts *RequestExportReportOptions, secopts *SecurityOptions) (*RequestExportReportResponse, error)
	// GetExportReportStatus - Get status of requested data exports.
	GetExportReportStatus(ctx context.Context, params GetExportReportStatusParameters, secopts *SecurityOptions) (*GetExportReportStatusResponse, error)
	// RetrieveDataExport Get report as a zip
	RetrieveDataExport(ctx context.Context, params RetrieveDataExportParameters, secopts *SecurityOptions) (*RetrieveDataExportResponse, error)
	// DeleteExportReport - Delete exported trades/ledgers report
	DeleteExportReport(ctx context.Context, params DeleteExportReportParameters, secopts *SecurityOptions) (*DeleteExportReportResponse, error)
	// Place a new order.
	AddOrder(ctx context.Context, params AddOrderParameters, opts *AddOrderOptions, secopts *SecurityOptions) (*AddOrderResponse, error)
	// Send an array of orders (max: 15). Any orders rejected due to order validations, will be dropped
	// and the rest of the batch is processed. All orders in batch should be limited to a single pair.
	// The order of returned txid's in the response array is the same as the order of the order list sent in request.
	AddOrderBatch(ctx context.Context, params AddOrderBatchParameters, opts *AddOrderBatchOptions, secopts *SecurityOptions) (*AddOrderBatchResponse, error)
	// Edit volume and price on open orders. Uneditable orders include margin orders, triggered stop/profit orders,
	// orders with conditional close terms attached, those already cancelled or filled, and those where the executed
	// volume is greater than the newly supplied volume. post-only flag is not retained from original order after
	// successful edit. post-only needs to be explicitly set on edit request.
	EditOrder(ctx context.Context, params EditOrderParameters, opts *EditOrderOptions, secopts *SecurityOptions) (*EditOrderResponse, error)
	// Cancel a particular open order (or set of open orders) by txid or userref
	CancelOrder(ctx context.Context, params CancelOrderParameters, secopts *SecurityOptions) (*CancelOrderResponse, error)
	// Cancel all open orders
	CancelAllOrders(ctx context.Context, secopts *SecurityOptions) (*CancelAllOrdersResponse, error)
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
	CancelAllOrdersAfterX(ctx context.Context, params CancelCancelAllOrdersAfterXParameters, secopts *SecurityOptions) (*CancelAllOrdersAfterXResponse, error)
	// Cancel multiple open orders by txid or userref
	CancelOrderBatch(ctx context.Context, params CancelOrderBatchParameters, secopts *SecurityOptions) (*CancelOrderBatchResponse, error)
	// Retrieve methods available for depositing a particular asset.
	GetDepositMethods(ctx context.Context, params GetDepositMethodsParameters, secopts *SecurityOptions) (*GetDepositMethodsResponse, error)
	// Retrieve (or generate a new) deposit addresses for a particular asset and method.
	GetDepositAddresses(ctx context.Context, params GetDepositAddressesParameters, opts *GetDepositAddressesOptions, secopts *SecurityOptions) (*GetDepositAddressesResponse, error)
	// Retrieve information about recent deposits made.
	GetStatusOfRecentDeposits(ctx context.Context, params GetStatusOfRecentDepositsParameters, opts *GetStatusOfRecentDepositsOptions, secopts *SecurityOptions) (*GetStatusOfRecentDepositsResponse, error)
	// Retrieve fee information about potential withdrawals for a particular asset, key and amount.
	GetWithdrawalInformation(ctx context.Context, params GetWithdrawalInformationParameters, secopts *SecurityOptions) (*GetWithdrawalInformationResponse, error)
	// Make a withdrawal request.
	WithdrawFunds(ctx context.Context, params WithdrawFundsParameters, secopts *SecurityOptions) (*WithdrawFundsResponse, error)
	// Retrieve information about recently requests withdrawals.
	GetStatusOfRecentWithdrawals(ctx context.Context, params GetStatusOfRecentWithdrawalsParameters, opts *GetStatusOfRecentWithdrawalsOptions, secopts *SecurityOptions) (*GetStatusOfRecentWithdrawalsResponse, error)
	// Cancel a recently requested withdrawal, if it has not already been successfully processed.
	RequestWithdrawalCancellation(ctx context.Context, params RequestWithdrawalCancellationParameters, secopts *SecurityOptions) (*RequestWithdrawalCancellationResponse, error)
	// Transfer from Kraken spot wallet to Kraken Futures holding wallet. Note that a transfer in the other direction must be requested via the Kraken Futures API endpoint.
	RequestWalletTransfer(ctx context.Context, params RequestWalletTransferParameters, secopts *SecurityOptions) (*RequestWalletTransferResponse, error)
	// StakeAsset stake an asset from spot wallet.
	StakeAsset(ctx context.Context, params StakeAssetParameters, secopts *SecurityOptions) (*StakeAssetResponse, error)
	// UnstakeAsset unstake an asset from your staking wallet.
	UnstakeAsset(ctx context.Context, params UnstakeAssetParameters, secopts *SecurityOptions) (*UnstakeAssetResponse, error)
	// ListOfStakeableAssets returns the list of assets that the user is able to stake.
	ListOfStakeableAssets(ctx context.Context, secopts *SecurityOptions) (*ListOfStakeableAssetsResponse, error)
	// GetPendingStakingTransactions returns the list of pending staking transactions.
	GetPendingStakingTransactions(ctx context.Context, secopts *SecurityOptions) (*GetPendingStakingTransactionsResponse, error)
	// ListOfStakingTransactions returns the list of 1000 recent staking transactions from past 90 days.
	ListOfStakingTransactions(ctx context.Context, secopts *SecurityOptions) (*ListOfStakingTransactionsResponse, error)
}

/*************************************************************************************************/
/* COMMON TYPES                                                                                  */
/*************************************************************************************************/

/*************************************************************************************************/
/* ENUMS                                                                                         */
/*************************************************************************************************/

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
