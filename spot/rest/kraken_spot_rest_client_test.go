package rest

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gbdevw/gosette"
	"github.com/gbdevw/purple-goctopus/spot/rest/account"
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
	"github.com/gbdevw/purple-goctopus/spot/rest/trading"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* UNIT TEST SUITE                                                                               */
/*************************************************************************************************/

// Test constants
const (
	// API key used to sign requests
	apiKey = "API_KEY"
	// API key secret used to sign requests
	secretB64 = "kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg=="
	// User-Agent value for test
	usrAgent = "TST"
)

// Unit test suite for NewKrakenSpotRESTClient
type KrakenSpotRESTClientTestSuite struct {
	suite.Suite
	// Mock HTTP server
	srv *gosette.HTTPTestServer
	// Kraken API client configured to use mock HTTP server
	client *KrakenSpotRESTClient
}

// Configure and run unit test suite
func TestKrakenSpotRESTClientTestSuite(t *testing.T) {

	// Test server with default httptest.Server
	tstsrv := gosette.NewHTTPTestServer(nil)
	// Start the test server - Need this because the server base url is set only when server starts
	tstsrv.Start()
	defer tstsrv.Close()
	// Build authorizer with secret from the API documentation and tracing disabled
	authorizer, err := WithInstrumentedAuthorizer(apiKey, secretB64, nil)
	if err != nil {
		panic(err)
	}
	// Build Kraken client with :
	//	- The test server base url as base url
	//	- A used defined value for the USer-Agent header (TST)
	//	- A retryable http client as http client to use
	httpclient := retryablehttp.NewClient()
	httpclient.RetryWaitMax = 1 * time.Second
	httpclient.RetryWaitMin = 1 * time.Second
	httpclient.RetryMax = 3
	httpclient.Logger = log.New(io.Discard, "", 0) // Silent debug logs
	client := NewKrakenSpotRESTClient(authorizer, &KrakenSpotRESTClientConfiguration{
		BaseURL: tstsrv.GetBaseURL(),
		Agent:   usrAgent,
		Client:  httpclient.StandardClient(),
	})
	// Run unit test suite
	suite.Run(t, &KrakenSpotRESTClientTestSuite{
		Suite:  suite.Suite{},
		srv:    tstsrv,
		client: client,
	})
}

// Clean the server predefined responses and records before each test.
func (suite *KrakenSpotRESTClientTestSuite) BeforeTest(suiteName, testName string) {
	// Clear responses & requests from test server
	suite.srv.Clear()
}

/*************************************************************************************************/
/* UNIT TESTS - UTILITIES                                                                        */
/*************************************************************************************************/

// Test EncodeNonceAndSecurityOptions helper function.
//
// Test will verify provided nonce and security options are encoded as expected in the provided
// form data
func (suite *KrakenSpotRESTClientTestSuite) TestEncodeNonceAndSecurityOptions() {
	// Expectecations
	expectedNonce := int64(1)
	expectedOtp := "otp"
	// Build empty form data
	form := url.Values{}
	// Encode form data with the helper function
	EncodeNonceAndSecurityOptions(form, expectedNonce, &common.SecurityOptions{SecondFactor: expectedOtp})
	// Verify form data
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), form.Get("nonce"))
	require.Equal(suite.T(), expectedOtp, form.Get("otp"))
}

// Test forgeAndAuthorizeKrakenAPIRequest method when an authorizer is set.
//
// Test will ensure the returned HTTP request is configured as expected and contains the expected
// authorization data.
func (suite *KrakenSpotRESTClientTestSuite) TestForgeAndAuthorizeKrakenAPIRequestWithAuthorizer() {
	// Expectations
	expectedHttpMethod := http.MethodPost
	expectedFormData := url.Values{
		"nonce":     []string{"1616492376594"},
		"ordertype": []string{"limit"},
		"pair":      []string{"XBTUSD"},
		"price":     []string{"37500"},
		"type":      []string{"buy"},
		"volume":    []string{"1.25"},
	}
	expectedEncodedFormData := "nonce=1616492376594&ordertype=limit&pair=XBTUSD&price=37500&type=buy&volume=1.25"
	require.Equal(suite.T(), expectedEncodedFormData, expectedFormData.Encode())
	expectedSignature := "4/dpxb3iT4tp/ZCVEwSnEsLxx0bqyhLpdfOpc6fn7OR8+UClSV5n9E6aSS8MPtnRfp32bAb0nmbRn6H8ndwLUQ=="
	expectedPath := "/0/private/AddOrder" // /0 is added as the client base url is set to the test server base url
	// Forge request
	req, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		expectedPath,
		expectedHttpMethod,
		"application/x-www-form-urlencoded",
		nil,
		strings.NewReader(expectedFormData.Encode()))
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), req)
	// Check forged request method, path and headers
	require.Equal(suite.T(), expectedHttpMethod, req.Method)
	require.Equal(suite.T(), expectedPath, req.URL.Path)
	require.Equal(suite.T(), expectedSignature, req.Header[managedHeaderAPISign][0])
	require.Equal(suite.T(), apiKey, req.Header[managedHeaderAPIKey][0])
	require.Equal(suite.T(), usrAgent, req.Header.Get(managedHeaderUserAgent))
	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
	require.Equal(suite.T(), expectedFormData.Encode(), req.PostForm.Encode())
}

// Test forgeAndAuthorizeKrakenAPIRequest method when no authorizer is set.
//
// Test will ensure the returned HTTP request is configured as expected and does not contain
// authorization data.
func (suite *KrakenSpotRESTClientTestSuite) TestForgeAndAuthorizeKrakenAPIRequestWithoutAuthorizer() {
	// Expectations
	expectedHttpMethod := http.MethodGet
	expectedQueryStringAsset := "XBT,ETH"
	expectedQueryString := url.Values{
		"asset": []string{expectedQueryStringAsset},
	}
	expectedPath := "/0/public/Assets"
	// Create new client without any authorizer and with default options
	client := NewKrakenSpotRESTClient(nil, nil)
	// Forge request
	req, err := client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		"/public/Assets",
		expectedHttpMethod,
		"",
		expectedQueryString,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), req)
	// Check forged request method, path
	require.Equal(suite.T(), expectedHttpMethod, req.Method)
	require.Equal(suite.T(), expectedPath, req.URL.Path)
	// Parse form data (as they are not parsed by the authorizer) and check on them
	require.NoError(suite.T(), req.ParseForm())
	require.Equal(suite.T(), expectedQueryString.Encode(), req.Form.Encode())
	// Check request headers
	require.Empty(suite.T(), req.Header[managedHeaderAPISign])
	require.Empty(suite.T(), req.Header[managedHeaderAPIKey])
	require.Equal(suite.T(), DefaultUserAgent, req.Header.Get(managedHeaderUserAgent))
	require.Empty(suite.T(), req.Header.Get(managedHeaderContentType))
}

// Test forgeAndAuthorizeKrakenAPIRequest method when wrong inputs lead to a malformed request.
//
// Test will ensure the method returns an error and no request when it fails to create the http.Request.
func (suite *KrakenSpotRESTClientTestSuite) TestForgeAndAuthorizeKrakenAPIRequestWithMalformedInputs() {
	// Forge request
	req, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		"",
		"application/x-www-form-urlencoded", // Use "application/x-www-form-urlencoded" as method to trigger the expected error
		"",
		nil,
		nil)
	require.Error(suite.T(), err)
	require.Nil(suite.T(), req)
	require.Contains(suite.T(), err.Error(), `net/http: invalid method "application/x-www-form-urlencoded"`)
}

// Test doKrakenAPIRequest method with a valid request that will be sent to the test server. Test
// server will be configured to reply with a valid JSON response.
//
// Test will ensure:
//   - Request is sent by the client
//   - Recorded request on the test server side matches the expected request settings (path, method, ...)
//   - Valid JSON response is parsed by the client and populates the provided receiver.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithJsonResponse() {
	// Expected response
	expectedResponseBody := `
	{
		"error": [],
		"result": {
		  "descr": {
			"order": "buy 1.25000000 XBTUSD @ limit 27500.0"
		  },
		  "txid": [
			"OU22CG-KLAF2-FWUDD7"
		  ]
		}
	}`
	expectedStatusCode := http.StatusOK
	expectedContentType := "application/json"
	expectedOrderDescr := "buy 1.25000000 XBTUSD @ limit 27500.0"
	expectedOrderTxId := []string{"OU22CG-KLAF2-FWUDD7"}
	// Configure test server response
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status: expectedStatusCode,
		Headers: map[string][]string{
			"Content-Type": {expectedContentType},
		},
		Body: []byte(expectedResponseBody),
	})
	// Request expectations
	expectedHttpMethod := http.MethodPost
	expectedFormData := url.Values{
		"nonce":     []string{"1616492376594"},
		"ordertype": []string{"limit"},
		"pair":      []string{"XBTUSD"},
		"price":     []string{"37500"},
		"type":      []string{"buy"},
		"volume":    []string{"1.25"},
	}
	expectedEncodedFormData := "nonce=1616492376594&ordertype=limit&pair=XBTUSD&price=37500&type=buy&volume=1.25"
	require.Equal(suite.T(), expectedEncodedFormData, expectedFormData.Encode())
	expectedSignature := "4/dpxb3iT4tp/ZCVEwSnEsLxx0bqyhLpdfOpc6fn7OR8+UClSV5n9E6aSS8MPtnRfp32bAb0nmbRn6H8ndwLUQ=="
	expectedPath := "/0/private/AddOrder" // /0 is added as the client base url is set to the test server base url
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		expectedPath,
		expectedHttpMethod,
		"application/x-www-form-urlencoded",
		nil,
		strings.NewReader(expectedFormData.Encode()))
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request
	receiver := new(trading.AddOrderResponse)
	resp, err := suite.client.doKrakenAPIRequest(context.Background(), baseReq, receiver)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
	// Check response
	require.Equal(suite.T(), expectedStatusCode, resp.StatusCode)
	require.Equal(suite.T(), expectedContentType, resp.Header.Get("Content-Type"))
	// Check body is closed
	_, err = resp.Body.Read(make([]byte, 0))
	require.Error(suite.T(), err)
	// Check receiver
	require.Equal(suite.T(), expectedOrderDescr, receiver.Result.Description.Order)
	require.ElementsMatch(suite.T(), expectedOrderTxId, receiver.Result.TransactionIDs)
	// Pop server record & check recorded request
	record := suite.srv.PopServerRecord()
	require.Equal(suite.T(), expectedHttpMethod, record.Request.Method)
	require.Equal(suite.T(), expectedPath, record.Request.URL.Path)
	// Check signature headers -> WARN: Recorded request has the headers in their canonical form
	require.Equal(suite.T(), expectedSignature, record.Request.Header.Get("Api-Sign"))
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key"))
	require.Equal(suite.T(), usrAgent, record.Request.Header.Get(managedHeaderUserAgent))
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get(managedHeaderContentType))
	// Check recorded request body
	recBody, err := io.ReadAll(record.RequestBody)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), expectedEncodedFormData, string(recBody))
}

// Test doKrakenAPIRequest method with a valid request that will be sent to the test server. Test
// server will be configured to reply with a valid binary response (application/octet-stream).
//
// Test will ensure:
//   - Request is sent by the client
//   - Recorded request on the test server side matches the expected request settings (path, method, ...)
//   - Response body is not closed and contains the response data.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithBinaryStreamResponse() {
	// Expected response
	expectedResponse := "hello"
	expectedStatusCode := http.StatusOK
	expectedContentType := "application/octet-stream"
	// Configure test server response
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status: expectedStatusCode,
		Headers: map[string][]string{
			"Content-Type": {expectedContentType},
		},
		Body: []byte(expectedResponse),
	})
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		"",
		http.MethodGet,
		"",
		nil,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request
	resp, err := suite.client.doKrakenAPIRequest(context.Background(), baseReq, nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), expectedResponse, string(body))
}

// Test doKrakenAPIRequest method with a valid request that will be sent to the test server. Test
// server will be configured to reply with a valid binary response (application/zip).
//
// Test will ensure:
//   - Request is sent by the client
//   - Recorded request on the test server side matches the expected request settings (path, method, ...)
//   - Response body is not closed and contains the response data.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithBinaryZipResponse() {
	// Expected response
	expectedResponse := "hello"
	expectedStatusCode := http.StatusOK
	expectedContentType := "application/zip"
	// Configure test server response
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status: expectedStatusCode,
		Headers: map[string][]string{
			"Content-Type": {expectedContentType},
		},
		Body: []byte(expectedResponse),
	})
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		"",
		http.MethodGet,
		"",
		nil,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request
	resp, err := suite.client.doKrakenAPIRequest(context.Background(), baseReq, nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), expectedResponse, string(body))
}

// Test doKrakenAPIRequest method when called with an expired context.
//
// Test will ensure an error is returned in such a case. Test will also ensure the request
// is not sent if the context has expired.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithCanceledContext() {
	// Configure expired context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(), // Do not provide the canceled context -> will fail
		serverTimePath,
		http.MethodGet,
		"",
		nil,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request and expect an error
	receiver := new(market.GetServerTimeResponse)
	_, err = suite.client.doKrakenAPIRequest(ctx, baseReq, receiver)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "aborting request")
	// Check no request has been received by the test server
	require.Nil(suite.T(), suite.srv.PopServerRecord())
}

// Test doKrakenAPIRequest method when the Do method of the underlying httpClient fails.
//
// Test will ensure an error is returned in such a case.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithHttpClientDoError() {
	// Configure test server response
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status: http.StatusServiceUnavailable,
	})
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		serverTimePath,
		http.MethodGet,
		"",
		nil,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request and expect an error
	receiver := new(market.GetServerTimeResponse)
	_, err = suite.client.doKrakenAPIRequest(context.Background(), baseReq, receiver)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "failed to process HTTP request")
}

