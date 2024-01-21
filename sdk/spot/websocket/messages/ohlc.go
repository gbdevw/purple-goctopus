package messages

import (
	"encoding/json"
	"fmt"
)

/*************************************************************************************************/
/* OHLC MESSAGE                                                                                  */
/*************************************************************************************************/

// Data of a ohlc message from the websocket API.
type OHLC struct {
	// Channel ID of subscription.
	//
	// Deprecated: use channelName and pair
	ChannelId int
	// Name of subscription - Should be "ohlc-*"
	Name string
	// Asset pair
	Pair string
	// OHLC data
	Data OHLCData
}

// Custom JSON marshaller for OHLC
func (ohlc *OHLC) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		ohlc.ChannelId,
		ohlc.Data,
		ohlc.Name,
		ohlc.Pair,
	})
}

// Custom JSON unmarshaller for OHLC
func (o *OHLC) UnmarshalJSON(data []byte) error {
	// 1. Prepare an array objects that will be used as target by the unmarshaller
	tmp := []interface{}{
		0.0,           // The channel ID is understood as a float by the parser
		new(OHLCData), // OHLC data
		"",            // Expect a string for channel name
		"",            // Expect a string for pair
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
	// 3 Encode OHLC
	o.ChannelId = int(cid)
	o.Name = cname
	o.Pair = pair
	o.Data = *tmp[1].(*OHLCData)
	return nil
}

/*************************************************************************************************/
/* OHLC DATA                                                                                     */
/*************************************************************************************************/

// Data of a single OHLC indicator
type OHLCData struct {
	// Candle last update time, in seconds since epoch (seconds + decimal nanoseconds)
	Start json.Number
	//  End time of interval, in seconds since epoch (seconds + decimal nanoseconds)
	End json.Number
	// Price of the first trade
	Open json.Number
	// Highest trade price
	High json.Number
	// Lowest trade price
	Low json.Number
	// Price of the last trade
	Close json.Number
	// Volume average price
	VolumeAveragePrice json.Number
	// Volume
	Volume json.Number
	// Number of trades used to build the indicator
	TradesCount int64
}

// Marshal a single OHLC indicator as an array of strings to produce the same JSON data as the API.
//
// [string <time>, string <etime>, string <open>, string <high>, string <low>, string <close>, string <vwap>, string <volume>, int <count>]
func (ohlc OHLCData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		ohlc.Start.String(),
		ohlc.End.String(),
		ohlc.Open.String(),
		ohlc.High.String(),
		ohlc.Low.String(),
		ohlc.Close.String(),
		ohlc.VolumeAveragePrice.String(),
		ohlc.Volume.String(),
		ohlc.TradesCount,
	})
}

// Unmarshal a single OHLC indicator from the API raw JSON data.
//
// [int <unixsec>, string <open>, string <high>, string <low>, string <close>, string <vwap>, string <volume>, int <count>]
func (ohlc *OHLCData) UnmarshalJSON(data []byte) error {
	// Create an array of interface with values that will help parser
	// picking the right type.
	tmp := []interface{}{
		"",           // time
		"",           // etime
		"",           // open
		"",           // high
		"",           // low
		"",           // close
		"",           // vwap
		"",           // volume
		float64(0.0), // count - yes, float64 is needed here
	}
	// Unmarshal data into target array
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// Encode OHLC and exit
	ohlc.Start = json.Number(tmp[0].(string))
	ohlc.End = json.Number(tmp[1].(string))
	ohlc.Open = json.Number(tmp[2].(string))
	ohlc.High = json.Number(tmp[3].(string))
	ohlc.Low = json.Number(tmp[4].(string))
	ohlc.Close = json.Number(tmp[5].(string))
	ohlc.VolumeAveragePrice = json.Number(tmp[6].(string))
	ohlc.Volume = json.Number(tmp[7].(string))
	ohlc.TradesCount = int64(tmp[8].(float64))
	return nil
}
