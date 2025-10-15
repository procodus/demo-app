// Package mq provides end-to-end tests for the RabbitMQ client with Docker container management.
package mq

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitmqImage      = "rabbitmq:3-management-alpine"
	rabbitmqPort       = "5672"
	rabbitmqMgmtPort   = "15672"
	containerNameBase  = "rabbitmq-e2e-test"
	defaultRabbitMQURL = "amqp://guest:guest@localhost:5672/"
)

// isDockerAvailable checks if Docker is installed and running.
func isDockerAvailable(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "info")
	err := cmd.Run()
	return err == nil
}

// startRabbitMQ starts a RabbitMQ container and returns the container ID and connection URL.
func startRabbitMQ(ctx context.Context) (string, string, error) {
	// Generate unique container name
	containerName := containerNameBase + "-" + strconv.FormatInt(time.Now().Unix(), 10)

	// Pull the RabbitMQ image
	pullCmd := exec.CommandContext(ctx, "docker", "pull", rabbitmqImage)
	if err := pullCmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to pull RabbitMQ image: %w", err)
	}

	// Start RabbitMQ container
	// #nosec G204 - docker command with controlled arguments for testing
	args := []string{
		"run",
		"-d",
		"--rm",
		"--name", containerName,
		"-p", rabbitmqPort + ":" + rabbitmqPort,
		"-p", rabbitmqMgmtPort + ":" + rabbitmqMgmtPort,
		"-e", "RABBITMQ_DEFAULT_USER=guest",
		"-e", "RABBITMQ_DEFAULT_PASS=guest",
		rabbitmqImage,
	}

	cmd := exec.CommandContext(ctx, "docker", args...) // #nosec G204 - controlled docker args
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to start RabbitMQ container: %w (output: %s)", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	return containerID, defaultRabbitMQURL, nil
}

// stopRabbitMQ stops and removes the RabbitMQ container.
func stopRabbitMQ(ctx context.Context, containerID string) error {
	if containerID == "" {
		return nil
	}

	// Stop container (--rm flag will auto-remove it)
	cmd := exec.CommandContext(ctx, "docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop RabbitMQ container: %w (output: %s)", err, string(output))
	}

	// Wait a moment for removal
	time.Sleep(1 * time.Second)

	return nil
}

var errRabbitMQNotReady = errors.New("timeout waiting for RabbitMQ to be ready")

// waitForRabbitMQ waits for RabbitMQ to be ready by attempting to connect.
func waitForRabbitMQ(ctx context.Context, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errRabbitMQNotReady

		case <-ticker.C:
			// Try to connect
			conn, err := amqp.Dial(url)
			if err == nil {
				if closeErr := conn.Close(); closeErr != nil {
					// Connection worked, ignore close error
					return nil
				}
				return nil
			}
			// Continue waiting if connection failed
		}
	}
}

// cleanupOrphanedContainers removes any orphaned test containers.
func cleanupOrphanedContainers(ctx context.Context) error {
	// List containers with our naming pattern
	// #nosec G204 - docker command with controlled filter argument
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "-q",
		"--filter", "name="+containerNameBase)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	containerIDs := strings.Fields(string(output))
	for _, id := range containerIDs {
		// #nosec G204 - docker rm with container ID from docker ps output
		stopCmd := exec.CommandContext(ctx, "docker", "rm", "-f", id)
		// Best effort cleanup - ignore errors
		if err := stopCmd.Run(); err != nil {
			// Continue cleaning up other containers
			continue
		}
	}

	return nil
}