// Test doKrakenAPIRequest method when the response contains an status code other than OK.
//
// Test will ensure an error is returned in such a case.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithUnexpectedStatusCode() {
	// Configure test server response
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status: http.StatusAccepted,
	})
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		serverTimePath,
		http.MethodGet,
		"",
		nil,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request and expect an error
	receiver := new(market.GetServerTimeResponse)
	_, err = suite.client.doKrakenAPIRequest(context.Background(), baseReq, receiver)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "unexpected status code received from Kraken API")
}

// Test doKrakenAPIRequest method when the method fails to parse response content-type
//
// Test will ensure an error is returned if the Content-Type cannot be parsed.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithParseContentTypeFail() {
	// Configure test server response
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {""},
		},
	})
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		serverTimePath,
		http.MethodGet,
		"",
		nil,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request and expect an error
	receiver := new(market.GetServerTimeResponse)
	_, err = suite.client.doKrakenAPIRequest(context.Background(), baseReq, receiver)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "could not decode the response Content-Type header")
}

// Test doKrakenAPIRequest method when response has an unexpected content-type.
//
// Test will ensure an error is returned in such as case.
func (suite *KrakenSpotRESTClientTestSuite) TestDoKrakenAPIRequestWithWronContentType() {
	// Configure test server response
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"text/plain"},
		},
	})
	// Forge request
	baseReq, err := suite.client.forgeAndAuthorizeKrakenAPIRequest(
		context.Background(),
		serverTimePath,
		http.MethodGet,
		"",
		nil,
		nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), baseReq)
	// Do request and expect an error
	receiver := new(market.GetServerTimeResponse)
	_, err = suite.client.doKrakenAPIRequest(context.Background(), baseReq, receiver)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "response Content-Type is")
}

/*************************************************************************************************/
/* UNIT TESTS - MARKET DATA                                                                      */
/*************************************************************************************************/

// Test GetServerTime when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetServerTime() {

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

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetServerTime(context.Background())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Equal(suite.T(), expUnixTime, resp.Result.Unixtime)
	require.Equal(suite.T(), expRFC1123, resp.Result.Rfc1123)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, serverTimePath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
}

// Test GetSystemStatus when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetSystemStatus() {

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

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetSystemStatus(context.Background())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Equal(suite.T(), expStatus, resp.Result.Status)
	require.Equal(suite.T(), expTimestamp, resp.Result.Timestamp)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, systemStatusPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
}

// Test GetAssetInfo when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetAssetInfo() {

	// Test parameters
	options := &market.GetAssetInfoRequestOptions{
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

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetAssetInfo(context.Background(), options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	for _, asset := range options.Assets {
		require.NotEmpty(suite.T(), resp.Result[asset].Altname)
	}

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, assetInfoPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())

	// Check request query string
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), options.AssetClass, record.Request.URL.Query().Get("aclass"))
	require.Equal(suite.T(), strings.Join(options.Assets, ","), record.Request.URL.Query().Get("asset"))
}

// Test GetTradableAssetPairs when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetTradableAssetPairs() {

	// Test parameters
	options := &market.GetTradableAssetPairsRequestOptions{
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

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetTradableAssetPairs(context.Background(), options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	for _, pair := range options.Pairs {
		require.NotEmpty(suite.T(), resp.Result[pair].AlternativeName)
	}

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, tradableAssetPairsPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())

	// Check request query string
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), options.Info, record.Request.URL.Query().Get("info"))
	require.Equal(suite.T(), strings.Join(options.Pairs, ","), record.Request.URL.Query().Get("pair"))
}

// Test GetTickerInformation when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetTickerInformation() {

	// Test parameters
	opts := &market.GetTickerInformationRequestOptions{
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

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetTickerInformation(context.Background(), opts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	for _, pair := range opts.Pairs {
		require.NotEmpty(suite.T(), resp.Result[pair].GetTodayOpen())
	}

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, tickerInformationPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())

	// Check request query string
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strings.Join(opts.Pairs, ","), record.Request.URL.Query().Get("pair"))
}

// Test GetOHLCData when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetOHLCData() {

	// Test parameters
	params := market.GetOHLCDataRequestParameters{
		Pair: "XXBTZUSD",
	}
	options := &market.GetOHLCDataRequestOptions{
		Interval: int64(market.M1),
		Since:    time.Now().Unix(),
	}

	// Predefined server response
	expectedJSONResponse := `{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  1688671200,
			  "30306.1",
			  "30306.2",
			  "30305.7",
			  "30305.7",
			  "30306.1",
			  "3.39243896",
			  23
			],
			[
			  1688671260,
			  "30304.5",
			  "30304.5",
			  "30300.0",
			  "30300.0",
			  "30300.0",
			  "4.42996871",
			  18
			],
			[
			  1688671320,
			  "30300.3",
			  "30300.4",
			  "30291.4",
			  "30291.4",
			  "30294.7",
			  "2.13024789",
			  25
			],
			[
			  1688671380,
			  "30291.8",
			  "30295.1",
			  "30291.8",
			  "30295.0",
			  "30293.8",
			  "1.01836275",
			  9
			]
		  ],
		  "last": 1688672160
		}
	  }`
	expectedXXBTZUSDLength := 4

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetOHLCData(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result.Data, expectedXXBTZUSDLength)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, ohlcDataPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())

	// Check request query string
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), params.Pair, record.Request.URL.Query().Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(options.Interval, 10), record.Request.URL.Query().Get("interval"))
	require.Equal(suite.T(), strconv.FormatInt(options.Since, 10), record.Request.URL.Query().Get("since"))
}

// Test GetOrderBook when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetOrderBook() {

	// Test parameters
	params := market.GetOrderBookRequestParameters{
		Pair: "XXBTZUSD",
	}
	options := &market.GetOrderBookRequestOptions{
		Count: 2,
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "XXBTZUSD": {
			"asks": [
			  [
				"30384.10000",
				"2.059",
				1688671659
			  ],
			  [
				"30387.90000",
				"1.500",
				1688671380
			  ],
			  [
				"30393.70000",
				"9.871",
				1688671261
			  ]
			],
			"bids": [
			  [
				"30297.00000",
				"1.115",
				1688671636
			  ],
			  [
				"30296.70000",
				"2.002",
				1688671674
			  ],
			  [
				"30289.80000",
				"5.001",
				1688671673
			  ]
			]
		  }
		}
	}`
	expectedLength := 3

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetOrderBook(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Equal(suite.T(), params.Pair, resp.Result.PairId)
	require.Len(suite.T(), resp.Result.Asks, expectedLength)
	require.Len(suite.T(), resp.Result.Bids, expectedLength)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, orderBookPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())

	// Check request query string
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), params.Pair, record.Request.URL.Query().Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(int64(options.Count), 10), record.Request.URL.Query().Get("count"))
}

// Test GetRecentTrades when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetRecentTrades() {

	// Test parameters
	params := market.GetRecentTradesRequestParameters{
		Pair: "XXBTZUSD",
	}
	options := &market.GetRecentTradesRequestOptions{
		Since: time.Now().Unix(),
		Count: 10,
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  "30243.40000",
			  "0.34507674",
			  1688669597.8277369,
			  "b",
			  "m",
			  "",
			  61044952
			],
			[
			  "30243.30000",
			  "0.00376960",
			  1688669598.2804112,
			  "s",
			  "l",
			  "",
			  61044953
			],
			[
			  "30243.30000",
			  "0.01235716",
			  1688669602.698379,
			  "s",
			  "m",
			  "",
			  61044956
			]
		  ],
		  "last": "1688671969993150842"
		}
	}`
	expectedLength := 3

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetRecentTrades(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Equal(suite.T(), params.Pair, resp.Result.PairId)
	require.Len(suite.T(), resp.Result.Trades, expectedLength)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, recentTradesPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())

	// Check request query string
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), params.Pair, record.Request.URL.Query().Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(int64(options.Count), 10), record.Request.URL.Query().Get("count"))
	require.Equal(suite.T(), strconv.FormatInt(options.Since, 10), record.Request.URL.Query().Get("since"))
}

// Test GetRecentTrades when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetRecentSpreadsHappyPath() {

	// Test parameters
	params := market.GetRecentSpreadsRequestParameters{
		Pair: "XXBTZUSD",
	}
	options := &market.GetRecentSpreadsRequestOptions{
		Since: time.Now().Unix(),
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  1688671834,
			  "30292.10000",
			  "30297.50000"
			],
			[
			  1688671834,
			  "30292.10000",
			  "30296.70000"
			],
			[
			  1688671834,
			  "30292.70000",
			  "30296.70000"
			]
		  ],
		  "last": 1688672106
		}
	}`
	expectedLength := 3

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetRecentSpreads(context.Background(), params, options)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Equal(suite.T(), params.Pair, resp.Result.PairId)
	require.Len(suite.T(), resp.Result.Spreads, expectedLength)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, recentSpreadsPath)
	require.Equal(suite.T(), http.MethodGet, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())

	// Check request query string
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), params.Pair, record.Request.URL.Query().Get("pair"))
	require.Equal(suite.T(), strconv.FormatInt(options.Since, 10), record.Request.URL.Query().Get("since"))
}

/*************************************************************************************************/
/* UNIT TESTS - MARKET DATA                                                                      */
/*************************************************************************************************/

// Test GetAccountBalance when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetAccountBalance() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "ZUSD": "171288.6158",
		  "ZEUR": "504861.8946",
		  "XXBT": "1011.1908877900",
		  "XETH": "818.5500000000",
		  "USDT": "500000.00000000",
		  "DAI": "9999.9999999999",
		  "DOT": "2.5000000000",
		  "ETH2.S": "198.3970800000",
		  "ETH2": "2.5885574330",
		  "USD.M": "1213029.2780"
		}
	}`
	expectedLength := 10
	expectedZUSDBalance := "171288.6158"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetAccountBalance(context.Background(), expectedNonce, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedLength)
	require.Equal(suite.T(), expectedZUSDBalance, resp.Result["ZUSD"].String())

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getAccountBalancePath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
}

// Test GetExtendedBalance when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetExtendedBalance() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Predefined server response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "ZUSD": {
			"balance": 25435.21,
			"hold_trade": 8249.76
		  },
		  "XXBT": {
			"balance": 1.2435,
			"hold_trade": 0.8423
		  }
		}
	  }`
	expectedLength := 2
	expectedZUSDBalance := "25435.21"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetExtendedBalance(context.Background(), expectedNonce, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedLength)
	require.NotNil(suite.T(), expectedZUSDBalance, resp.Result["ZUSD"])
	require.NotNil(suite.T(), expectedZUSDBalance, resp.Result["ZUSD"].Balance.String())

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getExtendedBalancePath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
}

// Test GetTradeBalance when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetTradeBalance() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}
	// Expected options
	options := &account.GetTradeBalanceRequestOptions{
		Asset: "ZUSD",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `{
		"error": [],
		"result": {
		  "eb": "1101.3425",
		  "tb": "392.2264",
		  "m": "7.0354",
		  "n": "-10.0232",
		  "c": "21.1063",
		  "v": "31.1297",
		  "e": "382.2032",
		  "mf": "375.1678",
		  "ml": "5432.57"
		}
	  }`
	expectedEquivalentBalance := "1101.3425"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetTradeBalance(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedEquivalentBalance, resp.Result.EquivalentBalance.String())
	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getTradeBalancePath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), options.Asset, record.Request.Form.Get("asset"))
}

// Test GetOpenOrders when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetOpenOrders() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}
	// Expected options
	options := &account.GetOpenOrdersRequestOptions{
		Trades:        true,
		UserReference: new(int64),
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "open": {
			"OQCLML-BW3P3-BUCMWZ": {
			  "refid": "None",
			  "userref": 0,
			  "status": "open",
			  "opentm": 1688666559.8974,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTUSD",
				"type": "buy",
				"ordertype": "limit",
				"price": "30010.0",
				"price2": "0",
				"leverage": "none",
				"order": "buy 1.25000000 XBTUSD @ limit 30010.0",
				"close": ""
			  },
			  "vol": "1.25000000",
			  "vol_exec": "0.37500000",
			  "cost": "11253.7",
			  "fee": "0.00000",
			  "price": "30010.0",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq",
			  "trades": [
				"TCCCTY-WE2O6-P3NB37"
			  ]
			},
			"OB5VMB-B4U2U-DK2WRW": {
			  "refid": "None",
			  "userref": 45326,
			  "status": "open",
			  "opentm": 1688665899.5699,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTUSD",
				"type": "buy",
				"ordertype": "limit",
				"price": "14500.0",
				"price2": "0",
				"leverage": "5:1",
				"order": "buy 0.27500000 XBTUSD @ limit 14500.0 with 5:1 leverage",
				"close": ""
			  },
			  "vol": "0.27500000",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "price": "0.00000",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq"
			}
		  }
		}
	}`
	expectedLength := 2
	targetOrder := "OB5VMB-B4U2U-DK2WRW"
	expectedTargetOrderDescrPair := "XBTUSD"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetOpenOrders(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result.Open, expectedLength)
	require.NotNil(suite.T(), resp.Result.Open[targetOrder])
	require.Equal(suite.T(), expectedTargetOrderDescrPair, resp.Result.Open[targetOrder].Description.Pair)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getOpenOrdersPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), record.Request.Form.Get("trades"))
	require.Equal(suite.T(), strconv.FormatInt(*options.UserReference, 10), record.Request.Form.Get("userref"))
}

