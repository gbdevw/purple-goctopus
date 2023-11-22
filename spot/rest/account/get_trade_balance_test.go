package account

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetTradeBalance DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetTradeBalanceTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetTradeBalanceTestSuite(t *testing.T) {
	suite.Run(t, new(GetTradeBalanceTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetTradeBalance.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetTradeBalanceResponse struct.
func (suite *GetTradeBalanceTestSuite) TestGetTradeBalanceUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "eb": "1101.3425",
		  "tb": "392.2264",
		  "m": "7.0354",
		  "n": "-10.0232",
		  "c": "21.1063",
		  "v": "31.1297",
		  "e": "382.2032",
		  "mf": "375.1678",
		  "ml": "5432.57",
		  "uv": "42.42"
		}
	}`
	expectedEquivalentBalance := "1101.3425"
	expectedTradeBalance := "392.2264"
	expectedUsedMargin := "7.0354"
	expectedUnrealizedPnl := "-10.0232"
	expectedPositionsCost := "21.1063"
	expectedPositionsValuation := "31.1297"
	expectedEquity := "382.2032"
	expectedFreeMargin := "375.1678"
	expectedMarginLevel := "5432.57"
	expectedUnexecutedValue := "42.42"
	// Unmarshal payload into struct
	response := new(GetTradeBalanceResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Equal(suite.T(), expectedEquivalentBalance, response.Result.EquivalentBalance.String())
	require.Equal(suite.T(), expectedTradeBalance, response.Result.TradeBalance.String())
	require.Equal(suite.T(), expectedUsedMargin, response.Result.MarginAmount.String())
	require.Equal(suite.T(), expectedUnrealizedPnl, response.Result.UnrealizedNetPNL.String())
	require.Equal(suite.T(), expectedPositionsCost, response.Result.CostBasis.String())
	require.Equal(suite.T(), expectedPositionsValuation, response.Result.FloatingValuation.String())
	require.Equal(suite.T(), expectedEquity, response.Result.Equity.String())
	require.Equal(suite.T(), expectedFreeMargin, response.Result.FreeMargin.String())
	require.Equal(suite.T(), expectedMarginLevel, response.Result.MarginLevel.String())
	require.Equal(suite.T(), expectedUnexecutedValue, response.Result.UnexecutedValue.String())
}
