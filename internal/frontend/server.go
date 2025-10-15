package frontend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"procodus.dev/demo-app/pkg/iot"
)

// Server represents the frontend HTTP server.
type Server struct {
	logger     *slog.Logger
	httpServer *http.Server
	grpcClient iot.IoTServiceClient
	grpcConn   *grpc.ClientConn
	config     *ServerConfig
}

// ServerConfig holds the configuration for the Server.
type ServerConfig struct {
	Logger *slog.Logger

	// HTTP server configuration
	HTTPPort int

	// Backend gRPC configuration
	BackendGRPCAddr string
}

// NewServer creates a new frontend Server instance.
func NewServer(cfg *ServerConfig) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("server config cannot be nil")
	}

	if cfg.Logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	if cfg.HTTPPort <= 0 {
		return nil, errors.New("HTTP port must be positive")
	}

	if cfg.BackendGRPCAddr == "" {
		return nil, errors.New("backend gRPC address cannot be empty")
	}

	return &Server{
		logger: cfg.Logger,
		config: cfg,
	}, nil
}

// Run starts the frontend server and blocks until shutdown.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting frontend server")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Connect to backend gRPC server
	s.logger.Info("connecting to backend gRPC server", "address", s.config.BackendGRPCAddr)
	conn, err := grpc.NewClient(
		s.config.BackendGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to backend: %w", err)
	}
	s.grpcConn = conn
	s.grpcClient = iot.NewIoTServiceClient(conn)

	s.logger.Info("connected to backend gRPC server")

	// Create HTTP router
	mux := s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	s.logger.Info("starting HTTP server", "address", s.httpServer.Addr)

	// Start HTTP server in goroutine
	httpErr := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpErr <- fmt.Errorf("HTTP server error: %w", err)
		}
		close(httpErr)
	}()

	s.logger.Info("frontend server started successfully")

	// Wait for shutdown signal or HTTP error
	select {
	case sig := <-sigChan:
		s.logger.Info("received shutdown signal", "signal", sig.String())
		cancel()
	case <-ctx.Done():
		s.logger.Info("context canceled")
	case err := <-httpErr:
		if err != nil {
			s.logger.Error("HTTP server error", "error", err)
			cancel()
			return err
		}
	}

	// Shutdown
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	s.logger.Info("shutting down frontend server")

	var shutdownErr error

	// Shutdown HTTP server
	if s.httpServer != nil {
		s.logger.Info("stopping HTTP server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("failed to shutdown HTTP server", "error", err)
			shutdownErr = fmt.Errorf("HTTP server shutdown error: %w", err)
		}
		s.logger.Info("HTTP server stopped")
	}

	// Close gRPC connection
	if s.grpcConn != nil {
		s.logger.Info("closing gRPC connection")
		if err := s.grpcConn.Close(); err != nil {
			s.logger.Error("failed to close gRPC connection", "error", err)
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("%w; gRPC connection close error: %w", shutdownErr, err)
			} else {
				shutdownErr = fmt.Errorf("gRPC connection close error: %w", err)
			}
		}
	}

	if shutdownErr != nil {
		s.logger.Error("frontend server shutdown completed with errors", "error", shutdownErr)
		return shutdownErr
	}

	s.logger.Info("frontend server shutdown completed successfully")
	return nil
}

// setupRoutes configures the HTTP routes.
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", s.handleHealth)

	// API endpoints for htmx
	mux.HandleFunc("GET /api/devices", s.handleAPIDevices)
	mux.HandleFunc("GET /api/device/{id}/readings", s.handleAPIDeviceReadings)

	// Main pages
	mux.HandleFunc("GET /devices", s.handleDevices)
	mux.HandleFunc("GET /device/{id}", s.handleDevice)

	// Serve static files (must be before catch-all routes)
	mux.HandleFunc("GET /static/", s.handleStatic)

	// Index page (catch-all, must be last)
	mux.HandleFunc("GET /{$}", s.handleIndex)

	return mux
}
