package krakenapiclient

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	mockhttpserver "gitlab.com/lake42/mock-http-server"
)

// Unit test suite for Kraken API client
type KrakenAPIClientUnitTestSuite struct {
	suite.Suite
	// Mock HTTP server
	srv *mockhttpserver.MockHTTPServer
	// Kraken API client configured to use mock HTTP server
	client *KrakenAPIClient
	// Fake API key used for tests
	key string
	// Fake API Key secret used for tests
	secret []byte
}

// Run unit test suite
func TestKrakenAPIClientUnitTestSuite(t *testing.T) {

	// Build fixtures
	fakeKey := "FAKE_KEY"
	fakeSecret := []byte("FAKE_SECRET")
	mockSrv := mockhttpserver.GetRunningMockHTTPServer()
	defer mockSrv.Close()
	client := NewWithCredentialsAndOptions(fakeKey, fakeSecret, &KrakenAPIClientOptions{BaseURL: mockSrv.GetMockHTTPServerBaseURL()})

	// Run unit test suite
	suite.Run(t, &KrakenAPIClientUnitTestSuite{
		srv:    mockSrv,
		client: client,
		key:    fakeKey,
		secret: fakeSecret,
	})
}

// Before every test in the suite
func (suite *KrakenAPIClientUnitTestSuite) BeforeTest(suiteName, testName string) {
	// Clear responses & requests from mock http server
	suite.srv.Clear()
}

// Tear down after all tests in the suite
func (suite *KrakenAPIClientUnitTestSuite) TearDownSuite() {
	// Close the mock http server
	suite.srv.Close()
}

/*****************************************************************************/
/* UNIT TESTS - MARKET DATA                                                  */
/*****************************************************************************/

// TestGetServerTimeHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetServerTimeHappyPath() {

	// Predefined server response
	expectedJSONResponse := `
	{
		"error": [ ],
		"result": {
			"unixtime": 1616336594,
			"rfc1123": "Sun, 21 Mar 21 14:23:14 +0000"
		}
	}`

	// Expected data
	expUnixTime := int64(1616336594)
	expRFC1123 := "Sun, 21 Mar 21 14:23:14 +0000"

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetServerTime()

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getServerTime)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expRFC1123, resp.Result.Rfc1123)
	require.Equal(suite.T(), expUnixTime, resp.Result.Unixtime)
}

// TestGetServerTimeErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetServerTimeErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	_, err := suite.client.GetServerTime()

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
}

// TestGetSystemStatusHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetSystemStatusHappyPath() {

	// Predefined server response
	expectedJSONResponse := `
	{
		"error": [ ],
		"result": {
			"status": "online",
			"timestamp": "2021-03-21T15:33:02Z"
		}
	}`

	// Expected data
	expStatus := "online"
	expTimestamp := "2021-03-21T15:33:02Z"

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetSystemStatus()

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getSystemStatus)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())

	// Check response
	require.Equal(suite.T(), expStatus, resp.Result.Status)
	require.Equal(suite.T(), expTimestamp, resp.Result.Timestamp)
}

// TestGetServerTimeErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetSystemStatusErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusBadRequest,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(`{"error": "wrong"}`),
	})

	// Make request
	_, err := suite.client.GetSystemStatus()

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
}

// TestGetAssetInfoHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetAssetInfoHappyPath() {

	// Test parameters
	options := GetAssetInfoOptions{
		Assets:     []string{"XETH", "XXBT"},
		AssetClass: "currency",
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error":[],
		"result":{
			"XETH":{
				"aclass":"currency",
				"altname":"ETH",
				"decimals":10,
				"display_decimals":5,
				"collateral_value":1.0
			},
			"XXBT":{
				"aclass":"currency",
				"altname":"XBT",
				"decimals":10,
				"display_decimals":5,
				"collateral_value":1.0
			}
		}
	}`

	expData := map[string]AssetInfo{
		options.Assets[0]: {
			AssetClass:      options.AssetClass,
			Altname:         "ETH",
			Decimals:        10,
			DisplayDecimals: 5,
			CollateralValue: 1.0,
		},
		options.Assets[1]: {
			AssetClass:      options.AssetClass,
			Altname:         "XBT",
			Decimals:        10,
			DisplayDecimals: 5,
			CollateralValue: 1.0,
		},
	}

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetAssetInfo(&options)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getAssetInfo)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.ElementsMatch(suite.T(), options.Assets, strings.Split(req.Form.Get("asset"), ","))
	require.Equal(suite.T(), options.AssetClass, req.Form.Get("aclass"))

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetAssetInfoErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetAssetInfoErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	_, err := suite.client.GetAssetInfo(nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
}

// TestGetTradableAssetPairsHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradableAssetPairsHappyPath() {

	// Test parameters
	options := GetTradableAssetPairsOptions{
		Pairs: []string{"XETHXXBT", "XXBTZUSD"},
		Info:  "all",
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error":[],
		"result":{
			"XETHXXBT":{
				"altname":"ETHXBT",
				"wsname":"ETH/XBT",
				"aclass_base":"currency",
				"base":"XETH",
				"aclass_quote":"currency",
				"quote":"XXBT",
				"pair_decimals":5,
				"lot_decimals":8,
				"lot_multiplier":1,
				"leverage_buy":[2,3,4,5],
				"leverage_sell":[2,3,4,5],
				"fees":[[0,0.26],[50000,0.24],[100000,0.22],[250000,0.2],[500000,0.18],[1000000,0.16],[2500000,0.14],[5000000,0.12],[10000000,0.1]],
				"fees_maker":[[0,0.16],[50000,0.14],[100000,0.12],[250000,0.1],[500000,0.08],[1000000,0.06],[2500000,0.04],[5000000,0.02],[10000000,0.0]],
				"fee_volume_currency":"ZUSD",
				"margin_call":80,
				"margin_stop":40,
				"ordermin":"0.01"
			},
			"XXBTZUSD":{
				"altname":"XBTUSD",
				"wsname":"XBT/USD",
				"aclass_base":"currency",
				"base":"XXBT",
				"aclass_quote":"currency",
				"quote":"ZUSD",
				"pair_decimals":1,
				"lot_decimals":8,
				"lot_multiplier":1,
				"leverage_buy":[2,3,4,5],
				"leverage_sell":[2,3,4,5],
				"fees":[[0,0.26],[50000,0.24],[100000,0.22],[250000,0.2],[500000,0.18],[1000000,0.16],[2500000,0.14],[5000000,0.12],[10000000,0.1]],
				"fees_maker":[[0,0.16],[50000,0.14],[100000,0.12],[250000,0.1],[500000,0.08],[1000000,0.06],[2500000,0.04],[5000000,0.02],[10000000,0.0]],
				"fee_volume_currency":"ZUSD",
				"margin_call":80,
				"margin_stop":40,
				"ordermin":"0.0001"
			}
		}
	}`
	expData := map[string]AssetPairInfo{
		options.Pairs[0]: {
			Altname:           "ETHXBT",
			WsName:            "ETH/XBT",
			AssetClassBase:    "currency",
			Base:              "XETH",
			AssetClassQuote:   "currency",
			Quote:             "XXBT",
			PairDecimals:      5,
			LotDecimals:       8,
			LotMultiplier:     1,
			LeverageBuy:       []int{2, 3, 4, 5},
			LeverageSell:      []int{2, 3, 4, 5},
			Fees:              [][]float64{{0, 0.26}, {50000, 0.24}, {100000, 0.22}, {250000, 0.2}, {500000, 0.18}, {1000000, 0.16}, {2500000, 0.14}, {5000000, 0.12}, {10000000, 0.1}},
			FeesMaker:         [][]float64{{0, 0.16}, {50000, 0.14}, {100000, 0.12}, {250000, 0.1}, {500000, 0.08}, {1000000, 0.06}, {2500000, 0.04}, {5000000, 0.02}, {10000000, 0.0}},
			FeeVolumeCurrency: "ZUSD",
			MarginCall:        80,
			MarginStop:        40,
			OrderMin:          "0.01",
		},
		options.Pairs[1]: {
			Altname:           "XBTUSD",
			WsName:            "XBT/USD",
			AssetClassBase:    "currency",
			Base:              "XXBT",
			AssetClassQuote:   "currency",
			Quote:             "ZUSD",
			PairDecimals:      1,
			LotDecimals:       8,
			LotMultiplier:     1,
			LeverageBuy:       []int{2, 3, 4, 5},
			LeverageSell:      []int{2, 3, 4, 5},
			Fees:              [][]float64{{0, 0.26}, {50000, 0.24}, {100000, 0.22}, {250000, 0.2}, {500000, 0.18}, {1000000, 0.16}, {2500000, 0.14}, {5000000, 0.12}, {10000000, 0.1}},
			FeesMaker:         [][]float64{{0, 0.16}, {50000, 0.14}, {100000, 0.12}, {250000, 0.1}, {500000, 0.08}, {1000000, 0.06}, {2500000, 0.04}, {5000000, 0.02}, {10000000, 0.0}},
			FeeVolumeCurrency: "ZUSD",
			MarginCall:        80,
			MarginStop:        40,
			OrderMin:          "0.0001",
		},
	}

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetTradableAssetPairs(&options)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getTradableAssetPairs)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.ElementsMatch(suite.T(), options.Pairs, strings.Split(req.Form.Get("pair"), ","))
	require.Equal(suite.T(), options.Info, req.Form.Get("info"))

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetTradableAssetPairsErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradableAssetPairsErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	resp, err := suite.client.GetTradableAssetPairs(nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestGetTickerInformationHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTickerInformationHappyPath() {

	// Test parameters
	opts := &GetTickerInformationOptions{
		Pairs: []string{"XETHXXBT", "XXBTZUSD"},
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error":[],
		"result":{
			"XETHXXBT":{
				"a":["0.078870","2","2.000"],
				"b":["0.078860","1","1.000"],
				"c":["0.078860","0.00484694"],
				"v":["1487.94072797","3495.24626651"],
				"p":["0.079151","0.079541"],
				"t":[18730,28522],
				"l":["0.078320","0.078320"],
				"h":["0.080550","0.080960"],
				"o":"0.079630"
			},
			"XXBTZUSD":{
				"a":["24100.10000","10","10.000"],
				"b":["24100.00000","1","1.000"],
				"c":["24100.00000","0.00935269"],
				"v":["2870.28639431","3816.56628826"],
				"p":["24416.92776","24398.48513"],
				"t":[23035,31242],
				"l":["23900.00000","23900.00000"],
				"h":["25200.00000","25200.00000"],
				"o":"24315.30000"
			}
		}
	}`

	expData := map[string]AssetTickerInfo{
		opts.Pairs[0]: {
			Ask:                []string{"0.078870", "2", "2.000"},
			Bid:                []string{"0.078860", "1", "1.000"},
			Close:              []string{"0.078860", "0.00484694"},
			Volume:             []string{"1487.94072797", "3495.24626651"},
			VolumeAveragePrice: []string{"0.079151", "0.079541"},
			Trades:             []int{18730, 28522},
			Low:                []string{"0.078320", "0.078320"},
			High:               []string{"0.080550", "0.080960"},
			OpeningPrice:       "0.079630",
		},
		opts.Pairs[1]: {
			Ask:                []string{"24100.10000", "10", "10.000"},
			Bid:                []string{"24100.00000", "1", "1.000"},
			Close:              []string{"24100.00000", "0.00935269"},
			Volume:             []string{"2870.28639431", "3816.56628826"},
			VolumeAveragePrice: []string{"24416.92776", "24398.48513"},
			Trades:             []int{23035, 31242},
			Low:                []string{"23900.00000", "23900.00000"},
			High:               []string{"25200.00000", "25200.00000"},
			OpeningPrice:       "24315.30000",
		},
	}

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetTickerInformation(opts)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getTickerInformation)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.ElementsMatch(suite.T(), opts.Pairs, strings.Split(req.Form.Get("pair"), ","))

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetTickerInformationErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTickerInformationErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	resp, err := suite.client.GetTickerInformation(nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestGetOHLCDataHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOHLCDataHappyPath() {

	// Test parameters
	params := GetOHLCDataParameters{
		Pair: "XXBTZUSD",
	}
	now := time.Now().UTC()
	options := GetOHLCDataOptions{
		Interval: M60,
		Since:    &now,
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error":[],
		"result":{
			"XXBTZUSD":[
				[1660534620,"25055.3","25067.1","25054.0","25054.0","25059.8","0.07453437",10],
				[1660534680,"25041.0","25041.0","24988.4","24994.6","24992.0","1.41833093",98]
			],
			"last":1660577700
		}
	}`

	expData := GetOHLCDataResult{
		OHLC: map[string][]OHLCData{
			params.Pair: {
				{
					Timestamp: time.Unix(1660534620, 0).UTC(),
					Open:      "25055.3",
					High:      "25067.1",
					Low:       "25054.0",
					Close:     "25054.0",
					Avg:       "25059.8",
					Volume:    "0.07453437",
					Count:     10,
				},
				{
					Timestamp: time.Unix(1660534680, 0).UTC(),
					Open:      "25041.0",
					High:      "25041.0",
					Low:       "24988.4",
					Close:     "24994.6",
					Avg:       "24992.0",
					Volume:    "1.41833093",
					Count:     98,
				},
			},
		},
		Last: 1660577700,
	}

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetOHLCData(params, &options)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getOHLCData)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), params.Pair, req.Form.Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(int64(options.Interval), 10), req.Form.Get("interval"))
	require.Equal(suite.T(), strconv.FormatInt(options.Since.Unix(), 10), req.Form.Get("since"))

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetOHLCDataErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOHLCDataErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	resp, err := suite.client.GetOHLCData(GetOHLCDataParameters{}, nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestGetOrderBookHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOrderBookHappyPath() {

	// Test parameters
	params := GetOrderBookParameters{
		Pair: "XXBTZUSD",
	}
	options := GetOrderBookOptions{
		Count: 2,
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error":[],
		"result":{
			"XXBTZUSD":{
				"asks":[
					["23991.20000","3.039",1660634851],
					["23991.50000","3.127",1660634829]
				],
				"bids":[
					["23991.10000","0.011",1660634853],
					["23986.90000","0.651",1660634852]
				]
			}
		}
	}`

	expData := map[string]OrderBook{
		params.Pair: {
			Asks: []OrderBookEntry{
				{
					Timestamp: time.Unix(1660634851, 0).UTC(),
					Price:     "23991.20000",
					Volume:    "3.039",
				},
				{
					Timestamp: time.Unix(1660634829, 0).UTC(),
					Price:     "23991.50000",
					Volume:    "3.127",
				},
			},
			Bids: []OrderBookEntry{
				{
					Timestamp: time.Unix(1660634853, 0).UTC(),
					Price:     "23991.10000",
					Volume:    "0.011",
				},
				{
					Timestamp: time.Unix(1660634852, 0).UTC(),
					Price:     "23986.90000",
					Volume:    "0.651",
				},
			},
		},
	}

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetOrderBook(params, &options)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getOrderBook)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), params.Pair, req.Form.Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(int64(options.Count), 10), req.Form.Get("count"))

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetOrderBookErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOrderBookErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	resp, err := suite.client.GetOrderBook(GetOrderBookParameters{}, nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestGetRecentTradesHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetRecentTradesHappyPath() {

	// Test parameters
	params := GetRecentTradesParameters{
		Pair: "XXBTZUSD",
	}
	now := time.Now().UTC()
	options := GetRecentTradesOptions{
		Since: &now,
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error":[],
		"result":{
			"XXBTZUSD":[
				["24006.10000","0.00010000",1660639679.019115255,"s","l","", 41557503],
				["24006.20000","0.08329382",1660639679.494113755,"b","l","", 41557504],
				["24006.10000","0.02300000",1660639679.596130855,"b","m","", 41557505]
			],
			"last":"1660639679596130788"
		}
	}`

	expData := GetRecentTradesResult{
		Last: "1660639679596130788",
		Trades: map[string][]Trade{
			params.Pair: {
				{
					Timestamp:     1660639679019,
					Price:         "24006.10000",
					Volume:        "0.00010000",
					Side:          "s",
					Type:          "l",
					Miscellaneous: "",
					Id:            41557503,
				},
				{
					Timestamp:     1660639679494,
					Price:         "24006.20000",
					Volume:        "0.08329382",
					Side:          "b",
					Type:          "l",
					Miscellaneous: "",
					Id:            41557504,
				},
				{
					Timestamp:     1660639679596,
					Price:         "24006.10000",
					Volume:        "0.02300000",
					Side:          "b",
					Type:          "m",
					Miscellaneous: "",
					Id:            41557505,
				},
			},
		},
	}

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetRecentTrades(params, &options)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getRecentTrades)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), params.Pair, req.Form.Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(int64(options.Since.Unix()), 10), req.Form.Get("since"))

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetRecentTradesErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetRecentTradesErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	resp, err := suite.client.GetRecentTrades(GetRecentTradesParameters{}, nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestGetRecentSpreadsHappyPath Test will succeed if client handles well a valid response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetRecentSpreadsHappyPath() {

	// Test parameters
	params := GetRecentSpreadsParameters{
		Pair: "XXBTZUSD",
	}
	now := time.Now().UTC()
	options := GetRecentSpreadsOptions{
		Since: &now,
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error":[],
		"result":{
			"XXBTZUSD":[
				[1660641970,"24103.30000","24103.50000"],
				[1660641970,"24103.30000","24103.40000"]
			],
			"last":1660641970
		}
	}`

	expData := GetRecentSpreadsResult{
		Last: "1660641970",
		Spreads: map[string][]SpreadData{
			params.Pair: {
				{
					Timestamp: time.Unix(1660641970, 0).UTC(),
					BestBid:   "24103.50000",
					BestAsk:   "24103.30000",
				},
				{
					Timestamp: time.Unix(1660641970, 0).UTC(),
					BestBid:   "24103.40000",
					BestAsk:   "24103.30000",
				},
			},
		},
	}

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, err := suite.client.GetRecentSpreads(params, &options)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	require.Contains(suite.T(), req.URL.Path, getRecentSpreads)
	require.Equal(suite.T(), http.MethodGet, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), params.Pair, req.Form.Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(int64(options.Since.Unix()), 10), req.Form.Get("since"))

	// Check response
	require.Empty(suite.T(), resp.Error)
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetRecentSpreadsErrPath Test will succeed if client handles well an error response from server.
func (suite *KrakenAPIClientUnitTestSuite) TestGetRecentSpreadsErrPath() {

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Make request
	resp, err := suite.client.GetRecentSpreads(GetRecentSpreadsParameters{}, nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for error
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

/*****************************************************************************/
/* UNIT TESTS - USER DATA													 */
/*****************************************************************************/

// TestGetAccountBalanceHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetAccountBalanceHappyPath() {

	// 2FA
	secopts := SecurityOptions{
		SecondFactor: "N0PE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
		  "ZUSD": "171288.6158",
		  "ZEUR": "504861.8946",
		  "ZGBP": "459567.9171",
		  "ZAUD": "500000.0000",
		  "ZCAD": "500000.0000",
		  "CHF": "500000.0000",
		  "XXBT": "1011.1908877900",
		  "XXRP": "100000.00000000",
		  "XLTC": "2000.0000000000",
		  "XETH": "818.5500000000",
		  "XETC": "1000.0000000000",
		  "XREP": "1000.0000000000",
		  "XXMR": "1000.0000000000",
		  "USDT": "500000.00000000",
		  "DASH": "1000.0000000000",
		  "GNO": "1000.0000000000",
		  "EOS": "1000.0000000000",
		  "BCH": "1016.6005000000",
		  "ADA": "100000.00000000",
		  "QTUM": "1000.0000000000",
		  "XTZ": "100000.00000000",
		  "ATOM": "100000.00000000",
		  "SC": "9999.9999999999",
		  "LSK": "1000.0000000000",
		  "WAVES": "1000.0000000000",
		  "ICX": "1000.0000000000",
		  "BAT": "1000.0000000000",
		  "OMG": "1000.0000000000",
		  "LINK": "1000.0000000000",
		  "DAI": "9999.9999999999",
		  "PAXG": "1000.0000000000",
		  "ALGO": "100000.00000000",
		  "USDC": "100000.00000000",
		  "TRX": "100000.00000000",
		  "DOT": "2.5000000000",
		  "OXT": "1000.0000000000",
		  "ETH2.S": "198.3970800000",
		  "ETH2": "2.5885574330",
		  "USD.M": "1213029.2780"
		}
	}`

	// Configure mock server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetAccountBalance(&secopts)
	require.NoError(suite.T(), err)

	// Get client request recorded by mock http server
	req := suite.srv.PopRecordedRequest()

	// Log request
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetAccountBalance)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

	// Unmarshall predefined response
	expectedJSONData := struct {
		Result map[string]string `json:"result"`
	}{}
	json.Unmarshal([]byte(expectedJSONResponse), &expectedJSONData)

	// Check that there is the same number of entries in predefined response and the response provided by client
	require.Equal(suite.T(), expectedJSONData.Result, resp.Result)
}

