package mq

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	containerID  string
	rabbitmqURL  string
	testLogger   *slog.Logger
	dockerExists bool
)

func TestMQE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MQ E2E Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()

	// Create logger for tests
	testLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Check if Docker is available
	if !isDockerAvailable(ctx) {
		Skip("Docker is not available - skipping E2E tests")
		return
	}
	dockerExists = true

	// Clean up any orphaned test containers first
	testLogger.Info("cleaning up orphaned containers")
	_ = cleanupOrphanedContainers(ctx)

	testLogger.Info("starting RabbitMQ container for E2E tests")

	// Start RabbitMQ container
	var err error
	containerID, rabbitmqURL, err = startRabbitMQ(ctx)
	if err != nil {
		Fail(fmt.Sprintf("Failed to start RabbitMQ container: %v", err))
	}

	testLogger.Info("RabbitMQ container started",
		"container_id", containerID,
		"url", rabbitmqURL,
	)

	// Wait for RabbitMQ to be ready
	testLogger.Info("waiting for RabbitMQ to be ready")
	if err := waitForRabbitMQ(ctx, rabbitmqURL, 30*time.Second); err != nil {
		// Clean up on failure
		_ = stopRabbitMQ(ctx, containerID)
		Fail(fmt.Sprintf("RabbitMQ did not become ready: %v", err))
	}

	testLogger.Info("RabbitMQ is ready for testing")
})

var _ = AfterSuite(func() {
	if !dockerExists {
		return
	}

	if containerID != "" {
		ctx := context.Background()
		testLogger.Info("stopping RabbitMQ container", "container_id", containerID)

		if err := stopRabbitMQ(ctx, containerID); err != nil {
			testLogger.Error("failed to stop RabbitMQ container",
				"container_id", containerID,
				"error", err,
			)
		} else {
			testLogger.Info("RabbitMQ container stopped and removed")
		}
	}
})
