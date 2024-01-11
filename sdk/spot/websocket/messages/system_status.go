package messages

import "encoding/json"

// Enum for the trading engine statuses
type EngineStatusEnum string

// Values for StatusEnum
const (
	StatusOnline      EngineStatusEnum = "online"
	StatusMaintenance EngineStatusEnum = "maintenance"
	StatusCancelOnly  EngineStatusEnum = "cancel_only"
	StatusLimitOnly   EngineStatusEnum = "limit_only"
	StatusPostOnly    EngineStatusEnum = "post_only"
)

// Publication: Status sent on connection or system status changes.
type SystemStatus struct {
	// Event type
	Event string `json:"event"`
	// Optional - Connection ID (will appear only in initial connection status message)
	ConnectionId json.Number `json:"connectionID,omitempty"`
	// Status. Cf. EngineStatusEnum for values.
	Status string `json:"status"`
	// API version
	Version string `json:"version"`
}
