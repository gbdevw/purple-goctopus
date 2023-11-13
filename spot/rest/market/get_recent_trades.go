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
	return json.Marshal([]interface{}{
		trade.Price,
		trade.Volume,
		trade.Timestamp.Unix(),
		trade.Side,
		trade.Type,
		trade.Miscellaneous,
		trade.Id,
	})
}

// Unmarshal trade data from an array of data from the API.
//
// [<price>, <volume>, <time>, <buy/sell>, <market/limit>, <miscellaneous>, <trade_id>]
func unmarshalTradeFromArray(input []interface{}) (Trade, error) {
	// Convert timestamp to string
	tsstr, ok := input[3].(string)
	if !ok {
		return Trade{}, fmt.Errorf("could not parse timestamp as text. Got %v", input[3])
	}
	// Split sec & nsec + parse each part as int64
	tssplits := strings.Split(tsstr, ".")
	if len(tssplits) != 2 {
		return Trade{}, fmt.Errorf("could not split timestamp seconds and nanosec parts. Got %v", tssplits)
	}
	sec, err := strconv.ParseInt(tssplits[0], 10, 64)
	if err != nil {
		return Trade{}, fmt.Errorf("could not parse timestamp.seconds as int64: %w", err)
	}
	nsec, err := strconv.ParseInt(tssplits[1], 10, 64)
	if err != nil {
		return Trade{}, fmt.Errorf("could not parse timestamp.nanoseconds as int64: %w", err)
	}
	// Convert Id to int64
	id, ok := input[6].(int64)
	if !ok {
		return Trade{}, fmt.Errorf("could not parse trade id as int64: %w", err)
	}
	// Convert other items to string
	price, ok := input[0].(string)
	if !ok {
		return Trade{}, fmt.Errorf("could not parse trade price as text. Got %v", input[0])
	}
	volume, ok := input[1].(string)
	if !ok {
		return Trade{}, fmt.Errorf("could not parse trade volume as text. Got %v", input[1])
	}
	side, ok := input[3].(string)
	if !ok {
		return Trade{}, fmt.Errorf("could not parse trade side as text. Got %v", input[3])
	}
	typ, ok := input[4].(string)
	if !ok {
		return Trade{}, fmt.Errorf("could not parse trade type as text. Got %v", input[4])
	}
	misc, ok := input[5].(string)
	if !ok {
		return Trade{}, fmt.Errorf("could not parse trade miscellaneous as text. Got %v", input[5])
	}
	// Build and return trade data
	return Trade{
		Price:         price,
		Volume:        volume,
		Timestamp:     time.Unix(sec, nsec),
		Side:          side,
		Type:          typ,
		Miscellaneous: misc,
		Id:            id,
	}, nil
}

// Recent trade data last + asset > [[<price>, <volume>, <time>, <buy/sell>, <market/limit>, <miscellaneous>, <trade_id>]]
type RecentTrades struct {
	// Asset pair ID
	PairId string
	// Timestamp to be used as since to fetch next trades data
	Last time.Time
	// Trades
	Trades []Trade
}

// Marshal trades data to produce the same raw data as the API.
func (trades *RecentTrades) MarshalJSON() ([]byte, error) {
	// Put data into a map
	base := map[string]interface{}{
		trades.PairId: trades.Trades,
		"last":        trades.Last.Unix(),
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
	// Cast last as int64
	ts, ok := tmp["last"].(int64)
	if !ok {
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprintf("%v", tmp["last"]),
			Type:   reflect.TypeOf(trades),
			Offset: int64(len(data)),
			Struct: "RecentTrades",
			Field:  ".",
		}
	}
	trades.Last = time.Unix(ts, 0)
	// Convert OHLC data as array of arrays
	trades.Trades = []Trade{}
	tdata := tmp[trades.PairId].([][]interface{})
	for _, raw := range tdata {
		parsed, err := unmarshalTradeFromArray(raw)
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

// GetRecentTrades required parameters
type GetRecentTradesParameters struct {
	// Asset pair to get data for.
	Pair string
}

// GetRecentTrades options
type GetRecentTradesOptions struct {
	// Return up to 1000 recent trades since given timestamp
	//
	// By default, return the most recent trades. A zero value triggers default behavior.
	Since time.Time
}

// GetRecentTrades Response
type GetRecentTradesResponse struct {
	common.KrakenSpotRESTResponse
	// Recent trades data
	Result *RecentTrades `json:"result,omitempty"`
}
