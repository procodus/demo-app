package producer

import (
	"math/rand"
	"time"

	"procodus.dev/demo-app/pkg/generator"
	"procodus.dev/demo-app/pkg/mq"
)

type Producer struct {
	MQClient   *mq.Client
	IoTDevices []*generator.IoTDevice
}

func NewProducer(mqClient *mq.Client) *Producer {
	var iotDevices []*generator.IoTDevice
	for range rand.Intn(5) {
		iotDevices = append(iotDevices, generator.NewIoTDevice())
	}
	return &Producer{
		MQClient:   mqClient,
		IoTDevices: iotDevices,
	}
}

func (p *Producer) RandomDataPoint() error {
	deviceId := p.IoTDevices[rand.Intn(len(p.IoTDevices))].DeviceID
	iotDataGen := generator.NewIoTGenerator(deviceId)
	reading := iotDataGen.GenerateCorrelatedReading(time.Now())
	return p.MQClient.Push(reading)
}
