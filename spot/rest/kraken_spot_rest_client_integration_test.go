package rest

import (
	"context"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
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
}

// Configure and run unit test suite
func TestKrakenSpotRESTClientIntegrationTestSuite(t *testing.T) {

	// Load credentials for Kraken spot REST API
	key := os.Getenv("KRAKEN_API_KEY")
	b64Secret := os.Getenv("KRAKEN_API_SECRET")
	otp := os.Getenv("KRAKEN_API_OTP")

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
		Suite:  suite.Suite{},
		client: client,
		fa2:    fa2,
	})
}

/*************************************************************************************************/
/* INTEGRATION TESTS - MARKET DATA                                                               */
/*************************************************************************************************/

// Integration test for GetServerTime.
func (suite *KrakenSpotRESTClientIntegrationTestSuite) TestGetServerTimeIntegration() {
	// Call API
	resp, httpresp, err := suite.client.GetServerTime(context.Background())
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
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
	suite.T().Logf("received HTTP response: %v", httpresp)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)

	// Check response
	require.NotEmpty(suite.T(), resp.Result.PairId)
	require.NotEmpty(suite.T(), resp.Result.Spreads)
	require.NotEmpty(suite.T(), resp.Result.Last)
}
