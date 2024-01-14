package websocket

import (
	"context"

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
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server.
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
	// Subscribe to the ownTrades channel. In case of success, a channel with the provided capacity
	// will be created and returned.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- snapshot: If true, upon subscription, the 50 most recent user trades will be published.
	//	- consolidateTaker: Whether to consolidate order fills by root taker trade(s).
	//	- capacity: Desired channel capacity. Can be 0 (not recommended).
	//
	// # Return
	//
	// In case of success, a channel with the desired capacity will be returned. Received data will
	// be published on that channel.
	//
	// An error (and no channel) is returned when:
	//
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active susbscription.
	//
	//	- A nil value MUST be published on the channel ONLY when the websocket connection is closed
	//    even if the client implementation has a mechanism to automatically reconnect to the
	//    websocket server. This nil value will serve as a cue for the consumer to detect
	//    interruptions in the stream of data and react to these interruptions.
	//
	//	- The websocket client implementation CAN either use blocking writes or discard messages in
	//    case the publish channel is full. It is up to the client implementation to be clear about
	//    how it deals with congestion.
	//
	//	- If the client implementation has a mechanism to automatically reconnect to the server AND
	//    resubscribe to previously subscribed channels, then, the client implementation MUST reuse
	//    the channel that has been previously created and returned to the user.
	//
	//	- The client MUST drop the channel if the user has used the corresponding Unsubscribe method.
	//    If the user use the subscribe method again, then, a new channel MUST be created and the
	//    older one MUST NOT be used anymore.
	SubscribeOwnTrades(ctx context.Context, snapshot bool, consolidateTaker bool, capacity int) (chan *messages.OwnTrades, error)
	// # Description
	//
	// Subscribe to the openOrders channel. In case of success, a channel with the provided
	// capacity will be created and returned.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- rateCounter: Whether to send rate-limit counter in updates.
	//	- capacity: Desired channel capacity. Can be 0 (not recommended).
	//
	// # Return
	//
	// In case of success, a channel with the desired capacity will be returned. Received data will
	// be published on that channel.
	//
	// An error (and no channel) is returned when:
	//
	//	- An error occurs when sending the subscription message.
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server (OperationError).
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST return an error if there is already an active susbscription.
	//
	//	- A nil value MUST be published on the channel ONLY when the websocket connection is closed
	//    even if the client implementation has a mechanism to automatically reconnect to the
	//    websocket server. This nil value will serve as a cue for the consumer to detect
	//    interruptions in the stream of data and react to these interruptions.
	//
	//	- The websocket client implementation CAN either use blocking writes or discard messages in
	//    case the publish channel is full. It is up to the client implementation to be clear about
	//    how it deals with congestion.
	//
	//	- If the client implementation has a mechanism to automatically reconnect to the server AND
	//    resubscribe to previously subscribed channels, then, the client implementation MUST reuse
	//    the channel that has been previously created and returned to the user.
	//
	//	- The client MUST drop the channel if the user has used the corresponding Unsubscribe method.
	//    If the user use the subscribe method again, then, a new channel MUST be created and the
	//    older one MUST NOT be used anymore.
	SubscribeOpenOrders(ctx context.Context, rateCounter bool, capacity int) (chan *messages.OpenOrders, error)
	// # Description
	//
	// Unsubscribe from the ownTrades channel. The previously used channel can be dropped as it
	// must not be used again.
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
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the message.
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server.
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST drop the channel that was used by the canceled subscription.
	//
	//	- The client MUST return an error if channel was not subscribed to.
	UnsubscribeOwnTrades(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the openOrders channel. The previously used channel can be dropped as it
	// must not be used again.
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
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the message.
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server.
	//
	// # Implementation and usage guidelines
	//
	//	- The client MUST drop the channel that was used by the canceled subscription.
	//
	//	- The client MUST return an error if channel was not subscribed to.
	UnsubscribeOpenOrders(ctx context.Context) error
}
