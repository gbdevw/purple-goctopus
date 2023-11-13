package market

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Data of a single spread
type Spread struct {
	// Timestamp
	Timestamp time.Time
	// Best bid
	BestBid string
	// Best ask
	BestAsk string
}

// Marshal spread data to produce the same payload as the API.
//
// [int <unixsec>, string <bid>, string <ask>]
func (spread *Spread) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		spread.Timestamp.Unix(),
		spread.BestBid,
		spread.BestAsk,
	})
}

// Unmarshal spread data from an array of data from the API.
//
// [int <unixsec>, string <bid>, string <ask>]
func unmarshalSpreadFromArray(input []interface{}) (Spread, error) {
	// Convert timestamp to int64
	ts, ok := input[0].(int64)
	if !ok {
		return Spread{}, fmt.Errorf("could not parse timestamp as int64. Got %v", input[0])
	}
	// Convert other items to string
	bid, ok := input[1].(string)
	if !ok {
		return Spread{}, fmt.Errorf("could not parse spread best bid as text. Got %v", input[0])
	}
	ask, ok := input[2].(string)
	if !ok {
		return Spread{}, fmt.Errorf("could not parse spread best ask as text. Got %v", input[1])
	}
	// Build and return trade data
	return Spread{
		Timestamp: time.Unix(ts, 0),
		BestBid:   bid,
		BestAsk:   ask,
	}, nil
}

// Spread data returned by the API.
type SpreadData struct {
	// ID to be used as since when polling for new spread data
	Last time.Time
	// Asset pair ID
	PairId string
	// Spreads by pair
	Spreads []Spread
}

// Marshal spreads data to produce the same raw data as the API.
func (spreads *SpreadData) MarshalJSON() ([]byte, error) {
	// Put data into a map
	base := map[string]interface{}{
		spreads.PairId: spreads.Spreads,
		"last":         spreads.Last.Unix(),
	}
	// Marshal map
	return json.Marshal(base)
}

// Unmarshal spreads data from the API raw data.
func (spreads *SpreadData) UnmarshalJSON(data []byte) error {
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
			spreads.PairId = key
		}
	}
	if len(tmp) != 2 {
		// Return an error because of malformatted data
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", keys),
			Type:   reflect.TypeOf(spreads),
			Offset: int64(len(data)),
			Struct: "SpreadData",
			Field:  ".",
		}
	}
	// Cast last as int64
	ts, ok := tmp["last"].(int64)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp["last"]),
			Type:   reflect.TypeOf(spreads),
			Offset: int64(len(data)),
			Struct: "SpreadData",
			Field:  ".",
		}
	}
	spreads.Last = time.Unix(ts, 0)
	// Convert OHLC data as array of arrays
	spreads.Spreads = []Spread{}
	tdata := tmp[spreads.PairId].([][]interface{})
	for _, raw := range tdata {
		parsed, err := unmarshalSpreadFromArray(raw)
		if err != nil {
			return &json.UnmarshalTypeError{
				Value:  fmt.Sprintf("%v", raw),
				Type:   reflect.TypeOf(Spread{}),
				Offset: int64(len(data)),
				Struct: "Spread",
				Field:  ".",
			}
		}
		spreads.Spreads = append(spreads.Spreads, parsed)
	}
	// Exit
	return nil
}

// GetRecentSpreads required parameters
type GetRecentSpreadsParameters struct {
	// Asset pair to get data for.
	Pair string
}

// GetRecentSpreads options
type GetRecentSpreadsOptions struct {
	// Return up to 1000 recent spreads since given timestamp.
	//
	// By default, return the most recent spreads. A zero value triggers default behavior.
	Since time.Time
}

// GetRecentSpreads response
type GetRecentSpreadsResponse struct {
	common.KrakenSpotRESTResponse
	Result *SpreadData `json:"result,omitempty"`
}