// Test GetClosedOrders when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetClosedOrders() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}
	// Expected options
	options := &account.GetClosedOrdersRequestOptions{
		Trades:           true,
		UserReference:    new(int64),
		Start:            strconv.FormatInt(time.Now().Unix(), 10),
		End:              strconv.FormatInt(time.Now().Unix(), 10),
		Offset:           10,
		Closetime:        string(account.UseBoth),
		ConsolidateTaker: true,
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "closed": {
			"O37652-RJWRT-IMO74O": {
			  "refid": "None",
			  "userref": 1,
			  "status": "canceled",
			  "reason": "User requested",
			  "opentm": 1688148493.7708,
			  "closetm": 1688148610.0482,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTGBP",
				"type": "buy",
				"ordertype": "stop-loss-limit",
				"price": "23667.0",
				"price2": "0",
				"leverage": "none",
				"order": "buy 0.00100000 XBTGBP @ limit 23667.0",
				"close": ""
			  },
			  "vol": "0.00100000",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "price": "0.00000",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq",
			  "trigger": "index"
			},
			"O6YDQ5-LOMWU-37YKEE": {
			  "refid": "None",
			  "userref": 36493663,
			  "status": "canceled",
			  "reason": "User requested",
			  "opentm": 1688148493.7708,
			  "closetm": 1688148610.0477,
			  "starttm": 0,
			  "expiretm": 0,
			  "descr": {
				"pair": "XBTEUR",
				"type": "buy",
				"ordertype": "take-profit-limit",
				"price": "27743.0",
				"price2": "0",
				"leverage": "none",
				"order": "buy 0.00100000 XBTEUR @ limit 27743.0",
				"close": ""
			  },
			  "vol": "0.00100000",
			  "vol_exec": "0.00000000",
			  "cost": "0.00000",
			  "fee": "0.00000",
			  "price": "0.00000",
			  "stopprice": "0.00000",
			  "limitprice": "0.00000",
			  "misc": "",
			  "oflags": "fciq",
			  "trigger": "index"
			}
		  },
		  "count": 2
		}
	}`
	expectedCount := 2
	targetOrder := "O6YDQ5-LOMWU-37YKEE"
	expectedTargetOrderDescr := "buy 0.00100000 XBTEUR @ limit 27743.0"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetClosedOrders(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result.Closed, expectedCount)
	require.Equal(suite.T(), resp.Result.Count, expectedCount)
	require.NotNil(suite.T(), resp.Result.Closed[targetOrder])
	require.Equal(suite.T(), expectedTargetOrderDescr, resp.Result.Closed[targetOrder].Description.OrderDescription)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getClosedOrdersPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), record.Request.Form.Get("trades"))
	require.Equal(suite.T(), strconv.FormatInt(*options.UserReference, 10), record.Request.Form.Get("userref"))
	require.Equal(suite.T(), options.Start, record.Request.Form.Get("start"))
	require.Equal(suite.T(), options.End, record.Request.Form.Get("end"))
	require.Equal(suite.T(), strconv.FormatInt(options.Offset, 10), record.Request.Form.Get("ofs"))
	require.Equal(suite.T(), options.Closetime, record.Request.Form.Get("closetime"))
	require.Equal(suite.T(), strconv.FormatBool(options.ConsolidateTaker), record.Request.Form.Get("consolidate_taker"))
}

// Test QueryOrdersInfo when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestQueryOrdersInfo() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	taker := true
	options := &account.QueryOrdersInfoRequestOptions{
		Trades:           true,
		UserReference:    new(int64),
		ConsolidateTaker: &taker,
	}

	// Expected parameters
	params := account.QueryOrdersInfoParameters{
		TxId: []string{"txid1", "txid2"},
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "OBCMZD-JIEE7-77TH3F": {
			"refid": "None",
			"userref": 0,
			"status": "closed",
			"reason": null,
			"opentm": 1688665496.7808,
			"closetm": 1688665499.1922,
			"starttm": 0,
			"expiretm": 0,
			"descr": {
			  "pair": "XBTUSD",
			  "type": "buy",
			  "ordertype": "stop-loss-limit",
			  "price": "27500.0",
			  "price2": "0",
			  "leverage": "none",
			  "order": "buy 1.25000000 XBTUSD @ limit 27500.0",
			  "close": ""
			},
			"vol": "1.25000000",
			"vol_exec": "1.25000000",
			"cost": "27526.2",
			"fee": "26.2",
			"price": "27500.0",
			"stopprice": "0.00000",
			"limitprice": "0.00000",
			"misc": "",
			"oflags": "fciq",
			"trigger": "index",
			"trades": [
			  "TZX2WP-XSEOP-FP7WYR"
			]
		  },
		  "OMMDB2-FSB6Z-7W3HPO": {
			"refid": "None",
			"userref": 0,
			"status": "closed",
			"reason": null,
			"opentm": 1688592012.2317,
			"closetm": 1688592012.2335,
			"starttm": 0,
			"expiretm": 0,
			"descr": {
			  "pair": "XBTUSD",
			  "type": "sell",
			  "ordertype": "market",
			  "price": "0",
			  "price2": "0",
			  "leverage": "none",
			  "order": "sell 0.25000000 XBTUSD @ market",
			  "close": ""
			},
			"vol": "0.25000000",
			"vol_exec": "0.25000000",
			"cost": "7500.0",
			"fee": "7.5",
			"price": "30000.0",
			"stopprice": "0.00000",
			"limitprice": "0.00000",
			"misc": "",
			"oflags": "fcib",
			"trades": [
			  "TJUW2K-FLX2N-AR2FLU"
			]
		  }
		}
	}`
	expectedCount := 2
	targetOrder := "OMMDB2-FSB6Z-7W3HPO"
	expectedTargetOrderTrades := []string{"TJUW2K-FLX2N-AR2FLU"}

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.QueryOrdersInfo(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[targetOrder])
	require.ElementsMatch(suite.T(), expectedTargetOrderTrades, resp.Result[targetOrder].Trades)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, queryOrdersInfosPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), record.Request.Form.Get("trades"))
	require.Equal(suite.T(), strconv.FormatInt(*options.UserReference, 10), record.Request.Form.Get("userref"))
	require.Equal(suite.T(), strconv.FormatBool(*options.ConsolidateTaker), record.Request.Form.Get("consolidate_taker"))
	require.Equal(suite.T(), strings.Join(params.TxId, ","), record.Request.Form.Get("txid"))
}

// Test GetTradesHistory when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetTradesHistory() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &account.GetTradesHistoryRequestOptions{
		Type:             string(account.TradeTypeAll),
		Trades:           true,
		Start:            strconv.FormatInt(time.Now().Unix(), 10),
		End:              strconv.FormatInt(time.Now().Unix(), 10),
		Offset:           10,
		ConsolidateTaker: true,
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "trades": {
			"THVRQM-33VKH-UCI7BS": {
			  "ordertxid": "OQCLML-BW3P3-BUCMWZ",
			  "postxid": "TKH2SE-M7IF5-CFI7LT",
			  "pair": "XXBTZUSD",
			  "time": 1688667796.8802,
			  "type": "buy",
			  "ordertype": "limit",
			  "price": "30010.00000",
			  "cost": "600.20000",
			  "fee": "0.00000",
			  "vol": "0.02000000",
			  "margin": "0.00000",
			  "misc": ""
			},
			"TCWJEG-FL4SZ-3FKGH6": {
			  "ordertxid": "OQCLML-BW3P3-BUCMWZ",
			  "postxid": "TKH2SE-M7IF5-CFI7LT",
			  "pair": "XXBTZUSD",
			  "time": 1688667769.6396,
			  "type": "buy",
			  "ordertype": "limit",
			  "price": "30010.00000",
			  "cost": "300.10000",
			  "fee": "0.00000",
			  "vol": "0.01000000",
			  "margin": "0.00000",
			  "misc": ""
			}
		  }
		}
	}`
	expectedCount := 2
	targetOrder := "TCWJEG-FL4SZ-3FKGH6"
	expectedTargetOrderTxID := "OQCLML-BW3P3-BUCMWZ"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetTradesHistory(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result.Trades, expectedCount)
	require.Equal(suite.T(), 0, resp.Result.Count) // from doc, we should expect count to be populated but the provided response does not have a count field
	require.NotNil(suite.T(), resp.Result.Trades[targetOrder])
	require.Equal(suite.T(), expectedTargetOrderTxID, resp.Result.Trades[targetOrder].OrderTransactionId)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getTradesHistoryPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), record.Request.Form.Get("trades"))
	require.Equal(suite.T(), options.Start, record.Request.Form.Get("start"))
	require.Equal(suite.T(), options.End, record.Request.Form.Get("end"))
	require.Equal(suite.T(), strconv.FormatInt(options.Offset, 10), record.Request.Form.Get("ofs"))
	require.Equal(suite.T(), strconv.FormatBool(options.ConsolidateTaker), record.Request.Form.Get("consolidate_taker"))
}

// Test QueryTradesInfo when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestQueryTradesInfo() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &account.QueryTradesRequestOptions{
		Trades: true,
	}

	// Expected parameters
	params := account.QueryTradesRequestParameters{
		TransactionIds: []string{"THVRQM-33VKH-UCI7BS", "OH76VO-UKWAD-PSBDX6"},
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "THVRQM-33VKH-UCI7BS": {
			"ordertxid": "OQCLML-BW3P3-BUCMWZ",
			"postxid": "TKH2SE-M7IF5-CFI7LT",
			"pair": "XXBTZUSD",
			"time": 1688667796.8802,
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
			"postxid": "TKH2SE-M7IF5-CFI7LT",
			"pair": "XXBTZEUR",
			"time": 1688082549.3138,
			"type": "buy",
			"ordertype": "limit",
			"price": "27732.00000",
			"cost": "0.20020",
			"fee": "0.00000",
			"vol": "0.00020000",
			"margin": "0.00000",
			"misc": ""
		  }
		}
	}`
	expectedCount := 2
	targetOrder := "TTEUX3-HDAAA-RC2RUO"
	expectedTargetOrderTxID := "OH76VO-UKWAD-PSBDX6"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.QueryTradesInfo(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[targetOrder])
	require.Equal(suite.T(), expectedTargetOrderTxID, resp.Result[targetOrder].OrderTransactionId)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, queryTradesInfoPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), record.Request.Form.Get("trades"))
	require.Equal(suite.T(), strings.Join(params.TransactionIds, ","), record.Request.Form.Get("txid"))
}

// Test GetOpenPositions when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetOpenPositions() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &account.GetOpenPositionsRequestOptions{
		TransactionIds: []string{"TF5GVO-T7ZZ2-6NBKBI", "T24DOR-TAFLM-ID3NYP"},
		DoCalcs:        true,
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
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
		  },
		  "TYMRFG-URRG5-2ZTQSD": {
			"ordertxid": "OF5WFH-V57DP-QANDAC",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1610448039.8374,
			"type": "buy",
			"ordertype": "limit",
			"cost": "0.00240",
			"fee": "0.00000",
			"vol": "0.00000010",
			"vol_closed": "0.00000000",
			"margin": "0.00048",
			"value": "0",
			"net": "+0.0006",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  },
		  "TAFGBN-TZNFC-7CCYIM": {
			"ordertxid": "OF5WFH-V57DP-QANDAC",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1610448039.8448,
			"type": "buy",
			"ordertype": "limit",
			"cost": "2.40000",
			"fee": "0.00264",
			"vol": "0.00010000",
			"vol_closed": "0.00000000",
			"margin": "0.48000",
			"value": "3.0",
			"net": "+0.6015",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  },
		  "T4O5L3-4VGS4-IRU2UL": {
			"ordertxid": "OF5WFH-V57DP-QANDAC",
			"posstatus": "open",
			"pair": "XXBTZUSD",
			"time": 1610448040.7722,
			"type": "buy",
			"ordertype": "limit",
			"cost": "21.59760",
			"fee": "0.02376",
			"vol": "0.00089990",
			"vol_closed": "0.00000000",
			"margin": "4.31952",
			"value": "27.0",
			"net": "+5.4133",
			"terms": "0.0100% per 4 hours",
			"rollovertm": "1616672637",
			"misc": "",
			"oflags": ""
		  }
		}
	}`
	expectedCount := 5
	targetPosition := "T4O5L3-4VGS4-IRU2UL"
	expectedTargetOrderTxID := "OF5WFH-V57DP-QANDAC"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetOpenPositions(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[targetPosition])
	require.Equal(suite.T(), expectedTargetOrderTxID, resp.Result[targetPosition].OrderTransactionId)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getOpenPositionsPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.DoCalcs), record.Request.Form.Get("docalcs"))
	require.Equal(suite.T(), strings.Join(options.TransactionIds, ","), record.Request.Form.Get("txid"))
}

// Test GetLedgersInfo when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetLedgersInfo() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &account.GetLedgersInfoRequestOptions{
		Assets:       []string{"XXBT"},
		AssetClass:   "currency",
		Type:         string(account.LedgerAll),
		Start:        strconv.FormatInt(time.Now().Unix(), 10),
		End:          strconv.FormatInt(time.Now().Unix(), 10),
		Offset:       10,
		WithoutCount: true,
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "ledger": {
			"L4UESK-KG3EQ-UFO4T5": {
			  "refid": "TJKLXF-PGMUI-4NTLXU",
			  "time": 1688464484.1787,
			  "type": "trade",
			  "subtype": "",
			  "aclass": "currency",
			  "asset": "ZGBP",
			  "amount": "-24.5000",
			  "fee": "0.0490",
			  "balance": "459567.9171"
			},
			"LMKZCZ-Z3GVL-CXKK4H": {
			  "refid": "TBZIP2-F6QOU-TMB6FY",
			  "time": 1688444262.8888,
			  "type": "trade",
			  "subtype": "",
			  "aclass": "currency",
			  "asset": "ZUSD",
			  "amount": "0.9852",
			  "fee": "0.0010",
			  "balance": "52732.1132"
			}
		  },
		  "count": 2
		}
	}`
	expectedCount := 2
	targetLedger := "LMKZCZ-Z3GVL-CXKK4H"
	expectedTargetRefId := "TBZIP2-F6QOU-TMB6FY"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetLedgersInfo(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result.Ledgers, expectedCount)
	require.Equal(suite.T(), expectedCount, resp.Result.Count)
	require.NotNil(suite.T(), resp.Result.Ledgers[targetLedger])
	require.Equal(suite.T(), expectedTargetRefId, resp.Result.Ledgers[targetLedger].ReferenceId)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getLedgersInfoPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strings.Join(options.Assets, ","), record.Request.Form.Get("asset"))
	require.Equal(suite.T(), options.AssetClass, record.Request.Form.Get("aclass"))
	require.Equal(suite.T(), options.Type, record.Request.Form.Get("type"))
	require.Equal(suite.T(), options.Start, record.Request.Form.Get("start"))
	require.Equal(suite.T(), options.End, record.Request.Form.Get("end"))
	require.Equal(suite.T(), strconv.FormatInt(options.Offset, 10), record.Request.Form.Get("ofs"))
	require.Equal(suite.T(), strconv.FormatBool(options.WithoutCount), record.Request.Form.Get("without_count"))
}

