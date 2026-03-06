package middleware

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracing returns a chi middleware that creates a span for every HTTP request.
// The span includes the HTTP method, path, status code, and user agent.
func Tracing(serviceName string) func(next http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract any incoming trace context from request headers.
			// This allows distributed tracing across multiple services.
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Create the span name from the HTTP method and path.
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

			// Start a new span for this request.
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(r.Method),
					semconv.URLPath(r.URL.Path),
					semconv.UserAgentOriginal(r.UserAgent()),
				),
			)
			defer span.End()

			// Wrap the ResponseWriter to capture the status code.
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call the next handler with the traced context.
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Add the response status code as an attribute.
			span.SetAttributes(
				semconv.HTTPResponseStatusCode(wrapped.statusCode),
			)

			// If the status code indicates an error, mark the span as error.
			if wrapped.statusCode >= 500 {
				span.SetAttributes(attribute.Bool("error", true))
			}
		})
	}
}
