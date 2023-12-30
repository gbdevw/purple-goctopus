package rest

import (
	"archive/zip"
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gbdevw/purple-goctopus/noncegen"
	"github.com/gbdevw/purple-goctopus/spot/rest/account"
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/earn"
	"github.com/gbdevw/purple-goctopus/spot/rest/funding"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
	"github.com/gbdevw/purple-goctopus/spot/rest/trading"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* INTEGRATION TEST SUITE                                                                        */
/*************************************************************************************************/

// Integration test suite for KrakenSpotRESTClient
type KrakenSpotRESTClientIntegrationTestSuite struct {
	suite.Suite
	// Decorated kraken API client
	client KrakenSpotRESTClientIface
	// Security options to use for private endpoints
	fa2 *common.SecurityOptions
	// Nonce generator
	noncegen noncegen.NonceGenerator
}

// Configure and run unit test suite
func TestKrakenSpotRESTClientIntegrationTestSuite(t *testing.T) {
	// Skip integration tests if short flag is used
	if testing.Short() {
		t.SkipNow()
	}

	// Load credentials for Kraken spot REST API
	// key := os.Getenv("KRAKEN_API_KEY")
	// b64Secret := os.Getenv("KRAKEN_API_SECRET")
	// otp := os.Getenv("KRAKEN_API_OTP")

	key := `ocIEujBuivw2YfBNSYGaDIaHoZlR2p3/Obn4MUgvIaZy0iPRcAYOrLji`
	b64Secret := `szTSP19f5oC463Pt6jWD4zc3D2BzPvG+lleUN3Pfi/v1TSC6KBxtpkP661ZZ3Kb2H5bfEndAvMKH+s33FvnEuw==`
	otp := `ApxP2td2!gbwHmYUt-PaWy*qgvb3VH8W2ag2Mj@ZFoh.omg3W9ECsE@kqLVc7XFsAwHN3426.dnoFLCxAsX8feujNYBCJx@HLg!E`

	// If an OTP is provided, set 2FA
	var fa2 *common.SecurityOptions = nil
	if otp != "" {
		fa2 = &common.SecurityOptions{SecondFactor: otp}
	}

	// Create instrumented authorizer
	auth, err := NewKrakenSpotRESTClientAuthorizer(key, b64Secret)
	require.NoError(t, err)
	require.NotNil(t, auth)
	authorizer := InstrumentKrakenSpotRESTClientAuthorizer(auth, nil)

	// Build and configure a retryable http client
	httpclient := retryablehttp.NewClient()
	httpclient.RetryWaitMax = 1 * time.Second
	httpclient.RetryWaitMin = 1 * time.Second
	httpclient.RetryMax = 3
	httpclient.Logger = log.New(io.Discard, "", 0) // Silent debug logs

	// Create an instrumented Kraken spot REST API client with a retryable http client.
	// The client will use KRaken production environment as target for the tests
	client := InstrumentKrakenSpotRESTClient(
		NewKrakenSpotRESTClient(
			authorizer,
			&KrakenSpotRESTClientConfiguration{
				BaseURL: KrakenProductionV0BaseUrl,
				Client:  httpclient.StandardClient(),
			}),
		nil)

	// Run unit test suite
	suite.Run(t, &KrakenSpotRESTClientIntegrationTestSuite{
		Suite:    suite.Suite{},
		client:   client,
		fa2:      fa2,
		noncegen: noncegen.NewHFNonceGenerator(),
	})
}

/*************************************************************************************************/
/* INTEGRATION TESTS - MARKET DATA                                                               */
/*************************************************************************************************/

// Integration test for GetServerTime.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetServerTimeIntegration() {
	// Call API
	resp, httpresp, err := suite.client.GetServerTime(context.Background())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	// Check response
	require.NotEmpty(suite.T(), resp.Result.Rfc1123)
	require.Greater(suite.T(), resp.Result.Unixtime, int64(0))
}

