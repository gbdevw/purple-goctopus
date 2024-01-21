package rest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for KrakenSpotRESTClientAuthorizerInstrumentationDecorator.
type KrakenSpotRESTClientAuthorizerInstrumentationDecoratorTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestKrakenSpotRESTClientAuthorizerInstrumentationDecoratorTestSuite(t *testing.T) {
	suite.Run(t, new(KrakenSpotRESTClientAuthorizerInstrumentationDecoratorTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test interface compliance.
func (suite *KrakenSpotRESTClientAuthorizerInstrumentationDecoratorTestSuite) TestIFaceCompliance() {
	// Configure decorate and assign it to interface{}
	var instance interface{} = InstrumentKrakenSpotRESTClientAuthorizer(NewMockKrakenSpotRESTClientAuthorizer(), nil)
	// Cast interface{} to the interface type and ensure it is OK
	_, ok := instance.(KrakenSpotRESTClientAuthorizerIface)
	require.True(suite.T(), ok)
}

// Test panic when no decorated is provided.
func (suite *KrakenSpotRESTClientAuthorizerInstrumentationDecoratorTestSuite) TestFactoryValidation() {
	require.Panics(suite.T(), func() {
		InstrumentKrakenSpotRESTClientAuthorizer(nil, nil)
	})
}

// Test the Authorize method when decorated returns a request and no error.
//
// Test will ensure mock works as expected.
func (suite *KrakenSpotRESTClientAuthorizerInstrumentationDecoratorTestSuite) TestAuthorize() {
	// Configure mock
	m := NewMockKrakenSpotRESTClientAuthorizer()
	// Set mock expectations and returned values
	ireq, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	require.NoError(suite.T(), err)
	m.On("Authorize", mock.Anything, mock.Anything).Return(ireq, err)
	// Decorate mock
	dec := InstrumentKrakenSpotRESTClientAuthorizer(m, nil)
	// Call decorator method and check results
	req, err := dec.Authorize(context.Background(), ireq)
	require.NotNil(suite.T(), req)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), ireq.Method, req.Method)
	// Verify mock has been called
	m.AssertNumberOfCalls(suite.T(), "Authorize", 1)
}

// Test the Authorize method when decorated returns nil and an error.
//
// Test will ensure mock works as expected.
func (suite *KrakenSpotRESTClientAuthorizerInstrumentationDecoratorTestSuite) TestAuthorizeWithError() {
	// Configure mock
	m := NewMockKrakenSpotRESTClientAuthorizer()
	// Set mock expectations and returned values
	ireq, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	require.NoError(suite.T(), err)
	m.On("Authorize", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("fail"))
	// Decorate mock
	dec := InstrumentKrakenSpotRESTClientAuthorizer(m, nil)
	// Call decorator method and check results
	req, err := dec.Authorize(context.Background(), ireq)
	require.Nil(suite.T(), req)
	require.Error(suite.T(), err)
	// Verify mock has been called
	m.AssertNumberOfCalls(suite.T(), "Authorize", 1)
}
