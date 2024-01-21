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

// Unit test suite for AddOrder
type AddOrderUnitTestSuite struct {
	suite.Suite
}

// Run the unit test suite
func TestAddOrderUnitTestSuite(t *testing.T) {
	suite.Run(t, new(AddOrderUnitTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test marshalling an example AddOrderRequest message to the same payload as documentation
func (suite *AddOrderUnitTestSuite) TestAddOrderRequestMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "addOrder",
		"token": "0000000000000000000000000000000000000000",
		"ordertype": "limit",
		"type": "buy",
		"pair": "XBT/USD",
		"price": "9000",
		"volume": "10",
		"close[ordertype]": "limit",
		"close[price]": "9100"
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Unmarshal to target
	target := new(AddOrderRequest)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example of a successfull AddOrderResponse and then test marshalling it to get the same
// payload as the API.
func (suite *AddOrderUnitTestSuite) TestAddOrderResponseMarshalJson() {
	// Payload to marshal
	payload := `{
		"event": "addOrderStatus",
		"status": "ok",
		"txid": "ONPNXH-KMKMU-F4MR5V",
		"descr": "buy 0.01770000 XBTUSD @ limit 4000"
	  }`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Expectations
	expectedEvent := string(EventTypeAddOrderStatus)
	expectedStatus := string(Ok)
	expectedTxId := "ONPNXH-KMKMU-F4MR5V"
	// Remove whitespaces for test
	expectedOrderDescr := matchesWhitespacesRegex.ReplaceAllString("buy 0.01770000 XBTUSD @ limit 4000", "")
	// Unmarshal to target
	target := new(AddOrderResponse)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedTxId, target.TxId)
	require.Equal(suite.T(), expectedStatus, target.Status)
	require.Equal(suite.T(), expectedOrderDescr, target.Description)
	require.Empty(suite.T(), target.Err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}

// Test unmarshalling an example of a AddOrderResponse with an error message and then test marshalling it to get the same
// payload as the API.
func (suite *AddOrderUnitTestSuite) TestAddOrderResponseMarshalJsonWithError() {
	// Payload to marshal
	payload := `{
		"event": "addOrderStatus",
		"status": "error",
		"errorMessage": "EOrder:Order minimum not met"
	}`
	// Remove whitespaces from payload
	payload = matchesWhitespacesRegex.ReplaceAllString(payload, "")
	// Expectations
	expectedEvent := string(EventTypeAddOrderStatus)
	expectedStatus := string(Err)
	expectedErr := matchesWhitespacesRegex.ReplaceAllString("EOrder:Order minimum not met", "")
	// Unmarshal to target
	target := new(AddOrderResponse)
	err := json.Unmarshal([]byte(payload), target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), expectedEvent, target.Event)
	require.Equal(suite.T(), expectedStatus, target.Status)
	require.Equal(suite.T(), expectedErr, target.Err)
	// Marshal target
	actual, err := json.Marshal(target)
	require.NoError(suite.T(), err)
	// Check data
	require.Equal(suite.T(), payload, string(actual))
}
