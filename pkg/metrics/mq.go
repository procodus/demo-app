package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// MQMetrics contains Prometheus metrics for the MQ client.
type MQMetrics struct {
	MessagesPushed      *prometheus.CounterVec
	PushFailures        *prometheus.CounterVec
	ReconnectAttempts   prometheus.Counter
	PushDuration        *prometheus.HistogramVec
	ConnectionStatus    prometheus.Gauge
	MessagesConsumed    *prometheus.CounterVec
	ConsumptionFailures *prometheus.CounterVec
	ConsumeDuration     *prometheus.HistogramVec
}

// NewMQMetrics creates and registers MQ client metrics.
func NewMQMetrics(namespace string) *MQMetrics {
	m := &MQMetrics{
		MessagesPushed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "messages_pushed_total",
				Help:      "Total number of messages pushed to RabbitMQ",
			},
			[]string{"queue"},
		),
		PushFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "push_failures_total",
				Help:      "Total number of failed message pushes",
			},
			[]string{"queue", "reason"},
		),
		ReconnectAttempts: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "reconnect_attempts_total",
				Help:      "Total number of reconnection attempts",
			},
		),
		PushDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "push_duration_seconds",
				Help:      "Duration of message push operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"queue"},
		),
		ConnectionStatus: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "connection_status",
				Help:      "Current connection status (1=connected, 0=disconnected)",
			},
		),
		MessagesConsumed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "messages_consumed_total",
				Help:      "Total number of messages consumed from RabbitMQ",
			},
			[]string{"queue"},
		),
		ConsumptionFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "consumption_failures_total",
				Help:      "Total number of failed message consumptions",
			},
			[]string{"queue", "reason"},
		),
		ConsumeDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "mq",
				Name:      "consume_duration_seconds",
				Help:      "Duration of message consumption operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"queue"},
		),
	}

	MustRegister(
		m.MessagesPushed,
		m.PushFailures,
		m.ReconnectAttempts,
		m.PushDuration,
		m.ConnectionStatus,
		m.MessagesConsumed,
		m.ConsumptionFailures,
		m.ConsumeDuration,
	)

	return m
}
