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

// Unit test suite for GetOrderBook DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetOrderBookTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetOrderBookTestSuite(t *testing.T) {
	suite.Run(t, new(GetOrderBookTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller of GetOrderBookResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetOrderBookResponse struct.
func (suite *GetOrderBookTestSuite) TestGetOrderBookResponseUnmarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": {
			"asks": [
			  [
				"30384.10000",
				"2.059",
				1688671659
			  ],
			  [
				"30387.90000",
				"1.500",
				1688671380
			  ],
			  [
				"30393.70000",
				"9.871",
				1688671261
			  ]
			],
			"bids": [
			  [
				"30297.00000",
				"1.115",
				1688671636
			  ],
			  [
				"30296.70000",
				"2.002",
				1688671674
			  ],
			  [
				"30289.80000",
				"5.001",
				1688671673
			  ]
			]
		  }
		}
	}`
	expectedResultCountPerSide := 3
	expectedPairId := "XXBTZUSD"
	expectedAsk1Timestamp := int64(1688671659)
	// Unmarshal payload into struct
	response := new(GetOrderBookResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	// Check data
	require.Empty(suite.T(), response.Error)
	require.NotNil(suite.T(), response.Result)
	require.Len(suite.T(), response.Result.Asks, expectedResultCountPerSide)
	require.Len(suite.T(), response.Result.Bids, expectedResultCountPerSide)
	require.Equal(suite.T(), expectedPairId, response.Result.PairId)
	require.Equal(suite.T(), expectedAsk1Timestamp, response.Result.Asks[0].Timestamp)
}

// Test the JSON marshaller of GetOrderBookResponse.
//
// The test will ensure:
//   - GetOrderBookResponse can be marshalled into the some JSON payload as the API.
func (suite *GetOrderBookTestSuite) TestGetOrderBookResponseMarshalJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": {
			"asks": [
			  [
				"30384.10000",
				"2.059",
				1688671659
			  ],
			  [
				"30387.90000",
				"1.500",
				1688671380
			  ],
			  [
				"30393.70000",
				"9.871",
				1688671261
			  ]
			],
			"bids": [
			  [
				"30297.00000",
				"1.115",
				1688671636
			  ],
			  [
				"30296.70000",
				"2.002",
				1688671674
			  ],
			  [
				"30289.80000",
				"5.001",
				1688671673
			  ]
			]
		  }
		}
	}`
	// Compact payload so we can compare it with the output of Marshal
	target := &bytes.Buffer{}
	err := json.Compact(target, []byte(payload))
	require.NoError(suite.T(), err)
	expectedCompactPayload, err := io.ReadAll(target)
	require.NoError(suite.T(), err)
	// Unmarshal payload into struct
	response := new(GetOrderBookResponse)
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
