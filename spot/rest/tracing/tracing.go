package tracing

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Package name used as instrumentation ID
	PackageName = "goctopus_sdk_spot_rest"
	// Package version
	PackageVersion = "0.0.0"
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
