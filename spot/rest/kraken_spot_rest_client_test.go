package rest

import (
	"context"
	"fmt"
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
	"github.com/gbdevw/purple-goctopus/spot/rest/earn"
	"github.com/gbdevw/purple-goctopus/spot/rest/funding"
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

// Test interface compliance with KrakenSpotRESTClientIface
func (suite *KrakenSpotRESTClientTestSuite) TestInterfaceCompliance() {
	var instance interface{} = NewKrakenSpotRESTClient(nil, nil)
	_, ok := instance.(KrakenSpotRESTClientIface)
	require.True(suite.T(), ok)
}

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

/*************************************************************************************************/
/* UNIT TESTS - TRADING                                                                          */
/*************************************************************************************************/

// Test AddOrder when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestAddOrder() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := trading.AddOrderRequestParameters{
		Pair: "XXBTZUSD",
		Order: trading.Order{
			UserReference:      new(int64),
			OrderType:          string(trading.StopLossLimit),
			Type:               string(trading.Sell),
			Volume:             "0.1",
			DisplayedVolume:    "0.001",
			Price:              "42500.0",
			Price2:             "43000.0",
			Trigger:            string(trading.Index),
			Leverage:           "5:1",
			ReduceOnly:         true,
			StpType:            string(trading.STPCancelBoth),
			OrderFlags:         strings.Join([]string{string(trading.OFlagFeeInQuote), string(account.OFlagNoMarketPriceProtection)}, ","),
			TimeInForce:        string(trading.ImmediateOrCancel),
			ScheduledStartTime: "+2",
			ExpirationTime:     "+2",
			Close:              &trading.CloseOrder{OrderType: string(trading.StopLossLimit), Price: "42500.0", Price2: "43000.0"},
		},
	}

	// Expected options
	options := &trading.AddOrderRequestOptions{
		Validate: true,
		Deadline: time.Now().UTC().Add(15 * time.Second),
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
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
	expectedOrderDescr := "buy 1.25000000 XBTUSD @ limit 27500.0"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.AddOrder(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedOrderDescr, resp.Result.Description.Order)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, addOrderPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Validate), record.Request.Form.Get("validate"))
	require.Equal(suite.T(), options.Deadline.Format(time.RFC3339), record.Request.Form.Get("deadline"))
	require.Equal(suite.T(), strconv.FormatInt(*params.Order.UserReference, 10), record.Request.Form.Get("userref"))
	require.Equal(suite.T(), params.Order.OrderType, record.Request.Form.Get("ordertype"))
	require.Equal(suite.T(), params.Order.Type, record.Request.Form.Get("type"))
	require.Equal(suite.T(), params.Order.Volume, record.Request.Form.Get("volume"))
	require.Equal(suite.T(), params.Order.DisplayedVolume, record.Request.Form.Get("displayvol"))
	require.Equal(suite.T(), params.Pair, record.Request.Form.Get("pair"))
	require.Equal(suite.T(), params.Order.Price, record.Request.Form.Get("price"))
	require.Equal(suite.T(), params.Order.Price2, record.Request.Form.Get("price2"))
	require.Equal(suite.T(), params.Order.Trigger, record.Request.Form.Get("trigger"))
	require.Equal(suite.T(), params.Order.Leverage, record.Request.Form.Get("leverage"))
	require.Equal(suite.T(), strconv.FormatBool(params.Order.ReduceOnly), record.Request.Form.Get("reduce_only"))
	require.Equal(suite.T(), params.Order.StpType, record.Request.Form.Get("stptype"))
	require.Equal(suite.T(), params.Order.OrderFlags, record.Request.Form.Get("oflags"))
	require.Equal(suite.T(), params.Order.TimeInForce, record.Request.Form.Get("timeinforce"))
	require.Equal(suite.T(), params.Order.ScheduledStartTime, record.Request.Form.Get("starttm"))
	require.Equal(suite.T(), params.Order.ExpirationTime, record.Request.Form.Get("expiretm"))
	require.Equal(suite.T(), params.Order.Close.OrderType, record.Request.Form.Get("close[ordertype]"))
	require.Equal(suite.T(), params.Order.Close.Price, record.Request.Form.Get("close[price]"))
	require.Equal(suite.T(), params.Order.Close.Price2, record.Request.Form.Get("close[price2]"))
}

