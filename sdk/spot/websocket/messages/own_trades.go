package messages

import (
	"encoding/json"
	"fmt"
)

// Data of a ownTrades message from the websocket server.
type OwnTrades struct {
	// Channel name. Should be "ownTrades"
	ChannelName string
	// Sequence number used to verify no message was lost
	SequenceId SequenceId
	// Message data which contains a collection of maps where keys are trades IDs and values the related trades.
	Data []map[string]OwnTradeData
}

// Custom JSON marshaller which produces the same JSON payloads as the API.
func (owt OwnTrades) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		owt.Data,
		owt.ChannelName,
		owt.SequenceId,
	})
}

// Custom JSON unmarshaller to parse ownTrades messages from the websocket server
func (owt *OwnTrades) UnmarshalJSON(data []byte) error {
	// Prepare an array of objects to parse the payload
	tmp := []interface{}{
		&[]map[string]OwnTradeData{}, // Trades
		"",                           // Channel name
		&SequenceId{},                // Sequence Id object
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return fmt.Errorf("failed to parse data as OwnTrades: %w", err)
	}
	// Encode target
	owt.ChannelName = tmp[1].(string)
	owt.Data = *tmp[0].(*[]map[string]OwnTradeData)
	owt.SequenceId = *tmp[2].(*SequenceId)
	return nil
}

// Data of a single trade
type OwnTradeData struct {
	// Order responsible for execution of trade
	OrderTransactionId string `json:"ordertxid"`
	// Position responsible for execution of trade
	PositionId string `json:"postxid,omitempty"`
	// Asset pair
	Pair string `json:"pair"`
	// Unix timestamp for the trade - As <sec>.<nsec> decimal
	Timestamp string `json:"time"`
	// Trade direction (buy/sell). Cf. SideEnum for values
	Type string `json:"type"`
	// Order type. Cf. OrderTypeEnum for values
	OrderType string `json:"ordertype"`
	// Average price order was executed at
	Price string `json:"price"`
	// Total cost of order
	Cost string `json:"cost,omitempty"`
	// Total fee
	Fee string `json:"fee"`
	// Volume
	Volume string `json:"vol"`
	// Initial margin
	Margin string `json:"margin,omitempty"`
	// Optional user reference ID
	UserReference *int64 `json:"userref,omitempty"`
}
