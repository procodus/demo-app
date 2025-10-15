package mq

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"

	e2econtainers "procodus.dev/demo-app/test/e2e/testcontainers"
)

var (
	rabbitmqURL string
	testLogger  *slog.Logger
	mqContainer testcontainers.Container
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

	testLogger.Info("starting RabbitMQ container for E2E tests")

	// Start RabbitMQ container using helper
	var err error
	mqContainer, rabbitmqURL, err = e2econtainers.StartRabbitMQ(ctx, &e2econtainers.RabbitMQConfig{
		User:          "guest",
		Password:      "guest",
		ContainerName: "rabbitmq-e2e-test",
	})

	if err != nil {
		Fail(fmt.Sprintf("Failed to start RabbitMQ container: %v", err))
	}

	testLogger.Info("RabbitMQ container started",
		"container_id", mqContainer.GetContainerID(),
		"url", rabbitmqURL,
	)

	testLogger.Info("RabbitMQ is ready for testing")
})

var _ = AfterSuite(func() {
	if mqContainer != nil {
		ctx := context.Background()
		testLogger.Info("stopping RabbitMQ container", "container_id", mqContainer.GetContainerID())
		err := mqContainer.Terminate(ctx)
		if err != nil {
			testLogger.Error("failed to stop RabbitMQ container", "error", err)
		}
	}
})
