package producer_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/producer"
	"procodus.dev/demo-app/pkg/mq"
	"procodus.dev/demo-app/pkg/mq/mock"
)

var _ = Describe("IoT Producer", func() {
	var (
		mqClient       mq.ClientInterface
		deviceMQClient mq.ClientInterface
	)

	Describe("NewProducer", func() {
		BeforeEach(func() {
			// Create mock MQ clients for unit tests
			mqClient = mock.NewMockClient()
			deviceMQClient = mock.NewMockClient()
		})

		AfterEach(func() {
			// No cleanup needed for mocks
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
			mqClient = mock.NewMockClient()
			deviceMQClient = mock.NewMockClient()
			prod = producer.NewProducer(mqClient, deviceMQClient)
		})

		Context("with successful push", func() {
			It("should successfully push data", func() {
				ctx := context.Background()
				err := prod.RandomDataPoint(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify Push was called
				mockClient := mqClient.(*mock.MockClient)
				Expect(mockClient.PushCalls).To(HaveLen(1))
			})
		})

		Context("with context", func() {
			It("should accept a context parameter", func() {
				ctx := context.Background()
				err := prod.RandomDataPoint(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify context was passed through
				mockClient := mqClient.(*mock.MockClient)
				Expect(mockClient.PushCalls).To(HaveLen(1))
				Expect(mockClient.PushCalls[0].Ctx).To(Equal(ctx))
			})

			It("should accept a canceled context", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				err := prod.RandomDataPoint(ctx)
				// Should pass context through even if canceled
				Expect(err).NotTo(HaveOccurred())

				mockClient := mqClient.(*mock.MockClient)
				Expect(mockClient.PushCalls).To(HaveLen(1))
			})
		})
	})

	Describe("Producer Integration", func() {
		It("should have valid device data structure", func() {
			mockClient := mock.NewMockClient()
			mockDeviceClient := mock.NewMockClient()

			prod := producer.NewProducer(mockClient, mockDeviceClient)

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
			mockClient := mock.NewMockClient()
			mockDeviceClient := mock.NewMockClient()

			prod := producer.NewProducer(mockClient, mockDeviceClient)
			initialCount := len(prod.IoTDevices)

			// Call RandomDataPoint multiple times
			ctx := context.Background()
			for i := 0; i < 5; i++ {
				err := prod.RandomDataPoint(ctx)
				Expect(err).NotTo(HaveOccurred())
			}

			// Device count should remain the same
			Expect(len(prod.IoTDevices)).To(Equal(initialCount))

			// Verify Push was called 5 times
			Expect(mockClient.PushCalls).To(HaveLen(5))
		})
	})

	Describe("Concurrent Access", func() {
		It("should handle concurrent RandomDataPoint calls", func() {
			mockClient := mock.NewMockClient()
			mockDeviceClient := mock.NewMockClient()

			prod := producer.NewProducer(mockClient, mockDeviceClient)
			ctx := context.Background()

			// Launch multiple goroutines
			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func() {
					err := prod.RandomDataPoint(ctx)
					Expect(err).NotTo(HaveOccurred())
					done <- true
				}()
			}

			// Wait for all to complete
			for i := 0; i < 5; i++ {
				Eventually(done).Should(Receive())
			}

			// Verify all 5 calls were made (MockClient is thread-safe)
			Expect(mockClient.PushCalls).To(HaveLen(5))
		})
	})
})
