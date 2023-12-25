package market

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Order book data
type OrderBook struct {
	// Book pair ID
	PairId string
	// Ask side of the order book
	Asks []*OrderBookEntry
	// Bid side of the order book
	Bids []*OrderBookEntry
}

// Marshal book data to produce the same JSON data as the API
func (book *OrderBook) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]map[string][]*OrderBookEntry{
		book.PairId: {
			"asks": book.Asks,
			"bids": book.Bids,
		},
	})
}

func (book *OrderBook) UnmarshalJSON(data []byte) error {
	// Unmarshal in tmp struct
	tmp := map[string]map[string][]*OrderBookEntry{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// Encode book
	keys := []string{}
	for key := range tmp {
		keys = append(keys, key) // Push keys for error msg if needed
		if len(tmp) == 1 {
			// Save key as pair ID, extract bids/asks and exit
			book.PairId = key
			book.Asks = tmp[key]["asks"]
			book.Bids = tmp[key]["bids"]
			return nil
		}
	}
	// Wrong number of keys -> malformatted data
	return &json.UnmarshalTypeError{
		Value:  fmt.Sprintf("%v", keys),
		Type:   reflect.TypeOf(book),
		Offset: int64(len(data)),
		Struct: "OrderBook",
		Field:  ".",
	}
}

// Order book entry
type OrderBookEntry struct {
	// Price level
	Price string
	// Volume
	Volume string
	// Last update timestamp as a Unix timestamp (seconds)
	Timestamp int64
}

// Marshal book entry data to produce the same JSON data as the API.
//
// [price<string>, volume<string>, timestamp<unixsec>]
func (entry *OrderBookEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		entry.Price,
		entry.Volume,
		entry.Timestamp,
	})
}

// Unmarshal OHLC data from the API raw data.
func (entry *OrderBookEntry) UnmarshalJSON(data []byte) error {
	// Unmarshal in array of strings
	tmp := []json.Number{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// Parse timestmap as int64
	ts, err := tmp[2].Int64()
	if err != nil {
		return &json.UnmarshalTypeError{
			Value:  tmp[2].String(),
			Type:   reflect.TypeOf(int64(0)),
			Offset: int64(len(data)),
			Struct: "int64",
			Field:  ".[2]",
		}
	}
	// Encode entry & exit
	entry.Timestamp = ts
	entry.Price = tmp[0].String()
	entry.Volume = tmp[1].String()
	return nil
}

// GetOrderBook request parameters
type GetOrderBookRequestParameters struct {
	// Asset pair to get data for.
	Pair string `json:"pair"`
}

// GetOrderBook request options
type GetOrderBookRequestOptions struct {
	// Maximum number of bid/ask entries : [1,500].
	//
	// Defaults to 100. A zero value (= 0) triggers default behavior.
	Count int `json:"count,omitempty"`
}

// GetOrderBook Response
type GetOrderBookResponse struct {
	common.KrakenSpotRESTResponse
	// Order book data
	Result *OrderBook `json:"result,omitempty"`
}
