package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// contextKey is an unexported type used as a key for storing the logger in context.
// Using a custom type prevents collisions with other packages that use context.
type contextKey struct{}

// New creates a new zap.Logger configured for the given log level.
// The level is read from the LOG_LEVEL environment variable.
// Valid levels: debug, info, warn, error, dpanic, panic, fatal.
// Defaults to "info" if not set or invalid.
func New() (*zap.Logger, error) {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	// Parse the log level string into a zapcore.Level
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		// If the level is invalid, default to info and log a warning
		zapLevel = zapcore.InfoLevel
	}

	// Configure the encoder (how logs are formatted)
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.TimeKey = "ts"
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Human-readable timestamps
	encodeConfig.EncodeDuration = zapcore.MillisDurationEncoder // Durations in ms
	encodeConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// Build the core: JSON encoder, writing to stdout, at the configured level
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encodeConfig),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Create the logger with caller info (file:line) and stacktrace on errors
	log := zap.New(core,
		zap.AddCaller(),                       // Adds "caller": "main.go:42"
		zap.AddStacktrace(zapcore.ErrorLevel)) // Stack traces only for error+

	return log, nil
}

// WithContext returns a new context with the logger attached.
func WithContext(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns a no-op logger (never nil).
func FromContext(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(contextKey{}).(*zap.Logger); ok {
		return log
	}
	return zap.NewNop() // Return a no-op logger to avoid nil pointer panics
}
