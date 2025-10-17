# Configuration Guide

This document describes all configuration options for the Demo App IoT Data Pipeline.

## Table of Contents

- [Configuration Methods](#configuration-methods)
- [Generator Configuration](#generator-configuration)
- [Backend Configuration](#backend-configuration)
- [Frontend Configuration](#frontend-configuration)
- [Global Settings](#global-settings)
- [Environment Variables](#environment-variables)
- [Configuration Examples](#configuration-examples)

## Configuration Methods

The application supports multiple configuration methods with the following priority (highest to lowest):

1. **Command-line flags** - Explicit values passed as flags
2. **Environment variables** - OS environment variables with `APP_` prefix
3. **Configuration file** - YAML file specified with `--config` flag
4. **Default values** - Built-in defaults

### Configuration File

Create a `config.yaml` file:

```yaml
generator:
  rabbitmq_url: amqp://localhost:5672
  device_queue: device-data
  sensor_queue: sensor-data
  interval: 5s
  num_devices: 10

backend:
  grpc_port: 50051
  metrics_port: 9090
  db:
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    database: iot_db
    sslmode: disable
  rabbitmq:
    url: amqp://localhost:5672
    sensor_queue: sensor-data
    device_queue: device-data

frontend:
  http_port: 8080
  backend_url: localhost:50051

log:
  level: info
  format: json
```

Use the config file:
```bash
./demo-app generator --config=config.yaml
```

### Command-Line Flags

Override specific settings:
```bash
./demo-app backend \
  --grpc-port=50052 \
  --db-host=postgres.example.com \
  --log-level=debug
```

### Environment Variables

Set environment variables with `APP_` prefix:
```bash
export APP_BACKEND_GRPC_PORT=50051
export APP_BACKEND_DB_HOST=localhost
export APP_BACKEND_DB_PASSWORD=secret
export APP_LOG_LEVEL=info

./demo-app backend
```

**Naming Convention**:
- Replace `.` with `_`
- Convert to UPPERCASE
- Add `APP_` prefix

Examples:
- `backend.grpc_port` → `APP_BACKEND_GRPC_PORT`
- `backend.db.host` → `APP_BACKEND_DB_HOST`
- `log.level` → `APP_LOG_LEVEL`

## Generator Configuration

The generator service creates synthetic IoT device data and publishes to RabbitMQ.

### Generator Flags

| Flag | Environment Variable | Type | Default | Description |
|------|---------------------|------|---------|-------------|
| `--rabbitmq-url` | `APP_GENERATOR_RABBITMQ_URL` | string | `amqp://localhost:5672` | RabbitMQ connection URL |
| `--device-queue` | `APP_GENERATOR_DEVICE_QUEUE` | string | `device-data` | Queue name for device messages |
| `--sensor-queue` | `APP_GENERATOR_SENSOR_QUEUE` | string | `sensor-data` | Queue name for sensor readings |
| `--interval` | `APP_GENERATOR_INTERVAL` | duration | `5s` | Interval between sensor readings |
| `--num-devices` | `APP_GENERATOR_NUM_DEVICES` | int | `10` | Number of devices to simulate |
| `--metrics-port` | `APP_GENERATOR_METRICS_PORT` | int | `9091` | Prometheus metrics HTTP port |
| `--enable-metrics` | `APP_GENERATOR_ENABLE_METRICS` | bool | `true` | Enable Prometheus metrics |

### Generator Example

**Config File**:
```yaml
generator:
  rabbitmq_url: amqp://user:pass@rabbitmq.example.com:5672
  device_queue: iot-devices
  sensor_queue: iot-sensors
  interval: 10s
  num_devices: 50
  metrics_port: 9091
  enable_metrics: true
```

**Command Line**:
```bash
./demo-app generator \
  --rabbitmq-url=amqp://rabbitmq.example.com:5672 \
  --device-queue=iot-devices \
  --sensor-queue=iot-sensors \
  --interval=10s \
  --num-devices=50 \
  --metrics-port=9091 \
  --enable-metrics=true
```

**Environment Variables**:
```bash
export APP_GENERATOR_RABBITMQ_URL=amqp://rabbitmq.example.com:5672
export APP_GENERATOR_DEVICE_QUEUE=iot-devices
export APP_GENERATOR_SENSOR_QUEUE=iot-sensors
export APP_GENERATOR_INTERVAL=10s
export APP_GENERATOR_NUM_DEVICES=50
export APP_GENERATOR_METRICS_PORT=9091
export APP_GENERATOR_ENABLE_METRICS=true

./demo-app generator
```

### Generator Behavior

**Device Generation**:
- Generates `num_devices` unique devices on startup
- Each device has:
  - Unique ID (`device-001`, `device-002`, etc.)
  - Random location (city name)
  - Random MAC address
  - Random IP address
  - Random firmware version
  - Random latitude/longitude

**Sensor Reading Generation**:
- Generates readings every `interval` (e.g., 5s)
- Each reading includes:
  - Temperature: -40°C to 85°C
  - Humidity: 0% to 100%
  - Pressure: 300 hPa to 1100 hPa
  - Battery Level: 0% to 100% (decreases over time)

## Backend Configuration

The backend service consumes messages, persists data, and provides gRPC API.

### Backend Flags

| Flag | Environment Variable | Type | Default | Description |
|------|---------------------|------|---------|-------------|
| **gRPC Server** |
| `--grpc-port` | `APP_BACKEND_GRPC_PORT` | int | `50051` | gRPC server port |
| `--metrics-port` | `APP_BACKEND_METRICS_PORT` | int | `9090` | Prometheus metrics HTTP port |
| `--enable-metrics` | `APP_BACKEND_ENABLE_METRICS` | bool | `true` | Enable Prometheus metrics |
| **Database** |
| `--db-host` | `APP_BACKEND_DB_HOST` | string | `localhost` | PostgreSQL host |
| `--db-port` | `APP_BACKEND_DB_PORT` | int | `5432` | PostgreSQL port |
| `--db-user` | `APP_BACKEND_DB_USER` | string | `postgres` | Database username |
| `--db-password` | `APP_BACKEND_DB_PASSWORD` | string | `postgres` | Database password |
| `--db-name` | `APP_BACKEND_DB_DATABASE` | string | `iot_db` | Database name |
| `--db-sslmode` | `APP_BACKEND_DB_SSLMODE` | string | `disable` | SSL mode (disable, require, verify-ca, verify-full) |
| **RabbitMQ** |
| `--rabbitmq-url` | `APP_BACKEND_RABBITMQ_URL` | string | `amqp://localhost:5672` | RabbitMQ connection URL |
| `--sensor-queue` | `APP_BACKEND_SENSOR_QUEUE` | string | `sensor-data` | Queue for sensor readings |
| `--device-queue` | `APP_BACKEND_DEVICE_QUEUE` | string | `device-data` | Queue for device messages |

### Backend Example

**Config File**:
```yaml
backend:
  grpc_port: 50051
  metrics_port: 9090
  enable_metrics: true
  db:
    host: postgres.example.com
    port: 5432
    user: iot_user
    password: secret_password
    database: iot_production
    sslmode: require
  rabbitmq:
    url: amqp://user:pass@rabbitmq.example.com:5672
    sensor_queue: sensor-data
    device_queue: device-data
```

**Command Line**:
```bash
./demo-app backend \
  --grpc-port=50051 \
  --metrics-port=9090 \
  --db-host=postgres.example.com \
  --db-port=5432 \
  --db-user=iot_user \
  --db-password=secret_password \
  --db-name=iot_production \
  --db-sslmode=require \
  --rabbitmq-url=amqp://rabbitmq.example.com:5672 \
  --sensor-queue=sensor-data \
  --device-queue=device-data
```

**Environment Variables**:
```bash
export APP_BACKEND_GRPC_PORT=50051
export APP_BACKEND_METRICS_PORT=9090
export APP_BACKEND_DB_HOST=postgres.example.com
export APP_BACKEND_DB_PORT=5432
export APP_BACKEND_DB_USER=iot_user
export APP_BACKEND_DB_PASSWORD=secret_password
export APP_BACKEND_DB_DATABASE=iot_production
export APP_BACKEND_DB_SSLMODE=require
export APP_BACKEND_RABBITMQ_URL=amqp://rabbitmq.example.com:5672
export APP_BACKEND_SENSOR_QUEUE=sensor-data
export APP_BACKEND_DEVICE_QUEUE=device-data

./demo-app backend
```

### Backend Behavior

**Database Migrations**:
- Auto-migrates tables on startup
- Creates `iot_devices` and `sensor_readings` tables
- Idempotent (safe to run multiple times)

**Consumer Behavior**:
- Runs two independent consumers:
  1. **Device Consumer**: Processes device creation (upsert)
  2. **Sensor Consumer**: Processes sensor readings (insert)
- Manual acknowledgment after successful processing
- Automatic reconnection on connection failure
- Retry logic with exponential backoff

**gRPC Server**:
- Listens on `grpc_port`
- Three methods: `GetAllDevice`, `GetDevice`, `GetSensorReadingByDeviceID`
- Graceful shutdown on SIGINT/SIGTERM

## Frontend Configuration

The frontend service provides web UI for visualizing IoT data.

### Frontend Flags

| Flag | Environment Variable | Type | Default | Description |
|------|---------------------|------|---------|-------------|
| `--http-port` | `APP_FRONTEND_HTTP_PORT` | int | `8080` | HTTP server port |
| `--backend-url` | `APP_FRONTEND_BACKEND_URL` | string | `localhost:50051` | Backend gRPC server address |
| `--enable-metrics` | `APP_FRONTEND_ENABLE_METRICS` | bool | `true` | Enable Prometheus metrics at `/metrics` |

### Frontend Example

**Config File**:
```yaml
frontend:
  http_port: 8080
  backend_url: backend.example.com:50051
  enable_metrics: true
```

**Command Line**:
```bash
./demo-app frontend \
  --http-port=8080 \
  --backend-url=backend.example.com:50051 \
  --enable-metrics=true
```

**Environment Variables**:
```bash
export APP_FRONTEND_HTTP_PORT=8080
export APP_FRONTEND_BACKEND_URL=backend.example.com:50051
export APP_FRONTEND_ENABLE_METRICS=true

./demo-app frontend
```

### Frontend Behavior

**HTTP Routes**:
- `/` - Home page
- `/devices` - Device list
- `/devices/{device_id}` - Device detail with sensor readings
- `/metrics` - Prometheus metrics (if enabled)

**gRPC Client**:
- Connects to backend at `backend_url`
- Retries transient failures
- Context timeout: 10 seconds per request

## Global Settings

Global settings apply to all subcommands.

### Global Flags

| Flag | Environment Variable | Type | Default | Description |
|------|---------------------|------|---------|-------------|
| `--config` | `APP_CONFIG` | string | - | Path to config file |
| `--log-level` | `APP_LOG_LEVEL` | string | `info` | Log level (debug, info, warn, error) |
| `--log-format` | `APP_LOG_FORMAT` | string | `json` | Log format (json, text) |

### Logging Configuration

**Log Levels**:
- `debug` - Verbose debugging information
- `info` - General informational messages (default)
- `warn` - Warning messages
- `error` - Error messages only

**Log Formats**:
- `json` - Structured JSON logs (production)
- `text` - Human-readable text logs (development)

**Example**:
```bash
# Debug logging with JSON format
./demo-app backend --log-level=debug --log-format=json

# Error logging with text format
./demo-app backend --log-level=error --log-format=text
```

**JSON Log Example**:
```json
{
  "time": "2025-10-17T10:30:45Z",
  "level": "INFO",
  "msg": "device consumer started",
  "queue": "device-data"
}
```

**Text Log Example**:
```
2025-10-17T10:30:45Z INFO device consumer started queue=device-data
```

## Environment Variables

### Complete Environment Variable List

**Generator**:
```bash
APP_GENERATOR_RABBITMQ_URL=amqp://localhost:5672
APP_GENERATOR_DEVICE_QUEUE=device-data
APP_GENERATOR_SENSOR_QUEUE=sensor-data
APP_GENERATOR_INTERVAL=5s
APP_GENERATOR_NUM_DEVICES=10
APP_GENERATOR_METRICS_PORT=9091
APP_GENERATOR_ENABLE_METRICS=true
```

**Backend**:
```bash
APP_BACKEND_GRPC_PORT=50051
APP_BACKEND_METRICS_PORT=9090
APP_BACKEND_ENABLE_METRICS=true
APP_BACKEND_DB_HOST=localhost
APP_BACKEND_DB_PORT=5432
APP_BACKEND_DB_USER=postgres
APP_BACKEND_DB_PASSWORD=postgres
APP_BACKEND_DB_DATABASE=iot_db
APP_BACKEND_DB_SSLMODE=disable
APP_BACKEND_RABBITMQ_URL=amqp://localhost:5672
APP_BACKEND_SENSOR_QUEUE=sensor-data
APP_BACKEND_DEVICE_QUEUE=device-data
```

**Frontend**:
```bash
APP_FRONTEND_HTTP_PORT=8080
APP_FRONTEND_BACKEND_URL=localhost:50051
APP_FRONTEND_ENABLE_METRICS=true
```

**Global**:
```bash
APP_CONFIG=/path/to/config.yaml
APP_LOG_LEVEL=info
APP_LOG_FORMAT=json
```

### Environment Variable File (.env)

Create `.env` file for local development:

```bash
# Generator
APP_GENERATOR_RABBITMQ_URL=amqp://localhost:5672
APP_GENERATOR_INTERVAL=10s
APP_GENERATOR_NUM_DEVICES=20

# Backend
APP_BACKEND_GRPC_PORT=50051
APP_BACKEND_DB_HOST=localhost
APP_BACKEND_DB_PASSWORD=postgres

# Frontend
APP_FRONTEND_HTTP_PORT=8080
APP_FRONTEND_BACKEND_URL=localhost:50051

# Global
APP_LOG_LEVEL=debug
APP_LOG_FORMAT=json
```

Load with:
```bash
export $(cat .env | xargs)
./demo-app backend
```

## Configuration Examples

### Development Environment

**config.dev.yaml**:
```yaml
generator:
  rabbitmq_url: amqp://localhost:5672
  interval: 5s
  num_devices: 5
  enable_metrics: true

backend:
  grpc_port: 50051
  db:
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    database: iot_dev
    sslmode: disable
  rabbitmq:
    url: amqp://localhost:5672
  enable_metrics: true

frontend:
  http_port: 8080
  backend_url: localhost:50051
  enable_metrics: true

log:
  level: debug
  format: text
```

### Production Environment

**config.prod.yaml**:
```yaml
generator:
  rabbitmq_url: amqp://user:${RABBITMQ_PASSWORD}@rabbitmq.prod.example.com:5672
  device_queue: prod-device-data
  sensor_queue: prod-sensor-data
  interval: 30s
  num_devices: 1000
  metrics_port: 9091
  enable_metrics: true

backend:
  grpc_port: 50051
  metrics_port: 9090
  db:
    host: postgres.prod.example.com
    port: 5432
    user: iot_prod_user
    password: ${DB_PASSWORD}
    database: iot_production
    sslmode: verify-full
  rabbitmq:
    url: amqp://user:${RABBITMQ_PASSWORD}@rabbitmq.prod.example.com:5672
    sensor_queue: prod-sensor-data
    device_queue: prod-device-data
  enable_metrics: true

frontend:
  http_port: 8080
  backend_url: backend.prod.example.com:50051
  enable_metrics: true

log:
  level: info
  format: json
```

### Kubernetes ConfigMap

**configmap.yaml**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: demo-app-config
data:
  config.yaml: |
    generator:
      rabbitmq_url: amqp://rabbitmq.default.svc.cluster.local:5672
      interval: 10s
      num_devices: 100

    backend:
      grpc_port: 50051
      db:
        host: postgres.default.svc.cluster.local
        port: 5432
        user: postgres
        database: iot_db
      rabbitmq:
        url: amqp://rabbitmq.default.svc.cluster.local:5672

    frontend:
      http_port: 8080
      backend_url: backend.default.svc.cluster.local:50051

    log:
      level: info
      format: json
```

**secrets.yaml**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: demo-app-secrets
type: Opaque
stringData:
  db-password: your-secret-password
  rabbitmq-password: your-rabbitmq-password
```

Use in deployment:
```yaml
env:
  - name: APP_BACKEND_DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: demo-app-secrets
        key: db-password
volumeMounts:
  - name: config
    mountPath: /etc/demo-app/config.yaml
    subPath: config.yaml
volumes:
  - name: config
    configMap:
      name: demo-app-config
```

## Validation

The application validates configuration on startup and will fail fast with clear error messages.

**Common Validation Errors**:

```bash
# Invalid interval format
Error: invalid interval: time: invalid duration "5"
Solution: Use valid duration format (e.g., "5s", "10m", "1h")

# Invalid port number
Error: invalid port: 99999
Solution: Use port between 1-65535

# Missing required database configuration
Error: database host is required
Solution: Set --db-host or APP_BACKEND_DB_HOST

# Cannot connect to RabbitMQ
Error: failed to connect to AMQP server: dial tcp: lookup rabbitmq: no such host
Solution: Verify RabbitMQ URL and network connectivity
```

## Best Practices

1. **Use Config Files for Base Settings**: Store common configuration in YAML files
2. **Environment Variables for Secrets**: Never commit passwords to version control
3. **Command-Line Flags for Overrides**: Use flags for quick testing
4. **Separate Configs per Environment**: dev, staging, production configs
5. **Version Control Configs**: Track config changes (except secrets)
6. **Document Custom Settings**: Add comments to config files
7. **Validate Before Deploy**: Test configuration in staging first

## Next Steps

- [API Reference](./api.md) - Explore the gRPC API
- [Deployment Guide](./deployment.md) - Deploy to production
- [Monitoring Guide](./monitoring.md) - Set up monitoring
- [Troubleshooting](./troubleshooting.md) - Fix common issues
