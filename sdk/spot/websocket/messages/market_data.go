package messages

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// Static initialization of regexes used by the parser
var extractChannelNameRegex = regexp.MustCompile(`^\[.*,.*,\"(.*)\",.*\]$`)
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
	default:
		return fmt.Errorf("failed to parse market data: unknown channel name: %s", matches[0])
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