// Test QueryLedgers when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestQueryLedgers() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &account.QueryLedgersRequestOptions{
		Trades: true,
	}

	// Expected parameters
	params := account.QueryLedgersRequestParameters{
		Id: []string{"L4UESK-KG3EQ-UFO4T5", "L4UESK-KG3EQ-UFO4T5"},
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "L4UESK-KG3EQ-UFO4T5": {
			"refid": "TJKLXF-PGMUI-4NTLXU",
			"time": 1688464484.1787,
			"type": "trade",
			"subtype": "",
			"aclass": "currency",
			"asset": "ZGBP",
			"amount": "-24.5000",
			"fee": "0.0490",
			"balance": "459567.9171"
		  }
		}
	}`
	expectedCount := 1
	targetLedger := "L4UESK-KG3EQ-UFO4T5"
	expectedTargetRefId := "TJKLXF-PGMUI-4NTLXU"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.QueryLedgers(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[targetLedger])
	require.Equal(suite.T(), expectedTargetRefId, resp.Result[targetLedger].ReferenceId)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, queryLedgersPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strings.Join(params.Id, ","), record.Request.Form.Get("id"))
	require.Equal(suite.T(), strconv.FormatBool(options.Trades), record.Request.Form.Get("trades"))
}

// Test GetTradeVolume when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetTradeVolume() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &account.GetTradeVolumeRequestOptions{
		Pairs: []string{"XXBTZUSD", "XETHZUSD"},
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "currency": "ZUSD",
		  "volume": "200709587.4223",
		  "fees": {
			"XXBTZUSD": {
			  "fee": "0.1000",
			  "minfee": "0.1000",
			  "maxfee": "0.2600",
			  "nextfee": null,
			  "nextvolume": null,
			  "tiervolume": "10000000.0000"
			}
		  },
		  "fees_maker": {
			"XXBTZUSD": {
			  "fee": "0.0000",
			  "minfee": "0.0000",
			  "maxfee": "0.1600",
			  "nextfee": null,
			  "nextvolume": null,
			  "tiervolume": "10000000.0000"
			}
		  }
		}
	}`
	targetFee := "XXBTZUSD"
	expectedTargetTierVolume := "10000000.0000"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetTradeVolume(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.NotNil(suite.T(), resp.Result.Fees[targetFee])
	require.Equal(suite.T(), expectedTargetTierVolume, resp.Result.Fees[targetFee].TierVolume.String())

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getTradeVolumePath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strings.Join(options.Pairs, ","), record.Request.Form.Get("pair"))
}

// Test RequestExportReport when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestRequestExportReport() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &account.RequestExportReportRequestOptions{
		Format:  string(account.TSV),
		Fields:  []string{string(account.FieldsAmount), string(account.FieldsBalance)},
		StartTm: time.Now().Unix(),
		EndTm:   time.Now().Unix(),
	}

	// Expected params
	params := account.RequestExportReportRequestParameters{
		Report:      string(account.ReportLedgers),
		Description: "Lorem",
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "id": "TCJA"
		}
	}`
	expectedId := "TCJA"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.RequestExportReport(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedId, resp.Result.Id)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, requestExportReportPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Report, record.Request.Form.Get("report"))
	require.Equal(suite.T(), params.Description, record.Request.Form.Get("description"))
	require.Equal(suite.T(), options.Format, record.Request.Form.Get("format"))
	require.Equal(suite.T(), strings.Join(options.Fields, ","), record.Request.Form.Get("fields"))
	require.Equal(suite.T(), strconv.FormatInt(options.StartTm, 10), record.Request.Form.Get("starttm"))
	require.Equal(suite.T(), strconv.FormatInt(options.EndTm, 10), record.Request.Form.Get("endtm"))
}

// Test GetExportReportStatus when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetExportReportStatus() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := account.GetExportReportStatusRequestParameters{
		Report: string(account.ReportLedgers),
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": [
		  {
			"id": "VSKC",
			"descr": "my_trades_1",
			"format": "CSV",
			"report": "trades",
			"subtype": "all",
			"status": "Processed",
			"flags": "0",
			"fields": "all",
			"createdtm": "1688669085",
			"expiretm": "1688878685",
			"starttm": "1688669093",
			"completedtm": "1688669093",
			"datastarttm": "1683556800",
			"dataendtm": "1688669085",
			"aclass": "forex",
			"asset": "all"
		  },
		  {
			"id": "TCJA",
			"descr": "my_trades_1",
			"format": "CSV",
			"report": "trades",
			"subtype": "all",
			"status": "Processed",
			"flags": "0",
			"fields": "all",
			"createdtm": "1688363637",
			"expiretm": "1688573237",
			"starttm": "1688363664",
			"completedtm": "1688363664",
			"datastarttm": "1683235200",
			"dataendtm": "1688363637",
			"aclass": "forex",
			"asset": "all"
		  }
		]
	}`
	expectedCount := 2
	expectedItem0Id := "VSKC"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetExportReportStatus(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[0])
	require.Equal(suite.T(), expectedItem0Id, resp.Result[0].Id)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getExportReportStatusPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Report, record.Request.Form.Get("report"))
}

// Test RetrieveDataExport when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestRetrieveDataExport() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := account.RetrieveDataExportParameters{
		Id: "42",
	}

	// Expected API response
	expectedResponseBody := "hello world"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/octet-stream"}},
		Body:    []byte(expectedResponseBody),
	})

	// Make request
	resp, httpresp, err := suite.client.RetrieveDataExport(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Read body
	require.NotNil(suite.T(), resp.Report)
	body, err := io.ReadAll(resp.Report)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), expectedResponseBody, string(body))

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, retrieveDataExportPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Id, record.Request.Form.Get("id"))
}