// TestGetAccountBalanceEmptyResponse Test will succeed if a valid empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetAccountBalanceEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
			"error": [],
			"result": {}
		}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetAccountBalance(nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetAccountBalance)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))

	// Check response - number of entries must be 0
	require.Empty(suite.T(), resp.Result)

	// Check that requested asset provide golang default value
	require.Empty(suite.T(), resp.Result["XXBT"])
}

// TestGetAccountBalanceErrPath Test will succeed if an invalid response from server triggers an error.
func (suite *KrakenAPIClientUnitTestSuite) TestGetAccountBalanceErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetAccountBalance(nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check error and response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestGetTradeBalanceHapyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradeBalanceHapyPath() {

	// Test parameters
	secopts := SecurityOptions{
		SecondFactor: "N0PE",
	}
	options := GetTradeBalanceOptions{
		Asset: "ZUSD",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"eb": "3224744.0162",
			"tb": "3224744.0162",
			"m": "0.0000",
			"n": "0.0000",
			"c": "0.0000",
			"v": "0.0000",
			"e": "3224744.0162",
			"mf": "3224744.0162",
			"ml": "0.0000"
		}
	}`

	expResp := GetTradeBalanceResponse{
		KrakenAPIResponse: KrakenAPIResponse{Error: []string{}},
		Result: GetTradeBalanceResult{
			EquivalentBalance: "3224744.0162",
			TradeBalance:      "3224744.0162",
			MarginAmount:      "0.0000",
			UnrealizedNetPNL:  "0.0000",
			CostBasis:         "0.0000",
			FloatingValuation: "0.0000",
			Equity:            "3224744.0162",
			FreeMargin:        "3224744.0162",
			MarginLevel:       "0.0000",
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetTradeBalance(&options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetTradeBalance)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), options.Asset, req.Form.Get("asset"))

	// Check response
	require.Equal(suite.T(), expResp, *resp)
}

// TestGetTradeBalanceErrPath Test will succeed if an invalid response from server triggers an error.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradeBalanceErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetTradeBalance(nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check error and response
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestGetOpenOrdersHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOpenOrdersHappyPath() {

	// Test parameters
	secopts := SecurityOptions{
		SecondFactor: "N0PE",
	}
	options := GetOpenOrdersOptions{
		Trades:        true,
		UserReference: new(int64),
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"open": {
				"txid1": {
					"refid": "string",
					"userref": "string",
					"status": "pending",
					"opentm": 0,
					"starttm": 0,
					"expiretm": 0,
					"descr": {
						"pair": "string",
						"type": "buy",
						"ordertype": "market",
						"price": "string",
						"price2": "string",
						"leverage": "string",
						"order": "string",
						"close": "string"
					},
					"vol": "string",
					"vol_exec": "string",
					"cost": "string",
					"fee": "string",
					"price": "string",
					"stopprice": "string",
					"limitprice": "string",
					"trigger": "last",
					"misc": "string",
					"oflags": "string",
					"trades": [
						"string"
					]
				},
				"txid2": {
					"refid": "string",
					"userref": "string",
					"status": "pending",
					"opentm": 0,
					"starttm": 0,
					"expiretm": 0,
					"descr": {
						"pair": "string",
						"type": "buy",
						"ordertype": "market",
						"price": "string",
						"price2": "string",
						"leverage": "string",
						"order": "string",
						"close": "string"
					},
					"vol": "string",
					"vol_exec": "string",
					"cost": "string",
					"fee": "string",
					"price": "string",
					"stopprice": "string",
					"limitprice": "string",
					"trigger": "last",
					"misc": "string",
					"oflags": "string",
					"trades": []
				}
			}
		}
	}`

	expResp := GetOpenOrdersResponse{
		KrakenAPIResponse: KrakenAPIResponse{Error: []string{}},
		Result: GetOpenOrdersResult{
			Open: map[string]OrderInfo{
				"txid1": {
					ReferralOrderTransactionId: "string",
					UserReferenceId:            "string",
					Status:                     "pending",
					OpenTimestamp:              0,
					StartTimestamp:             0,
					ExpireTimestamp:            0,
					Description: OrderInfoDescription{
						Pair:                  "string",
						Type:                  "buy",
						OrderType:             "market",
						Price:                 "string",
						Price2:                "string",
						Leverage:              "string",
						OrderDescription:      "string",
						CloseOrderDescription: "string",
					},
					Volume:         "string",
					VolumeExecuted: "string",
					Cost:           "string",
					Fee:            "string",
					Price:          "string",
					StopPrice:      "string",
					LimitPrice:     "string",
					Trigger:        "last",
					Miscellaneous:  "string",
					OrderFlags:     "string",
					Trades:         []string{"string"},
				},
				"txid2": {
					ReferralOrderTransactionId: "string",
					UserReferenceId:            "string",
					Status:                     "pending",
					OpenTimestamp:              0,
					StartTimestamp:             0,
					ExpireTimestamp:            0,
					Description: OrderInfoDescription{
						Pair:                  "string",
						Type:                  "buy",
						OrderType:             "market",
						Price:                 "string",
						Price2:                "string",
						Leverage:              "string",
						OrderDescription:      "string",
						CloseOrderDescription: "string",
					},
					Volume:         "string",
					VolumeExecuted: "string",
					Cost:           "string",
					Fee:            "string",
					Price:          "string",
					StopPrice:      "string",
					LimitPrice:     "string",
					Trigger:        "last",
					Miscellaneous:  "string",
					OrderFlags:     "string",
					Trades:         []string{},
				},
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetOpenOrders(&options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response received by server : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetOpenOrders)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), req.Form.Get("trades"))
	require.Equal(suite.T(), strconv.FormatInt(*options.UserReference, 10), req.Form.Get("userref"))

	// Check response
	require.Equal(suite.T(), expResp, *resp)
}

// TestGetOpenOrdersEmptyResponse Test will succeed if a valid response without any open orders from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOpenOrdersEmptyResponse() {

	// Expected API response with no open orders
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"open": {}
		}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetOpenOrders(nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response received by server : %#v", resp)

	// Check response
	require.NotNil(suite.T(), resp.Result.Open)
	require.Empty(suite.T(), resp.Result.Open)
}

