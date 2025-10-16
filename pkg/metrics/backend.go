package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// BackendMetrics contains Prometheus metrics for the backend service.
type BackendMetrics struct {
	GRPCRequestsTotal     *prometheus.CounterVec
	GRPCRequestDuration   *prometheus.HistogramVec
	GRPCRequestsInFlight  *prometheus.GaugeVec
	ConsumerMessagesTotal *prometheus.CounterVec
	ConsumerErrors        *prometheus.CounterVec
	ProcessingDuration    *prometheus.HistogramVec
	DBOperationsTotal     *prometheus.CounterVec
	DBOperationDuration   *prometheus.HistogramVec
	DBConnectionsActive   prometheus.Gauge
	ActiveConsumers       prometheus.Gauge
}

// NewBackendMetrics creates and registers backend service metrics.
func NewBackendMetrics(namespace string) *BackendMetrics {
	m := &BackendMetrics{
		GRPCRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "grpc",
				Name:      "requests_total",
				Help:      "Total number of gRPC requests",
			},
			[]string{"method", "status"}, // status: success, error
		),
		GRPCRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "grpc",
				Name:      "request_duration_seconds",
				Help:      "Duration of gRPC requests",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		GRPCRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "grpc",
				Name:      "requests_in_flight",
				Help:      "Number of gRPC requests currently being processed",
			},
			[]string{"method"},
		),
		ConsumerMessagesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "consumer",
				Name:      "messages_total",
				Help:      "Total number of messages consumed",
			},
			[]string{"queue", "status"}, // status: success, error
		),
		ConsumerErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "consumer",
				Name:      "errors_total",
				Help:      "Total number of consumer errors",
			},
			[]string{"queue", "error_type"},
		),
		ProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "consumer",
				Name:      "processing_duration_seconds",
				Help:      "Duration of message processing",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"queue"},
		),
		DBOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "operations_total",
				Help:      "Total number of database operations",
			},
			[]string{"operation", "table", "status"}, // operation: insert, update, select, delete
		),
		DBOperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "operation_duration_seconds",
				Help:      "Duration of database operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		DBConnectionsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "db",
				Name:      "connections_active",
				Help:      "Number of active database connections",
			},
		),
		ActiveConsumers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "consumer",
				Name:      "active_consumers",
				Help:      "Number of active message consumers",
			},
		),
	}

	MustRegister(
		m.GRPCRequestsTotal,
		m.GRPCRequestDuration,
		m.GRPCRequestsInFlight,
		m.ConsumerMessagesTotal,
		m.ConsumerErrors,
		m.ProcessingDuration,
		m.DBOperationsTotal,
		m.DBOperationDuration,
		m.DBConnectionsActive,
		m.ActiveConsumers,
	)

	return m
}
