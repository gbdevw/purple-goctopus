package funding

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetStatusOfRecentWithdrawals DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetStatusOfRecentWithdrawalsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetStatusOfRecentWithdrawalsTestSuite(t *testing.T) {
	suite.Run(t, new(GetStatusOfRecentWithdrawalsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetStatusOfRecentWithdrawalsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetStatusOfRecentWithdrawalsResponse struct.
func (suite *GetStatusOfRecentWithdrawalsTestSuite) TestGetStatusOfRecentWithdrawalsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": [
		  {
			"method": "Bitcoin",
			"aclass": "currency",
			"asset": "XXBT",
			"refid": "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg",
			"txid": "THVRQM-33VKH-UCI7BS",
			"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
			"amount": "0.72485000",
			"fee": "0.00020000",
			"time": 1688014586,
			"status": "Pending"
		  },
		  {
			"method": "Bitcoin",
			"aclass": "currency",
			"asset": "XXBT",
			"refid": "FTQcuak-V6Za8qrPnhsTx47yYLz8Tg",
			"txid": "KLETXZ-33VKH-UCI7BS",
			"info": "mzp6yUVMRxfasyfwzTZjjy38dHqMX7Z3GR",
			"amount": "0.72485000",
			"fee": "0.00020000",
			"time": 1688015423,
			"status": "Failure",
			"status-prop": "canceled"
		  }
		]
	}`
	expectedCount := 2
	expectedItem1Method := "Bitcoin"
	// Unmarshal payload into struct
	response := new(GetStatusOfRecentWithdrawalsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.Len(suite.T(), response.Result, expectedCount)
	require.Equal(suite.T(), expectedItem1Method, response.Result[0].Method)
}
