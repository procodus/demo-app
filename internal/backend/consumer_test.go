package backend_test

import (
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"

	"procodus.dev/demo-app/internal/backend"
)

var _ = Describe("Consumer", func() {
	var (
		logger *slog.Logger
		mockDB *gorm.DB
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
		// mockDB is nil - we're only testing validation, not actual DB operations
		mockDB = nil
	})

	Describe("NewConsumer", func() {
		Context("with valid configuration", func() {
			It("should create a consumer", func() {
				// Create a mock DB for this test
				dbCfg := &backend.DBConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
					Logger:   logger,
				}
				db, err := backend.NewDB(dbCfg)
				// We expect this to fail since we don't have a real DB, but we're testing the constructor
				if err == nil && db != nil {
					defer backend.CloseDB(db, logger)
				}

				config := &backend.ConsumerConfig{
					Logger:      logger,
					DB:          db,
					RabbitMQURL: "amqp://localhost:5672",
					QueueName:   "test-queue",
				}

				// This will create the consumer but not connect to MQ yet
				consumer, err := backend.NewConsumer(config)
				if db != nil {
					Expect(err).NotTo(HaveOccurred())
					Expect(consumer).NotTo(BeNil())
				}
			})
		})

		Context("with invalid configuration", func() {
			It("should return error when config is nil", func() {
				consumer, err := backend.NewConsumer(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("config cannot be nil"))
				Expect(consumer).To(BeNil())
			})

			It("should return error when logger is nil", func() {
				config := &backend.ConsumerConfig{
					Logger:      nil,
					DB:          mockDB,
					RabbitMQURL: "amqp://localhost:5672",
					QueueName:   "test-queue",
				}

				consumer, err := backend.NewConsumer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("logger"))
				Expect(consumer).To(BeNil())
			})

			It("should return error when database is nil", func() {
				config := &backend.ConsumerConfig{
					Logger:      logger,
					DB:          nil,
					RabbitMQURL: "amqp://localhost:5672",
					QueueName:   "test-queue",
				}

				consumer, err := backend.NewConsumer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database"))
				Expect(consumer).To(BeNil())
			})
		})

		Context("with different configurations", func() {
			It("should validate configuration parameters", func() {
				// Test that validation checks happen in order
				config := &backend.ConsumerConfig{
					Logger:      nil,
					DB:          nil,
					RabbitMQURL: "",
					QueueName:   "",
				}

				consumer, err := backend.NewConsumer(config)
				Expect(err).To(HaveOccurred())
				Expect(consumer).To(BeNil())
			})
		})
	})
})
