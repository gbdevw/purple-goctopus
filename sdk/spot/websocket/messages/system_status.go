package messages

// Publication: Status sent on connection or system status changes.
type SystemStatus struct {
	// Event type
	Event string `json:"event"`
	// Optional - Connection ID (will appear only in initial connection status message)
	ConnectionId int64 `json:"connectionID,omitempty"`
	// Status. Cf. StatusEnum for values.
	Status string `json:"status"`
	// API version
	Version string `json:"version"`
}
