package tracing

import (
	"fmt"
	"net/http"

	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Package name used as instrumentation ID
	PackageName = "goctopus_sdk_spot_rest"
	// Package version
	PackageVersion = "0.0.0"
	// Span & events namespace
	TracesNamespace = "goctopus.spot.rest"
	// Kraken Sport REST API tracing header
	ResponseTracingHeader = "x-trace-id"
)

// A helper function to trace an error (if any) and set the provided span status
// accordingly.
//
// The function is emant to replace code blocks like this:
//
//	if err != nil {
//		span.RecordError(err)
//		span.SetStatus(codes.Error, codes.Error.String())
//	} else {
//		span.SetStatus(codes.Ok, codes.Ok.String())
//	}
//
// By this:
//
// TraceErrorAndSetStatus(span, err)
func TraceErrorAndSetStatus(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, codes.Error.String())
	} else {
		span.SetStatus(codes.Ok, codes.Ok.String())
	}
}

// # Description
//
// Helper function which handles the common tracing logic for the API operations outcome. The function will
// set the provided span status depending the API operation outcome.
//
// # Inputs
//
//   - span: Span to populate
//   - apiresp: Parsed API response. Can be nil and shoud be nil in case of error.
//   - apiresp: Received HTTP response. Can be nil and might not be nil in case of error.
//   - error: Error returned by the API operation. Can be nil in case of success.
func TraceApiOperationAndSetStatus(span trace.Span, apiresp *common.KrakenSpotRESTResponse, httpresp *http.Response, err error) {
	if httpresp != nil {
		// Add synthetic tracing data about the received http response if any -> status code and tracing header
		// This must not overlap too much with what an instrumented http.Client would report
		span.AddEvent(TracesNamespace+".http.response", trace.WithAttributes(
			attribute.Int("http.response.status_code", httpresp.StatusCode),
			attribute.String("http.response.header."+ResponseTracingHeader, httpresp.Header.Get(ResponseTracingHeader)),
		))
	}
	if err != nil {
		// Record error if any and set span status to error
		span.RecordError(err)
		span.SetStatus(codes.Error, codes.Error.String())
	} else {
		// Check the API response
		if apiresp != nil {
			if len(apiresp.Error) > 0 {
				// Records errors and set span to error if api response contains errors
				span.RecordError(fmt.Errorf("received api response has errors: %v", apiresp.Error))
				span.SetStatus(codes.Error, codes.Error.String())
			} else {
				// Set span status to ok if api response does not contain business errors
				span.SetStatus(codes.Ok, codes.Ok.String())
			}
		} else {
			// Record an error if no response is available despite no error is reported.
			span.RecordError(fmt.Errorf("no api response is available despite no error has occured"))
			span.SetStatus(codes.Error, codes.Error.String())
		}
	}
}
