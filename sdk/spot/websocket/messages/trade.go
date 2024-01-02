package messages

import (
	"encoding/json"
	"fmt"
)

/*************************************************************************************************/
/* TRADE MESSAGE                                                                                 */
/*************************************************************************************************/

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

// Custom JSON marshaller for Trade
func (t Trade) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		t.ChannelId,
		t.Data,
		t.Name,
		t.Pair,
	})
}

// Custom JSON unmarshaller for Trade
func (t *Trade) UnmarshalJSON(data []byte) error {
	// 1. Prepare an array objects that will be used as target by the unmarshaller
	tmp := []interface{}{
		0.0,              // The channel ID is understood as a float by the parser
		new([]TradeData), // Trade data
		"",               // Expect a string for channel name
		"",               // Expect a string for pair
	}
	// 2. Unmarshal data into the target array of objects
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// 3. Extract data
	// Extract channel ID: index 0
	cid, ok := tmp[0].(float64) // Yes, it is understood like that by the parser
	if !ok {
		return fmt.Errorf("failed to extract channel ID from parsed data: %s", string(data))
	}
	// Extract channel name: string - index 2
	cname, ok := tmp[2].(string)
	if !ok {
		return fmt.Errorf("failed to extract channel name from parsed data: %s", string(data))
	}
	// Extract pair: string - index 3
	pair, ok := tmp[3].(string)
	if !ok {
		return fmt.Errorf("failed to extract pair from parsed data: %s", string(data))
	}
	// 3 Encode trade
	t.ChannelId = int(cid)
	t.Name = cname
	t.Pair = pair
	t.Data = *tmp[1].(*[]TradeData)
	return nil
}

/*************************************************************************************************/
/* TRADE DATA                                                                                    */
/*************************************************************************************************/

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
func (trade TradeData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{
		trade.Price.String(),
		trade.Volume.String(),
		trade.Timestamp.String(),
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
