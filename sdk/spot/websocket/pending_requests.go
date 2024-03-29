package websocket

import (
	"github.com/gbdevw/purple-goctopus/sdk/spot/websocket/messages"
)

// Container for pending websocket requests.
type pendingRequests struct {
	// Pending Ping requests per Request ID
	pendingPing map[int64]*pendingPing
	// Pending Subscribe requests per Request ID
	pendingSubscribe map[int64]*pendingSubscribe
	// Pending Unsubscribe requests per Request ID
	pendingUnsubscribe map[int64]*pendingUnsubscribe
	// Pending AddOrder requests per Request ID
	pendingAddOrderRequests map[int64]*pendingAddOrderRequest
	// Pending EditOrder requests per Request ID
	pendingEditOrderRequests map[int64]*pendingEditOrderRequest
	// Pending CancelOrder requests per Request ID
	pendingCancelOrderRequests map[int64]*pendingCancelOrderRequest
	// Pending CancelAllOrders requests per Request ID
	pendingCancelAllOrdersRequests map[int64]*pendingCancelAllOrdersRequest
	// Pending CancelAllOrdersAfterX requests per Request ID
	pendingCancelAllOrdersAfterXRequests map[int64]*pendingCancelAllOrdersAfterXRequest
}

// Data of a pending Ping request which contains channels whch can be used to provide the
// request results.
type pendingPing struct {
	// Channel to use to push the received response to requester.
	resp chan *messages.Pong
	// Channel used to push errors to requester.
	err chan error
}

// Data of a pending Subscribe request which contains channels whch can be used to provide the
// request results.
type pendingSubscribe struct {
	// Request pairs
	pairs []string
	// Map which tracks whether a response has been received for the given pair
	served map[string]bool
	// Map which records error messages received when some pairs could not be subscribed to
	errPerPair map[string]error
	// Channel used to push errors to requester.
	err chan error
}

// Data of a pending Unsubscribe request which contains channels whch can be used to provide the
// request results.
type pendingUnsubscribe struct {
	// Request pairs
	pairs []string
	// Map which tracks whether a response has been received for the given pair
	served map[string]bool
	// Map which records error messages received when some pairs could not be subscribed to
	errPerPair map[string]error
	// Channel used to push errors to requester.
	err chan error
}

// Data of a pending AddOrder request which contains channels whch can be used to provide the
// request results.
type pendingAddOrderRequest struct {
	// Channel to use to push the received response to requester.
	resp chan *messages.AddOrderResponse
	// Channel used to push errors to requester.
	err chan error
}

// Data of a pending EditOrder request which contains channels whch can be used to provide the
// request results.
type pendingEditOrderRequest struct {
	// Channel to use to push the received response to requester.
	resp chan *messages.EditOrderResponse
	// Channel used to push errors to requester.
	err chan error
}

// Data of a pending CancelOrder request which contains channels whch can be used to provide the
// request results.
type pendingCancelOrderRequest struct {
	// Channel to use to push the received response to requester.
	resp chan *messages.CancelOrderResponse
	// Channel used to push errors to requester.
	err chan error
}

// Data of a pending CancelAllOrders request which contains channels whch can be used to provide the
// request results.
type pendingCancelAllOrdersRequest struct {
	// Channel to use to push the received response to requester.
	resp chan *messages.CancelAllOrdersResponse
	// Channel used to push errors to requester.
	err chan error
}

// Data of a pending CancelAllOrdersAfterX request which contains channels whch can be used to provide the
// request results.
type pendingCancelAllOrdersAfterXRequest struct {
	// Channel to use to push the received response to requester.
	resp chan *messages.CancelAllOrdersAfterXResponse
	// Channel used to push errors to requester.
	err chan error
}
