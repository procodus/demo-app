# Monitoring Guide

This guide covers monitoring the Demo App IoT Data Pipeline with Prometheus and creating Grafana dashboards.

## Table of Contents

- [Overview](#overview)
- [Prometheus Setup](#prometheus-setup)
- [Metrics Reference](#metrics-reference)
- [Grafana Dashboards](#grafana-dashboards)
- [Alerting](#alerting)
- [Troubleshooting](#troubleshooting)

## Overview

The application exposes comprehensive Prometheus metrics for all services:

| Service | Endpoint | Port | Description |
|---------|----------|------|-------------|
| **Generator** | `/metrics` | 9091 | Message generation metrics |
| **Backend** | `/metrics` | 9090 | Consumer and gRPC metrics |
| **Frontend** | `/metrics` | 8080 | HTTP and template metrics |

## Prometheus Setup

### Local Setup

**1. Install Prometheus**:
```bash
# macOS
brew install prometheus

# Linux
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar -xzf prometheus-2.45.0.linux-amd64.tar.gz
cd prometheus-2.45.0.linux-amd64
```

**2. Configure Prometheus** (`prometheus.yml`):
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'demo-generator'
    static_configs:
      - targets: ['localhost:9091']
        labels:
          service: 'generator'

  - job_name: 'demo-backend'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          service: 'backend'

  - job_name: 'demo-frontend'
    static_configs:
      - targets: ['localhost:8080']
        labels:
          service: 'frontend'
```

**3. Start Prometheus**:
```bash
prometheus --config.file=prometheus.yml
```

Access at: http://localhost:9000

### Kubernetes Setup

**Using Prometheus Operator**:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: demo-app-metrics
  namespace: demo-app
spec:
  selector:
    matchLabels:
      monitoring: enabled
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
```

**Using Helm**:
```bash
# Install Prometheus Stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace

# Configure scraping for demo-app
kubectl apply -f servicemonitor.yaml
```

## Metrics Reference

### Generator Metrics (6 metrics)

**Message Generation**:
```promql
# Total messages generated (device + sensor)
demo_app_producer_messages_generated_total{type="device"}
demo_app_producer_messages_generated_total{type="sensor_reading"}

# Generation rate (messages/second)
rate(demo_app_producer_messages_generated_total[5m])

# Failures
demo_app_producer_generation_failures_total{type="device",reason="publish_failed"}

# Generation duration (seconds)
demo_app_producer_generation_duration_seconds_bucket
demo_app_producer_generation_duration_seconds_sum
demo_app_producer_generation_duration_seconds_count
```

**Producer Status**:
```promql
# Active producers
demo_app_producer_active_producers

# Devices generated
demo_app_producer_devices_generated_total

# Sensor readings created
demo_app_producer_sensor_readings_created_total
```

### Backend Metrics (10 metrics)

**Consumer Metrics**:
```promql
# Messages processed
demo_app_backend_consumer_messages_total{queue="sensor-data",status="success"}
demo_app_backend_consumer_messages_total{queue="device-data",status="success"}

# Consumer errors
demo_app_backend_consumer_errors_total{queue="sensor-data",error_type="database_error"}

# Processing duration (seconds)
demo_app_backend_processing_duration_seconds_bucket{queue="sensor-data"}
demo_app_backend_processing_duration_seconds_sum{queue="sensor-data"}
demo_app_backend_processing_duration_seconds_count{queue="sensor-data"}

# Active consumers
demo_app_backend_active_consumers
```

**gRPC API Metrics**:
```promql
# Request count
demo_app_backend_grpc_requests_total{method="/iot.SensorService/GetAllDevice",status="success"}

# Request duration (seconds)
demo_app_backend_grpc_request_duration_seconds_bucket{method="/iot.SensorService/GetDevice"}
demo_app_backend_grpc_request_duration_seconds_sum{method="/iot.SensorService/GetDevice"}
demo_app_backend_grpc_request_duration_seconds_count{method="/iot.SensorService/GetDevice"}

# In-flight requests
demo_app_backend_grpc_requests_in_flight{method="/iot.SensorService/GetAllDevice"}
```

### Frontend Metrics (9 metrics)

**HTTP Server Metrics**:
```promql
# HTTP requests
demo_app_frontend_http_requests_total{method="GET",path="/devices",status_code="200"}

# Request duration (seconds)
demo_app_frontend_http_request_duration_seconds_bucket{method="GET",path="/"}

# In-flight requests
demo_app_frontend_http_requests_in_flight{method="GET",path="/devices"}

# Response size (bytes)
demo_app_frontend_http_response_size_bytes_bucket{path="/devices"}
```

**gRPC Client Metrics**:
```promql
# gRPC calls to backend
demo_app_frontend_grpc_client_calls_total{method="GetAllDevice",status="success"}

# Call duration (seconds)
demo_app_frontend_grpc_client_call_duration_seconds_bucket{method="GetDevice"}

# Client errors
demo_app_frontend_grpc_client_errors_total{method="GetDevice",error_type="deadline_exceeded"}
```

**Template Rendering**:
```promql
# Template render duration
demo_app_frontend_template_render_duration_seconds_bucket{template="devices"}

# Template errors
demo_app_frontend_template_render_errors_total{template="device",error_type="render_error"}
```

### MQ Client Metrics (8 metrics)

**Connection Status**:
```promql
# Connection status (1=connected, 0=disconnected)
demo_app_mq_connection_status

# Reconnection attempts
demo_app_mq_reconnect_attempts_total
```

**Push Operations**:
```promql
# Messages pushed
demo_app_mq_messages_pushed_total{queue="sensor-data"}

# Push failures
demo_app_mq_push_failures_total{queue="sensor-data",reason="max_retries_exceeded"}

# Push duration (seconds)
demo_app_mq_push_duration_seconds_bucket{queue="sensor-data"}
```

**Consume Operations**:
```promql
# Messages consumed
demo_app_mq_messages_consumed_total{queue="sensor-data"}

# Consumption failures
demo_app_mq_consumption_failures_total{queue="sensor-data",reason="decode_error"}

# Consume duration (seconds)
demo_app_mq_consume_duration_seconds_bucket{queue="sensor-data"}
```

## Grafana Dashboards

### Install Grafana

```bash
# macOS
brew install grafana
brew services start grafana

# Docker
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana-oss

# Access at http://localhost:3000 (admin/admin)
```

### Add Prometheus Data Source

1. Navigate to **Configuration** → **Data Sources**
2. Click **Add data source**
3. Select **Prometheus**
4. Set URL: `http://localhost:9000`
5. Click **Save & Test**

### Dashboard 1: System Overview

**Panels**:

1. **Message Flow Rate** (Time Series):
```promql
rate(demo_app_producer_messages_generated_total[5m])
```

2. **Consumer Processing Rate** (Time Series):
```promql
rate(demo_app_backend_consumer_messages_total{status="success"}[5m])
```

3. **gRPC Request Rate** (Time Series):
```promql
rate(demo_app_backend_grpc_requests_total[5m])
```

4. **HTTP Request Rate** (Time Series):
```promql
rate(demo_app_frontend_http_requests_total[5m])
```

5. **Active Components** (Stat):
```promql
demo_app_producer_active_producers
demo_app_backend_active_consumers
```

6. **Connection Status** (Stat):
```promql
demo_app_mq_connection_status
```

### Dashboard 2: Performance Metrics

**Panels**:

1. **gRPC Request Duration (p95)** (Time Series):
```promql
histogram_quantile(0.95,
  rate(demo_app_backend_grpc_request_duration_seconds_bucket[5m])
)
```

2. **Consumer Processing Duration (p95)** (Time Series):
```promql
histogram_quantile(0.95,
  rate(demo_app_backend_processing_duration_seconds_bucket[5m])
)
```

3. **HTTP Request Duration (p95)** (Time Series):
```promql
histogram_quantile(0.95,
  rate(demo_app_frontend_http_request_duration_seconds_bucket[5m])
)
```

4. **Template Render Time (p99)** (Time Series):
```promql
histogram_quantile(0.99,
  rate(demo_app_frontend_template_render_duration_seconds_bucket[5m])
)
```

### Dashboard 3: Error Monitoring

**Panels**:

1. **Generation Failures** (Time Series):
```promql
rate(demo_app_producer_generation_failures_total[5m])
```

2. **Consumer Errors** (Time Series):
```promql
rate(demo_app_backend_consumer_errors_total[5m])
```

3. **gRPC Errors** (Time Series):
```promql
rate(demo_app_backend_grpc_requests_total{status="error"}[5m])
```

4. **HTTP 5xx Errors** (Time Series):
```promql
rate(demo_app_frontend_http_requests_total{status_code=~"5.."}[5m])
```

5. **MQ Push Failures** (Time Series):
```promql
rate(demo_app_mq_push_failures_total[5m])
```

### Dashboard 4: Resource Utilization

**Panels**:

1. **Message Queue Depth** (requires RabbitMQ exporter):
```promql
rabbitmq_queue_messages{queue="sensor-data"}
```

2. **Database Connections** (requires PostgreSQL exporter):
```promql
pg_stat_activity_count
```

3. **Go Memory Usage** (Gauge):
```promql
go_memstats_alloc_bytes
```

4. **Go Goroutines** (Gauge):
```promql
go_goroutines
```

## Alerting

### Prometheus Alert Rules

Create `alerts.yml`:

```yaml
groups:
  - name: demo-app-alerts
    interval: 30s
    rules:
      # Connection lost
      - alert: MQConnectionDown
        expr: demo_app_mq_connection_status == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "RabbitMQ connection is down"
          description: "Service {{ $labels.service }} lost connection to RabbitMQ"

      # High error rate
      - alert: HighErrorRate
        expr: |
          rate(demo_app_backend_consumer_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High consumer error rate"
          description: "Consumer error rate is {{ $value }}/sec"

      # Consumer not running
      - alert: NoActiveConsumers
        expr: demo_app_backend_active_consumers < 1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "No active consumers running"
          description: "Backend has 0 active consumers"

      # High gRPC latency
      - alert: HighGRPCLatency
        expr: |
          histogram_quantile(0.95,
            rate(demo_app_backend_grpc_request_duration_seconds_bucket[5m])
          ) > 1.0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High gRPC API latency"
          description: "p95 latency is {{ $value }}s"

      # High message queue depth
      - alert: HighQueueDepth
        expr: rabbitmq_queue_messages{queue="sensor-data"} > 10000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High message queue depth"
          description: "Queue has {{ $value }} messages pending"
```

Load alerts:
```bash
prometheus --config.file=prometheus.yml --storage.tsdb.path=data/ --web.enable-lifecycle
```

### Alertmanager Configuration

Create `alertmanager.yml`:

```yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'
    - match:
        severity: warning
      receiver: 'slack'

receivers:
  - name: 'default'
    email_configs:
      - to: 'ops@example.com'

  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'
        title: 'Demo App Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_KEY'
```

## Key Metrics to Monitor

### Golden Signals

**1. Latency**:
```promql
# p95 request duration
histogram_quantile(0.95,
  rate(demo_app_backend_grpc_request_duration_seconds_bucket[5m])
)
```

**2. Traffic**:
```promql
# Requests per second
rate(demo_app_backend_grpc_requests_total[5m])
```

**3. Errors**:
```promql
# Error rate
rate(demo_app_backend_grpc_requests_total{status="error"}[5m])
  /
rate(demo_app_backend_grpc_requests_total[5m])
```

**4. Saturation**:
```promql
# In-flight requests
demo_app_backend_grpc_requests_in_flight

# Consumer lag (queue depth)
rabbitmq_queue_messages{queue="sensor-data"}
```

### SLIs (Service Level Indicators)

**Availability**:
```promql
# Uptime
up{job="demo-backend"}

# Connection status
demo_app_mq_connection_status
```

**Success Rate**:
```promql
# Percentage of successful requests
sum(rate(demo_app_backend_grpc_requests_total{status="success"}[5m]))
  /
sum(rate(demo_app_backend_grpc_requests_total[5m]))
* 100
```

**Response Time**:
```promql
# p99 latency
histogram_quantile(0.99,
  rate(demo_app_backend_grpc_request_duration_seconds_bucket[5m])
)
```

## Troubleshooting

### No Metrics Available

**Check metrics endpoint**:
```bash
curl http://localhost:9090/metrics
curl http://localhost:9091/metrics
curl http://localhost:8080/metrics
```

**Verify Prometheus scraping**:
- Open Prometheus UI: http://localhost:9000
- Go to **Status** → **Targets**
- Verify targets are `UP`

### High Memory Usage

**Check memory metrics**:
```promql
go_memstats_alloc_bytes
go_memstats_heap_inuse_bytes
```

**Check goroutines**:
```promql
go_goroutines
```

### Missing Labels

Ensure labels are consistent in queries:
```promql
# Correct
demo_app_backend_grpc_requests_total{method="/iot.SensorService/GetDevice"}

# Incorrect (typo)
demo_app_backend_grpc_requests_total{method="GetDevice"}
```

## Next Steps

- [Troubleshooting Guide](./troubleshooting.md) - Debug issues
- [Performance Guide](./performance.md) - Optimize performance
- [API Reference](./api.md) - Understand metrics context
