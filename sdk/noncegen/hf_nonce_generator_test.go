package noncegen

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test HFNonceGenerator compliance with NonceGenerator interface
func TestHFNonceGeneratorInterfaceCompliance(t *testing.T) {
	var instance interface{} = NewHFNonceGenerator()
	_, ok := instance.(NonceGenerator)
	require.True(t, ok)
}

// Test HFNonceGenerator GenerateNonce
func TestHFNonceGenerator(t *testing.T) {
	// Save current time as UNIX nanosec timestamp
	now := time.Now().UnixNano()
	// Create a HFNonceGenerator
	gen := NewHFNonceGenerator()
	// Generate a Nonce
	nonce := gen.GenerateNonce()
	// Check generated nonce:
	// - nonce must be greater than the timestamp (or equal)
	// - nonce - gen.base must be equal to 0
	// - inc must be equal to 1
	require.GreaterOrEqual(t, nonce, now)
	require.Equal(t, int64(0), nonce-gen.base)
	require.Equal(t, int64(1), gen.inc)
}
