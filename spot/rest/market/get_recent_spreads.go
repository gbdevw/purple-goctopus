package market

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Data of a single spread
type Spread struct {
	// Timestamp as a Unix timestamp (seconds)
	Timestamp int64
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
		spread.Timestamp,
		spread.BestBid,
		spread.BestAsk,
	})
}

// Unmarshal spread data from an array of data from the API.
//
// [int <unixsec>, string <bid>, string <ask>]
func unmarshalSpreadFromArray(input []interface{}) (*Spread, error) {
	// Convert timestamp to int64
	ts, ok := input[0].(float64)
	if !ok {
		return &Spread{}, fmt.Errorf("could not parse timestamp as int64. Got %v", input[0])
	}
	// Convert other items to string
	bid, ok := input[1].(string)
	if !ok {
		return &Spread{}, fmt.Errorf("could not parse spread best bid as text. Got %v", input[0])
	}
	ask, ok := input[2].(string)
	if !ok {
		return &Spread{}, fmt.Errorf("could not parse spread best ask as text. Got %v", input[1])
	}
	// Build and return trade data
	return &Spread{
		Timestamp: int64(ts),
		BestBid:   bid,
		BestAsk:   ask,
	}, nil
}

// Spread data returned by the API.
type SpreadData struct {
	// ID to be used as since when polling for new spread data
	Last int64
	// Asset pair ID
	PairId string
	// Spreads by pair
	Spreads []*Spread
}

// Marshal spreads data to produce the same raw data as the API.
func (spreads *SpreadData) MarshalJSON() ([]byte, error) {
	// Put data into a map
	base := map[string]interface{}{
		spreads.PairId: spreads.Spreads,
		"last":         spreads.Last,
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
	// Cast last as float64
	ts, ok := tmp["last"].(float64)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp["last"]),
			Type:   reflect.TypeOf(spreads),
			Offset: int64(len(data)),
			Struct: "SpreadData",
			Field:  ".",
		}
	}
	spreads.Last = int64(ts)
	// Convert OHLC data as an array of object
	spreads.Spreads = []*Spread{}
	tdata, ok := tmp[spreads.PairId].([]interface{})
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp[spreads.PairId]),
			Type:   reflect.TypeOf(Spread{}),
			Offset: int64(len(data)),
			Struct: "Spread",
			Field:  ".",
		}
	}
	for _, raw := range tdata {
		// Cast raw to an array of object
		item, ok := raw.([]interface{})
		if !ok {
			return &json.UnmarshalTypeError{
				Value:  fmt.Sprintf("%v", raw),
				Type:   reflect.TypeOf(Spread{}),
				Offset: int64(len(data)),
				Struct: "Spread",
				Field:  ".",
			}
		}
		parsed, err := unmarshalSpreadFromArray(item)
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

// GetRecentSpreads request parameters
type GetRecentSpreadsRequestParameters struct {
	// Asset pair to get data for.
	Pair string `json:"pair"`
}

// GetRecentSpreads request options
type GetRecentSpreadsRequestOptions struct {
	// Return up to 1000 recent spreads since given uniix timestamp.
	//
	// By default, return the most recent spreads. A zero value triggers default behavior.
	Since int64 `json:"since,omitempty"`
}

// GetRecentSpreads response
type GetRecentSpreadsResponse struct {
	common.KrakenSpotRESTResponse
	Result *SpreadData `json:"result,omitempty"`
}
