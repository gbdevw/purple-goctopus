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

// Unit test suite for EditOrder
type EditOrderUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestEditOrderUnitTestSuite(t *testing.T) {
	suite.Run(t, new(EditOrderUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example EditOrderRequest message to the same payload as documentation
func (suite *EditOrderUnitTestSuite) TestEditOrderRequestMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "editOrder",
		"token": "0000000000000000000000000000000000000000",
		"orderid": "O26VH7-COEPR-YFYXLK",
		"reqid": 3,
		"pair": "XBT/USD",
		"price": "900",
		"newuserref": "666"
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(EditOrderRequest)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example of a successfull EditOrderResponse and then test marshalling it to get the same
// payload as the API.
func (suite *EditOrderUnitTestSuite) TestEditOrderResponseMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "editOrderStatus",
		"txid": "OTI672-HJFAO-XOIPPK",
		"originaltxid": "O65KZW-J4AW3-VFS74A",
		"reqid": 3,
		"status": "ok",
		"descr": "order edited price = 9000.00000000"
	  }`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(EditOrderResponse)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}
