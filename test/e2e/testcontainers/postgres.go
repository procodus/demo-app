package testcontainers

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresConfig holds configuration for PostgreSQL test container.
type PostgresConfig struct {
	// User is the PostgreSQL username (default: postgres)
	User string
	// Password is the PostgreSQL password (default: postgres)
	Password string
	// Database is the database name (default: testdb)
	Database string
	// ContainerName is the name of the container (optional)
	ContainerName string
}

// StartPostgres starts a PostgreSQL container for testing and returns the container and DSN.
func StartPostgres(ctx context.Context, config *PostgresConfig) (testcontainers.Container, string, error) {
	// Set defaults
	if config == nil {
		config = &PostgresConfig{}
	}
	if config.User == "" {
		config.User = "postgres"
	}
	if config.Password == "" {
		config.Password = "postgres"
	}
	if config.Database == "" {
		config.Database = "testdb"
	}

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("5432/tcp"),
				wait.ForLog("database system is ready to accept connections"),
			),
			Env: map[string]string{
				"POSTGRES_USER":     config.User,
				"POSTGRES_PASSWORD": config.Password,
				"POSTGRES_DB":       config.Database,
			},
			Name: config.ContainerName,
		},
		Started: true,
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Get host and port
	host, err := container.Host(ctx)
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			return nil, "", fmt.Errorf("failed to get container host: %w (cleanup error: %w)", err, termErr)
		}
		return nil, "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			return nil, "", fmt.Errorf("failed to get container port: %w (cleanup error: %w)", err, termErr)
		}
		return nil, "", fmt.Errorf("failed to get container port: %w", err)
	}

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), config.User, config.Password, config.Database)

	return container, dsn, nil
}

// GetPostgresConnectionInfo returns connection information for the PostgreSQL container.
func GetPostgresConnectionInfo(ctx context.Context, container testcontainers.Container, config *PostgresConfig) (host string, port int, user, password, database string, err error) {
	if config == nil {
		config = &PostgresConfig{
			User:     "postgres",
			Password: "postgres",
			Database: "testdb",
		}
	}

	host, err = container.Host(ctx)
	if err != nil {
		return "", 0, "", "", "", fmt.Errorf("failed to get host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return "", 0, "", "", "", fmt.Errorf("failed to get port: %w", err)
	}

	return host, mappedPort.Int(), config.User, config.Password, config.Database, nil
}