// Test AddOrderBatch when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestAddOrderBatch() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := trading.AddOrderBatchRequestParameters{
		Pair: "XXBTZUSD",
		Orders: []trading.Order{
			{
				UserReference:      new(int64),
				OrderType:          string(trading.StopLossLimit),
				Type:               string(trading.Sell),
				Volume:             "0.1",
				DisplayedVolume:    "0.001",
				Price:              "42500.0",
				Price2:             "43000.0",
				Trigger:            string(trading.Index),
				Leverage:           "5:1",
				ReduceOnly:         true,
				StpType:            string(trading.STPCancelBoth),
				OrderFlags:         strings.Join([]string{string(trading.OFlagFeeInQuote), string(account.OFlagNoMarketPriceProtection)}, ","),
				TimeInForce:        string(trading.ImmediateOrCancel),
				ScheduledStartTime: "+2",
				ExpirationTime:     "+2",
				Close:              &trading.CloseOrder{OrderType: string(trading.StopLossLimit), Price: "42500.0", Price2: "43000.0"},
			},
			{
				UserReference:      new(int64),
				OrderType:          string(trading.Limit),
				Type:               string(trading.Buy),
				Volume:             "1",
				DisplayedVolume:    "0.005",
				Price:              "44500.0",
				Price2:             "4000.0",
				Trigger:            string(trading.Last),
				Leverage:           "3:1",
				ReduceOnly:         true,
				StpType:            string(trading.STPCancelNewest),
				OrderFlags:         strings.Join([]string{string(trading.OFlagFeeInBase), string(account.OFlagNoMarketPriceProtection)}, ","),
				TimeInForce:        string(trading.GoodTilCanceled),
				ScheduledStartTime: "+6",
				ExpirationTime:     "+6",
				Close:              &trading.CloseOrder{OrderType: string(trading.Market), Price: "40500.0", Price2: "40000.0"},
			},
		},
	}

	// Expected options
	options := &trading.AddOrderBatchOptions{
		Validate: true,
		Deadline: time.Now().UTC().Add(15 * time.Second),
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "orders": [
			{
			  "txid": "O5OR23-ADFAD-Y2G61C",
			  "descr": {
				"order": "buy 0.80300000 XBTUSD @ limit 28300.0"
			  },
			  "close": "close position @ stop loss 27000.0 -> limit 26000.0"
			},
			{
			  "txid": "9K6KFS-5H3PL-XBRC7A",
			  "descr": {
				"order": "sell 0.10500000 XBTUSD @ limit 36000.0"
			  }
			}
		  ]
		}
	}`
	expectedCount := 2
	expectedItem1Descr := "sell 0.10500000 XBTUSD @ limit 36000.0"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.AddOrderBatch(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Len(suite.T(), resp.Result.Orders, expectedCount)
	require.Equal(suite.T(), expectedItem1Descr, resp.Result.Orders[1].Description.Order)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, addOrderBatchPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Validate), record.Request.Form.Get("validate"))
	require.Equal(suite.T(), options.Deadline.Format(time.RFC3339), record.Request.Form.Get("deadline"))
	require.Equal(suite.T(), params.Pair, record.Request.Form.Get("pair"))
	for index, iorder := range params.Orders {
		require.Equal(suite.T(), strconv.FormatInt(*iorder.UserReference, 10), record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "userref")))
		require.Equal(suite.T(), iorder.OrderType, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "ordertype")))
		require.Equal(suite.T(), iorder.Type, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "type")))
		require.Equal(suite.T(), iorder.Volume, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "volume")))
		require.Equal(suite.T(), iorder.DisplayedVolume, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "displayvol")))
		require.Equal(suite.T(), iorder.Price, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "price")))
		require.Equal(suite.T(), iorder.Price2, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "price2")))
		require.Equal(suite.T(), iorder.Trigger, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "trigger")))
		require.Equal(suite.T(), iorder.Leverage, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "leverage")))
		require.Equal(suite.T(), strconv.FormatBool(iorder.ReduceOnly), record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "reduce_only")))
		require.Equal(suite.T(), iorder.StpType, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "stptype")))
		require.Equal(suite.T(), iorder.OrderFlags, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "oflags")))
		require.Equal(suite.T(), iorder.TimeInForce, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "timeinforce")))
		require.Equal(suite.T(), iorder.ScheduledStartTime, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "starttm")))
		require.Equal(suite.T(), iorder.ExpirationTime, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s]", index, "expiretm")))
		require.Equal(suite.T(), iorder.Close.OrderType, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "ordertype")))
		require.Equal(suite.T(), iorder.Close.Price, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price")))
		require.Equal(suite.T(), iorder.Close.Price2, record.Request.Form.Get(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price2")))
	}
}

// Test EditOrder when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestEditOrder() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := trading.EditOrderRequestParameters{
		Pair: "XXBTZUSD",
		Id:   "OHYO67-6LP66-HMQ437",
	}

	// Expected options
	options := &trading.EditOrderRequestOptions{
		NewUserReference: "5",
		NewVolume:        "5",
		Price:            "42",
		Price2:           "42",
		OFlags:           []string{string(account.OFlagFeeInBase)},
		Validate:         true,
		CancelResponse:   true,
		Deadline:         time.Now().UTC().Add(15 * time.Second),
	}

	// Expected API response from API documentation
	expectedJSONResponse := `
	{
		"error": [],
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
	expectedOrderDescr := "buy 0.00030000 XXBTZGBP @ limit 19500.0"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.EditOrder(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedOrderDescr, resp.Result.Description.Order)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, editOrderPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Pair, record.Request.Form.Get("pair"))
	require.Equal(suite.T(), params.Id, record.Request.Form.Get("txid"))
	require.Equal(suite.T(), options.NewUserReference, record.Request.Form.Get("userref"))
	require.Equal(suite.T(), options.NewVolume, record.Request.Form.Get("volume"))
	require.Equal(suite.T(), options.NewDisplayedVolume, record.Request.Form.Get("displayvol"))
	require.Equal(suite.T(), options.Price, record.Request.Form.Get("price"))
	require.Equal(suite.T(), options.Price2, record.Request.Form.Get("price2"))
	require.Equal(suite.T(), strings.Join(options.OFlags, ","), record.Request.Form.Get("oflags"))
	require.Equal(suite.T(), strconv.FormatBool(options.CancelResponse), record.Request.Form.Get("cancel_response"))
	require.Equal(suite.T(), strconv.FormatBool(options.Validate), record.Request.Form.Get("validate"))
	require.Equal(suite.T(), options.Deadline.Format(time.RFC3339), record.Request.Form.Get("deadline"))
}

