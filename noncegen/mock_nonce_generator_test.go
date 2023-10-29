package noncegen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test MockNonceGenerator compliance with NonceGenerator interface
func TestMockNonceGeneratorInterfaceCompliance(t *testing.T) {
	var instance interface{} = NewMockNonceGenerator()
	_, ok := instance.(NonceGenerator)
	require.True(t, ok)
}

// Test MockNonceGenerator GenerateNonce
func TestMockNonceGenerator(t *testing.T) {
	// Create and configure mock
	expected := 1
	mock := NewMockNonceGenerator()
	mock.On("GenerateNonce").Return(expected)
	// Call mocked generate nonce
	nonce := mock.GenerateNonce()
	// Check
	require.Equal(t, int64(expected), nonce)
	mock.AssertCalled(t, "GenerateNonce")
}
