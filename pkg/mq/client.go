// Package mq provides a RabbitMQ client with automatic reconnection and error handling.
package mq

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	amqp "github.com/rabbitmq/amqp091-go"

	"procodus.dev/demo-app/pkg/metrics"
)

// Client is a RabbitMQ client that handles connection management,
// automatic reconnection, and provides methods for publishing and consuming messages.
type Client struct {
	m               *sync.Mutex
	infolog         *slog.Logger
	errlog          *slog.Logger
	connection      *amqp.Connection
	channel         *amqp.Channel
	done            chan bool
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	queueName       string
	isReady         bool
	metrics         *metrics.MQMetrics // Optional metrics
}

const (
	// When reconnecting to the server after connection failure.
	reconnectDelay = 5 * time.Second

	// When setting up the channel after a channel exception.
	reInitDelay = 2 * time.Second

	// Initial backoff delay for Push retries.
	initialBackoff = 100 * time.Millisecond

	// Maximum backoff delay for Push retries.
	maxBackoff = 10 * time.Second

	// Backoff multiplier for exponential backoff.
	backoffMultiplier = 2

	// Maximum number of retry attempts before giving up.
	maxRetryAttempts = 5
)

var (
	errNotConnected       = errors.New("not connected to a server")
	errAlreadyClosed      = errors.New("already closed: not connected to the server")
	errShutdown           = errors.New("client is shutting down")
	errMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")
)

// New creates a new consumer state instance, and automatically
// attempts to connect to the server.
func New(queueName, addr string, l *slog.Logger) *Client {
	client := Client{
		m:         &sync.Mutex{},
		infolog:   l,
		errlog:    l,
		queueName: queueName,
		done:      make(chan bool),
	}
	go client.handleReconnect(addr)
	return &client
}