// Test CancelOrder when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestCancelOrder() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := trading.CancelOrderRequestParameters{
		Id: "OHYO67-6LP66-HMQ437",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "count": 1
		}
	}`
	expectedCountCount := 1

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.CancelOrder(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedCountCount, resp.Result.Count)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, cancelOrderPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Id, record.Request.Form.Get("txid"))
}

// Test CancelAllOrders when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestCancelAllOrders() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "count": 4
		}
	}`
	expectedCountCount := 4

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.CancelAllOrders(context.Background(), expectedNonce, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedCountCount, resp.Result.Count)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, cancelAllOrdersPath)
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

// Test CancelAllOrdersAfterX when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestCancelAllOrdersAfterX() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := trading.CancelAllOrdersAfterXRequestParameters{
		Timeout: 60,
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "currentTime": "2023-03-24T17:41:56Z",
		  "triggerTime": "2023-03-24T17:42:56Z"
		}
	}`
	expectedCurrentTime := "2023-03-24T17:41:56Z"
	expectedTriggerTime := "2023-03-24T17:42:56Z"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.CancelAllOrdersAfterX(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedCurrentTime, resp.Result.CurrentTime.Format(time.RFC3339))
	require.Equal(suite.T(), expectedTriggerTime, resp.Result.TriggerTime.Format(time.RFC3339))

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, cancelAllOrdersPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatInt(params.Timeout, 10), record.Request.Form.Get("timeout"))
}

// Test CancelOrderBatch when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestCancelOrderBatch() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := trading.CancelOrderBatchRequestParameters{
		OrderIds: []string{"1", "2", "3"},
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "count": 3
		}
	}`
	expectedCount := 3

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.CancelOrderBatch(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedCount, resp.Result.Count)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, cancelOrderBatchPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strings.Join(params.OrderIds, ","), record.Request.Form.Get("orders"))
}

