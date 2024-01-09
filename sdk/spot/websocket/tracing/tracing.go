package tracing

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Package name used as instrumentation ID
	PackageName = "goctopus_sdk_spot_websocket"
	// Package version
	PackageVersion = "0.0.0"
	// Span & events namespace
	TracesNamespace = "goctopus.spot.websocket"
)

// A helper function to trace an error (if any) and set the provided span status
// accordingly.
//
// The function is meant to replace code blocks like this:
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
	// Panic if provided span is nil
	if span == nil {
		panic("provided span must not be nil.")
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, codes.Error.String())
	} else {
		span.SetStatus(codes.Ok, codes.Ok.String())
	}
}

// A helper function to trace an error (if any), set the provided span status
// accordingly and return the input error.
//
// The function is meant to replace code blocks like this:
//
//	if err != nil {
//		span.RecordError(err)
//		span.SetStatus(codes.Error, codes.Error.String())
//		return err
//	}
//	span.SetStatus(codes.Ok, codes.Ok.String())
//	return nil
//
// By this:
//
// return HandleAndTraceError(span, err)
func HandleAndTraceError(span trace.Span, err error) error {
	// Panic if provided span is nil
	if span == nil {
		panic("provided span must not be nil.")
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, codes.Error.String())
	} else {
		span.SetStatus(codes.Ok, codes.Ok.String())
	}
	return err
}
