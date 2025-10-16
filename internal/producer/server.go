package producer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"procodus.dev/demo-app/pkg/metrics"
	"procodus.dev/demo-app/pkg/mq"
)

// ServerConfig holds the configuration for the producer server.
type ServerConfig struct {
	// Logger is the structured logger
	Logger *slog.Logger
	// RabbitMQURL is the connection string for RabbitMQ
	RabbitMQURL string
	// QueueName is the name of the queue to publish sensor readings to
	QueueName string
	// DeviceQueueName is the name of the queue to publish device creation messages to
	DeviceQueueName string
	// Interval is the time between data point generation
	Interval time.Duration
	// ProducerCount is the number of concurrent producers
	ProducerCount int
	// Metrics is the optional Prometheus metrics collector
	Metrics *metrics.ProducerMetrics
	// MQMetrics is the optional Prometheus metrics collector for MQ operations
	MQMetrics *metrics.MQMetrics
}

// Server manages multiple producer instances.
type Server struct {
	logger        *slog.Logger
	config        *ServerConfig
	producers     []*Producer
	clients       []*mq.Client
	deviceClients []*mq.Client
	wg            sync.WaitGroup
	metrics       *metrics.ProducerMetrics
}

var (
	errInvalidProducerCount = errors.New("producer count must be greater than 0")
	errInvalidInterval      = errors.New("interval must be greater than 0")
	errLoggerRequired       = errors.New("logger is required")
)

// NewServer creates a new producer server with the given configuration.
func NewServer(cfg *ServerConfig) (*Server, error) {
	if cfg.ProducerCount <= 0 {
		return nil, errInvalidProducerCount
	}

	if cfg.Interval <= 0 {
		return nil, errInvalidInterval
	}

	if cfg.Logger == nil {
		return nil, errLoggerRequired
	}

	s := &Server{
		config:        cfg,
		producers:     make([]*Producer, 0, cfg.ProducerCount),
		clients:       make([]*mq.Client, 0, cfg.ProducerCount),
		deviceClients: make([]*mq.Client, 0, cfg.ProducerCount),
		logger:        cfg.Logger,
		metrics:       cfg.Metrics,
	}

	// Create producer instances with their own MQ clients
	for i := 0; i < cfg.ProducerCount; i++ {
		// Create MQ client for sensor readings
		client := mq.New(cfg.QueueName, cfg.RabbitMQURL, cfg.Logger.With(
			slog.String("component", "mq-client"),
			slog.Int("producer_id", i),
		))

		// Enable MQ metrics if configured
		if cfg.MQMetrics != nil {
			client.SetMetrics(cfg.MQMetrics)
		}

		// Create MQ client for device creation messages
		deviceClient := mq.New(cfg.DeviceQueueName, cfg.RabbitMQURL, cfg.Logger.With(
			slog.String("component", "device-mq-client"),
			slog.Int("producer_id", i),
		))

		// Enable MQ metrics if configured
		if cfg.MQMetrics != nil {
			deviceClient.SetMetrics(cfg.MQMetrics)
		}

		// Create producer with both clients
		producer := NewProducer(client, deviceClient)

		// Enable producer metrics if configured
		if cfg.Metrics != nil {
			producer.SetMetrics(cfg.Metrics)
		}

		s.clients = append(s.clients, client)
		s.deviceClients = append(s.deviceClients, deviceClient)
		s.producers = append(s.producers, producer)

		s.logger.Info("created producer instance",
			"producer_id", i,
			"queue", cfg.QueueName,
			"device_queue", cfg.DeviceQueueName,
			"device_count", len(producer.IoTDevices),
		)
	}

	return s, nil
}

// Run starts all producers and blocks until shutdown signal is received.
func (s *Server) Run(ctx context.Context) error {
	// Create context that can be canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start all producers
	for i, producer := range s.producers {
		s.wg.Add(1)
		go s.runProducer(ctx, i, producer)
	}

	s.logger.Info("producer server started",
		"producer_count", len(s.producers),
		"interval", s.config.Interval,
	)

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		s.logger.Info("received shutdown signal", "signal", sig.String())
		cancel()
	case <-ctx.Done():
		s.logger.Info("context canceled, shutting down")
	}

	// Wait for all producers to finish
	s.logger.Info("waiting for producers to shut down...")
	s.wg.Wait()

	// Close all MQ clients
	s.logger.Info("closing MQ clients...")
	s.closeClients()

	s.logger.Info("producer server stopped")
	return nil
}

// runProducer runs a single producer instance, generating data points at configured intervals.
func (s *Server) runProducer(ctx context.Context, id int, producer *Producer) {
	defer s.wg.Done()

	// Track active producer
	if s.metrics != nil {
		s.metrics.ActiveProducers.Inc()
		defer s.metrics.ActiveProducers.Dec()
	}

	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	producerLogger := s.logger.With(slog.Int("producer_id", id))
	producerLogger.Info("producer started")

	for {
		select {
		case <-ctx.Done():
			producerLogger.Info("producer shutting down")
			return

		case <-ticker.C:
			if err := producer.RandomDataPoint(ctx); err != nil {
				producerLogger.Error("failed to generate data point",
					"error", err,
				)
				// Continue on error - don't stop the producer
				continue
			}

			producerLogger.Debug("data point generated and sent")
		}
	}
}

// closeClients closes all MQ clients gracefully.
func (s *Server) closeClients() {
	var wg sync.WaitGroup

	// Close sensor reading clients
	for i, client := range s.clients {
		wg.Add(1)
		go func(id int, c *mq.Client) {
			defer wg.Done()

			if err := c.Close(); err != nil {
				s.logger.Error("failed to close MQ client",
					"producer_id", id,
					"error", err,
				)
				return
			}

			s.logger.Info("MQ client closed", "producer_id", id)
		}(i, client)
	}

	// Close device clients
	for i, deviceClient := range s.deviceClients {
		wg.Add(1)
		go func(id int, c *mq.Client) {
			defer wg.Done()

			if err := c.Close(); err != nil {
				s.logger.Error("failed to close device MQ client",
					"producer_id", id,
					"error", err,
				)
				return
			}

			s.logger.Info("device MQ client closed", "producer_id", id)
		}(i, deviceClient)
	}

	wg.Wait()
}

// Shutdown initiates a graceful shutdown of the server.
// This is an alternative to sending OS signals.
func (s *Server) Shutdown() error {
	s.logger.Info("shutdown requested")

	// Close all MQ clients
	s.closeClients()

	return nil
}
