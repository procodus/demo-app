// Package main provides the unified CLI entry point for the demo-app services.
package main

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"procodus.dev/demo-app/internal/producer"
)

var generatorCmd = &cobra.Command{
	Use:   "generator",
	Short: "Run the data generator",
	Long: `Run the data generator that:
- Generates synthetic IoT sensor readings
- Publishes sensor data to RabbitMQ
- Publishes device creation messages to RabbitMQ
- Supports multiple concurrent producers`,
	RunE: runGenerator,
}

func init() {
	rootCmd.AddCommand(generatorCmd)

	// Generator-specific flags
	generatorCmd.Flags().String("rabbitmq-url", "amqp://localhost:5672", "RabbitMQ URL")
	generatorCmd.Flags().String("queue-name", "sensor-data", "RabbitMQ queue name for sensor readings")
	generatorCmd.Flags().String("device-queue-name", "device-data", "RabbitMQ queue name for device creation messages")
	generatorCmd.Flags().Int("producer-count", 5, "Number of concurrent producers")
	generatorCmd.Flags().Duration("interval", 5*time.Second, "Interval between data generation")

	// Bind flags to viper
	if err := viper.BindPFlag("generator.rabbitmq.url", generatorCmd.Flags().Lookup("rabbitmq-url")); err != nil {
		log.Fatalf("failed to bind rabbitmq-url flag: %v", err)
	}
	if err := viper.BindPFlag("generator.rabbitmq.queue_name", generatorCmd.Flags().Lookup("queue-name")); err != nil {
		log.Fatalf("failed to bind queue-name flag: %v", err)
	}
	if err := viper.BindPFlag("generator.rabbitmq.device_queue_name", generatorCmd.Flags().Lookup("device-queue-name")); err != nil {
		log.Fatalf("failed to bind device-queue-name flag: %v", err)
	}
	if err := viper.BindPFlag("generator.producer_count", generatorCmd.Flags().Lookup("producer-count")); err != nil {
		log.Fatalf("failed to bind producer-count flag: %v", err)
	}
	if err := viper.BindPFlag("generator.interval", generatorCmd.Flags().Lookup("interval")); err != nil {
		log.Fatalf("failed to bind interval flag: %v", err)
	}
}

func runGenerator(_ *cobra.Command, _ []string) error {
	logger := GetLogger()
	logger.Info("starting generator service")

	// Create producer configuration from viper
	config := &producer.ServerConfig{
		Logger:          logger,
		RabbitMQURL:     viper.GetString("generator.rabbitmq.url"),
		QueueName:       viper.GetString("generator.rabbitmq.queue_name"),
		DeviceQueueName: viper.GetString("generator.rabbitmq.device_queue_name"),
		ProducerCount:   viper.GetInt("generator.producer_count"),
		Interval:        viper.GetDuration("generator.interval"),
	}

	// Create and run server
	server, err := producer.NewServer(config)
	if err != nil {
		logger.Error("failed to create generator server", "error", err)
		return err
	}

	logger.Info("generator server configuration",
		"rabbitmq_url", config.RabbitMQURL,
		"sensor_queue", config.QueueName,
		"device_queue", config.DeviceQueueName,
		"producer_count", config.ProducerCount,
		"interval", config.Interval,
	)

	if err := server.Run(context.Background()); err != nil {
		logger.Error("generator server error", "error", err)
		return err
	}

	logger.Info("generator server stopped")
	return nil
}
