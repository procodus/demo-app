package backend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"procodus.dev/demo-app/pkg/iot"
	"procodus.dev/demo-app/pkg/mq"
)

// Consumer consumes messages from RabbitMQ and persists them to PostgreSQL.
type Consumer struct {
	logger   *slog.Logger
	db       *gorm.DB
	mqClient *mq.Client
	done     chan struct{}
}

// ConsumerConfig holds the configuration for the Consumer.
type ConsumerConfig struct {
	Logger      *slog.Logger
	DB          *gorm.DB
	RabbitMQURL string
	QueueName   string
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(cfg *ConsumerConfig) (*Consumer, error) {
	if cfg == nil {
		return nil, errors.New("consumer config cannot be nil")
	}

	if cfg.Logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	if cfg.DB == nil {
		return nil, errors.New("database cannot be nil")
	}

	if cfg.RabbitMQURL == "" {
		return nil, errors.New("rabbitmq URL cannot be empty")
	}

	if cfg.QueueName == "" {
		return nil, errors.New("queue name cannot be empty")
	}

	// Create MQ client
	mqClient := mq.New(cfg.QueueName, cfg.RabbitMQURL, cfg.Logger)

	return &Consumer{
		logger:   cfg.Logger,
		db:       cfg.DB,
		mqClient: mqClient,
		done:     make(chan struct{}),
	}, nil
}

// Start begins consuming messages from RabbitMQ.
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("starting consumer")

	// Wait for MQ client to be ready
	time.Sleep(2 * time.Second)

	// Start consuming messages
	deliveries, err := c.mqClient.Consume()
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Info("consumer started, waiting for messages")

	// Process messages in a goroutine
	go c.processMessages(ctx, deliveries)

	return nil
}

// processMessages processes incoming messages from the deliveries channel.
func (c *Consumer) processMessages(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("context canceled, stopping message processing")
			close(c.done)
			return

		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.Warn("deliveries channel closed")
				close(c.done)
				return
			}

			c.handleDelivery(ctx, delivery)
		}
	}
}

// handleDelivery processes a single message delivery.
func (c *Consumer) handleDelivery(ctx context.Context, delivery amqp.Delivery) {
	// Parse the protobuf message
	reading := &iot.SensorReading{}
	if err := proto.Unmarshal(delivery.Body, reading); err != nil {
		c.logger.Error("failed to unmarshal sensor reading",
			"error", err,
		)
		// Acknowledge message even on parse error to avoid reprocessing
		if ackErr := delivery.Ack(false); ackErr != nil {
			c.logger.Error("failed to ack message", "error", ackErr)
		}
		return
	}

	// Log the received reading
	c.logger.Info("received sensor reading",
		"device_id", reading.GetDeviceId(),
		"timestamp", reading.GetTimestamp(),
		"temperature", reading.GetTemperature(),
	)

	// Save to database
	if err := c.saveSensorReading(ctx, reading); err != nil {
		c.logger.Error("failed to save sensor reading",
			"device_id", reading.GetDeviceId(),
			"error", err,
		)
		// Nack the message so it can be reprocessed
		if nackErr := delivery.Nack(false, true); nackErr != nil {
			c.logger.Error("failed to nack message", "error", nackErr)
		}
		return
	}

	// Acknowledge successful processing
	if err := delivery.Ack(false); err != nil {
		c.logger.Error("failed to ack message", "error", err)
		return
	}

	c.logger.Debug("sensor reading saved successfully",
		"device_id", reading.GetDeviceId(),
	)
}

// saveSensorReading saves a sensor reading to the database.
func (c *Consumer) saveSensorReading(ctx context.Context, reading *iot.SensorReading) error {
	// Convert protobuf timestamp to time.Time
	timestamp := time.Unix(reading.GetTimestamp(), 0).UTC()

	// Create database model
	dbReading := &SensorReading{
		DeviceID:     reading.GetDeviceId(),
		Timestamp:    timestamp,
		Temperature:  reading.GetTemperature(),
		Humidity:     reading.GetHumidity(),
		Pressure:     reading.GetPressure(),
		BatteryLevel: reading.GetBatteryLevel(),
	}

	// Save to database
	if err := c.db.WithContext(ctx).Create(dbReading).Error; err != nil {
		return fmt.Errorf("failed to create sensor reading: %w", err)
	}

	return nil
}

// Stop stops the consumer and closes the MQ client.
func (c *Consumer) Stop() error {
	c.logger.Info("stopping consumer")

	// Close MQ client
	if err := c.mqClient.Close(); err != nil {
		return fmt.Errorf("failed to close mq client: %w", err)
	}

	// Wait for message processing to complete
	<-c.done

	c.logger.Info("consumer stopped")
	return nil
}
