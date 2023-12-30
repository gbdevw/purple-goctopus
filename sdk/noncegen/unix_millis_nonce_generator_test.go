package noncegen

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test UnixMillisNonceGenerator compliance with NonceGenerator interface
func TestUnixMillisNonceGeneratornterfaceCompliance(t *testing.T) {
	var instance interface{} = NewUnixMillisNonceGenerator()
	_, ok := instance.(NonceGenerator)
	require.True(t, ok)
}

// Test UnixMillisNonceGenerator GenerateNonce
func TestUnixMillisNonceGenerator(t *testing.T) {
	// Save current time as UNIX nanosec & millisec timestamps
	nownano := time.Now().UnixNano()
	nowmillis := time.Now().UnixMilli()
	// Create a UnixMillisNonceGenerator
	gen := NewUnixMillisNonceGenerator()
	// Generate a Nonce
	nonce := gen.GenerateNonce()
	// Check generated nonce:
	// - nonce must be greater than the millisec timestamp (or equal)
	// - nonce must be lesser than the nanosec timestamp
	require.GreaterOrEqual(t, nonce, nowmillis)
	require.Less(t, nonce, nownano)
}
