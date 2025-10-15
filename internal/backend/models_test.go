package backend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/backend"
)

var _ = Describe("Models", func() {
	Describe("SensorReading", func() {
		Context("table name", func() {
			It("should return sensor_readings", func() {
				reading := backend.SensorReading{}
				Expect(reading.TableName()).To(Equal("sensor_readings"))
			})
		})

		Context("struct initialization", func() {
			It("should initialize with zero values", func() {
				reading := backend.SensorReading{}
				Expect(reading.DeviceID).To(BeEmpty())
				Expect(reading.Temperature).To(BeZero())
				Expect(reading.Humidity).To(BeZero())
				Expect(reading.Pressure).To(BeZero())
				Expect(reading.BatteryLevel).To(BeZero())
				Expect(reading.ID).To(BeZero())
			})

			It("should allow setting values", func() {
				reading := backend.SensorReading{
					DeviceID:     "device-001",
					Temperature:  22.5,
					Humidity:     65.0,
					Pressure:     1013.25,
					BatteryLevel: 85.0,
				}

				Expect(reading.DeviceID).To(Equal("device-001"))
				Expect(reading.Temperature).To(Equal(22.5))
				Expect(reading.Humidity).To(Equal(65.0))
				Expect(reading.Pressure).To(Equal(1013.25))
				Expect(reading.BatteryLevel).To(Equal(85.0))
			})
		})

		Context("field types", func() {
			It("should have correct field types", func() {
				reading := backend.SensorReading{
					DeviceID:     "device-001",
					Temperature:  -15.5,
					Humidity:     100.0,
					Pressure:     950.0,
					BatteryLevel: 0.0,
				}

				Expect(reading.DeviceID).To(BeAssignableToTypeOf(""))
				Expect(reading.Temperature).To(BeAssignableToTypeOf(float64(0)))
				Expect(reading.Humidity).To(BeAssignableToTypeOf(float64(0)))
				Expect(reading.Pressure).To(BeAssignableToTypeOf(float64(0)))
				Expect(reading.BatteryLevel).To(BeAssignableToTypeOf(float64(0)))
			})
		})
	})

	Describe("IoTDevice", func() {
		Context("table name", func() {
			It("should return iot_devices", func() {
				device := backend.IoTDevice{}
				Expect(device.TableName()).To(Equal("iot_devices"))
			})
		})

		Context("struct initialization", func() {
			It("should initialize with zero values", func() {
				device := backend.IoTDevice{}
				Expect(device.DeviceID).To(BeEmpty())
				Expect(device.Location).To(BeEmpty())
				Expect(device.MACAddress).To(BeEmpty())
				Expect(device.IPAddress).To(BeEmpty())
				Expect(device.Firmware).To(BeEmpty())
				Expect(device.Latitude).To(BeZero())
				Expect(device.Longitude).To(BeZero())
				Expect(device.ID).To(BeZero())
			})

			It("should allow setting values", func() {
				device := backend.IoTDevice{
					DeviceID:   "device-001",
					Location:   "Building A - Floor 2",
					MACAddress: "00:1B:44:11:3A:B7",
					IPAddress:  "192.168.1.100",
					Firmware:   "v1.2.3",
					Latitude:   37.7749,
					Longitude:  -122.4194,
				}

				Expect(device.DeviceID).To(Equal("device-001"))
				Expect(device.Location).To(Equal("Building A - Floor 2"))
				Expect(device.MACAddress).To(Equal("00:1B:44:11:3A:B7"))
				Expect(device.IPAddress).To(Equal("192.168.1.100"))
				Expect(device.Firmware).To(Equal("v1.2.3"))
				Expect(device.Latitude).To(BeNumerically("~", 37.7749, 0.0001))
				Expect(device.Longitude).To(BeNumerically("~", -122.4194, 0.0001))
			})
		})

		Context("field types", func() {
			It("should have correct field types", func() {
				device := backend.IoTDevice{
					DeviceID:   "device-001",
					Location:   "Location",
					MACAddress: "MAC",
					IPAddress:  "IP",
					Firmware:   "v1.0.0",
					Latitude:   0.0,
					Longitude:  0.0,
				}

				Expect(device.DeviceID).To(BeAssignableToTypeOf(""))
				Expect(device.Location).To(BeAssignableToTypeOf(""))
				Expect(device.MACAddress).To(BeAssignableToTypeOf(""))
				Expect(device.IPAddress).To(BeAssignableToTypeOf(""))
				Expect(device.Firmware).To(BeAssignableToTypeOf(""))
				Expect(device.Latitude).To(BeAssignableToTypeOf(float32(0)))
				Expect(device.Longitude).To(BeAssignableToTypeOf(float32(0)))
			})
		})

		Context("coordinate validation", func() {
			It("should accept valid coordinates", func() {
				coords := []struct {
					lat float32
					lng float32
				}{
					{0.0, 0.0},
					{90.0, 180.0},
					{-90.0, -180.0},
					{37.7749, -122.4194},
				}

				for _, coord := range coords {
					device := backend.IoTDevice{
						DeviceID:   "device-001",
						Location:   "Location",
						MACAddress: "MAC",
						IPAddress:  "IP",
						Firmware:   "v1.0.0",
						Latitude:   coord.lat,
						Longitude:  coord.lng,
					}

					Expect(device.Latitude).To(Equal(coord.lat))
					Expect(device.Longitude).To(Equal(coord.lng))
				}
			})
		})
	})
})
