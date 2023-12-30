package noncegen

import "time"

// Nonce generator which returns UNIX millisecond timestamps as nonces.
//
// This nonce generator has a low risk of nonce collision / bad nonce error in case of use
// at high frequency or in a distributed system.
type UnixMillisNonceGenerator struct{}

// Factory which returns a new UnixMillisNonceGenerator.
func NewUnixMillisNonceGenerator() *UnixMillisNonceGenerator {
	return new(UnixMillisNonceGenerator)
}

// Generate a new nonce that is a UNIX millisecond timestamp.
func (g *UnixMillisNonceGenerator) GenerateNonce() int64 {
	return time.Now().UnixMilli()
}
