package noncegen

import (
	"sync"
	"time"
)

// A thread-safe nonce generator with no collision risk when used at high frequency. The nonce
// generator generate nonce from two numbers that are added:
//   - base: The UNIX nanosec timestamp of the moment when the generator has been created. This
//     ensures generated nonce will always increase even in case of application restart (without
//     a persistence layer).
//   - inc: An atomic counter which increases each time a nonce is generated. This ensures the
//     generator thread-safety.
//
// WARNING: The nonce generator has no risk of collision in case of usage in a monolithic system
// only. In case several applications need to access Kraken API, use different API keys for each
// application.
type HFNonceGenerator struct {
	// Base used to compute nonce.
	//
	// Base should be set when nonce generator is created as a UNIX nanosec timestamp. That way,
	// generated nonce will always use an increased base at each appplication restart.
	base int64
	// A value which increments each time a nonce is produced
	inc int64
	// Mutex used to protect inc and make the nonce generator thread-safe.
	mu sync.Mutex
}

// Factory which returns a new ready-to-use HFNonceGenerator.
func NewHFNonceGenerator() *HFNonceGenerator {
	return &HFNonceGenerator{
		base: time.Now().UnixNano(),
		inc:  0,
		mu:   sync.Mutex{},
	}
}

// Generate a new nonce.
func (g *HFNonceGenerator) GenerateNonce() int64 {
	// Lock mutex and defer Unlock
	g.mu.Lock()
	defer g.mu.Unlock()
	// Generate nonce -> add base and inc, increase counter and return nonce
	nonce := g.base + g.inc
	g.inc = g.inc + 1
	return nonce
}
