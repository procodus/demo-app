package logger_test

import (
	"log/slog"
	"os"

	"procodus.dev/demo-app/pkg/logger"
)

func ExampleNew() {
	// Create a logger with custom configuration.
	cfg := &logger.Config{
		Level:  slog.LevelDebug,
		Output: os.Stdout,
	}
	log := logger.New(cfg)

	log.Debug("debug message")
	log.Info("info message")
}

func ExampleNewDefault() {
	// Create a logger with default configuration (Info level, stdout).
	log := logger.NewDefault()

	log.Info("application started", "version", "1.0.0")
}

func ExampleNewWithLevel() {
	// Create a logger with a specific log level.
	log := logger.NewWithLevel(slog.LevelWarn)

	// This will not be logged (below Warn level).
	log.Info("this won't appear")

	// This will be logged.
	log.Warn("warning message")
}

func ExampleParseLevel() {
	// Parse log level from string (useful for configuration).
	level := logger.ParseLevel("debug")

	log := logger.NewWithLevel(level)
	log.Debug("debug enabled")
}

func ExampleWithContext() {
	// Create a logger with contextual fields that appear in all log messages.
	baseLogger := logger.NewDefault()

	// Add context fields
	requestLogger := logger.WithContext(baseLogger,
		slog.String("request_id", "req-123"),
		slog.String("user_id", "user-456"),
	)

	// All logs will include request_id and user_id.
	requestLogger.Info("processing request")
	requestLogger.Info("request completed")
}
