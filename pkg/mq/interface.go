package mq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ClientInterface defines the interface for message queue operations.
// This interface enables easier testing through mocking and dependency injection.
type ClientInterface interface {
	// Push will push data onto the queue, and wait for a confirmation.
	// This will block until the server sends a confirmation.
	// The context is used for cancellation and timeout.
	Push(ctx context.Context, data []byte) error

	// UnsafePush will push to the queue without checking for confirmation.
	// It returns an error if it fails to connect.
	// No guarantees are provided for whether the server will receive the message.
	// The context is used for cancellation and timeout.
	UnsafePush(ctx context.Context, data []byte) error

	// Consume will continuously put queue items on the channel.
	// It is required to call delivery.Ack when it has been successfully processed,
	// or delivery.Nack when it fails.
	Consume() (<-chan amqp.Delivery, error)

	// Close will cleanly shut down the channel and connection.
	Close() error
}

// Ensure Client implements ClientInterface.
var _ ClientInterface = (*Client)(nil)
