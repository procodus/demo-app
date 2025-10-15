package backend

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	"procodus.dev/demo-app/pkg/iot"
)

var _ = Describe("Backend Consumer E2E", func() {
	Context("Sensor Consumer", func() {
		It("should consume and save sensor reading messages", func() {
			ctx := context.Background()

			deviceID := "device-001"

			// Step 1: Create the device first (required due to foreign key constraint).
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Test Location",
				MacAddress: "AA:BB:CC:DD:EE:01",
				IpAddress:  "192.168.1.1",
				Firmware:   "v1.0.0",
				Latitude:   40.0,
				Longitude:  -120.0,
			}

			deviceBytes, err := proto.Marshal(device)
			Expect(err).NotTo(HaveOccurred())

			err = mqChannel.PublishWithContext(
				ctx,
				"",
				deviceQueueName,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/protobuf",
					Body:         deviceBytes,
					DeliveryMode: amqp.Persistent,
				},
			)
			Expect(err).NotTo(HaveOccurred())

			testLogger.Info("published device message", "device_id", deviceID)

			// Poll until device exists in database.
			Eventually(func() error {
				resp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
					DeviceId: deviceID,
				})
				if err != nil {
					return err
				}
				if resp.GetDevice() == nil {
					return fmt.Errorf("device not yet created")
				}
				return nil
			}, 10*time.Second, 500*time.Millisecond).Should(Succeed())

			testLogger.Info("device created, now publishing sensor reading")

			// Step 2: Create a sensor reading message for the device.
			sensorReading := &iot.SensorReading{
				DeviceId:     deviceID,
				Timestamp:    time.Now().Unix(),
				Temperature:  25.5,
				Humidity:     60.0,
				Pressure:     1013.25,
				BatteryLevel: 85.0,
			}

			// Marshal message.
			msgBytes, err := proto.Marshal(sensorReading)
			Expect(err).NotTo(HaveOccurred())

			// Publish message to sensor queue.
			err = mqChannel.PublishWithContext(
				ctx,
				"",              // exchange
				sensorQueueName, // routing key
				false,           // mandatory
				false,           // immediate
				amqp.Publishing{
					ContentType:  "application/protobuf",
					Body:         msgBytes,
					DeliveryMode: amqp.Persistent,
				},
			)
			Expect(err).NotTo(HaveOccurred())

			testLogger.Info("published sensor reading message", "device_id", sensorReading.GetDeviceId())

			// Poll until sensor reading appears in database.
			Eventually(func() int {
				resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId: deviceID,
				})
				if err != nil {
					return 0
				}
				return len(resp.GetReading())
			}, 30*time.Second, 500*time.Millisecond).Should(BeNumerically(">=", 1))

			// Verify the data.
			resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.GetReading()).NotTo(BeEmpty())

			reading := resp.GetReading()[0]
			Expect(reading.GetDeviceId()).To(Equal(deviceID))
			Expect(reading.GetTemperature()).To(BeNumerically("~", 25.5, 0.01))
			Expect(reading.GetHumidity()).To(BeNumerically("~", 60.0, 0.01))
			Expect(reading.GetPressure()).To(BeNumerically("~", 1013.25, 0.01))
			Expect(reading.GetBatteryLevel()).To(BeNumerically("~", 85.0, 0.01))

			testLogger.Info("sensor reading successfully consumed and saved")
		})

		It("should consume and save multiple sensor readings", func() {
			ctx := context.Background()

			deviceID := "device-002"
			numReadings := 5

			// Step 1: Create the device first.
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Test Location 2",
				MacAddress: "AA:BB:CC:DD:EE:02",
				IpAddress:  "192.168.1.2",
				Firmware:   "v1.0.0",
				Latitude:   40.1,
				Longitude:  -120.1,
			}

			deviceBytes, err := proto.Marshal(device)
			Expect(err).NotTo(HaveOccurred())

			err = mqChannel.PublishWithContext(
				ctx,
				"",
				deviceQueueName,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/protobuf",
					Body:         deviceBytes,
					DeliveryMode: amqp.Persistent,
				},
			)
			Expect(err).NotTo(HaveOccurred())

			testLogger.Info("published device message", "device_id", deviceID)

			// Poll until device exists in database.
			Eventually(func() error {
				resp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
					DeviceId: deviceID,
				})
				if err != nil {
					return err
				}
				if resp.GetDevice() == nil {
					return fmt.Errorf("device not yet created")
				}
				return nil
			}, 10*time.Second, 500*time.Millisecond).Should(Succeed())

			testLogger.Info("device created, now publishing sensor readings")

			// Step 2: Publish multiple sensor readings.
			for i := 0; i < numReadings; i++ {
				sensorReading := &iot.SensorReading{
					DeviceId:     deviceID,
					Timestamp:    time.Now().Add(time.Duration(i) * time.Second).Unix(),
					Temperature:  20.0 + float64(i),
					Humidity:     50.0 + float64(i),
					Pressure:     1000.0 + float64(i),
					BatteryLevel: 80.0 - float64(i),
				}

				msgBytes, err := proto.Marshal(sensorReading)
				Expect(err).NotTo(HaveOccurred())

				err = mqChannel.PublishWithContext(
					ctx,
					"",
					sensorQueueName,
					false,
					false,
					amqp.Publishing{
						ContentType:  "application/protobuf",
						Body:         msgBytes,
						DeliveryMode: amqp.Persistent,
					},
				)
				Expect(err).NotTo(HaveOccurred())
			}

			testLogger.Info("published multiple sensor readings", "count", numReadings, "device_id", deviceID)

			// Poll until all sensor readings appear in database.
			Eventually(func() int {
				resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId: deviceID,
				})
				if err != nil {
					return 0
				}
				return len(resp.GetReading())
			}, 30*time.Second, 500*time.Millisecond).Should(BeNumerically(">=", numReadings))

			testLogger.Info("multiple sensor readings successfully consumed and saved")
		})

		It("should handle sensor readings from different devices", func() {
			ctx := context.Background()

			devices := []string{"device-003", "device-004", "device-005"}

			// Step 1: Create all devices first.
			for i, deviceID := range devices {
				device := &iot.IoTDevice{
					DeviceId:   deviceID,
					Timestamp:  time.Now().Unix(),
					Location:   fmt.Sprintf("Location %d", i+3),
					MacAddress: fmt.Sprintf("AA:BB:CC:DD:EE:%02d", i+3),
					IpAddress:  fmt.Sprintf("192.168.1.%d", i+3),
					Firmware:   "v1.0.0",
					Latitude:   float32(40.0 + i),
					Longitude:  float32(-120.0 + i),
				}

				deviceBytes, err := proto.Marshal(device)
				Expect(err).NotTo(HaveOccurred())

				err = mqChannel.PublishWithContext(
					ctx,
					"",
					deviceQueueName,
					false,
					false,
					amqp.Publishing{
						ContentType:  "application/protobuf",
						Body:         deviceBytes,
						DeliveryMode: amqp.Persistent,
					},
				)
				Expect(err).NotTo(HaveOccurred())
			}

			testLogger.Info("published device messages", "count", len(devices))

			// Poll until all devices exist in database.
			for _, deviceID := range devices {
				Eventually(func() error {
					resp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
						DeviceId: deviceID,
					})
					if err != nil {
						return err
					}
					if resp.GetDevice() == nil {
						return fmt.Errorf("device %s not yet created", deviceID)
					}
					return nil
				}, 10*time.Second, 500*time.Millisecond).Should(Succeed())
			}

			testLogger.Info("all devices created, now publishing sensor readings")

			// Step 2: Publish sensor readings from different devices.
			for _, deviceID := range devices {
				sensorReading := &iot.SensorReading{
					DeviceId:     deviceID,
					Timestamp:    time.Now().Unix(),
					Temperature:  25.0,
					Humidity:     55.0,
					Pressure:     1010.0,
					BatteryLevel: 90.0,
				}

				msgBytes, err := proto.Marshal(sensorReading)
				Expect(err).NotTo(HaveOccurred())

				err = mqChannel.PublishWithContext(
					ctx,
					"",
					sensorQueueName,
					false,
					false,
					amqp.Publishing{
						ContentType:  "application/protobuf",
						Body:         msgBytes,
						DeliveryMode: amqp.Persistent,
					},
				)
				Expect(err).NotTo(HaveOccurred())
			}

			testLogger.Info("published sensor readings from multiple devices", "count", len(devices))

			// Poll until each device has readings.
			for _, deviceID := range devices {
				Eventually(func() int {
					resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
						DeviceId: deviceID,
					})
					if err != nil {
						return 0
					}
					return len(resp.GetReading())
				}, 30*time.Second, 500*time.Millisecond).Should(BeNumerically(">=", 1))

				testLogger.Info("verified sensor reading for device", "device_id", deviceID)
			}

			testLogger.Info("sensor readings from multiple devices successfully processed")
		})
	})

	Context("Device Consumer", func() {
		It("should consume and save device creation messages", func() {
			ctx := context.Background()

			// Create a device message
			device := &iot.IoTDevice{
				DeviceId:   "device-101",
				Timestamp:  time.Now().Unix(),
				Location:   "Office A",
				MacAddress: "00:11:22:33:44:55",
				IpAddress:  "192.168.1.101",
				Firmware:   "v1.0.0",
				Latitude:   37.7749,
				Longitude:  -122.4194,
			}

			// Marshal message
			msgBytes, err := proto.Marshal(device)
			Expect(err).NotTo(HaveOccurred())

			// Publish message to device queue
			err = mqChannel.PublishWithContext(
				ctx,
				"",
				deviceQueueName,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/protobuf",
					Body:         msgBytes,
					DeliveryMode: amqp.Persistent,
				},
			)
			Expect(err).NotTo(HaveOccurred())

			testLogger.Info("published device creation message", "device_id", device.GetDeviceId())

			// Wait for message to be consumed and processed
			time.Sleep(3 * time.Second)

			// Verify device was saved via gRPC API
			resp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
				DeviceId: "device-101",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.GetDevice()).NotTo(BeNil())

			// Verify device data
			savedDevice := resp.GetDevice()
			Expect(savedDevice.GetDeviceId()).To(Equal("device-101"))
			Expect(savedDevice.GetLocation()).To(Equal("Office A"))
			Expect(savedDevice.GetMacAddress()).To(Equal("00:11:22:33:44:55"))
			Expect(savedDevice.GetIpAddress()).To(Equal("192.168.1.101"))
			Expect(savedDevice.GetFirmware()).To(Equal("v1.0.0"))
			Expect(savedDevice.GetLatitude()).To(BeNumerically("~", 37.7749, 0.0001))
			Expect(savedDevice.GetLongitude()).To(BeNumerically("~", -122.4194, 0.0001))

			testLogger.Info("device successfully consumed and saved")
		})

		It("should update existing device on duplicate message (upsert)", func() {
			ctx := context.Background()

			deviceID := "device-102"

			// Create initial device
			device1 := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Office B",
				MacAddress: "00:11:22:33:44:66",
				IpAddress:  "192.168.1.102",
				Firmware:   "v1.0.0",
				Latitude:   37.7750,
				Longitude:  -122.4195,
			}

			msgBytes, err := proto.Marshal(device1)
			Expect(err).NotTo(HaveOccurred())

			err = mqChannel.PublishWithContext(
				ctx,
				"",
				deviceQueueName,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/protobuf",
					Body:         msgBytes,
					DeliveryMode: amqp.Persistent,
				},
			)
			Expect(err).NotTo(HaveOccurred())

			testLogger.Info("published initial device message", "device_id", deviceID)
			time.Sleep(3 * time.Second)

			// Create updated device (same ID, different firmware and location)
			device2 := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Office C",
				MacAddress: "00:11:22:33:44:66",
				IpAddress:  "192.168.1.102",
				Firmware:   "v2.0.0", // Updated firmware
				Latitude:   37.7751,  // Updated location
				Longitude:  -122.4196,
			}

			msgBytes2, err := proto.Marshal(device2)
			Expect(err).NotTo(HaveOccurred())

			err = mqChannel.PublishWithContext(
				ctx,
				"",
				deviceQueueName,
				false,
				false,
				amqp.Publishing{
					ContentType:  "application/protobuf",
					Body:         msgBytes2,
					DeliveryMode: amqp.Persistent,
				},
			)
			Expect(err).NotTo(HaveOccurred())

			testLogger.Info("published updated device message", "device_id", deviceID)
			time.Sleep(3 * time.Second)

			// Verify device was updated
			resp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.GetDevice()).NotTo(BeNil())

			savedDevice := resp.GetDevice()
			Expect(savedDevice.GetDeviceId()).To(Equal(deviceID))
			Expect(savedDevice.GetLocation()).To(Equal("Office C")) // Should be updated.
			Expect(savedDevice.GetFirmware()).To(Equal("v2.0.0"))   // Should be updated.
			Expect(savedDevice.GetLatitude()).To(BeNumerically("~", 37.7751, 0.0001))
			Expect(savedDevice.GetLongitude()).To(BeNumerically("~", -122.4196, 0.0001))

			testLogger.Info("device successfully updated via upsert")
		})

		It("should consume and save multiple devices", func() {
			ctx := context.Background()

			numDevices := 5

			// Publish multiple devices
			for i := 0; i < numDevices; i++ {
				device := &iot.IoTDevice{
					DeviceId:   fmt.Sprintf("device-20%d", i),
					Timestamp:  time.Now().Unix(),
					Location:   fmt.Sprintf("Location %d", i),
					MacAddress: fmt.Sprintf("00:11:22:33:44:%02d", i),
					IpAddress:  fmt.Sprintf("192.168.1.%d", 200+i),
					Firmware:   "v1.0.0",
					Latitude:   37.0 + float32(i)*0.1,
					Longitude:  -122.0 + float32(i)*0.1,
				}

				msgBytes, err := proto.Marshal(device)
				Expect(err).NotTo(HaveOccurred())

				err = mqChannel.PublishWithContext(
					ctx,
					"",
					deviceQueueName,
					false,
					false,
					amqp.Publishing{
						ContentType:  "application/protobuf",
						Body:         msgBytes,
						DeliveryMode: amqp.Persistent,
					},
				)
				Expect(err).NotTo(HaveOccurred())
			}

			testLogger.Info("published multiple device messages", "count", numDevices)

			// Wait for all messages to be processed
			time.Sleep(3 * time.Second)

			// Verify all devices were saved
			resp, err := grpcClient.GetAllDevice(ctx, &iot.GetAllDevicesRequest{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.GetDevices()).NotTo(BeEmpty())
			Expect(len(resp.GetDevices())).To(BeNumerically(">=", numDevices))

			testLogger.Info("multiple devices successfully consumed and saved")
		})
	})
})
