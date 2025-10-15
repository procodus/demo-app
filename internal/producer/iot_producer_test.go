package producer_test

import (
	"context"
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/producer"
	"procodus.dev/demo-app/pkg/mq"
)

var _ = Describe("IoT Producer", func() {
	var (
		logger         *slog.Logger
		mqClient       *mq.Client
		deviceMQClient *mq.Client
	)

	BeforeEach(func() {
		// Create a logger that discards output for tests
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	})

	Describe("NewProducer", func() {
		BeforeEach(func() {
			// Create MQ clients (won't actually connect in unit tests)
			mqClient = mq.New("test-queue", "amqp://invalid:5672", logger)
			deviceMQClient = mq.New("device-queue", "amqp://invalid:5672", logger)
		})

		AfterEach(func() {
			_ = mqClient.Close()
			_ = deviceMQClient.Close()
		})

		It("should create a producer with a valid MQ client", func() {
			prod := producer.NewProducer(mqClient, deviceMQClient)
			Expect(prod).NotTo(BeNil())
		})

		It("should create a producer with IoT devices", func() {
			prod := producer.NewProducer(mqClient, deviceMQClient)
			Expect(prod.IoTDevices).NotTo(BeEmpty())
			Expect(len(prod.IoTDevices)).To(BeNumerically(">=", 1))
			Expect(len(prod.IoTDevices)).To(BeNumerically("<=", 5))
		})

		It("should create a producer with the provided MQ client", func() {
			prod := producer.NewProducer(mqClient, deviceMQClient)
			Expect(prod.MQClient).To(Equal(mqClient))
		})

		It("should create different device sets on multiple calls", func() {
			prod1 := producer.NewProducer(mqClient, deviceMQClient)
			prod2 := producer.NewProducer(mqClient, deviceMQClient)

			// At least one device should be different (highly likely with UUIDs)
			allSame := true
			if len(prod1.IoTDevices) != len(prod2.IoTDevices) {
				allSame = false
			} else {
				for i := range prod1.IoTDevices {
					if prod1.IoTDevices[i].DeviceID != prod2.IoTDevices[i].DeviceID {
						allSame = false
						break
					}
				}
			}
			Expect(allSame).To(BeFalse())
		})
	})

	Describe("RandomDataPoint", func() {
		var prod *producer.Producer

		BeforeEach(func() {
			mqClient = mq.New("test-queue", "amqp://invalid:5672", logger)
			deviceMQClient = mq.New("device-queue", "amqp://invalid:5672", logger)
			prod = producer.NewProducer(mqClient, deviceMQClient)
		})

		AfterEach(func() {
			_ = mqClient.Close()
			_ = deviceMQClient.Close()
		})

		Context("with disconnected MQ client", func() {
			It("should return an error", func() {
				ctx := context.Background()
				err := prod.RandomDataPoint(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with context", func() {
			It("should accept a context parameter", func() {
				ctx := context.Background()
				err := prod.RandomDataPoint(ctx)
				// Will error because not connected, but that's ok
				Expect(err).NotTo(BeNil())
			})

			It("should accept a canceled context", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				err := prod.RandomDataPoint(ctx)
				// Will error because not connected
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Producer Integration", func() {
		It("should have valid device data structure", func() {
			mqClient := mq.New("test-queue", "amqp://invalid:5672", logger)
			deviceMQClient := mq.New("device-queue", "amqp://invalid:5672", logger)
			defer func() {
				_ = mqClient.Close()
				_ = deviceMQClient.Close()
			}()

			prod := producer.NewProducer(mqClient, deviceMQClient)

			// Verify device structure
			for _, device := range prod.IoTDevices {
				Expect(device.DeviceID).NotTo(BeEmpty())
				Expect(device.Location).NotTo(BeEmpty())
				Expect(device.MacAddress).NotTo(BeEmpty())
				Expect(device.IPAddress).NotTo(BeEmpty())
				Expect(device.Firmware).NotTo(BeEmpty())
			}
		})

		It("should maintain consistent device list", func() {
			mqClient := mq.New("test-queue", "amqp://invalid:5672", logger)
			deviceMQClient := mq.New("device-queue", "amqp://invalid:5672", logger)
			defer func() {
				_ = mqClient.Close()
				_ = deviceMQClient.Close()
			}()

			prod := producer.NewProducer(mqClient, deviceMQClient)
			initialCount := len(prod.IoTDevices)

			// Call RandomDataPoint multiple times
			ctx := context.Background()
			for i := 0; i < 5; i++ {
				_ = prod.RandomDataPoint(ctx)
			}

			// Device count should remain the same
			Expect(len(prod.IoTDevices)).To(Equal(initialCount))
		})
	})

	Describe("Concurrent Access", func() {
		It("should handle concurrent RandomDataPoint calls", func() {
			mqClient := mq.New("test-queue", "amqp://invalid:5672", logger)
			deviceMQClient := mq.New("device-queue", "amqp://invalid:5672", logger)
			defer func() {
				_ = mqClient.Close()
				_ = deviceMQClient.Close()
			}()

			prod := producer.NewProducer(mqClient, deviceMQClient)
			ctx := context.Background()

			// Launch multiple goroutines
			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func() {
					_ = prod.RandomDataPoint(ctx)
					done <- true
				}()
			}

			// Wait for all to complete
			for i := 0; i < 5; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})
})
