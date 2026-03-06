package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

// InitTracer sets up the OpenTelemetry trace pipeline.
// It configures an OTLP HTTP exporter that sends spans to the given endpoint (e.g., Jaeger).
//
// Returns a shutdown function that MUST be deferred in main() to flush pending spans.
//
// Usage:
//
//	shutdown, err := tracer.InitTracer("blog-api", "http://localhost:4318")
//	if err != nil { log.Fatal(err) }
//	defer shutdown()
func InitTracer(serviceName, endpoint string) (func(), error) {
	ctx := context.Background()

	// Create the OTLP HTTP exporter.
	// This sends spans to Jaeger (or any OTLP-compatible collector) via HTTP.
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // No TLS for local development
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create a resource that describes this service.
	// Resources are metadata attached to every span (service name, version, etc.).
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource %w", err)
	}

	// Create the TracerProvider.
	// BatchSpanProcessor batches spans before sending to reduce network overhead.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample 100% for development
	)

	// Register the TracerProvider globally.
	// This allows otel.Tracer("name") to work anywhere in the application.
	otel.SetTracerProvider(tp)

	// Set up W3C TraceContext propagation.
	// This ensures trace IDs are passed between services via HTTP headers.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return a shutdown function that flushes pending spans.
	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := tp.Shutdown(ctx); shutdownErr != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", shutdownErr)
		}
	}

	return shutdown, nil
}
