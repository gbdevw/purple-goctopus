package noncegen

import "github.com/stretchr/testify/mock"

// A mock for NonceGenerator interface
type MockNonceGenerator struct {
	mock.Mock
}

// Factory which creates a new MockNonceGenerator without any expectations set.
func NewMockNonceGenerator() *MockNonceGenerator {
	return &MockNonceGenerator{mock.Mock{}}
}

// Mocked GenerateNonce method
func (m *MockNonceGenerator) GenerateNonce() int64 {
	args := m.Called()
	return int64(args.Int(0))
}
