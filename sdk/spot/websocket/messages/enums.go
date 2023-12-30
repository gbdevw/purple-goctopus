// This package contains definitions of messages exchanged when interacting
// with Kraken spot websocket API.
package messages

// Enum for the event types supported on Kraken spot websocket API.
type EventTypeEnum string

// Values for EventTypeEnum
const (
	EventTypePing         EventTypeEnum = "ping"
	EventTypePong         EventTypeEnum = "pong"
	EventTypeHeartbeat    EventTypeEnum = "heartbeat"
	EventTypeSystemStatus EventTypeEnum = "systemStatus"
)

// Enum for the API statuses
type StatusEnum string

// Values for StatusEnum
const (
	StatusOnline      StatusEnum = "online"
	StatusMaintenance StatusEnum = "maintenance"
	StatusCancelOnly  StatusEnum = "cancel_only"
	StatusLimitOnly   StatusEnum = "limit_only"
	StatusPostOnly    StatusEnum = "post_only"
)
