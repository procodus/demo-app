// Package producer provides IoT data generation and publishing functionality.
package producer

import (
	"context"
	"math/rand"
	"time"

	"google.golang.org/protobuf/proto"

	"procodus.dev/demo-app/pkg/generator"
	"procodus.dev/demo-app/pkg/mq"
)

// Producer manages IoT devices and publishes sensor data to a message queue.
type Producer struct {
	MQClient   *mq.Client
	IoTDevices []*generator.IoTDevice
}

// NewProducer creates a new producer with a random number of IoT devices.
// Note: Uses math/rand for device generation which is acceptable for simulation data.
func NewProducer(mqClient *mq.Client) *Producer {
	deviceCount := rand.Intn(5) + 1 // #nosec G404 - weak random is acceptable for test data generation
	iotDevices := make([]*generator.IoTDevice, 0, deviceCount)
	for range deviceCount {
		iotDevices = append(iotDevices, generator.NewIoTDevice())
	}
	return &Producer{
		MQClient:   mqClient,
		IoTDevices: iotDevices,
	}
}

// RandomDataPoint generates a random sensor reading and publishes it to the message queue.
// Note: Uses math/rand for device selection which is acceptable for simulation data.
// The context parameter is currently unused but maintained for interface consistency
// and future compatibility when the MQ client is updated to support context-aware operations.
func (p *Producer) RandomDataPoint(_ context.Context) error {
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
	// TODO: Pass context to Push when MQ client supports context-aware operations
	//nolint:contextcheck // MQ client Push does not currently accept context
	return p.MQClient.Push(message)
}