// Test DeleteExportReport when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestDeleteExportReport() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := account.DeleteExportReportRequestParameters{
		Id:   "42",
		Type: string(account.DeleteReport),
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "delete": true
		}
	}`

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.DeleteExportReport(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.True(suite.T(), resp.Result.Delete)
	require.False(suite.T(), resp.Result.Cancel)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, deleteExportReportPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Id, record.Request.Form.Get("id"))
}

// // TestDeleteExportReportErrPath Test will succeed if a error response from server is well processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestDeleteExportReportErrPath() {

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.DeleteExportReport(DeleteExportReportParameters{}, nil)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Check response
// 	require.Nil(suite.T(), resp)
// 	require.Error(suite.T(), err)
// }

// /*****************************************************************************/
// /* UNIT TESTS - USER TRADING												 */
// /*****************************************************************************/

// // Test Add Order Method - Happy path
// //
// // Test will succeed if all provided input parameters are in request sent by client
// // and if predefined response from server is correctly processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderHappyPath() {

// 	// Test parameters
// 	pair := "XXBTZUSD"
// 	userref := new(int64)
// 	*userref = 42
// 	deadline := time.Now().UTC()
// 	validate := true
// 	otp := "Once"
// 	order := &Order{
// 		UserReference:      userref,
// 		OrderType:          OTypeLimit,
// 		Type:               Buy,
// 		Volume:             "2.1234",
// 		Price:              "45000.1",
// 		Price2:             "45000.1",
// 		Trigger:            TriggerLast,
// 		Leverage:           "2:1",
// 		StpType:            StpCancelNewest,
// 		OrderFlags:         strings.Join([]string{OFlagFeeInQuote, OFlagPost}, ","),
// 		TimeInForce:        GoodTilCanceled,
// 		ScheduledStartTime: "0",
// 		ExpirationTime:     "0",
// 		Close:              &CloseOrder{OrderType: OTypeStopLossLimit, Price: "38000.42", Price2: "36000"},
// 	}

// 	// Predefined server response
// 	expectedJSONResponse := `{
// 		"error": [ ],
// 		"result": {
// 			"descr": {
// 				"order": "buy 2.12340000 XBTUSD @ limit 45000.1 with 2:1 leverage",
// 				"close": "close position @ stop loss 38000.42 -> limit 36000.0"
// 			},
// 			"txid": [
// 				"OUF4EM-FRGI2-MQMWZD",
// 				"OUF4EM-FRGI2-MQMW42"
// 			]
// 		}
// 	}`

// 	// Expected response data
// 	expOrderDescr := "buy 2.12340000 XBTUSD @ limit 45000.1 with 2:1 leverage"
// 	expCloseDescr := "close position @ stop loss 38000.42 -> limit 36000.0"
// 	expTxID := []string{"OUF4EM-FRGI2-MQMWZD", "OUF4EM-FRGI2-MQMW42"}

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.AddOrder(
// 		AddOrderParameters{Pair: pair, Order: *order},
// 		&AddOrderOptions{Deadline: &deadline, Validate: validate},
// 		&SecurityOptions{SecondFactor: otp})

// 	// Get and log request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check client error and log response
// 	require.NoError(suite.T(), err)
// 	suite.T().Log(resp)

// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postAddOrder)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))
// 	require.Equal(suite.T(), pair, req.Form.Get("pair"))
// 	actUserRef, err := strconv.ParseInt(req.Form.Get("userref"), 10, 64)
// 	require.NoError(suite.T(), err)
// 	require.Equal(suite.T(), *order.UserReference, actUserRef)
// 	require.Equal(suite.T(), order.OrderType, req.Form.Get("ordertype"))
// 	require.Equal(suite.T(), order.Type, req.Form.Get("type"))
// 	require.Equal(suite.T(), order.Volume, req.Form.Get("volume"))
// 	require.Equal(suite.T(), order.Price, req.Form.Get("price"))
// 	require.Equal(suite.T(), order.Price2, req.Form.Get("price2"))
// 	require.Equal(suite.T(), order.Trigger, req.Form.Get("trigger"))
// 	require.Equal(suite.T(), order.Leverage, req.Form.Get("leverage"))
// 	require.Equal(suite.T(), order.StpType, req.Form.Get("stp_type"))
// 	require.Equal(suite.T(), order.OrderFlags, req.Form.Get("oflags"))
// 	require.Equal(suite.T(), order.TimeInForce, req.Form.Get("timeinforce"))
// 	require.Equal(suite.T(), order.ScheduledStartTime, req.Form.Get("starttm"))
// 	require.Equal(suite.T(), order.ExpirationTime, req.Form.Get("expiretm"))
// 	require.Equal(suite.T(), order.Close.OrderType, req.Form.Get("close[ordertype]"))
// 	require.Equal(suite.T(), order.Close.Price, req.Form.Get("close[price]"))
// 	require.Equal(suite.T(), order.Close.Price2, req.Form.Get("close[price2]"))
// 	actValidate, err := strconv.ParseBool(req.Form.Get("validate"))
// 	require.NoError(suite.T(), err)
// 	require.Equal(suite.T(), validate, actValidate)
// 	actDeadline, err := time.Parse(time.RFC3339, req.Form.Get("deadline"))
// 	require.NoError(suite.T(), err)
// 	// Nanoseconds are not provided
// 	require.Equal(suite.T(), deadline.Truncate(time.Second), actDeadline)

// 	// Check server response
// 	require.Equal(suite.T(), expOrderDescr, resp.Result.Description.Order)
// 	require.Equal(suite.T(), expCloseDescr, resp.Result.Description.Close)
// 	require.ElementsMatch(suite.T(), expTxID, resp.Result.TransactionIDs)
// }

// // Test Add Order Method - Empty parameters
// //
// // Test will succeed if a valid request is sent by client, if empty parameters are not included
// // in request and if predefined response from server is correctly processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderEmptyParameters() {

// 	// Test parameters
// 	pair := "XXBTZUSD"
// 	order := &Order{
// 		OrderType: OTypeMarket,
// 		Type:      Buy,
// 		Volume:    "2.1234",
// 	}

// 	// Predefined server response
// 	expectedJSONResponse := `{
// 		"error": [ ],
// 		"result": {
// 			"descr": {
// 				"order": "buy 2.12340000 XBTUSD @ market"
// 			},
// 			"txid": [
// 				"OUF4EM-FRGI2-MQMWZD"
// 			]
// 		}
// 	}`

// 	// Expected response data
// 	expOrderDescr := "buy 2.12340000 XBTUSD @ market"
// 	expTxID := []string{"OUF4EM-FRGI2-MQMWZD"}

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.AddOrder(AddOrderParameters{Pair: pair, Order: *order}, nil, nil)

// 	// Get and log client request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check for client error & log response
// 	require.NoError(suite.T(), err)
// 	suite.T().Log(resp)

// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postAddOrder)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Empty(suite.T(), req.Form.Get("otp"))
// 	require.Equal(suite.T(), pair, req.Form.Get("pair"))
// 	require.Nil(suite.T(), order.UserReference)
// 	require.Empty(suite.T(), req.Form.Get("userref"))
// 	require.Equal(suite.T(), order.OrderType, req.Form.Get("ordertype"))
// 	require.Equal(suite.T(), order.Type, req.Form.Get("type"))
// 	require.Equal(suite.T(), order.Volume, req.Form.Get("volume"))
// 	require.Empty(suite.T(), order.Price)
// 	require.Empty(suite.T(), order.Price2)
// 	require.Empty(suite.T(), req.Form.Get("price"))
// 	require.Empty(suite.T(), req.Form.Get("price2"))
// 	require.Empty(suite.T(), order.Trigger)
// 	require.Empty(suite.T(), req.Form.Get("trigger"))
// 	require.Empty(suite.T(), order.Leverage)
// 	require.Empty(suite.T(), req.Form.Get("leverage"))
// 	require.Empty(suite.T(), order.StpType)
// 	require.Empty(suite.T(), req.Form.Get("stp_type"))
// 	require.Empty(suite.T(), order.OrderFlags)
// 	require.Empty(suite.T(), req.Form.Get("oflags"))
// 	require.Empty(suite.T(), order.TimeInForce)
// 	require.Empty(suite.T(), req.Form.Get("timeinforce"))
// 	require.Empty(suite.T(), order.ScheduledStartTime)
// 	require.Empty(suite.T(), req.Form.Get("starttm"))
// 	require.Empty(suite.T(), order.ExpirationTime)
// 	require.Empty(suite.T(), req.Form.Get("expiretm"))
// 	require.Nil(suite.T(), order.Close)
// 	require.Empty(suite.T(), req.Form.Get("close[ordertype]"))
// 	require.Empty(suite.T(), req.Form.Get("close[price]"))
// 	require.Empty(suite.T(), req.Form.Get("close[price2]"))
// 	require.Empty(suite.T(), req.Form.Get("validate"))
// 	require.Empty(suite.T(), req.Form.Get("deadline"))

// 	// Check server response
// 	require.Equal(suite.T(), expOrderDescr, resp.Result.Description.Order)
// 	require.Empty(suite.T(), resp.Result.Description.Close)
// 	require.ElementsMatch(suite.T(), expTxID, resp.Result.TransactionIDs)
// }

// // Test Add Order Method - Error path
// //
// // Test will succeed if API call fail because of invalid server response.
// func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderErrPath() {

// 	// Test parameters
// 	pair := "XXBTZUSD"
// 	validate := true
// 	order := &Order{
// 		OrderType: OTypeMarket,
// 		Type:      Buy,
// 		Volume:    "2.1234",
// 	}

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Perform request
// 	_, err := suite.client.AddOrder(
// 		AddOrderParameters{Pair: pair, Order: *order},
// 		&AddOrderOptions{Deadline: nil, Validate: validate},
// 		nil)
// 	require.Error(suite.T(), err)
// }

// // Test Add Order Batch - Happy Path
// //
// // Test will succeed if a request sent by client is well formatted, contains
// // all input parameters and if server response is well processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchHappyPath() {

// 	// Test parameters
// 	pair := "XXBTZUSD"
// 	userref := new(int64)
// 	*userref = 123
// 	deadline := time.Now().UTC()
// 	validate := true
// 	otp := "otp"
// 	orders := []Order{
// 		{
// 			OrderType:   OTypeLimit,
// 			Type:        Buy,
// 			Volume:      "1.2",
// 			Price:       "40000",
// 			Price2:      "40000.42",
// 			Trigger:     TriggerIndex,
// 			StpType:     StpCancelNewest,
// 			TimeInForce: GoodTilCanceled,
// 			Close: &CloseOrder{
// 				OrderType: OTypeStopLossLimit,
// 				Price:     "37000",
// 				Price2:    "36000",
// 			},
// 		},
// 		{
// 			UserReference:      userref,
// 			OrderType:          OTypeLimit,
// 			Type:               Sell,
// 			Volume:             "1.2",
// 			Price:              "42000",
// 			Leverage:           "2:1",
// 			OrderFlags:         "fciq,post",
// 			ScheduledStartTime: "0",
// 			ExpirationTime:     "0",
// 			Close:              nil,
// 		},
// 	}

// 	// Predefined response
// 	expectedJSONResponse := `{
// 		"error": [ ],
// 		"result": {
// 			"orders": [
// 				{
// 					"descr": {
// 						"order": "buy 1.2 BTCUSD @ limit 40000",
// 						"close": "close position @ stop loss 37000.0 -> limit 36000.0"
// 					},
// 					"txid": "OUF4EM-FRGI2-MQMWZD"
// 				},
// 				{
// 					"descr": {
// 						"order": "sell 1.2 BTCUSD @ limit 42000"
// 					},
// 					"txid": ["OCF5EM-FRGI2-MQWEDD", "OUF4EM-FRGI2-MQMWZD"]
// 				}
// 			]
// 		}
// 	}`

// 	expResp := &AddOrderBatchResponse{}
// 	err := json.Unmarshal([]byte(expectedJSONResponse), expResp)
// 	require.NoError(suite.T(), err)

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.AddOrderBatch(
// 		AddOrderBatchParameters{Pair: pair, Orders: orders},
// 		&AddOrderBatchOptions{Deadline: &deadline, Validate: validate},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Get and log client request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check client error and log response
// 	require.NoError(suite.T(), err)

// 	// Check request
// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postAddOrderBatch)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))
// 	require.Equal(suite.T(), pair, req.Form.Get("pair"))
// 	require.Equal(suite.T(), orders[0].OrderType, req.Form.Get("orders[0][ordertype]"))
// 	require.Equal(suite.T(), orders[0].Type, req.Form.Get("orders[0][type]"))
// 	require.Equal(suite.T(), orders[0].Volume, req.Form.Get("orders[0][volume]"))
// 	require.Equal(suite.T(), orders[0].Price, req.Form.Get("orders[0][price]"))
// 	require.Equal(suite.T(), orders[0].Price2, req.Form.Get("orders[0][price2]"))
// 	require.Equal(suite.T(), orders[0].Trigger, req.Form.Get("orders[0][trigger]"))
// 	require.Equal(suite.T(), orders[0].StpType, req.Form.Get("orders[0][stp_type]"))
// 	require.Equal(suite.T(), orders[0].TimeInForce, req.Form.Get("orders[0][timeinforce]"))
// 	require.Equal(suite.T(), orders[0].Close.OrderType, req.Form.Get("orders[0][close][ordertype]"))
// 	require.Equal(suite.T(), orders[0].Close.Price, req.Form.Get("orders[0][close][price]"))
// 	require.Equal(suite.T(), orders[0].Close.Price2, req.Form.Get("orders[0][close][price2]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[0][userref]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[0][leverage]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[0][oflags]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[0][starttm]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[0][expiretm]"))
// 	require.Equal(suite.T(), orders[1].OrderType, req.Form.Get("orders[1][ordertype]"))
// 	require.Equal(suite.T(), orders[1].Type, req.Form.Get("orders[1][type]"))
// 	require.Equal(suite.T(), orders[1].Volume, req.Form.Get("orders[1][volume]"))
// 	require.Equal(suite.T(), orders[1].Price, req.Form.Get("orders[1][price]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[1][price2]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[1][trigger]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[1][stp_type]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[1][timeinforce]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[1][close][ordertype]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[1][close][price]"))
// 	require.Empty(suite.T(), req.Form.Get("orders[1][close][price2]"))
// 	require.Equal(suite.T(), strconv.FormatInt(*orders[1].UserReference, 10), req.Form.Get("orders[1][userref]"))
// 	require.Equal(suite.T(), orders[1].Leverage, req.Form.Get("orders[1][leverage]"))
// 	require.Equal(suite.T(), orders[1].OrderFlags, req.Form.Get("orders[1][oflags]"))
// 	require.Equal(suite.T(), orders[1].ScheduledStartTime, req.Form.Get("orders[1][starttm]"))
// 	require.Equal(suite.T(), orders[1].ExpirationTime, req.Form.Get("orders[1][expiretm]"))
// 	actValidate, err := strconv.ParseBool(req.Form.Get("validate"))
// 	require.NoError(suite.T(), err)
// 	require.Equal(suite.T(), validate, actValidate)
// 	actDeadline, err := time.Parse(time.RFC3339, req.Form.Get("deadline"))
// 	require.NoError(suite.T(), err)
// 	// Nanoseconds are not provided
// 	require.Equal(suite.T(), deadline.Truncate(time.Second), actDeadline)

// 	// Check response
// 	require.Equal(suite.T(), expResp, resp)
// }

// // Test Add Order Batch - Empty orders
// //
// // Test will succeed if an error is thrown by client when
// // an empty list of orders is submitted.
// func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchEmptyOrders() {

// 	// Perform request
// 	_, err := suite.client.AddOrderBatch(AddOrderBatchParameters{}, nil, nil)
// 	require.Error(suite.T(), err)

// 	// Ensure no request has been sent
// 	req := suite.srv.PopRecordedRequest()
// 	require.Nil(suite.T(), req)
// }

// // Test Add Order Batch - Partial success
// //
// // Test will succeed if a response with some failed orders is well processed
// // by client. Response should contain non-failed orders and a non nil error.
// func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchPartialSuccess() {

// 	// Test parameters
// 	pair := "XXBTZUSD"
// 	validate := false
// 	orders := []Order{
// 		{
// 			OrderType: OTypeMarket,
// 			Type:      Buy,
// 			Volume:    "100000000.00000000",
// 		},
// 		{
// 			OrderType: OTypeMarket,
// 			Type:      Sell,
// 			Volume:    "100000000.00000000",
// 		},
// 	}

// 	// Predefined server response
// 	expectedJSONResponse := `{
// 		"error":[],
// 		"result":{
// 			"orders":[
// 				{
// 					"error":"EOrder:Insufficient funds",
// 					"descr":{
// 						"order":"buy 100000000.00000000 XBTEUR @ market"
// 					}
// 				},
// 				{
// 					"error":["EOrder:Insufficient funds", "wrong"],
// 					"descr":{
// 						"order":"sell 100000000.00000000 XBTEUR @ market"
// 					}
// 				}
// 			]
// 		}
// 	}`

// 	expResp := &AddOrderBatchResponse{}
// 	err := json.Unmarshal([]byte(expectedJSONResponse), expResp)
// 	require.NoError(suite.T(), err)

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, _ := suite.client.AddOrderBatch(
// 		AddOrderBatchParameters{Pair: pair, Orders: orders},
// 		&AddOrderBatchOptions{Deadline: nil, Validate: validate},
// 		nil)

// 	// Get and log request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check response
// 	require.Equal(suite.T(), expResp, resp)
// }

// // Test Add Order Batch - Error path
// //
// // Test will succeed if an error is returned when an invalid response is
// // received from server.
// func (suite *KrakenAPIClientUnitTestSuite) TestAddOrderBatchErrPath() {

// 	// Test parameters
// 	pair := "XXBTZUSD"
// 	validate := false
// 	orders := []Order{
// 		{
// 			OrderType: OTypeMarket,
// 			Type:      Buy,
// 			Volume:    "100000000.00000000",
// 		},
// 		{
// 			OrderType: OTypeMarket,
// 			Type:      Sell,
// 			Volume:    "100000000.00000000",
// 		},
// 	}

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Perform request
// 	resp, err := suite.client.AddOrderBatch(
// 		AddOrderBatchParameters{Pair: pair, Orders: orders},
// 		&AddOrderBatchOptions{Deadline: nil, Validate: validate},
// 		nil)

// 	// Check client error and response
// 	require.Error(suite.T(), err)
// 	require.Nil(suite.T(), resp)
// }

// // Test Edit Order - Happy Path
// //
// // Test will succeed if request sent by client is well formatted, contains all input
// // all input parameters and if server response is well processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestEditOrderHappyPath() {

// 	// Test parameters
// 	pair := "XXBTZUSD"
// 	userref := new(int64)
// 	*userref = 123
// 	originalTxID := "OHYO67-6LP66-HMQ437"
// 	volume := "0.00030000"
// 	price := "19500.0"
// 	price2 := "32500.0"
// 	oflags := []string{"fcib", "post"}
// 	deadline := time.Now().UTC()
// 	cancelResponse := true
// 	validate := true
// 	otp := "otp"

// 	// Predefined response
// 	expectedJSONResponse := `{
// 			"error": [ ],
// 			"result": {
// 				"status": "ok",
// 				"txid": "OFVXHJ-KPQ3B-VS7ELA",
// 				"originaltxid": "OHYO67-6LP66-HMQ437",
// 				"volume": "0.00030000",
// 				"price": "19500.0",
// 				"price2": "32500.0",
// 				"orders_cancelled": 1,
// 				"descr": {
// 				"order": "buy 0.00030000 XXBTZGBP @ limit 19500.0"
// 			}
// 		}
// 	}`

// 	// Expected response data
// 	expResp := &EditOrderResponse{}
// 	err := json.Unmarshal([]byte(expectedJSONResponse), &expResp)
// 	require.NoError(suite.T(), err)

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.EditOrder(
// 		EditOrderParameters{Pair: pair, Id: originalTxID},
// 		&EditOrderOptions{
// 			NewUserReference: strconv.FormatInt(*userref, 10),
// 			NewVolume:        volume,
// 			Price:            price,
// 			Price2:           price2,
// 			OFlags:           oflags,
// 			Deadline:         &deadline,
// 			CancelResponse:   cancelResponse,
// 			Validate:         validate,
// 		},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Get and log client request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check client error and log response
// 	require.NoError(suite.T(), err)
// 	suite.T().Log(resp)

// 	// Check request
// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postEditOrder)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))
// 	require.Equal(suite.T(), pair, req.Form.Get("pair"))
// 	require.Equal(suite.T(), strconv.FormatInt(*userref, 10), req.Form.Get("userref"))
// 	require.Equal(suite.T(), originalTxID, req.Form.Get("txid"))
// 	require.Equal(suite.T(), volume, req.Form.Get("volume"))
// 	require.Equal(suite.T(), price, req.Form.Get("price"))
// 	require.Equal(suite.T(), price2, req.Form.Get("price2"))
// 	require.Equal(suite.T(), strings.Join(oflags, ","), req.Form.Get("oflags"))
// 	actCancelResponse, err := strconv.ParseBool(req.Form.Get("cancel_response"))
// 	require.NoError(suite.T(), err)
// 	require.Equal(suite.T(), cancelResponse, actCancelResponse)
// 	actValidate, err := strconv.ParseBool(req.Form.Get("validate"))
// 	require.NoError(suite.T(), err)
// 	require.Equal(suite.T(), validate, actValidate)
// 	actDeadline, err := time.Parse(time.RFC3339, req.Form.Get("deadline"))
// 	require.NoError(suite.T(), err)
// 	// Nanoseconds are not provided
// 	require.Equal(suite.T(), deadline.Truncate(time.Second), actDeadline)

// 	// Check response
// 	require.Equal(suite.T(), expResp, resp)
// }

// // Test Cancel Order - Happy Path
// //
// // Test will succeed if request sent by client is well formatted, contains all input
// // all input parameters and if server response is well processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestCancelOrderHappyPath() {

// 	// Test parameters
// 	txID := "OHYO67-6LP66-HMQ437"
// 	otp := "otp"

// 	// Predefined response
// 	expectedJSONResponse := `{
// 		"result": {
// 			"count": 0,
// 			"pending": true
// 		},
// 		"error": []
// 	}`

// 	// Expected response data
// 	expCount := 0
// 	expPending := true

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.CancelOrder(CancelOrderParameters{Id: txID}, &SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Get and log client request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check client error and log response
// 	require.NoError(suite.T(), err)
// 	suite.T().Log(resp)

// 	// Check request
// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postCancelOrder)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))
// 	require.Equal(suite.T(), txID, req.Form.Get("txid"))

// 	// Check response
// 	require.Equal(suite.T(), expCount, resp.Result.Count)
// 	require.Equal(suite.T(), expPending, resp.Result.Pending)
// }

// // Test Cancel All Orders - Happy Path
// //
// // Test will succeed if request sent by client is well formatted, contains all input
// // all input parameters and if server response is well processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestCancelAllOrdersHappyPath() {

// 	// Test parameters
// 	otp := "otp"

// 	// Predefined response
// 	expectedJSONResponse := `{
// 		"result": {
// 			"count": 4
// 		},
// 		"error": []
// 	}`

// 	// Expected response data
// 	expCount := 4

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.CancelAllOrders(&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Get and log client request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check client error and log response
// 	require.NoError(suite.T(), err)
// 	suite.T().Log(resp)

// 	// Check request
// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postCancelAllOrders)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expCount, resp.Result.Count)
// }

// // Test Cancel All Orders After X - Happy Path
// //
// // Test will succeed if request sent by client is well formatted, contains all input
// // all input parameters and if server response is well processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestCancelAllOrdersAfterXHappyPath() {

// 	// Test parameters
// 	timeout := int64(60)
// 	otp := "otp"

// 	// Predefined response
// 	expectedJSONResponse := `{
// 		"result": {
// 			"currentTime": "2021-03-24T17:41:56Z",
// 			"triggerTime": "2021-03-24T17:42:56Z"
// 		},
// 		"error": []
// 	}`

// 	// Expected response data
// 	expYear := 2021
// 	expMonth := time.Month(3)
// 	expDayOfMonth := 24
// 	expHour := 17
// 	expSecond := 56
// 	expNanosec := 0
// 	expCurrMinute := 41
// 	expTrigMinute := 42

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.CancelAllOrdersAfterX(CancelCancelAllOrdersAfterXParameters{Timeout: timeout}, &SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Get and log client request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check client error and log response
// 	require.NoError(suite.T(), err)
// 	suite.T().Log(resp)

// 	// Check request
// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postCancelAllOrdersAfterX)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))
// 	require.Equal(suite.T(), strconv.FormatInt(timeout, 10), req.Form.Get("timeout"))

// 	// Check response
// 	require.Equal(suite.T(), expYear, resp.Result.CurrentTime.Year())
// 	require.Equal(suite.T(), expYear, resp.Result.TriggerTime.Year())
// 	require.Equal(suite.T(), expMonth, resp.Result.CurrentTime.Month())
// 	require.Equal(suite.T(), expMonth, resp.Result.TriggerTime.Month())
// 	require.Equal(suite.T(), expDayOfMonth, resp.Result.CurrentTime.Day())
// 	require.Equal(suite.T(), expDayOfMonth, resp.Result.TriggerTime.Day())
// 	require.Equal(suite.T(), expHour, resp.Result.CurrentTime.Hour())
// 	require.Equal(suite.T(), expHour, resp.Result.TriggerTime.Hour())
// 	require.Equal(suite.T(), expCurrMinute, resp.Result.CurrentTime.Minute())
// 	require.Equal(suite.T(), expTrigMinute, resp.Result.TriggerTime.Minute())
// 	require.Equal(suite.T(), expSecond, resp.Result.CurrentTime.Second())
// 	require.Equal(suite.T(), expSecond, resp.Result.TriggerTime.Second())
// 	require.Equal(suite.T(), expNanosec, resp.Result.CurrentTime.Nanosecond())
// 	require.Equal(suite.T(), expNanosec, resp.Result.TriggerTime.Nanosecond())
// }

// // Test Cancel Order Batch - Happy Path
// //
// // Test will succeed if request sent by client is well formatted, contains all input
// // all input parameters and if server response is well processed by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestCancelOrderBatchHappyPath() {

// 	// Test parameters
// 	orders := []string{"42", "43"}
// 	otp := "otp"

// 	// Predefined response
// 	expectedJSONResponse := `{
// 		"result": {
// 			"count": 2
// 		},
// 		"error": []
// 	}`

// 	// Expected response data
// 	expCount := 2

// 	// Configure Mock HTTP Server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Perform request
// 	resp, err := suite.client.CancelOrderBatch(CancelOrderBatchParameters{OrderIds: orders}, &SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Get and log client request
// 	req := suite.srv.PopRecordedRequest()
// 	require.NotNil(suite.T(), req)
// 	suite.T().Log(req)

// 	// Check client error and log response
// 	require.NoError(suite.T(), err)
// 	suite.T().Log(resp)

// 	// Check request
// 	// Check URL, Method and some Headers
// 	require.Contains(suite.T(), req.URL.Path, postCancelOrderBatch)
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Equal(suite.T(), suite.client.agent, req.UserAgent())
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.client.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))

// 	// Check request body
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))
// 	reqOrders := []string{
// 		req.Form.Get("orders[0]"),
// 		req.Form.Get("orders[1]"),
// 	}
// 	require.ElementsMatch(suite.T(), orders, reqOrders)

// 	// Check response
// 	require.Equal(suite.T(), expCount, resp.Result.Count)
// }

// // Test Cancel Order Batch - Empty list
// //
// // Test will succeed if request is rejected by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestCancelOrderBatchEmptyList() {

// 	// Perform request
// 	_, err := suite.client.CancelOrderBatch(CancelOrderBatchParameters{}, nil)
// 	require.Error(suite.T(), err)

// 	// Check no request sent
// 	require.Nil(suite.T(), suite.srv.PopRecordedRequest())
// }

// /*****************************************************************************/
// /* UNIT TESTS - USER FUNDING												 */
// /*****************************************************************************/

// // Test Get Deposit Methods - Empty limit
// func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositMethodsEmptyLimit() {

// 	// Test parameters
// 	asset := "XXBT"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": [
// 			{
// 				"method": "Bitcoin Lightning",
// 				"limit": false,
// 				"fee": "0.000000000"
// 			}
// 		]
// 	}`

// 	expectedMethod := "Bitcoin Lightning"
// 	expectedFee := "0.000000000"
// 	expectedAddrSetupFee := ""
// 	expectedGenAddr := ""
// 	expectedLimit := "false"

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetDepositMethods(GetDepositMethodsParameters{Asset: asset}, nil)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body contains asset
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))

// 	// Check response
// 	require.Equal(suite.T(), 1, len(resp.Result))
// 	require.Equal(suite.T(), expectedMethod, (resp.Result)[0].Method)
// 	require.Equal(suite.T(), expectedLimit, (resp.Result)[0].Limit)
// 	require.Equal(suite.T(), expectedFee, (resp.Result)[0].Fee)
// 	require.Equal(suite.T(), expectedAddrSetupFee, (resp.Result)[0].AddressSetupFee)
// 	require.Equal(suite.T(), expectedGenAddr, (resp.Result)[0].GenAddress)
// }

// // Test Get Deposit Methods - Predefined limit
// func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositMethodsFloatLimit() {

// 	// Test parameters
// 	asset := "XXBT"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": [
// 			{
// 				"method": "Bitcoin",
// 				"limit": "342.42",
// 				"fee": "4",
// 				"address-setup-fee": "1.2",
// 				"gen-address": true
// 			}
// 		]
// 	}`

// 	expectedLimit := "342.42"
// 	expectedMethod := "Bitcoin"
// 	expectedFee := "4"
// 	expectedAddrSetupFee := "1.2"
// 	expectedGenAddr := "true"

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetDepositMethods(GetDepositMethodsParameters{Asset: asset}, nil)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body contains asset
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))

