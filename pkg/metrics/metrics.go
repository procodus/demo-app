// Package metrics provides Prometheus metrics collection for all services.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry is the global Prometheus registry for all metrics.
var Registry = prometheus.NewRegistry()

func init() {
	// Register default Go metrics collectors
	Registry.MustRegister(collectors.NewGoCollector())
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// Handler returns an HTTP handler for exposing Prometheus metrics.
func Handler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// MustRegister registers collectors with the global registry.
// Panics if registration fails.
func MustRegister(collectors ...prometheus.Collector) {
	Registry.MustRegister(collectors...)
}