/*************************************************************************************************/
/* UNIT TESTS - FUNDING                                                                          */
/*************************************************************************************************/

// Test GetDepositMethods when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetDepositMethods() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := funding.GetDepositMethodsRequestParameters{
		Asset: "XXBT",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": [
		  {
			"method": "Bitcoin",
			"limit": false,
			"fee": "0.0000000000",
			"gen-address": true,
			"minimum": "0.00010000"
		  },
		  {
			"method": "Bitcoin Lightning",
			"limit": false,
			"fee": "0.00000000",
			"minimum": "0.00010000"
		  }
		]
	}`
	expectedCount := 2
	expectedItem1Method := "Bitcoin Lightning"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetDepositMethods(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[1].Method)
	require.Equal(suite.T(), expectedItem1Method, resp.Result[1].Method)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getDepositMethodsPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, record.Request.Form.Get("asset"))
}

// Test GetDepositAddresses when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetDepositAddresses() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := funding.GetDepositAddressesRequestParameters{
		Asset:  "XXBT",
		Method: "Bitcoin Lightning",
	}

	// Expected options
	options := &funding.GetDepositAddressesRequestOptions{
		New:    true,
		Amount: "0.1",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": [
		  {
			"address": "2N9fRkx5JTWXWHmXzZtvhQsufvoYRMq9ExV",
			"expiretm": "0",
			"new": true
		  },
		  {
			"address": "2NCpXUCEYr8ur9WXM1tAjZSem2w3aQeTcAo",
			"expiretm": "0",
			"new": true
		  },
		  {
			"address": "2Myd4eaAW96ojk38A2uDK4FbioCayvkEgVq",
			"expiretm": "0"
		  },
		  {
			"address": "rLHzPsX3oXdzU2qP17kHCH2G4csZv1rAJh",
			"expiretm": "0",
			"new": true,
			"tag": "1361101127"
		  },
		  {
			"address": "krakenkraken",
			"expiretm": "0",
			"memo": "4150096490"
		  }
		]
	}`
	expectedCount := 5
	expectedItem1Address := "2NCpXUCEYr8ur9WXM1tAjZSem2w3aQeTcAo"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetDepositAddresses(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[1].Address)
	require.Equal(suite.T(), expectedItem1Address, resp.Result[1].Address)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getDepositAddressesPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, record.Request.Form.Get("asset"))
	require.Equal(suite.T(), params.Method, record.Request.Form.Get("method"))
	require.Equal(suite.T(), options.Amount, record.Request.Form.Get("amount"))
	require.Equal(suite.T(), strconv.FormatBool(options.New), record.Request.Form.Get("new"))
}

