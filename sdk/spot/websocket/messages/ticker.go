package messages

import (
	"encoding/json"
	"fmt"
)

/*************************************************************************************************/
/* TICKER MESSAGE                                                                                */
/*************************************************************************************************/

// Data of a ticker message from the websocket API.
type Ticker struct {
	// Channel ID of subscription.
	//
	// Deprecated: use channelName and pair
	ChannelId int
	// Name of subscription - Should be "ticker"
	Name string
	// Asset pair
	Pair string
	// Ticker data
	Data TickerData
}

// Custom JSON marshaller for Ticker
func (t Ticker) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		t.ChannelId,
		t.Data,
		t.Name,
		t.Pair,
	})
}

// Custom JSON unmarshaller for Ticker
func (t *Ticker) UnmarshalJSON(data []byte) error {
	// 1. Prepare an array objects that will be used as target by the unmarshaller
	tmp := []interface{}{
		0.0,             // The channel ID is understood as a float by the parser
		new(TickerData), // Ticker
		"",              // Expect a string for channel name
		"",              // Expect a string for pair
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
	// 3 Encode ticker
	t.ChannelId = int(cid)
	t.Name = cname
	t.Pair = pair
	t.Data = *tmp[1].(*TickerData)
	return nil
}

/*************************************************************************************************/
/* TICKER DATA                                                                                    */
/*************************************************************************************************/

// Ticker data
type TickerData struct {
	// Ask array(<price>, <whole lot volume>, <lot volume>)
	Ask []json.Number `json:"a"`
	// Bid array(<price>, <whole lot volume>, <lot volume>)
	Bid []json.Number `json:"b"`
	// Last trade closed array(<price>, <lot volume>)
	Close []json.Number `json:"c"`
	// Volume array(<today>, <last 24 hours>)
	Volume []json.Number `json:"v"`
	// Volume weighted average price array(<today>, <last 24 hours>)
	VolumeAveragePrice []json.Number `json:"p"`
	// Number of trades array(<today>, <last 24 hours>)
	Trades []json.Number `json:"t"`
	// Low array(<today>, <last 24 hours>)
	Low []json.Number `json:"l"`
	// High array(<today>, <last 24 hours>)
	High []json.Number `json:"h"`
	// Open array(<today>, <last 24 hours>)
	Open []json.Number `json:"o"`
}

// Intermediate struct used to marshal TickerData to the same payloads as the API.
type marshalTickerData struct {
	// Ask array(<price>, <whole lot volume>, <lot volume>)
	Ask []interface{} `json:"a"`
	// Bid array(<price>, <whole lot volume>, <lot volume>)
	Bid []interface{} `json:"b"`
	// Last trade closed array(<price>, <lot volume>)
	Close []interface{} `json:"c"`
	// Volume array(<today>, <last 24 hours>)
	Volume []interface{} `json:"v"`
	// Volume weighted average price array(<today>, <last 24 hours>)
	VolumeAveragePrice []interface{} `json:"p"`
	// Number of trades array(<today>, <last 24 hours>)
	Trades []interface{} `json:"t"`
	// Low array(<today>, <last 24 hours>)
	Low []interface{} `json:"l"`
	// High array(<today>, <last 24 hours>)
	High []interface{} `json:"h"`
	// Open array(<today>, <last 24 hours>)
	Open []interface{} `json:"o"`
}

// Custom JSON marshaller for TickerData
func (t TickerData) MarshalJSON() ([]byte, error) {
	awl, err := t.GetAskWholeLotVolume().Int64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert AskWholeLotVolume to int64: %w", err)
	}
	bwl, err := t.GetBidWholeLotVolume().Int64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert BidWholeLotVolume to int64: %w", err)
	}
	ttc, err := t.GetTodayTradeCount().Int64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert TodayTradeCount to int64: %w", err)
	}
	t24c, err := t.GetPast24HTradeCount().Int64()
	if err != nil {
		return nil, fmt.Errorf("failed to convert Past24HTradeCount to int64: %w", err)
	}
	return json.Marshal(&marshalTickerData{
		Ask:                []interface{}{t.GetAskPrice().String(), awl, t.GetAskLotVolume().String()},
		Bid:                []interface{}{t.GetBidPrice().String(), bwl, t.GetBidLotVolume().String()},
		Close:              []interface{}{t.GetLastTradePrice().String(), t.GetLastTradeLotVolume().String()},
		Volume:             []interface{}{t.GetTodayVolume().String(), t.GetPast24HVolume().String()},
		VolumeAveragePrice: []interface{}{t.GetTodayVolumeAveragePrice().String(), t.GetPast24HVolumeAveragePrice().String()},
		Trades:             []interface{}{ttc, t24c},
		Low:                []interface{}{t.GetTodayLow().String(), t.GetPast24HLow().String()},
		High:               []interface{}{t.GetTodayHigh().String(), t.GetPast24HHigh().String()},
		Open:               []interface{}{t.GetTodayOpen().String(), t.GetPast24HOpen().String()},
	})
}

