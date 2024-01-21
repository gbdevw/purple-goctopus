package messages

// Publication: Server heartbeat sent if no subscription traffic within 1 second (approximately)
type Heartbeat struct {
	// Event type
	Event string `json:"event"`
}
