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

// Unit test suite for GetTradableAssetPairs DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetTradableAssetPairsTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetTradableAssetPairsTestSuite(t *testing.T) {
	suite.Run(t, new(GetTradableAssetPairsTestSuite))
}

/*************************************************************************************************/
/* UNIT TESTS                                                                                    */
/*************************************************************************************************/

// Test the JSON unmarshaller and marshaller of GetTradableAssetPairsResponse.
//
// The test will ensure:
//   - A valid JSON response from the API can be unmarshalled into the corresponding GetTradableAssetPairsResponse struct.
//   - A GetTradableAssetPairsResponse can be marshalled to the same JSON paylaod as the API.
func (suite *GetTradableAssetPairsTestSuite) TestGetTradableAssetPairsResponseJSON() {
	// Test settings, expectations, ...
	payload := `{
		"error": [],
		"result": {
		  "XETHXXBT": {
			"altname": "ETHXBT",
			"wsname": "ETH/XBT",
			"aclass_base": "currency",
			"base": "XETH",
			"aclass_quote": "currency",
			"quote": "XXBT",
			"cost_decimals": 6,
			"pair_decimals": 5,
			"lot_decimals": 8,
			"lot_multiplier": 1,
			"leverage_buy": [
			  2,
			  3,
			  4,
			  5
			],
			"leverage_sell": [
			  2,
			  3,
			  4,
			  5
			],
			"fees": [
			  [
				0,
				0.26
			  ],
			  [
				50000,
				0.24
			  ],
			  [
				100000,
				0.22
			  ],
			  [
				250000,
				0.2
			  ],
			  [
				500000,
				0.18
			  ],
			  [
				1000000,
				0.16
			  ],
			  [
				2500000,
				0.14
			  ],
			  [
				5000000,
				0.12
			  ],
			  [
				10000000,
				0.1
			  ]
			],
			"fees_maker": [
			  [
				0,
				0.16
			  ],
			  [
				50000,
				0.14
			  ],
			  [
				100000,
				0.12
			  ],
			  [
				250000,
				0.1
			  ],
			  [
				500000,
				0.08
			  ],
			  [
				1000000,
				0.06
			  ],
			  [
				2500000,
				0.04
			  ],
			  [
				5000000,
				0.02
			  ],
			  [
				10000000,
				0
			  ]
			],
			"fee_volume_currency": "ZUSD",
			"margin_call": 80,
			"margin_stop": 40,
			"ordermin": "0.01",
			"costmin": "0.00002",
			"tick_size": "0.00001",
			"status": "online",
			"long_position_limit": 1100,
			"short_position_limit": 400
		  },
		  "XXBTZUSD": {
			"altname": "XBTUSD",
			"wsname": "XBT/USD",
			"aclass_base": "currency",
			"base": "XXBT",
			"aclass_quote": "currency",
			"quote": "ZUSD",
			"cost_decimals": 5,
			"pair_decimals": 1,
			"lot_decimals": 8,
			"lot_multiplier": 1,
			"leverage_buy": [
			  2,
			  3,
			  4,
			  5
			],
			"leverage_sell": [
			  2,
			  3,
			  4,
			  5
			],
			"fees": [
			  [
				0,
				0.26
			  ],
			  [
				50000,
				0.24
			  ],
			  [
				100000,
				0.22
			  ],
			  [
				250000,
				0.2
			  ],
			  [
				500000,
				0.18
			  ],
			  [
				1000000,
				0.16
			  ],
			  [
				2500000,
				0.14
			  ],
			  [
				5000000,
				0.12
			  ],
			  [
				10000000,
				0.1
			  ]
			],
			"fees_maker": [
			  [
				0,
				0.16
			  ],
			  [
				50000,
				0.14
			  ],
			  [
				100000,
				0.12
			  ],
			  [
				250000,
				0.1
			  ],
			  [
				500000,
				0.08
			  ],
			  [
				1000000,
				0.06
			  ],
			  [
				2500000,
				0.04
			  ],
			  [
				5000000,
				0.02
			  ],
			  [
				10000000,
				0
			  ]
			],
			"fee_volume_currency": "ZUSD",
			"margin_call": 80,
			"margin_stop": 40,
			"ordermin": "0.0001",
			"costmin": "0.5",
			"tick_size": "0.1",
			"status": "online",
			"long_position_limit": 250,
			"short_position_limit": 200
		  }
		}
	}`
	// Compact payload so we can compare it with the output of Marshal
	target := &bytes.Buffer{}
	err := json.Compact(target, []byte(payload))
	require.NoError(suite.T(), err)
	expectedCompactPayload, err := io.ReadAll(target)
	require.NoError(suite.T(), err)
	// Unmarshal into GetTickerInformationResponse and check
	response := new(GetTradableAssetPairsResponse)
	err = json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), response.Error)
	require.NotEmpty(suite.T(), response.Result)
	// Marshal and compare
	result, err := json.Marshal(response)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), string(expectedCompactPayload), string(result))
}