// TestGetOpenOrdersErrPath Test will succeed if an invalid response from server triggers an error.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOpenOrdersErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetOpenOrders(nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestGetClosedOrdersHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetClosedOrdersHappyPath() {

	// Test parameters
	secopts := SecurityOptions{
		SecondFactor: "N0PE",
	}
	options := GetClosedOrdersOptions{
		Trades:        true,
		UserReference: new(int64),
		Start:         &time.Time{},
		End:           &time.Time{},
		Offset:        new(int64),
		Closetime:     "open",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"closed": {
				"txid1": {
					"refid": "string",
					"userref": "string",
					"status": "pending",
					"opentm": 0,
					"starttm": 0,
					"expiretm": 0,
					"descr": {
						"pair": "string",
						"type": "buy",
						"ordertype": "market",
						"price": "string",
						"price2": "string",
						"leverage": "string",
						"order": "string",
						"close": "string"
					},
					"vol": "string",
					"vol_exec": "string",
					"cost": "string",
					"fee": "string",
					"price": "string",
					"stopprice": "string",
					"limitprice": "string",
					"trigger": "last",
					"misc": "string",
					"oflags": "string",
					"trades": [
						"string"
					],
					"closetm": 0,
					"reason": "string"
				},
				"txid2": {
					"refid": "string",
					"userref": "string",
					"status": "pending",
					"opentm": 0,
					"starttm": 0,
					"expiretm": 0,
					"descr": {
						"pair": "string",
						"type": "buy",
						"ordertype": "market",
						"price": "string",
						"price2": "string",
						"leverage": "string",
						"order": "string",
						"close": "string"
					},
					"vol": "string",
					"vol_exec": "string",
					"cost": "string",
					"fee": "string",
					"price": "string",
					"stopprice": "string",
					"limitprice": "string",
					"trigger": "last",
					"misc": "string",
					"oflags": "string",
					"trades": [],
					"closetm": 0,
					"reason": "string"
				}
			},
			"count": 2
		}
	}`

	expResp := GetClosedOrdersResponse{
		KrakenAPIResponse: KrakenAPIResponse{Error: []string{}},
		Result: GetClosedOrdersResult{
			Closed: map[string]OrderInfo{
				"txid1": {
					ReferralOrderTransactionId: "string",
					UserReferenceId:            "string",
					Status:                     "pending",
					OpenTimestamp:              0,
					StartTimestamp:             0,
					ExpireTimestamp:            0,
					Description: OrderInfoDescription{
						Pair:                  "string",
						Type:                  "buy",
						OrderType:             "market",
						Price:                 "string",
						Price2:                "string",
						Leverage:              "string",
						OrderDescription:      "string",
						CloseOrderDescription: "string",
					},
					Volume:         "string",
					VolumeExecuted: "string",
					Cost:           "string",
					Fee:            "string",
					Price:          "string",
					StopPrice:      "string",
					LimitPrice:     "string",
					Trigger:        "last",
					Miscellaneous:  "string",
					OrderFlags:     "string",
					Trades:         []string{"string"},
					CloseTimestamp: 0,
					Reason:         "string",
				},
				"txid2": {
					ReferralOrderTransactionId: "string",
					UserReferenceId:            "string",
					Status:                     "pending",
					OpenTimestamp:              0,
					StartTimestamp:             0,
					ExpireTimestamp:            0,
					Description: OrderInfoDescription{
						Pair:                  "string",
						Type:                  "buy",
						OrderType:             "market",
						Price:                 "string",
						Price2:                "string",
						Leverage:              "string",
						OrderDescription:      "string",
						CloseOrderDescription: "string",
					},
					Volume:         "string",
					VolumeExecuted: "string",
					Cost:           "string",
					Fee:            "string",
					Price:          "string",
					StopPrice:      "string",
					LimitPrice:     "string",
					Trigger:        "last",
					Miscellaneous:  "string",
					OrderFlags:     "string",
					Trades:         []string{},
					CloseTimestamp: 0,
					Reason:         "string",
				},
			},
			Count: 2,
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetClosedOrders(&options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetClosedOrders)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), req.Form.Get("trades"))
	require.Equal(suite.T(), strconv.FormatInt(*options.UserReference, 10), req.Form.Get("userref"))
	require.Equal(suite.T(), strconv.FormatInt(options.Start.Unix(), 10), req.Form.Get("start"))
	require.Equal(suite.T(), strconv.FormatInt(options.End.Unix(), 10), req.Form.Get("end"))
	require.Equal(suite.T(), strconv.FormatInt(*options.Offset, 10), req.Form.Get("ofs"))
	require.Equal(suite.T(), string(options.Closetime), req.Form.Get("closetime"))

	// Check response
	require.Equal(suite.T(), expResp, *resp)
}

// TestGetClosedOrdersEmptyResponse Test will succeed if a valid, empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetClosedOrdersEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"closed": {},
			"count": 0
		}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetClosedOrders(nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Result.Closed)
	require.Equal(suite.T(), 0, resp.Result.Count)
}

// TestGetClosedOrdersErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetClosedOrdersErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetClosedOrders(nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestQueryOrdersInfoHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryOrdersInfoHappyPath() {

	// Test parameters
	params := QueryOrdersParameters{
		TransactionIds: []string{"txid1", "txid2"},
	}

	options := QueryOrdersOptions{
		Trades:        true,
		UserReference: new(int64),
	}

	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"txid1": {
				"refid": "string",
				"userref": "string",
				"status": "pending",
				"opentm": 0,
				"starttm": 0,
				"expiretm": 0,
				"descr": {
					"pair": "string",
					"type": "buy",
					"ordertype": "market",
					"price": "string",
					"price2": "string",
					"leverage": "string",
					"order": "string",
					"close": "string"
				},
				"vol": "string",
				"vol_exec": "string",
				"cost": "string",
				"fee": "string",
				"price": "string",
				"stopprice": "string",
				"limitprice": "string",
				"trigger": "last",
				"misc": "string",
				"oflags": "string",
				"trades": [
					"string"
				],
				"closetm": 0,
				"reason": "string"
			},
			"txid2": {
				"refid": "string",
				"userref": "string",
				"status": "pending",
				"opentm": 0,
				"starttm": 0,
				"expiretm": 0,
				"descr": {
					"pair": "string",
					"type": "buy",
					"ordertype": "market",
					"price": "string",
					"price2": "string",
					"leverage": "string",
					"order": "string",
					"close": "string"
				},
				"vol": "string",
				"vol_exec": "string",
				"cost": "string",
				"fee": "string",
				"price": "string",
				"stopprice": "string",
				"limitprice": "string",
				"trigger": "last",
				"misc": "string",
				"oflags": "string",
				"trades": [],
				"closetm": 0,
				"reason": "string"
			}
		}
	}`

	expData := QueryOrdersInfoResponse{
		Result: map[string]OrderInfo{
			"txid1": {
				ReferralOrderTransactionId: "string",
				UserReferenceId:            "string",
				Status:                     "pending",
				OpenTimestamp:              0,
				StartTimestamp:             0,
				ExpireTimestamp:            0,
				Description: OrderInfoDescription{
					Pair:                  "string",
					Type:                  "buy",
					OrderType:             "market",
					Price:                 "string",
					Price2:                "string",
					Leverage:              "string",
					OrderDescription:      "string",
					CloseOrderDescription: "string",
				},
				Volume:         "string",
				VolumeExecuted: "string",
				Cost:           "string",
				Fee:            "string",
				Price:          "string",
				StopPrice:      "string",
				LimitPrice:     "string",
				Trigger:        "last",
				Miscellaneous:  "string",
				OrderFlags:     "string",
				Trades:         []string{"string"},
				CloseTimestamp: 0,
				Reason:         "string",
			},
			"txid2": {
				ReferralOrderTransactionId: "string",
				UserReferenceId:            "string",
				Status:                     "pending",
				OpenTimestamp:              0,
				StartTimestamp:             0,
				ExpireTimestamp:            0,
				Description: OrderInfoDescription{
					Pair:                  "string",
					Type:                  "buy",
					OrderType:             "market",
					Price:                 "string",
					Price2:                "string",
					Leverage:              "string",
					OrderDescription:      "string",
					CloseOrderDescription: "string",
				},
				Volume:         "string",
				VolumeExecuted: "string",
				Cost:           "string",
				Fee:            "string",
				Price:          "string",
				StopPrice:      "string",
				LimitPrice:     "string",
				Trigger:        "last",
				Miscellaneous:  "string",
				OrderFlags:     "string",
				Trades:         []string{},
				CloseTimestamp: 0,
				Reason:         "string",
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.QueryOrdersInfo(params, &options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postQueryOrdersInfos)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), req.Form.Get("trades"))
	require.Equal(suite.T(), strconv.FormatInt(*options.UserReference, 10), req.Form.Get("userref"))
	require.Equal(suite.T(), strings.Join(params.TransactionIds, ","), req.Form.Get("txid"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// TestQueryOrdersInfoEmptyResponse Test will succeed if a valid, empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryOrdersInfoEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.QueryOrdersInfo(QueryOrdersParameters{}, nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Result)
}

// TestQueryOrdersInfoErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryOrdersInfoErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.QueryOrdersInfo(QueryOrdersParameters{}, nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestGetTradesHistoryHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradesHistoryHappyPath() {

	// Test parameters
	now := time.Now().UTC()
	offset := int64(0)
	options := GetTradesHistoryOptions{
		Type:   "all",
		Trades: true,
		Start:  &now,
		End:    &now,
		Offset: &offset,
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"trades": {
				"txid1": {
					"ordertxid": "string",
					"pair": "string",
					"time": 0,
					"type": "string",
					"ordertype": "string",
					"price": "string",
					"cost": "string",
					"fee": "string",
					"vol": "string",
					"margin": "string",
					"misc": "string",
					"posstatus": "string",
					"cprice": null,
					"ccost": null,
					"cfee": null,
					"cvol": null,
					"cmargin": null,
					"net": null,
					"trades": ["string"]
				},
				"txid2": {
					"ordertxid": "string",
					"pair": "string",
					"time": 0,
					"type": "string",
					"ordertype": "string",
					"price": "string",
					"cost": "string",
					"fee": "string",
					"vol": "string",
					"margin": "string",
					"misc": "string",
					"posstatus": "string",
					"cprice": null,
					"ccost": null,
					"cfee": null,
					"cvol": null,
					"cmargin": null,
					"net": null,
					"trades": []
				}
			},
			"count": 0
		}
	}`

	expData := GetTradesHistoryResult{
		Trades: map[string]TradeInfo{
			"txid1": {
				OrderTransactionId: "string",
				Pair:               "string",
				Timestamp:          0,
				Type:               "string",
				OrderType:          "string",
				Price:              "string",
				Cost:               "string",
				Fee:                "string",
				Volume:             "string",
				Margin:             "string",
				Miscellaneous:      "string",
				PositionStatus:     "string",
				ClosedPrice:        "",
				ClosedCost:         "",
				ClosedFee:          "",
				ClosedVolume:       "",
				ClosedMargin:       "",
				ClosedNetPNL:       "",
				ClosingTrades:      []string{"string"},
			},
			"txid2": {
				OrderTransactionId: "string",
				Pair:               "string",
				Timestamp:          0,
				Type:               "string",
				OrderType:          "string",
				Price:              "string",
				Cost:               "string",
				Fee:                "string",
				Volume:             "string",
				Margin:             "string",
				Miscellaneous:      "string",
				PositionStatus:     "string",
				ClosedPrice:        "",
				ClosedCost:         "",
				ClosedFee:          "",
				ClosedVolume:       "",
				ClosedMargin:       "",
				ClosedNetPNL:       "",
				ClosingTrades:      []string{},
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetTradesHistory(&options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetTradesHistory)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), options.Type, req.Form.Get("type"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), req.Form.Get("trades"))
	require.Equal(suite.T(), strconv.FormatInt(now.Unix(), 10), req.Form.Get("start"))
	require.Equal(suite.T(), strconv.FormatInt(now.Unix(), 10), req.Form.Get("end"))
	require.Equal(suite.T(), strconv.FormatInt(offset, 10), req.Form.Get("ofs"))

	// Check response
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetTradesHistoryEmptyResponse Test will succeed if a valid, empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradesHistoryEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetTradesHistory(nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Result.Trades)
}

// TestGetTradesHistoryErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradesHistoryErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetTradesHistory(nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestQueryTradesInfoHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryTradesInfoHappyPath() {

	// Test parameters
	params := QueryTradesParameters{
		TransactionIds: []string{"THVRQM-33VKH-UCI7BS", "OH76VO-UKWAD-PSBDX6"},
	}
	options := QueryTradesOptions{
		Trades: true,
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"THVRQM-33VKH-UCI7BS": {
				"ordertxid": "OQCLML-BW3P3-BUCMWZ",
				"pair": "XXBTZUSD",
				"time": 1616667796.8802,
				"type": "buy",
				"ordertype": "limit",
				"price": "30010.00000",
				"cost": "600.20000",
				"fee": "0.00000",
				"vol": "0.02000000",
				"margin": "0.00000",
				"misc": ""
				},
			"TTEUX3-HDAAA-RC2RUO": {
				"ordertxid": "OH76VO-UKWAD-PSBDX6",
				"pair": "XXBTZEUR",
				"time": 1614082549.3138,
				"type": "buy",
				"ordertype": "limit",
				"price": "1001.00000",
				"cost": "0.20020",
				"fee": "0.00000",
				"vol": "0.00020000",
				"margin": "0.00000",
				"misc": ""
			}
		}
	}`

	expData := QueryTradesInfoResponse{
		Result: map[string]TradeInfo{
			"THVRQM-33VKH-UCI7BS": {
				OrderTransactionId: "OQCLML-BW3P3-BUCMWZ",
				Pair:               "XXBTZUSD",
				Timestamp:          1616667796.8802,
				Type:               "buy",
				OrderType:          "limit",
				Price:              "30010.00000",
				Cost:               "600.20000",
				Fee:                "0.00000",
				Volume:             "0.02000000",
				Margin:             "0.00000",
				Miscellaneous:      "",
			},
			"TTEUX3-HDAAA-RC2RUO": {
				OrderTransactionId: "OH76VO-UKWAD-PSBDX6",
				Pair:               "XXBTZEUR",
				Timestamp:          1614082549.3138,
				Type:               "buy",
				OrderType:          "limit",
				Price:              "1001.00000",
				Cost:               "0.20020",
				Fee:                "0.00000",
				Volume:             "0.00020000",
				Margin:             "0.00000",
				Miscellaneous:      "",
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.QueryTradesInfo(params, &options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postQueryTradesInfo)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.ElementsMatch(suite.T(), params.TransactionIds, strings.Split(req.Form.Get("txid"), ","))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), req.Form.Get("trades"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// Test QueryTradesInfo Test will succeed if a valid, empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryTradesInfoEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.QueryTradesInfo(QueryTradesParameters{}, nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Result)
}

// TestQueryTradesInfoErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryTradesInfoErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.QueryTradesInfo(QueryTradesParameters{}, nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestGetOpenPositionsHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOpenPositionsHappyPath() {

	// Test parameters
	options := GetOpenPositionsOptions{
		TransactionIds: []string{"TF5GVO-T7ZZ2-6NBKBI", "T24DOR-TAFLM-ID3NYP"},
		DoCalcs:        true,
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"TF5GVO-T7ZZ2-6NBKBI": {
				"ordertxid": "OLWNFG-LLH4R-D6SFFP",
				"posstatus": "open",
				"pair": "XXBTZUSD",
				"time": 1605280097.8294,
				"type": "buy",
				"ordertype": "limit",
				"cost": "104610.52842",
				"fee": "289.06565",
				"vol": "8.82412861",
				"vol_closed": "0.20200000",
				"margin": "20922.10568",
				"value": "258797.5",
				"net": "+154186.9728",
				"terms": "0.0100% per 4 hours",
				"rollovertm": "1616672637",
				"misc": "",
				"oflags": ""
			},
			"T24DOR-TAFLM-ID3NYP": {
				"ordertxid": "OIVYGZ-M5EHU-ZRUQXX",
				"posstatus": "open",
				"pair": "XXBTZUSD",
				"time": 1607943827.3172,
				"type": "buy",
				"ordertype": "limit",
				"cost": "145756.76856",
				"fee": "335.24057",
				"vol": "8.00000000",
				"vol_closed": "0.00000000",
				"margin": "29151.35371",
				"value": "240124.0",
				"net": "+94367.2314",
				"terms": "0.0100% per 4 hours",
				"rollovertm": "1616672637",
				"misc": "",
				"oflags": ""
			}
		}
	}`

	expData := GetOpenPositionsResponse{
		Result: map[string]PositionInfo{
			"TF5GVO-T7ZZ2-6NBKBI": {
				OrderTransactionId: "OLWNFG-LLH4R-D6SFFP",
				PositionStatus:     "open",
				Pair:               "XXBTZUSD",
				Timestamp:          1605280097.8294,
				Type:               "buy",
				OrderType:          "limit",
				Cost:               "104610.52842",
				Fee:                "289.06565",
				Volume:             "8.82412861",
				ClosedVolume:       "0.20200000",
				Margin:             "20922.10568",
				Value:              "258797.5",
				Net:                "+154186.9728",
				Terms:              "0.0100% per 4 hours",
				RolloverTimestamp:  "1616672637",
				Miscellaneous:      "",
				OrderFlags:         "",
			},
			"T24DOR-TAFLM-ID3NYP": {
				OrderTransactionId: "OIVYGZ-M5EHU-ZRUQXX",
				PositionStatus:     "open",
				Pair:               "XXBTZUSD",
				Timestamp:          1607943827.3172,
				Type:               "buy",
				OrderType:          "limit",
				Cost:               "145756.76856",
				Fee:                "335.24057",
				Volume:             "8.00000000",
				ClosedVolume:       "0.00000000",
				Margin:             "29151.35371",
				Value:              "240124.0",
				Net:                "+94367.2314",
				Terms:              "0.0100% per 4 hours",
				RolloverTimestamp:  "1616672637",
				Miscellaneous:      "",
				OrderFlags:         "",
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetOpenPositions(&options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetOpenPositions)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.ElementsMatch(suite.T(), options.TransactionIds, strings.Split(req.Form.Get("txid"), ","))
	require.Equal(suite.T(), strconv.FormatBool(options.DoCalcs), req.Form.Get("docalcs"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// TestGetOpenPositionsEmptyResponse Test will succeed if a valid, empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOpenPositionsEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetOpenPositions(nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Result)
}

// TestGetOpenPositionsErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetOpenPositionsErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetOpenPositions(nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestGetLedgersInfoHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetLedgersInfoHappyPath() {

	// Test parameters
	now := time.Now().UTC()
	offset := int64(0)
	options := GetLedgersInfoOptions{
		Assets:     []string{"string"},
		AssetClass: "string",
		Type:       "all",
		Start:      &now,
		End:        &now,
		Offset:     &offset,
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"ledger": {
				"ledger_id1": {
					"refid": "string",
					"time": 0,
					"type": "trade",
					"subtype": "string",
					"aclass": "string",
					"asset": "string",
					"amount": "string",
					"fee": "string",
					"balance": "string"
				},
				"ledger_id2": {
					"refid": "string",
					"time": 0,
					"type": "trade",
					"subtype": "string",
					"aclass": "string",
					"asset": "string",
					"amount": "string",
					"fee": "string",
					"balance": "string"
				}
			},
			"count": 0				
		}
	}`

	expData := GetLedgersInfoResult{
		Ledgers: map[string]LedgerEntry{
			"ledger_id1": {
				ReferenceId: "string",
				Timestamp:   0,
				Type:        "trade",
				SubType:     "string",
				AssetClass:  "string",
				Asset:       "string",
				Amount:      "string",
				Fee:         "string",
				Balance:     "string",
			},
			"ledger_id2": {
				ReferenceId: "string",
				Timestamp:   0,
				Type:        "trade",
				SubType:     "string",
				AssetClass:  "string",
				Asset:       "string",
				Amount:      "string",
				Fee:         "string",
				Balance:     "string",
			},
		},
		Count: 0,
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetLedgersInfo(&options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetLedgersInfo)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.ElementsMatch(suite.T(), options.Assets, strings.Split(req.Form.Get("asset"), ","))
	require.Equal(suite.T(), options.AssetClass, req.Form.Get("aclass"))
	require.Equal(suite.T(), options.Type, req.Form.Get("type"))
	require.Equal(suite.T(), strconv.FormatInt(now.Unix(), 10), req.Form.Get("start"))
	require.Equal(suite.T(), strconv.FormatInt(now.Unix(), 10), req.Form.Get("end"))
	require.Equal(suite.T(), strconv.FormatInt(*options.Offset, 10), req.Form.Get("ofs"))

	// Check response
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetLedgersInfoEmptyResponse Test will succeed if a valid, empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetLedgersInfoEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"ledger": {},
			"count": 0
		}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetLedgersInfo(nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Result.Ledgers)
	require.Zero(suite.T(), resp.Result.Count)
}

// TestGetLedgersInfoErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetLedgersInfoErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetLedgersInfo(nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestQueryLedgersHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryLedgersHappyPath() {

	// Test parameters
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"ledger_id1": {
				"refid": "string",
				"time": 0,
				"type": "trade",
				"subtype": "string",
				"aclass": "string",
				"asset": "string",
				"amount": "string",
				"fee": "string",
				"balance": "string"
			},
			"ledger_id2": {
				"refid": "string",
				"time": 0,
				"type": "trade",
				"subtype": "string",
				"aclass": "string",
				"asset": "string",
				"amount": "string",
				"fee": "string",
				"balance": "string"
			}		
		}
	}`

	expData := QueryLedgersResponse{
		Result: map[string]LedgerEntry{
			"ledger_id1": {
				ReferenceId: "string",
				Timestamp:   0,
				Type:        "trade",
				SubType:     "string",
				AssetClass:  "string",
				Asset:       "string",
				Amount:      "string",
				Fee:         "string",
				Balance:     "string",
			},
			"ledger_id2": {
				ReferenceId: "string",
				Timestamp:   0,
				Type:        "trade",
				SubType:     "string",
				AssetClass:  "string",
				Asset:       "string",
				Amount:      "string",
				Fee:         "string",
				Balance:     "string",
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.QueryLedgers(QueryLedgersParameters{
		LedgerIds: []string{"1", "2"},
	},
		nil,
		&secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetLedgersInfo)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// TestQueryLedgersEmptyResponse Test will succeed if a valid, empty response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryLedgersEmptyResponse() {

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {}
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.QueryLedgers(QueryLedgersParameters{}, nil, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.NotNil(suite.T(), resp)
	require.Empty(suite.T(), resp.Result)
}

// TestQueryLedgersErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestQueryLedgersErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.QueryLedgers(QueryLedgersParameters{}, nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestGetTradeVolumeHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradeVolumeHappyPath() {

	// Test parameters
	params := GetTradeVolumeParameters{
		Pairs: []string{"pair1", "pair2"},
	}
	options := GetTradeVolumeOptions{
		FeeInfo: true,
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"currency": "string",
			"volume": "string",
			"fees": {
				"pair1": {
					"fee": "string",
					"min_fee": "string",
					"max_fee": "string",
					"next_fee": "string",
					"tier_volume": "string",
					"next_volume": "string"
				},
				"pair2": {
					"fee": "string",
					"min_fee": "string",
					"max_fee": "string",
					"next_fee": "string",
					"tier_volume": "string",
					"next_volume": "string"
				}
			},
			"fees_maker": {
				"pair1": {
					"fee": "string",
					"min_fee": "string",
					"max_fee": "string",
					"next_fee": "string",
					"tier_volume": "string",
					"next_volume": "string"
				},
				"pair2": {
					"fee": "string",
					"min_fee": "string",
					"max_fee": "string",
					"next_fee": "string",
					"tier_volume": "string",
					"next_volume": "string"
				}
			}				
		}
	}`

	expData := GetTradeVolumeResult{
		Currency: "string",
		Volume:   "string",
		Fees: map[string]FeeTierInfo{
			"pair1": {
				Fee:            "string",
				MinimumFee:     "string",
				MaximumFee:     "string",
				NextFee:        "string",
				TierVolume:     "string",
				NextTierVolume: "string",
			},
			"pair2": {
				Fee:            "string",
				MinimumFee:     "string",
				MaximumFee:     "string",
				NextFee:        "string",
				TierVolume:     "string",
				NextTierVolume: "string",
			},
		},
		FeesMaker: map[string]FeeTierInfo{
			"pair1": {
				Fee:            "string",
				MinimumFee:     "string",
				MaximumFee:     "string",
				NextFee:        "string",
				TierVolume:     "string",
				NextTierVolume: "string",
			},
			"pair2": {
				Fee:            "string",
				MinimumFee:     "string",
				MaximumFee:     "string",
				NextFee:        "string",
				TierVolume:     "string",
				NextTierVolume: "string",
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetTradeVolume(params, &options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetTradeVolume)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.ElementsMatch(suite.T(), params.Pairs, strings.Split(req.Form.Get("pair"), ","))
	require.Equal(suite.T(), strconv.FormatBool(options.FeeInfo), req.Form.Get("fee-info"))

	// Check response
	require.Equal(suite.T(), expData, resp.Result)
}

// TestGetTradeVolumeErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetTradeVolumeErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetTradeVolume(GetTradeVolumeParameters{}, nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestRequestExportReportHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestRequestExportReportHappyPath() {

	// Test parameters
	params := RequestExportReportParameters{
		Report:      "trades",
		Description: "testing",
	}
	now := time.Now().UTC()
	options := RequestExportReportOptions{
		Format:  "csv",
		Fields:  []string{"all"},
		StartTm: &now,
		EndTm:   &now,
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"id": "TCJA"				
		}
	}`

	expData := RequestExportReportResult{
		Id: "TCJA",
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.RequestExportReport(params, &options, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postRequestExportReport)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), params.Description, req.Form.Get("description"))
	require.Equal(suite.T(), params.Report, req.Form.Get("report"))
	require.Equal(suite.T(), options.Format, req.Form.Get("format"))
	require.ElementsMatch(suite.T(), options.Fields, strings.Split(req.Form.Get("fields"), ","))
	require.Equal(suite.T(), strconv.FormatInt(options.StartTm.Unix(), 10), req.Form.Get("starttm"))
	require.Equal(suite.T(), strconv.FormatInt(options.EndTm.Unix(), 10), req.Form.Get("endtm"))

	// Check response
	require.Equal(suite.T(), expData, resp.Result)
}

// TestRequestExportReportErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestRequestExportReportErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.RequestExportReport(RequestExportReportParameters{}, nil, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestGetExportReportStatusHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetExportReportStatusHappyPath() {

	// Test parameters
	params := GetExportReportStatusParameters{
		Report: "trades",
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": [
			{
				"id": "VSKC",
				"descr": "my_trades_1",
				"format": "CSV",
				"report": "trades",
				"subtype": "all",
				"status": "Processed",
				"fields": "all",
				"createdtm": "1616669085",
				"starttm": "1616669093",
				"completedtm": "1616669093",
				"datastarttm": "1614556800",
				"dataendtm": "1616669085",
				"asset": "all"
			},
			{
				"id": "TCJA",
				"descr": "my_trades_1",
				"format": "CSV",
				"report": "trades",
				"subtype": "all",
				"status": "Processed",
				"fields": "all",
				"createdtm": "1617363637",
				"starttm": "1617363664",
				"completedtm": "1617363664",
				"datastarttm": "1617235200",
				"dataendtm": "1617363637",
				"asset": "all"
			}
		]
	}`

	expData := GetExportReportStatusResponse{
		Result: []ExportReportStatus{
			{
				Id:                 "VSKC",
				Description:        "my_trades_1",
				Format:             "CSV",
				Report:             "trades",
				SubType:            "all",
				Status:             "Processed",
				Fields:             "all",
				RequestTimestamp:   "1616669085",
				StartTimestamp:     "1616669093",
				CompletedTimestamp: "1616669093",
				DataStartTimestamp: "1614556800",
				DataEndTimestamp:   "1616669085",
				Asset:              "all",
			},
			{
				Id:                 "TCJA",
				Description:        "my_trades_1",
				Format:             "CSV",
				Report:             "trades",
				SubType:            "all",
				Status:             "Processed",
				Fields:             "all",
				RequestTimestamp:   "1617363637",
				StartTimestamp:     "1617363664",
				CompletedTimestamp: "1617363664",
				DataStartTimestamp: "1617235200",
				DataEndTimestamp:   "1617363637",
				Asset:              "all",
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetExportReportStatus(params, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetExportReportStatus)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), params.Report, req.Form.Get("report"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// TestGetExportReportStatus Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetExportReportStatusErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetExportReportStatus(GetExportReportStatusParameters{}, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestRetrieveDataExportHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestRetrieveDataExportHappyPath() {

	// Test parameters
	params := RetrieveDataExportParameters{
		Id: "VSKC",
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedBytesResponse := []byte{78, 48, 80, 69}

	expData := RetrieveDataExportResponse{
		Report: []byte{78, 48, 80, 69},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/octet-stream"}},
		Body:    expectedBytesResponse,
	})

	// Call API endpoint
	resp, err := suite.client.RetrieveDataExport(params, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postRetrieveDataExport)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), params.Id, req.Form.Get("id"))

	// Check response
	require.Equal(suite.T(), expData, *resp)
}

// TestRetrieveDataExportErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestRetrieveDataExportErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.RetrieveDataExport(RetrieveDataExportParameters{}, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

// TestDeleteExportReportHappyPath Test will succeed if a valid response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestDeleteExportReportHappyPath() {

	// Test parameters
	params := DeleteExportReportParameters{
		Id:   "VSKC",
		Type: "delete",
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [ ],
		"result": {
			"delete": true
		}
	}`

	expData := DeleteExportReportResult{
		Delete: true,
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.DeleteExportReport(params, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postDeleteExportReport)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), params.Id, req.Form.Get("id"))
	require.Equal(suite.T(), params.Type, req.Form.Get("type"))

	// Check response
	require.Equal(suite.T(), expData, resp.Result)
}

// TestDeleteExportReportErrPath Test will succeed if a error response from server is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestDeleteExportReportErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.DeleteExportReport(DeleteExportReportParameters{}, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Nil(suite.T(), resp)
	require.Error(suite.T(), err)
}

/*****************************************************************************/
/* UNIT TESTS - USER TRADING												 */
/*****************************************************************************/

// Test Add Order Method - Happy path
//
// Test will succeed if all provided input parameters are in request sent by client
// and if predefined response from server is correctly processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderHappyPath() {

	// Test parameters
	pair := "XXBTZUSD"
	userref := new(int64)
	*userref = 42
	deadline := time.Now().UTC()
	validate := true
	otp := "Once"
	order := &Order{
		UserReference:      userref,
		OrderType:          OTypeLimit,
		Type:               Buy,
		Volume:             "2.1234",
		Price:              "45000.1",
		Price2:             "45000.1",
		Trigger:            TriggerLast,
		Leverage:           "2:1",
		StpType:            StpCancelNewest,
		OrderFlags:         strings.Join([]string{OFlagFeeInQuote, OFlagPost}, ","),
		TimeInForce:        GoodTilCanceled,
		ScheduledStartTime: "0",
		ExpirationTime:     "0",
		Close:              &CloseOrder{OrderType: OTypeStopLossLimit, Price: "38000.42", Price2: "36000"},
	}

	// Predefined server response
	expectedJSONResponse := `{
		"error": [ ],
		"result": {
			"descr": {
				"order": "buy 2.12340000 XBTUSD @ limit 45000.1 with 2:1 leverage",
				"close": "close position @ stop loss 38000.42 -> limit 36000.0"
			},
			"txid": [
				"OUF4EM-FRGI2-MQMWZD",
				"OUF4EM-FRGI2-MQMW42"
			]
		}
	}`

	// Expected response data
	expOrderDescr := "buy 2.12340000 XBTUSD @ limit 45000.1 with 2:1 leverage"
	expCloseDescr := "close position @ stop loss 38000.42 -> limit 36000.0"
	expTxID := []string{"OUF4EM-FRGI2-MQMWZD", "OUF4EM-FRGI2-MQMW42"}

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.AddOrder(
		AddOrderParameters{Pair: pair, Order: *order},
		&AddOrderOptions{Deadline: &deadline, Validate: validate},
		&SecurityOptions{SecondFactor: otp})

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check client error and log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postAddOrder)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))
	require.Equal(suite.T(), pair, req.Form.Get("pair"))
	actUserRef, err := strconv.ParseInt(req.Form.Get("userref"), 10, 64)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), *order.UserReference, actUserRef)
	require.Equal(suite.T(), order.OrderType, req.Form.Get("ordertype"))
	require.Equal(suite.T(), order.Type, req.Form.Get("type"))
	require.Equal(suite.T(), order.Volume, req.Form.Get("volume"))
	require.Equal(suite.T(), order.Price, req.Form.Get("price"))
	require.Equal(suite.T(), order.Price2, req.Form.Get("price2"))
	require.Equal(suite.T(), order.Trigger, req.Form.Get("trigger"))
	require.Equal(suite.T(), order.Leverage, req.Form.Get("leverage"))
	require.Equal(suite.T(), order.StpType, req.Form.Get("stp_type"))
	require.Equal(suite.T(), order.OrderFlags, req.Form.Get("oflags"))
	require.Equal(suite.T(), order.TimeInForce, req.Form.Get("timeinforce"))
	require.Equal(suite.T(), order.ScheduledStartTime, req.Form.Get("starttm"))
	require.Equal(suite.T(), order.ExpirationTime, req.Form.Get("expiretm"))
	require.Equal(suite.T(), order.Close.OrderType, req.Form.Get("close[ordertype]"))
	require.Equal(suite.T(), order.Close.Price, req.Form.Get("close[price]"))
	require.Equal(suite.T(), order.Close.Price2, req.Form.Get("close[price2]"))
	actValidate, err := strconv.ParseBool(req.Form.Get("validate"))
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), validate, actValidate)
	actDeadline, err := time.Parse(time.RFC3339, req.Form.Get("deadline"))
	require.NoError(suite.T(), err)
	// Nanoseconds are not provided
	require.Equal(suite.T(), deadline.Truncate(time.Second), actDeadline)

	// Check server response
	require.Equal(suite.T(), expOrderDescr, resp.Result.Description.Order)
	require.Equal(suite.T(), expCloseDescr, resp.Result.Description.Close)
	require.ElementsMatch(suite.T(), expTxID, resp.Result.TransactionIDs)
}

// Test Add Order Method - Empty parameters
//
// Test will succeed if a valid request is sent by client, if empty parameters are not included
// in request and if predefined response from server is correctly processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderEmptyParameters() {

	// Test parameters
	pair := "XXBTZUSD"
	order := &Order{
		OrderType: OTypeMarket,
		Type:      Buy,
		Volume:    "2.1234",
	}

	// Predefined server response
	expectedJSONResponse := `{
		"error": [ ],
		"result": {
			"descr": {
				"order": "buy 2.12340000 XBTUSD @ market"
			},
			"txid": [
				"OUF4EM-FRGI2-MQMWZD"
			]
		}
	}`

	// Expected response data
	expOrderDescr := "buy 2.12340000 XBTUSD @ market"
	expTxID := []string{"OUF4EM-FRGI2-MQMWZD"}

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.AddOrder(AddOrderParameters{Pair: pair, Order: *order}, nil, nil)

	// Get and log client request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check for client error & log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postAddOrder)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Empty(suite.T(), req.Form.Get("otp"))
	require.Equal(suite.T(), pair, req.Form.Get("pair"))
	require.Nil(suite.T(), order.UserReference)
	require.Empty(suite.T(), req.Form.Get("userref"))
	require.Equal(suite.T(), order.OrderType, req.Form.Get("ordertype"))
	require.Equal(suite.T(), order.Type, req.Form.Get("type"))
	require.Equal(suite.T(), order.Volume, req.Form.Get("volume"))
	require.Empty(suite.T(), order.Price)
	require.Empty(suite.T(), order.Price2)
	require.Empty(suite.T(), req.Form.Get("price"))
	require.Empty(suite.T(), req.Form.Get("price2"))
	require.Empty(suite.T(), order.Trigger)
	require.Empty(suite.T(), req.Form.Get("trigger"))
	require.Empty(suite.T(), order.Leverage)
	require.Empty(suite.T(), req.Form.Get("leverage"))
	require.Empty(suite.T(), order.StpType)
	require.Empty(suite.T(), req.Form.Get("stp_type"))
	require.Empty(suite.T(), order.OrderFlags)
	require.Empty(suite.T(), req.Form.Get("oflags"))
	require.Empty(suite.T(), order.TimeInForce)
	require.Empty(suite.T(), req.Form.Get("timeinforce"))
	require.Empty(suite.T(), order.ScheduledStartTime)
	require.Empty(suite.T(), req.Form.Get("starttm"))
	require.Empty(suite.T(), order.ExpirationTime)
	require.Empty(suite.T(), req.Form.Get("expiretm"))
	require.Nil(suite.T(), order.Close)
	require.Empty(suite.T(), req.Form.Get("close[ordertype]"))
	require.Empty(suite.T(), req.Form.Get("close[price]"))
	require.Empty(suite.T(), req.Form.Get("close[price2]"))
	require.Empty(suite.T(), req.Form.Get("validate"))
	require.Empty(suite.T(), req.Form.Get("deadline"))

	// Check server response
	require.Equal(suite.T(), expOrderDescr, resp.Result.Description.Order)
	require.Empty(suite.T(), resp.Result.Description.Close)
	require.ElementsMatch(suite.T(), expTxID, resp.Result.TransactionIDs)
}

// Test Add Order Method - Error path
//
// Test will succeed if API call fail because of invalid server response.
func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderErrPath() {

	// Test parameters
	pair := "XXBTZUSD"
	validate := true
	order := &Order{
		OrderType: OTypeMarket,
		Type:      Buy,
		Volume:    "2.1234",
	}

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Perform request
	_, err := suite.client.AddOrder(
		AddOrderParameters{Pair: pair, Order: *order},
		&AddOrderOptions{Deadline: nil, Validate: validate},
		nil)
	require.Error(suite.T(), err)
}

// Test Add Order Batch - Happy Path
//
// Test will succeed if a request sent by client is well formatted, contains
// all input parameters and if server response is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchHappyPath() {

	// Test parameters
	pair := "XXBTZUSD"
	userref := new(int64)
	*userref = 123
	deadline := time.Now().UTC()
	validate := true
	otp := "otp"
	orders := []Order{
		{
			OrderType:   OTypeLimit,
			Type:        Buy,
			Volume:      "1.2",
			Price:       "40000",
			Price2:      "40000.42",
			Trigger:     TriggerIndex,
			StpType:     StpCancelNewest,
			TimeInForce: GoodTilCanceled,
			Close: &CloseOrder{
				OrderType: OTypeStopLossLimit,
				Price:     "37000",
				Price2:    "36000",
			},
		},
		{
			UserReference:      userref,
			OrderType:          OTypeLimit,
			Type:               Sell,
			Volume:             "1.2",
			Price:              "42000",
			Leverage:           "2:1",
			OrderFlags:         "fciq,post",
			ScheduledStartTime: "0",
			ExpirationTime:     "0",
			Close:              nil,
		},
	}

	// Predefined response
	expectedJSONResponse := `{
		"error": [ ],
		"result": {
			"orders": [
				{
					"descr": {
						"order": "buy 1.2 BTCUSD @ limit 40000",	
						"close": "close position @ stop loss 37000.0 -> limit 36000.0"
					},
					"txid": "OUF4EM-FRGI2-MQMWZD"
				},
				{
					"descr": {
						"order": "sell 1.2 BTCUSD @ limit 42000"
					},
					"txid": ["OCF5EM-FRGI2-MQWEDD", "OUF4EM-FRGI2-MQMWZD"]
				}
			]
		}
	}`

	expResp := &AddOrderBatchResponse{}
	err := json.Unmarshal([]byte(expectedJSONResponse), expResp)
	require.NoError(suite.T(), err)

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.AddOrderBatch(
		AddOrderBatchParameters{Pair: pair, Orders: orders},
		&AddOrderBatchOptions{Deadline: &deadline, Validate: validate},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Get and log client request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check client error and log response
	require.NoError(suite.T(), err)

	// Check request
	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postAddOrderBatch)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))
	require.Equal(suite.T(), pair, req.Form.Get("pair"))
	require.Equal(suite.T(), orders[0].OrderType, req.Form.Get("orders[0][ordertype]"))
	require.Equal(suite.T(), orders[0].Type, req.Form.Get("orders[0][type]"))
	require.Equal(suite.T(), orders[0].Volume, req.Form.Get("orders[0][volume]"))
	require.Equal(suite.T(), orders[0].Price, req.Form.Get("orders[0][price]"))
	require.Equal(suite.T(), orders[0].Price2, req.Form.Get("orders[0][price2]"))
	require.Equal(suite.T(), orders[0].Trigger, req.Form.Get("orders[0][trigger]"))
	require.Equal(suite.T(), orders[0].StpType, req.Form.Get("orders[0][stp_type]"))
	require.Equal(suite.T(), orders[0].TimeInForce, req.Form.Get("orders[0][timeinforce]"))
	require.Equal(suite.T(), orders[0].Close.OrderType, req.Form.Get("orders[0][close][ordertype]"))
	require.Equal(suite.T(), orders[0].Close.Price, req.Form.Get("orders[0][close][price]"))
	require.Equal(suite.T(), orders[0].Close.Price2, req.Form.Get("orders[0][close][price2]"))
	require.Empty(suite.T(), req.Form.Get("orders[0][userref]"))
	require.Empty(suite.T(), req.Form.Get("orders[0][leverage]"))
	require.Empty(suite.T(), req.Form.Get("orders[0][oflags]"))
	require.Empty(suite.T(), req.Form.Get("orders[0][starttm]"))
	require.Empty(suite.T(), req.Form.Get("orders[0][expiretm]"))
	require.Equal(suite.T(), orders[1].OrderType, req.Form.Get("orders[1][ordertype]"))
	require.Equal(suite.T(), orders[1].Type, req.Form.Get("orders[1][type]"))
	require.Equal(suite.T(), orders[1].Volume, req.Form.Get("orders[1][volume]"))
	require.Equal(suite.T(), orders[1].Price, req.Form.Get("orders[1][price]"))
	require.Empty(suite.T(), req.Form.Get("orders[1][price2]"))
	require.Empty(suite.T(), req.Form.Get("orders[1][trigger]"))
	require.Empty(suite.T(), req.Form.Get("orders[1][stp_type]"))
	require.Empty(suite.T(), req.Form.Get("orders[1][timeinforce]"))
	require.Empty(suite.T(), req.Form.Get("orders[1][close][ordertype]"))
	require.Empty(suite.T(), req.Form.Get("orders[1][close][price]"))
	require.Empty(suite.T(), req.Form.Get("orders[1][close][price2]"))
	require.Equal(suite.T(), strconv.FormatInt(*orders[1].UserReference, 10), req.Form.Get("orders[1][userref]"))
	require.Equal(suite.T(), orders[1].Leverage, req.Form.Get("orders[1][leverage]"))
	require.Equal(suite.T(), orders[1].OrderFlags, req.Form.Get("orders[1][oflags]"))
	require.Equal(suite.T(), orders[1].ScheduledStartTime, req.Form.Get("orders[1][starttm]"))
	require.Equal(suite.T(), orders[1].ExpirationTime, req.Form.Get("orders[1][expiretm]"))
	actValidate, err := strconv.ParseBool(req.Form.Get("validate"))
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), validate, actValidate)
	actDeadline, err := time.Parse(time.RFC3339, req.Form.Get("deadline"))
	require.NoError(suite.T(), err)
	// Nanoseconds are not provided
	require.Equal(suite.T(), deadline.Truncate(time.Second), actDeadline)

	// Check response
	require.Equal(suite.T(), expResp, resp)
}

// Test Add Order Batch - Empty orders
//
// Test will succeed if an error is thrown by client when
// an empty list of orders is submitted.
func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchEmptyOrders() {

	// Perform request
	_, err := suite.client.AddOrderBatch(AddOrderBatchParameters{}, nil, nil)
	require.Error(suite.T(), err)

	// Ensure no request has been sent
	req := suite.srv.PopRecordedRequest()
	require.Nil(suite.T(), req)
}

// Test Add Order Batch - Partial success
//
// Test will succeed if a response with some failed orders is well processed
// by client. Response should contain non-failed orders and a non nil error.
func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchPartialSuccess() {

	// Test parameters
	pair := "XXBTZUSD"
	validate := false
	orders := []Order{
		{
			OrderType: OTypeMarket,
			Type:      Buy,
			Volume:    "100000000.00000000",
		},
		{
			OrderType: OTypeMarket,
			Type:      Sell,
			Volume:    "100000000.00000000",
		},
	}

	// Predefined server response
	expectedJSONResponse := `{
		"error":[],
		"result":{
			"orders":[
				{
					"error":"EOrder:Insufficient funds",
					"descr":{
						"order":"buy 100000000.00000000 XBTEUR @ market"
					}
				},
				{
					"error":["EOrder:Insufficient funds", "wrong"],
					"descr":{
						"order":"sell 100000000.00000000 XBTEUR @ market"
					}
				}
			]
		}
	}`

	expResp := &AddOrderBatchResponse{}
	err := json.Unmarshal([]byte(expectedJSONResponse), expResp)
	require.NoError(suite.T(), err)

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, _ := suite.client.AddOrderBatch(
		AddOrderBatchParameters{Pair: pair, Orders: orders},
		&AddOrderBatchOptions{Deadline: nil, Validate: validate},
		nil)

	// Get and log request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check response
	require.Equal(suite.T(), expResp, resp)
}

// Test Add Order Batch - Error path
//
// Test will succeed if an error is returned when an invalid response is
// received from server.
func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchErrPath() {

	// Test parameters
	pair := "XXBTZUSD"
	validate := false
	orders := []Order{
		{
			OrderType: OTypeMarket,
			Type:      Buy,
			Volume:    "100000000.00000000",
		},
		{
			OrderType: OTypeMarket,
			Type:      Sell,
			Volume:    "100000000.00000000",
		},
	}

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Perform request
	resp, err := suite.client.AddOrderBatch(
		AddOrderBatchParameters{Pair: pair, Orders: orders},
		&AddOrderBatchOptions{Deadline: nil, Validate: validate},
		nil)

	// Check client error and response
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// Test Edit Order - Happy Path
//
// Test will succeed if request sent by client is well formatted, contains all input
// all input parameters and if server response is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestEditOrderHappyPath() {

	// Test parameters
	pair := "XXBTZUSD"
	userref := new(int64)
	*userref = 123
	originalTxID := "OHYO67-6LP66-HMQ437"
	volume := "0.00030000"
	price := "19500.0"
	price2 := "32500.0"
	oflags := []string{"fcib", "post"}
	deadline := time.Now().UTC()
	cancelResponse := true
	validate := true
	otp := "otp"

	// Predefined response
	expectedJSONResponse := `{
			"error": [ ],
			"result": {
				"status": "ok",
				"txid": "OFVXHJ-KPQ3B-VS7ELA",
				"originaltxid": "OHYO67-6LP66-HMQ437",
				"volume": "0.00030000",
				"price": "19500.0",
				"price2": "32500.0",
				"orders_cancelled": 1,
				"descr": {
				"order": "buy 0.00030000 XXBTZGBP @ limit 19500.0"
			}
		}
	}`

	// Expected response data
	expResp := &EditOrderResponse{}
	err := json.Unmarshal([]byte(expectedJSONResponse), &expResp)
	require.NoError(suite.T(), err)

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.EditOrder(
		EditOrderParameters{Pair: pair, Id: originalTxID},
		&EditOrderOptions{
			NewUserReference: strconv.FormatInt(*userref, 10),
			NewVolume:        volume,
			Price:            price,
			Price2:           price2,
			OFlags:           oflags,
			Deadline:         &deadline,
			CancelResponse:   cancelResponse,
			Validate:         validate,
		},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Get and log client request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check client error and log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postEditOrder)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))
	require.Equal(suite.T(), pair, req.Form.Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(*userref, 10), req.Form.Get("userref"))
	require.Equal(suite.T(), originalTxID, req.Form.Get("txid"))
	require.Equal(suite.T(), volume, req.Form.Get("volume"))
	require.Equal(suite.T(), price, req.Form.Get("price"))
	require.Equal(suite.T(), price2, req.Form.Get("price2"))
	require.Equal(suite.T(), strings.Join(oflags, ","), req.Form.Get("oflags"))
	actCancelResponse, err := strconv.ParseBool(req.Form.Get("cancel_response"))
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), cancelResponse, actCancelResponse)
	actValidate, err := strconv.ParseBool(req.Form.Get("validate"))
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), validate, actValidate)
	actDeadline, err := time.Parse(time.RFC3339, req.Form.Get("deadline"))
	require.NoError(suite.T(), err)
	// Nanoseconds are not provided
	require.Equal(suite.T(), deadline.Truncate(time.Second), actDeadline)

	// Check response
	require.Equal(suite.T(), expResp, resp)
}

// Test Cancel Order - Happy Path
//
// Test will succeed if request sent by client is well formatted, contains all input
// all input parameters and if server response is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestCancelOrderHappyPath() {

	// Test parameters
	txID := "OHYO67-6LP66-HMQ437"
	otp := "otp"

	// Predefined response
	expectedJSONResponse := `{
		"result": {
			"count": 0,
			"pending": true
		},
		"error": []
	}`

	// Expected response data
	expCount := 0
	expPending := true

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.CancelOrder(CancelOrderParameters{Id: txID}, &SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Get and log client request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check client error and log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postCancelOrder)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))
	require.Equal(suite.T(), txID, req.Form.Get("txid"))

	// Check response
	require.Equal(suite.T(), expCount, resp.Result.Count)
	require.Equal(suite.T(), expPending, resp.Result.Pending)
}

// Test Cancel All Orders - Happy Path
//
// Test will succeed if request sent by client is well formatted, contains all input
// all input parameters and if server response is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestCancelAllOrdersHappyPath() {

	// Test parameters
	otp := "otp"

	// Predefined response
	expectedJSONResponse := `{
		"result": {
			"count": 4
		},
		"error": []
	}`

	// Expected response data
	expCount := 4

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.CancelAllOrders(&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Get and log client request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check client error and log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postCancelAllOrders)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expCount, resp.Result.Count)
}

// Test Cancel All Orders After X - Happy Path
//
// Test will succeed if request sent by client is well formatted, contains all input
// all input parameters and if server response is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestCancelAllOrdersAfterXHappyPath() {

	// Test parameters
	timeout := int64(60)
	otp := "otp"

	// Predefined response
	expectedJSONResponse := `{
		"result": {
			"currentTime": "2021-03-24T17:41:56Z",
			"triggerTime": "2021-03-24T17:42:56Z"
		},
		"error": []
	}`

	// Expected response data
	expYear := 2021
	expMonth := time.Month(3)
	expDayOfMonth := 24
	expHour := 17
	expSecond := 56
	expNanosec := 0
	expCurrMinute := 41
	expTrigMinute := 42

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.CancelAllOrdersAfterX(CancelCancelAllOrdersAfterXParameters{Timeout: timeout}, &SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Get and log client request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check client error and log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postCancelAllOrdersAfterX)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatInt(timeout, 10), req.Form.Get("timeout"))

	// Check response
	require.Equal(suite.T(), expYear, resp.Result.CurrentTime.Year())
	require.Equal(suite.T(), expYear, resp.Result.TriggerTime.Year())
	require.Equal(suite.T(), expMonth, resp.Result.CurrentTime.Month())
	require.Equal(suite.T(), expMonth, resp.Result.TriggerTime.Month())
	require.Equal(suite.T(), expDayOfMonth, resp.Result.CurrentTime.Day())
	require.Equal(suite.T(), expDayOfMonth, resp.Result.TriggerTime.Day())
	require.Equal(suite.T(), expHour, resp.Result.CurrentTime.Hour())
	require.Equal(suite.T(), expHour, resp.Result.TriggerTime.Hour())
	require.Equal(suite.T(), expCurrMinute, resp.Result.CurrentTime.Minute())
	require.Equal(suite.T(), expTrigMinute, resp.Result.TriggerTime.Minute())
	require.Equal(suite.T(), expSecond, resp.Result.CurrentTime.Second())
	require.Equal(suite.T(), expSecond, resp.Result.TriggerTime.Second())
	require.Equal(suite.T(), expNanosec, resp.Result.CurrentTime.Nanosecond())
	require.Equal(suite.T(), expNanosec, resp.Result.TriggerTime.Nanosecond())
}

// Test Cancel Order Batch - Happy Path
//
// Test will succeed if request sent by client is well formatted, contains all input
// all input parameters and if server response is well processed by client.
func (suite *KrakenAPIClientUnitTestSuite) TestCancelOrderBatchHappyPath() {

	// Test parameters
	orders := []string{"42", "43"}
	otp := "otp"

	// Predefined response
	expectedJSONResponse := `{
		"result": {
			"count": 2
		},
		"error": []
	}`

	// Expected response data
	expCount := 2

	// Configure Mock HTTP Server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Perform request
	resp, err := suite.client.CancelOrderBatch(CancelOrderBatchParameters{OrderIds: orders}, &SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Get and log client request
	req := suite.srv.PopRecordedRequest()
	require.NotNil(suite.T(), req)
	suite.T().Log(req)

	// Check client error and log response
	require.NoError(suite.T(), err)
	suite.T().Log(resp)

	// Check request
	// Check URL, Method and some Headers
	require.Contains(suite.T(), req.URL.Path, postCancelOrderBatch)
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

	// Check request body
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))
	reqOrders := []string{
		req.Form.Get("orders[0]"),
		req.Form.Get("orders[1]"),
	}
	require.ElementsMatch(suite.T(), orders, reqOrders)

	// Check response
	require.Equal(suite.T(), expCount, resp.Result.Count)
}

// Test Cancel Order Batch - Empty list
//
// Test will succeed if request is rejected by client.
func (suite *KrakenAPIClientUnitTestSuite) TestCancelOrderBatchEmptyList() {

	// Perform request
	_, err := suite.client.CancelOrderBatch(CancelOrderBatchParameters{}, nil)
	require.Error(suite.T(), err)

	// Check no request sent
	require.Nil(suite.T(), suite.srv.PopRecordedRequest())
}

/*****************************************************************************/
/* UNIT TESTS - USER FUNDING												 */
/*****************************************************************************/

// Test Get Deposit Methods - Empty limit
func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositMethodsEmptyLimit() {

	// Test parameters
	asset := "XXBT"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": [
			{
				"method": "Bitcoin Lightning",
				"limit": false,
				"fee": "0.000000000"
			}
		]
	}`

	expectedMethod := "Bitcoin Lightning"
	expectedFee := "0.000000000"
	expectedAddrSetupFee := ""
	expectedGenAddr := ""
	expectedLimit := "false"

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetDepositMethods(GetDepositMethodsParameters{Asset: asset}, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body contains asset
	require.Equal(suite.T(), asset, req.Form.Get("asset"))

	// Check response
	require.Equal(suite.T(), 1, len(resp.Result))
	require.Equal(suite.T(), expectedMethod, (resp.Result)[0].Method)
	require.Equal(suite.T(), expectedLimit, (resp.Result)[0].Limit)
	require.Equal(suite.T(), expectedFee, (resp.Result)[0].Fee)
	require.Equal(suite.T(), expectedAddrSetupFee, (resp.Result)[0].AddressSetupFee)
	require.Equal(suite.T(), expectedGenAddr, (resp.Result)[0].GenAddress)
}

// Test Get Deposit Methods - Predefined limit
func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositMethodsFloatLimit() {

	// Test parameters
	asset := "XXBT"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": [
			{
				"method": "Bitcoin",
				"limit": "342.42",
				"fee": "4",
				"address-setup-fee": "1.2",
				"gen-address": true
			}
		]
	}`

	expectedLimit := "342.42"
	expectedMethod := "Bitcoin"
	expectedFee := "4"
	expectedAddrSetupFee := "1.2"
	expectedGenAddr := "true"

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetDepositMethods(GetDepositMethodsParameters{Asset: asset}, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body contains asset
	require.Equal(suite.T(), asset, req.Form.Get("asset"))

	// Check response
	require.Equal(suite.T(), 1, len(resp.Result))
	require.Equal(suite.T(), expectedMethod, (*resp).Result[0].Method)
	require.Equal(suite.T(), expectedLimit, (*resp).Result[0].Limit)
	require.Equal(suite.T(), expectedFee, (*resp).Result[0].Fee)
	require.Equal(suite.T(), expectedAddrSetupFee, (*resp).Result[0].AddressSetupFee)
	require.Equal(suite.T(), expectedGenAddr, (*resp).Result[0].GenAddress)
}

// Test Get Deposit Methods - No limit field
func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositMethodsNoLimit() {

	// Test parameters
	asset := "XXBT"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": [
			{
				"method": "Bitcoin",
				"fee": "4",
				"address-setup-fee": "1.2",
				"gen-address": true
			}
		]
	}`

	expectedMethod := "Bitcoin"
	expectedFee := "4"
	expectedAddrSetupFee := "1.2"
	expectedGenAddr := "true"

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetDepositMethods(GetDepositMethodsParameters{Asset: asset}, nil)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body contains asset
	require.Equal(suite.T(), asset, req.Form.Get("asset"))

	// Check response
	require.Equal(suite.T(), 1, len(resp.Result))
	require.Equal(suite.T(), expectedMethod, (*resp).Result[0].Method)
	require.Empty(suite.T(), (*resp).Result[0].Limit)
	require.Equal(suite.T(), expectedFee, (*resp).Result[0].Fee)
	require.Equal(suite.T(), expectedAddrSetupFee, (*resp).Result[0].AddressSetupFee)
	require.Equal(suite.T(), expectedGenAddr, (*resp).Result[0].GenAddress)
}

