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

// Unit test suite for RequestWalletTransfer DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type RequestWalletTransferTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestRequestWalletTransferTestSuite(t *testing.T) {
	suite.Run(t, new(RequestWalletTransferTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of RequestWalletTransferResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding RequestWalletTransferResponse struct.
func (suite *RequestWalletTransferTestSuite) TestRequestWalletTransferResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "refid": "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"
		}
	}`
	expectedRefId := "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"
	// Unmarshal payload into struct
	response := new(RequestWalletTransferResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedRefId, response.Result.ReferenceID)
}
