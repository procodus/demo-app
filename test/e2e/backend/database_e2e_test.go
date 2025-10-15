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

var _ = Describe("Backend Database Relationships E2E", func() {
	Context("Device and SensorReading Relationship", func() {
		It("should maintain one-to-many relationship (device has many sensor readings)", func() {
			ctx := context.Background()

			deviceID := "db-device-001"

			// Step 1: Create a device
			device := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Database Test Location",
				MacAddress: "AA:BB:CC:DD:EE:01",
				IpAddress:  "192.168.10.1",
				Firmware:   "v1.0.0",
				Latitude:   40.7128,
				Longitude:  -74.0060,
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

			testLogger.Info("published device for relationship test", "device_id", deviceID)
			time.Sleep(3 * time.Second)

			// Step 2: Create multiple sensor readings for this device
			numReadings := 10
			for i := 0; i < numReadings; i++ {
				reading := &iot.SensorReading{
					DeviceId:     deviceID,
					Timestamp:    time.Now().Add(time.Duration(i) * time.Minute).Unix(),
					Temperature:  20.0 + float64(i)*0.5,
					Humidity:     50.0 + float64(i)*2.0,
					Pressure:     1000.0 + float64(i)*5.0,
					BatteryLevel: 100.0 - float64(i)*5.0,
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
			time.Sleep(3 * time.Second)

			// Step 3: Verify device exists
			deviceResp, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(deviceResp.GetDevice()).NotTo(BeNil())
			Expect(deviceResp.GetDevice().GetDeviceId()).To(Equal(deviceID))

			// Step 4: Verify all sensor readings belong to this device
			readingsResp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(readingsResp.GetReading()).NotTo(BeEmpty())
			Expect(len(readingsResp.GetReading())).To(BeNumerically(">=", numReadings))

			// Verify all readings have correct device_id
			for _, reading := range readingsResp.GetReading() {
				Expect(reading.GetDeviceId()).To(Equal(deviceID))
			}

			testLogger.Info("verified one-to-many relationship: device has many sensor readings")
		})

		It("should handle sensor readings for multiple devices independently", func() {
			ctx := context.Background()

			// Create multiple devices
			devices := []struct {
				ID       string
				Location string
			}{
				{"db-device-101", "Location A"},
				{"db-device-102", "Location B"},
				{"db-device-103", "Location C"},
			}

			// Create each device
			for i, d := range devices {
				device := &iot.IoTDevice{
					DeviceId:   d.ID,
					Timestamp:  time.Now().Unix(),
					Location:   d.Location,
					MacAddress: fmt.Sprintf("AA:BB:CC:DD:EE:%02d", 10+i),
					IpAddress:  fmt.Sprintf("192.168.20.%d", 10+i),
					Firmware:   "v1.0.0",
					Latitude:   float32(40.0 + i),
					Longitude:  float32(-74.0 + i),
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

			testLogger.Info("published multiple devices", "count", len(devices))
			time.Sleep(3 * time.Second)

			// Create different numbers of sensor readings for each device
			readingCounts := map[string]int{
				"db-device-101": 5,
				"db-device-102": 8,
				"db-device-103": 3,
			}

			for deviceID, count := range readingCounts {
				for i := 0; i < count; i++ {
					reading := &iot.SensorReading{
						DeviceId:     deviceID,
						Timestamp:    time.Now().Add(time.Duration(i) * time.Minute).Unix(),
						Temperature:  25.0,
						Humidity:     60.0,
						Pressure:     1015.0,
						BatteryLevel: 90.0,
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

				testLogger.Info("published sensor readings for device", "device_id", deviceID, "count", count)
			}

			time.Sleep(3 * time.Second)

			// Verify each device has the correct number of readings
			for deviceID, expectedCount := range readingCounts {
				resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId: deviceID,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.GetReading()).NotTo(BeEmpty())
				Expect(len(resp.GetReading())).To(BeNumerically(">=", expectedCount))

				// Verify all readings belong to the correct device
				for _, reading := range resp.GetReading() {
					Expect(reading.GetDeviceId()).To(Equal(deviceID))
				}

				testLogger.Info("verified readings for device", "device_id", deviceID, "count", len(resp.GetReading()))
			}

			testLogger.Info("verified multiple devices have independent sensor readings")
		})

		It("should handle sensor readings without corresponding device", func() {
			ctx := context.Background()

			// Publish sensor reading for non-existent device
			// This tests if the foreign key relationship allows orphaned readings
			// or if the system creates a device implicitly
			orphanDeviceID := "db-orphan-device-999"

			reading := &iot.SensorReading{
				DeviceId:     orphanDeviceID,
				Timestamp:    time.Now().Unix(),
				Temperature:  22.0,
				Humidity:     55.0,
				Pressure:     1012.0,
				BatteryLevel: 88.0,
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

			testLogger.Info("published sensor reading for orphan device", "device_id", orphanDeviceID)
			time.Sleep(3 * time.Second)

			// Try to get readings for orphan device
			resp, err := grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
				DeviceId: orphanDeviceID,
			})

			// The behavior depends on foreign key constraints
			// If foreign key is enforced, reading won't be saved
			// If not enforced, reading will exist without parent device
			if err == nil && len(resp.GetReading()) > 0 {
				testLogger.Info("orphan sensor reading was saved (no FK constraint enforcement)")
			} else {
				testLogger.Info("orphan sensor reading was not saved or returned empty (FK constraint enforced)")
			}
		})

		It("should preserve device data integrity when sensor readings are added", func() {
			ctx := context.Background()

			deviceID := "db-device-201"

			// Create device with specific attributes
			originalDevice := &iot.IoTDevice{
				DeviceId:   deviceID,
				Timestamp:  time.Now().Unix(),
				Location:   "Original Location",
				MacAddress: "BB:BB:BB:BB:BB:01",
				IpAddress:  "192.168.30.1",
				Firmware:   "v1.0.0",
				Latitude:   35.0,
				Longitude:  -115.0,
			}

			msgBytes, err := proto.Marshal(originalDevice)
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

			testLogger.Info("published device for integrity test", "device_id", deviceID)
			time.Sleep(3 * time.Second)

			// Verify device initial state
			deviceResp1, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(deviceResp1.GetDevice()).NotTo(BeNil())

			originalLocation := deviceResp1.GetDevice().GetLocation()
			originalFirmware := deviceResp1.GetDevice().GetFirmware()

			// Add sensor readings
			for i := 0; i < 5; i++ {
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

			time.Sleep(3 * time.Second)

			// Verify device attributes haven't changed after adding sensor readings
			deviceResp2, err := grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
				DeviceId: deviceID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(deviceResp2.GetDevice()).NotTo(BeNil())

			Expect(deviceResp2.GetDevice().GetLocation()).To(Equal(originalLocation))
			Expect(deviceResp2.GetDevice().GetFirmware()).To(Equal(originalFirmware))
			Expect(deviceResp2.GetDevice().GetDeviceId()).To(Equal(deviceID))

			testLogger.Info("verified device data integrity preserved after adding sensor readings")
		})
	})
})
