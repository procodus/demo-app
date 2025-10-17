# Demo App - IoT Data Pipeline

[![Go Report Card](https://goreportcard.com/badge/github.com/procodus/demo-app)](https://goreportcard.com/report/github.com/procodus/demo-app)
[![GoDoc](https://pkg.go.dev/badge/github.com/procodus/demo-app)](https://pkg.go.dev/github.com/procodus/demo-app)
[![Coverage](https://img.shields.io/badge/coverage-15.5%25-orange)](./coverage_summary.md)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25.3-blue)](https://go.dev/dl/)

[![PR Checks](https://github.com/procodus/demo-app/actions/workflows/pr-checks.yml/badge.svg)](https://github.com/procodus/demo-app/actions/workflows/pr-checks.yml)
[![Release](https://github.com/procodus/demo-app/actions/workflows/release.yml/badge.svg)](https://github.com/procodus/demo-app/actions/workflows/release.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/procodus/demo-app)](https://github.com/procodus/demo-app/pkgs/container/demo-app)

A production-ready, cloud-native IoT data pipeline built with Go, demonstrating modern microservices architecture with Kubernetes orchestration, comprehensive observability, and automated CI/CD.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)
- [Testing](#testing)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Overview

Demo App is a complete IoT data pipeline consisting of three microservices unified into a single CLI application:

| Component | Purpose | Technology | Ports |
|-----------|---------|------------|-------|
| **Generator** | Generates synthetic IoT sensor data and publishes to message queue | Go, Protobuf, RabbitMQ | 9091 (metrics) |
| **Backend** | Consumes messages, persists to database, serves gRPC API | Go, GORM, PostgreSQL, gRPC | 50051 (gRPC), 9090 (metrics) |
| **Frontend** | Web UI for visualizing IoT data | Go, htmx, Templ | 8080 (HTTP + metrics) |

**Data Flow**: Generator → RabbitMQ → Backend → PostgreSQL ← gRPC ← Frontend

## 🏗️ Architecture

```
┌─────────────┐
│  Generator  │  Generates IoT sensor data
│   (Port:    │  - Temperature, humidity, pressure
│    9091)    │  - Device metadata
└──────┬──────┘  - Publishes to RabbitMQ
       │
       ▼
┌─────────────┐
│  RabbitMQ   │  Message Queue
│   (AMQP)    │  - sensor-data queue
└──────┬──────┘  - device-data queue
       │
       ▼
┌─────────────┐     ┌──────────────┐
│   Backend   │────▶│ PostgreSQL   │
│  (Port:     │     │   Database   │
│   50051,    │     │              │
│   9090)     │     │  - Devices   │
└──────┬──────┘     │  - Readings  │
       │            └──────────────┘
       │ gRPC
       ▼
┌─────────────┐
│  Frontend   │  Web UI (htmx + Templ)
│  (Port:     │  - Device list
│   8080)     │  - Real-time readings
└─────────────┘  - Historical data
```

### Key Technologies

| Category | Technology |
|----------|-----------|
| **Language** | Go 1.25.3 |
| **Message Queue** | RabbitMQ 3 |
| **Database** | PostgreSQL 16 + GORM |
| **API** | gRPC + Protocol Buffers |
| **Frontend** | htmx + Templ (server-side rendering) |
| **CLI** | Cobra + Viper |
| **Observability** | Prometheus (33 metrics), slog (structured logging) |
| **Testing** | Ginkgo + Gomega + testcontainers-go |
| **Container** | Docker (multi-stage Alpine, 30MB) |
| **Orchestration** | Kubernetes + Helm |
| **CI/CD** | GitHub Actions + GoReleaser |

## ✨ Features

### Core Functionality
- ✅ **Synthetic IoT Data Generation** - Configurable concurrent producers
- ✅ **Message Queue Integration** - RabbitMQ with automatic reconnection and retry logic (max 5 attempts)
- ✅ **Database Persistence** - PostgreSQL with GORM ORM, automatic migrations
- ✅ **gRPC API** - High-performance API for device and sensor reading queries
- ✅ **Web UI** - Responsive interface with htmx for dynamic updates
- ✅ **Multi-tenancy Ready** - Device isolation with foreign key constraints

### Observability
- ✅ **Prometheus Metrics** - 33 metrics across all services (connection status, request rates, durations, errors)
- ✅ **Structured Logging** - JSON format with slog (Go standard library)
- ✅ **Health Checks** - Liveness and readiness probes for Kubernetes
- ✅ **Distributed Tracing Ready** - Context propagation throughout the pipeline

### Reliability & Performance
- ✅ **Exponential Backoff** - Retry logic with maximum attempts (prevents infinite loops)
- ✅ **Graceful Shutdown** - Proper cleanup of connections and resources
- ✅ **Connection Pooling** - Efficient database and message queue connections
- ✅ **Concurrent Processing** - Multiple producers and consumers
- ✅ **Non-blocking Operations** - Async message processing

### Security
- ✅ **Non-root Containers** - Runs as UID 1000
- ✅ **Read-only Filesystem** - Kubernetes security contexts
- ✅ **Static Binaries** - No CGO dependencies
- ✅ **Security Scanning Ready** - Trivy, Docker Scout compatible
- ✅ **Minimal Attack Surface** - Alpine base (5MB + 25MB binary)

### DevOps & Deployment
- ✅ **Multi-arch Support** - linux/amd64, linux/arm64 (AWS Graviton, Raspberry Pi)
- ✅ **Helm Chart** - Production-ready with ServiceMonitor, HPA, Ingress
- ✅ **CI/CD** - Automated testing, linting, and releases
- ✅ **Local Development** - Kind cluster scripts with full infrastructure
- ✅ **GitOps Ready** - Declarative Kubernetes manifests

## 🚀 Quick Start

### Prerequisites

- Go 1.25.3+
- Docker (for local development)
- kubectl & Helm (for Kubernetes deployment)

### 1. Run Locally (Docker Compose)

```bash
# Clone repository
git clone git@github.com:procodus/demo-app.git
cd demo-app

# Start infrastructure
docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine

# Build and run
go build -o bin/demo-app ./cmd

# Terminal 1: Start backend
./bin/demo-app backend

# Terminal 2: Start generator
./bin/demo-app generator

# Terminal 3: Start frontend
./bin/demo-app frontend

# Access UI
open http://localhost:8080
```

### 2. Run with Kind (Kubernetes)

```bash
# Setup Kind cluster with infrastructure
./hacks/01-setup-kind-cluster.sh

# Build and push Docker image
./hacks/02-build-push-images.sh

# Deploy with Helm
./hacks/03-deploy-helm.sh

# Access services
kubectl port-forward -n demo-app svc/demo-app-frontend 8080:8080
open http://localhost:8080

# Cleanup when done
./hacks/04-cleanup.sh
```

### 3. Install from Release

```bash
# Download latest release (Linux amd64 example)
VERSION=1.0.0
curl -L -o demo-app.tar.gz \
  https://github.com/procodus/demo-app/releases/download/v${VERSION}/demo-app_${VERSION}_Linux_x86_64.tar.gz

# Extract and install
tar -xzf demo-app.tar.gz
sudo mv demo-app /usr/local/bin/

# Verify
demo-app --help
```

## 📦 Installation

### Option 1: Pre-built Binaries

Download from [GitHub Releases](https://github.com/procodus/demo-app/releases):

- **Linux** (amd64, arm64)
- **macOS** (Intel, Apple Silicon)

### Option 2: Docker Image

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/procodus/demo-app:latest

# Run generator
docker run --rm \
  -e RABBITMQ_URL=amqp://rabbitmq:5672 \
  ghcr.io/procodus/demo-app:latest generator

# Run backend
docker run --rm \
  -p 50051:50051 \
  -e RABBITMQ_URL=amqp://rabbitmq:5672 \
  -e DB_HOST=postgres \
  ghcr.io/procodus/demo-app:latest backend

# Run frontend
docker run --rm \
  -p 8080:8080 \
  -e BACKEND_ADDR=backend:50051 \
  ghcr.io/procodus/demo-app:latest frontend
```

**Platforms**: `linux/amd64`, `linux/arm64`

### Option 3: Helm Chart (Kubernetes)

```bash
# Install from OCI registry
helm install demo-app \
  oci://ghcr.io/procodus/charts/demo-app \
  --version 1.0.0 \
  --namespace demo-app \
  --create-namespace

# With custom values
helm install demo-app \
  oci://ghcr.io/procodus/charts/demo-app \
  --version 1.0.0 \
  --namespace demo-app \
  --create-namespace \
  --set generator.replicaCount=3 \
  --set metrics.enabled=true \
  --set metrics.serviceMonitor.enabled=true

# Check deployment
kubectl get pods -n demo-app
```

**Chart Features**:
- ServiceMonitor (Prometheus Operator)
- HorizontalPodAutoscaler (HPA)
- Ingress support
- Configurable resources and replicas
- External/in-cluster RabbitMQ and PostgreSQL

### Option 4: Build from Source

```bash
# Clone repository
git clone git@github.com:procodus/demo-app.git
cd demo-app

# Build
go build -o bin/demo-app ./cmd

# Run
./bin/demo-app --help
```

## 📖 Usage

### Command Line

```bash
# Show help
demo-app --help

# Run generator (5 concurrent producers, 5s interval)
demo-app generator \
  --producer-count 5 \
  --interval 5s \
  --rabbitmq-url amqp://localhost:5672

# Run backend
demo-app backend \
  --grpc-port 50051 \
  --db-host localhost \
  --db-password secret \
  --rabbitmq-url amqp://localhost:5672

# Run frontend
demo-app frontend \
  --http-port 8080 \
  --backend-addr localhost:50051

# With config file
demo-app backend --config ./config.yaml

# Enable debug logging
demo-app generator --log-level debug
```

### Configuration File

Create `config.yaml`:

```yaml
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

generator:
  producer_count: 5
  interval: 5s
  metrics_port: 9091
  rabbitmq:
    url: amqp://localhost:5672
    sensor_queue: sensor-data
    device_queue: device-data

frontend:
  http_port: 8080
  backend_addr: localhost:50051

log:
  level: info
```

### Environment Variables

Override config with environment variables (prefix `APP_`):

```bash
# Backend configuration
export APP_BACKEND_GRPC_PORT=50051
export APP_BACKEND_DB_PASSWORD=secret
export APP_BACKEND_RABBITMQ_URL=amqp://rabbitmq:5672

# Generator configuration
export APP_GENERATOR_PRODUCER_COUNT=10
export APP_GENERATOR_INTERVAL=3s

# Frontend configuration
export APP_FRONTEND_HTTP_PORT=8080
export APP_FRONTEND_BACKEND_ADDR=backend:50051

# Global configuration
export APP_LOG_LEVEL=debug
```

### API Examples

#### gRPC API (Backend)

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Get all devices
grpcurl -plaintext localhost:50051 iot.IoTService/GetAllDevice

# Get device by ID
grpcurl -plaintext -d '{"device_id": "device-001"}' \
  localhost:50051 iot.IoTService/GetDevice

# Get sensor readings (with pagination)
grpcurl -plaintext -d '{"device_id": "device-001", "page_size": 10}' \
  localhost:50051 iot.IoTService/GetSensorReadingByDeviceID
```

#### Web UI (Frontend)

- **Home**: http://localhost:8080/
- **Devices**: http://localhost:8080/devices
- **Device Detail**: http://localhost:8080/device/{device_id}

#### Metrics Endpoints

```bash
# Generator metrics
curl http://localhost:9091/metrics

# Backend metrics
curl http://localhost:9090/metrics

# Frontend metrics
curl http://localhost:8080/metrics
```

## ⚙️ Configuration

### Generator Options

| Flag | Environment | Default | Description |
|------|-------------|---------|-------------|
| `--producer-count` | `APP_GENERATOR_PRODUCER_COUNT` | `5` | Number of concurrent producers |
| `--interval` | `APP_GENERATOR_INTERVAL` | `5s` | Data generation interval |
| `--rabbitmq-url` | `APP_GENERATOR_RABBITMQ_URL` | `amqp://localhost:5672` | RabbitMQ connection URL |
| `--metrics-port` | `APP_GENERATOR_METRICS_PORT` | `9091` | Prometheus metrics port |

### Backend Options

| Flag | Environment | Default | Description |
|------|-------------|---------|-------------|
| `--grpc-port` | `APP_BACKEND_GRPC_PORT` | `50051` | gRPC server port |
| `--metrics-port` | `APP_BACKEND_METRICS_PORT` | `9090` | Prometheus metrics port |
| `--db-host` | `APP_BACKEND_DB_HOST` | `localhost` | PostgreSQL host |
| `--db-port` | `APP_BACKEND_DB_PORT` | `5432` | PostgreSQL port |
| `--db-user` | `APP_BACKEND_DB_USER` | `postgres` | PostgreSQL user |
| `--db-password` | `APP_BACKEND_DB_PASSWORD` | `postgres` | PostgreSQL password |
| `--db-name` | `APP_BACKEND_DB_NAME` | `iot_db` | PostgreSQL database name |
| `--rabbitmq-url` | `APP_BACKEND_RABBITMQ_URL` | `amqp://localhost:5672` | RabbitMQ connection URL |

### Frontend Options

| Flag | Environment | Default | Description |
|------|-------------|---------|-------------|
| `--http-port` | `APP_FRONTEND_HTTP_PORT` | `8080` | HTTP server port |
| `--backend-addr` | `APP_FRONTEND_BACKEND_ADDR` | `localhost:50051` | Backend gRPC address |

### Global Options

| Flag | Environment | Default | Description |
|------|-------------|---------|-------------|
| `--config` | `APP_CONFIG` | `./config.yaml` | Config file path |
| `--log-level` | `APP_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

## 🛠️ Development

### Setup

```bash
# Install dependencies
go mod download

# Install development tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/a-h/templ/cmd/templ@latest

# Start local infrastructure
docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine
```

### Build

```bash
# Build binary
go build -o bin/demo-app ./cmd

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o bin/demo-app-linux-amd64 ./cmd
GOOS=linux GOARCH=arm64 go build -o bin/demo-app-linux-arm64 ./cmd
GOOS=darwin GOARCH=amd64 go build -o bin/demo-app-darwin-amd64 ./cmd
GOOS=darwin GOARCH=arm64 go build -o bin/demo-app-darwin-arm64 ./cmd

# Build Docker image
docker build -t demo-app:latest .

# Build multi-arch Docker image
docker buildx build --platform linux/amd64,linux/arm64 -t demo-app:latest .
```

### Code Generation

```bash
# Generate gRPC code from protobuf
protoc --go_out=. --go-grpc_out=. api/proto/*.proto
cp api/proto/*.pb.go pkg/iot/

# Generate Templ templates
templ generate -path ./internal/frontend
```

### Linting

```bash
# Run linter (expected: 0 issues)
golangci-lint run ./...

# Auto-fix issues
golangci-lint run --fix ./...

# Format code
go fmt ./...
goimports -w .
```

### Local Development Workflow

```bash
# 1. Make code changes
vim internal/backend/server.go

# 2. Run tests
go test -v -race ./...

# 3. Lint code
golangci-lint run ./...

# 4. Build and test locally
go build -o bin/demo-app ./cmd
./bin/demo-app backend

# 5. For Kubernetes testing
./hacks/02-build-push-images.sh
./hacks/03-deploy-helm.sh
kubectl logs -f -n demo-app -l app.kubernetes.io/component=backend
```

## 🧪 Testing

### Run All Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run with race detection
go test -race ./...

# Run verbose
go test -v ./...
```

### Unit Tests

```bash
# Run unit tests only
go test -short ./...

# Run specific package
go test ./internal/producer/...
go test ./pkg/mq/...
go test ./pkg/logger/...

# Run with Ginkgo
ginkgo -v ./internal/producer/
```

### E2E Tests

E2E tests use **testcontainers-go** to automatically manage Docker containers:

```bash
# Prerequisites: Docker must be running
docker info

# Run all E2E tests
go test ./test/e2e/...

# Run backend E2E tests
go test -v ./test/e2e/backend/...

# Run MQ client E2E tests
go test -v ./test/e2e/mq/...

# Run specific test
go test -v -ginkgo.focus="should consume and save sensor reading" ./test/e2e/backend/...

# With coverage
go test -coverprofile=coverage.out ./test/e2e/...
```

### CI/CD Testing

GitHub Actions automatically runs tests on:
- Every pull request
- Push to main branch
- Tag creation (releases)

**Required checks** (merge blocked if fails):
- ✅ Unit tests with race detection
- ✅ E2E tests with real infrastructure
- ✅ Coverage threshold (≥15%)
- ✅ Linting (0 issues)
- ✅ Build verification

### Component Documentation

| Component | Location |
|-----------|----------|
| **Backend** | [internal/backend/](internal/backend/) |
| **Frontend** | [internal/frontend/](internal/frontend/) |
| **Generator** | [internal/producer/](internal/producer/) + [internal/producer/README.md](internal/producer/README.md) |
| **MQ Client** | [pkg/mq/](pkg/mq/) + [pkg/mq/README.md](pkg/mq/README.md) |
| **Metrics** | [pkg/metrics/](pkg/metrics/) + [pkg/metrics/README.md](pkg/metrics/README.md) |
| **Logger** | [pkg/logger/](pkg/logger/) + [pkg/logger/README.md](pkg/logger/README.md) |

### Infrastructure Documentation

| Document | Description |
|----------|-------------|
| [Dockerfile.README.md](deployments/Dockerfile.README.md) | Docker build documentation |
| [hacks/README.md](hacks/README.md) | Local development scripts |
| [deployments/helm/demo-app/README.md](deployments/helm/demo-app/README.md) | Helm chart guide |
| [deployments/helm/demo-app/INSTALL.md](deployments/helm/demo-app/INSTALL.md) | Quick installation |

### CI/CD Documentation

| Document | Description |
|----------|-------------|
| [.github/WORKFLOWS.md](.github/WORKFLOWS.md) | Complete workflow guide |
| [.github/BRANCH_PROTECTION.md](.github/BRANCH_PROTECTION.md) | Branch protection setup |
| [.github/QUICK_REFERENCE.md](.github/QUICK_REFERENCE.md) | Daily cheat sheet |
| [.github/README.md](.github/README.md) | CI/CD overview |

### API Documentation

```bash
# View protobuf definitions
cat api/proto/sensor.proto

# Generate Go documentation
godoc -http=:6060
open http://localhost:6060/pkg/github.com/procodus/demo-app/
```

### Architecture Decision Records

For significant architectural decisions, see:
- [test/e2e/REFACTORING.md](test/e2e/REFACTORING.md) - E2E test architecture

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

### Development Workflow

1. **Fork the repository**
   ```bash
   gh repo fork procodus/demo-app --clone
   cd demo-app
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make your changes**
   - Follow the coding standards (see [CLAUDE.md](CLAUDE.md))
   - Add tests for new functionality
   - Update documentation as needed

4. **Run checks locally**
   ```bash
   go test -v -race ./...
   golangci-lint run ./...
   go build -o bin/demo-app ./cmd
   ```

5. **Commit with conventional commits**
   ```bash
   git commit -m "feat: add amazing feature"
   git commit -m "fix: resolve bug in sensor consumer"
   git commit -m "docs: update README with new examples"
   ```

6. **Push and create PR**
   ```bash
   git push origin feature/amazing-feature
   gh pr create --fill
   ```

### Conventional Commits

Use conventional commit format for automatic changelog generation:

| Prefix | Description | Example |
|--------|-------------|---------|
| `feat:` | New feature | `feat: add device filtering` |
| `fix:` | Bug fix | `fix: handle nil pointer in consumer` |
| `perf:` | Performance improvement | `perf: optimize database queries` |
| `docs:` | Documentation | `docs: update installation guide` |
| `test:` | Tests | `test: add E2E tests for backend` |
| `chore:` | Maintenance | `chore: update dependencies` |
| `ci:` | CI/CD changes | `ci: add Docker build caching` |

### Code Standards

- **Linting**: 0 issues (run `golangci-lint run ./...`)
- **Testing**: Add unit tests (Ginkgo + Gomega)
- **Coverage**: Maintain or improve coverage
- **Logging**: Use slog (never `fmt.Print*` or `log.Print*`)
- **Errors**: Wrap errors with `fmt.Errorf("%w", err)`
- **Comments**: Add godoc comments for exported functions

### PR Requirements

All PRs must pass:
- ✅ Tests (unit + E2E)
- ✅ Linting (0 issues)
- ✅ Build (binary compiles)
- ✅ Coverage (≥15%)
- ✅ Code review (1 approval)

**Branch protection** prevents merging until all checks pass.

## 🏆 Project Status

### Completed ✅

- **Core Functionality**
  - Unified CLI with Cobra + Viper
  - Generator with concurrent producers
  - Backend with dual consumers (sensor + device)
  - Frontend with htmx + Templ
  - gRPC API (3 endpoints)
  - PostgreSQL persistence with GORM

- **Observability**
  - Prometheus metrics (33 metrics)
  - Structured JSON logging (slog)
  - Health probes

- **Reliability**
  - Exponential backoff with max retries
  - Graceful shutdown
  - Connection pooling

- **Testing**
  - 18 backend E2E tests
  - MQ client E2E tests
  - Unit tests (Ginkgo + Gomega)
  - testcontainers-go integration

- **DevOps**
  - GitHub Actions (PR checks + releases)
  - GoReleaser (multi-platform binaries)
  - Docker multi-arch images
  - Helm OCI chart
  - Kind cluster automation

### In Progress 🚧

- Improving unit test coverage (15.5% → 60%+)
- Adding unit tests for `pkg/generator`
- Improving `pkg/mq` coverage
- Adding backend/frontend unit tests with mocks

### Planned 🔮

- Circuit breaker pattern
- Distributed tracing with OpenTelemetry
- Grafana dashboards
- ArgoCD GitOps deployment
- API documentation with OpenAPI
- Performance benchmarks
- Rate limiting

## 📊 Metrics & Monitoring

### Prometheus Metrics

**33 metrics** across all services:

| Service | Metrics | Examples |
|---------|---------|----------|
| **MQ Client** (8) | Connection status, push/consume counters, failures, duration | `mq_connection_status`, `mq_messages_pushed_total` |
| **Producer** (6) | Messages generated, failures, active producers | `producer_messages_generated_total`, `producer_active_producers` |
| **Backend** (10) | Consumer messages, gRPC requests, in-flight, errors | `backend_grpc_requests_total`, `backend_consumer_messages_total` |
| **Frontend** (9) | HTTP requests, gRPC client calls, template renders | `frontend_http_requests_total`, `frontend_grpc_client_calls_total` |

### Example PromQL Queries

```promql
# Message generation rate
rate(demo_app_producer_messages_generated_total[5m])

# Active producers
demo_app_producer_active_producers

# gRPC request latency (p95)
histogram_quantile(0.95, rate(demo_app_backend_grpc_request_duration_seconds_bucket[5m]))

# MQ connection status (1=connected, 0=disconnected)
demo_app_mq_connection_status

# HTTP request rate by path
rate(demo_app_frontend_http_requests_total[5m])
```

### Grafana Dashboards

Import pre-configured dashboards:
- **System Overview** - All services health
- **Generator Metrics** - Message generation rates
- **Backend Performance** - gRPC latency, consumer throughput
- **Frontend Traffic** - HTTP requests, page views

See [pkg/metrics/README.md](pkg/metrics/README.md) for dashboard JSON.

## 🔒 Security

### Practices

- ✅ Non-root containers (UID 1000)
- ✅ Read-only filesystem support
- ✅ Static binaries (no CGO)
- ✅ Minimal Alpine base (30MB total)
- ✅ No secrets in code or images
- ✅ Dependency scanning (Dependabot)
- ✅ Security linting (gosec)

### Scanning

```bash
# Scan Docker image
trivy image ghcr.io/procodus/demo-app:latest
docker scout cves ghcr.io/procodus/demo-app:latest
grype ghcr.io/procodus/demo-app:latest

# Scan Go dependencies
go list -json -m all | nancy sleuth

# Static analysis
gosec ./...
```

## 📄 License

Copyright © 2025 Procodus Demo App Team

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
