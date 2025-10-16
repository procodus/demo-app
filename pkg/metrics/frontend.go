package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// FrontendMetrics contains Prometheus metrics for the frontend service.
type FrontendMetrics struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec
	HTTPResponseSize     *prometheus.HistogramVec
	GRPCClientCalls      *prometheus.CounterVec
	GRPCClientDuration   *prometheus.HistogramVec
	GRPCClientErrors     *prometheus.CounterVec
	TemplateRenderTime   *prometheus.HistogramVec
	TemplateRenderErrors *prometheus.CounterVec
}

// NewFrontendMetrics creates and registers frontend service metrics.
func NewFrontendMetrics(namespace string) *FrontendMetrics {
	m := &FrontendMetrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Duration of HTTP requests",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "Size of HTTP responses in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to ~10MB
			},
			[]string{"path"},
		),
		GRPCClientCalls: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "grpc_client",
				Name:      "calls_total",
				Help:      "Total number of gRPC client calls",
			},
			[]string{"method", "status"}, // status: success, error
		),
		GRPCClientDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "grpc_client",
				Name:      "call_duration_seconds",
				Help:      "Duration of gRPC client calls",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		GRPCClientErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "grpc_client",
				Name:      "errors_total",
				Help:      "Total number of gRPC client errors",
			},
			[]string{"method", "error_type"},
		),
		TemplateRenderTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "template",
				Name:      "render_duration_seconds",
				Help:      "Duration of template rendering",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"template"},
		),
		TemplateRenderErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "template",
				Name:      "render_errors_total",
				Help:      "Total number of template rendering errors",
			},
			[]string{"template", "error_type"},
		),
	}

	MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.HTTPResponseSize,
		m.GRPCClientCalls,
		m.GRPCClientDuration,
		m.GRPCClientErrors,
		m.TemplateRenderTime,
		m.TemplateRenderErrors,
	)

	return m
}
