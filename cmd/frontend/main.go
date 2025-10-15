package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"procodus.dev/demo-app/internal/frontend"
)

func main() {
	// Parse command-line flags
	httpPort := flag.Int("http-port", 8080, "HTTP server port")
	backendAddr := flag.String("backend-addr", "localhost:9090", "Backend gRPC server address")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Set up logger
	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	// Create server configuration
	config := &frontend.ServerConfig{
		Logger:          logger,
		HTTPPort:        *httpPort,
		BackendGRPCAddr: *backendAddr,
	}

	// Create server
	server, err := frontend.NewServer(config)
	if err != nil {
		logger.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	// Run server
	logger.Info("starting frontend server",
		"http_port", *httpPort,
		"backend_addr", *backendAddr,
	)

	if err := server.Run(context.Background()); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("frontend server stopped")
}
