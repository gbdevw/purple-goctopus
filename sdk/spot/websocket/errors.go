package websocket

import "fmt"

// This error is used when the reply from the server to a request contains an error message.
//
// In this case, the error contains the error message from the server.
type OperationError struct {
	Operation string
	Root      error
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("%s failed: %s", e.Operation, e.Root.Error())
}

func (e *OperationError) Unwrap() error { return e.Root }

// This error is used when the reply from the server to a specific request could not be received
// because either connection with the websocket server was lost while waiting for the server reply
// or because a timeout has occured while waiting for the reply. The error will embed the root
// error, either the cancel error from the context or the error from the websocket engine.
//
// When such error occurs, the client cannot tell whether its request has been successfully processed
// or not by the websocket server.
//
// Recommendation is:
//   - Always provide user defined IDs when adding/editing orders.
//   - Use the websocket/rest client to reconciliate state. Query open or closed orders by using
//     the expected user ID and then retry the operation if needed.
type OperationInterruptedError struct {
	Operation string
	Root      error
}

func (e *OperationInterruptedError) Error() string {
	return fmt.Sprintf("%s has been interrupted: %s", e.Operation, e.Root.Error())
}

func (e *OperationInterruptedError) Unwrap() error { return e.Root }

// This error is used to carry information about pairs for which subscribe or unsubscribe failed.
type SubscriptionError struct {
	// Map where keys are pairs for which subscribe/unsubscribe failed and value are the cause.
	Errs map[string]error
}

func (e *SubscriptionError) Error() string {
	return fmt.Sprintf("subscription failed for the following pairs: %v", e.Errs)
}

func (e *SubscriptionError) Unwrap() error { return nil }
