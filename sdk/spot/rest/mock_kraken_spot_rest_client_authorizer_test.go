package rest

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for MockKrakenSpotRESTClientAuthorizer.
type MockKrakenSpotRESTClientAuthorizerTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestMockKrakenSpotRESTClientAuthorizerTestSuite(t *testing.T) {
	suite.Run(t, new(MockKrakenSpotRESTClientAuthorizerTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test interface compliance.
func (suite *MockKrakenSpotRESTClientAuthorizerTestSuite) TestIFaceCompliance() {
	// Configure mock and assign it to interface{}
	var instance interface{} = NewMockKrakenSpotRESTClientAuthorizer()
	// Cast interface{} to the interface type and ensure it is OK
	_, ok := instance.(KrakenSpotRESTClientAuthorizerIface)
	require.True(suite.T(), ok)
}

// Test the Authorize method.
//
// Test will ensure mock works as expected.
func (suite *MockKrakenSpotRESTClientAuthorizerTestSuite) TestAuthorize() {
	// Configure mock
	m := NewMockKrakenSpotRESTClientAuthorizer()
	// Set mock expectations and returned values
	m.On("Authorize", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("fail"))
	// Call mocked method
	req, err := m.Authorize(context.Background(), nil)
	require.Nil(suite.T(), req)
	require.Error(suite.T(), err)
	m.AssertNumberOfCalls(suite.T(), "Authorize", 1)
}
