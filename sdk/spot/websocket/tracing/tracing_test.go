package tracing

import (
	"context"
	"fmt"
	"testing"

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

// Test HandleAndTraceError.
//
// This test only ensures all paths work as expected.
func (suite *TracingUnitTestSuite) TestHandleAndTraceError() {
	// Start a span
	_, span := suite.tracer.Start(context.Background(), "test", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	// Test function with no error
	require.NoError(suite.T(), HandleAndTraceError(span, nil))
	// Test function with an error
	require.Error(suite.T(), HandleAndTraceError(span, fmt.Errorf("fail")))
	// Test panic when span is nil
	require.Panics(suite.T(), func() { HandleAndTraceError(nil, nil) })
}
