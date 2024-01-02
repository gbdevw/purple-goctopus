package messages

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Used to determine what kind of payload has been parsed
type payloadType string

// Values for payloadType
const (
	// Payload is not a book update
	Other payloadType = "other"
	// Payload is a book update with only asks
	AsksOnly payloadType = "asks"
	// Payload is a book update with only bids
	BidsOnly payloadType = "bids"
	// Payload is a book update with asks and bids
	Mixed payloadType = "mixed"
)

// Static regex which extracts the channel name without trailing -***
var extractChannelNameRegex = regexp.MustCompile(`\"(ticker|ohlc|trade|spread|book)[-,0-9]*\"`)

// Static regex used to matches whitespaces
var matchesWhitespacesRegex = regexp.MustCompile(`\s`)

// Abstract canevas for market updates published by the websocket API (trades, book, ticker, ...).
//
// This canevas can be used when receiving a message from the websocket server to parse its content and
// determine its content (book, trade, ...). Helper methods are then provided to convert the market update
// to its real type (Ticker, Trade, ...).
type MarketData struct {
	// Channel ID of subscription.
	//
	// Deprecated: use channelName and pair
	ChannelId int
	// Name of subscription - Should be "ticker"
	Name string
	// Asset pair
	Pair string
	// Market data
	Data interface{}
}

// Custom JSON unmarshaller for market data
func (m *MarketData) UnmarshalJSON(data []byte) error {
	// 0. Set paylaod type to other
	payloadType := Other
	// 1. Remove whitespaces
	src := string(matchesWhitespacesRegex.ReplaceAll(data, []byte("")))
	// 2. Use a regex to extract the channel name in order to know what kind
	// of data the parser will have to parse. A single match is expected.
	matches := extractChannelNameRegex.FindStringSubmatch(src)
	if len(matches) != 2 {
		return fmt.Errorf("failed to extract channel name from parsed data: %s", src)
	}
	// 2. Prepare an array objects that will be used as target by the unmarshaller
	tmp := []interface{}{
		0.0, // The channel ID is understood as a float by the parser
		nil, // Will be set later, expected type of data
		"",  // Expect a string for channel name
		"",  // Expect a string for pair
	}
	// 3. Depending on the channel name, set the real type of data to parse
	switch matches[1] {
	case string(ChannelTicker):
		// Market data should contain ticker data
		tmp[1] = new(AssetTickerInfo)
	case string(ChannelOHLC):
		// Market data should contain ohlc data
		tmp[1] = new(OHLCData)
	case string(ChannelTrade):
		// Market data should contain trade data
		tmp[1] = &[]TradeData{}
	case string(ChannelSpread):
		// Market data should contain spread data
		tmp[1] = new(SpreadData)
	case string(ChannelBook):
		// Check if parsed data contain a "as" which means we have a book snapshot
		if strings.Contains(string(data), `"as"`) {
			tmp[1] = new(BookSnapshotData)
		} else {
			// We have a book update
			if strings.Contains(string(data), `"a"`) {
				// We have a book update with asks updates
				tmp[1] = new(bookAsksUpdate)
				payloadType = AsksOnly
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
					payloadType = Mixed
				}
			} else {
				// We have a book update with only bids updates
				tmp[1] = new(bookBidsUpdate)
				payloadType = BidsOnly
			}
		}
	default:
		// Unknown channel name
		return fmt.Errorf("failed to parse market data: unknown channel name: %s", matches[1])
	}
	// 4. Unmarshal data into the array of objects
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	// 5. Extract common data depending on the array length
	cid, ok := tmp[0].(float64) // Yes, it is understood like that by the parser
	if !ok {
		return fmt.Errorf("failed to extract channel ID from parsed data: %s", src)
	}
	// Depending on the array length (4 or 5)
	var cname string
	var pair string
	if len(tmp) == 4 {
		// Extract channel name: string - index 2
		cname, ok = tmp[2].(string)
		if !ok {
			return fmt.Errorf("failed to extract channel name from parsed data: %s", src)
		}
		// Extract pair: string - index 3
		pair, ok = tmp[3].(string)
		if !ok {
			return fmt.Errorf("failed to extract pair from parsed data: %s", src)
		}
	} else {
		// Extract channel name: string - index 3
		cname, ok = tmp[3].(string)
		if !ok {
			return fmt.Errorf("failed to extract channel name from parsed data: %s", src)
		}
		// Extract pair: string - index 4
		pair, ok = tmp[4].(string)
		if !ok {
			return fmt.Errorf("failed to extract pair from parsed data: %s", src)
		}
	}
	// 6. Endcode data depending on the payload type
	m.ChannelId = int(cid)
	m.Name = cname
	m.Pair = pair
	switch payloadType {
	case Other:
		// Use the parsed data as is
		m.Data = tmp[1]
	case BidsOnly:
		// Build a BookUpdate with bids as the parsed data
		bids, ok := tmp[1].(*bookBidsUpdate)
		if !ok {
			return fmt.Errorf("failed to extract bids update from parsed data: %s", src)
		}
		m.Data = &BookUpdateData{
			Asks:     []BookMessageEntry{},
			Bids:     bids.Bids,
			Checksum: bids.Checksum,
		}
	case AsksOnly:
		// Build a BookUpdate with asks as the parsed data
		asks, ok := tmp[1].(*bookAsksUpdate)
		if !ok {
			return fmt.Errorf("failed to extract asks update from parsed data: %s", src)
		}
		m.Data = &BookUpdateData{
			Asks:     asks.Asks,
			Bids:     []BookMessageEntry{},
			Checksum: asks.Checksum,
		}
	case Mixed:
		// Build a BookUpdate with both asks and bids as the parsed data
		asks, ok := tmp[1].(*bookAsksUpdate)
		if !ok {
			return fmt.Errorf("failed to extract asks update from parsed data: %s", src)
		}
		bids, ok := tmp[2].(*bookBidsUpdate)
		if !ok {
			return fmt.Errorf("failed to extract bids update from parsed data: %s", src)
		}
		m.Data = &BookUpdateData{
			Asks:     asks.Asks,
			Bids:     bids.Bids,
			Checksum: bids.Checksum, // Checksum will be in bids, the last container of the message
		}
	}
	return nil
}

