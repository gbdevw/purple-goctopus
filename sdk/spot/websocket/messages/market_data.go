package messages

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Static regex which extracts the channel name without trailing -***
var extractChannelNameRegex = regexp.MustCompile(`^\[.*,.*,\"([a-z]*)[-,\"].*,.*\]$`)

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
		tmp[1] = []TradeData{}
	case string(ChannelSpread):
		// Market data should contain spread data
		tmp[1] = new(SpreadData)
	case string(ChannelBook):
		// Check if parsed data contain a "as" which means we have a book snapshot
		if strings.Contains(string(data), "as") {
			tmp[1] = new(BookSnapshotData)
		} else {
			// We have a book update
			tmp[1] = new(BookUpdateData)
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
	// Extract channel ID: float64 - index 0
	cid, ok := tmp[0].(float64) // Yes, it is understood like that by the parser
	if !ok {
		return fmt.Errorf("failed to extract channel ID from parsed data: %s", src)
	}
	// Extract channel name: string - index 2
	cname, ok := tmp[2].(string)
	if !ok {
		return fmt.Errorf("failed to extract channel name from parsed data: %s", src)
	}
	// Extract pair: string - index 3
	pair, ok := tmp[3].(string)
	if !ok {
		return fmt.Errorf("failed to extract pair from parsed data: %s", src)
	}
	// 5. Encode market data struct with parsed and converted data
	m.ChannelId = int(cid)
	m.Data = tmp[1]
	m.Name = cname
	m.Pair = pair
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
func (m *MarketData) AsTicker() (Ticker, error) {
	// Check channel name is ticker
	if m.Name != string(ChannelTicker) {
		return Ticker{}, fmt.Errorf("cannot convert to Ticker: market data channel name is not %s: %s", string(ChannelTicker), m.Name)
	}
	ticker, ok := m.Data.(*AssetTickerInfo)
	if !ok {
		return Ticker{}, fmt.Errorf("cannot convert market data to Ticker: market data are of type %T", m.Data)
	}
	// Return ticker
	return Ticker{
		ChannelId: 0,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *ticker,
	}, nil
}

// Convert this market data into a ohlc.
func (m *MarketData) AsOHLC() (OHLC, error) {
	// Check channel name is ohlc
	if strings.Contains(m.Name, string(ChannelOHLC)) {
		return OHLC{}, fmt.Errorf("cannot convert to OHLC: market data channel name does not contain %s: %s", string(ChannelOHLC), m.Name)
	}
	ohlc, ok := m.Data.(*OHLCData)
	if !ok {
		return OHLC{}, fmt.Errorf("cannot convert market data to OHLC: market data are of type %T", m.Data)
	}
	// Return ticker
	return OHLC{
		ChannelId: 0,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *ohlc,
	}, nil
}

// Convert this market data into a trade.
func (m *MarketData) AsTrade() (Trade, error) {
	// Check channel name is trade
	if strings.Contains(m.Name, string(ChannelTrade)) {
		return Trade{}, fmt.Errorf("cannot convert to Trade: market data channel name does not contain %s: %s", string(ChannelTrade), m.Name)
	}
	trade, ok := m.Data.([]TradeData)
	if !ok {
		return Trade{}, fmt.Errorf("cannot convert market data to Trade: market data are of type %T", m.Data)
	}
	// Return ticker
	return Trade{
		ChannelId: 0,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      trade,
	}, nil
}

// Convert this market data into a spread.
func (m *MarketData) AsSpread() (Spread, error) {
	// Check channel name is spread
	if strings.Contains(m.Name, string(ChannelSpread)) {
		return Spread{}, fmt.Errorf("cannot convert to Spread: market data channel name does not contain %s: %s", string(ChannelSpread), m.Name)
	}
	spread, ok := m.Data.(*SpreadData)
	if !ok {
		return Spread{}, fmt.Errorf("cannot convert market data to Spread: market data are of type %T", m.Data)
	}
	// Return ticker
	return Spread{
		ChannelId: 0,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *spread,
	}, nil
}

// Convert this market data into a book snapshot.
func (m *MarketData) AsBookSnapshot() (BookSnapshot, error) {
	// Check channel name is book-*
	if strings.Contains(m.Name, string(ChannelBook)) {
		return BookSnapshot{}, fmt.Errorf("cannot convert to BookSnapshot: market data channel name does not contain %s: %s", string(ChannelBook), m.Name)
	}
	book, ok := m.Data.(*BookSnapshotData)
	if !ok {
		return BookSnapshot{}, fmt.Errorf("cannot convert market data to Spread: market data are of type %T", m.Data)
	}
	// Return ticker
	return BookSnapshot{
		ChannelId: 0,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *book,
	}, nil
}

// Convert this market data into a book update.
func (m *MarketData) AsBookUpdate() (BookUpdate, error) {
	// Check channel name is book-*
	if strings.Contains(m.Name, string(ChannelBook)) {
		return BookUpdate{}, fmt.Errorf("cannot convert to BookUpdate: market data channel name does not contain %s: %s", string(ChannelBook), m.Name)
	}
	book, ok := m.Data.(*BookUpdateData)
	if !ok {
		return BookUpdate{}, fmt.Errorf("cannot convert market data to Spread: market data are of type %T", m.Data)
	}
	// Return ticker
	return BookUpdate{
		ChannelId: 0,
		Name:      m.Name,
		Pair:      m.Pair,
		Data:      *book,
	}, nil
}
