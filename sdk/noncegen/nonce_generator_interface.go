// This package provides an interface and several implementations for a nonce generator.
package noncegen

// Interface which defines a method to get a unique incrementing nonce which will be used to sign
// a request to Kraken API.
type NonceGenerator interface {
	// Generate a new nonce.
	GenerateNonce() int64
}
