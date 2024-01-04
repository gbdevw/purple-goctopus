package messages

// Response. Server pong response to a ping to determine whether connection is alive. This is an
// application level pong as opposed to default pong in websockets standard which is sent by client
// in response to a ping.
type Pong struct {
	// Event type
	Event string `json:"event"`
	// Optional - matching client originated request ID
	ReqId *int64 `json:"reqid,omitempty"`
}
