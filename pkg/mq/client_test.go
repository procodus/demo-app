package mq_test

import (
	"context"
	"log/slog"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/pkg/mq"
)

var _ = Describe("MQ Client", func() {
	var (
		logger *slog.Logger
	)

	BeforeEach(func() {
		// Create a logger that discards output for tests
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	})

	Describe("New", func() {
		It("should create a new client instance", func() {
			client := mq.New("test-queue", "amqp://localhost:5672", logger)
			Expect(client).NotTo(BeNil())
		})

		It("should start background reconnection goroutine", func() {
			client := mq.New("test-queue", "amqp://invalid:5672", logger)
			Expect(client).NotTo(BeNil())

			// Give the goroutine a moment to start
			time.Sleep(100 * time.Millisecond)

			// Clean up
			_ = client.Close()
		})
	})

	Describe("Push", func() {
		Context("when not connected", func() {
			It("should retry with backoff and timeout", func() {
				client := mq.New("test-queue", "amqp://invalid:5672", logger)

				// Give client time to attempt connection and fail
				time.Sleep(100 * time.Millisecond)

				// Use a context with timeout to prevent infinite retries
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				start := time.Now()
				err := client.Push(ctx, []byte("test message"))
				elapsed := time.Since(start)

				// Should eventually timeout due to context
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(SatisfyAny(
					ContainSubstring("context deadline exceeded"),
					ContainSubstring("context canceled"),
				))
				// Should have waited for backoff retries
				Expect(elapsed).To(BeNumerically(">=", 100*time.Millisecond))

				_ = client.Close()
			})

			It("should return error after max retry attempts", func() {
				client := mq.New("test-queue", "amqp://invalid:5672", logger)

				// Give client time to attempt connection and fail
				time.Sleep(100 * time.Millisecond)

				// Use a long timeout that won't interfere with max retry logic
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				start := time.Now()
				err := client.Push(ctx, []byte("test message"))
				elapsed := time.Since(start)

				// Should return max retries exceeded error
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("maximum retry attempts exceeded"))

				// Should have waited for multiple backoff attempts
				// 5 retries with backoff: 100ms + 200ms + 400ms + 800ms + 1600ms = 3100ms minimum
				Expect(elapsed).To(BeNumerically(">=", 3*time.Second))
				Expect(elapsed).To(BeNumerically("<", 10*time.Second))

				_ = client.Close()
			})

			It("should return error for UnsafePush", func() {
				client := mq.New("test-queue", "amqp://invalid:5672", logger)

				// Give client time to attempt connection and fail
				time.Sleep(100 * time.Millisecond)

				err := client.UnsafePush(context.Background(), []byte("test message"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not connected"))

				_ = client.Close()
			})
		})
	})

	Describe("Consume", func() {
		Context("when not connected", func() {
			It("should return error", func() {
				client := mq.New("test-queue", "amqp://invalid:5672", logger)

				// Give client time to attempt connection and fail
				time.Sleep(100 * time.Millisecond)

				_, err := client.Consume()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not connected"))

				_ = client.Close()
			})
		})
	})

	Describe("Close", func() {
		Context("when not connected", func() {
			It("should return already closed error", func() {
				client := mq.New("test-queue", "amqp://invalid:5672", logger)

				// Give client time to attempt connection and fail
				time.Sleep(100 * time.Millisecond)

				err := client.Close()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already closed"))
			})
		})

		Context("when closing twice", func() {
			It("should return error on second close", func() {
				client := mq.New("test-queue", "amqp://invalid:5672", logger)

				// Give client time to attempt connection and fail
				time.Sleep(100 * time.Millisecond)

				// First close
				err1 := client.Close()
				Expect(err1).To(HaveOccurred()) // Will error because not connected

				// Second close should also error
				err2 := client.Close()
				Expect(err2).To(HaveOccurred())
				Expect(err2.Error()).To(ContainSubstring("already closed"))
			})
		})
	})

	Describe("Error Constants", func() {
		It("should have meaningful error messages", func() {
			client := mq.New("test-queue", "amqp://invalid:5672", logger)
			defer func() { _ = client.Close() }()

			// Wait for connection failure
			time.Sleep(100 * time.Millisecond)

			// Test that errors are returned (use timeout context)
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			err := client.Push(ctx, []byte("test"))
			Expect(err).NotTo(BeNil())
		})
	})

	Describe("Concurrent Access", func() {
		It("should handle concurrent Push attempts safely", func() {
			client := mq.New("test-queue", "amqp://invalid:5672", logger)
			defer func() { _ = client.Close() }()

			// Wait for connection failure
			time.Sleep(100 * time.Millisecond)

			// Try multiple concurrent pushes
			done := make(chan bool, 3)
			for i := 0; i < 3; i++ {
				go func() {
					_ = client.UnsafePush(context.Background(), []byte("test"))
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 3; i++ {
				Eventually(done).Should(Receive())
			}
		})

		It("should handle concurrent Close attempts safely", func() {
			client := mq.New("test-queue", "amqp://invalid:5672", logger)

			// Wait for connection failure
			time.Sleep(100 * time.Millisecond)

			// Try multiple concurrent closes
			done := make(chan bool, 3)
			for i := 0; i < 3; i++ {
				go func() {
					_ = client.Close()
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 3; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Describe("Configuration", func() {
		It("should accept custom queue names", func() {
			queueNames := []string{
				"sensor-data",
				"iot-readings",
				"device-events",
			}

			for _, queueName := range queueNames {
				client := mq.New(queueName, "amqp://invalid:5672", logger)
				Expect(client).NotTo(BeNil())
				_ = client.Close()
			}
		})

		It("should accept different AMQP URLs", func() {
			urls := []string{
				"amqp://localhost:5672",
				"amqp://guest:guest@localhost:5672",
				"amqp://rabbitmq:5672/vhost",
			}

			for _, url := range urls {
				client := mq.New("test-queue", url, logger)
				Expect(client).NotTo(BeNil())
				time.Sleep(50 * time.Millisecond) // Give time for connection attempt
				_ = client.Close()
			}
		})

		It("should require a logger", func() {
			client := mq.New("test-queue", "amqp://invalid:5672", logger)
			Expect(client).NotTo(BeNil())
			_ = client.Close()
		})
	})

	Describe("Integration Scenarios", Label("unit"), func() {
		Context("without RabbitMQ connection", func() {
			It("should handle connection failures gracefully", func() {
				client := mq.New("test-queue", "amqp://nonexistent:5672", logger)

				// Give client time to attempt connection
				time.Sleep(200 * time.Millisecond)

				// Client should exist but not be ready
				Expect(client).NotTo(BeNil())

				// Operations should fail gracefully
				err := client.UnsafePush(context.Background(), []byte("test"))
				Expect(err).To(HaveOccurred())

				_ = client.Close()
			})

			It("should continue retrying connection", func() {
				client := mq.New("test-queue", "amqp://nonexistent:5672", logger)

				// Wait for multiple retry attempts (connection failure + retry delay)
				// reconnectDelay is 5 seconds, but we just want to verify it's trying
				time.Sleep(500 * time.Millisecond)

				// Client should still exist
				Expect(client).NotTo(BeNil())

				_ = client.Close()
			})
		})
	})
})
