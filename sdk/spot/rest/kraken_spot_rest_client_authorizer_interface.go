package rest

import (
	"context"
	"net/http"
)

// Interface for a component which manages authorizations for requests to Kraken Spot REST API.
//
// This component is meant to allow users to plug-in their own authorization logic. It can be
// extended to manage:
//   - Outgoing request filtering and post-processing based on user's own criterias.
//   - Proxy/Egress: Send (unsigned or authorized by another means) requests to a proxy/egress gateway
//     that will add the signature and send the request to Kraken.
//   - ...
type KrakenSpotRESTClientAuthorizerIface interface {
	// # Description
	//
	// Authorize and post-process the outgoing http.Request. The returned http.Request reference will be used
	// to send the request.
	//
	// In case the authorizer returns an error, the client MUST NOT send the http.Request.
	//
	// # Inputs
	//
	//	- ctx: Context used for tracing and coordination purpose.
	//	- req: Reference to the http.Request to authorize. Must not be nil.
	//
	// # Returns
	//
	//	- A reference to the authorized http.Request to send.
	//	- An error if any.
	Authorize(ctx context.Context, req *http.Request) (*http.Request, error)
}
