package market

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*************************************************************************************************/
/* TEST SUITE                                                                                    */
/*************************************************************************************************/

// Unit test suite for GetRecentSpreads DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetRecentSpreadsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetRecentSpreadsTestSuite(t *testing.T) {
	suite.Run(t, new(GetRecentSpreadsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetRecentSpreadsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetRecentSpreadsResponse struct.
func (suite *GetRecentSpreadsTestSuite) TestGetRecentSpreadsResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  1688671834,
			  "30292.10000",
			  "30297.50000"
			],
			[
			  1688671834,
			  "30292.10000",
			  "30296.70000"
			],
			[
			  1688671834,
			  "30292.70000",
			  "30296.70000"
			]
		  ],
		  "last": 1688672106
		}
	}`
	expectedResultCount := 3
	expectedPairId := "XXBTZUSD"
	expectedSpread1Timestamp := int64(1688671834)
	expectedSpread1Ask := "30297.50000"
	expectedLast := int64(1688672106)
	// Unmarshal payload into struct
	response := new(GetRecentSpreadsResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Spreads, expectedResultCount)
	require.Equal(suite.T(), expectedPairId, response.Result.PairId)
	require.Equal(suite.T(), expectedSpread1Timestamp, response.Result.Spreads[0].Timestamp)
	require.Equal(suite.T(), expectedSpread1Ask, response.Result.Spreads[0].BestAsk)
	require.Equal(suite.T(), expectedLast, response.Result.Last)
}

// Test the JSON marshaller of GetRecentSpreadsResponse.
//
// The test will ensure:
//   - GetRecentSpreadsResponse can be marshalled into the some JSON payload as the API.
func (suite *GetRecentSpreadsTestSuite) TestGetRecentSpreadsResponseMarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  1688671834,
			  "30292.10000",
			  "30297.50000"
			],
			[
			  1688671834,
			  "30292.10000",
			  "30296.70000"
			],
			[
			  1688671834,
			  "30292.70000",
			  "30296.70000"
			]
		  ],
		  "last": 1688672106
		}
	}`
	// Compact payload so we can compare it with the output of Marshal
	target := &bytes.Buffer{}
	err := json.Compact(target, []byte(payload))
	require.NoError(suite.T(), err)
	expectedCompactPayload, err := io.ReadAll(target)
	require.NoError(suite.T(), err)
	// Unmarshal payload into struct
	response := new(GetRecentSpreadsResponse)
	err = json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	// Marshal and compare
	result, err := json.Marshal(response)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), string(expectedCompactPayload), string(result))
}
