package frontend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"procodus.dev/demo-app/pkg/iot"
	"procodus.dev/demo-app/pkg/metrics"
)

// Server represents the frontend HTTP server.
type Server struct {
	logger     *slog.Logger
	httpServer *http.Server
	grpcClient iot.IoTServiceClient
	grpcConn   *grpc.ClientConn
	config     *ServerConfig
	metrics    *metrics.FrontendMetrics // Optional metrics
}

// ServerConfig holds the configuration for the Server.
type ServerConfig struct {
	// Backend gRPC configuration
	BackendGRPCAddr string

	Logger *slog.Logger

	// HTTP server configuration
	HTTPPort int

	// Metrics configuration (optional)
	Metrics *metrics.FrontendMetrics
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
		logger:  cfg.Logger,
		config:  cfg,
		metrics: cfg.Metrics,
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

	// Shutdown with timeout context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	//nolint:contextcheck // Intentionally creating new context for shutdown with timeout
	return s.Shutdown(shutdownCtx)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down frontend server")

	var shutdownErr error

	// Shutdown HTTP server
	if s.httpServer != nil {
		s.logger.Info("stopping HTTP server")
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
func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", s.handleHealth)

	// Prometheus metrics endpoint (if metrics enabled)
	if s.metrics != nil {
		mux.Handle("GET /metrics", metrics.Handler())
	}

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

	// Wrap with metrics middleware if metrics are enabled
	if s.metrics != nil {
		return s.metricsMiddleware(mux)
	}

	return mux
}

// metricsMiddleware wraps HTTP handlers with Prometheus metrics tracking.
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track in-flight requests
		s.metrics.HTTPRequestsInFlight.WithLabelValues(r.Method, r.URL.Path).Inc()
		defer s.metrics.HTTPRequestsInFlight.WithLabelValues(r.Method, r.URL.Path).Dec()

		// Track duration
		timer := prometheus.NewTimer(s.metrics.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path))
		defer timer.ObserveDuration()

		// Create response writer wrapper to capture status code and size
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Track request completion
		s.metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode)).Inc()
		s.metrics.HTTPResponseSize.WithLabelValues(r.URL.Path).Observe(float64(rw.bytesWritten))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures bytes written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// callGetAllDevice wraps gRPC GetAllDevice call with metrics.
func (s *Server) callGetAllDevice(ctx context.Context, req *iot.GetAllDevicesRequest) (*iot.GetAllDevicesResponse, error) {
	if s.metrics == nil {
		return s.grpcClient.GetAllDevice(ctx, req)
	}

	// Track duration
	timer := prometheus.NewTimer(s.metrics.GRPCClientDuration.WithLabelValues("GetAllDevice"))
	defer timer.ObserveDuration()

	// Make the call
	resp, err := s.grpcClient.GetAllDevice(ctx, req)

	// Track result
	if err != nil {
		s.metrics.GRPCClientCalls.WithLabelValues("GetAllDevice", "error").Inc()
		// Categorize error type
		if st, ok := status.FromError(err); ok {
			s.metrics.GRPCClientErrors.WithLabelValues("GetAllDevice", st.Code().String()).Inc()
		} else {
			s.metrics.GRPCClientErrors.WithLabelValues("GetAllDevice", "unknown").Inc()
		}
		return nil, err
	}

	s.metrics.GRPCClientCalls.WithLabelValues("GetAllDevice", "success").Inc()
	return resp, nil
}

// callGetDevice wraps gRPC GetDevice call with metrics.
func (s *Server) callGetDevice(ctx context.Context, req *iot.GetDeviceByIDRequest) (*iot.GetDeviceByIDResponse, error) {
	if s.metrics == nil {
		return s.grpcClient.GetDevice(ctx, req)
	}

	// Track duration
	timer := prometheus.NewTimer(s.metrics.GRPCClientDuration.WithLabelValues("GetDevice"))
	defer timer.ObserveDuration()

	// Make the call
	resp, err := s.grpcClient.GetDevice(ctx, req)

	// Track result
	if err != nil {
		s.metrics.GRPCClientCalls.WithLabelValues("GetDevice", "error").Inc()
		// Categorize error type
		if st, ok := status.FromError(err); ok {
			s.metrics.GRPCClientErrors.WithLabelValues("GetDevice", st.Code().String()).Inc()
		} else {
			s.metrics.GRPCClientErrors.WithLabelValues("GetDevice", "unknown").Inc()
		}
		return nil, err
	}

	s.metrics.GRPCClientCalls.WithLabelValues("GetDevice", "success").Inc()
	return resp, nil
}

// callGetSensorReadingByDeviceID wraps gRPC GetSensorReadingByDeviceID call with metrics.
func (s *Server) callGetSensorReadingByDeviceID(ctx context.Context, req *iot.GetSensorReadingByDeviceIDRequest) (*iot.GetSensorReadingByDeviceIDResponse, error) {
	if s.metrics == nil {
		return s.grpcClient.GetSensorReadingByDeviceID(ctx, req)
	}

	// Track duration
	timer := prometheus.NewTimer(s.metrics.GRPCClientDuration.WithLabelValues("GetSensorReadingByDeviceID"))
	defer timer.ObserveDuration()

	// Make the call
	resp, err := s.grpcClient.GetSensorReadingByDeviceID(ctx, req)

	// Track result
	if err != nil {
		s.metrics.GRPCClientCalls.WithLabelValues("GetSensorReadingByDeviceID", "error").Inc()
		// Categorize error type
		if st, ok := status.FromError(err); ok {
			s.metrics.GRPCClientErrors.WithLabelValues("GetSensorReadingByDeviceID", st.Code().String()).Inc()
		} else {
			s.metrics.GRPCClientErrors.WithLabelValues("GetSensorReadingByDeviceID", "unknown").Inc()
		}
		return nil, err
	}

	s.metrics.GRPCClientCalls.WithLabelValues("GetSensorReadingByDeviceID", "success").Inc()
	return resp, nil
}