// Integration test for GetSystemStatus.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetSystemStatusIntegration() {
	// Call API
	resp, httpresp, err := suite.client.GetSystemStatus(context.Background())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	// Check response
	require.NotEmpty(suite.T(), resp.Result.Status)
	require.NotEmpty(suite.T(), resp.Result.Timestamp)
}

// Integration test for GetAssetInfo.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetAssetInfoIntegration() {
	// Call API
	options := &market.GetAssetInfoRequestOptions{
		Assets:     []string{"XXBT", "XETH"},
		AssetClass: "currency",
	}
	resp, httpresp, err := suite.client.GetAssetInfo(context.Background(), options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	// Check response
	require.Len(suite.T(), resp.Result, len(options.Assets))
	for _, asset := range options.Assets {
		require.NotNil(suite.T(), resp.Result[asset])
	}
}

// Integration test for GetTradableAssetPairs.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetTradableAssetPairsIntegration() {
	// Call API
	options := &market.GetTradableAssetPairsRequestOptions{
		Pairs: []string{"XXBTZUSD", "XETHZEUR"},
		Info:  string(market.InfoFees),
	}
	resp, httpresp, err := suite.client.GetTradableAssetPairs(context.Background(), options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	// Check response
	require.Len(suite.T(), resp.Result, len(options.Pairs))
	for _, pair := range options.Pairs {
		require.NotNil(suite.T(), resp.Result[pair])
	}
}

// Integration test for GetTickerInformation.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetTickerInformationIntegration() {
	// Call API
	options := &market.GetTickerInformationRequestOptions{
		Pairs: []string{"XXBTZUSD", "XETHZEUR"},
	}
	resp, httpresp, err := suite.client.GetTickerInformation(context.Background(), options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	// Check response
	require.Len(suite.T(), resp.Result, len(options.Pairs))
	for _, pair := range options.Pairs {
		require.NotNil(suite.T(), resp.Result[pair])
	}
}

// Integration test for GetOHLCData.
//
// Test is OK but htere is a flagging issue -> an error is sometime returned by the API despite the URL is OK and lead
// to a correct response in the browser.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetOHLCDataIntegration() {
	// Call API
	params := market.GetOHLCDataRequestParameters{
		Pair: "XBTUSD",
	}
	options := &market.GetOHLCDataRequestOptions{
		Interval: int64(market.M60),
		Since:    1548111600,
	}
	resp, httpresp, err := suite.client.GetOHLCData(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)

	// Check response
	require.NotEmpty(suite.T(), resp.Result.PairId) // PairId use canonical names not like the ones that must be provided
	require.Greater(suite.T(), resp.Result.Last, int64(0))
	require.NotEmpty(suite.T(), resp.Result.Data)
}

// Integration test for GetOrderBook.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetOrderBookIntegration() {
	// Call API
	params := market.GetOrderBookRequestParameters{
		Pair: "XBTUSD",
	}
	options := &market.GetOrderBookRequestOptions{
		Count: 2,
	}
	resp, httpresp, err := suite.client.GetOrderBook(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)

	// Check response
	require.NotEmpty(suite.T(), resp.Result.PairId)
	require.Len(suite.T(), resp.Result.Asks, options.Count)
	require.Len(suite.T(), resp.Result.Bids, options.Count)
}

// Integration test for GetRecentTrades.
//
// Test is OK but htere is a flagging issue -> an error is sometime returned by the API despite the URL is OK and lead
// to a correct response in the browser.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetRecentTradesIntegration() {
	// Call API
	params := market.GetRecentTradesRequestParameters{
		Pair: "XBTUSD",
	}
	options := &market.GetRecentTradesRequestOptions{
		Count: 2,
		Since: 1548111600,
	}
	resp, httpresp, err := suite.client.GetRecentTrades(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)

	// Check response
	require.NotEmpty(suite.T(), resp.Result.PairId)
	require.Len(suite.T(), resp.Result.Trades, options.Count)
	require.NotEmpty(suite.T(), resp.Result.Last)
}

