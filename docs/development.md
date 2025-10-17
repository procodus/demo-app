# Development Guide

This guide covers setting up a development environment and contributing to the Demo App.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Code Generation](#code-generation)
- [Testing](#testing)
- [Code Style](#code-style)

## Prerequisites

### Required Tools

```bash
# Go 1.25.3+
go version

# Protocol Buffers compiler
brew install protobuf  # macOS
# or download from https://github.com/protocolbuffers/protobuf/releases

# Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Templ (for frontend templates)
go install github.com/a-h/templ/cmd/templ@latest

# golangci-lint (for linting)
brew install golangci-lint  # macOS
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker (for local infrastructure)
# https://docs.docker.com/get-docker/
```

### Optional Tools

```bash
# grpcurl (for testing gRPC API)
brew install grpcurl  # macOS

# Air (for hot reload during development)
go install github.com/air-verse/air@latest

# goreleaser (for releases)
brew install goreleaser  # macOS
```

## Development Setup

### 1. Clone Repository

```bash
git clone https://github.com/procodus/demo-app.git
cd demo-app
```

### 2. Install Dependencies

```bash
go mod download
go mod verify
```

### 3. Start Infrastructure

```bash
# Start PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=iot_db \
  -p 5432:5432 \
  postgres:16-alpine

# Start RabbitMQ
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management-alpine

# Or use helper script
./hacks/start-local-infrastructure.sh
```

### 4. Build Application

```bash
go build -o bin/demo-app ./cmd
```

### 5. Run Services

```bash
# Terminal 1 - Backend
./bin/demo-app backend

# Terminal 2 - Generator
./bin/demo-app generator

# Terminal 3 - Frontend
./bin/demo-app frontend
```

### 6. Verify Setup

```bash
# Check frontend
curl http://localhost:8080

# Check backend gRPC
grpcurl -plaintext localhost:50051 list

# Check metrics
curl http://localhost:9090/metrics
curl http://localhost:9091/metrics
```

## Project Structure

```
demo-app/
├── cmd/                       # CLI application entry point
│   ├── main.go               # Main entry with Execute()
│   ├── root.go               # Root command setup
│   ├── config.go             # Configuration management
│   ├── backend.go            # Backend subcommand
│   ├── generator.go          # Generator subcommand
│   └── frontend.go           # Frontend subcommand
├── internal/                  # Private application code
│   ├── producer/             # Generator logic
│   │   ├── server.go         # Producer server
│   │   ├── iot_producer.go   # IoT data producer
│   │   └── *_test.go         # Unit tests
│   ├── backend/              # Backend logic
│   │   ├── server.go         # Backend server
│   │   ├── consumer.go       # Sensor consumer
│   │   ├── device_consumer.go # Device consumer
│   │   ├── grpc_service.go   # gRPC implementation
│   │   ├── db.go             # Database init
│   │   ├── models.go         # GORM models
│   │   └── *_test.go         # Unit tests
│   └── frontend/             # Frontend logic
│       ├── server.go         # HTTP server
│       ├── handlers.go       # Request handlers
│       ├── *.templ           # Templ templates
│       └── *_test.go         # Unit tests
├── api/                       # API definitions
│   └── proto/                # gRPC protobuf
│       ├── sensor.proto      # Protobuf schema
│       ├── sensor.pb.go      # Generated code
│       └── sensor_grpc.pb.go # Generated gRPC
├── pkg/                       # Public shared libraries
│   ├── generator/            # Device generation
│   ├── iot/                  # Protobuf (copied from api/)
│   ├── logger/               # Logging utilities
│   ├── mq/                   # RabbitMQ client
│   └── metrics/              # Prometheus metrics
├── test/                      # Test files
│   └── e2e/                  # End-to-end tests
│       ├── testcontainers/   # Container helpers
│       ├── backend/          # Backend E2E tests
│       └── mq/               # MQ client E2E tests
├── deployments/               # Deployment configs
│   ├── Dockerfile            # Container image
│   ├── helm/                 # Helm chart
│   └── k8s/                  # Kubernetes manifests
├── hacks/                     # Development scripts
├── docs/                      # Documentation
├── .golangci.yaml            # Linter configuration
├── .goreleaser.yaml          # Release configuration
├── go.mod                    # Go dependencies
└── go.sum                    # Dependency checksums
```

## Development Workflow

### Adding a New Feature

**1. Create Feature Branch**:
```bash
git checkout -b feature/my-new-feature
```

**2. Make Changes**:
- Write code following style guidelines
- Add unit tests
- Update documentation

**3. Run Tests**:
```bash
go test ./...
```

**4. Run Linter**:
```bash
golangci-lint run ./...
```

**5. Commit Changes**:
```bash
git add .
git commit -m "feat: add new feature description"
```

**6. Push and Create PR**:
```bash
git push origin feature/my-new-feature
# Create PR on GitHub
```

### Hot Reload Development

Use Air for automatic reloading:

Create `.air.toml`:
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["backend"]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

Run with Air:
```bash
air
```

## Code Generation

### Generate Protocol Buffers

After modifying `api/proto/sensor.proto`:

```bash
# Generate Go code
protoc --go_out=. --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  api/proto/*.proto

# Copy to pkg/iot
cp api/proto/*.pb.go pkg/iot/
```

Or use helper script:
```bash
./hacks/generate-proto.sh
```

### Generate Templ Templates

After modifying `.templ` files:

```bash
templ generate -path ./internal/frontend
```

Or watch for changes:
```bash
templ generate -path ./internal/frontend -watch
```

### Generate Mocks

For testing with mocks:

```bash
# Install mockgen
go install github.com/golang/mock/mockgen@latest

# Generate mocks
mockgen -source=pkg/mq/client.go -destination=pkg/mq/mock/client.go -package=mock
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Specific Tests

```bash
# Run tests in specific package
go test ./internal/backend/...

# Run specific test
go test -run TestDeviceConsumer ./internal/backend/

# Run tests with verbose output
go test -v ./internal/backend/...
```

### Run Tests with Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
```

### Run E2E Tests

E2E tests require Docker:

```bash
# Verify Docker is running
docker info

# Run E2E tests
go test ./test/e2e/...

# Run with verbose output
go test -v ./test/e2e/backend/...

# Run specific E2E test
go test -v -ginkgo.focus="should consume and save" ./test/e2e/backend/...
```

### Run Tests with Race Detection

```bash
go test -race ./...
```

### Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...
```

## Code Style

### Formatting

```bash
# Format all Go files
go fmt ./...

# Use goimports for import organization
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

### Linting

The project uses golangci-lint v2.2.0 with strict configuration:

```bash
# Run linter
golangci-lint run ./...

# Run with verbose output
golangci-lint run -v ./...

# Fix auto-fixable issues
golangci-lint run --fix ./...
```

**Key Rules**:
- Use slog for logging (no fmt.Print*)
- Check all errors
- Use context properly
- Follow import order (standard → external → local)
- Add package comments
- End comments with periods

### Commit Messages

Follow Conventional Commits:

```bash
feat: add new feature
fix: fix bug in consumer
docs: update API documentation
style: format code
refactor: simplify error handling
test: add unit tests for generator
chore: update dependencies
```

### Code Review Checklist

Before submitting PR:

- [ ] All tests pass
- [ ] Linter passes (0 issues)
- [ ] Code is formatted
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No debug code left
- [ ] Error handling is complete
- [ ] Tests cover new code

## Debugging

### Debug with Delve

Install Delve:
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Debug backend:
```bash
dlv debug ./cmd -- backend --db-host=localhost
```

Debug tests:
```bash
dlv test ./internal/backend
```

### VS Code Debug Configuration

Add to `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd",
      "args": ["backend"],
      "env": {
        "APP_BACKEND_DB_HOST": "localhost",
        "APP_BACKEND_DB_PASSWORD": "postgres"
      }
    },
    {
      "name": "Debug Generator",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd",
      "args": ["generator"]
    }
  ]
}
```

### Profiling

**CPU Profile**:
```bash
go test -cpuprofile=cpu.prof -bench=. ./internal/backend
go tool pprof cpu.prof
```

**Memory Profile**:
```bash
go test -memprofile=mem.prof -bench=. ./internal/backend
go tool pprof mem.prof
```

**Live Profiling**:
```bash
# Enable pprof endpoint (already enabled)
# Access at http://localhost:9090/debug/pprof/

# Download profile
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

## Building

### Local Build

```bash
go build -o bin/demo-app ./cmd
```

### Build with Flags

```bash
go build -ldflags="-s -w -X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)" \
  -o bin/demo-app ./cmd
```

### Cross-Compile

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o bin/demo-app-linux-amd64 ./cmd

# Linux arm64
GOOS=linux GOARCH=arm64 go build -o bin/demo-app-linux-arm64 ./cmd

# macOS amd64
GOOS=darwin GOARCH=amd64 go build -o bin/demo-app-darwin-amd64 ./cmd

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/demo-app-darwin-arm64 ./cmd
```

### Docker Build

```bash
# Single architecture
docker build -t demo-app:latest -f deployments/Dockerfile .

# Multi-architecture
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t demo-app:latest \
  -f deployments/Dockerfile \
  --push \
  .
```

## Release Process

Releases are automated via GitHub Actions:

1. Create and push a tag:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

2. GitHub Actions will:
   - Run tests and linting
   - Build binaries (4 platforms)
   - Build Docker images (multi-arch)
   - Package Helm chart
   - Create GitHub Release
   - Push images to ghcr.io
   - Push Helm chart to ghcr.io

See `.github/workflows/release.yml` for details.

## Next Steps

- [Testing Guide](./testing.md) - Detailed testing strategies
- [Contributing Guide](./contributing.md) - Contribution guidelines
- [API Reference](./api.md) - Understand the API
- [Architecture](./architecture.md) - System design
