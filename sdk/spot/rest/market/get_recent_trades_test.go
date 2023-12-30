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

// Unit test suite for GetRecentTrades DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetRecentTradesTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetRecentTradesTestSuite(t *testing.T) {
	suite.Run(t, new(GetRecentTradesTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetRecentTradesResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetRecentTradesResponse struct.
func (suite *GetRecentTradesTestSuite) TestGetRecentTradesResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  "30243.40000",
			  "0.34507674",
			  1688669597.827736900,
			  "b",
			  "m",
			  "",
			  61044952
			],
			[
			  "30243.30000",
			  "0.00376960",
			  1688669598.2804112,
			  "s",
			  "l",
			  "",
			  61044953
			],
			[
			  "30243.30000",
			  "0.01235716",
			  1688669602.698379,
			  "s",
			  "m",
			  "",
			  61044956
			]
		  ],
		  "last": "1688671969993150842"
		}
	}`
	expectedResultCount := 3
	expectedPairId := "XXBTZUSD"
	expectedTrade1Timestamp := int64(1688669597827736900)
	expectedTrade1Price := "30243.40000"
	expectedLast := int64(1688671969993150842)
	// Unmarshal payload into struct
	response := new(GetRecentTradesResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Trades, expectedResultCount)
	require.Equal(suite.T(), expectedPairId, response.Result.PairId)
	// Ensure exact trade timestamp amd rendered trade timestamp as a unix nanosec timestamp are equal +- 1 microsecond
	require.InDelta(suite.T(), expectedTrade1Timestamp, response.Result.Trades[0].Timestamp.UnixNano(), 1000)
	require.Equal(suite.T(), expectedTrade1Price, response.Result.Trades[0].Price)
	require.Equal(suite.T(), expectedLast, response.Result.Last)
}

// Test the JSON marshaller of GetRecentTradesResponse.
//
// The test will ensure:
//   - GetRecentTradesResponse can be marshalled into the some JSON payload as the API.
func (suite *GetRecentTradesTestSuite) TestGetRecentTradesResponseMarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": [
			[
			  "30243.40000",
			  "0.34507674",
			  1688669597.8277369,
			  "b",
			  "m",
			  "",
			  61044952
			],
			[
			  "30243.30000",
			  "0.00376960",
			  1688669598.2804112,
			  "s",
			  "l",
			  "",
			  61044953
			],
			[
			  "30243.30000",
			  "0.01235716",
			  1688669602.698379,
			  "s",
			  "m",
			  "",
			  61044956
			]
		  ],
		  "last": "1688671969993150842"
		}
	  }`
	// Compact payload so we can compare it with the output of Marshal
	target := &bytes.Buffer{}
	err := json.Compact(target, []byte(payload))
	require.NoError(suite.T(), err)
	expectedCompactPayload, err := io.ReadAll(target)
	require.NoError(suite.T(), err)
	// Unmarshal payload into struct
	response := new(GetRecentTradesResponse)
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
