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
	// Subscribe to the ticker channel.
	//
	// In case of success, a channel with the desired capacity is created and returned. The channel
	// will be used to publish subscription's data.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
	//	- capacity: Desired channel capacity. Can be 0 (not recommended).
	//
	// # Return
	//
	// In case of success, a channel with the desired capacity will be returned. Received data will
	// be published on that channel.
	//
	// An error (and no channel) is returned when:
	//
	//	- A subscription is already active.
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
	SubscribeTicker(ctx context.Context, pairs []string, capacity int) (chan *messages.Ticker, error)
	// # Description
	//
	// Subscribe to the ohlc channel.
	//
	// In case of success, a channel with the desired capacity is created and returned. The channel
	// will be used to publish subscription's data.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel
	//    will be watched for timeout/cancel signal.
	//	- pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
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
	//	- A subscription is already active.
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
	SubscribeOHLC(ctx context.Context, pairs []string, interval messages.IntervalEnum, capacity int) (chan *messages.OHLC, error)
	// # Description
	//
	// Subscribe to the trade channel.
	//
	// In case of success, a channel with the desired capacity is created and returned. The channel
	// will be used to publish subscription's data.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
	//	- pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
	//	- capacity: Desired channel capacity. Can be 0 (not recommended).
	//
	// # Return
	//
	// In case of success, a channel with the desired capacity will be returned. Received data will
	// be published on that channel.
	//
	// An error (and no channel) is returned when:
	//
	//	- A subscription is already active.
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
	SubscribeTrade(ctx context.Context, pairs []string, capacity int) (chan *messages.Trade, error)
	// # Description
	//
	// Subscribe to the spread channel.
	//
	// In case of success, a channel with the desired capacity is created and returned. The channel
	// will be used to publish subscription's data.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
	//	- pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
	//	- capacity: Desired channel capacity. Can be 0 (not recommended).
	//
	// # Return
	//
	// In case of success, a channel with the desired capacity will be returned. Received data will
	// be published on that channel.
	//
	// An error (and no channel) is returned when:
	//
	//	- A subscription is already active.
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
	SubscribeSpread(ctx context.Context, pairs []string, capacity int) (chan *messages.Spread, error)
	// # Description
	//
	// Subscribe to the book channel.
	//
	// In case of success, a channel with the desired capacity is created and returned. The channel
	// will be used to publish subscription's data.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done channel will be watched for timeout/cancel signal.
	//	- pairs: Array of currency pairs to subscribe to. Format of each pair is "A/B".
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
	SubscribeBook(ctx context.Context, pairs []string, depth messages.DepthEnum, capacity int) (chan *messages.BookSnapshot, chan *messages.BookUpdate, error)
	// # Description
	//
	// Unsubscribe from  the ticker topic. The previously used channel will be dropped and user
	// must stop using it.
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
	UnsubscribeTicker(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from  the ohlc topic. The previously used channel will be dropped and user
	// must stop using it.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose. The provided context Done
	//    channel will be watched for timeout/cancel signal.
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
	UnsubscribeOHLC(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from  the trade topic. The previously used channel will be dropped and user
	// must stop using it.
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
	UnsubscribeTrade(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from  the spread topic. The previously used channel will be dropped and user
	// must stop using it.
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
	UnsubscribeSpread(ctx context.Context) error
	// # Description
	//
	// Unsubscribe from  the book topic. The previously used channel will be dropped and user
	// must stop using it.
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
	UnsubscribeBook(ctx context.Context) error
	// # Description
	//
	// Get the client's built-in channel to publish received system status updates.
	//
	// # Implemetation and usage guidelines
	//
	//	- As the channel is automatically subscribed to, the client implementation CAN discard messages
	//    in case of congestion in the publication channel. The client implementation must be clear
	//    about how it deals with congestion.
	//
	// # Return
	//
	// The client's built-in channel used to publish received system status updates.
	GetSystemStatusChannel() chan *messages.SystemStatus
	// # Description
	//
	// Get the client's built-in channel to publish received heartbeats.
	//
	// # Implemetation and usage guidelines
	//
	//	- As the channel is automatically subscribed to, the client implementation CAN discard messages
	//    in case of congestion in the publication channel. The client implementation must be clear
	//    about how it deals with congestion.
	//
	// # Return
	//
	// The client's built-in channel used to publish received heartbeats.
	GetHeartbeatChannel() chan *messages.Heartbeat
}
