package messages

import (
	"encoding/json"
	"fmt"
)

// Data of a single book entry in a websocket message
type BookMessageEntry struct {
	// Price level
	Price json.Number
	// Price level volume, for updates volume = 0 for level removal/deletion
	Volume json.Number
	// Price level last updated, seconds since epoch (seconds + decimal nanoseconds)
	Timestamp json.Number
	// Optional - "r" in case update is a republished update
	UpdateType string
}

// Unmarshal a single book entry.
func (b *BookMessageEntry) UnmarshalJSON(data []byte) error {
	// Prepare an array of strings to unmarshal raw data.
	//
	// [string<price>, string<volume>, string<timestamp, optional string<updateType>]
	tmp := make([]string, 4)
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return fmt.Errorf("cannot parse data as a book entry: %w. Got %s", err, string(data))
	}
	// Encode struct
	b.Price = json.Number(tmp[0])
	b.Volume = json.Number(tmp[1])
	b.Timestamp = json.Number(tmp[2])
	if len(tmp) == 4 {
		b.UpdateType = tmp[3]
	}
	return nil
}

// Marshal a book entry to get the same JSON payload as the API.
func (b BookMessageEntry) MarshalJSON() ([]byte, error) {
	// Create an array of strings with the base data.
	data := []string{
		b.Price.String(),
		b.Volume.String(),
		b.Timestamp.String(),
	}
	// If an update type is here, add it to the data
	if b.UpdateType != "" {
		data = append(data, b.UpdateType)
	}
	// Marshal the array of strings to get the same payload as the API
	return json.Marshal(data)
}

// Data of a book snapshot message
type BookSnapshotData struct {
	// Ask side of the book
	Asks []BookMessageEntry `json:"as"`
	// Bid side of the book
	Bids []BookMessageEntry `json:"bs"`
}

// book snapshot message from the websocket server.
type BookSnapshot struct {
	// Channel ID of subscription.
	//
	// Deprecated: use channelName and pair
	ChannelId int
	// Name of subscription - Should be "book-*"
	Name string
	// Asset pair
	Pair string
	// Book snapshot data
	Data BookSnapshotData
}

// Custom JSON marshaller for BookSnapshot
func (bs BookSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		bs.ChannelId,
		bs.Data,
		bs.Name,
		bs.Pair,
	})
}

// Data of a book asks update message
type bookAsksUpdate struct {
	// Asks updates
	Asks []BookMessageEntry `json:"a"`
	// Book checksum as a quoted unsigned 32-bit integer
	Checksum string `json:"c,omitempty"`
}

// Data of a book bids update message
type bookBidsUpdate struct {
	// Bids updates
	Bids []BookMessageEntry `json:"b"`
	// Book checksum as a quoted unsigned 32-bit integer
	Checksum string `json:"c,omitempty"`
}

// Data of a book update message
type BookUpdateData struct {
	// Asks updates
	Asks []BookMessageEntry `json:"a"`
	// Bids updates
	Bids []BookMessageEntry `json:"b"`
	// Book checksum as a quoted unsigned 32-bit integer
	Checksum string `json:"c"`
}

// book update message from the websocket server.
type BookUpdate struct {
	// Channel ID of subscription.
	//
	// Deprecated: use channelName and pair
	ChannelId int
	// Name of subscription - Should be "book-*"
	Name string
	// Asset pair
	Pair string
	// Book update data
	Data BookUpdateData
}

// Custom JSON marshaller for BookUpdate
func (bu BookUpdate) MarshalJSON() ([]byte, error) {
	// Create an array of objects with channel id as first item
	target := []interface{}{bu.ChannelId}
	// Depending on the update content, add asks and/or bids
	if len(bu.Data.Asks) > 0 {
		if len(bu.Data.Bids) > 0 {
			// Both asks and bids
			target = append(
				target,
				&bookAsksUpdate{
					Asks: bu.Data.Asks,
				},
				&bookBidsUpdate{
					Bids:     bu.Data.Bids,
					Checksum: bu.Data.Checksum,
				})
		} else {
			// Only asks
			target = append(target, &bookAsksUpdate{
				Asks:     bu.Data.Asks,
				Checksum: bu.Data.Checksum,
			})
		}
	} else {
		// Only bids are expected
		target = append(target, &bookBidsUpdate{
			Bids:     bu.Data.Bids,
			Checksum: bu.Data.Checksum,
		})
	}
	// Append channel name and pair and marshal array
	return json.Marshal(append(target,
		bu.Name,
		bu.Pair,
	))
}
