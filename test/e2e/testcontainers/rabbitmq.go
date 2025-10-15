// Package testcontainers provides helper functions for managing test containers across e2e tests.
package testcontainers

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RabbitMQConfig holds configuration for RabbitMQ test container.
type RabbitMQConfig struct {
	// User is the RabbitMQ username (default: guest)
	User string
	// Password is the RabbitMQ password (default: guest)
	Password string
	// ContainerName is the name of the container (optional)
	ContainerName string
}

// StartRabbitMQ starts a RabbitMQ container for testing and returns the container and connection URL.
func StartRabbitMQ(ctx context.Context, config *RabbitMQConfig) (testcontainers.Container, string, error) {
	// Set defaults
	if config == nil {
		config = &RabbitMQConfig{}
	}
	if config.User == "" {
		config.User = "guest"
	}
	if config.Password == "" {
		config.Password = "guest"
	}

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:3-management-alpine",
			ExposedPorts: []string{"5672/tcp", "15672/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("5672/tcp"),
				wait.ForLog("Server startup complete"),
			),
			Env: map[string]string{
				"RABBITMQ_DEFAULT_USER": config.User,
				"RABBITMQ_DEFAULT_PASS": config.Password,
			},
			Name: config.ContainerName,
		},
		Started: true,
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to start RabbitMQ container: %w", err)
	}

	// Get host and port
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5672")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", fmt.Errorf("failed to get container port: %w", err)
	}

	// Build connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.User, config.Password, host, port.Port())

	return container, url, nil
}
