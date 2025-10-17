# Getting Started

This guide will help you get the Demo App IoT Data Pipeline up and running quickly.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Installation Methods](#installation-methods)
- [First Steps](#first-steps)
- [Next Steps](#next-steps)

## Prerequisites

### Required

- **Go 1.25.3+** - [Download](https://go.dev/dl/)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **RabbitMQ** - Message queue (via Docker or local install)
- **PostgreSQL 16+** - Database (via Docker or local install)

### Optional

- **Kubernetes** (Kind, Minikube, or production cluster)
- **Helm 3+** - For Kubernetes deployments
- **Docker Compose** - For multi-container local setup

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Disk | 10 GB | 20+ GB |

## Quick Start

The fastest way to get started is running all services locally with Docker containers for infrastructure.

### Step 1: Start Infrastructure

Start RabbitMQ and PostgreSQL using Docker:

```bash
# Start RabbitMQ
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management-alpine

# Start PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=iot_db \
  -p 5432:5432 \
  postgres:16-alpine
```

**Verify services are running**:
```bash
# Check RabbitMQ
curl http://localhost:15672
# Login: guest/guest

# Check PostgreSQL
docker exec -it postgres psql -U postgres -d iot_db -c "SELECT version();"
```

### Step 2: Clone and Build

```bash
# Clone repository
git clone https://github.com/procodus/demo-app.git
cd demo-app

# Install dependencies
go mod download

# Build the application
go build -o bin/demo-app ./cmd
```

### Step 3: Run Services

Open three terminal windows and run each service:

**Terminal 1 - Backend**:
```bash
./bin/demo-app backend \
  --db-host=localhost \
  --db-port=5432 \
  --db-user=postgres \
  --db-password=postgres \
  --db-name=iot_db \
  --rabbitmq-url=amqp://guest:guest@localhost:5672 \
  --grpc-port=50051
```

**Terminal 2 - Generator**:
```bash
./bin/demo-app generator \
  --rabbitmq-url=amqp://guest:guest@localhost:5672 \
  --interval=5s \
  --num-devices=10
```

**Terminal 3 - Frontend**:
```bash
./bin/demo-app frontend \
  --backend-url=localhost:50051 \
  --http-port=8080
```

### Step 4: Verify Installation

1. **Check Frontend UI**:
   - Open browser: http://localhost:8080
   - You should see the device list

2. **Check RabbitMQ Messages**:
   - Open: http://localhost:15672
   - Login: guest/guest
   - Go to "Queues" tab - should see messages flowing

3. **Check Database**:
   ```bash
   docker exec -it postgres psql -U postgres -d iot_db -c "SELECT COUNT(*) FROM iot_devices;"
   docker exec -it postgres psql -U postgres -d iot_db -c "SELECT COUNT(*) FROM sensor_readings;"
   ```

4. **Check Metrics**:
   - Generator: http://localhost:9091/metrics
   - Backend: http://localhost:9090/metrics
   - Frontend: http://localhost:8080/metrics

**Success!** You should now have:
- Devices being generated and saved
- Sensor readings flowing through the system
- Web UI showing live data

## Installation Methods

### Method 1: Pre-built Binaries

Download from [GitHub Releases](https://github.com/procodus/demo-app/releases):

```bash
# Linux (amd64)
wget https://github.com/procodus/demo-app/releases/download/v1.0.0/demo-app_Linux_x86_64.tar.gz
tar -xzf demo-app_Linux_x86_64.tar.gz
sudo mv demo-app /usr/local/bin/

# macOS (amd64)
wget https://github.com/procodus/demo-app/releases/download/v1.0.0/demo-app_Darwin_x86_64.tar.gz
tar -xzf demo-app_Darwin_x86_64.tar.gz
sudo mv demo-app /usr/local/bin/

# macOS (arm64 - Apple Silicon)
wget https://github.com/procodus/demo-app/releases/download/v1.0.0/demo-app_Darwin_arm64.tar.gz
tar -xzf demo-app_Darwin_arm64.tar.gz
sudo mv demo-app /usr/local/bin/
```

### Method 2: Docker Images

Pull and run using Docker:

```bash
# Pull images
docker pull ghcr.io/procodus/demo-app:latest

# Run backend
docker run -d \
  --name demo-backend \
  -p 50051:50051 \
  -p 9090:9090 \
  -e DB_HOST=postgres \
  -e RABBITMQ_URL=amqp://rabbitmq:5672 \
  ghcr.io/procodus/demo-app:latest backend

# Run generator
docker run -d \
  --name demo-generator \
  -p 9091:9091 \
  -e RABBITMQ_URL=amqp://rabbitmq:5672 \
  ghcr.io/procodus/demo-app:latest generator

# Run frontend
docker run -d \
  --name demo-frontend \
  -p 8080:8080 \
  -e BACKEND_URL=backend:50051 \
  ghcr.io/procodus/demo-app:latest frontend
```

### Method 3: Build from Source

```bash
# Prerequisites: Go 1.25.3+
git clone https://github.com/procodus/demo-app.git
cd demo-app

# Install dependencies
go mod download

# Build
go build -o bin/demo-app ./cmd

# Verify
./bin/demo-app --version
```

### Method 4: Kubernetes with Helm

```bash
# Add Helm repository (OCI)
helm pull oci://ghcr.io/procodus/charts/demo-app --version 1.0.0
helm install demo-app oci://ghcr.io/procodus/charts/demo-app --version 1.0.0

# Or with custom values
helm install demo-app oci://ghcr.io/procodus/charts/demo-app \
  --version 1.0.0 \
  --set generator.replicaCount=2 \
  --set backend.database.host=my-postgres.example.com
```

For detailed Kubernetes deployment, see [Deployment Guide](./deployment.md).

## First Steps

### 1. Explore the Web UI

Navigate to http://localhost:8080:

- **Home Page** (`/`): Overview and navigation
- **Devices** (`/devices`): List all IoT devices
- **Device Details** (`/devices/{device_id}`): View sensor readings for specific device

### 2. Query the gRPC API

Install grpcurl for testing:

```bash
# Install grpcurl
brew install grpcurl  # macOS
# or
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Get all devices
grpcurl -plaintext localhost:50051 iot.SensorService/GetAllDevice

# Get specific device
grpcurl -plaintext -d '{"device_id": "device-001"}' \
  localhost:50051 iot.SensorService/GetDevice

# Get sensor readings
grpcurl -plaintext -d '{"device_id": "device-001"}' \
  localhost:50051 iot.SensorService/GetSensorReadingByDeviceID
```

### 3. Monitor with Prometheus

If you have Prometheus installed:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'demo-generator'
    static_configs:
      - targets: ['localhost:9091']

  - job_name: 'demo-backend'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'demo-frontend'
    static_configs:
      - targets: ['localhost:8080']
```

Start Prometheus:
```bash
prometheus --config.file=prometheus.yml
```

Access at http://localhost:9090

**Example Queries**:
```promql
# Message generation rate
rate(demo_app_producer_messages_generated_total[5m])

# Backend processing duration
histogram_quantile(0.95, rate(demo_app_backend_processing_duration_seconds_bucket[5m]))

# HTTP request rate
rate(demo_app_frontend_http_requests_total[5m])
```

### 4. Check Logs

All services use structured JSON logging with slog:

```bash
# View backend logs
./bin/demo-app backend 2>&1 | jq .

# Filter for errors only
./bin/demo-app backend 2>&1 | jq 'select(.level == "ERROR")'

# Watch specific device processing
./bin/demo-app backend 2>&1 | jq 'select(.device_id == "device-001")'
```

### 5. Test with Custom Configuration

Create a config file `config.yaml`:

```yaml
generator:
  rabbitmq_url: amqp://localhost:5672
  device_queue: device-data
  sensor_queue: sensor-data
  interval: 10s
  num_devices: 50

backend:
  grpc_port: 50051
  db:
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    database: iot_db
  rabbitmq:
    url: amqp://localhost:5672
    sensor_queue: sensor-data
    device_queue: device-data

frontend:
  http_port: 8080
  backend_url: localhost:50051

log:
  level: debug
```

Run with config file:
```bash
./bin/demo-app generator --config=config.yaml
./bin/demo-app backend --config=config.yaml
./bin/demo-app frontend --config=config.yaml
```

## Common Commands

### Start/Stop Services

```bash
# Stop infrastructure
docker stop rabbitmq postgres
docker rm rabbitmq postgres

# Clean database
docker exec -it postgres psql -U postgres -d iot_db -c "DROP TABLE IF EXISTS sensor_readings CASCADE;"
docker exec -it postgres psql -U postgres -d iot_db -c "DROP TABLE IF EXISTS iot_devices CASCADE;"

# Restart backend (migrations will recreate tables)
./bin/demo-app backend
```

### Development Workflow

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run ./...

# Run tests
go test ./...

# Run specific test
go test -v -run TestDeviceGeneration ./pkg/generator/

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Helper Scripts

The project includes helper scripts in `hacks/`:

```bash
# Start local infrastructure
./hacks/start-local-infrastructure.sh

# Stop local infrastructure
./hacks/stop-local-infrastructure.sh

# Run all services
./hacks/run-all-services.sh

# Clean build artifacts
./hacks/clean.sh

# Rebuild everything
./hacks/rebuild.sh
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080
# Kill process
kill -9 <PID>
```

### Cannot Connect to RabbitMQ

```bash
# Check RabbitMQ is running
docker ps | grep rabbitmq

# Check logs
docker logs rabbitmq

# Restart RabbitMQ
docker restart rabbitmq
```

### Cannot Connect to PostgreSQL

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check logs
docker logs postgres

# Test connection
docker exec -it postgres psql -U postgres -d iot_db -c "SELECT 1;"
```

### Services Not Starting

```bash
# Check Go version
go version  # Should be 1.25.3+

# Rebuild
go clean
go build -o bin/demo-app ./cmd

# Check dependencies
go mod verify
go mod tidy
```

For more troubleshooting help, see [Troubleshooting Guide](./troubleshooting.md).

## Next Steps

Now that you have the system running:

1. **Explore the API**: Read the [API Reference](./api.md)
2. **Deploy to Kubernetes**: Follow the [Deployment Guide](./deployment.md)
3. **Set up Monitoring**: Configure [Monitoring](./monitoring.md)
4. **Customize Configuration**: See [Configuration Guide](./configuration.md)
5. **Contribute**: Check [Contributing Guidelines](./contributing.md)

## Getting Help

- **Documentation**: Browse this docs folder
- **Issues**: [GitHub Issues](https://github.com/procodus/demo-app/issues)
- **Discussions**: [GitHub Discussions](https://github.com/procodus/demo-app/discussions)
