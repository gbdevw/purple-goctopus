package messages

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* UNIT TEST SUITE                                                                               */
/*************************************************************************************************/

// Unit test suite for CancelOrder
type CancelOrderUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestCancelOrderUnitTestSuite(t *testing.T) {
	suite.Run(t, new(CancelOrderUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example CancelOrderRequest message to the same payload as documentation
func (suite *CancelOrderUnitTestSuite) TestCancelOrderRequestMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "cancelOrder",
		"token": "0000000000000000000000000000000000000000",
		"txid": [
		  "OGTT3Y-C6I3P-XRI6HX",
		  "OGTT3Y-C6I3P-X2I6HX"
		]
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(CancelOrderRequest)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example of a successfull CancelOrderResponse and then test marshalling it to get the same
// payload as the API.
func (suite *CancelOrderUnitTestSuite) TestCancelOrderResponseMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "cancelOrderStatus",
		"status": "error",
		"errorMessage": "EOrder:Unknown order"
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(CancelOrderResponse)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}
