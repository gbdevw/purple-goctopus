package messages

import (
	"encoding/json"
	"fmt"
	"strings"
)

/*************************************************************************************************/
/* BOOK ENTRY MESSAGE                                                                            */
/*************************************************************************************************/

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

/*************************************************************************************************/
/* BOOK SNAPSHOT MESSAGE                                                                         */
/*************************************************************************************************/

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

// Custom JSON unmarshaller for BookSnapshot
func (bs *BookSnapshot) UnmarshalJSON(data []byte) error {
	// 1. Prepare an array objects that will be used as target by the unmarshaller
	tmp := []interface{}{
		0.0,                   // The channel ID is understood as a float by the parser
		new(BookSnapshotData), // BookSnapshot data
		"",                    // Expect a string for channel name
		"",                    // Expect a string for pair
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
	bs.ChannelId = int(cid)
	bs.Name = cname
	bs.Pair = pair
	bs.Data = *tmp[1].(*BookSnapshotData)
	return nil
}

/*************************************************************************************************/
/* BOOK UPDATE MESSAGE                                                                           */
/*************************************************************************************************/

// Used to determine what kind of content is inside the book update
type bookUpdateContent string

// Values for bookUpdateContent
const (
	// Payload is a book update with only asks
	AsksOnly bookUpdateContent = "asks"
	// Payload is a book update with only bids
	BidsOnly bookUpdateContent = "bids"
	// Payload is a book update with asks and bids
	Mixed bookUpdateContent = "mixed"
)

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

// Custom JSON unmarshaller for BookUpdate
func (bu *BookUpdate) UnmarshalJSON(data []byte) error {

	// 1. Prepare the array used as target to unmarshal data
	tmp := []interface{}{
		0.0, // The channel ID is understood as a float by the parser
		nil, // Will be set later
		"",  // Expect a string for channel name
		"",  // Expect a string for pair
	}
	var typ bookUpdateContent
	if strings.Contains(string(data), `"a"`) {
		// We have a book update with asks updates
		tmp[1] = new(bookAsksUpdate)
		typ = AsksOnly
		if strings.Contains(string(data), `"b"`) {
			// We have a book update with bids updates as well.
			// We have to create a new array to parse data as length differ from all others
			tmp = []interface{}{
				0.0, // The channel ID is understood as a float by the parser
				new(bookAsksUpdate),
				new(bookBidsUpdate),
				"", // Expect a string for channel name
				"", // Expect a string for pair
			}
			typ = Mixed
		}
	} else {
		// We have a book update with only bids updates
		tmp[1] = new(bookBidsUpdate)
		typ = BidsOnly
	}
	// 2. Unmarshal data into the array of objects
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// 3. Extract common data depending on the content type
	cid, ok := tmp[0].(float64) // Yes, it is understood like that by the parser
	if !ok {
		return fmt.Errorf("failed to extract channel ID from parsed data: %s", string(data))
	}
	// Depending on the content type
	var cname string
	var pair string
	if typ != Mixed {
		// Extract channel name: string - index 2
		cname, ok = tmp[2].(string)
		if !ok {
			return fmt.Errorf("failed to extract channel name from parsed data: %s", string(data))
		}
		// Extract pair: string - index 3
		pair, ok = tmp[3].(string)
		if !ok {
			return fmt.Errorf("failed to extract pair from parsed data: %s", string(data))
		}
	} else {
		// Extract channel name: string - index 3
		cname, ok = tmp[3].(string)
		if !ok {
			return fmt.Errorf("failed to extract channel name from parsed data: %s", string(data))
		}
		// Extract pair: string - index 4
		pair, ok = tmp[4].(string)
		if !ok {
			return fmt.Errorf("failed to extract pair from parsed data: %s", string(data))
		}
	}
	// 4. Endcode data depending on the payload type
	bu.ChannelId = int(cid)
	bu.Name = cname
	bu.Pair = pair
	switch typ {
	case BidsOnly:
		// Build a BookUpdate with bids as the parsed data
		bids, ok := tmp[1].(*bookBidsUpdate)
		if !ok {
			return fmt.Errorf("failed to extract bids update from parsed data: %s", string(data))
		}
		bu.Data = BookUpdateData{
			Asks:     []BookMessageEntry{},
			Bids:     bids.Bids,
			Checksum: bids.Checksum,
		}
	case AsksOnly:
		// Build a BookUpdate with asks as the parsed data
		asks, ok := tmp[1].(*bookAsksUpdate)
		if !ok {
			return fmt.Errorf("failed to extract asks update from parsed data: %s", string(data))
		}
		bu.Data = BookUpdateData{
			Asks:     asks.Asks,
			Bids:     []BookMessageEntry{},
			Checksum: asks.Checksum,
		}
	case Mixed:
		// Build a BookUpdate with both asks and bids as the parsed data
		asks, ok := tmp[1].(*bookAsksUpdate)
		if !ok {
			return fmt.Errorf("failed to extract asks update from parsed data: %s", string(data))
		}
		bids, ok := tmp[2].(*bookBidsUpdate)
		if !ok {
			return fmt.Errorf("failed to extract bids update from parsed data: %s", string(data))
		}
		bu.Data = BookUpdateData{
			Asks:     asks.Asks,
			Bids:     bids.Bids,
			Checksum: bids.Checksum, // Checksum will be in bids, the last container of the message
		}
	}
	return nil
}
