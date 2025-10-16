package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// ProducerMetrics contains Prometheus metrics for the producer/generator service.
type ProducerMetrics struct {
	MessagesGenerated     *prometheus.CounterVec
	GenerationFailures    *prometheus.CounterVec
	GenerationDuration    *prometheus.HistogramVec
	ActiveProducers       prometheus.Gauge
	DevicesGenerated      prometheus.Counter
	SensorReadingsCreated prometheus.Counter
}

// NewProducerMetrics creates and registers producer metrics.
func NewProducerMetrics(namespace string) *ProducerMetrics {
	m := &ProducerMetrics{
		MessagesGenerated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "producer",
				Name:      "messages_generated_total",
				Help:      "Total number of messages generated",
			},
			[]string{"type"}, // type: device, sensor_reading
		),
		GenerationFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "producer",
				Name:      "generation_failures_total",
				Help:      "Total number of message generation failures",
			},
			[]string{"type", "reason"},
		),
		GenerationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "producer",
				Name:      "generation_duration_seconds",
				Help:      "Duration of message generation operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"type"},
		),
		ActiveProducers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "producer",
				Name:      "active_producers",
				Help:      "Number of currently active producers",
			},
		),
		DevicesGenerated: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "producer",
				Name:      "devices_generated_total",
				Help:      "Total number of devices generated",
			},
		),
		SensorReadingsCreated: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "producer",
				Name:      "sensor_readings_created_total",
				Help:      "Total number of sensor readings created",
			},
		),
	}

	MustRegister(
		m.MessagesGenerated,
		m.GenerationFailures,
		m.GenerationDuration,
		m.ActiveProducers,
		m.DevicesGenerated,
		m.SensorReadingsCreated,
	)

	return m
}