// 	// Check response
// 	require.Equal(suite.T(), 1, len(resp.Result))
// 	require.Equal(suite.T(), expectedMethod, (*resp).Result[0].Method)
// 	require.Equal(suite.T(), expectedLimit, (*resp).Result[0].Limit)
// 	require.Equal(suite.T(), expectedFee, (*resp).Result[0].Fee)
// 	require.Equal(suite.T(), expectedAddrSetupFee, (*resp).Result[0].AddressSetupFee)
// 	require.Equal(suite.T(), expectedGenAddr, (*resp).Result[0].GenAddress)
// }

// // Test Get Deposit Methods - No limit field
// func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositMethodsNoLimit() {

// 	// Test parameters
// 	asset := "XXBT"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": [
// 			{
// 				"method": "Bitcoin",
// 				"fee": "4",
// 				"address-setup-fee": "1.2",
// 				"gen-address": true
// 			}
// 		]
// 	}`

// 	expectedMethod := "Bitcoin"
// 	expectedFee := "4"
// 	expectedAddrSetupFee := "1.2"
// 	expectedGenAddr := "true"

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetDepositMethods(GetDepositMethodsParameters{Asset: asset}, nil)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body contains asset
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))

// 	// Check response
// 	require.Equal(suite.T(), 1, len(resp.Result))
// 	require.Equal(suite.T(), expectedMethod, (*resp).Result[0].Method)
// 	require.Empty(suite.T(), (*resp).Result[0].Limit)
// 	require.Equal(suite.T(), expectedFee, (*resp).Result[0].Fee)
// 	require.Equal(suite.T(), expectedAddrSetupFee, (*resp).Result[0].AddressSetupFee)
// 	require.Equal(suite.T(), expectedGenAddr, (*resp).Result[0].GenAddress)
// }

// // Test Get Deposit Addresses - Happy path
// func (suite *KrakenAPIClientUnitTestSuite) TestGetDepositAddressesHappyPath() {

// 	// Test parameters
// 	asset := "XXBT"
// 	method := "Bitcoin"
// 	new := true
// 	otp := "NOPE"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": [
// 			{
// 				"address": "2N9fRkx5JTWXWHmXzZtvhQsufvoYRMq9ExV",
// 				"expiretm": "0",
// 				"new": true
// 			},
// 			{
// 				"address": "2NCpXUCEYr8ur9WXM1tAjZSem2w3aQeTcAo",
// 				"expiretm": "1658736768",
// 				"new": false
// 			},
// 			{
// 				"address": "2Myd4eaAW96ojk38A2uDK4FbioCayvkEgVq",
// 				"expiretm": "0"
// 			}
// 		]
// 	}`