// Test GetStatusOfRecentDeposits when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetStatusOfRecentDeposits() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &funding.GetStatusOfRecentDepositsRequestOptions{
		Asset:  "XXBT",
		Method: "Bitcoin Lightning",
		Start:  "42",
		End:    "42",
		Cursor: "false", // Set to false to ensure it will be forced to true (forced pagination)
		Limit:  10,
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
			"deposit": [
				{
					"method": "Bitcoin",
					"aclass": "currency",
					"asset": "XXBT",
					"refid": "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg",
					"txid": "6544b41b607d8b2512baf801755a3a87b6890eacdb451be8a94059fb11f0a8d9",
					"info": "2Myd4eaAW96ojk38A2uDK4FbioCayvkEgVq",
					"amount": "0.78125000",
					"fee": "0.0000000000",
					"time": 1688992722,
					"status": "Success",
					"status-prop": "return"
				},
				{
					"method": "Ether (Hex)",
					"aclass": "currency",
					"asset": "XETH",
					"refid": "FTQcuak-V6Za8qrPnhsTx47yYLz8Tg",
					"txid": "0x339c505eba389bf2c6bebb982cc30c6d82d0bd6a37521fa292890b6b180affc0",
					"info": "0xca210f4121dc891c9154026c3ae3d1832a005048",
					"amount": "0.1383862742",
					"time": 1688992722,
					"status": "Settled",
					"status-prop": "onhold",
					"originators": [
						"0x70b6343b104785574db2c1474b3acb3937ab5de7346a5b857a78ee26954e0e2d",
						"0x5b32f6f792904a446226b17f607850d0f2f7533cdc35845bfe432b5b99f55b66"
					]
				}
			]
		}
	}`
	expectedCount := 2
	expectedItem0TxId := "6544b41b607d8b2512baf801755a3a87b6890eacdb451be8a94059fb11f0a8d9"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetStatusOfRecentDeposits(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result.Deposits, expectedCount)
	require.NotNil(suite.T(), resp.Result.Deposits[0])
	require.Equal(suite.T(), expectedItem0TxId, resp.Result.Deposits[0].TransactionID)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getStatusOfRecentDepositsPath)
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
	require.Equal(suite.T(), options.Method, record.Request.Form.Get("method"))
	require.Equal(suite.T(), options.Start, record.Request.Form.Get("start"))
	require.Equal(suite.T(), options.End, record.Request.Form.Get("end"))
	require.Equal(suite.T(), strconv.FormatBool(true), record.Request.Form.Get("cursor"))
	require.Equal(suite.T(), strconv.FormatInt(options.Limit, 10), record.Request.Form.Get("limit"))
}

// Test GetWithdrawalMethods when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetWithdrawalMethods() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &funding.GetWithdrawalMethodsRequestOptions{
		Asset:   "XXBT",
		Network: "Bitcoin",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": [
		  {
			"asset": "XXBT",
			"method": "Bitcoin",
			"network": "Bitcoin",
			"minimum": "0.0004"
		  },
		  {
			"asset": "XXBT",
			"method": "Bitcoin Lightning",
			"network": "Lightning",
			"minimum": "0.00001"
		  }
		]
	}`
	expectedCount := 2
	expectedItem0Asset := "XXBT"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetWithdrawalMethods(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[0])
	require.Equal(suite.T(), expectedItem0Asset, resp.Result[0].Asset)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getWithdrawalMethodsPath)
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
	require.Equal(suite.T(), options.Network, record.Request.Form.Get("network"))
}

// Test GetWithdrawalAddresses when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetWithdrawalAddresses() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &funding.GetWithdrawalAddressesRequestOptions{
		Asset:    "XXBT",
		Method:   "Bitcoin",
		Key:      "btc-wallet-1",
		Verified: true,
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": [
		  {
			"address": "bc1qxdsh4sdd29h6ldehz0se5c61asq8cgwyjf2y3z",
			"asset": "XBT",
			"method": "Bitcoin",
			"key": "btc-wallet-1",
			"verified": true
		  }
		]
	}`
	expectedCount := 1
	expectedItem0Address := "bc1qxdsh4sdd29h6ldehz0se5c61asq8cgwyjf2y3z"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetWithdrawalAddresses(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[0])
	require.Equal(suite.T(), expectedItem0Address, resp.Result[0].Address)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getWithdrawalAddressesPath)
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
	require.Equal(suite.T(), options.Method, record.Request.Form.Get("method"))
	require.Equal(suite.T(), options.Key, record.Request.Form.Get("key"))
	require.Equal(suite.T(), strconv.FormatBool(options.Verified), record.Request.Form.Get("verified"))
}

// Test GetWithdrawalInformation when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetWithdrawalInformation() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := funding.GetWithdrawalInformationRequestParameters{
		Asset:  "XXBT",
		Key:    "btc-wallet-1",
		Amount: "0.1",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "method": "Bitcoin",
		  "limit": "332.00956139",
		  "amount": "0.72480000",
		  "fee": "0.00020000"
		}
	}`
	expectedLimit := "332.00956139"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetWithdrawalInformation(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedLimit, resp.Result.Limit)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getWithdrawalInformationPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, record.Request.Form.Get("asset"))
	require.Equal(suite.T(), params.Key, record.Request.Form.Get("key"))
	require.Equal(suite.T(), params.Amount, record.Request.Form.Get("amount"))
}

