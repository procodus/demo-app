# Architecture Overview

This document describes the architecture of the Demo App IoT Data Pipeline, a distributed system for generating, processing, and visualizing IoT sensor data.

## Table of Contents

- [System Architecture](#system-architecture)
- [Components](#components)
- [Data Flow](#data-flow)
- [Communication Patterns](#communication-patterns)
- [Technology Stack](#technology-stack)
- [Design Decisions](#design-decisions)

## System Architecture

The Demo App consists of three microservices that work together to form a complete IoT data pipeline:

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  Generator  │         │   Backend   │         │  Frontend   │
│  (Producer) │────────▶│  (Consumer) │◀────────│   (Web UI)  │
│             │  AMQP   │             │  gRPC   │             │
└─────────────┘         └─────────────┘         └─────────────┘
       │                       │                        │
       │                       │                        │
       ▼                       ▼                        │
┌─────────────┐         ┌─────────────┐                │
│  RabbitMQ   │         │ PostgreSQL  │                │
│   (Queue)   │         │  (Database) │                │
└─────────────┘         └─────────────┘                │
                                                        │
       ┌────────────────────────────────────────────────┘
       ▼
┌─────────────┐
│ Prometheus  │
│  (Metrics)  │
└─────────────┘
```

## Components

### 1. Generator (Producer Service)

**Purpose**: Generates synthetic IoT sensor data and publishes to message queues.

**Responsibilities**:
- Generate realistic IoT device metadata (location, MAC address, firmware)
- Create synthetic sensor readings (temperature, humidity, pressure, battery)
- Publish device creation messages to `device-data` queue
- Publish sensor reading messages to `sensor-data` queue
- Support multiple concurrent producers for load testing

**Technology**:
- Go 1.25.3
- Protocol Buffers (protobuf) for message serialization
- RabbitMQ client with auto-reconnection

**Ports**:
- `9091` - Prometheus metrics endpoint

**Configuration**:
```yaml
generator:
  rabbitmq_url: amqp://localhost:5672
  device_queue: device-data
  sensor_queue: sensor-data
  interval: 5s
  num_devices: 10
```

### 2. Backend (Consumer Service)

**Purpose**: Consumes messages, persists data, and provides query API.

**Responsibilities**:
- Run two independent consumers:
  - **Device Consumer**: Process device creation messages
  - **Sensor Consumer**: Process sensor reading messages
- Persist data to PostgreSQL with GORM ORM
- Provide gRPC API for querying devices and readings
- Enforce foreign key relationships (sensor readings must have valid device)
- Automatic database migrations

**Technology**:
- Go 1.25.3
- GORM (Object-Relational Mapping)
- gRPC server
- PostgreSQL 16

**Ports**:
- `50051` - gRPC API server
- `9090` - Prometheus metrics endpoint

**Database Tables**:
- `iot_devices` - Device metadata (device_id is primary key)
- `sensor_readings` - Time-series sensor data with FK to iot_devices

**Configuration**:
```yaml
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
```

### 3. Frontend (Web UI)

**Purpose**: Provides web interface for visualizing IoT data.

**Responsibilities**:
- Display device list and details
- Show sensor readings with pagination
- Real-time updates with htmx
- Query backend via gRPC
- Server-side rendering with Templ templates

**Technology**:
- Go 1.25.3
- htmx (dynamic UI updates)
- Templ (type-safe templates)
- gRPC client

**Ports**:
- `8080` - HTTP server
- `8080/metrics` - Prometheus metrics (when enabled)

**Configuration**:
```yaml
frontend:
  http_port: 8080
  backend_url: localhost:50051
```

## Data Flow

### 1. Device Creation Flow

```
┌──────────┐    Device     ┌──────────┐    Consume    ┌──────────┐
│Generator │─────────────▶│RabbitMQ  │──────────────▶│ Backend  │
│          │   (protobuf)  │  Queue   │               │ Consumer │
└──────────┘               └──────────┘               └──────────┘
                                                             │
                                                             │ Insert/Upsert
                                                             ▼
                                                       ┌──────────┐
                                                       │PostgreSQL│
                                                       │ iot_     │
                                                       │ devices  │
                                                       └──────────┘
```

**Message Format** (Protobuf):
```protobuf
message IoTDevice {
  string device_id = 1;
  string location = 2;
  string mac_address = 3;
  string ip_address = 4;
  string firmware = 5;
  float latitude = 6;
  float longitude = 7;
  int64 last_seen = 8;
}
```

### 2. Sensor Reading Flow

```
┌──────────┐   Reading     ┌──────────┐    Consume    ┌──────────┐
│Generator │─────────────▶│RabbitMQ  │──────────────▶│ Backend  │
│          │   (protobuf)  │  Queue   │               │ Consumer │
└──────────┘               └──────────┘               └──────────┘
                                                             │
                                                             │ Insert
                                                             ▼
                                                       ┌──────────┐
                                                       │PostgreSQL│
                                                       │ sensor_  │
                                                       │ readings │
                                                       └──────────┘
```

**Message Format** (Protobuf):
```protobuf
message SensorReading {
  string device_id = 1;
  int64 timestamp = 2;
  double temperature = 3;
  double humidity = 4;
  double pressure = 5;
  double battery_level = 6;
}
```

### 3. Query Flow

```
┌──────────┐   HTTP       ┌──────────┐    gRPC      ┌──────────┐
│ Browser  │─────────────▶│Frontend  │─────────────▶│ Backend  │
│          │   GET        │  Server  │   Request    │  gRPC    │
└──────────┘              └──────────┘              └──────────┘
     ▲                          │                          │
     │                          │                          │ Query
     │      HTML                │                          ▼
     │    (htmx/Templ)          │                    ┌──────────┐
     │                          │                    │PostgreSQL│
     └──────────────────────────┘                    │ Database │
                                                     └──────────┘
```

**gRPC Methods**:
- `GetAllDevice()` - Retrieve all devices
- `GetDevice(device_id)` - Get specific device
- `GetSensorReadingByDeviceID(device_id, page_token)` - Get readings with pagination

## Communication Patterns

### 1. Asynchronous Messaging (Generator → Backend)

**Pattern**: Publish-Subscribe via RabbitMQ

**Benefits**:
- Decouples producer from consumer
- Handles backpressure automatically
- Message persistence (durable queues)
- Automatic reconnection on failure

**Queue Configuration**:
- **Durability**: Non-durable (for demo purposes)
- **Auto-delete**: False
- **Exclusive**: False
- **Acknowledgment**: Manual ack after processing

**Retry Logic**:
- Exponential backoff (100ms → 10s max)
- Maximum 5 retry attempts
- Context cancellation support

### 2. Synchronous RPC (Frontend ↔ Backend)

**Pattern**: gRPC (HTTP/2 + Protocol Buffers)

**Benefits**:
- Type-safe API with protobuf contracts
- Efficient binary serialization
- Built-in code generation
- Streaming support (future enhancement)

**Error Handling**:
- gRPC status codes
- Detailed error messages
- Retry on transient failures

### 3. Metrics Collection (All Services → Prometheus)

**Pattern**: Pull-based metrics scraping

**Benefits**:
- Centralized monitoring
- Time-series data
- Flexible querying with PromQL
- Alerting capabilities

**Metrics Endpoints**:
- Generator: `http://localhost:9091/metrics`
- Backend: `http://localhost:9090/metrics`
- Frontend: `http://localhost:8080/metrics`

## Technology Stack

### Core Technologies

| Layer | Technology | Version | Purpose |
|-------|------------|---------|---------|
| **Language** | Go | 1.25.3 | All services |
| **Message Queue** | RabbitMQ | 3 | Async messaging |
| **Database** | PostgreSQL | 16 | Data persistence |
| **RPC** | gRPC | Latest | Frontend-Backend communication |
| **Serialization** | Protocol Buffers | v3 | Message format |
| **ORM** | GORM | Latest | Database access |

### Infrastructure

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Containerization** | Docker | Multi-arch images |
| **Orchestration** | Kubernetes | Production deployment |
| **Package Manager** | Helm | Kubernetes deployments |
| **CI/CD** | GitHub Actions | Automated testing and releases |
| **Monitoring** | Prometheus | Metrics collection |
| **Logging** | slog (JSON) | Structured logging |

### Frontend Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Templates** | Templ | Type-safe HTML templates |
| **Dynamic UI** | htmx | Partial page updates |
| **Rendering** | Server-side | Go templates |

## Design Decisions

### 1. Why Monorepo with Unified CLI?

**Decision**: Single repository with one binary and multiple subcommands

**Rationale**:
- Simplified dependency management (single `go.mod`)
- Shared code reuse (`pkg/` packages)
- Easier testing across services
- Single build process
- Cobra provides excellent CLI experience

**Trade-offs**:
- Larger binary size (~22MB)
- All services deploy together (mitigated by subcommands)

### 2. Why RabbitMQ for Messaging?

**Decision**: RabbitMQ over Kafka, NATS, or direct HTTP

**Rationale**:
- Mature, battle-tested message broker
- Excellent Go client library
- Built-in management UI
- Message persistence
- Simple deployment

**Trade-offs**:
- Single point of failure (mitigate with clustering)
- Not designed for high-throughput streaming (acceptable for IoT demo)

### 3. Why gRPC for Backend API?

**Decision**: gRPC over REST/JSON

**Rationale**:
- Type-safe contracts with protobuf
- Efficient binary serialization
- Code generation for clients
- HTTP/2 multiplexing
- Streaming support for future features

**Trade-offs**:
- Less human-readable than JSON
- Requires code generation step
- Browser support requires gRPC-web proxy

### 4. Why PostgreSQL over NoSQL?

**Decision**: PostgreSQL with GORM ORM

**Rationale**:
- Strong consistency for device-reading relationships
- Foreign key constraints prevent orphaned data
- Excellent JSON support for flexible fields
- GORM provides excellent Go integration
- Time-series data support with indexing

**Trade-offs**:
- Vertical scaling limits (acceptable for demo scale)
- More complex than key-value stores

### 5. Why Server-Side Rendering (htmx + Templ)?

**Decision**: Server-side rendering over React/Vue SPA

**Rationale**:
- Simpler architecture (no separate frontend build)
- Type-safe templates with Templ
- Progressive enhancement with htmx
- Faster initial page loads
- Consistent Go stack

**Trade-offs**:
- Less rich interactivity than SPAs
- Server load for every render

### 6. Why Unified CLI with Cobra?

**Decision**: Cobra framework with subcommands

**Rationale**:
- Clean CLI interface
- Flexible configuration (flags, env vars, config files)
- Built-in help and documentation
- Viper integration for config management

**Example Usage**:
```bash
./demo-app generator --interval=10s
./demo-app backend --db-host=postgres.example.com
./demo-app frontend --http-port=3000
```

## Scalability Considerations

### Horizontal Scaling

| Component | Scalability | Notes |
|-----------|-------------|-------|
| **Generator** | ✅ Excellent | Stateless, scale to N instances |
| **Backend Consumer** | ⚠️ Limited | Queue consumer competition |
| **Backend gRPC** | ✅ Excellent | Stateless, load balance with K8s |
| **Frontend** | ✅ Excellent | Stateless, scale with replicas |
| **RabbitMQ** | ⚠️ Requires clustering | Single instance SPOF |
| **PostgreSQL** | ⚠️ Vertical only | Read replicas for queries |

### Performance Optimizations

1. **Database Indexes**:
   - `idx_device_timestamp` on sensor_readings
   - `idx_timestamp` for time-range queries
   - `idx_last_seen` on iot_devices

2. **Connection Pooling**:
   - GORM manages PostgreSQL connection pool
   - RabbitMQ client connection reuse

3. **Pagination**:
   - Cursor-based pagination for sensor readings
   - Prevents large result sets

4. **Metrics Caching**:
   - Prometheus scrapes at intervals
   - Avoids real-time calculation overhead

## Security Considerations

1. **Authentication**: Not implemented (demo purposes)
2. **Encryption**:
   - Use TLS for gRPC in production
   - Use AMQPS for RabbitMQ in production
3. **Input Validation**: Protobuf schema validation
4. **SQL Injection**: GORM prevents SQL injection
5. **Secrets Management**: Environment variables or Kubernetes secrets

## Future Enhancements

1. **Streaming gRPC**: Real-time sensor data updates
2. **Authentication**: JWT-based auth for frontend
3. **Multi-tenancy**: Support multiple organizations
4. **Data Retention**: Time-series data compaction
5. **Event Sourcing**: Audit log for all changes
6. **WebSockets**: Real-time frontend updates
7. **GraphQL**: Alternative query API
8. **Distributed Tracing**: OpenTelemetry integration

## References

- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers](https://protobuf.dev/)
- [GORM Documentation](https://gorm.io/docs/)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/getstarted.html)
- [htmx Documentation](https://htmx.org/docs/)
- [Templ Documentation](https://templ.guide/)
