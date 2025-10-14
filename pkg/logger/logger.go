// Package logger provides a shared structured logging implementation using slog.
package logger

import (
	"io"
	"log/slog"
	"os"
)

// Config holds the configuration for the logger.
type Config struct {
	// Output is the writer to send logs to (defaults to os.Stdout).
	Output io.Writer
	// Level is the minimum log level to output.
	Level slog.Level
	// AddSource adds source code position to log records.
	AddSource bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Level:     slog.LevelInfo,
		Output:    os.Stdout,
		AddSource: false,
	}
}

// New creates a new JSON logger with the provided configuration.
func New(cfg *Config) *slog.Logger {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	// Create JSON handler
	handler := slog.NewJSONHandler(cfg.Output, opts)

	// Create and return logger
	return slog.New(handler)
}

// NewDefault creates a new JSON logger with default configuration.
func NewDefault() *slog.Logger {
	return New(DefaultConfig())
}

// NewWithLevel creates a new JSON logger with the specified log level.
func NewWithLevel(level slog.Level) *slog.Logger {
	cfg := DefaultConfig()
	cfg.Level = level
	return New(cfg)
}

// ParseLevel converts a string to a slog.Level.
// Supported values: "debug", "info", "warn", "error".
// Returns slog.LevelInfo if the level string is not recognized.
func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext returns a new logger with the provided context fields.
// Fields persist across all subsequent log messages.
func WithContext(logger *slog.Logger, attrs ...slog.Attr) *slog.Logger {
	// Convert []slog.Attr to []any for logger.With()
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return logger.With(args...)
}
