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

// Unit test suite for GetStatusOfRecentDeposits DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetStatusOfRecentDepositsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetStatusOfRecentDepositsTestSuite(t *testing.T) {
	suite.Run(t, new(GetStatusOfRecentDepositsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetStatusOfRecentDepositsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetStatusOfRecentDepositsResponse struct.
func (suite *GetStatusOfRecentDepositsTestSuite) TestGetStatusOfRecentDepositsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
			"next_cursor": "AAAA",
			"deposit": [
				{
					"method": "Bitcoin",
					"aclass": "currency",
					"asset": "XXBT",
					"refid": "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg",
					"txid": "6544b41b607d8b2512baf801755a3a87b6890eacdb451be8a94059fb11f0a8d9",
					"info": "2Myd4eaAW96ojk38A2uDK4FbioCayvkEgVq",
					"amount": "0.78125000",
					"fee": "0.0000000000",
					"time": 1688992722,
					"status": "Success",
					"status-prop": "return"
				},
				{
					"method": "Ether (Hex)",
					"aclass": "currency",
					"asset": "XETH",
					"refid": "FTQcuak-V6Za8qrPnhsTx47yYLz8Tg",
					"txid": "0x339c505eba389bf2c6bebb982cc30c6d82d0bd6a37521fa292890b6b180affc0",
					"info": "0xca210f4121dc891c9154026c3ae3d1832a005048",
					"amount": "0.1383862742",
					"time": 1688992722,
					"status": "Settled",
					"status-prop": "onhold",
					"originators": [
					"0x70b6343b104785574db2c1474b3acb3937ab5de7346a5b857a78ee26954e0e2d",
					"0x5b32f6f792904a446226b17f607850d0f2f7533cdc35845bfe432b5b99f55b66"
					]
				}
			]
		}
	}`
	expectedNextCursor := "AAAA"
	expectedCount := 2
	expectedItem1Method := "Bitcoin"
	expectedItem2OriginatorsCount := 2
	expectedItem2Originators1 := "0x70b6343b104785574db2c1474b3acb3937ab5de7346a5b857a78ee26954e0e2d"
	// Unmarshal payload into struct
	response := new(GetStatusOfRecentDepositsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.NotEmpty(suite.T(), response.Result.Deposits)
	require.Equal(suite.T(), expectedNextCursor, response.Result.NextCursor)
	require.Len(suite.T(), response.Result.Deposits, expectedCount)
	require.Equal(suite.T(), expectedItem1Method, response.Result.Deposits[0].Method)
	require.Len(suite.T(), response.Result.Deposits[1].Originators, expectedItem2OriginatorsCount)
	require.Equal(suite.T(), expectedItem2Originators1, response.Result.Deposits[1].Originators[0])
}