// Integration test for GetRecentSpreads.
//
// Test is OK but htere is a flagging issue -> an error is sometime returned by the API despite the URL is OK and lead
// to a correct response in the browser.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetRecentSpreadsIntegration() {
	// Call API
	params := market.GetRecentSpreadsRequestParameters{
		Pair: "XBTUSD",
	}
	options := &market.GetRecentSpreadsRequestOptions{
		Since: 1548111600,
	}
	resp, httpresp, err := suite.client.GetRecentSpreads(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)

	// Check response
	require.NotEmpty(suite.T(), resp.Result.PairId)
	require.NotEmpty(suite.T(), resp.Result.Spreads)
	require.NotEmpty(suite.T(), resp.Result.Last)
}

/*************************************************************************************************/
/* INTEGRATION TESTS - ACCOUNT DATA                                                              */
/*************************************************************************************************/

// Integration test for GetAccountBalance.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetAccountBalanceIntegration() {
	// Call API
	resp, httpresp, err := suite.client.GetAccountBalance(context.Background(), suite.noncegen.GenerateNonce(), suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetExtendedBalance.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetExtendedBalanceIntegration() {
	// Call API
	resp, httpresp, err := suite.client.GetExtendedBalance(context.Background(), suite.noncegen.GenerateNonce(), suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetTradeBalance.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetTradeBalanceIntegration() {
	// Call API
	options := &account.GetTradeBalanceRequestOptions{
		Asset: "ZEUR",
	}
	resp, httpresp, err := suite.client.GetTradeBalance(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetOpenOrders.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetOpenOrdersIntegration() {
	// Call API
	options := &account.GetOpenOrdersRequestOptions{
		Trades:        true,
		UserReference: new(int64),
	}
	*options.UserReference = 10
	resp, httpresp, err := suite.client.GetOpenOrders(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetClosedOrders.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetClosedOrdersIntegration() {
	// Call API
	options := &account.GetClosedOrdersRequestOptions{
		Trades:           true,
		UserReference:    new(int64),
		Start:            strconv.FormatInt(time.Now().Add(-1*3*time.Hour).Unix(), 10),
		End:              strconv.FormatInt(time.Now().Add(-1*1*time.Hour).Unix(), 10),
		Offset:           10,
		Closetime:        string(account.UseOpen),
		ConsolidateTaker: true,
	}
	*options.UserReference = 10
	resp, httpresp, err := suite.client.GetClosedOrders(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for QueryOrdersInfo.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestQueryOrdersInfoIntegration() {
	// Call API
	options := &account.QueryOrdersInfoRequestOptions{
		Trades:           true,
		UserReference:    new(int64),
		ConsolidateTaker: new(bool),
	}
	*options.UserReference = 10
	params := account.QueryOrdersInfoParameters{
		TxId: []string{"OBCMZD-JIEE7-77TH3F", "OBCMZD-JIEE7-77TH3A"},
	}
	resp, httpresp, err := suite.client.QueryOrdersInfo(context.Background(), suite.noncegen.GenerateNonce(), params, options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetTradesHistory.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetTradesHistoryIntegration() {
	// Call API
	options := &account.GetTradesHistoryRequestOptions{
		Type:             string(account.TradeTypeNoPosition),
		Trades:           true,
		Start:            strconv.FormatInt(time.Now().Add(-1*3*time.Hour).Unix(), 10),
		End:              strconv.FormatInt(time.Now().Add(-1*1*time.Hour).Unix(), 10),
		Offset:           10,
		ConsolidateTaker: true,
	}
	resp, httpresp, err := suite.client.GetTradesHistory(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for QueryTradesInfo.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestQueryTradesInfoIntegration() {
	// Call API
	options := &account.QueryTradesRequestOptions{
		Trades: true,
	}
	params := account.QueryTradesRequestParameters{
		TransactionIds: []string{"THVRQM-33VKH-UCI7BS", "THVRQM-33VKH-UCI7BA"},
	}
	resp, httpresp, err := suite.client.QueryTradesInfo(context.Background(), suite.noncegen.GenerateNonce(), params, options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetOpenPositions.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetOpenPositionsIntegration() {
	// Call API
	options := &account.GetOpenPositionsRequestOptions{
		TransactionIds: []string{"TF5GVO-T7ZZ2-6NBKBI", "TF5GVO-T7ZZ2-6NBKBA"},
		DoCalcs:        true,
	}
	resp, httpresp, err := suite.client.GetOpenPositions(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetLedgersInfo.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetLedgersInfoIntegration() {
	// Call API
	options := &account.GetLedgersInfoRequestOptions{
		Assets:       []string{"XXBT", "XETH"},
		AssetClass:   "currency",
		Type:         string(account.LedgerSettled),
		Start:        strconv.FormatInt(time.Now().Add(-1*3*time.Hour).Unix(), 10),
		End:          strconv.FormatInt(time.Now().Add(-1*1*time.Hour).Unix(), 10),
		Offset:       10,
		WithoutCount: true,
	}
	resp, httpresp, err := suite.client.GetLedgersInfo(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetTradeVolume.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetTradeVolumeIntegration() {
	// Call API
	options := &account.GetTradeVolumeRequestOptions{
		Pairs: []string{"XXBTZUSD", "XETHZEUR"},
	}
	resp, httpresp, err := suite.client.GetTradeVolume(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.NotEmpty(suite.T(), resp.Result.Currency)
	require.NotEmpty(suite.T(), resp.Result.Volume.String())
	for _, pair := range options.Pairs {
		require.NotNil(suite.T(), resp.Result.Fees[pair])
	}
}

// Integration test for data export API operations:
//   - RequestExportReport
//   - GetExportReportStatus
//   - RetrieveDataExport
//   - DeleteExportReport
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestDataExportIntegration() {
	// 1. Request data export
	requestExportOptions := &account.RequestExportReportRequestOptions{
		Format:  string(account.CSV),
		Fields:  []string{string(account.FieldsAmount), string(account.FieldsBalance)},
		StartTm: time.Now().Add(-1 * 3 * time.Hour).Unix(),
		EndTm:   time.Now().Add(-1 * time.Hour).Unix(),
	}
	requestExportParams := account.RequestExportReportRequestParameters{
		Report:      string(account.ReportLedgers),
		Description: "integration test",
	}
	requestExportResp, httpresp, err := suite.client.RequestExportReport(context.Background(), suite.noncegen.GenerateNonce(), requestExportParams, requestExportOptions, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), requestExportResp)
	require.Empty(suite.T(), requestExportResp.Error)
	require.NotNil(suite.T(), requestExportResp.Result)
	require.NotEmpty(suite.T(), requestExportResp.Result.Id)
	// Save export ID
	exportId := requestExportResp.Result.Id

	// 2. Poll status
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	available := false
	getExportStatusParams := account.GetExportReportStatusRequestParameters{
		Report: string(account.ReportLedgers),
	}
	for !available {
		select {
		case <-ctx.Done():
			// Timeout
			suite.FailNow("Data export takes too long to be processed")
		default:
			// Poll export statuses
			getExportStatusResp, httpresp, err := suite.client.GetExportReportStatus(context.Background(), suite.noncegen.GenerateNonce(), getExportStatusParams, suite.fa2)
			require.NoError(suite.T(), err)
			require.NotNil(suite.T(), httpresp)
			// Override sensitive data in request to prevent credentials leak in logs
			httpresp.Request.Header["API-Key"][0] = "SECRET"
			httpresp.Request.Header["API-Sign"][0] = "SECRET"
			httpresp.Request.Form.Set("otp", "SECRET")
			httpresp.Request.PostForm.Set("otp", "SECRET")
			suite.T().Logf("sent HTTP request: %v", httpresp.Request)
			suite.T().Logf("received HTTP response: %v", httpresp)
			// Check results
			require.NotNil(suite.T(), getExportStatusResp)
			require.Empty(suite.T(), getExportStatusResp.Error)
			for _, status := range getExportStatusResp.Result {
				// Available will be true only if an entry with the corresponding ID and a status equal to 'Processed' is found
				available = (status.Id == exportId && status.Status == string(account.Processed))
			}
			if !available {
				// Sleep 3 seconds before retry
				suite.T().Log("requested data export is not ready yet. Retrying in 3 seconds")
				time.Sleep(3 * time.Second)
			}
		}
	}

	// 3. Retrieve data export
	retrieveExportParams := account.RetrieveDataExportParameters{
		Id: exportId,
	}
	// Retrieve data export
	retrieveExportResp, httpresp, err := suite.client.RetrieveDataExport(context.Background(), suite.noncegen.GenerateNonce(), retrieveExportParams, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), retrieveExportResp.Report)
	// Download data, store them in local file and open zip
	data, err := io.ReadAll(retrieveExportResp.Report)
	require.NoError(suite.T(), err)
	os.WriteFile("ledgers.zip", data, 0644)
	zipped, err := zip.OpenReader("ledgers.zip")
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), zipped.File)
	// Remove the archive
	os.Remove("ledgers.zip")

	// 4. Delete data export
	deleteExportParams := account.DeleteExportReportRequestParameters{
		Id:   exportId,
		Type: string(account.DeleteReport),
	}
	deleteExportResp, httpresp, err := suite.client.DeleteExportReport(context.Background(), suite.noncegen.GenerateNonce(), deleteExportParams, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.Empty(suite.T(), deleteExportResp.Error)
	require.NotNil(suite.T(), deleteExportResp.Result)
	require.True(suite.T(), deleteExportResp.Result.Delete)
}

/*************************************************************************************************/
/* INTEGRATION TESTS - TRADING                                                                   */
/*************************************************************************************************/

// Integration test for AddOrder.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestAddOrderIntegration() {
	// Call API
	params := trading.AddOrderRequestParameters{
		Pair: "XXBTZUSD",
		Order: trading.Order{
			UserReference:      new(int64),
			OrderType:          string(trading.StopLossLimit),
			Type:               string(trading.Buy),
			Volume:             "0.0002",
			DisplayedVolume:    "",
			Price:              "#5%",
			Price2:             "#.5%",
			Trigger:            string(trading.Last),
			Leverage:           "5:1",
			ReduceOnly:         true,
			StpType:            string(trading.STPCancelBoth),
			OrderFlags:         strings.Join([]string{string(trading.OFlagFeeInQuote)}, ","),
			TimeInForce:        string(trading.GoodTilDate),
			ScheduledStartTime: "+2",
			ExpirationTime:     "+30",
			Close: &trading.CloseOrder{
				OrderType: string(trading.StopLossLimit),
				Price:     "#1000",
				Price2:    "#0.5%",
			},
		},
	}
	options := &trading.AddOrderRequestOptions{
		Validate: true,
		Deadline: time.Now().Add(15 * time.Second),
	}
	// Ensure validate is set to true -> purpose is not to post a real order
	require.NotNil(suite.T(), options)
	require.True(suite.T(), options.Validate)
	resp, httpresp, err := suite.client.AddOrder(context.Background(), suite.noncegen.GenerateNonce(), params, options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Empty(suite.T(), resp.Result.TransactionIDs)
	require.NotEmpty(suite.T(), resp.Result.Description.Order)
	require.NotEmpty(suite.T(), resp.Result.Description.Close)
}

// Integration test for AddOrderBatch.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestAddOrderBatchIntegration() {
	// Call API
	params := trading.AddOrderBatchRequestParameters{
		Pair: "XXBTZUSD",
		Orders: []trading.Order{
			{
				UserReference:      new(int64),
				OrderType:          string(trading.StopLossLimit),
				Type:               string(trading.Buy),
				Volume:             "0.0002",
				DisplayedVolume:    "",
				Price:              "#5%",
				Price2:             "#.5%",
				Trigger:            string(trading.Last),
				Leverage:           "5:1",
				ReduceOnly:         true,
				StpType:            string(trading.STPCancelBoth),
				OrderFlags:         strings.Join([]string{string(trading.OFlagFeeInQuote)}, ","),
				TimeInForce:        string(trading.GoodTilDate),
				ScheduledStartTime: "+2",
				ExpirationTime:     "+30",
				Close: &trading.CloseOrder{
					OrderType: string(trading.StopLossLimit),
					Price:     "#1000",
					Price2:    "#0.5%",
				},
			},
			{
				UserReference: new(int64),
				OrderType:     string(trading.Market),
				Type:          string(trading.Buy),
				Volume:        "0.0002",
				OrderFlags:    strings.Join([]string{string(trading.OFlagFeeInQuote)}, ","),
			},
		},
	}
	options := &trading.AddOrderBatchRequestOptions{
		Validate: true,
		Deadline: time.Now().Add(15 * time.Second),
	}
	// Ensure validate is set to true -> purpose is not to post a real order
	require.NotNil(suite.T(), options)
	require.True(suite.T(), options.Validate)
	resp, httpresp, err := suite.client.AddOrderBatch(context.Background(), suite.noncegen.GenerateNonce(), params, options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Len(suite.T(), resp.Result.Orders, len(params.Orders))
	for _, order := range resp.Result.Orders {
		require.NotEmpty(suite.T(), order.Description.Order)
	}
}

// Integration test for EditOrder.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestEditOrderIntegration() {
	// Call API
	params := trading.EditOrderRequestParameters{
		Id:   "OYVGEW-VYV5B-UUEXSK",
		Pair: "XXBTZUSD",
	}
	options := &trading.EditOrderRequestOptions{
		NewUserReference: "10",
		NewVolume:        "0.003",
		Price:            "+1000",
		Price2:           "-1000",
		OFlags:           []string{string(account.OFlagNoMarketPriceProtection)},
		Validate:         true,
		CancelResponse:   false,
		Deadline:         time.Now().Add(30 * time.Second),
	}
	// Ensure validate is set to true -> purpose is not to post a real order
	require.NotNil(suite.T(), options)
	require.True(suite.T(), options.Validate)
	resp, httpresp, err := suite.client.EditOrder(context.Background(), suite.noncegen.GenerateNonce(), params, options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.NotEmpty(suite.T(), resp.Result.Description)
}

// Integration test for CancelOrder.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestCancelOrderIntegration() {
	// Call API
	params := trading.CancelOrderRequestParameters{
		Id: "OYVGEW-VYV5B-UUEXSK",
	}
	resp, httpresp, err := suite.client.CancelOrder(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Contains(suite.T(), resp.Error, "EOrder:Unknown order")
}

// Integration test for CancelAllOrders.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestCancelAllOrdersIntegration() {
	// Call API
	resp, httpresp, err := suite.client.CancelAllOrders(context.Background(), suite.noncegen.GenerateNonce(), suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for CancelAllOrdersAfterX.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestCancelAllOrdersAfterXIntegration() {
	// Call API
	params := trading.CancelAllOrdersAfterXRequestParameters{
		Timeout: 60,
	}
	resp, httpresp, err := suite.client.CancelAllOrdersAfterX(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.NotEmpty(suite.T(), resp.Result.CurrentTime)
	require.NotEmpty(suite.T(), resp.Result.TriggerTime)
}

// Integration test for CancelOrderBatch.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestCancelOrderBatchIntegration() {
	// Call API
	params := trading.CancelOrderBatchRequestParameters{
		OrderIds: []string{"OYVGEW-VYV5B-UUEXSK", "OYVGEW-VYV5B-UUEXSA"},
	}
	resp, httpresp, err := suite.client.CancelOrderBatch(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Contains(suite.T(), resp.Error, "EGeneral:Invalid arguments:orders")
}

/*************************************************************************************************/
/* INTEGRATION TESTS - FUNDING                                                                   */
/*************************************************************************************************/

// Integration test for GetDepositMethods.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetDepositMethodsIntegration() {
	// Call API
	params := funding.GetDepositMethodsRequestParameters{
		Asset: "XXBT",
	}
	resp, httpresp, err := suite.client.GetDepositMethods(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetDepositAddresses.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetDepositAddressesIntegration() {
	// Call API
	params := funding.GetDepositAddressesRequestParameters{
		Asset:  "XXBT",
		Method: "Bitcoin Lightning",
	}
	options := &funding.GetDepositAddressesRequestOptions{
		New:    true,
		Amount: "0.1",
	}
	resp, httpresp, err := suite.client.GetDepositAddresses(context.Background(), suite.noncegen.GenerateNonce(), params, options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotEmpty(suite.T(), resp.Result)
}

// Integration test for GetStatusOfRecentDeposits.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetStatusOfRecentDepositsIntegration() {
	// Call API
	options := &funding.GetStatusOfRecentDepositsRequestOptions{
		Asset:  "XBT",
		Method: "Bitcoin",
		Start:  strconv.FormatInt(time.Now().Add(-1*3*time.Hour).Unix(), 10),
		End:    strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10),
	}
	resp, httpresp, err := suite.client.GetStatusOfRecentDeposits(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetWithdrawalMethods.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetWithdrawalMethodsIntegration() {
	// Call API
	options := &funding.GetWithdrawalMethodsRequestOptions{
		Asset:   "XXBT",
		Network: "Bitcoin",
	}
	resp, httpresp, err := suite.client.GetWithdrawalMethods(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for GetWithdrawalInformation.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetWithdrawalInformationIntegration() {
	// Call API
	params := funding.GetWithdrawalInformationRequestParameters{
		Asset:  "XBT",
		Key:    "TEST",
		Amount: "0.01",
	}
	resp, httpresp, err := suite.client.GetWithdrawalInformation(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	// A withdrawal address must be configured or an error will be returned.
	// Test will not check the API response in order to not block tests.
	//require.Empty(suite.T(), resp.Error)
	//require.NotNil(suite.T(), resp.Result)
}

// Integration test for WithdrawFunds.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestWithdrawFundsIntegration() {
	// Call API
	params := funding.WithdrawFundsRequestParameters{
		Asset:  "XBT",
		Key:    "TEST",
		Amount: "21000000",
	}
	options := &funding.WithdrawFundsRequestOptions{
		Address: "XXXXXXX",
		MaxFee:  "1",
	}
	resp, httpresp, err := suite.client.WithdrawFunds(context.Background(), suite.noncegen.GenerateNonce(), params, options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.NotEmpty(suite.T(), resp.Error)
	// You must configure a withdrawal address in order to really test the operation.
	// Test is set in a way we expect the API to return an "insufficient funds" error.
	//require.Contains(suite.T(), resp.Error, "EFunding:Insufficient funds")
	require.Nil(suite.T(), resp.Result)
}

// Integration test for GetStatusOfRecentWithdrawals.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetStatusOfRecentWithdrawalsIntegration() {
	// Call API
	options := &funding.GetStatusOfRecentWithdrawalsRequestOptions{
		Method: "Bitcoin",
		Asset:  "XBT",
		Start:  strconv.FormatInt(time.Now().Add(-1*3*time.Hour).Unix(), 10),
		End:    strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10),
	}
	resp, httpresp, err := suite.client.GetStatusOfRecentWithdrawals(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}

// Integration test for RequestWithdrawalCancellation.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestRequestWithdrawalCancellationIntegration() {
	// Call API
	params := funding.RequestWithdrawalCancellationRequestParameters{
		Asset:       "XBT",
		ReferenceId: "TEST",
	}
	resp, httpresp, err := suite.client.RequestWithdrawalCancellation(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Contains(suite.T(), resp.Error, "EFunding:Unknown reference id")
	require.False(suite.T(), resp.Result)
}

// Integration test for RequestWalletTransfer.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestRequestWalletTransferIntegration() {
	// Call API
	params := funding.RequestWalletTransferRequestParameters{
		Asset:  "XBT",
		From:   string(funding.Spot),
		To:     string(funding.Futures),
		Amount: "2100000",
	}
	resp, httpresp, err := suite.client.RequestWalletTransfer(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.NotEmpty(suite.T(), resp.Error)
	require.Nil(suite.T(), resp.Result)
}

/*************************************************************************************************/
/* INTEGRATION TESTS - EARN                                                                      */
/*************************************************************************************************/

// Integration test for AllocateEarnFunds.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestAllocateEarnFundsIntegration() {
	// Call API
	params := earn.AllocateFundsRequestParameters{
		Amount:     "100000",
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}
	resp, httpresp, err := suite.client.AllocateEarnFunds(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Contains(suite.T(), resp.Error, "EFunding:Insufficient funds")
	require.False(suite.T(), resp.Result)
}

// Integration test for DeallocateEarnFunds.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestDeallocateEarnFundsIntegration() {
	// Call API
	params := earn.DeallocateFundsRequestParameters{
		Amount:     "100000",
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}
	resp, httpresp, err := suite.client.DeallocateEarnFunds(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Contains(suite.T(), resp.Error, "EFunding:Insufficient funds")
	require.False(suite.T(), resp.Result)
}

// Integration test for GetAllocationStatus.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetAllocationStatusIntegration() {
	// Call API
	params := earn.GetAllocationStatusRequestParameters{
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}
	resp, httpresp, err := suite.client.GetAllocationStatus(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.False(suite.T(), resp.Result.Pending)
}

// Integration test for GetDeallocationStatus.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetDeallocationStatusIntegration() {
	// Call API
	params := earn.GetDeallocationStatusRequestParameters{
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}
	resp, httpresp, err := suite.client.GetDeallocationStatus(context.Background(), suite.noncegen.GenerateNonce(), params, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.False(suite.T(), resp.Result.Pending)
}

// Integration test for ListEarnStrategies.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestListEarnStrategiesIntegration() {
	// Call API
	options := &earn.ListEarnStrategiesRequestOptions{
		Ascending: true,
		Asset:     "DOT",
		LockType:  []string{string(earn.Flex), string(earn.Instant)},
	}
	resp, httpresp, err := suite.client.ListEarnStrategies(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.NotEmpty(suite.T(), resp.Result.Items)
}

// Integration test for ListEarnAllocations.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestListEarnAllocationsIntegration() {
	// Call API
	options := &earn.ListEarnAllocationsRequestOptions{
		Ascending:           true,
		ConvertedAsset:      "EUR",
		HideZeroAllocations: true,
	}
	resp, httpresp, err := suite.client.ListEarnAllocations(context.Background(), suite.noncegen.GenerateNonce(), options, suite.fa2)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	// Override sensitive data in request to prevent credentials leak in logs
	httpresp.Request.Header["API-Key"][0] = "SECRET"
	httpresp.Request.Header["API-Sign"][0] = "SECRET"
	httpresp.Request.Form.Set("otp", "SECRET")
	httpresp.Request.PostForm.Set("otp", "SECRET")
	suite.T().Logf("sent HTTP request: %v", httpresp.Request)
	suite.T().Logf("received HTTP response: %v", httpresp)
	// Check results
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
}
