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

// Unit test suite for AddOrder DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type AddOrderTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestAddOrderTestSuite(t *testing.T) {
	suite.Run(t, new(AddOrderTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of AddOrder.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding AddOrderResponse struct.
func (suite *AddOrderTestSuite) TestAddOrderUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
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
	expectedDescrOrder := "buy 1.25000000 XBTUSD @ limit 27500.0"
	expectedTxs := []string{"OU22CG-KLAF2-FWUDD7"}
	// Unmarshal payload into struct
	response := new(AddOrderResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedDescrOrder, response.Result.Description.Order)
	require.Empty(suite.T(), response.Result.Description.Close)
	require.ElementsMatch(suite.T(), response.Result.TransactionIDs, expectedTxs)
}
