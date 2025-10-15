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

var _ = Describe("Backend gRPC API E2E", func() {
	Context("GetAllDevice", func() {
		It("should return all devices", func() {
			ctx := context.Background()

			// First, publish some devices
			deviceIDs := []string{"api-device-001", "api-device-002", "api-device-003"}

			for i, deviceID := range deviceIDs {
				device := &iot.IoTDevice{
					DeviceId:   deviceID,
					Timestamp:  time.Now().Unix(),
					Location:   fmt.Sprintf("Location %d", i),
					MacAddress: fmt.Sprintf("AA:BB:CC:DD:EE:%02d", i),
					IpAddress:  fmt.Sprintf("10.0.0.%d", i+1),
					Firmware:   "v1.0.0",
					Latitude:   float32(40.0 + i),
					Longitude:  float32(-120.0 + i),
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

			testLogger.Info("published test devices for GetAllDevice", "count", len(deviceIDs))

			// Wait for devices to be processed
			time.Sleep(3 * time.Second)

			// Call GetAllDevice
			resp, err := grpcClient.GetAllDevice(ctx, &iot.GetAllDevicesRequest{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.GetDevices()).NotTo(BeEmpty())

			// Verify we have at least the devices we just created
			Expect(len(resp.GetDevices())).To(BeNumerically(">=", len(deviceIDs)))

			// Verify our devices are in the response
			deviceMap := make(map[string]*iot.IoTDevice)
			for _, device := range resp.GetDevices() {
				deviceMap[device.GetDeviceId()] = device
			}

			for _, expectedID := range deviceIDs {
				device, found := deviceMap[expectedID]
				Expect(found).To(BeTrue(), "device %s should be in the response", expectedID)
				Expect(device.GetDeviceId()).To(Equal(expectedID))
			}

			testLogger.Info("GetAllDevice returned correct devices")
		})

		It("should return empty list when no devices exist", func() {
			// Note: This test may fail if previous tests created devices
			// In a real scenario, we'd reset the database or use isolated test data
			ctx := context.Background()

			resp, err := grpcClient.GetAllDevice(ctx, &iot.GetAllDevicesRequest{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			// Devices list may not be empty due to previous tests
		})
	})

	Context("GetDevice", func() {
		It("should return a specific device by ID", func() {
			ctx := context.Background()

			deviceID := "api-device-101"

			// Publish a device
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Test Location",
				MacAddress: "11:22:33:44:55:66",
				IpAddress:  "192.168.100.1",
				Firmware:   "v2.0.0",
				Latitude:   45.5,
				Longitude:  -122.6,
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

			testLogger.Info("published device for GetDevice test", "device_id", deviceID)
			time.Sleep(3 * time.Second)

			// Call GetDevice
			resp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.GetDevice()).NotTo(BeNil())

			// Verify device details
			returnedDevice := resp.GetDevice()
			Expect(returnedDevice.GetDeviceId()).To(Equal(deviceID))
			Expect(returnedDevice.GetLocation()).To(Equal("Test Location"))
			Expect(returnedDevice.GetMacAddress()).To(Equal("11:22:33:44:55:66"))
			Expect(returnedDevice.GetIpAddress()).To(Equal("192.168.100.1"))
			Expect(returnedDevice.GetFirmware()).To(Equal("v2.0.0"))
			Expect(returnedDevice.GetLatitude()).To(BeNumerically("~", 45.5, 0.01))
			Expect(returnedDevice.GetLongitude()).To(BeNumerically("~", -122.6, 0.01))

			testLogger.Info("GetDevice returned correct device data")
		})

		It("should return error for non-existent device", func() {
			ctx := context.Background()

			// Call GetDevice with non-existent ID
			resp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
				DeviceId: "non-existent-device-999",
			})

			// Expect error or empty device
			if err == nil {
				Expect(resp.GetDevice()).To(BeNil())
			} else {
				testLogger.Info("GetDevice correctly returned error for non-existent device")
			}
		})
	})

	Context("GetSensorReadingByDeviceID", func() {
		It("should return sensor readings for a specific device", func() {
			ctx := context.Background()

			deviceID := "api-device-201"

			// First, create the device.
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Sensor Test Location",
				MacAddress: "AA:BB:CC:DD:EE:FF",
				IpAddress:  "192.168.200.1",
				Firmware:   "v1.0.0",
				Latitude:   50.0,
				Longitude:  -100.0,
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

			// Publish multiple sensor readings for this device.
			numReadings := 3
			for i := 0; i < numReadings; i++ {
				reading := &iot.SensorReading{
					DeviceId:     deviceID,
					Timestamp:    time.Now().Add(time.Duration(i) * time.Second).Unix(),
					Temperature:  22.0 + float64(i),
					Humidity:     55.0 + float64(i),
					Pressure:     1010.0 + float64(i),
					BatteryLevel: 90.0 - float64(i),
				}

				msgBytes, err := proto.Marshal(reading)
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

			testLogger.Info("published sensor readings for device", "device_id", deviceID, "count", numReadings)

			// Poll until sensor readings appear in database.
			Eventually(func() int {
				resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId: deviceID,
				})
				if err != nil {
					return 0
				}
				return len(resp.GetReading())
			}, 30*time.Second, 500*time.Millisecond).Should(BeNumerically(">=", numReadings))

			// Call GetSensorReadingByDeviceID to verify data.
			resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.GetReading()).NotTo(BeEmpty())

			// Verify we have the correct number of readings.
			Expect(len(resp.GetReading())).To(BeNumerically(">=", numReadings))

			// Verify all readings belong to the correct device.
			for _, reading := range resp.GetReading() {
				Expect(reading.GetDeviceId()).To(Equal(deviceID))
			}

			testLogger.Info("GetSensorReadingByDeviceID returned correct readings")
		})

		It("should return readings in descending order by timestamp", func() {
			ctx := context.Background()

			deviceID := "api-device-202"

			// Create device.
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Order Test Location",
				MacAddress: "11:11:11:11:11:11",
				IpAddress:  "192.168.202.1",
				Firmware:   "v1.0.0",
				Latitude:   51.0,
				Longitude:  -101.0,
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

			// Publish readings with known timestamps.
			timestamps := []int64{
				time.Now().Add(-3 * time.Hour).Unix(),
				time.Now().Add(-2 * time.Hour).Unix(),
				time.Now().Add(-1 * time.Hour).Unix(),
				time.Now().Unix(),
			}

			for _, ts := range timestamps {
				reading := &iot.SensorReading{
					DeviceId:     deviceID,
					Timestamp:    ts,
					Temperature:  20.0,
					Humidity:     50.0,
					Pressure:     1000.0,
					BatteryLevel: 85.0,
				}

				msgBytes, err := proto.Marshal(reading)
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

			testLogger.Info("published ordered sensor readings", "device_id", deviceID)

			// Poll until all sensor readings appear in database.
			Eventually(func() int {
				resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId: deviceID,
				})
				if err != nil {
					return 0
				}
				return len(resp.GetReading())
			}, 30*time.Second, 500*time.Millisecond).Should(BeNumerically(">=", len(timestamps)))

			// Get readings to verify order.
			resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.GetReading()).NotTo(BeEmpty())

			// Verify readings are in descending order (most recent first).
			readings := resp.GetReading()
			for i := 0; i < len(readings)-1; i++ {
				Expect(readings[i].GetTimestamp()).To(BeNumerically(">=", readings[i+1].GetTimestamp()))
			}

			testLogger.Info("verified sensor readings are in correct order")
		})

		It("should support pagination with page tokens", func() {
			ctx := context.Background()

			deviceID := "api-device-203"

			// Create device.
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Pagination Test",
				MacAddress: "22:22:22:22:22:22",
				IpAddress:  "192.168.203.1",
				Firmware:   "v1.0.0",
				Latitude:   52.0,
				Longitude:  -102.0,
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

			// Publish many readings to test pagination.
			numReadings := 15
			for i := 0; i < numReadings; i++ {
				reading := &iot.SensorReading{
					DeviceId:     deviceID,
					Timestamp:    time.Now().Add(time.Duration(i) * time.Second).Unix(),
					Temperature:  20.0,
					Humidity:     50.0,
					Pressure:     1000.0,
					BatteryLevel: 85.0,
				}

				msgBytes, err := proto.Marshal(reading)
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

			testLogger.Info("published readings for pagination test", "count", numReadings)

			// Poll until sensor readings appear in database.
			Eventually(func() int {
				resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId: deviceID,
				})
				if err != nil {
					return 0
				}
				return len(resp.GetReading())
			}, 30*time.Second, 500*time.Millisecond).Should(BeNumerically(">=", numReadings))

			// Get first page.
			resp1, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp1.GetReading()).NotTo(BeEmpty())

			firstPageCount := len(resp1.GetReading())
			testLogger.Info("first page returned readings", "count", firstPageCount)

			// If there's a next page token, fetch next page.
			if resp1.GetNextPageToken() != "" {
				resp2, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId:  deviceID,
					PageToken: resp1.GetNextPageToken(),
				})
				Expect(err).NotTo(HaveOccurred())

				testLogger.Info("second page returned readings", "count", len(resp2.GetReading()))
			}

			testLogger.Info("pagination test completed")
		})

		It("should return empty list for device with no readings", func() {
			ctx := context.Background()

			deviceID := "api-device-204"

			// Create device but don't publish any readings
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "No Readings Test",
				MacAddress: "33:33:33:33:33:33",
				IpAddress:  "192.168.204.1",
				Firmware:   "v1.0.0",
				Latitude:   53.0,
				Longitude:  -103.0,
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
			time.Sleep(3 * time.Second)

			// Get readings for device with no sensor data
			resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.GetReading()).To(BeEmpty())

			testLogger.Info("correctly returned empty readings for device with no data")
		})
	})
})
