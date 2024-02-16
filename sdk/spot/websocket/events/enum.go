package events

// Enum for event types used by the websocket client
type WebsocketClientEventTypeEnum string

const (
	// Event type used by events produced to warn consumers that the stream of data has been
	// interrupted because the connection with the websocket server has been interrupted. This will
	// be used as a cue for consumer to react to interruptions in the stream of data.
	ConnectionInterrupted WebsocketClientEventTypeEnum = "connection_interrupted"
	// Event typed used when a new heartbeat is received from the server
	Heartbeat WebsocketClientEventTypeEnum = "heartbeat"
	// Event type used when a new system status is received from the server.
	SystemStatus WebsocketClientEventTypeEnum = "system_status"
	// Event type when new message is received on the own trades channel.
	OwnTrades WebsocketClientEventTypeEnum = "own_trades"
	// Event type used when a new message is received on the open orders channel.
	OpenOrders WebsocketClientEventTypeEnum = "open_orders"
	// Event type used when a new message is received on the tickers channel.
	Ticker WebsocketClientEventTypeEnum = "ticker"
	// Event type used when a new message is received on a ohlc channel.
	OHLC WebsocketClientEventTypeEnum = "ohlc"
	// Event type used when a new message is received on the trades channel.
	Trade WebsocketClientEventTypeEnum = "trade"
	// Event type used when a new message is received on the spreads channel.
	Spread WebsocketClientEventTypeEnum = "spread"
	// Event type used when a new message is received on the book channel (snapshot).
	BookSnapshot WebsocketClientEventTypeEnum = "book_snapshot"
	// Event type used when a new message is received on the book channel (update).
	BookUpdate WebsocketClientEventTypeEnum = "book_update"
)