// Custom JSON marshaller for market data
func (m *MarketData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		m.ChannelId,
		m.Data,
		m.Name,
		m.Pair,
	})
}

// Convert this market data into a ticker.
func (m *MarketData) AsTicker() (*Ticker, error) {
	// Check channel name is ticker
	if m.Name != string(ChannelTicker) {
		return nil, fmt.Errorf("cannot convert to Ticker: market data channel name is not %s: %s", string(ChannelTicker), m.Name)
	}
	ticker, ok := m.Data.(*AssetTickerInfo)
	if !ok {
		return nil, fmt.Errorf("cannot convert market data to Ticker: market data are of type %T", m.Data)
	}
	// Return ticker
	return &Ticker{
		ChannelId: m.ChannelId,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *ticker,
	}, nil
}

// Convert this market data into a ohlc.
func (m *MarketData) AsOHLC() (*OHLC, error) {
	// Check channel name is ohlc
	if !strings.Contains(m.Name, string(ChannelOHLC)) {
		return nil, fmt.Errorf("cannot convert to OHLC: market data channel name does not contain %s: %s", string(ChannelOHLC), m.Name)
	}
	ohlc, ok := m.Data.(*OHLCData)
	if !ok {
		return nil, fmt.Errorf("cannot convert market data to OHLC: market data are of type %T", m.Data)
	}
	// Return ohlc
	return &OHLC{
		ChannelId: m.ChannelId,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *ohlc,
	}, nil
}

// Convert this market data into a trade.
func (m *MarketData) AsTrade() (*Trade, error) {
	// Check channel name is trade
	if !strings.Contains(m.Name, string(ChannelTrade)) {
		return nil, fmt.Errorf("cannot convert to Trade: market data channel name does not contain %s: %s", string(ChannelTrade), m.Name)
	}
	trade, ok := m.Data.(*[]TradeData)
	if !ok {
		return nil, fmt.Errorf("cannot convert market data to Trade: market data are of type %T", m.Data)
	}
	// Return trade
	return &Trade{
		ChannelId: m.ChannelId,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *trade,
	}, nil
}

// Convert this market data into a spread.
func (m *MarketData) AsSpread() (*Spread, error) {
	// Check channel name is spread
	if !strings.Contains(m.Name, string(ChannelSpread)) {
		return nil, fmt.Errorf("cannot convert to Spread: market data channel name does not contain %s: %s", string(ChannelSpread), m.Name)
	}
	spread, ok := m.Data.(*SpreadData)
	if !ok {
		return nil, fmt.Errorf("cannot convert market data to Spread: market data are of type %T", m.Data)
	}
	// Return spread
	return &Spread{
		ChannelId: m.ChannelId,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *spread,
	}, nil
}

// Convert this market data into a book snapshot.
func (m *MarketData) AsBookSnapshot() (*BookSnapshot, error) {
	// Check channel name is book-*
	if !strings.Contains(m.Name, string(ChannelBook)) {
		return nil, fmt.Errorf("cannot convert to BookSnapshot: market data channel name does not contain %s: %s", string(ChannelBook), m.Name)
	}
	book, ok := m.Data.(*BookSnapshotData)
	if !ok {
		return nil, fmt.Errorf("cannot convert market data to Spread: market data are of type %T", m.Data)
	}
	// Return ticker
	return &BookSnapshot{
		ChannelId: m.ChannelId,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *book,
	}, nil
}

// Convert this market data into a book update.
func (m *MarketData) AsBookUpdate() (*BookUpdate, error) {
	// Check channel name is book-*
	if !strings.Contains(m.Name, string(ChannelBook)) {
		return nil, fmt.Errorf("cannot convert to BookUpdate: market data channel name does not contain %s: %s", string(ChannelBook), m.Name)
	}
	book, ok := m.Data.(*BookUpdateData)
	if !ok {
		return nil, fmt.Errorf("cannot convert market data to Spread: market data are of type %T", m.Data)
	}
	// Return ticker
	return &BookUpdate{
		ChannelId: m.ChannelId,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *book,
	}, nil
}
