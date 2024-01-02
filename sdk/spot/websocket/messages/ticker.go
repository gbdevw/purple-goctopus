package messages

import (
	"encoding/json"
	"fmt"
)

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
	Data AssetTickerInfo
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

// Asset Ticker Info
type AssetTickerInfo struct {
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

// Intermediate struct used to marshal AssetTickerInfo to the same payloads as the API.
type marshalAssetTickerInfo struct {
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

// Custom JSON marshaller for AssetTickerInfo
func (t AssetTickerInfo) MarshalJSON() ([]byte, error) {
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
	return json.Marshal(&marshalAssetTickerInfo{
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

// Get the price of the best ask out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetAskPrice() json.Number {
	return ati.Ask[0]
}

// Get the whole lot volume of the best ask out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetAskWholeLotVolume() json.Number {
	return ati.Ask[1]
}

// Get the lot volume of the best ask out of an AssetTickerInfo
func (ati *AssetTickerInfo) GetAskLotVolume() json.Number {
	return ati.Ask[2]
}

// Get the price of the best bid out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetBidPrice() json.Number {
	return ati.Bid[0]
}

// Get the whole lot volume of the best bid out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetBidWholeLotVolume() json.Number {
	return ati.Bid[1]
}

// Get the lot volume of the best bid out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetBidLotVolume() json.Number {
	return ati.Bid[2]
}

// Get the price of the last trade out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetLastTradePrice() json.Number {
	return ati.Close[0]
}

// Get the lot volume of the last trade out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetLastTradeLotVolume() json.Number {
	return ati.Close[1]
}

// Get today's traded volume out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayVolume() json.Number {
	return ati.Volume[0]
}

// Get past 24h traded volume out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HVolume() json.Number {
	return ati.Volume[1]
}

// Get today's volume average price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayVolumeAveragePrice() json.Number {
	return ati.VolumeAveragePrice[0]
}

// Get past 24h volume average price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HVolumeAveragePrice() json.Number {
	return ati.VolumeAveragePrice[1]
}

// Get today's trade count out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayTradeCount() json.Number {
	return ati.Trades[0]
}

// Get today's trade count out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HTradeCount() json.Number {
	return ati.Trades[1]
}

// Get today's low price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayLow() json.Number {
	return ati.Low[0]
}

// Get past 24h low price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HLow() json.Number {
	return ati.Low[1]
}

// Get today's high price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayHigh() json.Number {
	return ati.High[0]
}

// Get past 24h high price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HHigh() json.Number {
	return ati.High[1]
}

// Get today's opening price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetTodayOpen() json.Number {
	return ati.Open[0]
}

// Get past 24 hours opening price out of this AssetTickerInfo
func (ati *AssetTickerInfo) GetPast24HOpen() json.Number {
	return ati.Open[1]
}
