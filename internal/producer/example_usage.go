package producer

// This file provides example usage of the producer server.
// DO NOT include this in builds - it's for documentation only.

import (
	"context"
	"time"

	"procodus.dev/demo-app/pkg/logger"
)

// ExampleServerUsage demonstrates how to create and run the producer server.
func ExampleServerUsage() {
	// Create logger
	log := logger.NewWithLevel(logger.ParseLevel("info"))

	// Create server configuration
	config := &ServerConfig{
		Logger:        log,
		RabbitMQURL:   "amqp://guest:guest@localhost:5672",
		QueueName:     "iot-sensor-data",
		ProducerCount: 5,               // 5 concurrent producers
		Interval:      5 * time.Second, // Generate data every 5 seconds
	}

	// Create server
	server, err := NewServer(config)
	if err != nil {
		log.Error("failed to create server", "error", err)
		return
	}

	// Run server (blocks until shutdown signal)
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Error("server error", "error", err)
	}
}

// ExampleServerWithCustomInterval shows how to use a custom data generation interval.
func ExampleServerWithCustomInterval() {
	log := logger.NewDefault()

	config := &ServerConfig{
		Logger:        log,
		RabbitMQURL:   "amqp://localhost:5672",
		QueueName:     "sensor-readings",
		ProducerCount: 3,
		Interval:      10 * time.Second, // Generate data every 10 seconds
	}

	server, err := NewServer(config)
	if err != nil {
		log.Error("failed to create server", "error", err)
		return
	}

	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Error("server error", "error", err)
	}
}

// ExampleServerProgrammaticShutdown shows how to shutdown the server programmatically.
func ExampleServerProgrammaticShutdown() {
	log := logger.NewDefault()

	config := &ServerConfig{
		Logger:        log,
		RabbitMQURL:   "amqp://localhost:5672",
		QueueName:     "test-queue",
		ProducerCount: 2,
		Interval:      1 * time.Second,
	}

	server, err := NewServer(config)
	if err != nil {
		log.Error("failed to create server", "error", err)
		return
	}

	// Create context with timeout for automatic shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run server (will shutdown after 30 seconds or on signal)
	if err := server.Run(ctx); err != nil {
		log.Error("server error", "error", err)
	}
}
