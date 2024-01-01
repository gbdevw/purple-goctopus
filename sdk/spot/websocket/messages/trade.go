package messages

import (
	"encoding/json"
)

// Data of a trade message from the websocket API.
type Trade struct {
	// Channel ID of subscription.
	//
	// Deprecated: use channelName and pair
	ChannelId int
	// Name of subscription - Should be "trade"
	Name string
	// Asset pair
	Pair string
	// Trades
	Data []TradeData
}

// Data of a single trade
type TradeData struct {
	// Price
	Price json.Number
	// Volume
	Volume json.Number
	// Time, seconds since epoch (seconds + decimal nanoseconds)
	Timestamp json.Number
	// Triggering order side, buy/sell
	Side string
	// Triggering order type market/limit
	OrderType string
	// Miscellaneous
	Miscellaneous string
}

// Marshal a single trade as an array of strings to produce the same JSON data as the API.
//
// [string <price>, string <volume>, string <time>, string <side>, string <orderType>, string <misc>]
func (trade *TradeData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{
		trade.Price.String(),
		trade.Volume.String(),
		trade.Price.String(),
		trade.Side,
		trade.OrderType,
		trade.Miscellaneous,
	})
}

// Unmarshal a single trade from the API raw JSON data.
//
// [string <price>, string <volume>, string <time>, string <side>, string <orderType>, string <misc>]
func (trade *TradeData) UnmarshalJSON(data []byte) error {
	// Create an array of strings to unmarshal data
	tmp := make([]string, 6)
	// Unmarshal data into target array
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// Encode trade and exit
	trade.Price = json.Number(tmp[0])
	trade.Volume = json.Number(tmp[1])
	trade.Timestamp = json.Number(tmp[2])
	trade.Side = tmp[3]
	trade.OrderType = tmp[4]
	trade.Miscellaneous = tmp[5]
	return nil
}
