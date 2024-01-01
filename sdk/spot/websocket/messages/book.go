package messages

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

/*************************************************************************************************/
/* MESSAGES                                                                                      */
/*************************************************************************************************/

// Data of a single book entry
type BookEntry struct {
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
func (b *BookEntry) UnmarshalJSON(data []byte) error {
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
	b.UpdateType = tmp[3]
	return nil
}

// Marshal a book entry to get the same JSON payload as the API.
func (b *BookEntry) MarshalJSON() ([]byte, error) {
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
	Asks []BookEntry `json:"as"`
	// Bid side of the book
	Bids []BookEntry `json:"bs"`
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

// Data of a book update message
type BookUpdateData struct {
	// Asks updates
	Asks []BookEntry `json:"a"`
	// Bids updates
	Bids []BookEntry `json:"b"`
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

/*************************************************************************************************/
/* BOOK                                                                                          */
/*************************************************************************************************/

// Static regex used to check volume is 0
var checkVolumeIs0Regex = regexp.MustCompile(`^(0|0\.0*)$`)

// Static regex used to capture the input to put in the string used to calculate checksum
// Input strings MUST NOT contain dots (remove decimal first)
var captureChecksumInput = regexp.MustCompile(`^0*([1-9][0-9]*)$`)

// Order book.
//
// This class provides methods to keep the book updated with the messages received from the API and
// verify the book state with the updates checksums.
type Book struct {
	// Pair related to the order book.
	//
	// Used to ensure updates relate to the base snapshot.
	Pair string
	// Channel name from where source data shoud come from.
	//
	// Used to ensure updates relate to the base snapshot.
	SourceChannelName string
	// Ask side of the book by price level
	Asks map[string]BookEntry `json:"as"`
	// Bid side of the book by price level
	Bids map[string]BookEntry `json:"bs"`
}

// Factory which initialize a book from a snapshot message
func NewBookFromSnapshot(snapshot *BookSnapshot) *Book {
	// Build a new book
	book := &Book{
		Pair:              snapshot.Pair,
		SourceChannelName: snapshot.Name,
		Asks:              map[string]BookEntry{},
		Bids:              map[string]BookEntry{},
	}
	// Encode book with snapshot data
	for _, ask := range snapshot.Data.Asks {
		book.Asks[ask.Price.String()] = ask
	}
	for _, bid := range snapshot.Data.Bids {
		book.Bids[bid.Price.String()] = bid
	}
	// Return book
	return book
}

// # Description
//
// Apply the book update to the book.
//
// As an option, the method can verify the update data after applying the update to this book. In
// case one of the verifications would fail, an error will be returned by the method. In this case,
// recommendation is to resubscribe to the book feed and create a new book from the snapshot that
// will be sent by the websocket server.
//
// If verifications are enabled, the method will:
//   - Check whether rhe channel name and pair of the update match the ones from the base snapshot.
//   - Check the book checksum matches the checksum in the provided update.
//
// If verifications are disabled, the method will apply the update without verifying the update data.
// Therefore, the method will never return an error and returned value can be safely ignored.
//
// # Inputs
//
//   - update: Received update message to apply to this book
//   - verify: If true, verifications will be enabled.
//
// # Return
//
// The method returns an error only if verifications are enabled and if one of the
// verifications fail.
func (book *Book) ApplyBookUpdate(update *BookUpdate, verify bool) error {
	// Apply update
	for _, ask := range update.Data.Asks {
		if checkVolumeIs0Regex.MatchString(string(ask.Volume.String())) {
			// Remove entry from map as volume is 0
			delete(book.Asks, ask.Price.String())
		} else {
			// Update entry
			book.Asks[ask.Price.String()] = ask
		}
	}
	for _, bid := range update.Data.Bids {
		if checkVolumeIs0Regex.MatchString(string(bid.Volume.String())) {
			// Remove entry from map as volume is 0
			delete(book.Bids, bid.Price.String())
		} else {
			// Update entry
			book.Bids[bid.Price.String()] = bid
		}
	}
	// Verify if required to do so
	if verify {
		// Check channel names match
		if update.Name != book.SourceChannelName {
			return fmt.Errorf("book channel name %s differ from the update channel name. Got %s", book.SourceChannelName, update.Name)
		}
		// Check pairs match
		if update.Name != book.SourceChannelName {
			return fmt.Errorf("book pair %s differ from the update pair. Got %s", book.Pair, update.Pair)
		}
		// Checksum
		err := book.checksum(update.Data.Checksum)
		if err != nil {
			return fmt.Errorf("failed to verify the book checksum after applying the update: %w", err)
		}
	}
	return nil
}

// Helper method which verify the order book state matches
// the provided checksum.
func (book *Book) checksum(checksum string) error {
	// 0. Parse privded checksum as uint32
	crc32targetuint64, err := strconv.ParseUint(checksum, 10, 32)
	if err != nil {
		return fmt.Errorf("failed to parse provided checksum as uint32: %w", err)
	}
	crc32target := uint32(crc32targetuint64)
	// 1. Compute the top 10 asks sorted in ascending order (low to high)
	asks := []priceLevel{}
	for key, entry := range book.Asks {
		// Parse price level as a big.Float
		f, _, err := big.ParseFloat(key, 10, uint(big.Exact), big.AwayFromZero)
		if err != nil {
			return fmt.Errorf("failed to parse price level: %w", err)
		}
		// Append to asks
		asks = append(asks, priceLevel{
			Entry:   entry,
			Numeric: *f,
		})
	}
	// Sort asks in ascending price level order
	slices.SortFunc(asks, func(a, b priceLevel) int {
		return a.Numeric.Cmp(&b.Numeric)
	})
	if len(asks) > 10 {
		// Only keep the 10 first entries
		asks = asks[0:10]
	}
	// 2. For each ask in the top 10
	concat := ""
	volume := ""
	for _, ask := range asks {
		// Capture price and volume after removing the . from the price/volume
		// Captured price will be directly added to the concat string
		// Captured volume will be sotred aside and added later
		concat = concat + captureChecksumInput.FindString(strings.ReplaceAll(ask.Entry.Price.String(), ".", ""))
		volume = volume + captureChecksumInput.FindString(strings.ReplaceAll(ask.Entry.Volume.String(), ".", ""))
	}
	concat = concat + volume
	volume = "" // Reset volume
	// 3. Compute the top 10 bids in descending order (high to low)
	bids := []priceLevel{}
	for key, entry := range book.Bids {
		// Parse price level as a big.Float
		f, _, err := big.ParseFloat(key, 10, uint(big.Exact), big.AwayFromZero)
		if err != nil {
			return fmt.Errorf("failed to parse price level: %w", err)
		}
		// Append to bids
		bids = append(bids, priceLevel{
			Entry:   entry,
			Numeric: *f,
		})
	}
	// Sort bids in descending price level order
	slices.SortFunc(bids, func(a, b priceLevel) int {
		return a.Numeric.Cmp(&b.Numeric) * -1
	})
	if len(bids) > 10 {
		// Only keep the 10 first entries
		bids = bids[0:10]
	}
	// 4. For each bid in the top 10
	for _, bid := range bids {
		// Capture price and volume after removing the . from the price/volume
		// Captured price will be directly added to the concat string
		// Captured volume will be sotred aside and added later
		concat = concat + captureChecksumInput.FindString(strings.ReplaceAll(bid.Entry.Price.String(), ".", ""))
		volume = volume + captureChecksumInput.FindString(strings.ReplaceAll(bid.Entry.Volume.String(), ".", ""))
	}
	concat = concat + volume
	// 5. Compute checksum and compare
	actual := crc32.ChecksumIEEE([]byte(concat))
	if actual != crc32target {
		return fmt.Errorf("crc32 checksum for input string: '%s' differs from provided checksum %s", concat, checksum)
	}
	// Success
	return nil
}

// Price level data used to when checking the checksum
type priceLevel struct {
	// Price level as numeric value
	Numeric big.Float
	// Corresponding book entry
	Entry BookEntry
}
