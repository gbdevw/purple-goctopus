package trading

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for EditOrder DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type EditOrderTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestEditOrderTestSuite(t *testing.T) {
	suite.Run(t, new(EditOrderTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of EditOrder.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding EditOrderResponse struct.
func (suite *EditOrderTestSuite) TestEditOrderUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
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
	expectedStatus := string(Ok)
	expectedTxId := "OFVXHJ-KPQ3B-VS7ELA"
	expectedOriginalTxId := "OHYO67-6LP66-HMQ437"
	expectedVolume := "0.00030000"
	expectedPrice := "19500.0"
	expectedPrice2 := "32500.0"
	expectedCancelCount := 1
	expectedDescrOrder := "buy 0.00030000 XXBTZGBP @ limit 19500.0"
	// Unmarshal payload into struct
	response := new(EditOrderResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedStatus, response.Result.Status)
	require.Equal(suite.T(), expectedTxId, response.Result.TransactionID)
	require.Equal(suite.T(), expectedOriginalTxId, response.Result.OriginalTransactionID)
	require.Equal(suite.T(), expectedVolume, response.Result.Volume)
	require.Equal(suite.T(), expectedPrice, response.Result.Price)
	require.Equal(suite.T(), expectedPrice2, response.Result.Price2)
	require.Equal(suite.T(), expectedCancelCount, response.Result.OrdersCancelled)
	require.NotNil(suite.T(), response.Result.Description)
	require.Equal(suite.T(), expectedDescrOrder, response.Result.Description.Order)
	require.Nil(suite.T(), response.Result.NewUserReference)
	require.Nil(suite.T(), response.Result.OldUserReference)
}
