package messages

// Request. Client can ping server to determine whether connection is alive, server responds with
// pong. This is an application level ping as opposed to default ping in websockets standard which
// is server initiated.
type Ping struct {
	// Event type
	Event string `json:"event"`
	// Optional - client originated ID reflected in response message.
	//
	// A zero value means no user ID will be provided.
	ReqId int64 `json:"reqid,omitempty"`
}
