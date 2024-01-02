package messages

import (
	"encoding/json"
	"fmt"
)

/*************************************************************************************************/
/* SPREAD MESSAGE                                                                                */
/*************************************************************************************************/

// Data of a spread message from the websocket API.
type Spread struct {
	// Channel ID of subscription.
	//
	// Deprecated: use channelName and pair
	ChannelId int
	// Name of subscription - Should be "spread"
	Name string
	// Asset pair
	Pair string
	// Spread
	Data SpreadData
}

// Custom JSON marshaller for Spread
func (s Spread) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		s.ChannelId,
		s.Data,
		s.Name,
		s.Pair,
	})
}

// Custom JSON unmarshaller for Spread
func (s *Spread) UnmarshalJSON(data []byte) error {
	// 1. Prepare an array objects that will be used as target by the unmarshaller
	tmp := []interface{}{
		0.0,             // The channel ID is understood as a float by the parser
		new(SpreadData), // Spread data
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
	// 3 Encode Spread
	s.ChannelId = int(cid)
	s.Name = cname
	s.Pair = pair
	s.Data = *tmp[1].(*SpreadData)
	return nil
}

/*************************************************************************************************/
/* SPREAD DATA                                                                                   */
/*************************************************************************************************/

// Data of a spread
type SpreadData struct {
	// Best bid price
	BestBidPrice json.Number
	// Best ask price
	BestAskPrice json.Number
	// Time, seconds since epoch (seconds + decimal nanoseconds)
	Timestamp json.Number
	// Best bid volume
	BestBidVolume json.Number
	// Best ask volume
	BestAskVolume json.Number
}

// Marshal a spread as an array of strings to produce the same JSON data as the API.
//
// [string <bid>, string <ask>, string <time>, string <bidVolume>, string <askVolume>]
func (spread SpreadData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{
		spread.BestBidPrice.String(),
		spread.BestAskPrice.String(),
		spread.Timestamp.String(),
		spread.BestBidVolume.String(),
		spread.BestAskVolume.String(),
	})
}

// Unmarshal a spread from the API raw JSON data.
//
// [string <bid>, string <ask>, string <time>, string <bidVolume>, string <askVolume>]
func (spread *SpreadData) UnmarshalJSON(data []byte) error {
	// Create an array of strings to unmarshal data
	tmp := make([]string, 5)
	// Unmarshal data into target array
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// Encode spread and exit
	spread.BestBidPrice = json.Number(tmp[0])
	spread.BestAskPrice = json.Number(tmp[1])
	spread.Timestamp = json.Number(tmp[2])
	spread.BestBidVolume = json.Number(tmp[3])
	spread.BestAskVolume = json.Number(tmp[4])
	return nil
}
