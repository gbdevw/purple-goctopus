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

// Unit test suite for GetWithdrawalInformation DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetWithdrawalInformationTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetWithdrawalInformationTestSuite(t *testing.T) {
	suite.Run(t, new(GetWithdrawalInformationTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetWithdrawalInformationResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetWithdrawalInformationResponse struct.
func (suite *GetWithdrawalInformationTestSuite) TestGetWithdrawalInformationResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "method": "Bitcoin",
		  "limit": "332.00956139",
		  "amount": "0.72485000",
		  "fee": "0.00020000"
		}
	}`
	expectedItem1Method := "Bitcoin"
	// Unmarshal payload into struct
	response := new(GetWithdrawalInformationResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedItem1Method, response.Result.Method)
}
