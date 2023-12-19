package rest

import (
	"context"
	"net/http"

	"github.com/gbdevw/purple-goctopus/spot/rest/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// A decorator for KrakenSpotRESTClientAuthorizer which manages code instrumentation by using the
// OpenTelemetry framework.
type KrakenSpotRESTClientAuthorizerInstrumentationDecorator struct {
	// Decorated
	decorated KrakenSpotRESTClientAuthorizerIface
	// Tracer
	tracer trace.Tracer
}

// # Description
//
// Decorate the provided KrakenSpotRESTClientAuthorizerIface implementation. The function returns the decorator which
// manages tracing and code instrumentation of the decorated by using the OpenTelemetry framework.
//
// # Inputs
//
//   - decorated: The KrakenSpotRESTClientAuthorizerIface implentation to decorate. Must no be nil.
//   - tracerProvider: Tracer provider to use to get the tracer used by the decorator to instrument code. If nil, the global tracer provider will be used (can be a NoopTracerProvider).
//
// # Returns
//
// The decorator which decorates the provided KrakenSpotRESTClientAuthorizerIface implementation.
func DecorateKrakenSpotRESTClientAuthorizer(decorated KrakenSpotRESTClientAuthorizerIface, tracerProvider trace.TracerProvider) KrakenSpotRESTClientAuthorizerIface {
	if decorated == nil {
		// Panic if decorated is nil
		panic("decorated cannot be nil")
	}
	if tracerProvider == nil {
		// Use the global tracer provider if the provided tracer provider is nil.
		// In case the global tracer provider is not configured, its default behavior is to return a NoopTracerProvider.
		tracerProvider = otel.GetTracerProvider()
	}
	// Return decorator
	return &KrakenSpotRESTClientAuthorizerInstrumentationDecorator{
		decorated: decorated,
		tracer:    tracerProvider.Tracer(tracing.PackageName, trace.WithInstrumentationVersion(tracing.PackageVersion)),
	}
}

// Instrument the decorated Authorize method.
func (dec *KrakenSpotRESTClientAuthorizerInstrumentationDecorator) Authorize(ctx context.Context, req *http.Request) (*http.Request, error) {
	// Panic if provided request is nil
	if req == nil {
		panic("provided request must not be nil.")
	}
	// Start a span
	ctx, span := dec.tracer.Start(ctx, "authorize", trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			// Do not trace otp nor the form body value nor the headers
			// risk of sensitive informations leak
			attribute.String("path", req.URL.Path),
			attribute.String("nonce", req.FormValue("nonce")),
		))
	defer span.End()
	// Call decorated
	oreq, err := dec.decorated.Authorize(ctx, req)
	// Trace error and set span status
	tracing.TraceErrorAndSetStatus(span, err)
	// Return results
	return oreq, err
}
