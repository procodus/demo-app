package frontend

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"procodus.dev/demo-app/pkg/iot"
	"procodus.dev/demo-app/pkg/metrics"
)

// renderIndex renders the index page.
func renderIndex(ctx context.Context, w http.ResponseWriter, m *metrics.FrontendMetrics) error {
	return trackTemplateRender(ctx, w, m, "index", func() error {
		return index().Render(ctx, w)
	})
}

// renderDevices renders the devices page.
func renderDevices(ctx context.Context, w http.ResponseWriter, deviceList []*iot.IoTDevice, m *metrics.FrontendMetrics) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return trackTemplateRender(ctx, w, m, "devices", func() error {
		return devices(deviceList).Render(ctx, w)
	})
}

// renderDevice renders a single device detail page.
func renderDevice(ctx context.Context, w http.ResponseWriter, dev *iot.IoTDevice, readings []*iot.SensorReading, m *metrics.FrontendMetrics) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return trackTemplateRender(ctx, w, m, "device", func() error {
		return device(dev, readings).Render(ctx, w)
	})
}

// renderDevicesList renders the devices list fragment.
func renderDevicesList(ctx context.Context, w http.ResponseWriter, deviceList []*iot.IoTDevice, m *metrics.FrontendMetrics) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return trackTemplateRender(ctx, w, m, "devices_list", func() error {
		return devicesList(deviceList).Render(ctx, w)
	})
}

// renderReadingsList renders the readings list fragment.
func renderReadingsList(ctx context.Context, w http.ResponseWriter, readings []*iot.SensorReading, nextPageToken string, m *metrics.FrontendMetrics) error {
	//nolint:contextcheck // Context is passed to Templ's Render method
	return trackTemplateRender(ctx, w, m, "readings_list", func() error {
		return readingsList(readings, nextPageToken).Render(ctx, w)
	})
}

// trackTemplateRender wraps template rendering with metrics tracking.
func trackTemplateRender(ctx context.Context, w http.ResponseWriter, m *metrics.FrontendMetrics, templateName string, renderFunc func() error) error {
	// If metrics not enabled, just render
	if m == nil {
		return renderFunc()
	}

	// Track duration
	timer := prometheus.NewTimer(m.TemplateRenderTime.WithLabelValues(templateName))
	defer timer.ObserveDuration()

	// Render template
	err := renderFunc()

	// Track errors
	if err != nil {
		m.TemplateRenderErrors.WithLabelValues(templateName, "render_error").Inc()
		return err
	}

	return nil
}
