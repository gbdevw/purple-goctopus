package rest

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// A mock for KrakenSpotRESTClientAuthorizerIface.
type MockKrakenSpotRESTClientAuthorizer struct {
	mock.Mock
}

// Factory which creates a new MockKrakenSpotRESTClientAuthorizer without any expectations set.
func NewMockKrakenSpotRESTClientAuthorizerr() *MockKrakenSpotRESTClientAuthorizer {
	return &MockKrakenSpotRESTClientAuthorizer{mock.Mock{}}
}

// Mocked GenerateNonce method
func (m *MockKrakenSpotRESTClientAuthorizer) Authorize(ctx context.Context, req *http.Request) (*http.Request, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*http.Request), args.Error(1)
}
