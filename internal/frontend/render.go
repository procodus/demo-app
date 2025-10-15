package frontend

import (
	"context"
	"net/http"

	"procodus.dev/demo-app/pkg/iot"
)

// renderIndex renders the index page.
func renderIndex(w http.ResponseWriter, ctx context.Context) error {
	return index().Render(ctx, w)
}

// renderDevices renders the devices page.
func renderDevices(w http.ResponseWriter, ctx context.Context, deviceList []*iot.IoTDevice) error {
	return devices(deviceList).Render(ctx, w)
}

// renderDevice renders a single device detail page.
func renderDevice(w http.ResponseWriter, ctx context.Context, dev *iot.IoTDevice, readings []*iot.SensorReading) error {
	return device(dev, readings).Render(ctx, w)
}

// renderDevicesList renders the devices list fragment.
func renderDevicesList(w http.ResponseWriter, ctx context.Context, deviceList []*iot.IoTDevice) error {
	return devicesList(deviceList).Render(ctx, w)
}

// renderReadingsList renders the readings list fragment.
func renderReadingsList(w http.ResponseWriter, ctx context.Context, readings []*iot.SensorReading, nextPageToken string) error {
	return readingsList(readings, nextPageToken).Render(ctx, w)
}
