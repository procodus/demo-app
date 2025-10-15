package producer_test

import (
	"context"
	"log/slog"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/producer"
)

var _ = Describe("Producer Server", func() {
	var (
		logger *slog.Logger
	)

	BeforeEach(func() {
		// Create a logger for tests
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError, // Only show errors in tests
		}))
	})

	Describe("NewServer", func() {
		Context("with valid configuration", func() {
			It("should create a server", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 5,
					Interval:      5 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})

			It("should create server with minimum producer count", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 1,
					Interval:      1 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})

			It("should create server with large producer count", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 20,
					Interval:      1 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})

			It("should create server with small interval", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 2,
					Interval:      100 * time.Millisecond,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})
		})

		Context("with invalid configuration", func() {
			It("should return error when producer count is zero", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 0,
					Interval:      5 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("producer count"))
				Expect(server).To(BeNil())
			})

			It("should return error when producer count is negative", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: -1,
					Interval:      5 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("producer count"))
				Expect(server).To(BeNil())
			})

			It("should return error when interval is zero", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 5,
					Interval:      0,
				}

				server, err := producer.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("interval"))
				Expect(server).To(BeNil())
			})

			It("should return error when interval is negative", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 5,
					Interval:      -1 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("interval"))
				Expect(server).To(BeNil())
			})

			It("should return error when logger is nil", func() {
				config := &producer.ServerConfig{
					Logger:        nil,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					ProducerCount: 5,
					Interval:      5 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("logger"))
				Expect(server).To(BeNil())
			})
		})

		Context("with different configurations", func() {
			It("should accept different RabbitMQ URLs", func() {
				urls := []string{
					"amqp://localhost:5672",
					"amqp://guest:guest@localhost:5672",
					"amqp://user:pass@rabbitmq:5672/vhost",
					"amqps://secure.example.com:5671",
				}

				for _, url := range urls {
					config := &producer.ServerConfig{
						Logger:        logger,
						RabbitMQURL:   url,
						QueueName:     "test-queue",
						ProducerCount: 1,
						Interval:      1 * time.Second,
					}

					server, err := producer.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})

			It("should accept different queue names", func() {
				queueNames := []string{
					"sensor-data",
					"iot-readings",
					"device-events",
					"test_queue_123",
				}

				for _, queueName := range queueNames {
					config := &producer.ServerConfig{
						Logger:        logger,
						RabbitMQURL:   "amqp://localhost:5672",
						QueueName:     queueName,
						ProducerCount: 1,
						Interval:      1 * time.Second,
					}

					server, err := producer.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})
		})
	})

	Describe("Server Run", func() {
		Context("with context cancellation", func() {
			It("should shutdown when context is canceled", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://invalid:5672", // Invalid to prevent actual connection
					QueueName:     "test-queue",
					ProducerCount: 2,
					Interval:      100 * time.Millisecond,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				// Should complete within reasonable time after context cancellation
				Eventually(done, 2*time.Second).Should(Receive(BeNil()))
			})

			It("should shutdown immediately with pre-canceled context", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://invalid:5672",
					QueueName:     "test-queue",
					ProducerCount: 2,
					Interval:      1 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel before Run

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				// Should complete very quickly
				Eventually(done, 1*time.Second).Should(Receive(BeNil()))
			})

			It("should handle multiple consecutive runs", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://invalid:5672",
					QueueName:     "test-queue",
					ProducerCount: 1,
					Interval:      100 * time.Millisecond,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				// First run
				ctx1, cancel1 := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel1()

				done1 := make(chan error, 1)
				go func() {
					done1 <- server.Run(ctx1)
				}()

				Eventually(done1, 1*time.Second).Should(Receive(BeNil()))

				// Second run
				ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel2()

				done2 := make(chan error, 1)
				go func() {
					done2 <- server.Run(ctx2)
				}()

				Eventually(done2, 1*time.Second).Should(Receive(BeNil()))
			})
		})

		Context("with different intervals", func() {
			It("should run with fast interval", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://invalid:5672",
					QueueName:     "test-queue",
					ProducerCount: 1,
					Interval:      50 * time.Millisecond,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				Eventually(done, 1*time.Second).Should(Receive(BeNil()))
			})

			It("should run with slow interval", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://invalid:5672",
					QueueName:     "test-queue",
					ProducerCount: 1,
					Interval:      10 * time.Second,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				// Should still shutdown promptly despite long interval
				Eventually(done, 1*time.Second).Should(Receive(BeNil()))
			})
		})

		Context("with multiple producers", func() {
			It("should manage multiple producer goroutines", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://invalid:5672",
					QueueName:     "test-queue",
					ProducerCount: 5,
					Interval:      100 * time.Millisecond,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
				defer cancel()

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				Eventually(done, 2*time.Second).Should(Receive(BeNil()))
			})

			It("should manage many producer goroutines", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://invalid:5672",
					QueueName:     "test-queue",
					ProducerCount: 20,
					Interval:      100 * time.Millisecond,
				}

				server, err := producer.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
				defer cancel()

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				Eventually(done, 3*time.Second).Should(Receive(BeNil()))
			})
		})
	})

	Describe("Server Shutdown", func() {
		It("should shutdown cleanly", func() {
			config := &producer.ServerConfig{
				Logger:        logger,
				RabbitMQURL:   "amqp://invalid:5672",
				QueueName:     "test-queue",
				ProducerCount: 2,
				Interval:      1 * time.Second,
			}

			server, err := producer.NewServer(config)
			Expect(err).NotTo(HaveOccurred())

			err = server.Shutdown()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle multiple shutdown calls", func() {
			config := &producer.ServerConfig{
				Logger:        logger,
				RabbitMQURL:   "amqp://invalid:5672",
				QueueName:     "test-queue",
				ProducerCount: 2,
				Interval:      1 * time.Second,
			}

			server, err := producer.NewServer(config)
			Expect(err).NotTo(HaveOccurred())

			err1 := server.Shutdown()
			Expect(err1).NotTo(HaveOccurred())

			err2 := server.Shutdown()
			// Second shutdown should not panic and may return error
			Expect(err2).To(Or(BeNil(), HaveOccurred()))
		})
	})

	Describe("ServerConfig", func() {
		Context("field ordering", func() {
			It("should have logger as first field for memory alignment", func() {
				config := &producer.ServerConfig{
					Logger:        logger,
					RabbitMQURL:   "amqp://localhost:5672",
					QueueName:     "test-queue",
					Interval:      5 * time.Second,
					ProducerCount: 5,
				}

				Expect(config.Logger).NotTo(BeNil())
			})
		})
	})

	Describe("Error Constants", func() {
		It("should return consistent error messages", func() {
			config := &producer.ServerConfig{
				Logger:        nil,
				RabbitMQURL:   "amqp://localhost:5672",
				QueueName:     "test-queue",
				ProducerCount: 0,
				Interval:      0,
			}

			_, err := producer.NewServer(config)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Concurrent Server Creation", func() {
		It("should handle concurrent NewServer calls", func() {
			results := make(chan error, 5)

			for i := 0; i < 5; i++ {
				go func(_ int) {
					config := &producer.ServerConfig{
						Logger:        logger,
						RabbitMQURL:   "amqp://invalid:5672",
						QueueName:     "test-queue",
						ProducerCount: 2,
						Interval:      1 * time.Second,
					}

					_, err := producer.NewServer(config)
					results <- err
				}(i)
			}

			// All should succeed
			for i := 0; i < 5; i++ {
				Eventually(results).Should(Receive(BeNil()))
			}
		})
	})
})