// Test Get Deposit Addresses - Happy path
func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositAddressesHappyPath() {

	// Test parameters
	asset := "XXBT"
	method := "Bitcoin"
	new := true
	otp := "NOPE"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": [
			{
				"address": "2N9fRkx5JTWXWHmXzZtvhQsufvoYRMq9ExV",
				"expiretm": "0",
				"new": true
			},
			{
				"address": "2NCpXUCEYr8ur9WXM1tAjZSem2w3aQeTcAo",
				"expiretm": "1658736768",
				"new": false
			},
			{
				"address": "2Myd4eaAW96ojk38A2uDK4FbioCayvkEgVq",
				"expiretm": "0"
			}
		]
	}`

	expectedAddressesLen := 3

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetDepositAddresses(
		GetDepositAddressesParameters{Asset: asset, Method: method},
		&GetDepositAddressesOptions{New: new},
		&SecurityOptions{SecondFactor: otp})

	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body
	require.Equal(suite.T(), asset, req.Form.Get("asset"))
	require.Equal(suite.T(), method, req.Form.Get("method"))
	require.Equal(suite.T(), strconv.FormatBool(new), req.Form.Get("new"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expectedAddressesLen, len(resp.Result))
	for i, v := range resp.Result {
		require.NotEmpty(suite.T(), v.Address)
		require.GreaterOrEqual(suite.T(), v.Expiretm, int64(0))
		if i == 0 {
			require.True(suite.T(), v.New)
		} else {
			require.False(suite.T(), v.New)
		}
	}
}

// Test Get Status of Recent Deposits - Happy path
func (suite *KrakenAPIClientUnitTestSuite) TestGetStatusOfRecentDepositsHappyPath() {

	// Test parameters
	asset := "XXBT"
	method := "Bitcoin"
	otp := "Nope"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": [
			{
				"method": "Bitcoin",
				"aclass": "currency",
				"asset": "XXBT",
				"refid": "AGBSO6T-UFMTTQ-I7KGS6",
				"txid": "AGBSO6T-UFMTTQ-I7KGS6",
				"info": "SEPA",
				"amount": "1.42",
				"fee": null,
				"time": 1658736768,
				"status": "Initial",
				"status-prop": "return"
			}
		]
	}`

	expectedAClass := "currency"
	expectedRefID := "AGBSO6T-UFMTTQ-I7KGS6"
	expectedTxID := "AGBSO6T-UFMTTQ-I7KGS6"
	expectedInfo := "SEPA"
	expectedAmount := "1.42"
	expectedFee := ""
	expectedTime := int64(1658736768)
	expectedStatus := TxStateInitial
	expectedStatusProp := TxStatusReturn
	expectedAddressesLen := 1

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetStatusOfRecentDeposits(
		GetStatusOfRecentDepositsParameters{Asset: asset},
		&GetStatusOfRecentDepositsOptions{Method: method},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body
	require.Equal(suite.T(), asset, req.Form.Get("asset"))
	require.Equal(suite.T(), method, req.Form.Get("method"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expectedAddressesLen, len(resp.Result))
	for _, v := range resp.Result {
		require.Equal(suite.T(), method, v.Method)
		require.Equal(suite.T(), expectedAClass, v.AssetClass)
		require.Equal(suite.T(), asset, v.Asset)
		require.Equal(suite.T(), expectedRefID, v.ReferenceID)
		require.Equal(suite.T(), expectedTxID, v.TransactionID)
		require.Equal(suite.T(), expectedInfo, v.Info)
		require.Equal(suite.T(), expectedAmount, v.Amount)
		require.Equal(suite.T(), expectedFee, v.Fee)
		require.Equal(suite.T(), expectedTime, v.Time)
		require.Equal(suite.T(), expectedStatus, v.Status)
		require.Equal(suite.T(), expectedStatusProp, v.StatusProperty)
	}
}

// Test Get Withdrawal Information - Happy path
func (suite *KrakenAPIClientUnitTestSuite) TestGetWithdrawalInformationHappyPath() {

	// Test parameters
	asset := "XXBT"
	amount := "42.999999"
	key := "withdrawal_address"
	otp := "Nope"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"method": "Bitcoin",
			"limit": "332.00956139",
			"amount": "42.999999",
			"fee": "0.00015000"
		}
	}`

	expectedFee := "0.00015000"
	expectedLimit := "332.00956139"
	expectedMethod := "Bitcoin"

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetWithdrawalInformation(
		GetWithdrawalInformationParameters{Asset: asset, Amount: amount, Key: key},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body
	require.Equal(suite.T(), asset, req.Form.Get("asset"))
	require.Equal(suite.T(), amount, req.Form.Get("amount"))
	require.Equal(suite.T(), key, req.Form.Get("key"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expectedMethod, resp.Result.Method)
	require.Equal(suite.T(), amount, resp.Result.Amount)
	require.Equal(suite.T(), expectedFee, resp.Result.Fee)
	require.Equal(suite.T(), expectedLimit, resp.Result.Limit)
}

// Test Withdraw Funds - Happy path
func (suite *KrakenAPIClientUnitTestSuite) TestWithdrawFundsHappyPath() {

	// Test parameters
	asset := "XXBT"
	waddrn := "nevermind"
	amount := "42.999999"
	otp := "NOPE"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"refid": "AGBSO6T-UFMTTQ-I7KGS6"
		}
	}`

	expectedRefID := "AGBSO6T-UFMTTQ-I7KGS6"

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.WithdrawFunds(
		WithdrawFundsParameters{Asset: asset, Amount: amount, Key: waddrn},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body
	require.Equal(suite.T(), asset, req.Form.Get("asset"))
	require.Equal(suite.T(), amount, req.Form.Get("amount"))
	require.Equal(suite.T(), waddrn, req.Form.Get("key"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expectedRefID, resp.Result.ReferenceID)
}

// Test Get Status of Recent Withdrawal - Happy path
func (suite *KrakenAPIClientUnitTestSuite) TestGetStatusOfRecentWithdrawalsHappyPath() {

	// Test parameters
	asset := "XXBT"
	method := "Bitcoin"
	otp := "NOPE"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": [
			{
				"method": "Bitcoin",
				"aclass": "currency",
				"asset": "XXBT",
				"refid": "AGBZNBO-5P2XSB-RFVF6J",
				"txid": "THVRQM-33VKH-UCI7BS",
				"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
				"amount": "0.72485000",
				"fee": "0.00015000",
				"time": 1617014586,
				"status": "Pending"
			},
			{
				"method": "Bitcoin",
				"aclass": "currency",
				"asset": "XXBT",
				"refid": "AGBSO6T-UFMTTQ-I7KGS6",
				"txid": "KLETXZ-33VKH-UCI7BS",
				"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
				"amount": "0.72485000",
				"fee": "0.00015000",
				"time": 1617015423,
				"status": "Failure",
				"status-prop": "canceled"
			}
		]
	}`

	expectedResultLen := 2
	expectedAClass := "currency"
	expectedInfo := "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR"
	expectedAmount := "0.72485000"
	expectedFee := "0.00015000"

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetStatusOfRecentWithdrawals(
		GetStatusOfRecentWithdrawalsParameters{Asset: asset},
		&GetStatusOfRecentWithdrawalsOptions{Method: method},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body
	require.Equal(suite.T(), asset, req.Form.Get("asset"))
	require.Equal(suite.T(), method, req.Form.Get("method"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expectedResultLen, len(resp.Result))
	for _, v := range resp.Result {
		require.Equal(suite.T(), method, v.Method)
		require.Equal(suite.T(), expectedAClass, v.AssetClass)
		require.Equal(suite.T(), asset, v.Asset)
		require.NotEmpty(suite.T(), v.ReferenceID)
		require.NotEmpty(suite.T(), v.TransactionID)
		require.Equal(suite.T(), expectedInfo, v.Info)
		require.Equal(suite.T(), expectedAmount, v.Amount)
		require.Equal(suite.T(), expectedFee, v.Fee)
		require.GreaterOrEqual(suite.T(), v.Time, int64(0))
		require.True(suite.T(), v.Status == TxStatePending || v.Status == TxStateFailure)
		require.True(suite.T(), v.StatusProperty == "" || v.StatusProperty == TxCanceled)
	}
}

// Test Request Withdrawal Cancellation - Happy path
func (suite *KrakenAPIClientUnitTestSuite) TestRequestWithdrawalCancellationHappyPath() {

	// Test parameters
	asset := "XXBT"
	refid := "AGBZNBO-5P2XSB-RFVF6J"
	otp := "NOPE"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": true
	}`

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.RequestWithdrawalCancellation(
		RequestWithdrawalCancellationParameters{Asset: asset, ReferenceId: refid},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body
	require.Equal(suite.T(), asset, req.Form.Get("asset"))
	require.Equal(suite.T(), refid, req.Form.Get("refid"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.True(suite.T(), resp.Result)
}

// Test Request Wallet Transfer - Happy path
func (suite *KrakenAPIClientUnitTestSuite) TestRequestWalletTransferHappyPath() {

	// Test parameters
	asset := "XXBT"
	from := "Spot Wallet"
	to := "Future Wallet"
	amount := "42.24"
	otp := "NOPE"

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
			"refid": "BOG5AE5-KSCNR4-VPNPEV"
		}
	}`

	expectedRefID := "BOG5AE5-KSCNR4-VPNPEV"

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.RequestWalletTransfer(
		RequestWalletTransferParameters{Asset: asset, From: from, To: to, Amount: amount},
		&SecurityOptions{SecondFactor: otp})
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request body
	require.Equal(suite.T(), asset, req.Form.Get("asset"))
	require.Equal(suite.T(), from, req.Form.Get("from"))
	require.Equal(suite.T(), to, req.Form.Get("to"))
	require.Equal(suite.T(), amount, req.Form.Get("amount"))
	require.Equal(suite.T(), otp, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expectedRefID, resp.Result.ReferenceID)
}

/*****************************************************************************/
/* USER STAKING TESTS                                                        */
/*****************************************************************************/

// TestStakeAssetHappyPath is a unit test for Stake Asset. The test will succeed
// if client send a valid request and if a valid predefined response from server
// is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestStakeAssetHappyPath() {

	// Test params
	params := StakeAssetParameters{
		Asset:  "XXBT",
		Amount: "0.01",
		Method: "offchain",
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Server response
	expectedJSONResponse := `
	{
		"error": [ ],
		"result": {
		"refid": "BOG5AE5-KSCNR4-VPNPEV"
		}
	}`

	expResp := &StakeAssetResponse{
		KrakenAPIResponse: KrakenAPIResponse{Error: []string{}},
		Result: struct {
			ReferenceID string "json:\"refid\""
		}{ReferenceID: "BOG5AE5-KSCNR4-VPNPEV"},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.StakeAsset(params, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postStakeAsset)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, req.Form.Get("asset"))
	require.Equal(suite.T(), params.Amount, req.Form.Get("amount"))
	require.Equal(suite.T(), params.Method, req.Form.Get("method"))

	// Check response
	require.Equal(suite.T(), expResp, resp)
}

// TestStakeAssetErrPath is a unit test for Stake Asset. The test will succeed
// if an error response from server is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestStakeAssetErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.StakeAsset(StakeAssetParameters{}, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestUnstakeAssetHappyPath is a unit test for Unstake Asset. The test will succeed
// if client send a valid request and if a valid predefined response from server
// is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestUnstakeAssetHappyPath() {

	// Test params
	params := UnstakeAssetParameters{
		Asset:  "XXBT",
		Amount: "0.01",
	}
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Server response
	expectedJSONResponse := `
	{
		"error": [ ],
		"result": {
		"refid": "BOG5AE5-KSCNR4-VPNPEV"
		}
	}`

	expResp := &UnstakeAssetResponse{
		KrakenAPIResponse: KrakenAPIResponse{Error: []string{}},
		Result: struct {
			ReferenceID string "json:\"refid\""
		}{ReferenceID: "BOG5AE5-KSCNR4-VPNPEV"},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.UnstakeAsset(params, &secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postUnstakeAsset)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, req.Form.Get("asset"))
	require.Equal(suite.T(), params.Amount, req.Form.Get("amount"))

	// Check response
	require.Equal(suite.T(), expResp, resp)
}

// TestUnstakeAssetErrPath is a unit test for Unstake Asset. The test will succeed
// if an error response from server is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestUnstakeAssetErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.UnstakeAsset(UnstakeAssetParameters{}, nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestListOfStakeableAssetsHappyPath is a unit test for List Of Stakeable Assets.
// Test will succeed if client send a valid request and if a valid predefined
// response from server is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestListOfStakeableAssetsHappyPath() {

	// Test params
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Server response
	expectedJSONResponse := `
	{
		"result": [
			{
				"method": "fake",
				"asset": "FAKE",
				"staking_asset": "FAKE.S",
				"rewards": {
			  		"reward": "99.95",
			  		"type": "percentage"
				},
				"on_chain": false,
				"can_stake": false,
				"can_unstake": false,
				"minimum_amount": {
				  	"staking": "0.0000000000",
				  	"unstaking": "0.0000000000"
				},
				"lock": {
					"staking": {
						"days": 0.5,
						"percentage": 42.42
					},
					"unstaking": {
						"days": 0.5,
						"percentage": 42.42
					},
					"lockup": {
						"days": 0.5,
						"percentage": 42.42
					}
				},
				"enabled_for_user": false,
				"disabled": true
		  	},
		  	{
				"method": "kusama-staked",
				"asset": "KSM",
				"staking_asset": "KSM.S",
				"rewards": {
				 	"reward": "12.00",
			  		"type": "percentage"
				}
		  	},
			{
				"method": "nope",
				"asset": "NOPE",
				"staking_asset": "NOPE.S",
				"rewards": {
				 	"reward": "12.00",
			  		"type": "percentage"
				},
				"minimum_amount": {},
				"lock": {}
		  	}
		],
		"error": []
	}`

	expData := ListOfStakeableAssetsResponse{
		Result: []StakingAssetInformation{
			{
				Asset:        "FAKE",
				StakingAsset: "FAKE.S",
				Method:       "fake",
				OnChain:      false,
				CanStake:     false,
				CanUnstake:   false,
				MinAmount: &StakingAssetMinAmount{
					Unstaking: "0.0000000000",
					Staking:   "0.0000000000",
				},
				Lock: &StakingAssetLockup{
					Unstaking: &StakingAssetLockPeriod{
						Days:       0.5,
						Percentage: 42.42,
					},
					Staking: &StakingAssetLockPeriod{
						Days:       0.5,
						Percentage: 42.42,
					},
					Lockup: &StakingAssetLockPeriod{
						Days:       0.5,
						Percentage: 42.42,
					},
				},
				EnabledForUser: false,
				Disabled:       true,
				Rewards: StakingAssetReward{
					Reward: "99.95",
					Type:   "percentage",
				},
			},
			{
				Asset:          "KSM",
				StakingAsset:   "KSM.S",
				Method:         "kusama-staked",
				OnChain:        true,
				CanStake:       true,
				CanUnstake:     true,
				MinAmount:      nil,
				Lock:           nil,
				EnabledForUser: true,
				Disabled:       false,
				Rewards: StakingAssetReward{
					Reward: "12.00",
					Type:   "percentage",
				},
			},
			{
				Asset:        "NOPE",
				StakingAsset: "NOPE.S",
				Method:       "nope",
				OnChain:      true,
				CanStake:     true,
				CanUnstake:   true,
				MinAmount: &StakingAssetMinAmount{
					Unstaking: "0",
					Staking:   "0",
				},
				Lock:           &StakingAssetLockup{},
				EnabledForUser: true,
				Disabled:       false,
				Rewards: StakingAssetReward{
					Reward: "12.00",
					Type:   "percentage",
				},
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.ListOfStakeableAssets(&secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postListOfStakeableAssets)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// TestListOfStakeableAssetsErrPath is a unit test for List Of Stakeable Assets. Test
// will succeed if an error response from server is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestListOfStakeableAssetsErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.ListOfStakeableAssets(nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestGetPendingStackingTransactionsHappyPath is a unit test for Get Pending Stacking
// Transactions. Test will succeed if client sends a valid request and if a valid server
// response is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetPendingStackingTransactionsHappyPath() {

	// Test params
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Server response
	expectedJSONResponse := `
	{
		"result": [
			{
				"method": "ada-staked",
				"aclass": "currency",
				"asset": "ADA.S",
				"refid": "RUSB7W6-ESIXUX-K6PVTM",
				"amount": "0.34844300",
				"fee": "0.00000000",
				"time": 1622967367,
				"status": "Initial",
				"type": "bonding"
		 	},
		  	{
				"method": "xtz-staked",
				"aclass": "currency",
				"asset": "XTZ.S",
				"refid": "RUCXX7O-6MWQBO-CQPGAX",
				"amount": "0.00746900",
				"fee": "0.00000000",
				"time": 1623074402,
				"status": "Initial",
				"type": "bonding"
		  	}
		],
		"error": []
	  }`

	expData := GetPendingStakingTransactionsResponse{
		Result: []StakingTransactionInfo{
			{
				ReferenceId: "RUSB7W6-ESIXUX-K6PVTM",
				Asset:       "ADA.S",
				AssetClass:  "currency",
				Type:        "bonding",
				Method:      "ada-staked",
				Amount:      "0.34844300",
				Fee:         "0.00000000",
				Timestamp:   1622967367,
				Status:      "Initial",
			},
			{
				ReferenceId: "RUCXX7O-6MWQBO-CQPGAX",
				Asset:       "XTZ.S",
				AssetClass:  "currency",
				Type:        "bonding",
				Method:      "xtz-staked",
				Amount:      "0.00746900",
				Fee:         "0.00000000",
				Timestamp:   1623074402,
				Status:      "Initial",
			},
		},
	}

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.GetPendingStakingTransactions(&secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postGetPendingStakingTransactions)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// TestGetPendingStackingTransactionsErrPath is a unit test for Get Pending Stacking
// Transactions. Test will succeed if an error response from server is well handled
// by client.
func (suite *KrakenAPIClientUnitTestSuite) TestGetPendingStackingTransactionsErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.GetPendingStakingTransactions(nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

// TestListOfStackingTransactionsHappyPath is a unit test for List Of Stacking Transactions.
// Test will succeed if client sends a valid request and if a valid server response is well
// handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestListOfStackingTransactionsHappyPath() {

	// Test params
	secopts := SecurityOptions{
		SecondFactor: "NOPE",
	}

	// Server response
	expectedJSONResponse := `
	{
		"result": [
			{
				"method": "ada-staked",
				"aclass": "currency",
				"asset": "ADA.S",
				"refid": "RUSB7W6-ESIXUX-K6PVTM",
				"amount": "0.34844300",
				"fee": "0.00000000",
				"time": 1622967367,
				"status": "Initial",
				"type": "bonding",
				"bond_start": 1622971496,
				"bond_end": 1622971496
		 	},
		  	{
				"method": "xtz-staked",
				"aclass": "currency",
				"asset": "XTZ.S",
				"refid": "RUCXX7O-6MWQBO-CQPGAX",
				"amount": "0.00746900",
				"fee": "0.00000000",
				"time": 1623074402,
				"status": "Initial",
				"type": "bonding"
		  	}
		],
		"error": []
	  }`

	expData := ListOfStakingTransactionsResponse{
		Result: []StakingTransactionInfo{
			{
				ReferenceId: "RUSB7W6-ESIXUX-K6PVTM",
				Asset:       "ADA.S",
				AssetClass:  "currency",
				Type:        "bonding",
				Method:      "ada-staked",
				Amount:      "0.34844300",
				Fee:         "0.00000000",
				Timestamp:   1622967367,
				Status:      "Initial",
				BondStart:   new(int64),
				BondEnd:     new(int64),
			},
			{
				ReferenceId: "RUCXX7O-6MWQBO-CQPGAX",
				Asset:       "XTZ.S",
				AssetClass:  "currency",
				Type:        "bonding",
				Method:      "xtz-staked",
				Amount:      "0.00746900",
				Fee:         "0.00000000",
				Timestamp:   1623074402,
				Status:      "Initial",
				BondStart:   nil,
				BondEnd:     nil,
			},
		},
	}
	*expData.Result[0].BondStart = 1622971496
	*expData.Result[0].BondEnd = 1622971496

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Call API endpoint
	resp, err := suite.client.ListOfStakingTransactions(&secopts)
	require.NoError(suite.T(), err)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Log response
	suite.T().Logf("Response decoded by client : %#v", resp)

	// Check request
	require.Equal(suite.T(), http.MethodPost, req.Method)
	require.Contains(suite.T(), req.URL.Path, postListOfStakingTransactions)
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

	// Check response
	require.Equal(suite.T(), expData.Result, resp.Result)
}

// TestListOfStackingTransactionsErrPath is a unit test for List Of Stacking Transactions.
// Test will succeed if an error response from server is well handled by client.
func (suite *KrakenAPIClientUnitTestSuite) TestListOfStackingTransactionsErrPath() {

	// Configure mock http server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: http.StatusBadRequest,
	})

	// Call API endpoint
	resp, err := suite.client.ListOfStakingTransactions(nil)

	// Log request
	req := suite.srv.PopRecordedRequest()
	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

	// Check response
	require.Error(suite.T(), err)
	require.Nil(suite.T(), resp)
}

/*****************************************************************************/
/* GENERAL API CLIENT TESTS                                                  */
/*****************************************************************************/

// Test if client implents client interface
func (suite *KrakenAPIClientUnitTestSuite) TestClientInterfaceImplemented() {

	iface := reflect.TypeOf((*KrakenAPIClientIface)(nil)).Elem()
	ok := reflect.TypeOf(suite.client).Implements(iface)
	require.True(suite.T(), ok, "KranAPIClient does not fully implement KrakenAPIClient interface")
}

// Test when the client receives a 503 HTTP Error from server
func (suite *KrakenAPIClientUnitTestSuite) TestClientReceivesServiceUnavailableError() {

	// Expected error status
	expectedStatus := http.StatusServiceUnavailable

	// Configure server
	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
		Status: expectedStatus,
	})

	// Call API endpoint
	_, err := suite.client.GetSystemStatus()
	require.Error(suite.T(), err)
}

// Test when the client call a non existing server
func (suite *KrakenAPIClientUnitTestSuite) TestClientCallNotExistingServer() {

	// Create client to non-existing endpoint
	client := NewPublicWithOptions(&KrakenAPIClientOptions{BaseURL: "http://localhost:42422"})

	// Call API endpoint
	_, err := client.GetSystemStatus()
	require.Error(suite.T(), err)
}

// Test when the client experience a request timeout
func (suite *KrakenAPIClientUnitTestSuite) TestClientRequestTimeout() {

	// Create client with a timeout of 1 nanosecond
	client := NewPublicWithOptions(&KrakenAPIClientOptions{
		BaseURL: suite.srv.GetMockHTTPServerBaseURL(),
		Client:  &http.Client{Timeout: time.Duration(1)},
	})

	// Call API endpoint
	_, err := client.GetSystemStatus()
	require.Error(suite.T(), err)
}

/*****************************************************************************/
/* UTILITY FUNCTION TESTS													 */
/*****************************************************************************/

// Test the method used to forge a signature for a request
func (suite *KrakenAPIClientUnitTestSuite) TestRequestSignature() {

	// Signature parameters
	secret, _ := base64.StdEncoding.DecodeString("kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg==")
	nonce := "1616492376594"
	resource := "/0/private/AddOrder"
	encodedPayload := make(url.Values)
	encodedPayload.Set("nonce", nonce)
	encodedPayload.Set("ordertype", "limit")
	encodedPayload.Set("pair", "XBTUSD")
	encodedPayload.Set("price", "37500")
	encodedPayload.Set("type", "buy")
	encodedPayload.Set("volume", "1.25")

	// Expected signature - from documentation
	// https://docs.kraken.com/rest/#section/Authentication/Headers-and-Signature
	expected := "4/dpxb3iT4tp/ZCVEwSnEsLxx0bqyhLpdfOpc6fn7OR8+UClSV5n9E6aSS8MPtnRfp32bAb0nmbRn6H8ndwLUQ=="

	// Forge & compare signature
	require.Equal(suite.T(), expected, GetKrakenSignature(resource, encodedPayload, secret))
}