// Test WithdrawFunds when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestWithdrawFunds() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := funding.WithdrawFundsRequestParameters{
		Asset:  "XXBT",
		Key:    "btc-wallet-1",
		Amount: "0.1",
	}

	// Expected options
	options := &funding.WithdrawFundsRequestOptions{
		Address: "test",
		MaxFee:  "0.005",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "refid": "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"
		}
	  }`
	expectedRefId := "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.WithdrawFunds(context.Background(), expectedNonce, params, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedRefId, resp.Result.ReferenceID)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, withdrawFundsPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, record.Request.Form.Get("asset"))
	require.Equal(suite.T(), params.Key, record.Request.Form.Get("key"))
	require.Equal(suite.T(), params.Amount, record.Request.Form.Get("amount"))
	require.Equal(suite.T(), options.Address, record.Request.Form.Get("address"))
	require.Equal(suite.T(), options.MaxFee, record.Request.Form.Get("max_fee"))
}

// Test GetStatusOfRecentWithdrawals when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetStatusOfRecentWithdrawals() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &funding.GetStatusOfRecentWithdrawalsRequestOptions{
		Method: "Bitcoin",
		Asset:  "XXBT",
		Start:  "42",
		End:    "42",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": [
		  {
			"method": "Bitcoin",
			"aclass": "currency",
			"asset": "XXBT",
			"refid": "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg",
			"txid": "THVRQM-33VKH-UCI7BS",
			"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
			"amount": "0.72485000",
			"fee": "0.00020000",
			"time": 1688014586,
			"status": "Pending",
			"key": "btc-wallet-1"
		  },
		  {
			"method": "Bitcoin",
			"aclass": "currency",
			"asset": "XXBT",
			"refid": "FTQcuak-V6Za8qrPnhsTx47yYLz8Tg",
			"txid": "KLETXZ-33VKH-UCI7BS",
			"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
			"amount": "0.72485000",
			"fee": "0.00020000",
			"time": 1688015423,
			"status": "Failure",
			"status-prop": "canceled",
			"key": "btc-wallet-2"
		  }
		]
	}`
	expectedCount := 2
	expectedItem0RefId := "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetStatusOfRecentWithdrawals(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.Len(suite.T(), resp.Result, expectedCount)
	require.NotNil(suite.T(), resp.Result[0])
	require.Equal(suite.T(), expectedItem0RefId, resp.Result[0].ReferenceID)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getStatusOfRecentWithdrawalsPath)
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
	require.Equal(suite.T(), options.Method, record.Request.Form.Get("method"))
	require.Equal(suite.T(), options.Start, record.Request.Form.Get("start"))
	require.Equal(suite.T(), options.End, record.Request.Form.Get("end"))
}

// Test RequestWithdrawalCancelation when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestRequestWithdrawalCancelation() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := funding.RequestWithdrawalCancellationRequestParameters{
		Asset:       "XXBT",
		ReferenceId: "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": true
	}`

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.RequestWithdrawalCancellation(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.True(suite.T(), resp.Result)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, requestWithdrawalCancellationPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, record.Request.Form.Get("asset"))
	require.Equal(suite.T(), params.ReferenceId, record.Request.Form.Get("refid"))
}

// Test RequestWalletTransfer when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestRequestWalletTransfer() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := funding.RequestWalletTransferRequestParameters{
		Asset:  "XXBT",
		From:   string(funding.Spot),
		To:     string(funding.Futures),
		Amount: "1.2",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "refid": "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"
		}
	}`
	expectedRefId := "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.RequestWalletTransfer(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedRefId, resp.Result.ReferenceID)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, requestWalletTransferPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Asset, record.Request.Form.Get("asset"))
	require.Equal(suite.T(), params.From, record.Request.Form.Get("from"))
	require.Equal(suite.T(), params.To, record.Request.Form.Get("to"))
	require.Equal(suite.T(), params.Amount, record.Request.Form.Get("amount"))
}

/*************************************************************************************************/
/* UNIT TESTS - EARN                                                                             */
/*************************************************************************************************/

