package market

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for OHLC data interval
type OHLCInterval int

// Values for OHLCInterval
const (
	// 1 minute
	M1 OHLCInterval = 1
	// 5 minutes
	M5 OHLCInterval = 5
	// 15 minutes
	M15 OHLCInterval = 15
	// 30 minutes
	M30 OHLCInterval = 30
	// 60 minutes (1 hour)
	M60 OHLCInterval = 60
	// 240 minutes (4 hours)
	M240 OHLCInterval = 240
	// 1440 minutes (1 day)
	M1440 OHLCInterval = 1440
	// 10080 minutes (1 week)
	M10080 OHLCInterval = 10080
	// 21600 minutes (2 weeks)
	M21600 OHLCInterval = 21600
)

// Data of a single OHLC indicator
type OHLC struct {
	// Start timestamp for the indicator
	Timestamp time.Time
	// Price of the first trade
	Open string
	// Highest trade price
	High string
	// Lowest trade price
	Low string
	// Price of the last trade
	Close string
	// Volume average price
	VolumeAveragePrice string
	// Volume
	Volume string
	// Number of trades used to build the indicator
	TradesCount int64
}

// Marshal a single OHLC indicator as an array of strings to produce the same JSON data as the API.
//
// [int <unixsec>, string <open>, string <high>, string <low>, string <close>, string <vwap>, string <volume>, int <count>]
func (ohlc *OHLC) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		ohlc.Timestamp.Unix(),
		ohlc.Open,
		ohlc.High,
		ohlc.Low,
		ohlc.Close,
		ohlc.VolumeAveragePrice,
		ohlc.Volume,
		ohlc.TradesCount,
	})
}

// Parse a single OHLC indicator from the API raw JSON data as an array of any.
//
// [int <unixsec>, string <open>, string <high>, string <low>, string <close>, string <vwap>, string <volume>, int <count>]
func parseOHLCFromArray(input []interface{}) (OHLC, error) {
	// Cast timestamp as int64
	ts, ok := input[0].(int64)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse timestamp as int64. Got %v", input[0])
	}
	// Parse count as int64
	count, ok := input[7].(int64)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse trades count as int64. Got %v", input[7])
	}
	// Convert other entries as strings
	open, ok := input[1].(string)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse open as string. Got %v", input[1])
	}
	high, ok := input[2].(string)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse high as string. Got %v", input[2])
	}
	low, ok := input[3].(string)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse low as string. Got %v", input[3])
	}
	close, ok := input[4].(string)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse close as string. Got %v", input[4])
	}
	vap, ok := input[5].(string)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse volume average price as string. Got %v", input[5])
	}
	volume, ok := input[6].(string)
	if !ok {
		return OHLC{}, fmt.Errorf("could not parse volume as string. Got %v", input[6])
	}
	return OHLC{
		Timestamp:          time.Unix(ts, 0),
		Open:               open,
		High:               high,
		Low:                low,
		Close:              close,
		VolumeAveragePrice: vap,
		Volume:             volume,
		TradesCount:        count,
	}, nil
}

// Unmarshal a single OHLC indicator from the API raw JSON data.
//
// [int <unixsec>, string <open>, string <high>, string <low>, string <close>, string <vwap>, string <volume>, int <count>]
func (ohlc *OHLC) UnmarshalJSON(data []byte) error {
	// Unmarshal data into an array of strings
	tmp := []string{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// Parse timestamp as int64
	ts, err := strconv.ParseInt(tmp[0], 10, 64)
	if err != nil {
		return &json.UnmarshalTypeError{
			Value:  tmp[0],
			Type:   reflect.TypeOf(ohlc),
			Offset: int64(len(data)),
			Struct: "OHLC",
			Field:  ".[0]",
		}
	}
	// Parse count as int64
	count, err := strconv.ParseInt(tmp[7], 10, 64)
	if err != nil {
		return &json.UnmarshalTypeError{
			Value:  tmp[7],
			Type:   reflect.TypeOf(ohlc),
			Offset: int64(len(data)),
			Struct: "OHLC",
			Field:  ".[7]",
		}
	}
	// Encode OHLC and exit
	ohlc.Timestamp = time.Unix(ts, 0)
	ohlc.Open = tmp[1]
	ohlc.High = tmp[2]
	ohlc.Low = tmp[3]
	ohlc.Close = tmp[4]
	ohlc.VolumeAveragePrice = tmp[5]
	ohlc.Volume = tmp[6]
	ohlc.TradesCount = count
	return nil
}

// OHLC data returned by the API
type OHLCData struct {
	// Timestamp to be used as since when polling for new, committed OHLC data
	Last time.Time
	// Asset pair ID
	PairId string
	// OHLC data
	Data []OHLC
}

// Marshal OHLC data to produce the same raw data as the API.
func (ohlc *OHLCData) MarshalJSON() ([]byte, error) {
	// Put data into a map
	base := map[string]interface{}{
		ohlc.PairId: ohlc.Data,
		"last":      ohlc.Last.Unix(),
	}
	// Marshal map
	return json.Marshal(base)
}

// Unmarshal OHLC data from the API raw data.
func (ohlc *OHLCData) UnmarshalJSON(data []byte) error {
	// Unmarshal data into a map
	tmp := map[string]interface{}{}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// Check there are 2 map entries and extract the Pair ID
	keys := []string{}
	for key := range tmp {
		keys = append(keys, key) // Push discovered keys - Used in case of error
		if len(tmp) == 2 && key != "last" {
			ohlc.PairId = key
		}
	}
	if len(tmp) != 2 {
		// Return an error because of malformatted data
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", keys),
			Type:   reflect.TypeOf(ohlc),
			Offset: int64(len(data)),
			Struct: "OHLCData",
			Field:  ".",
		}
	}
	// Cast last as int64
	ts, ok := tmp["last"].(int64)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp["last"]),
			Type:   reflect.TypeOf(ohlc),
			Offset: int64(len(data)),
			Struct: "OHLCData",
			Field:  ".",
		}
	}
	ohlc.Last = time.Unix(ts, 0)
	// Convert OHLC data as array of arrays
	ohlc.Data = []OHLC{}
	ohlcs := tmp[ohlc.PairId].([][]interface{})
	for _, raw := range ohlcs {
		parsed, err := parseOHLCFromArray(raw)
		if err != nil {
			return &json.UnmarshalTypeError{
				Value:  fmt.Sprintf("%v", raw),
				Type:   reflect.TypeOf(OHLC{}),
				Offset: int64(len(data)),
				Struct: "OHLC",
				Field:  ".",
			}
		}
		ohlc.Data = append(ohlc.Data, parsed)
	}
	// Exit
	return nil
}

// GetOHLCData required parameters
type GetOHLCDataParameters struct {
	// Asset pair to get OHLC data for.
	Pair string
}

// GetOHLCData options.
type GetOHLCDataOptions struct {
	// Time frame interval in minutes.
	//
	// Default to 1. A zero value (= 0) triggers default behavior.
	Interval OHLCInterval
	// Return up to 720 OHLC data points since given timestamp. By default, return the most recent
	// OHLC data points.
	//
	// A zero value (IsZero() returns true) triggers default behavior (= most recent data).
	Since time.Time
}

// GetOHLCData Response
type GetOHLCDataResponse struct {
	common.KrakenSpotRESTResponse
	// OHLC data
	Result *OHLCData `json:"result,omitempty"`
}
