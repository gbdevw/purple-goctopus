package rest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for KrakenSpotRESTClientAuthorizer.
//
// The test suite ensures the authorizer produces valid signatures for the requests to authorize.
type KrakenSpotRESTClientAuthorizerTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestKrakenSpotRESTClientAuthorizerTestSuite(t *testing.T) {
	suite.Run(t, new(KrakenSpotRESTClientAuthorizerTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test interface compliance
func (suite *KrakenSpotRESTClientAuthorizerTestSuite) TestIFaceCompliance() {
	// Inputs
	inputB64Secret := "kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg=="
	// Configure authorizer (no need API key to generate the signature) and assign it to an interface{} variable.
	var auth interface{}
	var err error
	auth, err = NewKrakenSpotRESTClientAuthorizer("", inputB64Secret)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), auth)
	// Cast interface{} to the interface type and ensure it is OK
	_, ok := auth.(KrakenSpotRESTClientAuthorizerIface)
	require.True(suite.T(), ok)
}

// Test the getKrakenSignature method.
//
// Test will ensure the method produces the expected signature given the example inputs from the doocumentation.
func (suite *KrakenSpotRESTClientAuthorizerTestSuite) TestGetKrakenSignature() {
	// Test settings, expectations, ...
	inputUriPath := "/0/private/AddOrder"
	inputForm := url.Values{
		"nonce":     []string{"1616492376594"},
		"ordertype": []string{"limit"},
		"pair":      []string{"XBTUSD"},
		"price":     []string{"37500"},
		"type":      []string{"buy"},
		"volume":    []string{"1.25"},
	}
	inputB64Secret := "kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg=="
	expectedEncodedPayload := "nonce=1616492376594&ordertype=limit&pair=XBTUSD&price=37500&type=buy&volume=1.25"
	expectedSignature := "4/dpxb3iT4tp/ZCVEwSnEsLxx0bqyhLpdfOpc6fn7OR8+UClSV5n9E6aSS8MPtnRfp32bAb0nmbRn6H8ndwLUQ=="
	// Check encoded payload
	require.Equal(suite.T(), expectedEncodedPayload, inputForm.Encode())
	// Configure authorizer (no need API key to generate the signature)
	auth, err := NewKrakenSpotRESTClientAuthorizer("", inputB64Secret)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), auth)
	// Generate signature
	signature, err := auth.getKrakenSignature(inputUriPath, inputForm)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), expectedSignature, signature)
}

// Test the Authorize method.
//
// Test will ensure the method uses well the request data to produce the expected signature. Test will also
// ensure the method adds the signing data well to the provided request.
func (suite *KrakenSpotRESTClientAuthorizerTestSuite) TestAuthorize() {
	// Test settings, expectations, ...
	inputUriPath := "/0/private/AddOrder"
	inputB64Secret := "kQH5HW/8p1uGOVjbgWA7FunAmGO8lsSUXNsu3eow76sz84Q18fWxnyRzBHCd3pd5nE9qa99HAZtuZuj6F1huXg=="
	inputPayload := "nonce=1616492376594&ordertype=limit&pair=XBTUSD&price=37500&type=buy&volume=1.25"
	expectedKey := "KEY"
	expectedSignature := "4/dpxb3iT4tp/ZCVEwSnEsLxx0bqyhLpdfOpc6fn7OR8+UClSV5n9E6aSS8MPtnRfp32bAb0nmbRn6H8ndwLUQ=="
	// Forge input request * Be sure to use POST method + set the Content-Type - Hostname and scheme do not matter.
	ireq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", "http://localhost", inputUriPath), strings.NewReader(inputPayload))
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), ireq)
	// Set content type to
	ireq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Configure authorizer.
	auth, err := NewKrakenSpotRESTClientAuthorizer(expectedKey, inputB64Secret)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), auth)
	// Authorize request
	oreq, err := auth.Authorize(context.Background(), ireq)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), oreq)
	// Check the right headers contain the signature and key
	require.Equal(suite.T(), expectedKey, oreq.Header[managedHeaderAPIKey][0])
	require.Equal(suite.T(), expectedSignature, oreq.Header[managedHeaderAPISign][0])
}