// 	expectedAddressesLen := 3

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetDepositAddresses(
// 		GetDepositAddressesParameters{Asset: asset, Method: method},
// 		&GetDepositAddressesOptions{New: new},
// 		&SecurityOptions{SecondFactor: otp})

// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), method, req.Form.Get("method"))
// 	require.Equal(suite.T(), strconv.FormatBool(new), req.Form.Get("new"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expectedAddressesLen, len(resp.Result))
// 	for i, v := range resp.Result {
// 		require.NotEmpty(suite.T(), v.Address)
// 		require.GreaterOrEqual(suite.T(), v.Expiretm, int64(0))
// 		if i == 0 {
// 			require.True(suite.T(), v.New)
// 		} else {
// 			require.False(suite.T(), v.New)
// 		}
// 	}
// }

// // Test Get Status of Recent Deposits - Happy path
// func (suite *KrakenAPIClientUnitTestSuite) TestGetStatusOfRecentDepositsHappyPath() {

// 	// Test parameters
// 	asset := "XXBT"
// 	method := "Bitcoin"
// 	otp := "Nope"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": [
// 			{
// 				"method": "Bitcoin",
// 				"aclass": "currency",
// 				"asset": "XXBT",
// 				"refid": "AGBSO6T-UFMTTQ-I7KGS6",
// 				"txid": "AGBSO6T-UFMTTQ-I7KGS6",
// 				"info": "SEPA",
// 				"amount": "1.42",
// 				"fee": null,
// 				"time": 1658736768,
// 				"status": "Initial",
// 				"status-prop": "return"
// 			}
// 		]
// 	}`

// 	expectedAClass := "currency"
// 	expectedRefID := "AGBSO6T-UFMTTQ-I7KGS6"
// 	expectedTxID := "AGBSO6T-UFMTTQ-I7KGS6"
// 	expectedInfo := "SEPA"
// 	expectedAmount := "1.42"
// 	expectedFee := ""
// 	expectedTime := int64(1658736768)
// 	expectedStatus := TxStateInitial
// 	expectedStatusProp := TxStatusReturn
// 	expectedAddressesLen := 1

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetStatusOfRecentDeposits(
// 		GetStatusOfRecentDepositsParameters{Asset: asset},
// 		&GetStatusOfRecentDepositsOptions{Method: method},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), method, req.Form.Get("method"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expectedAddressesLen, len(resp.Result))
// 	for _, v := range resp.Result {
// 		require.Equal(suite.T(), method, v.Method)
// 		require.Equal(suite.T(), expectedAClass, v.AssetClass)
// 		require.Equal(suite.T(), asset, v.Asset)
// 		require.Equal(suite.T(), expectedRefID, v.ReferenceID)
// 		require.Equal(suite.T(), expectedTxID, v.TransactionID)
// 		require.Equal(suite.T(), expectedInfo, v.Info)
// 		require.Equal(suite.T(), expectedAmount, v.Amount)
// 		require.Equal(suite.T(), expectedFee, v.Fee)
// 		require.Equal(suite.T(), expectedTime, v.Time)
// 		require.Equal(suite.T(), expectedStatus, v.Status)
// 		require.Equal(suite.T(), expectedStatusProp, v.StatusProperty)
// 	}
// }

// // Test Get Withdrawal Information - Happy path
// func (suite *KrakenAPIClientUnitTestSuite) TestGetWithdrawalInformationHappyPath() {

// 	// Test parameters
// 	asset := "XXBT"
// 	amount := "42.999999"
// 	key := "withdrawal_address"
// 	otp := "Nope"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": {
// 			"method": "Bitcoin",
// 			"limit": "332.00956139",
// 			"amount": "42.999999",
// 			"fee": "0.00015000"
// 		}
// 	}`

// 	expectedFee := "0.00015000"
// 	expectedLimit := "332.00956139"
// 	expectedMethod := "Bitcoin"

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetWithdrawalInformation(
// 		GetWithdrawalInformationParameters{Asset: asset, Amount: amount, Key: key},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), amount, req.Form.Get("amount"))
// 	require.Equal(suite.T(), key, req.Form.Get("key"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expectedMethod, resp.Result.Method)
// 	require.Equal(suite.T(), amount, resp.Result.Amount)
// 	require.Equal(suite.T(), expectedFee, resp.Result.Fee)
// 	require.Equal(suite.T(), expectedLimit, resp.Result.Limit)
// }

// // Test Withdraw Funds - Happy path
// func (suite *KrakenAPIClientUnitTestSuite) TestWithdrawFundsHappyPath() {

// 	// Test parameters
// 	asset := "XXBT"
// 	waddrn := "nevermind"
// 	amount := "42.999999"
// 	otp := "NOPE"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": {
// 			"refid": "AGBSO6T-UFMTTQ-I7KGS6"
// 		}
// 	}`

// 	expectedRefID := "AGBSO6T-UFMTTQ-I7KGS6"

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.WithdrawFunds(
// 		WithdrawFundsParameters{Asset: asset, Amount: amount, Key: waddrn},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), amount, req.Form.Get("amount"))
// 	require.Equal(suite.T(), waddrn, req.Form.Get("key"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expectedRefID, resp.Result.ReferenceID)
// }

// // Test Get Status of Recent Withdrawal - Happy path
// func (suite *KrakenAPIClientUnitTestSuite) TestGetStatusOfRecentWithdrawalsHappyPath() {

// 	// Test parameters
// 	asset := "XXBT"
// 	method := "Bitcoin"
// 	otp := "NOPE"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": [
// 			{
// 				"method": "Bitcoin",
// 				"aclass": "currency",
// 				"asset": "XXBT",
// 				"refid": "AGBZNBO-5P2XSB-RFVF6J",
// 				"txid": "THVRQM-33VKH-UCI7BS",
// 				"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
// 				"amount": "0.72485000",
// 				"fee": "0.00015000",
// 				"time": 1617014586,
// 				"status": "Pending"
// 			},
// 			{
// 				"method": "Bitcoin",
// 				"aclass": "currency",
// 				"asset": "XXBT",
// 				"refid": "AGBSO6T-UFMTTQ-I7KGS6",
// 				"txid": "KLETXZ-33VKH-UCI7BS",
// 				"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
// 				"amount": "0.72485000",
// 				"fee": "0.00015000",
// 				"time": 1617015423,
// 				"status": "Failure",
// 				"status-prop": "canceled"
// 			}
// 		]
// 	}`

// 	expectedResultLen := 2
// 	expectedAClass := "currency"
// 	expectedInfo := "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR"
// 	expectedAmount := "0.72485000"
// 	expectedFee := "0.00015000"

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetStatusOfRecentWithdrawals(
// 		GetStatusOfRecentWithdrawalsParameters{Asset: asset},
// 		&GetStatusOfRecentWithdrawalsOptions{Method: method},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), method, req.Form.Get("method"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expectedResultLen, len(resp.Result))
// 	for _, v := range resp.Result {
// 		require.Equal(suite.T(), method, v.Method)
// 		require.Equal(suite.T(), expectedAClass, v.AssetClass)
// 		require.Equal(suite.T(), asset, v.Asset)
// 		require.NotEmpty(suite.T(), v.ReferenceID)
// 		require.NotEmpty(suite.T(), v.TransactionID)
// 		require.Equal(suite.T(), expectedInfo, v.Info)
// 		require.Equal(suite.T(), expectedAmount, v.Amount)
// 		require.Equal(suite.T(), expectedFee, v.Fee)
// 		require.GreaterOrEqual(suite.T(), v.Time, int64(0))
// 		require.True(suite.T(), v.Status == TxStatePending || v.Status == TxStateFailure)
// 		require.True(suite.T(), v.StatusProperty == "" || v.StatusProperty == TxCanceled)
// 	}
// }

// // Test Request Withdrawal Cancellation - Happy path
// func (suite *KrakenAPIClientUnitTestSuite) TestRequestWithdrawalCancellationHappyPath() {

// 	// Test parameters
// 	asset := "XXBT"
// 	refid := "AGBZNBO-5P2XSB-RFVF6J"
// 	otp := "NOPE"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": true
// 	}`

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.RequestWithdrawalCancellation(
// 		RequestWithdrawalCancellationParameters{Asset: asset, ReferenceId: refid},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), refid, req.Form.Get("refid"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.True(suite.T(), resp.Result)
// }

// // Test Request Wallet Transfer - Happy path
// func (suite *KrakenAPIClientUnitTestSuite) TestRequestWalletTransferHappyPath() {

// 	// Test parameters
// 	asset := "XXBT"
// 	from := "Spot Wallet"
// 	to := "Future Wallet"
// 	amount := "42.24"
// 	otp := "NOPE"

// 	// Expected API response from API documentation
// 	expectedJSONResponse := `{
// 		"error": [],
// 		"result": {
// 			"refid": "BOG5AE5-KSCNR4-VPNPEV"
// 		}
// 	}`

// 	expectedRefID := "BOG5AE5-KSCNR4-VPNPEV"

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.RequestWalletTransfer(
// 		RequestWalletTransferParameters{Asset: asset, From: from, To: to, Amount: amount},
// 		&SecurityOptions{SecondFactor: otp})
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request body
// 	require.Equal(suite.T(), asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), from, req.Form.Get("from"))
// 	require.Equal(suite.T(), to, req.Form.Get("to"))
// 	require.Equal(suite.T(), amount, req.Form.Get("amount"))
// 	require.Equal(suite.T(), otp, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expectedRefID, resp.Result.ReferenceID)
// }

// /*****************************************************************************/
// /* USER STAKING TESTS                                                        */
// /*****************************************************************************/

// // TestStakeAssetHappyPath is a unit test for Stake Asset. The test will succeed
// // if client send a valid request and if a valid predefined response from server
// // is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestStakeAssetHappyPath() {

// 	// Test params
// 	params := StakeAssetParameters{
// 		Asset:  "XXBT",
// 		Amount: "0.01",
// 		Method: "offchain",
// 	}
// 	secopts := SecurityOptions{
// 		SecondFactor: "NOPE",
// 	}

// 	// Server response
// 	expectedJSONResponse := `
// 	{
// 		"error": [ ],
// 		"result": {
// 		"refid": "BOG5AE5-KSCNR4-VPNPEV"
// 		}
// 	}`

// 	expResp := &StakeAssetResponse{
// 		KrakenAPIResponse: KrakenAPIResponse{Error: []string{}},
// 		Result: struct {
// 			ReferenceID string "json:\"refid\""
// 		}{ReferenceID: "BOG5AE5-KSCNR4-VPNPEV"},
// 	}

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.StakeAsset(params, &secopts)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Contains(suite.T(), req.URL.Path, postStakeAsset)
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
// 	require.Equal(suite.T(), params.Asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), params.Amount, req.Form.Get("amount"))
// 	require.Equal(suite.T(), params.Method, req.Form.Get("method"))

// 	// Check response
// 	require.Equal(suite.T(), expResp, resp)
// }

// // TestStakeAssetErrPath is a unit test for Stake Asset. The test will succeed
// // if an error response from server is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestStakeAssetErrPath() {

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.StakeAsset(StakeAssetParameters{}, nil)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Check response
// 	require.Error(suite.T(), err)
// 	require.Nil(suite.T(), resp)
// }

// // TestUnstakeAssetHappyPath is a unit test for Unstake Asset. The test will succeed
// // if client send a valid request and if a valid predefined response from server
// // is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestUnstakeAssetHappyPath() {

// 	// Test params
// 	params := UnstakeAssetParameters{
// 		Asset:  "XXBT",
// 		Amount: "0.01",
// 	}
// 	secopts := SecurityOptions{
// 		SecondFactor: "NOPE",
// 	}

// 	// Server response
// 	expectedJSONResponse := `
// 	{
// 		"error": [ ],
// 		"result": {
// 		"refid": "BOG5AE5-KSCNR4-VPNPEV"
// 		}
// 	}`

// 	expResp := &UnstakeAssetResponse{
// 		KrakenAPIResponse: KrakenAPIResponse{Error: []string{}},
// 		Result: struct {
// 			ReferenceID string "json:\"refid\""
// 		}{ReferenceID: "BOG5AE5-KSCNR4-VPNPEV"},
// 	}

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.UnstakeAsset(params, &secopts)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Contains(suite.T(), req.URL.Path, postUnstakeAsset)
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))
// 	require.Equal(suite.T(), params.Asset, req.Form.Get("asset"))
// 	require.Equal(suite.T(), params.Amount, req.Form.Get("amount"))

// 	// Check response
// 	require.Equal(suite.T(), expResp, resp)
// }

// // TestUnstakeAssetErrPath is a unit test for Unstake Asset. The test will succeed
// // if an error response from server is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestUnstakeAssetErrPath() {

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.UnstakeAsset(UnstakeAssetParameters{}, nil)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Check response
// 	require.Error(suite.T(), err)
// 	require.Nil(suite.T(), resp)
// }

// // TestListOfStakeableAssetsHappyPath is a unit test for List Of Stakeable Assets.
// // Test will succeed if client send a valid request and if a valid predefined
// // response from server is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestListOfStakeableAssetsHappyPath() {

// 	// Test params
// 	secopts := SecurityOptions{
// 		SecondFactor: "NOPE",
// 	}

// 	// Server response
// 	expectedJSONResponse := `
// 	{
// 		"result": [
// 			{
// 				"method": "fake",
// 				"asset": "FAKE",
// 				"staking_asset": "FAKE.S",
// 				"rewards": {
// 			  		"reward": "99.95",
// 			  		"type": "percentage"
// 				},
// 				"on_chain": false,
// 				"can_stake": false,
// 				"can_unstake": false,
// 				"minimum_amount": {
// 				  	"staking": "0.0000000000",
// 				  	"unstaking": "0.0000000000"
// 				},
// 				"lock": {
// 					"staking": {
// 						"days": 0.5,
// 						"percentage": 42.42
// 					},
// 					"unstaking": {
// 						"days": 0.5,
// 						"percentage": 42.42
// 					},
// 					"lockup": {
// 						"days": 0.5,
// 						"percentage": 42.42
// 					}
// 				},
// 				"enabled_for_user": false,
// 				"disabled": true
// 		  	},
// 		  	{
// 				"method": "kusama-staked",
// 				"asset": "KSM",
// 				"staking_asset": "KSM.S",
// 				"rewards": {
// 				 	"reward": "12.00",
// 			  		"type": "percentage"
// 				}
// 		  	},
// 			{
// 				"method": "nope",
// 				"asset": "NOPE",
// 				"staking_asset": "NOPE.S",
// 				"rewards": {
// 				 	"reward": "12.00",
// 			  		"type": "percentage"
// 				},
// 				"minimum_amount": {},
// 				"lock": {}
// 		  	}
// 		],
// 		"error": []
// 	}`

// 	expData := ListOfStakeableAssetsResponse{
// 		Result: []StakingAssetInformation{
// 			{
// 				Asset:        "FAKE",
// 				StakingAsset: "FAKE.S",
// 				Method:       "fake",
// 				OnChain:      false,
// 				CanStake:     false,
// 				CanUnstake:   false,
// 				MinAmount: &StakingAssetMinAmount{
// 					Unstaking: "0.0000000000",
// 					Staking:   "0.0000000000",
// 				},
// 				Lock: &StakingAssetLockup{
// 					Unstaking: &StakingAssetLockPeriod{
// 						Days:       0.5,
// 						Percentage: 42.42,
// 					},
// 					Staking: &StakingAssetLockPeriod{
// 						Days:       0.5,
// 						Percentage: 42.42,
// 					},
// 					Lockup: &StakingAssetLockPeriod{
// 						Days:       0.5,
// 						Percentage: 42.42,
// 					},
// 				},
// 				EnabledForUser: false,
// 				Disabled:       true,
// 				Rewards: StakingAssetReward{
// 					Reward: "99.95",
// 					Type:   "percentage",
// 				},
// 			},
// 			{
// 				Asset:          "KSM",
// 				StakingAsset:   "KSM.S",
// 				Method:         "kusama-staked",
// 				OnChain:        true,
// 				CanStake:       true,
// 				CanUnstake:     true,
// 				MinAmount:      nil,
// 				Lock:           nil,
// 				EnabledForUser: true,
// 				Disabled:       false,
// 				Rewards: StakingAssetReward{
// 					Reward: "12.00",
// 					Type:   "percentage",
// 				},
// 			},
// 			{
// 				Asset:        "NOPE",
// 				StakingAsset: "NOPE.S",
// 				Method:       "nope",
// 				OnChain:      true,
// 				CanStake:     true,
// 				CanUnstake:   true,
// 				MinAmount: &StakingAssetMinAmount{
// 					Unstaking: "0",
// 					Staking:   "0",
// 				},
// 				Lock:           &StakingAssetLockup{},
// 				EnabledForUser: true,
// 				Disabled:       false,
// 				Rewards: StakingAssetReward{
// 					Reward: "12.00",
// 					Type:   "percentage",
// 				},
// 			},
// 		},
// 	}

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.ListOfStakeableAssets(&secopts)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Contains(suite.T(), req.URL.Path, postListOfStakeableAssets)
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expData.Result, resp.Result)
// }

// // TestListOfStakeableAssetsErrPath is a unit test for List Of Stakeable Assets. Test
// // will succeed if an error response from server is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestListOfStakeableAssetsErrPath() {

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.ListOfStakeableAssets(nil)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Check response
// 	require.Error(suite.T(), err)
// 	require.Nil(suite.T(), resp)
// }

// // TestGetPendingStackingTransactionsHappyPath is a unit test for Get Pending Stacking
// // Transactions. Test will succeed if client sends a valid request and if a valid server
// // response is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestGetPendingStackingTransactionsHappyPath() {

// 	// Test params
// 	secopts := SecurityOptions{
// 		SecondFactor: "NOPE",
// 	}

// 	// Server response
// 	expectedJSONResponse := `
// 	{
// 		"result": [
// 			{
// 				"method": "ada-staked",
// 				"aclass": "currency",
// 				"asset": "ADA.S",
// 				"refid": "RUSB7W6-ESIXUX-K6PVTM",
// 				"amount": "0.34844300",
// 				"fee": "0.00000000",
// 				"time": 1622967367,
// 				"status": "Initial",
// 				"type": "bonding"
// 		 	},
// 		  	{
// 				"method": "xtz-staked",
// 				"aclass": "currency",
// 				"asset": "XTZ.S",
// 				"refid": "RUCXX7O-6MWQBO-CQPGAX",
// 				"amount": "0.00746900",
// 				"fee": "0.00000000",
// 				"time": 1623074402,
// 				"status": "Initial",
// 				"type": "bonding"
// 		  	}
// 		],
// 		"error": []
// 	  }`

// 	expData := GetPendingStakingTransactionsResponse{
// 		Result: []StakingTransactionInfo{
// 			{
// 				ReferenceId: "RUSB7W6-ESIXUX-K6PVTM",
// 				Asset:       "ADA.S",
// 				AssetClass:  "currency",
// 				Type:        "bonding",
// 				Method:      "ada-staked",
// 				Amount:      "0.34844300",
// 				Fee:         "0.00000000",
// 				Timestamp:   1622967367,
// 				Status:      "Initial",
// 			},
// 			{
// 				ReferenceId: "RUCXX7O-6MWQBO-CQPGAX",
// 				Asset:       "XTZ.S",
// 				AssetClass:  "currency",
// 				Type:        "bonding",
// 				Method:      "xtz-staked",
// 				Amount:      "0.00746900",
// 				Fee:         "0.00000000",
// 				Timestamp:   1623074402,
// 				Status:      "Initial",
// 			},
// 		},
// 	}

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetPendingStakingTransactions(&secopts)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Contains(suite.T(), req.URL.Path, postGetPendingStakingTransactions)
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expData.Result, resp.Result)
// }

// // TestGetPendingStackingTransactionsErrPath is a unit test for Get Pending Stacking
// // Transactions. Test will succeed if an error response from server is well handled
// // by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestGetPendingStackingTransactionsErrPath() {

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.GetPendingStakingTransactions(nil)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Check response
// 	require.Error(suite.T(), err)
// 	require.Nil(suite.T(), resp)
// }

// // TestListOfStackingTransactionsHappyPath is a unit test for List Of Stacking Transactions.
// // Test will succeed if client sends a valid request and if a valid server response is well
// // handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestListOfStackingTransactionsHappyPath() {

// 	// Test params
// 	secopts := SecurityOptions{
// 		SecondFactor: "NOPE",
// 	}

// 	// Server response
// 	expectedJSONResponse := `
// 	{
// 		"result": [
// 			{
// 				"method": "ada-staked",
// 				"aclass": "currency",
// 				"asset": "ADA.S",
// 				"refid": "RUSB7W6-ESIXUX-K6PVTM",
// 				"amount": "0.34844300",
// 				"fee": "0.00000000",
// 				"time": 1622967367,
// 				"status": "Initial",
// 				"type": "bonding",
// 				"bond_start": 1622971496,
// 				"bond_end": 1622971496
// 		 	},
// 		  	{
// 				"method": "xtz-staked",
// 				"aclass": "currency",
// 				"asset": "XTZ.S",
// 				"refid": "RUCXX7O-6MWQBO-CQPGAX",
// 				"amount": "0.00746900",
// 				"fee": "0.00000000",
// 				"time": 1623074402,
// 				"status": "Initial",
// 				"type": "bonding"
// 		  	}
// 		],
// 		"error": []
// 	  }`

// 	expData := ListOfStakingTransactionsResponse{
// 		Result: []StakingTransactionInfo{
// 			{
// 				ReferenceId: "RUSB7W6-ESIXUX-K6PVTM",
// 				Asset:       "ADA.S",
// 				AssetClass:  "currency",
// 				Type:        "bonding",
// 				Method:      "ada-staked",
// 				Amount:      "0.34844300",
// 				Fee:         "0.00000000",
// 				Timestamp:   1622967367,
// 				Status:      "Initial",
// 				BondStart:   new(int64),
// 				BondEnd:     new(int64),
// 			},
// 			{
// 				ReferenceId: "RUCXX7O-6MWQBO-CQPGAX",
// 				Asset:       "XTZ.S",
// 				AssetClass:  "currency",
// 				Type:        "bonding",
// 				Method:      "xtz-staked",
// 				Amount:      "0.00746900",
// 				Fee:         "0.00000000",
// 				Timestamp:   1623074402,
// 				Status:      "Initial",
// 				BondStart:   nil,
// 				BondEnd:     nil,
// 			},
// 		},
// 	}
// 	*expData.Result[0].BondStart = 1622971496
// 	*expData.Result[0].BondEnd = 1622971496

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status:  http.StatusOK,
// 		Headers: http.Header{"Content-Type": []string{"application/json"}},
// 		Body:    []byte(expectedJSONResponse),
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.ListOfStakingTransactions(&secopts)
// 	require.NoError(suite.T(), err)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Log response
// 	suite.T().Logf("Response decoded by client : %#v", resp)

// 	// Check request
// 	require.Equal(suite.T(), http.MethodPost, req.Method)
// 	require.Contains(suite.T(), req.URL.Path, postListOfStakingTransactions)
// 	require.Equal(suite.T(), "application/x-www-form-urlencoded", req.Header.Get(managedHeaderContentType))
// 	require.Equal(suite.T(), suite.key, req.Header.Get(managedHeaderAPIKey))
// 	require.NotEmpty(suite.T(), req.Header.Get(managedHeaderAPISign))
// 	require.NotEmpty(suite.T(), req.Form.Get("nonce"))
// 	require.Equal(suite.T(), secopts.SecondFactor, req.Form.Get("otp"))

// 	// Check response
// 	require.Equal(suite.T(), expData.Result, resp.Result)
// }

// // TestListOfStackingTransactionsErrPath is a unit test for List Of Stacking Transactions.
// // Test will succeed if an error response from server is well handled by client.
// func (suite *KrakenAPIClientUnitTestSuite) TestListOfStackingTransactionsErrPath() {

// 	// Configure mock http server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: http.StatusBadRequest,
// 	})

// 	// Call API endpoint
// 	resp, err := suite.client.ListOfStakingTransactions(nil)

// 	// Log request
// 	req := suite.srv.PopRecordedRequest()
// 	suite.T().Logf("Request received by mock HTTP server from API client. Got %#v", req)

// 	// Check response
// 	require.Error(suite.T(), err)
// 	require.Nil(suite.T(), resp)
// }

// /*****************************************************************************/
// /* GENERAL API CLIENT TESTS                                                  */
// /*****************************************************************************/

// // Test if client implents client interface
// func (suite *KrakenAPIClientUnitTestSuite) TestClientInterfaceImplemented() {

// 	iface := reflect.TypeOf((*KrakenAPIClientIface)(nil)).Elem()
// 	ok := reflect.TypeOf(suite.client).Implements(iface)
// 	require.True(suite.T(), ok, "KranAPIClient does not fully implement KrakenAPIClient interface")
// }

// // Test when the client receives a 503 HTTP Error from server
// func (suite *KrakenAPIClientUnitTestSuite) TestClientReceivesServiceUnavailableError() {

// 	// Expected error status
// 	expectedStatus := http.StatusServiceUnavailable

// 	// Configure server
// 	suite.srv.AddResponse(&mockhttpserver.ServerResponse{
// 		Status: expectedStatus,
// 	})

// 	// Call API endpoint
// 	_, err := suite.client.GetSystemStatus()
// 	require.Error(suite.T(), err)
// }

// // Test when the client call a non existing server
// func (suite *KrakenAPIClientUnitTestSuite) TestClientCallNotExistingServer() {

// 	// Create client to non-existing endpoint
// 	client := NewPublicWithOptions(&KrakenAPIClientOptions{BaseURL: "http://localhost:42422"})

// 	// Call API endpoint
// 	_, err := client.GetSystemStatus()
// 	require.Error(suite.T(), err)
// }

// // Test when the client experience a request timeout
// func (suite *KrakenAPIClientUnitTestSuite) TestClientRequestTimeout() {

// 	// Create client with a timeout of 1 nanosecond
// 	client := NewPublicWithOptions(&KrakenAPIClientOptions{
// 		BaseURL: suite.srv.GetMockHTTPServerBaseURL(),
// 		Client:  &http.Client{Timeout: time.Duration(1)},
// 	})

// 	// Call API endpoint
// 	_, err := client.GetSystemStatus()
// 	require.Error(suite.T(), err)
// }

// /*****************************************************************************/
// /* UTILITY FUNCTION TESTS													 */
// /*****************************************************************************/

// // Test the method used to forge a signature for a request
// func (suite *KrakenAPIClientUnitTestSuite) TestRequestSignature() {

// 	// Signature parameters
// 	secret, _ := base64.StdEncoding.DecodeString("kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg==")
// 	nonce := "1616492376594"
// 	resource := "/0/private/AddOrder"
// 	encodedPayload := make(url.Values)
// 	encodedPayload.Set("nonce", nonce)
// 	encodedPayload.Set("ordertype", "limit")
// 	encodedPayload.Set("pair", "XBTUSD")
// 	encodedPayload.Set("price", "37500")
// 	encodedPayload.Set("type", "buy")
// 	encodedPayload.Set("volume", "1.25")

// 	// Expected signature - from documentation
// 	// https://docs.kraken.com/rest/#section/Authentication/Headers-and-Signature
// 	expected := "4/dpxb3iT4tp/ZCVEwSnEsLxx0bqyhLpdfOpc6fn7OR8+UClSV5n9E6aSS8MPtnRfp32bAb0nmbRn6H8ndwLUQ=="

// 	// Forge & compare signature
// 	require.Equal(suite.T(), expected, GetKrakenSignature(resource, encodedPayload, secret))
// }
