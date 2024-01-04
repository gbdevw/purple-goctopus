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

// Unit test suite for CancelAllOrders
type CancelAllOrdersUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestCancelAllOrdersUnitTestSuite(t *testing.T) {
	suite.Run(t, new(CancelAllOrdersUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example CancelAllOrdersRequest message to the same payload as documentation
func (suite *CancelAllOrdersUnitTestSuite) TestCancelAllOrdersRequestMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "cancelAll",
		"token": "0000000000000000000000000000000000000000"
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(CancelAllOrdersRequest)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example of a successfull CancelAllOrdersResponse and then test marshalling it to get the same
// payload as the API.
func (suite *CancelAllOrdersUnitTestSuite) TestCancelAllOrdersResponseMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "cancelAllStatus",
		"count": 2,
		"status": "ok"
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(CancelAllOrdersResponse)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}
