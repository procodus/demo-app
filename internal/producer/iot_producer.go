// Package producer provides IoT data generation and publishing functionality.
package producer

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"google.golang.org/protobuf/proto"

	"procodus.dev/demo-app/pkg/generator"
	"procodus.dev/demo-app/pkg/iot"
	"procodus.dev/demo-app/pkg/mq"
)

// Producer manages IoT devices and publishes sensor data to a message queue.
type Producer struct {
	MQClient       mq.ClientInterface
	DeviceMQClient mq.ClientInterface
	IoTDevices     []*generator.IoTDevice
}

// NewProducer creates a new producer with a random number of IoT devices.
// It publishes device creation messages for each device.
// Note: Uses math/rand for device generation which is acceptable for simulation data.
func NewProducer(mqClient mq.ClientInterface, deviceMQClient mq.ClientInterface) *Producer {
	deviceCount := rand.Intn(5) + 1 // #nosec G404 - weak random is acceptable for test data generation
	iotDevices := make([]*generator.IoTDevice, 0, deviceCount)
	for range deviceCount {
		iotDevices = append(iotDevices, generator.NewIoTDevice())
	}

	producer := &Producer{
		MQClient:       mqClient,
		DeviceMQClient: deviceMQClient,
		IoTDevices:     iotDevices,
	}

	// Publish device creation messages
	for _, device := range iotDevices {
		if err := producer.publishDeviceCreation(device); err != nil {
			// Log error but continue with other devices
			slog.Error(err.Error())
			continue
		}
	}

	return producer
}

// publishDeviceCreation publishes an IoT device creation message to the device queue.
func (p *Producer) publishDeviceCreation(device *generator.IoTDevice) error {
	// Transform generator.IoTDevice to proto iot.IoTDevice
	protoDevice := &iot.IoTDevice{
		DeviceId:   device.DeviceID,
		Timestamp:  device.Timestamp.Unix(),
		Location:   device.Location,
		MacAddress: device.MacAddress,
		IpAddress:  device.IPAddress,
		Firmware:   device.Firmware,
		Latitude:   float32(device.Latitude),
		Longitude:  float32(device.Longitude),
	}

	// Marshal to protobuf
	message, err := proto.Marshal(protoDevice)
	if err != nil {
		return err
	}

	// Publish to device queue with timeout
	// Use very short timeout to avoid blocking during initialization in tests
	// Background reconnection will handle subsequent operations once connection is established
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	return p.DeviceMQClient.Push(ctx, message)
}

// RandomDataPoint generates a random sensor reading and publishes it to the message queue.
// Note: Uses math/rand for device selection which is acceptable for simulation data.
func (p *Producer) RandomDataPoint(ctx context.Context) error {
	// Select a random device
	deviceID := p.IoTDevices[rand.Intn(len(p.IoTDevices))].DeviceID // #nosec G404 - weak random is acceptable for simulation

	// Generate sensor reading
	iotDataGen := generator.NewIoTGenerator(deviceID)
	reading := iotDataGen.GenerateCorrelatedReading(time.Now())

	// Marshal to protobuf
	message, err := proto.Marshal(reading)
	if err != nil {
		return err
	}

	// Publish to message queue
	return p.MQClient.Push(ctx, message)
}
