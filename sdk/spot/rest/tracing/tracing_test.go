package tracing

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gbdevw/purple-goctopus/sdk/spot/rest/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

/*************************************************************************************************/
/* TRACING: UNIT TEST SUITE                                                                      */
/*************************************************************************************************/

// Unit test suite for Tracing.
type TracingUnitTestSuite struct {
	suite.Suite
	// Tracer used to generate spans
	tracer trace.Tracer
}

// Run unit test suite.
func TestTracingUnitTestSuite(t *testing.T) {
	// Get a tracer from global tracer provider.
	suite.Run(t, &TracingUnitTestSuite{
		tracer: otel.GetTracerProvider().Tracer(PackageName, trace.WithInstrumentationVersion(PackageVersion)),
	})
}

/*************************************************************************************************/
/* TRACING: UNIT TEST SUITE                                                                      */
/*************************************************************************************************/

// Test TraceErrorAndSetStatus.
//
// This test only ensures all paths work as expected.
func (suite *TracingUnitTestSuite) TestTraceErrorAndSetStatus() {
	// Start a span
	_, span := suite.tracer.Start(context.Background(), "test", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	// Test function with no error
	TraceErrorAndSetStatus(span, nil)
	// Test function with an error
	TraceErrorAndSetStatus(span, fmt.Errorf("fail"))
	// Test panic when span is nil
	require.Panics(suite.T(), func() { TraceErrorAndSetStatus(nil, nil) })
}

// Test TraceApiOperationAndSetStatus.
//
// This test only ensures all paths work as expected.
func (suite *TracingUnitTestSuite) TestTraceApiOperationAndSetStatus() {
	// Start a span
	_, span := suite.tracer.Start(context.Background(), "test", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	// Test function with no error and non nil responses
	TraceApiOperationAndSetStatus(span, new(common.KrakenSpotRESTResponse), new(http.Response), nil)
	// Test function with no error, non nil responses and errors in api response
	TraceApiOperationAndSetStatus(span, &common.KrakenSpotRESTResponse{Error: []string{"1", "2"}}, new(http.Response), nil)
	// Test function with an error
	TraceApiOperationAndSetStatus(span, nil, nil, fmt.Errorf("fail"))
	// Test all nil case
	TraceApiOperationAndSetStatus(span, nil, nil, nil)
	// Test panic when span is nil
	require.Panics(suite.T(), func() { TraceApiOperationAndSetStatus(nil, nil, nil, nil) })
}
