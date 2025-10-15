package backend_test

import (
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/backend"
)

var _ = Describe("Backend Server", func() {
	var (
		logger *slog.Logger
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	})

	Describe("NewServer", func() {
		Context("with valid configuration", func() {
			It("should create a server", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})

			It("should create server with SSL mode enabled", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "require",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})

			It("should create server with different database ports", func() {
				ports := []int{5432, 5433, 5434, 15432}

				for _, port := range ports {
					config := &backend.ServerConfig{
						Logger:          logger,
						DBHost:          "localhost",
						DBPort:          port,
						DBUser:          "test",
						DBPassword:      "password",
						DBName:          "testdb",
						DBSSLMode:       "disable",
						RabbitMQURL:     "amqp://localhost:5672",
						QueueName:       "test-queue",
						DeviceQueueName: "device-queue",
						GRPCPort:        9090,
					}

					server, err := backend.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})

			It("should create server with empty password", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})
		})

		Context("with invalid configuration", func() {
			It("should return error when config is nil", func() {
				server, err := backend.NewServer(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("config cannot be nil"))
				Expect(server).To(BeNil())
			})

			It("should return error when logger is nil", func() {
				config := &backend.ServerConfig{
					Logger:          nil,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("logger"))
				Expect(server).To(BeNil())
			})

			It("should return error when RabbitMQ URL is empty", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("rabbitmq URL"))
				Expect(server).To(BeNil())
			})

			It("should return error when queue name is empty", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "",
					DeviceQueueName: "device-queue",
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("queue name"))
				Expect(server).To(BeNil())
			})

			It("should return error when device queue name is empty", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("device queue name"))
				Expect(server).To(BeNil())
			})

			It("should return error when database host is empty", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database host"))
				Expect(server).To(BeNil())
			})

			It("should return error when database port is zero", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          0,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database port"))
				Expect(server).To(BeNil())
			})

			It("should return error when database port is negative", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          -1,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database port"))
				Expect(server).To(BeNil())
			})

			It("should return error when database user is empty", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database user"))
				Expect(server).To(BeNil())
			})

			It("should return error when database name is empty", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        9090,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database name"))
				Expect(server).To(BeNil())
			})

			It("should return error when gRPC port is zero", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        0,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("gRPC port"))
				Expect(server).To(BeNil())
			})

			It("should return error when gRPC port is negative", func() {
				config := &backend.ServerConfig{
					Logger:          logger,
					DBHost:          "localhost",
					DBPort:          5432,
					DBUser:          "test",
					DBPassword:      "password",
					DBName:          "testdb",
					DBSSLMode:       "disable",
					RabbitMQURL:     "amqp://localhost:5672",
					QueueName:       "test-queue",
					DeviceQueueName: "device-queue",
					GRPCPort:        -1,
				}

				server, err := backend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("gRPC port"))
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
					config := &backend.ServerConfig{
						Logger:          logger,
						DBHost:          "localhost",
						DBPort:          5432,
						DBUser:          "test",
						DBPassword:      "password",
						DBName:          "testdb",
						DBSSLMode:       "disable",
						RabbitMQURL:     url,
						QueueName:       "test-queue",
						DeviceQueueName: "device-queue",
						GRPCPort:        9090,
					}

					server, err := backend.NewServer(config)
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
					config := &backend.ServerConfig{
						Logger:          logger,
						DBHost:          "localhost",
						DBPort:          5432,
						DBUser:          "test",
						DBPassword:      "password",
						DBName:          "testdb",
						DBSSLMode:       "disable",
						RabbitMQURL:     "amqp://localhost:5672",
						QueueName:       queueName,
						DeviceQueueName: "device-queue",
						GRPCPort:        9090,
					}

					server, err := backend.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})

			It("should accept different database hosts", func() {
				hosts := []string{
					"localhost",
					"127.0.0.1",
					"postgres.example.com",
					"10.0.0.1",
				}

				for _, host := range hosts {
					config := &backend.ServerConfig{
						Logger:          logger,
						DBHost:          host,
						DBPort:          5432,
						DBUser:          "test",
						DBPassword:      "password",
						DBName:          "testdb",
						DBSSLMode:       "disable",
						RabbitMQURL:     "amqp://localhost:5672",
						QueueName:       "test-queue",
						DeviceQueueName: "device-queue",
						GRPCPort:        9090,
					}

					server, err := backend.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})

			It("should accept different SSL modes", func() {
				sslModes := []string{
					"disable",
					"require",
					"verify-ca",
					"verify-full",
				}

				for _, sslMode := range sslModes {
					config := &backend.ServerConfig{
						Logger:          logger,
						DBHost:          "localhost",
						DBPort:          5432,
						DBUser:          "test",
						DBPassword:      "password",
						DBName:          "testdb",
						DBSSLMode:       sslMode,
						RabbitMQURL:     "amqp://localhost:5672",
						QueueName:       "test-queue",
						DeviceQueueName: "device-queue",
						GRPCPort:        9090,
					}

					server, err := backend.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})
		})
	})

	Describe("Server Shutdown", func() {
		It("should shutdown cleanly with no initialized components", func() {
			config := &backend.ServerConfig{
				Logger:          logger,
				DBHost:          "localhost",
				DBPort:          5432,
				DBUser:          "test",
				DBPassword:      "password",
				DBName:          "testdb",
				DBSSLMode:       "disable",
				RabbitMQURL:     "amqp://localhost:5672",
				QueueName:       "test-queue",
				DeviceQueueName: "device-queue",
				GRPCPort:        9090,
			}

			server, err := backend.NewServer(config)
			Expect(err).NotTo(HaveOccurred())

			err = server.Shutdown()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle multiple shutdown calls", func() {
			config := &backend.ServerConfig{
				Logger:          logger,
				DBHost:          "localhost",
				DBPort:          5432,
				DBUser:          "test",
				DBPassword:      "password",
				DBName:          "testdb",
				DBSSLMode:       "disable",
				RabbitMQURL:     "amqp://localhost:5672",
				QueueName:       "test-queue",
				DeviceQueueName: "device-queue",
				GRPCPort:        9090,
			}

			server, err := backend.NewServer(config)
			Expect(err).NotTo(HaveOccurred())

			err1 := server.Shutdown()
			Expect(err1).NotTo(HaveOccurred())

			err2 := server.Shutdown()
			Expect(err2).NotTo(HaveOccurred())
		})
	})

	Describe("Concurrent Server Creation", func() {
		It("should handle concurrent NewServer calls", func() {
			results := make(chan error, 5)

			for i := 0; i < 5; i++ {
				go func(_ int) {
					config := &backend.ServerConfig{
						Logger:          logger,
						DBHost:          "localhost",
						DBPort:          5432,
						DBUser:          "test",
						DBPassword:      "password",
						DBName:          "testdb",
						DBSSLMode:       "disable",
						RabbitMQURL:     "amqp://localhost:5672",
						QueueName:       "test-queue",
						DeviceQueueName: "device-queue",
						GRPCPort:        9090,
					}

					_, err := backend.NewServer(config)
					results <- err
				}(i)
			}

			for i := 0; i < 5; i++ {
				Eventually(results).Should(Receive(BeNil()))
			}
		})
	})
})
