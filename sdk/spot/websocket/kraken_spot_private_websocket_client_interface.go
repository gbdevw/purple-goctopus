// This package provides interfaces and implementations for websocket clients
// using Kraken spot websocket API (both public and private environments)
package websocket

import (
	"context"

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
	// Subscribe to the ticker channel. In case of success, a channel with the provided capacity
	// will be created and returned.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
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
	SubscribeTicker(ctx context.Context, capacity int) (chan *messages.Ticker, error)
	// # Description
	//
	// Subscribe to the ohlc channel. In case of success, a channel with the provided capacity will
	// be created and returned.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- interval: The desired interval for OHLC indicators. Multiple subscriptions can be
	//    maintained for different intervals.
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
	SubscribeOHLC(ctx context.Context, interval messages.IntervalEnum, capacity int) (chan *messages.OHLC, error)
	// # Description
	//
	// Subscribe to the trade channel. In case of success, a channel with the provided capacity will be created and returned.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
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
	SubscribeTrade(ctx context.Context, capacity int) (chan *messages.Trade, error)
	// # Description
	//
	// Subscribe to the spread channel. In case of success, a channel with the provided capacity will be created and returned.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
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
	SubscribeSpread(ctx context.Context, capacity int) (chan *messages.Spread, error)
	// # Description
	//
	// Subscribe to the ticker channel. In case of success, a channel with the provided capacity will be created and returned.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
	//	- depth: Desired book depth. Multiple subscriptions can be maintained for different depths.
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
	SubscribeBook(ctx context.Context, depth messages.DepthEnum, capacity int) (chan *messages.BookSnapshot, chan *messages.BookUpdate, error)
	// # Description
	//
	// Unsubscribe from the ticker channel. The previously used channel can be dropped as it must
	// not be used again.
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
	UnsubscribeTicker(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the ohlc channel with the given interval. The previously used channel can
	// be dropped as it must not be used again.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done
	//    channel will be watched for timeout/cancel signal.
	//	- interval: Used to target the right OHLC subscription to cancel. Multiple subscriptions
	//    can be maintained for different intervals.
	//
	// # Return
	//
	// Nil in case of success. Otherwise, an error is returned when:
	//
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the message.
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server.
	UnsubscribeOHLC(ctx context.Context, interval messages.IntervalEnum) error
	// # Description
	//
	// Unsubscribe from the trade channel. The previously used channel can be dropped as it must
	// not be used again.
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
	UnsubscribeTrade(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the spread channel. The previously used channel can be dropped as it must
	// not be used again.
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
	UnsubscribeSpread(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from the book channel with the given depth. The previously used channel can be
	// dropped as it must not be used again.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- depth: Target book depth. Multiple subscriptions can be maintained for different depths.
	//
	// # Return
	//
	// Nil in case of success. Otherwise, an error is returned when:
	//
	//	- The channel has not been subscribed to.
	//	- An error occurs when sending the message.
	//	- The provided context expires (timeout/cancel).
	//	- An error message is received from the server.
	UnsubscribeBook(ctx context.Context, depth messages.DepthEnum) error
}
