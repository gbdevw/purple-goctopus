package market

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
)

// Enum for trade types
type TradeType string

// Values for TradeType
const (
	TradeTypeMarket TradeType = "market"
	TradeTypeLimit  TradeType = "limit"
)

// Trade data.
type Trade struct {
	// Trade price
	Price string
	// Trade volume
	Volume string
	// Trade timestamp
	Timestamp time.Time
	// Side: buy or sell
	Side string
	// Trade type: market or limit
	Type string
	// Misc.
	Miscellaneous string
	// Trade ID
	Id int64
}

// Marshal trade data to produce the same payload as the API.
//
// [<price>, <volume>, <time>, <buy/sell>, <market/limit>, <miscellaneous>, <trade_id>]
func (trade *Trade) MarshalJSON() ([]byte, error) {
	// Print trade
	return json.Marshal([]interface{}{
		trade.Price,
		trade.Volume,
		common.Must(strconv.ParseFloat(fmt.Sprintf("%d.%d", trade.Timestamp.Unix(), trade.Timestamp.Nanosecond()), 64)),
		trade.Side,
		trade.Type,
		trade.Miscellaneous,
		trade.Id,
	})
}

// Unmarshal trade data from an array of data from the API.
//
// [<price>, <volume>, <time>, <buy/sell>, <market/limit>, <miscellaneous>, <trade_id>]
func unmarshalTradeFromArray(input []interface{}) (*Trade, error) {
	// Convert timestamp to float64
	tsflo, ok := input[2].(float64)
	if !ok {
		return &Trade{}, fmt.Errorf("could not parse timestamp as float64. Got %v", input[3])
	}
	// Split decimal and integrer parts of the timestamp
	splits := strings.Split(strconv.FormatFloat(tsflo, 'f', 9, 64), ".")
	// Parse each split as int64 -> will not fail as input is checked
	sec := common.Must(strconv.ParseInt(splits[0], 10, 64))
	nsec := common.Must(strconv.ParseInt(splits[1], 10, 64))
	// Convert Id to int64
	id, ok := input[6].(float64)
	if !ok {
		return &Trade{}, fmt.Errorf("could not parse trade id as float64. Got %v", input[6])
	}
	// Convert other items to string
	price, ok := input[0].(string)
	if !ok {
		return &Trade{}, fmt.Errorf("could not parse trade price as text. Got %v", input[0])
	}
	volume, ok := input[1].(string)
	if !ok {
		return &Trade{}, fmt.Errorf("could not parse trade volume as text. Got %v", input[1])
	}
	side, ok := input[3].(string)
	if !ok {
		return &Trade{}, fmt.Errorf("could not parse trade side as text. Got %v", input[3])
	}
	typ, ok := input[4].(string)
	if !ok {
		return &Trade{}, fmt.Errorf("could not parse trade type as text. Got %v", input[4])
	}
	misc, ok := input[5].(string)
	if !ok {
		return &Trade{}, fmt.Errorf("could not parse trade miscellaneous as text. Got %v", input[5])
	}
	// Build and return trade data
	return &Trade{
		Price:         price,
		Volume:        volume,
		Timestamp:     time.Unix(sec, nsec),
		Side:          side,
		Type:          typ,
		Miscellaneous: misc,
		Id:            int64(id),
	}, nil
}

// Recent trade data last + asset > [[<price>, <volume>, <time>, <buy/sell>, <market/limit>, <miscellaneous>, <trade_id>]]
type RecentTrades struct {
	// Asset pair ID
	PairId string
	// Timestamp (Unix - nanoseconds) to be used as since to fetch next trades data
	Last int64
	// Trades
	Trades []*Trade
}

// Marshal trades data to produce the same raw data as the API.
func (trades *RecentTrades) MarshalJSON() ([]byte, error) {
	// Put data into a map
	base := map[string]interface{}{
		trades.PairId: trades.Trades,
		"last":        strconv.FormatInt(trades.Last, 10),
	}
	// Marshal map
	return json.Marshal(base)
}

// Unmarshal trades data from the API raw data.
func (trades *RecentTrades) UnmarshalJSON(data []byte) error {
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
			trades.PairId = key
		}
	}
	if len(tmp) != 2 {
		// Return an error because of malformatted data
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", keys),
			Type:   reflect.TypeOf(trades),
			Offset: int64(len(data)),
			Struct: "RecentTrades",
			Field:  ".",
		}
	}
	// Parse last as int64
	tsstr, ok := tmp["last"].(string)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp["last"]),
			Type:   reflect.TypeOf(""),
			Offset: int64(len(data)),
			Struct: "RecentTrades",
			Field:  ".",
		}
	}
	ts, err := strconv.ParseInt(tsstr, 10, 64)
	if err != nil {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp["last"]),
			Type:   reflect.TypeOf(int64(0)),
			Offset: int64(len(data)),
			Struct: "RecentTrades",
			Field:  ".",
		}
	}
	trades.Last = ts
	// Convert OHLC data as an array of objects
	trades.Trades = []*Trade{}
	tdata, ok := tmp[trades.PairId].([]interface{})
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp[trades.PairId]),
			Type:   reflect.TypeOf(trades),
			Offset: int64(len(data)),
			Struct: "RecentTrades",
			Field:  ".",
		}
	}
	for _, raw := range tdata {
		// Cast to an array of object
		item, ok := raw.([]interface{})
		if !ok {
			return &json.UnmarshalTypeError{
				Value:  fmt.Sprintf("%v", raw),
				Type:   reflect.TypeOf(Trade{}),
				Offset: int64(len(data)),
				Struct: "Trade",
				Field:  ".",
			}
		}
		parsed, err := unmarshalTradeFromArray(item)
		if err != nil {
			return &json.UnmarshalTypeError{
				Value:  fmt.Sprintf("%v", raw),
				Type:   reflect.TypeOf(Trade{}),
				Offset: int64(len(data)),
				Struct: "Trade",
				Field:  ".",
			}
		}
		trades.Trades = append(trades.Trades, parsed)
	}
	// Exit
	return nil
}

// GetRecentTrades request parameters
type GetRecentTradesRequestParameters struct {
	// Asset pair to get data for.
	Pair string `json:"pair"`
}

// GetRecentTrades request options
type GetRecentTradesRequestOptions struct {
	// Return up to 1000 recent trades since given unix timestamp.
	//
	// By default, return the most recent trades. A zero value triggers default behavior.
	Since int64 `json:"since,omitempty"`
	// Return specific number of trades, up to 1000.
	//
	// 1000 by default. A zero value triggers default behavior.
	Count int `json:"count,omitempty"`
}

// GetRecentTrades Response
type GetRecentTradesResponse struct {
	common.KrakenSpotRESTResponse
	// Recent trades data
	Result *RecentTrades `json:"result,omitempty"`
}
