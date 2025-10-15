package backend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"procodus.dev/demo-app/pkg/iot"
)

// Server represents the backend server that manages database, message queue, and gRPC.
type Server struct {
	logger     *slog.Logger
	db         *gorm.DB
	consumer   *Consumer
	grpcServer *grpc.Server
	config     *ServerConfig
}

// ServerConfig holds the configuration for the Server.
type ServerConfig struct {
	Logger *slog.Logger

	// Database configuration
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// RabbitMQ configuration
	RabbitMQURL string
	QueueName   string

	// gRPC configuration
	GRPCPort int

	// Database port
	DBPort int
}

// NewServer creates a new Server instance.
func NewServer(cfg *ServerConfig) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("server config cannot be nil")
	}

	if cfg.Logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	if cfg.RabbitMQURL == "" {
		return nil, errors.New("rabbitmq URL cannot be empty")
	}

	if cfg.QueueName == "" {
		return nil, errors.New("queue name cannot be empty")
	}

	if cfg.DBHost == "" {
		return nil, errors.New("database host cannot be empty")
	}

	if cfg.DBPort <= 0 {
		return nil, errors.New("database port must be positive")
	}

	if cfg.DBUser == "" {
		return nil, errors.New("database user cannot be empty")
	}

	if cfg.DBName == "" {
		return nil, errors.New("database name cannot be empty")
	}

	if cfg.GRPCPort <= 0 {
		return nil, errors.New("gRPC port must be positive")
	}

	return &Server{
		logger: cfg.Logger,
		config: cfg,
	}, nil
}

// Run starts the backend server and blocks until shutdown.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting backend server")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Initialize database
	dbCfg := &DBConfig{
		Host:     s.config.DBHost,
		Port:     s.config.DBPort,
		User:     s.config.DBUser,
		Password: s.config.DBPassword,
		DBName:   s.config.DBName,
		SSLMode:  s.config.DBSSLMode,
		Logger:   s.logger,
	}

	db, err := NewDB(dbCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	s.db = db

	s.logger.Info("database initialized successfully")

	// Initialize consumer
	consumerCfg := &ConsumerConfig{
		Logger:      s.logger,
		DB:          s.db,
		RabbitMQURL: s.config.RabbitMQURL,
		QueueName:   s.config.QueueName,
	}

	consumer, err := NewConsumer(consumerCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize consumer: %w", err)
	}
	s.consumer = consumer

	// Start consumer
	if err := s.consumer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	// Initialize gRPC service
	iotService, err := NewIoTService(s.logger, s.db)
	if err != nil {
		return fmt.Errorf("failed to initialize gRPC service: %w", err)
	}

	// Create gRPC server
	s.grpcServer = grpc.NewServer()
	iot.RegisterIoTServiceServer(s.grpcServer, iotService)

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", s.config.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcAddr, err)
	}

	s.logger.Info("starting gRPC server", "address", grpcAddr)

	// Start gRPC server in goroutine
	grpcErr := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			grpcErr <- fmt.Errorf("gRPC server error: %w", err)
		}
		close(grpcErr)
	}()

	s.logger.Info("backend server started successfully")

	// Wait for shutdown signal or gRPC error
	select {
	case sig := <-sigChan:
		s.logger.Info("received shutdown signal", "signal", sig.String())
		cancel()
	case <-ctx.Done():
		s.logger.Info("context canceled")
	case err := <-grpcErr:
		if err != nil {
			s.logger.Error("gRPC server error", "error", err)
			cancel()
			return err
		}
	}

	// Shutdown
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	s.logger.Info("shutting down backend server")

	var shutdownErr error

	// Stop gRPC server
	if s.grpcServer != nil {
		s.logger.Info("stopping gRPC server")
		s.grpcServer.GracefulStop()
		s.logger.Info("gRPC server stopped")
	}

	// Stop consumer
	if s.consumer != nil {
		s.logger.Info("stopping consumer")
		if err := s.consumer.Stop(); err != nil {
			s.logger.Error("failed to stop consumer", "error", err)
			shutdownErr = fmt.Errorf("consumer shutdown error: %w", err)
		}
	}

	// Close database
	if s.db != nil {
		s.logger.Info("closing database connection")
		if err := CloseDB(s.db, s.logger); err != nil {
			s.logger.Error("failed to close database", "error", err)
			if shutdownErr != nil {
				shutdownErr = fmt.Errorf("%w; database close error: %w", shutdownErr, err)
			} else {
				shutdownErr = fmt.Errorf("database close error: %w", err)
			}
		}
	}

	if shutdownErr != nil {
		s.logger.Error("backend server shutdown completed with errors", "error", shutdownErr)
		return shutdownErr
	}

	s.logger.Info("backend server shutdown completed successfully")
	return nil
}
