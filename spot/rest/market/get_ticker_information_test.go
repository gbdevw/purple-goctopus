package market

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test AssetTickerInfo helper methods and JSON unmarshalling.
//
// Test will ensure:
//   - A predefined JSON payload from doc. can be unmarshalled as a AssetTickerInfo
//   - Each piece of data of AssetTickerInfo is equal to what should be returned by the
//     corresponding helper method
func TestAssetTickerInfo(t *testing.T) {
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
	require.NoError(t, err)
	// Check each piece of data against the corresponding helper method
	// Check Ask
	require.Equal(t, "30300.10000", ticker.Ask[0])
	require.Equal(t, ticker.Ask[0], ticker.GetAskPrice())
	require.Equal(t, "1", ticker.Ask[1])
	require.Equal(t, ticker.Ask[1], ticker.GetAskWholeLotVolume())
	require.Equal(t, "1.000", ticker.Ask[2])
	require.Equal(t, ticker.Ask[2], ticker.GetAskLotVolume())
	// Check Bid
	require.Equal(t, "30300.10000", ticker.Bid[0])
	require.Equal(t, ticker.Bid[0], ticker.GetBidPrice())
	require.Equal(t, "1", ticker.Bid[1])
	require.Equal(t, ticker.Bid[1], ticker.GetBidWholeLotVolume())
	require.Equal(t, "1.000", ticker.Bid[2])
	require.Equal(t, ticker.Bid[2], ticker.GetBidLotVolume())
	// Check Close
	require.Equal(t, "30303.20000", ticker.Close[0])
	require.Equal(t, ticker.Close[0], ticker.GetLastTradePrice())
	require.Equal(t, "0.00067643", ticker.Close[1])
	require.Equal(t, ticker.Close[1], ticker.GetLastTradeLotVolume())
	// Check volume
	require.Equal(t, "4083.67001100", ticker.Volume[0])
	require.Equal(t, ticker.Volume[0], ticker.GetTodayVolume())
	require.Equal(t, "4412.73601799", ticker.Volume[1])
	require.Equal(t, ticker.Volume[1], ticker.GetPast24HVolume())
	// Check volume average price
	require.Equal(t, "30706.77771", ticker.VolumeAveragePrice[0])
	require.Equal(t, ticker.VolumeAveragePrice[0], ticker.GetTodayVolumeAveragePrice())
	require.Equal(t, "30689.13205", ticker.VolumeAveragePrice[1])
	require.Equal(t, ticker.VolumeAveragePrice[1], ticker.GetPast24HVolumeAveragePrice())
	// Check trades
	require.Equal(t, int64(34619), ticker.Trades[0])
	require.Equal(t, ticker.Trades[0], ticker.GetTodayTradeCount())
	require.Equal(t, int64(38907), ticker.Trades[1])
	require.Equal(t, ticker.Trades[1], ticker.GetPast24HTradeCount())
	// Check low
	require.Equal(t, "29868.30000", ticker.Low[0])
	require.Equal(t, ticker.Low[0], ticker.GetTodayLow())
	require.Equal(t, "29868.30000", ticker.Low[1])
	require.Equal(t, ticker.Low[1], ticker.GetPast24HLow())
	// Check high
	require.Equal(t, "31631.00000", ticker.High[0])
	require.Equal(t, ticker.High[0], ticker.GetTodayHigh())
	require.Equal(t, "31631.00000", ticker.High[1])
	require.Equal(t, ticker.High[1], ticker.GetPast24HHigh())
	// Check open
	require.Equal(t, "30502.80000", ticker.OpeningPrice)
	require.Equal(t, ticker.OpeningPrice, ticker.GetTodayOpen())
}
