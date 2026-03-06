package middleware

import (
	"net/http"
	"time"

	"github.com/Hapkiduki/blog-api/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
// The standard ResponseWriter doesn't expose the status code after WriteHeader is called.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code before writing it.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures a 200 status code on first write (default behavior of net/http).
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// RequestLogger returns a chi middleware that logs every HTTP request.
// It captures: method, path, status code, latency, and request ID.
// It also injects the logger (with request-scoped fields) into the request context.
func RequestLogger(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get the request ID from chi's RequestID middleware
			reqID := middleware.GetReqID(r.Context())

			// Create a request-scoped logger with the request ID pre-attached.
			// Every log line from this request will include the request ID.
			reqLogger := log.With(
				zap.String("request_id", reqID),
			)

			// Inject the request-scoped logger into the context.
			// Downstream handlers can retrieve it with logger.FromContext(ctx).
			ctx := logger.WithContext(r.Context(), reqLogger)

			// Wrap the ResponseWriter to capture the status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call the next handler
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log the completed request
			latency := time.Since(start)
			reqLogger.Info("request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("latency", latency),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}
