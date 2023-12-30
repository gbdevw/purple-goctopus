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

// Unit test suite for AddOrderBatch DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type AddOrderBatchTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestAddOrderBatchTestSuite(t *testing.T) {
	suite.Run(t, new(AddOrderBatchTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of AddOrderBatch.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding AddOrderBatchResponse struct.
func (suite *AddOrderBatchTestSuite) TestAddOrderBatchUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "orders": [
			{
			  "txid": "65LRD3-AHGRA-YAH8E5",
			  "descr": {
				"order": "buy 1.02010000 XBTUSD @ limit 29000.0"
			  }
			},
			{
			  "txid": "OK8HFF-5J2PL-XLR17S",
			  "descr": {
				"order": "sell 0.14000000 XBTUSD @ limit 40000.0"
			  }
			}
		  ]
		}
	}`
	expectedCount := 2
	expected2DescrOrder := "sell 0.14000000 XBTUSD @ limit 40000.0"
	expected2TxId := "OK8HFF-5J2PL-XLR17S"
	// Unmarshal payload into struct
	response := new(AddOrderBatchResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Orders, expectedCount)
	require.Equal(suite.T(), expected2TxId, response.Result.Orders[1].Id)
	require.Equal(suite.T(), expected2DescrOrder, response.Result.Orders[1].Description.Order)
	require.Empty(suite.T(), response.Result.Orders[1].Error)
}
