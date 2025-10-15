package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"procodus.dev/demo-app/internal/backend"
)

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Run the backend server",
	Long: `Run the backend server that:
- Consumes sensor readings from RabbitMQ
- Consumes device creation messages from RabbitMQ
- Persists data to PostgreSQL
- Serves gRPC API endpoints`,
	RunE: runBackend,
}

func init() {
	rootCmd.AddCommand(backendCmd)

	// Backend-specific flags
	backendCmd.Flags().String("db-host", "localhost", "PostgreSQL host")
	backendCmd.Flags().Int("db-port", 5432, "PostgreSQL port")
	backendCmd.Flags().String("db-user", "postgres", "PostgreSQL user")
	backendCmd.Flags().String("db-password", "", "PostgreSQL password")
	backendCmd.Flags().String("db-name", "iot", "PostgreSQL database name")
	backendCmd.Flags().String("db-sslmode", "disable", "PostgreSQL SSL mode")
	backendCmd.Flags().String("rabbitmq-url", "amqp://localhost:5672", "RabbitMQ URL")
	backendCmd.Flags().String("queue-name", "sensor-data", "RabbitMQ queue name for sensor readings")
	backendCmd.Flags().String("device-queue-name", "device-data", "RabbitMQ queue name for device creation messages")
	backendCmd.Flags().Int("grpc-port", 9090, "gRPC server port")

	// Bind flags to viper
	_ = viper.BindPFlag("backend.db.host", backendCmd.Flags().Lookup("db-host"))
	_ = viper.BindPFlag("backend.db.port", backendCmd.Flags().Lookup("db-port"))
	_ = viper.BindPFlag("backend.db.user", backendCmd.Flags().Lookup("db-user"))
	_ = viper.BindPFlag("backend.db.password", backendCmd.Flags().Lookup("db-password"))
	_ = viper.BindPFlag("backend.db.name", backendCmd.Flags().Lookup("db-name"))
	_ = viper.BindPFlag("backend.db.sslmode", backendCmd.Flags().Lookup("db-sslmode"))
	_ = viper.BindPFlag("backend.rabbitmq.url", backendCmd.Flags().Lookup("rabbitmq-url"))
	_ = viper.BindPFlag("backend.rabbitmq.queue_name", backendCmd.Flags().Lookup("queue-name"))
	_ = viper.BindPFlag("backend.rabbitmq.device_queue_name", backendCmd.Flags().Lookup("device-queue-name"))
	_ = viper.BindPFlag("backend.grpc.port", backendCmd.Flags().Lookup("grpc-port"))
}

func runBackend(_ *cobra.Command, _ []string) error {
	logger := GetLogger()
	logger.Info("starting backend service")

	// Create backend configuration from viper
	config := &backend.ServerConfig{
		Logger:          logger,
		DBHost:          viper.GetString("backend.db.host"),
		DBPort:          viper.GetInt("backend.db.port"),
		DBUser:          viper.GetString("backend.db.user"),
		DBPassword:      viper.GetString("backend.db.password"),
		DBName:          viper.GetString("backend.db.name"),
		DBSSLMode:       viper.GetString("backend.db.sslmode"),
		RabbitMQURL:     viper.GetString("backend.rabbitmq.url"),
		QueueName:       viper.GetString("backend.rabbitmq.queue_name"),
		DeviceQueueName: viper.GetString("backend.rabbitmq.device_queue_name"),
		GRPCPort:        viper.GetInt("backend.grpc.port"),
	}

	// Create and run server
	server, err := backend.NewServer(config)
	if err != nil {
		logger.Error("failed to create backend server", "error", err)
		return err
	}

	logger.Info("backend server configuration",
		"db_host", config.DBHost,
		"db_port", config.DBPort,
		"db_name", config.DBName,
		"rabbitmq_url", config.RabbitMQURL,
		"sensor_queue", config.QueueName,
		"device_queue", config.DeviceQueueName,
		"grpc_port", config.GRPCPort,
	)

	if err := server.Run(context.Background()); err != nil {
		logger.Error("backend server error", "error", err)
		return err
	}

	logger.Info("backend server stopped")
	return nil
}
