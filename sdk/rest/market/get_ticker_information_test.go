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

// Unit test suite for GetTickerInformation DTO.
//
// The test suite ensures all DTO can be marshalled/unmarshalled to/from JSON payloads used by the
// Kraken Spot REST API.
type GetTickerInformationTestSuite struct {
	suite.Suite
}

// Run unit test suite
func TestGetTickerInformationTestSuite(t *testing.T) {
	suite.Run(t, new(GetTickerInformationTestSuite))
}

// Test AssetTickerInfo helper methods and JSON unmarshalling.
//
// Test will ensure:
//   - A predefined JSON payload from doc. can be unmarshalled as a AssetTickerInfo
//   - Each piece of data of AssetTickerInfo is equal to what should be returned by the
//     corresponding helper method
func (suite *GetTickerInformationTestSuite) TestAssetTickerInfo() {
	// Predefined JSON payload
	payload := `{
		"a": [
		"30300.10000",
		"1",
		"1.000"
		],
		"b": [
		"30300.00000",
		"1",
		"1.000"
		],
		"c": [
		"30303.20000",
		"0.00067643"
		],
		"v": [
		"4083.67001100",
		"4412.73601799"
		],
		"p": [
		"30706.77771",
		"30689.13205"
		],
		"t": [
		34619,
		38907
		],
		"l": [
		"29868.30000",
		"29868.30000"
		],
		"h": [
		"31631.00000",
		"31631.00000"
		],
		"o": "30502.80000"
		}`
	// Unmarshal into AssetTickerInfo
	ticker := new(AssetTickerInfo)
	err := json.Unmarshal([]byte(payload), ticker)
	require.NoError(suite.T(), err)
	// Check each piece of data against the corresponding helper method
	// Check Ask
	require.Equal(suite.T(), "30300.10000", ticker.Ask[0])
	require.Equal(suite.T(), ticker.Ask[0], ticker.GetAskPrice())
	require.Equal(suite.T(), "1", ticker.Ask[1])
	require.Equal(suite.T(), ticker.Ask[1], ticker.GetAskWholeLotVolume())
	require.Equal(suite.T(), "1.000", ticker.Ask[2])
	require.Equal(suite.T(), ticker.Ask[2], ticker.GetAskLotVolume())
	// Check Bid
	require.Equal(suite.T(), "30300.00000", ticker.Bid[0])
	require.Equal(suite.T(), ticker.Bid[0], ticker.GetBidPrice())
	require.Equal(suite.T(), "1", ticker.Bid[1])
	require.Equal(suite.T(), ticker.Bid[1], ticker.GetBidWholeLotVolume())
	require.Equal(suite.T(), "1.000", ticker.Bid[2])
	require.Equal(suite.T(), ticker.Bid[2], ticker.GetBidLotVolume())
	// Check Close
	require.Equal(suite.T(), "30303.20000", ticker.Close[0])
	require.Equal(suite.T(), ticker.Close[0], ticker.GetLastTradePrice())
	require.Equal(suite.T(), "0.00067643", ticker.Close[1])
	require.Equal(suite.T(), ticker.Close[1], ticker.GetLastTradeLotVolume())
	// Check volume
	require.Equal(suite.T(), "4083.67001100", ticker.Volume[0])
	require.Equal(suite.T(), ticker.Volume[0], ticker.GetTodayVolume())
	require.Equal(suite.T(), "4412.73601799", ticker.Volume[1])
	require.Equal(suite.T(), ticker.Volume[1], ticker.GetPast24HVolume())
	// Check volume average price
	require.Equal(suite.T(), "30706.77771", ticker.VolumeAveragePrice[0])
	require.Equal(suite.T(), ticker.VolumeAveragePrice[0], ticker.GetTodayVolumeAveragePrice())
	require.Equal(suite.T(), "30689.13205", ticker.VolumeAveragePrice[1])
	require.Equal(suite.T(), ticker.VolumeAveragePrice[1], ticker.GetPast24HVolumeAveragePrice())
	// Check trades
	require.Equal(suite.T(), int64(34619), ticker.Trades[0])
	require.Equal(suite.T(), ticker.Trades[0], ticker.GetTodayTradeCount())
	require.Equal(suite.T(), int64(38907), ticker.Trades[1])
	require.Equal(suite.T(), ticker.Trades[1], ticker.GetPast24HTradeCount())
	// Check low
	require.Equal(suite.T(), "29868.30000", ticker.Low[0])
	require.Equal(suite.T(), ticker.Low[0], ticker.GetTodayLow())
	require.Equal(suite.T(), "29868.30000", ticker.Low[1])
	require.Equal(suite.T(), ticker.Low[1], ticker.GetPast24HLow())
	// Check high
	require.Equal(suite.T(), "31631.00000", ticker.High[0])
	require.Equal(suite.T(), ticker.High[0], ticker.GetTodayHigh())
	require.Equal(suite.T(), "31631.00000", ticker.High[1])
	require.Equal(suite.T(), ticker.High[1], ticker.GetPast24HHigh())
	// Check open
	require.Equal(suite.T(), "30502.80000", ticker.OpeningPrice)
	require.Equal(suite.T(), ticker.OpeningPrice, ticker.GetTodayOpen())
}

// Test GetTickerInformationResponse JSON unmarshalling.
//
// Test will ensure:
//   - A predefined JSON payload from doc. can be unmarshalled as a GetTickerInformationResponse
func (suite *GetTickerInformationTestSuite) TestGetTickerInformationUnmarshalJSON() {
	// Predefined JSON payload
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": {
			"a": [
			  "30300.10000",
			  "1",
			  "1.000"
			],
			"b": [
			  "30300.00000",
			  "1",
			  "1.000"
			],
			"c": [
			  "30303.20000",
			  "0.00067643"
			],
			"v": [
			  "4083.67001100",
			  "4412.73601799"
			],
			"p": [
			  "30706.77771",
			  "30689.13205"
			],
			"t": [
			  34619,
			  38907
			],
			"l": [
			  "29868.30000",
			  "29868.30000"
			],
			"h": [
			  "31631.00000",
			  "31631.00000"
			],
			"o": "30502.80000"
		  }
		}
	}`
	// Unmarshal into GetTickerInformationResponse and check
	response := new(GetTickerInformationResponse)
	err := json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), response.Error)
	require.NotEmpty(suite.T(), response.Result)
}

// Test GetTickerInformationResponse JSON marshalling.
//
// Test will ensure:
//   - A GetTickerInformationResponse can be marshalled to the exact same payload as a predefined response from the API.
func (suite *GetTickerInformationTestSuite) TestGetTickerInformationMarshalJSON() {
	// Predefined JSON payload
	payload := `{
		"error": [],
		"result": {
		  "XXBTZUSD": {
			"a": [
			  "30300.10000",
			  "1",
			  "1.000"
			],
			"b": [
			  "30300.00000",
			  "1",
			  "1.000"
			],
			"c": [
			  "30303.20000",
			  "0.00067643"
			],
			"v": [
			  "4083.67001100",
			  "4412.73601799"
			],
			"p": [
			  "30706.77771",
			  "30689.13205"
			],
			"t": [
			  34619,
			  38907
			],
			"l": [
			  "29868.30000",
			  "29868.30000"
			],
			"h": [
			  "31631.00000",
			  "31631.00000"
			],
			"o": "30502.80000"
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
	response := new(GetTickerInformationResponse)
	err = json.Unmarshal([]byte(payload), response)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), response.Error)
	require.NotEmpty(suite.T(), response.Result)
	// Marshal and compare
	result, err := json.Marshal(response)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), string(expectedCompactPayload), string(result))
}
