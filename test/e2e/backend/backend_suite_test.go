package backend

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"procodus.dev/demo-app/internal/backend"
	"procodus.dev/demo-app/pkg/iot"
	e2econtainers "procodus.dev/demo-app/test/e2e/testcontainers"
)

var (
	testLogger *slog.Logger

	// Containers.
	postgresContainer testcontainers.Container
	rabbitMQContainer testcontainers.Container

	// Connection info.
	postgresDSN string
	rabbitmqURL string

	// Backend server.
	backendServer *backend.Server
	serverCtx     context.Context
	serverCancel  context.CancelFunc

	// gRPC client.
	grpcConn   *grpc.ClientConn
	grpcClient iot.IoTServiceClient

	// RabbitMQ client for publishing test messages.
	mqConn    *amqp.Connection
	mqChannel *amqp.Channel

	// Queue names.
	sensorQueueName = "sensor-data-e2e-test"
	deviceQueueName = "device-data-e2e-test"

	// gRPC port.
	grpcPort = 19090
)

func TestBackendE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backend E2E Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()

	// Create logger for tests
	testLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	testLogger.Info("starting PostgreSQL container for E2E tests")

	// Start PostgreSQL container
	var err error
	postgresContainer, postgresDSN, err = e2econtainers.StartPostgres(ctx, &e2econtainers.PostgresConfig{
		User:          "testuser",
		Password:      "testpass",
		Database:      "testdb",
		ContainerName: "postgres-backend-e2e-test",
	})
	if err != nil {
		Fail(fmt.Sprintf("Failed to start PostgreSQL container: %v", err))
	}

	testLogger.Info("PostgreSQL container started",
		"container_id", postgresContainer.GetContainerID(),
		"dsn", postgresDSN,
	)

	testLogger.Info("starting RabbitMQ container for E2E tests")

	// Start RabbitMQ container
	rabbitMQContainer, rabbitmqURL, err = e2econtainers.StartRabbitMQ(ctx, &e2econtainers.RabbitMQConfig{
		User:          "guest",
		Password:      "guest",
		ContainerName: "rabbitmq-backend-e2e-test",
	})
	if err != nil {
		Fail(fmt.Sprintf("Failed to start RabbitMQ container: %v", err))
	}

	testLogger.Info("RabbitMQ container started",
		"container_id", rabbitMQContainer.GetContainerID(),
		"url", rabbitmqURL,
	)

	// Extract PostgreSQL connection parameters
	host, port, user, password, dbname, err := e2econtainers.GetPostgresConnectionInfo(
		ctx,
		postgresContainer,
		&e2econtainers.PostgresConfig{
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
		},
	)
	if err != nil {
		Fail(fmt.Sprintf("Failed to get PostgreSQL connection info: %v", err))
	}

	// Create backend server configuration
	serverConfig := &backend.ServerConfig{
		Logger:          testLogger,
		DBHost:          host,
		DBPort:          port,
		DBUser:          user,
		DBPassword:      password,
		DBName:          dbname,
		DBSSLMode:       "disable",
		RabbitMQURL:     rabbitmqURL,
		QueueName:       sensorQueueName,
		DeviceQueueName: deviceQueueName,
		GRPCPort:        grpcPort,
	}

	// Create backend server
	backendServer, err = backend.NewServer(serverConfig)
	if err != nil {
		Fail(fmt.Sprintf("Failed to create backend server: %v", err))
	}

	testLogger.Info("starting backend server")

	// Start backend server in background
	serverCtx, serverCancel = context.WithCancel(context.Background())
	serverErr := make(chan error, 1)
	go func() {
		if err := backendServer.Run(serverCtx); err != nil {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Wait for server to start (give it time to initialize both consumers)
	time.Sleep(5 * time.Second)

	// Check if server started successfully
	select {
	case err := <-serverErr:
		if err != nil {
			Fail(fmt.Sprintf("Backend server failed to start: %v", err))
		}
	default:
		// Server is running
	}

	testLogger.Info("backend server started successfully")

	// Create gRPC client
	grpcAddr := fmt.Sprintf("localhost:%d", grpcPort)
	grpcConn, err = grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Fail(fmt.Sprintf("Failed to create gRPC client: %v", err))
	}

	grpcClient = iot.NewIoTServiceClient(grpcConn)

	testLogger.Info("gRPC client connected", "address", grpcAddr)

	// Create RabbitMQ connection for publishing test messages
	mqConn, err = amqp.Dial(rabbitmqURL)
	if err != nil {
		Fail(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
	}

	mqChannel, err = mqConn.Channel()
	if err != nil {
		Fail(fmt.Sprintf("Failed to create RabbitMQ channel: %v", err))
	}

	// Note: Queues are automatically declared by the backend consumers
	// No need to declare them here as it would conflict with consumer declarations

	testLogger.Info("RabbitMQ client ready")
	testLogger.Info("backend E2E test environment ready")
})

var _ = AfterSuite(func() {
	testLogger.Info("cleaning up backend E2E test environment")

	// Close RabbitMQ channel and connection
	if mqChannel != nil {
		_ = mqChannel.Close()
	}
	if mqConn != nil {
		_ = mqConn.Close()
	}

	// Close gRPC client
	if grpcConn != nil {
		_ = grpcConn.Close()
	}

	// Stop backend server
	if serverCancel != nil {
		testLogger.Info("stopping backend server")
		serverCancel()
		time.Sleep(1 * time.Second) // Give server time to shut down
	}

	// Stop containers
	ctx := context.Background()

	if rabbitMQContainer != nil {
		testLogger.Info("stopping RabbitMQ container", "container_id", rabbitMQContainer.GetContainerID())
		err := rabbitMQContainer.Terminate(ctx)
		if err != nil {
			testLogger.Error("failed to stop RabbitMQ container", "error", err)
		}
	}

	if postgresContainer != nil {
		testLogger.Info("stopping PostgreSQL container", "container_id", postgresContainer.GetContainerID())
		err := postgresContainer.Terminate(ctx)
		if err != nil {
			testLogger.Error("failed to stop PostgreSQL container", "error", err)
		}
	}

	testLogger.Info("backend E2E test environment cleaned up")
})
