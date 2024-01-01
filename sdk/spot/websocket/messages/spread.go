package messages

import (
	"encoding/json"
)

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
func (spread *SpreadData) MarshalJSON() ([]byte, error) {
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
