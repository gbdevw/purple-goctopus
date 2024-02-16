// This package provides interfaces and implementations for websocket clients
// using Kraken spot websocket API (both public and private environments)
package websocket

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
)

// Interface for a websocket client using the public environment for Kraken spot websocket API.
//
// Public websocket client has access to:
//   - Ping
//   - Ticker feed
//   - OHLC feed
//   - Trades feed
//   - Spreads feed
//   - Order book feed
type KrakenSpotPublicWebsocketClientInterface interface {
	// # Description
	//
	// Send a ping to the websocket server and wait until a Pong response is received from the
	// server or until an error or a timeout occurs.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//
	// # Return
	//
	// Nil in case of success. Otherwise, an error is returned when:
	//
	//	- An error occurs when sending the message.
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server (OperationError).
	Ping(ctx context.Context) error
	// # Description
	//
	// Subscribe to the tickers channel. In case of success, the websocket client will start
	// publishing received events on the user's provided channel.
	//
	// Two types of events can be published on the channel:
	//	- connection_interrupted: This event type is used when connection with the sevrer has been
	//    interrupted. The event will not have any data. It only serves as a cue for the consumer
	//    to allow the consumer to react when the connection with the server is interrupted.
	//	- ticker: This event type is used when a message has been received from the server.
	//    Published events will contain both the received data and the tracing context to continue
	//    the tracing span from the source (= the websocket engine).
	//
	// In case when the connection with the server is lost, the websocket client will publish a
	// connection_interrupted event to warn consumer about the failure.
	//
	// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
	// when it reconnects to the server and it MUST reuse the previously provided channel to publish
	// received messages.
	//
	// Consumers should always watch to the event type to separate messages from the connection
	// failure events and react according the event type.
	//
	// Finally, the provided channel will be automatically closed by the client when:
	//	- The user unsubscribe from the topic by using UnsubscribeTickers
	//	- The websocket client definitely stops.
	//
	// Consumers should also watch channel closure to know when no more data will be delivered.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- connection_interrupted
	//	- ticker
	//
	// # Extract data
	//
	// Before parsing the data, check the event type to catch rare connection_interrupted events.
	//
	// The event data contains the JSON payload from the server and can be parsed into a structure
	// of type messages.Ticker like this:
	//
	//	ticker := new(messages.Ticker)
	//	err := event.DataAs(ticker)
	//
	// The event will also contain the tracing context from OpenTelemetry. This tracing context can
	// be extracted from the event to continue tracing the event processing from the source:
	//
	//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- pair: Pairs to subscribe to.
	//	- rcv: Channel used to publish ticker messages and connection_interrupted events.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- There is already an active subscription.
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active subscription.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	//
	//	- A connection_interrupted event MUST be published on the channel each time the websocket
	//    connection is closed.
	//
	//	- The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
	//
	//	- The websocket client implementation CAN either use blocking writes or discard messages in
	//    case the provided channel is full. It is up to the client implementation to be clear about
	//    how it deals with congestion.
	//
	//	- If the client implementation has a mechanism to automatically reconnect to the server,
	//    then the websocket client MUST resubscribe to previously subscribed channels and reuse
	//    the channel that has been provided when the user subscribed to the channel.
	SubscribeTicker(ctx context.Context, pairs []string, rcv chan event.Event) error
	// # Description
	//
	// Subscribe to the ohlc channel with the given interval. In case of success, the websocket
	// client will start publishing received events on the user's provided channel.
	//
	// The client supports multiple subscriptions but for different interval. The client will return
	// an error in case there's already a subscription for that interval.
	//
	// Two types of events can be published on the channel:
	//	- connection_interrupted: This event type is used when connection with the sevrer has been
	//    interrupted. The event will not have any data. It only serves as a cue for the consumer
	//    to allow the consumer to react when the connection with the server is interrupted.
	//	- ohlc: This event type is used when a message has been received from the server.
	//    Published events will contain both the received data and the tracing context to continue
	//    the tracing span from the source (= the websocket engine).
	//
	// In case when the connection with the server is lost, the websocket client will publish a
	// connection_interrupted event to warn consumer about the failure.
	//
	// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
	// when it reconnects to the server and it MUST reuse the previously provided channel to publish
	// received messages.
	//
	// Consumers should always watch to the event type to separate messages from the connection
	// failure events and react according the event type.
	//
	// Finally, the provided channel will be automatically closed by the client when:
	//	- The user unsubscribe from the topic by using UnsubscribeOHLC
	//	- The websocket client definitely stops.
	//
	// Consumers should also watch channel closure to know when no more data will be delivered.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- connection_interrupted
	//	- ohlc
	//
	// # Extract data
	//
	// Before parsing the data, check the event type to catch rare connection_interrupted events.
	//
	// The event data contains the JSON payload from the server and can be parsed into a structure
	// of type messages.OHLC like this:
	//
	//	ohlc := new(messages.OHLC)
	//	err := event.DataAs(ohlc)
	//
	// The event will also contain the tracing context from OpenTelemetry. This tracing context can
	// be extracted from the event to continue tracing the event processing from the source:
	//
	//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- pair: Pairs to subscribe to.
	//	- interval: Interval for produced OHLC indicators.
	//	- rcv: Channel used to publish ohlc messages and connection_interrupted events.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- There is already an active subscription for that interval.
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active subscription.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	//
	//	- A connection_interrupted event MUST be published on the channel each time the websocket
	//    connection is closed.
	//
	//	- The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
	//
	//	- The websocket client implementation CAN either use blocking writes or discard messages in
	//    case the provided channel is full. It is up to the client implementation to be clear about
	//    how it deals with congestion.
	//
	//	- If the client implementation has a mechanism to automatically reconnect to the server,
	//    then the websocket client MUST resubscribe to previously subscribed channels and reuse
	//    the channel that has been provided when the user subscribed to the channel.
	SubscribeOHLC(ctx context.Context, pairs []string, interval messages.IntervalEnum, rcv chan event.Event) error
	// # Description
	//
	// Subscribe to the trades channel. In case of success, the websocket client will start
	// publishing received events on the user's provided channel.
	//
	// Two types of events can be published on the channel:
	//	- connection_interrupted: This event type is used when connection with the sevrer has been
	//    interrupted. The event will not have any data. It only serves as a cue for the consumer
	//    to allow the consumer to react when the connection with the server is interrupted.
	//	- trade: This event type is used when a message has been received from the server.
	//    Published events will contain both the received data and the tracing context to continue
	//    the tracing span from the source (= the websocket engine).
	//
	// In case when the connection with the server is lost, the websocket client will publish a
	// connection_interrupted event to warn consumer about the failure.
	//
	// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
	// when it reconnects to the server and it MUST reuse the previously provided channel to publish
	// received messages.
	//
	// Consumers should always watch to the event type to separate messages from the connection
	// failure events and react according the event type.
	//
	// Finally, the provided channel will be automatically closed by the client when:
	//	- The user unsubscribe from the topic by using UnsubscribeTrades
	//	- The websocket client definitely stops.
	//
	// Consumers should also watch channel closure to know when no more data will be delivered.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- connection_interrupted
	//	- trade
	//
	// # Extract data
	//
	// Before parsing the data, check the event type to catch rare connection_interrupted events.
	//
	// The event data contains the JSON payload from the server and can be parsed into a structure
	// of type messages.Trade like this:
	//
	//	trade := new(messages.Trade)
	//	err := event.DataAs(trade)
	//
	// The event will also contain the tracing context from OpenTelemetry. This tracing context can
	// be extracted from the event to continue tracing the event processing from the source:
	//
	//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- pair: Pairs to subscribe to.
	//	- rcv: Channel used to publish trade messages and connection_interrupted events.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- There is already an active subscription.
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active subscription.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	//
	//	- A connection_interrupted event MUST be published on the channel each time the websocket
	//    connection is closed.
	//
	//	- The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
	//
	//	- The websocket client implementation CAN either use blocking writes or discard messages in
	//    case the provided channel is full. It is up to the client implementation to be clear about
	//    how it deals with congestion.
	//
	//	- If the client implementation has a mechanism to automatically reconnect to the server,
	//    then the websocket client MUST resubscribe to previously subscribed channels and reuse
	//    the channel that has been provided when the user subscribed to the channel.
	SubscribeTrade(ctx context.Context, pairs []string, rcv chan event.Event) error
	// # Description
	//
	// Subscribe to the spreads channel. In case of success, the websocket client will start
	// publishing received events on the user's provided channel.
	//
	// Two types of events can be published on the channel:
	//	- connection_interrupted: This event type is used when connection with the sevrer has been
	//    interrupted. The event will not have any data. It only serves as a cue for the consumer
	//    to allow the consumer to react when the connection with the server is interrupted.
	//	- spread: This event type is used when a message has been received from the server.
	//    Published events will contain both the received data and the tracing context to continue
	//    the tracing span from the source (= the websocket engine).
	//
	// In case when the connection with the server is lost, the websocket client will publish a
	// connection_interrupted event to warn consumer about the failure.
	//
	// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
	// when it reconnects to the server and it MUST reuse the previously provided channel to publish
	// received messages.
	//
	// Consumers should always watch to the event type to separate messages from the connection
	// failure events and react according the event type.
	//
	// Finally, the provided channel will be automatically closed by the client when:
	//	- The user unsubscribe from the topic by using UnsubscribeSpreads
	//	- The websocket client definitely stops.
	//
	// Consumers should also watch channel closure to know when no more data will be delivered.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- connection_interrupted
	//	- spread
	//
	// # Extract data
	//
	// Before parsing the data, check the event type to catch rare connection_interrupted events.
	//
	// The event data contains the JSON payload from the server and can be parsed into a structure
	// of type messages.Spread like this:
	//
	//	spread := new(messages.Spread)
	//	err := event.DataAs(spread)
	//
	// The event will also contain the tracing context from OpenTelemetry. This tracing context can
	// be extracted from the event to continue tracing the event processing from the source:
	//
	//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- pair: Pairs to subscribe to.
	//	- rcv: Channel used to publish spread messages and connection_interrupted events.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- There is already an active subscription.
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active subscription.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	//
	//	- A connection_interrupted event MUST be published on the channel each time the websocket
	//    connection is closed.
	//
	//	- The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
	//
	//	- The websocket client implementation CAN either use blocking writes or discard messages in
	//    case the provided channel is full. It is up to the client implementation to be clear about
	//    how it deals with congestion.
	//
	//	- If the client implementation has a mechanism to automatically reconnect to the server,
	//    then the websocket client MUST resubscribe to previously subscribed channels and reuse
	//    the channel that has been provided when the user subscribed to the channel.
	SubscribeSpread(ctx context.Context, pairs []string, rcv chan event.Event) error
	// # Description
	//
	// Subscribe to the book channel. In case of success, the websocket client will start
	// publishing received events on the user's provided channel.
	//
	// Three types of events can be published on the channel:
	//	- connection_interrupted: This event type is used when connection with the sevrer has been
	//    interrupted. The event will not have any data. It only serves as a cue for the consumer
	//    to allow the consumer to react when the connection with the server is interrupted.
	//	- book_snapshot: This event type is used when a snapshot of the order book is received from
	//    the websocket server.
	//	- book_update: This event is used when an update about the order book is received from the
	//    websocket server.
	//
	// In case when the connection with the server is lost, the websocket client will publish a
	// connection_interrupted event to warn consumer about the failure.
	//
	// If the websocket client has a auto-reconnect feature, it MUST resubscribe to the publication
	// when it reconnects to the server and it MUST reuse the previously provided channel to publish
	// received messages.
	//
	// Consumers should always watch to the event type to separate messages from the connection
	// failure events and react according the event type.
	//
	// Finally, the provided channel will be automatically closed by the client when:
	//	- The user unsubscribe from the topic by using UnsubscribeBook
	//	- The websocket client definitely stops.
	//
	// Consumers should also watch channel closure to know when no more data will be delivered.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- connection_interrupted
	//	- book_snapshot
	//	- book_update
	//
	// # Extract data
	//
	// Before parsing the data, check the event type to catch rare connection_interrupted events.
	//
	// The event data contains the JSON payload from the server and can be parsed into a structure
	// of type messages.BookUpdate or messages.BookSnapshot depending on the event type like this:
	//
	//	switch(WebsocketClientEventTypeEnum(event.Type)) {
	//		case events.BookSnapshot:
	//			snapshot := new(messages.BookSnapshot)
	//			err := event.DataAs(snapshot)
	//		case events.BookUpdate:
	//			update := new(messages.BookUpdate)
	//			err := event.DataAs(update)
	//		case events.ConnectionInterrupted:
	//			panic("connection lost")
	//		default:
	//			panic("unknown message type", event.Type)
	//	}
	//
	// The event will also contain the tracing context from OpenTelemetry. This tracing context can
	// be extracted from the event to continue tracing the event processing from the source:
	//
	//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- pair: Pairs to subscribe to.
	//	- rcv: Channel used to publish book_snapshot & book+update messages and
	//         connection_interrupted events.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- There is already an active subscription.
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active subscription.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	//
	//	- A connection_interrupted event MUST be published on the channel each time the websocket
	//    connection is closed.
	//
	//	- The provided channel MUST be closed upon unsubscribe or when the websocket client stops.
	//
	//	- The websocket client implementation CAN either use blocking writes or discard messages in
	//    case the provided channel is full. It is up to the client implementation to be clear about
	//    how it deals with congestion.
	//
	//	- If the client implementation has a mechanism to automatically reconnect to the server,
	//    then the websocket client MUST resubscribe to previously subscribed channels and reuse
	//    the channel that has been provided when the user subscribed to the channel.
	SubscribeBook(ctx context.Context, pairs []string, depth messages.DepthEnum, rcv chan event.Event) error
	// # Description
	//
	// Unsubscribe from the ticker channel. The channel provided on subscribe will be closed by
	// the websocket client.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the unsubscribe message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- In case of success, the client MUST close the channel used to publish events.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	UnsubscribeTicker(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the ohlc channel with the given interva. The channel provided on subscribe
	// will be closed by the websocket client.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- interval: Used to determine which subscription must be cancelled.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the unsubscribe message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- In case of success, the client MUST close the channel used to publish events.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	UnsubscribeOHLC(ctx context.Context, interval messages.IntervalEnum) error
	// # Description
	//
	// Unsubscribe from the trade channel. The channel provided on subscribe will be closed by
	// the websocket client.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the unsubscribe message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- In case of success, the client MUST close the channel used to publish events.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	UnsubscribeTrade(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the spread channel. The channel provided on subscribe will be closed by
	// the websocket client.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the unsubscribe message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- In case of success, the client MUST close the channel used to publish events.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	UnsubscribeSpread(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the book channel. The channel provided on subscribe will be closed by
	// the websocket client.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the unsubscribe message.
	//	- The provided context expires before subscription is completed (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- In case of success, the client MUST close the channel used to publish events.
	//
	//	- The client MUST use the right error type as described in the "Return" section.
	UnsubscribeBook(ctx context.Context) error
	// # Description
	//
	// Get the client's built-in channel used to publish received system status updates.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- system_status
	//
	//	# Return
	//
	// The client's built-in channel used to publish received system status updates.
	//
	// # Implemetation and usage guidelines
	//
	//	- The client MUST provide the channel it will use to publish heartbeats even though the
	//    cllient has not been started yet and is not connected to the server.
	//
	//	- The client MUST close the channel when it definitely stops.
	//
	//	- As the channel is automatically subscribed to, the client implementation must deal with
	//    possible channel congestion by discarding messages in a FIFO or LIFO fashion. The client
	//    must indicate how congestion is handled.
	GetSystemStatusChannel() chan event.Event
	// # Description
	//
	// Get the client's built-in channel to publish received heartbeats.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- heartbeat
	//
	//	# Return
	//
	// # Implemetation and usage guidelines
	//
	//	- The client MUST provide the channel it will use to publish heartbeats even though the
	//    cllient has not been started yet and is not connected to the server.
	//
	//	- The client MUST close the channel when it definitely stops.
	//
	//	- As the channel is automatically subscribed to, the client implementation must deal with
	//    possible channel congestion by discarding messages in a FIFO or LIFO fashion. The client
	//    must indicate how congestion is handled.
	//
	// # Return
	//
	// The client's built-in channel used to publish received heartbeats.
	GetHeartbeatChannel() chan event.Event
}
