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

// DeviceConsumer consumes device creation messages from RabbitMQ and persists them to PostgreSQL.
type DeviceConsumer struct {
	logger   *slog.Logger
	db       *gorm.DB
	mqClient mq.ClientInterface
	done     chan struct{}
}

// DeviceConsumerConfig holds the configuration for the DeviceConsumer.
type DeviceConsumerConfig struct {
	Logger      *slog.Logger
	DB          *gorm.DB
	RabbitMQURL string
	QueueName   string
}

// NewDeviceConsumer creates a new DeviceConsumer instance.
func NewDeviceConsumer(cfg *DeviceConsumerConfig) (*DeviceConsumer, error) {
	if cfg == nil {
		return nil, errors.New("device consumer config cannot be nil")
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

	return &DeviceConsumer{
		logger:   cfg.Logger,
		db:       cfg.DB,
		mqClient: mqClient,
		done:     make(chan struct{}),
	}, nil
}

// Start begins consuming device messages from RabbitMQ.
func (c *DeviceConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting device consumer")

	// Wait for MQ client to be ready
	time.Sleep(2 * time.Second)

	// Start consuming messages
	deliveries, err := c.mqClient.Consume()
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Info("device consumer started, waiting for messages")

	// Process messages in a goroutine
	go c.processMessages(ctx, deliveries)

	return nil
}

// processMessages processes incoming device messages from the deliveries channel.
func (c *DeviceConsumer) processMessages(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("context canceled, stopping device message processing")
			close(c.done)
			return

		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.Warn("device deliveries channel closed")
				close(c.done)
				return
			}

			c.handleDelivery(ctx, delivery)
		}
	}
}

// handleDelivery processes a single device message delivery.
func (c *DeviceConsumer) handleDelivery(ctx context.Context, delivery amqp.Delivery) {
	// Parse the protobuf message
	device := &iot.IoTDevice{}
	if err := proto.Unmarshal(delivery.Body, device); err != nil {
		c.logger.Error("failed to unmarshal device message",
			"error", err,
		)
		// Acknowledge message even on parse error to avoid reprocessing
		if ackErr := delivery.Ack(false); ackErr != nil {
			c.logger.Error("failed to ack message", "error", ackErr)
		}
		return
	}

	// Log the received device
	c.logger.Info("received device message",
		"device_id", device.GetDeviceId(),
		"location", device.GetLocation(),
	)

	// Save to database
	if err := c.saveIoTDevice(ctx, device); err != nil {
		c.logger.Error("failed to save device",
			"device_id", device.GetDeviceId(),
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

	c.logger.Debug("device saved successfully",
		"device_id", device.GetDeviceId(),
	)
}

// saveIoTDevice saves an IoT device to the database using upsert logic.
func (c *DeviceConsumer) saveIoTDevice(ctx context.Context, device *iot.IoTDevice) error {
	// Convert protobuf timestamp to time.Time
	timestamp := time.Unix(device.GetTimestamp(), 0).UTC()

	// Create database model
	dbDevice := &IoTDevice{
		DeviceID:   device.GetDeviceId(),
		Location:   device.GetLocation(),
		MACAddress: device.GetMacAddress(),
		IPAddress:  device.GetIpAddress(),
		Firmware:   device.GetFirmware(),
		LastSeen:   timestamp,
		Latitude:   device.GetLatitude(),
		Longitude:  device.GetLongitude(),
	}

	// Use upsert logic: create if not exists, update if exists
	// This handles the case where a device message might be received multiple times
	result := c.db.WithContext(ctx).
		Where("device_id = ?", dbDevice.DeviceID).
		Assign(map[string]interface{}{
			"location":    dbDevice.Location,
			"mac_address": dbDevice.MACAddress,
			"ip_address":  dbDevice.IPAddress,
			"firmware":    dbDevice.Firmware,
			"last_seen":   dbDevice.LastSeen,
			"latitude":    dbDevice.Latitude,
			"longitude":   dbDevice.Longitude,
		}).
		FirstOrCreate(dbDevice)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert device: %w", result.Error)
	}

	return nil
}

// Stop stops the device consumer and closes the MQ client.
func (c *DeviceConsumer) Stop() error {
	c.logger.Info("stopping device consumer")

	// Close MQ client
	if err := c.mqClient.Close(); err != nil {
		return fmt.Errorf("failed to close mq client: %w", err)
	}

	// Wait for message processing to complete
	<-c.done

	c.logger.Info("device consumer stopped")
	return nil
}
