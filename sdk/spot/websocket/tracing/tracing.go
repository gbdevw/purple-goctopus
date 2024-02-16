package tracing

import (
	"log"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Package name used as instrumentation ID
	PackageName = "goctopus.sdk.spot.websocket"
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

// A helper function to trace and log an error (if any), set the provided span status
// accordingly and return the input error.
//
// The function is meant to replace code blocks like this:
//
//	if err != nil {
//		logger.Println(err)
//		span.RecordError(err)
//		span.SetStatus(codes.Error, codes.Error.String())
//		return err
//	}
//	span.SetStatus(codes.Ok, codes.Ok.String())
//	return nil
//
// By this:
//
// return HandleAndTraLogError(span, err)
func HandleAndTraLogError(span trace.Span, logger *log.Logger, err error) error {
	// Panic if provided span or logger is nil
	if span == nil || logger == nil {
		panic("provided span and logger must not be nil.")
	}
	if err != nil {
		log.Println(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, codes.Error.String())
	} else {
		span.SetStatus(codes.Ok, codes.Ok.String())
	}
	return err
}