/*************************************************************************************************/
/* HELPER METHODS                                                                                */
/*************************************************************************************************/

// Get the price of the best ask out of this TickerData
func (ati *TickerData) GetAskPrice() json.Number {
	return ati.Ask[0]
}

// Get the whole lot volume of the best ask out of this TickerData
func (ati *TickerData) GetAskWholeLotVolume() json.Number {
	return ati.Ask[1]
}

// Get the lot volume of the best ask out of an TickerData
func (ati *TickerData) GetAskLotVolume() json.Number {
	return ati.Ask[2]
}

// Get the price of the best bid out of this TickerData
func (ati *TickerData) GetBidPrice() json.Number {
	return ati.Bid[0]
}

// Get the whole lot volume of the best bid out of this TickerData
func (ati *TickerData) GetBidWholeLotVolume() json.Number {
	return ati.Bid[1]
}

// Get the lot volume of the best bid out of this TickerData
func (ati *TickerData) GetBidLotVolume() json.Number {
	return ati.Bid[2]
}

// Get the price of the last trade out of this TickerData
func (ati *TickerData) GetLastTradePrice() json.Number {
	return ati.Close[0]
}

// Get the lot volume of the last trade out of this TickerData
func (ati *TickerData) GetLastTradeLotVolume() json.Number {
	return ati.Close[1]
}

// Get today's traded volume out of this TickerData
func (ati *TickerData) GetTodayVolume() json.Number {
	return ati.Volume[0]
}

// Get past 24h traded volume out of this TickerData
func (ati *TickerData) GetPast24HVolume() json.Number {
	return ati.Volume[1]
}

// Get today's volume average price out of this TickerData
func (ati *TickerData) GetTodayVolumeAveragePrice() json.Number {
	return ati.VolumeAveragePrice[0]
}

// Get past 24h volume average price out of this TickerData
func (ati *TickerData) GetPast24HVolumeAveragePrice() json.Number {
	return ati.VolumeAveragePrice[1]
}

// Get today's trade count out of this TickerData
func (ati *TickerData) GetTodayTradeCount() json.Number {
	return ati.Trades[0]
}

// Get today's trade count out of this TickerData
func (ati *TickerData) GetPast24HTradeCount() json.Number {
	return ati.Trades[1]
}

// Get today's low price out of this TickerData
func (ati *TickerData) GetTodayLow() json.Number {
	return ati.Low[0]
}

// Get past 24h low price out of this TickerData
func (ati *TickerData) GetPast24HLow() json.Number {
	return ati.Low[1]
}

// Get today's high price out of this TickerData
func (ati *TickerData) GetTodayHigh() json.Number {
	return ati.High[0]
}

// Get past 24h high price out of this TickerData
func (ati *TickerData) GetPast24HHigh() json.Number {
	return ati.High[1]
}

// Get today's opening price out of this TickerData
func (ati *TickerData) GetTodayOpen() json.Number {
	return ati.Open[0]
}

// Get past 24 hours opening price out of this TickerData
func (ati *TickerData) GetPast24HOpen() json.Number {
	return ati.Open[1]
}