// Test AllocateEarnFunds when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestAllocateEarnFunds() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := earn.AllocateFundsRequestParameters{
		Amount:     "1.2",
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": true
	}`

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.AllocateEarnFunds(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.True(suite.T(), resp.Result)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, allocateEarnFundsPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Amount, record.Request.Form.Get("amount"))
	require.Equal(suite.T(), params.StrategyId, record.Request.Form.Get("strategy_id"))
}

// Test DeallocateEarnFunds when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestDeallocateEarnFunds() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := earn.DeallocateFundsRequestParameters{
		Amount:     "1.2",
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": true
	}`

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.DeallocateEarnFunds(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.True(suite.T(), resp.Result)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, deallocateEarnFundsPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.Amount, record.Request.Form.Get("amount"))
	require.Equal(suite.T(), params.StrategyId, record.Request.Form.Get("strategy_id"))
}

// Test GetAllocationStatus when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetAllocationStatus() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := earn.GetAllocationStatusRequestParameters{
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "pending": true
		}
	}`

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetAllocationStatus(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.True(suite.T(), resp.Result.Pending)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getAllocationStatusPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.StrategyId, record.Request.Form.Get("strategy_id"))
}

// Test GetDeallocationStatus when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetDeallocationStatus() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected params
	params := earn.GetDeallocationStatusRequestParameters{
		StrategyId: "ESRFUO3-Q62XD-WIOIL7",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "pending": true
		}
	}`

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetDeallocationStatus(context.Background(), expectedNonce, params, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.True(suite.T(), resp.Result.Pending)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getDeallocationStatusPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), params.StrategyId, record.Request.Form.Get("strategy_id"))
}

// Test ListEarnStrategies when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestListEarnStrategies() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &earn.ListEarnStrategiesRequestOptions{
		Ascending: true,
		Asset:     "XXBT",
		Cursor:    "false", // Set to false and verify if true -> pagination use is forced
		Limit:     10,
		LockType:  string(earn.Bonded),
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "next_cursor": "2",
		  "items": [
			{
			  "id": "ESRFUO3-Q62XD-WIOIL7",
			  "asset": "DOT",
			  "lock_type": {
				"type": "instant",
				"payout_frequency": 604800
			  },
			  "apr_estimate": {
				"low": "8.0000",
				"high": "12.0000"
			  },
			  "user_min_allocation": "0.01",
			  "allocation_fee": "0.0000",
			  "deallocation_fee": "0.0000",
			  "auto_compound": {
				"type": "enabled"
			  },
			  "yield_source": {
				"type": "staking"
			  },
			  "can_allocate": true,
			  "can_deallocate": true,
			  "allocation_restriction_info": []
			}
		  ]
		}
	}`
	expectedCount := 1
	expectedNextCursor := "2"
	expectedItem0Id := "ESRFUO3-Q62XD-WIOIL7"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.ListEarnStrategies(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Len(suite.T(), resp.Result.Items, expectedCount)
	require.Equal(suite.T(), expectedNextCursor, resp.Result.NextCursor)
	require.NotNil(suite.T(), resp.Result.Items[0])
	require.Equal(suite.T(), expectedItem0Id, resp.Result.Items[0].Id)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, listEarnStartegiesPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Ascending), record.Request.Form.Get("ascending"))
	require.Equal(suite.T(), options.Asset, record.Request.Form.Get("asset"))
	require.Equal(suite.T(), strconv.FormatBool(true), record.Request.Form.Get("cursor"))
	require.Equal(suite.T(), strconv.FormatInt(int64(options.Limit), 10), record.Request.Form.Get("limit"))
	require.Equal(suite.T(), options.LockType, record.Request.Form.Get("lock_type"))
}

// Test ListEarnAllocations when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestListEarnAllocations() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Expected options
	options := &earn.ListEarnAllocationsRequestOptions{
		Ascending:           true,
		ConvertedAsset:      "ZUSD",
		HideZeroAllocations: true,
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "converted_asset": "USD",
		  "total_allocated": "49.2398",
		  "total_rewarded": "0.0675",
		  "next_cursor": "2",
		  "items": [
			{
			  "strategy_id": "ESDQCOL-WTZEU-NU55QF",
			  "native_asset": "ETH",
			  "amount_allocated": {
				"bonding": {
				  "native": "0.0210000000",
				  "converted": "39.0645",
				  "allocation_count": 2,
				  "allocations": [
					{
					  "created_at": "2023-07-06T10:52:05Z",
					  "expires": "2023-08-19T02:34:05.807Z",
					  "native": "0.0010000000",
					  "converted": "1.8602"
					},
					{
					  "created_at": "2023-08-01T11:25:52Z",
					  "expires": "2023-09-06T07:55:52.648Z",
					  "native": "0.0200000000",
					  "converted": "37.2043"
					}
				  ]
				},
				"total": {
				  "native": "0.0210000000",
				  "converted": "39.0645"
				}
			  },
			  "total_rewarded": {
				"native": "0",
				"converted": "0.0000"
			  }
			}
		  ]
		}
	}`
	expectedCount := 1
	expectedItem0AllocatedTotalNative := "0.0210000000"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.ListEarnAllocations(context.Background(), expectedNonce, options, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Len(suite.T(), resp.Result.Items, expectedCount)
	require.NotNil(suite.T(), resp.Result.Items[0])
	require.Equal(suite.T(), expectedItem0AllocatedTotalNative, resp.Result.Items[0].AmountAllocated.Total.Native)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, listEarnAllocationsPath)
	require.Equal(suite.T(), http.MethodPost, record.Request.Method)
	require.Equal(suite.T(), suite.client.agent, record.Request.UserAgent())
	require.Equal(suite.T(), "application/x-www-form-urlencoded", record.Request.Header.Get("Content-Type"))
	require.NotEmpty(suite.T(), record.Request.Header.Get("Api-Sign"))     // Headers are in canonical form in recorded request
	require.Equal(suite.T(), apiKey, record.Request.Header.Get("Api-Key")) // Headers are in canonical form in recorded request

	// Check request form body
	require.NoError(suite.T(), record.Request.ParseForm())
	require.Equal(suite.T(), strconv.FormatInt(expectedNonce, 10), record.Request.Form.Get("nonce"))
	require.Equal(suite.T(), expectedSecOpts.SecondFactor, record.Request.Form.Get("otp"))
	require.Equal(suite.T(), strconv.FormatBool(options.Ascending), record.Request.Form.Get("ascending"))
	require.Equal(suite.T(), strconv.FormatBool(options.HideZeroAllocations), record.Request.Form.Get("hide_zero_allocations"))
	require.Equal(suite.T(), options.ConvertedAsset, record.Request.Form.Get("converted_asset"))
}