// SetMetrics sets the metrics collector for this client.
// This should be called before the client starts processing messages.
func (client *Client) SetMetrics(m *metrics.MQMetrics) {
	client.metrics = m
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (client *Client) handleReconnect(addr string) {
	for {
		client.m.Lock()
		client.isReady = false
		client.m.Unlock()

		client.infolog.Info("attempting to connect")

		// Track reconnection attempt
		if client.metrics != nil {
			client.metrics.ReconnectAttempts.Inc()
		}

		conn, err := client.connect(addr)
		if err != nil {
			client.errlog.Error("failed to connect. Retrying...", "error", err)

			select {
			case <-client.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := client.handleReInit(conn); done {
			break
		}
	}
}

// connect will create a new AMQP connection.
func (client *Client) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		// Update connection status metric
		if client.metrics != nil {
			client.metrics.ConnectionStatus.Set(0)
		}
		return nil, err
	}

	client.changeConnection(conn)
	client.infolog.Info("connected")

	// Update connection status metric
	if client.metrics != nil {
		client.metrics.ConnectionStatus.Set(1)
	}

	return conn, nil
}

// handleReInit will wait for a channel error
// and then continuously attempt to re-initialize both channels.
func (client *Client) handleReInit(conn *amqp.Connection) bool {
	for {
		client.m.Lock()
		client.isReady = false
		client.m.Unlock()

		err := client.init(conn)
		if err != nil {
			client.errlog.Error("failed to initialize channel, retrying...", "error", err)

			select {
			case <-client.done:
				return true
			case <-client.notifyConnClose:
				client.infolog.Info("connection closed, reconnecting...")
				return false
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-client.done:
			return true
		case <-client.notifyConnClose:
			client.infolog.Info("connection closed, reconnecting...")
			return false
		case <-client.notifyChanClose:
			client.infolog.Info("channel closed, re-running init...")
		}
	}
}

// init will initialize channel & declare queue.
func (client *Client) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(
		client.queueName,
		false, // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		return err
	}

	client.changeChannel(ch)
	client.m.Lock()
	client.isReady = true
	client.m.Unlock()
	client.infolog.Info("client init done")

	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (client *Client) changeConnection(connection *amqp.Connection) {
	client.connection = connection
	client.notifyConnClose = make(chan *amqp.Error, 1)
	client.connection.NotifyClose(client.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (client *Client) changeChannel(channel *amqp.Channel) {
	client.channel = channel
	client.notifyChanClose = make(chan *amqp.Error, 1)
	client.notifyConfirm = make(chan amqp.Confirmation, 1)
	client.channel.NotifyClose(client.notifyChanClose)
	client.channel.NotifyPublish(client.notifyConfirm)
}

// Push will push data onto the queue, and wait for a confirmation.
// This will block until the server sends a confirmation. Errors are
// only returned if the push action itself fails, see UnsafePush.
// The context is used for cancellation and timeout.
// Uses exponential backoff retry when the client is not connected,
// allowing time for automatic reconnection to succeed.
// After maxRetryAttempts (5) failed attempts, returns a fatal error.
func (client *Client) Push(ctx context.Context, data []byte) error {
	// Track duration
	var timer *prometheus.Timer
	if client.metrics != nil {
		timer = prometheus.NewTimer(client.metrics.PushDuration.WithLabelValues(client.queueName))
		defer timer.ObserveDuration()
	}

	backoff := initialBackoff
	retryCount := 0

	for {
		// Check if max retries exceeded
		if retryCount >= maxRetryAttempts {
			client.errlog.Error("maximum retry attempts exceeded",
				"retry_count", retryCount,
				"max_attempts", maxRetryAttempts)

			// Track failure
			if client.metrics != nil {
				client.metrics.PushFailures.WithLabelValues(client.queueName, "max_retries_exceeded").Inc()
			}

			return errMaxRetriesExceeded
		}

		// Check if connected
		client.m.Lock()
		isReady := client.isReady
		client.m.Unlock()

		if !isReady {
			// Not connected - use exponential backoff to wait for reconnection
			client.infolog.Info("not connected, waiting for reconnection",
				"backoff", backoff,
				"retry_count", retryCount)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-client.done:
				return errShutdown
			case <-time.After(backoff):
				// Increase backoff exponentially
				backoff *= backoffMultiplier
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				retryCount++
				continue
			}
		}

		// Attempt to push
		err := client.UnsafePush(ctx, data)
		if err != nil {
			client.errlog.Error("push failed, retrying with backoff",
				"error", err,
				"backoff", backoff,
				"retry_count", retryCount)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-client.done:
				return errShutdown
			case <-time.After(backoff):
				// Increase backoff exponentially
				backoff *= backoffMultiplier
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				retryCount++
				continue
			}
		}

		// Wait for confirmation
		select {
		case <-ctx.Done():
			// Track failure
			if client.metrics != nil {
				client.metrics.PushFailures.WithLabelValues(client.queueName, "context_canceled").Inc()
			}
			return ctx.Err()
		case confirm := <-client.notifyConfirm:
			if confirm.Ack {
				// Track success
				if client.metrics != nil {
					client.metrics.MessagesPushed.WithLabelValues(client.queueName).Inc()
				}

				if retryCount > 0 {
					client.infolog.Info("push confirmed after retries",
						"delivery_tag", confirm.DeliveryTag,
						"retry_count", retryCount)
				} else {
					client.infolog.Info("push confirmed", "delivery_tag", confirm.DeliveryTag)
				}
				return nil
			}
			// Negative acknowledgment - retry with backoff
			client.errlog.Warn("push not acknowledged, retrying",
				"delivery_tag", confirm.DeliveryTag,
				"backoff", backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-client.done:
				return errShutdown
			case <-time.After(backoff):
				// Increase backoff exponentially
				backoff *= backoffMultiplier
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				retryCount++
				continue
			}
		}
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message. The context is used for cancellation and timeout.
func (client *Client) UnsafePush(ctx context.Context, data []byte) error {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return errNotConnected
	}
	client.m.Unlock()

	return client.channel.PublishWithContext(
		ctx,
		"",               // Exchange
		client.queueName, // Routing key
		false,            // Mandatory
		false,            // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}

// Consume will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (client *Client) Consume() (<-chan amqp.Delivery, error) {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return nil, errNotConnected
	}
	client.m.Unlock()

	if err := client.channel.Qos(
		1,     // prefetchCount
		0,     // prefetchSize
		false, // global
	); err != nil {
		return nil, err
	}

	return client.channel.Consume(
		client.queueName,
		"",    // Consumer
		false, // Auto-Ack
		false, // Exclusive
		false, // No-local
		false, // No-Wait
		nil,   // Args
	)
}

// Close will cleanly shut down the channel and connection.
func (client *Client) Close() error {
	client.m.Lock()
	// we read and write isReady in two locations, so we grab the lock and hold onto
	// it until we are finished
	defer client.m.Unlock()

	if !client.isReady {
		return errAlreadyClosed
	}
	close(client.done)
	err := client.channel.Close()
	if err != nil {
		return err
	}
	err = client.connection.Close()
	if err != nil {
		return err
	}

	client.isReady = false

	// Update connection status metric
	if client.metrics != nil {
		client.metrics.ConnectionStatus.Set(0)
	}

	return nil
}
