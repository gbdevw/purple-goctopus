// This package contains definitions of messages exchanged when interacting
// with Kraken spot websocket API.
package messages

// Enum for the event types supported on Kraken spot websocket API.
type EventTypeEnum string

// Values for EventTypeEnum
const (
	EventTypePing EventTypeEnum = "ping"
)