/*************************************************************************************************/
/* UNIT TESTS - WEBSOCKET                                                                        */
/*************************************************************************************************/

// Test GetWebsocketsToken when a valid response is received from the test server.
//
// Test will ensure:
//   - The request is well formatted and contains all inputs.
//   - The returned values contain the expected parsed response data.
func (suite *KrakenSpotRESTClientTestSuite) TestGetWebsocketsToken() {

	// Expected nonce and secopts
	expectedNonce := int64(42)
	expectedSecOpts := &common.SecurityOptions{
		SecondFactor: "42",
	}

	// Predefined response
	expectedJSONResponse := `
	{
		"error": [],
		"result": {
		  "token": "1Dwc4lzSwNWOAwkMdqhssNNFhs1ed606d1WcF3XfEMw",
		  "expires": 900
		}
	}`
	expectedExpire := int64(900)
	expectedToken := "1Dwc4lzSwNWOAwkMdqhssNNFhs1ed606d1WcF3XfEMw"

	// Configure test server
	suite.srv.PushPredefinedServerResponse(&gosette.PredefinedServerResponse{
		Status:  http.StatusOK,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
		Body:    []byte(expectedJSONResponse),
	})

	// Make request
	resp, httpresp, err := suite.client.GetWebsocketToken(context.Background(), expectedNonce, expectedSecOpts)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), httpresp)
	require.NotNil(suite.T(), resp)

	// Check parsed response
	require.NotNil(suite.T(), resp.Result)
	require.Equal(suite.T(), expectedExpire, resp.Result.Expires)
	require.Equal(suite.T(), expectedToken, resp.Result.Token)

	// Get the recorded request
	record := suite.srv.PopServerRecord()
	require.NotNil(suite.T(), record)

	// Check the request settings
	require.Contains(suite.T(), record.Request.URL.Path, getWebsocketTokenPath)
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
