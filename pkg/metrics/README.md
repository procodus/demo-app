# Metrics Package

This package provides Prometheus metrics collection for all services in the demo-app project.

## Overview

The metrics package offers:
- Pre-configured Prometheus metrics for each service type
- Standard Go runtime metrics (memory, goroutines, GC, etc.)
- Process metrics (CPU, file descriptors, etc.)
- HTTP handler for exposing metrics in Prometheus format

## Usage

### Basic Setup

```go
import (
    "net/http"
    "procodus.dev/demo-app/pkg/metrics"
)

func main() {
    // Expose metrics endpoint
    http.Handle("/metrics", metrics.Handler())
    go http.ListenAndServe(":2112", nil)

    // ... rest of your application
}
```

### Producer/Generator Metrics

```go
import "procodus.dev/demo-app/pkg/metrics"

// Initialize metrics
m := metrics.NewProducerMetrics("demo_app")

// Record metrics
m.MessagesGenerated.WithLabelValues("device").Inc()
m.SensorReadingsCreated.Inc()
m.ActiveProducers.Set(5)

// Record duration
timer := prometheus.NewTimer(m.GenerationDuration.WithLabelValues("sensor_reading"))
defer timer.ObserveDuration()
// ... do work ...
```

### Backend Service Metrics

```go
import "procodus.dev/demo-app/pkg/metrics"

// Initialize metrics
m := metrics.NewBackendMetrics("demo_app")

// Record gRPC request
m.GRPCRequestsTotal.WithLabelValues("GetAllDevice", "success").Inc()

// Record database operation
m.DBOperationsTotal.WithLabelValues("insert", "sensor_readings", "success").Inc()

// Track in-flight requests
m.GRPCRequestsInFlight.WithLabelValues("GetDevice").Inc()
defer m.GRPCRequestsInFlight.WithLabelValues("GetDevice").Dec()

// Record duration
timer := prometheus.NewTimer(m.GRPCRequestDuration.WithLabelValues("GetAllDevice"))
defer timer.ObserveDuration()
// ... handle request ...
```

### Frontend Service Metrics

```go
import "procodus.dev/demo-app/pkg/metrics"

// Initialize metrics
m := metrics.NewFrontendMetrics("demo_app")

// Record HTTP request
m.HTTPRequestsTotal.WithLabelValues("GET", "/devices", "200").Inc()

// Record gRPC client call
m.GRPCClientCalls.WithLabelValues("GetAllDevice", "success").Inc()

// Record template rendering
timer := prometheus.NewTimer(m.TemplateRenderTime.WithLabelValues("device_list"))
defer timer.ObserveDuration()
// ... render template ...
```

### MQ Client Metrics

```go
import "procodus.dev/demo-app/pkg/metrics"

// Initialize metrics
m := metrics.NewMQMetrics("demo_app")

// Record message push
m.MessagesPushed.WithLabelValues("sensor-data").Inc()

// Record failure
m.PushFailures.WithLabelValues("sensor-data", "max_retries_exceeded").Inc()

// Update connection status
m.ConnectionStatus.Set(1) // 1 = connected, 0 = disconnected

// Record duration
timer := prometheus.NewTimer(m.PushDuration.WithLabelValues("sensor-data"))
defer timer.ObserveDuration()
// ... push message ...
```

## Metrics Endpoint

Each service should expose a `/metrics` endpoint on a separate port (typically 2112):

```bash
# View metrics
curl http://localhost:2112/metrics

# Check specific metric
curl http://localhost:2112/metrics | grep demo_app_mq_messages_pushed_total
```

## Available Metrics

### Producer Metrics (`demo_app_producer_*`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `messages_generated_total` | Counter | `type` | Total messages generated |
| `generation_failures_total` | Counter | `type`, `reason` | Failed generations |
| `generation_duration_seconds` | Histogram | `type` | Generation duration |
| `active_producers` | Gauge | - | Active producers |
| `devices_generated_total` | Counter | - | Total devices generated |
| `sensor_readings_created_total` | Counter | - | Total sensor readings |

### Backend Metrics (`demo_app_*`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `grpc_requests_total` | Counter | `method`, `status` | Total gRPC requests |
| `grpc_request_duration_seconds` | Histogram | `method` | gRPC request duration |
| `grpc_requests_in_flight` | Gauge | `method` | In-flight gRPC requests |
| `consumer_messages_total` | Counter | `queue`, `status` | Messages consumed |
| `consumer_errors_total` | Counter | `queue`, `error_type` | Consumer errors |
| `consumer_processing_duration_seconds` | Histogram | `queue` | Processing duration |
| `db_operations_total` | Counter | `operation`, `table`, `status` | DB operations |
| `db_operation_duration_seconds` | Histogram | `operation`, `table` | DB operation duration |
| `db_connections_active` | Gauge | - | Active DB connections |
| `consumer_active_consumers` | Gauge | - | Active consumers |

### Frontend Metrics (`demo_app_*`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | `method`, `path`, `status_code` | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | `method`, `path` | HTTP request duration |
| `http_requests_in_flight` | Gauge | `method`, `path` | In-flight HTTP requests |
| `http_response_size_bytes` | Histogram | `path` | Response size |
| `grpc_client_calls_total` | Counter | `method`, `status` | gRPC client calls |
| `grpc_client_call_duration_seconds` | Histogram | `method` | gRPC call duration |
| `grpc_client_errors_total` | Counter | `method`, `error_type` | gRPC client errors |
| `template_render_duration_seconds` | Histogram | `template` | Template render time |
| `template_render_errors_total` | Counter | `template`, `error_type` | Template errors |

### MQ Metrics (`demo_app_mq_*`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `messages_pushed_total` | Counter | `queue` | Messages pushed |
| `push_failures_total` | Counter | `queue`, `reason` | Push failures |
| `reconnect_attempts_total` | Counter | - | Reconnection attempts |
| `push_duration_seconds` | Histogram | `queue` | Push duration |
| `connection_status` | Gauge | - | Connection status (1/0) |
| `messages_consumed_total` | Counter | `queue` | Messages consumed |
| `consumption_failures_total` | Counter | `queue`, `reason` | Consumption failures |
| `consume_duration_seconds` | Histogram | `queue` | Consume duration |

## Integration with Prometheus

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'demo-app-generator'
    static_configs:
      - targets: ['localhost:2112']

  - job_name: 'demo-app-backend'
    static_configs:
      - targets: ['localhost:2113']

  - job_name: 'demo-app-frontend'
    static_configs:
      - targets: ['localhost:2114']
```

## Best Practices

1. **Use consistent namespaces**: All metrics use `demo_app` as the namespace
2. **Label cardinality**: Keep label values bounded (avoid high-cardinality labels like timestamps or UUIDs)
3. **Separate metrics port**: Expose metrics on a different port than the main service
4. **Duration tracking**: Use `prometheus.NewTimer()` for accurate duration tracking
5. **Error categorization**: Use specific error types in labels for better debugging

## Testing

Metrics can be tested by checking the registry:

```go
import (
    "testing"
    "github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetrics(t *testing.T) {
    m := metrics.NewProducerMetrics("test")
    m.DevicesGenerated.Inc()

    // Verify counter value
    value := testutil.ToFloat64(m.DevicesGenerated)
    if value != 1 {
        t.Errorf("expected 1, got %v", value)
    }
}
```

## Related Documentation

- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Go Client Library](https://github.com/prometheus/client_golang)
- [CLAUDE.md](../../CLAUDE.md) - Project overview
