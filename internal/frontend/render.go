package frontend

import (
	"context"
	"net/http"

	"procodus.dev/demo-app/pkg/iot"
)

// renderIndex renders the index page.
func renderIndex(ctx context.Context, w http.ResponseWriter) error {
	return index().Render(ctx, w)
}

// renderDevices renders the devices page.
func renderDevices(ctx context.Context, w http.ResponseWriter, deviceList []*iot.IoTDevice) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return devices(deviceList).Render(ctx, w)
}

// renderDevice renders a single device detail page.
func renderDevice(ctx context.Context, w http.ResponseWriter, dev *iot.IoTDevice, readings []*iot.SensorReading) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return device(dev, readings).Render(ctx, w)
}

// renderDevicesList renders the devices list fragment.
func renderDevicesList(ctx context.Context, w http.ResponseWriter, deviceList []*iot.IoTDevice) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return devicesList(deviceList).Render(ctx, w)
}

// renderReadingsList renders the readings list fragment.
func renderReadingsList(ctx context.Context, w http.ResponseWriter, readings []*iot.SensorReading, nextPageToken string) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return readingsList(readings, nextPageToken).Render(ctx, w)
}
