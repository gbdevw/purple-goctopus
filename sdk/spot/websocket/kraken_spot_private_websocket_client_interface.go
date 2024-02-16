package websocket

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
)

// Interface for a websocket client using the private environment for Kraken spot websocket API.
//
// Private websocket client has access to:
//   - Ping
//   - OwnTrades feed
//   - OpenOrders feed
//   - Add order operation
//   - Edit order operation
//   - Cancel order operations
type KrakenSpotPrivateWebsocketClientInterface interface {
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
	//	- The provided context expires before pong is received (OperationInterruptedError).
	//	- An error message is received from the server (OperationError).
	Ping(ctx context.Context) error
	// # Description
	//
	// Add a new order and wait until a AddOrderResponse response is received from the server or
	// until an error or a timeout occurs.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- params: AddOrder request parameters.
	//
	// # Return
	//
	// The AddOrderResponse message from the server if any has been received. In case the response
	// has its error message set, an error with the error message will also be returned.
	//
	// An error is returned when:
	//
	//	- The client failed to send the request (no specific error type).
	//	- A timeout has occured before the request could be sent (no specific error type)
	//	- An error message is received from the server (OperationError).
	//	- A timeout or network failure occurs after sending the request to the server, while
	//    waiting for the server response. In this case, a OperationInterruptedError is returned.
	AddOrder(ctx context.Context, params AddOrderRequestParameters) (*messages.AddOrderResponse, error)
	// # Description
	//
	// Edit an existing order and wait until a EditOrderResponse response is received from the
	// server or until an error or a timeout occurs.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- params: EditOrder request parameters.
	//
	// # Return
	//
	// The EditOrderResponse message from the server if any has been received. In case the response
	// has its error message set, an error with the error message will also be returned.
	//
	// An error is returned when:
	//
	//	- The client failed to send the request (no specific error type).
	//	- A timeout has occured before the request could be sent (no specific error type)
	//	- An error message is received from the server (OperationError).
	//	- A timeout or network failure occurs after sending the request to the server, while
	//    waiting for the server response. In this case, a OperationInterruptedError is returned.
	EditOrder(ctx context.Context, params EditOrderRequestParameters) (*messages.EditOrderResponse, error)
	// # Description
	//
	// Cancel one or several existing orders and wait until a CancelOrderResponse response is
	// received from the server or until an error or a timeout occurs.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- params: CancelOrder request parameters.
	//
	// # Return
	//
	// The CancelOrderResponse message from the server if any has been received. In case the response
	// has its error message set, an error with the error message will also be returned.
	//
	// An error is returned when:
	//
	//	- The client failed to send the request (no specific error type).
	//	- A timeout has occured before the request could be sent (no specific error type)
	//	- An error message is received from the server (OperationError).
	//	- A timeout or network failure occurs after sending the request to the server, while
	//    waiting for the server response. In this case, a OperationInterruptedError is returned.
	CancelOrder(ctx context.Context, params CancelOrderRequestParameters) (*messages.CancelOrderResponse, error)
	// # Description
	//
	// Cancel all orders and wait until a CancelAllOrdersResponse response is received from the
	// server or until an error or a timeout occurs.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//
	// # Return
	//
	// The CancelAllOrdersResponse message from the server if any has been received. In case the response
	// has its error message set, an error with the error message will also be returned.
	//
	// An error is returned when:
	//
	//	- The client failed to send the request (no specific error type).
	//	- A timeout has occured before the request could be sent (no specific error type)
	//	- An error message is received from the server (OperationError).
	//	- A timeout or network failure occurs after sending the request to the server, while
	//    waiting for the server response. In this case, a OperationInterruptedError is returned.
	CancellAllOrders(ctx context.Context) (*messages.CancelAllOrdersResponse, error)
	// # Description
	//
	// Set, extend or unset a timer which cancels all orders when expiring and wait until a
	// response is received from the server or until an error or a timeout occurs.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- params: CancellAllOrdersAfterX request parameters.
	//
	// # Return
	//
	// The CancelAllOrdersAfterXResponse message from the server if any has been received. In case
	// the response has its error message set, an error with the error message is also be returned.
	//
	// An error is returned when:
	//
	//	- The client failed to send the request (no specific error type).
	//	- A timeout has occured before the request could be sent (no specific error type)
	//	- An error message is received from the server (OperationError).
	//	- A timeout or network failure occurs after sending the request to the server, while
	//    waiting for the server response. In this case, a OperationInterruptedError is returned.
	CancellAllOrdersAfterX(ctx context.Context, params CancelAllOrdersAfterXRequestParameters) (*messages.CancelAllOrdersAfterXResponse, error)
	// # Description
	//
	// Subscribe to the ownTrades channel. In case of success, the websocket client will start
	// publishing received events on the user's provided channel.
	//
	// Two types of events can be published on the channel:
	//	- connection_interrupted: This event type is used when connection with the sevrer has been
	//    interrupted. The event will not have any data. It only serves as a cue for the consumer
	//    to allow the consumer to react when the connection with the server is interrupted.
	//	- own_trades: This event type is used when a message has been received from the server.
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
	//	- The user unsubscribe from the topic by using UnsubscribeOwnTrades
	//	- The websocket client definitely stops.
	//
	// Consumers should also watch channel closure to know when no more data will be delivered.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- connection_interrupted
	//	- own_trades
	//
	// # Extract data
	//
	// Before parsing the data, check the event type to catch rare connection_interrupted events.
	//
	// The event data contains the JSON payload from the server and can be parsed into a structure
	// of type messages.OwnTrades like this:
	//
	//	ownTrade := new(messages.OwnTrades)
	//	err := event.DataAs(ownTrade)
	//
	//
	// The event will also contain the tracing context from OpenTelemetry. This tracing context can
	// be extracted from the event to continue tracing the event processing from the source:
	//
	//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- snapshot: If true, upon subscription, the 50 most recent user trades will be published.
	//	- consolidateTaker: Whether to consolidate order fills by root taker trade(s).
	//	- rcv: Channel used to publish own_trades messages and connection_interrupted events.
	//
	// # Return
	//
	// An error is returned when:
	//
	//	- There is already an active subscription.
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires (timeout/cancel) before subscription is completed.
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active susbscription.
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
	SubscribeOwnTrades(ctx context.Context, snapshot bool, consolidateTaker bool, rcv chan event.Event) error
	// # Description
	//
	// Subscribe to the openOrders channel. In case of success, the websocket client will start
	// publishing received events on the user's provided channel.
	//
	// Two types of events can be published on the channel:
	//	- connection_interrupted: This event type is used when connection with the sevrer has been
	//    interrupted. The event will not have any data. It only serves as a cue for the consumer
	//    to allow the consumer to react when the connection with the server is interrupted.
	//	- open_orders: This event type is used when a message has been received from the server.
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
	//	- The user unsubscribe from the topic by using UnsubscribeOpenOrders
	//	- The websocket client definitely stops.
	//
	// Consumers should also watch channel closure to know when no more data will be delivered.
	//
	// # Event types
	//
	// Only these types of events will be published on the channel (Cf. WebsocketClientEventTypeEnum):
	//	- connection_interrupted
	//	- open_orders
	//
	// # Extract data
	//
	// Before parsing the data, check the event type to catch rare connection_interrupted events.
	//
	// The event data contains the JSON payload from the server and can be parsed into a structure
	// of type messages.OpenOrders like this:
	//
	//	openOrders := new(messages.OpenOrders)
	//	err := event.DataAs(openOrders)
	//
	//
	// The event will also contain the tracing context from OpenTelemetry. This tracing context can
	// be extracted from the event to continue tracing the event processing from the source:
	//
	//	ctx := otelObs.ExtractDistributedTracingExtension(context.Background(), event)
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- rateCounter: If true, rate limiting information will be included in messages.
	//	- rcv: Channel used to publish open_orders messages and connection_interrupted events.
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
	SubscribeOpenOrders(ctx context.Context, rateCounter bool, rcv chan event.Event) error
	// # Description
	//
	// Unsubscribe from the ownTrades channel. The channel provided on subscribe will bbe closed by
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
	UnsubscribeOwnTrades(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the openOrders channel. The channel provided on subscribe will bbe closed by
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
	UnsubscribeOpenOrders(ctx context.Context) error
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
