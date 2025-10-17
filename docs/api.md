# API Reference

This document describes the gRPC API provided by the backend service.

## Table of Contents

- [Overview](#overview)
- [Protocol](#protocol)
- [Service Definition](#service-definition)
- [Data Models](#data-models)
- [API Methods](#api-methods)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Overview

The backend service exposes a gRPC API for querying IoT device information and sensor readings.

**Base URL**: `localhost:50051` (default)

**Protocol**: gRPC over HTTP/2

**Serialization**: Protocol Buffers (protobuf v3)

**Authentication**: None (demo purposes)

## Protocol

### gRPC vs REST

This API uses gRPC instead of REST for:
- **Type Safety**: Protobuf schemas ensure type correctness
- **Performance**: Binary serialization is faster than JSON
- **Code Generation**: Auto-generate clients in multiple languages
- **Streaming**: Support for bidirectional streaming (future)

### Protobuf Schema

The API is defined in `api/proto/sensor.proto`:

```protobuf
syntax = "proto3";

package iot;
option go_package = "procodus.dev/demo-app/pkg/iot";

service SensorService {
  rpc GetAllDevice (GetAllDeviceRequest) returns (GetAllDeviceResponse);
  rpc GetDevice (GetDeviceByIDRequest) returns (GetDeviceByIDResponse);
  rpc GetSensorReadingByDeviceID (GetSensorReadingByDeviceIDRequest) returns (GetSensorReadingByDeviceIDResponse);
}
```

## Service Definition

### SensorService

The main service for querying IoT data.

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `GetAllDevice` | `GetAllDeviceRequest` | `GetAllDeviceResponse` | Retrieve all devices |
| `GetDevice` | `GetDeviceByIDRequest` | `GetDeviceByIDResponse` | Get specific device |
| `GetSensorReadingByDeviceID` | `GetSensorReadingByDeviceIDRequest` | `GetSensorReadingByDeviceIDResponse` | Get sensor readings for device |

## Data Models

### IoTDevice

Represents an IoT device with metadata.

```protobuf
message IoTDevice {
  string device_id = 1;      // Unique device identifier (e.g., "device-001")
  string location = 2;       // Physical location (e.g., "San Francisco")
  string mac_address = 3;    // MAC address (e.g., "00:1B:44:11:3A:B7")
  string ip_address = 4;     // IP address (e.g., "192.168.1.100")
  string firmware = 5;       // Firmware version (e.g., "v1.2.3")
  float latitude = 6;        // GPS latitude (-90 to 90)
  float longitude = 7;       // GPS longitude (-180 to 180)
  int64 last_seen = 8;       // Unix timestamp of last contact
}
```

**Field Descriptions**:
- `device_id`: Primary key, unique across all devices
- `location`: Human-readable location name
- `mac_address`: Network interface MAC address
- `ip_address`: IPv4 or IPv6 address
- `firmware`: Software version running on device
- `latitude`/`longitude`: GPS coordinates
- `last_seen`: Timestamp when device last sent data (Unix seconds)

### SensorReading

Represents a single sensor reading from a device.

```protobuf
message SensorReading {
  string device_id = 1;      // Foreign key to IoTDevice
  int64 timestamp = 2;       // Reading timestamp (Unix seconds)
  double temperature = 3;    // Temperature in Celsius
  double humidity = 4;       // Relative humidity (0-100%)
  double pressure = 5;       // Atmospheric pressure (hPa)
  double battery_level = 6;  // Battery level (0-100%)
}
```

**Field Descriptions**:
- `device_id`: Must match an existing device
- `timestamp`: When the reading was taken
- `temperature`: Temperature in degrees Celsius (-40 to 85)
- `humidity`: Relative humidity percentage (0-100)
- `pressure`: Atmospheric pressure in hectopascals (300-1100)
- `battery_level`: Device battery percentage (0-100)

## API Methods

### GetAllDevice

Retrieve all devices in the system.

**Request**:
```protobuf
message GetAllDeviceRequest {
  // Empty - no parameters required
}
```

**Response**:
```protobuf
message GetAllDeviceResponse {
  repeated IoTDevice device = 1;  // List of all devices
}
```

**Use Case**: Display device list in UI, system monitoring

**Example**:
```bash
grpcurl -plaintext localhost:50051 iot.SensorService/GetAllDevice
```

**Response Example**:
```json
{
  "device": [
    {
      "device_id": "device-001",
      "location": "San Francisco",
      "mac_address": "00:1B:44:11:3A:B7",
      "ip_address": "192.168.1.100",
      "firmware": "v1.2.3",
      "latitude": 37.7749,
      "longitude": -122.4194,
      "last_seen": "1697548800"
    },
    {
      "device_id": "device-002",
      "location": "New York",
      "mac_address": "00:1B:44:22:4C:D8",
      "ip_address": "192.168.1.101",
      "firmware": "v1.2.3",
      "latitude": 40.7128,
      "longitude": -74.0060,
      "last_seen": "1697548805"
    }
  ]
}
```

**Performance**:
- Returns all devices in single response
- Suitable for <10,000 devices
- For larger datasets, pagination should be implemented

---

### GetDevice

Retrieve a specific device by ID.

**Request**:
```protobuf
message GetDeviceByIDRequest {
  string device_id = 1;  // Device ID to query
}
```

**Response**:
```protobuf
message GetDeviceByIDResponse {
  IoTDevice device = 1;  // Device details (null if not found)
}
```

**Use Case**: Device detail page, health checks

**Example**:
```bash
grpcurl -plaintext -d '{"device_id": "device-001"}' \
  localhost:50051 iot.SensorService/GetDevice
```

**Response Example**:
```json
{
  "device": {
    "device_id": "device-001",
    "location": "San Francisco",
    "mac_address": "00:1B:44:11:3A:B7",
    "ip_address": "192.168.1.100",
    "firmware": "v1.2.3",
    "latitude": 37.7749,
    "longitude": -122.4194,
    "last_seen": "1697548800"
  }
}
```

**Error Cases**:
- Device not found: Returns empty `device` field
- Invalid device_id format: Returns validation error

---

### GetSensorReadingByDeviceID

Retrieve sensor readings for a specific device with pagination.

**Request**:
```protobuf
message GetSensorReadingByDeviceIDRequest {
  string device_id = 1;      // Device ID to query
  string page_token = 2;     // Pagination token (empty for first page)
  int32 page_size = 3;       // Number of readings per page (default: 50)
}
```

**Response**:
```protobuf
message GetSensorReadingByDeviceIDResponse {
  repeated SensorReading reading = 1;  // List of sensor readings
  string next_page_token = 2;          // Token for next page (empty if last page)
}
```

**Use Case**: Time-series charts, historical analysis

**Example (First Page)**:
```bash
grpcurl -plaintext -d '{"device_id": "device-001", "page_size": 10}' \
  localhost:50051 iot.SensorService/GetSensorReadingByDeviceID
```

**Response Example**:
```json
{
  "reading": [
    {
      "device_id": "device-001",
      "timestamp": "1697548800",
      "temperature": 22.5,
      "humidity": 45.2,
      "pressure": 1013.25,
      "battery_level": 87.5
    },
    {
      "device_id": "device-001",
      "timestamp": "1697548805",
      "temperature": 22.6,
      "humidity": 45.1,
      "pressure": 1013.22,
      "battery_level": 87.4
    }
  ],
  "next_page_token": "eyJ0aW1lc3RhbXAiOjE2OTc1NDg4MDV9"
}
```

**Example (Next Page)**:
```bash
grpcurl -plaintext -d '{
  "device_id": "device-001",
  "page_size": 10,
  "page_token": "eyJ0aW1lc3RhbXAiOjE2OTc1NDg4MDV9"
}' localhost:50051 iot.SensorService/GetSensorReadingByDeviceID
```

**Pagination Details**:
- Default `page_size`: 50 readings
- Maximum `page_size`: 1000 readings
- Readings sorted by timestamp (newest first)
- `next_page_token` is empty on last page
- Token format: Base64-encoded JSON (opaque to client)

**Performance**:
- Efficient cursor-based pagination
- No OFFSET (avoids performance degradation on large datasets)
- Indexed by device_id and timestamp

## Error Handling

### gRPC Status Codes

The API uses standard gRPC status codes:

| Code | Status | Description | Example |
|------|--------|-------------|---------|
| `0` | `OK` | Success | - |
| `3` | `INVALID_ARGUMENT` | Invalid parameter | Malformed device_id |
| `5` | `NOT_FOUND` | Resource not found | Device does not exist |
| `13` | `INTERNAL` | Server error | Database connection failure |
| `14` | `UNAVAILABLE` | Service unavailable | Database is down |

### Error Response Format

Errors are returned as gRPC status with details:

```json
{
  "error": {
    "code": 5,
    "message": "device not found",
    "details": [
      {
        "@type": "type.googleapis.com/google.rpc.ErrorInfo",
        "reason": "DEVICE_NOT_FOUND",
        "domain": "iot.example.com",
        "metadata": {
          "device_id": "device-999"
        }
      }
    ]
  }
}
```

### Common Errors

**Device Not Found**:
```bash
# Request
grpcurl -plaintext -d '{"device_id": "nonexistent"}' \
  localhost:50051 iot.SensorService/GetDevice

# Error
ERROR:
  Code: NotFound
  Message: device not found: nonexistent
```

**Invalid Device ID**:
```bash
# Request
grpcurl -plaintext -d '{"device_id": ""}' \
  localhost:50051 iot.SensorService/GetDevice

# Error
ERROR:
  Code: InvalidArgument
  Message: device_id is required
```

**Service Unavailable**:
```bash
# Error
ERROR:
  Code: Unavailable
  Message: database connection failed
```

## Examples

### Using grpcurl

Install grpcurl:
```bash
brew install grpcurl  # macOS
# or
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**List services**:
```bash
grpcurl -plaintext localhost:50051 list
```

Output:
```
grpc.reflection.v1alpha.ServerReflection
iot.SensorService
```

**List methods**:
```bash
grpcurl -plaintext localhost:50051 list iot.SensorService
```

Output:
```
iot.SensorService.GetAllDevice
iot.SensorService.GetDevice
iot.SensorService.GetSensorReadingByDeviceID
```

**Describe service**:
```bash
grpcurl -plaintext localhost:50051 describe iot.SensorService
```

**Get all devices**:
```bash
grpcurl -plaintext localhost:50051 iot.SensorService/GetAllDevice
```

**Get specific device**:
```bash
grpcurl -plaintext -d '{"device_id": "device-001"}' \
  localhost:50051 iot.SensorService/GetDevice
```

**Get sensor readings (paginated)**:
```bash
grpcurl -plaintext -d '{
  "device_id": "device-001",
  "page_size": 20
}' localhost:50051 iot.SensorService/GetSensorReadingByDeviceID
```

### Using Go Client

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "procodus.dev/demo-app/pkg/iot"
)

func main() {
    // Connect to backend
    conn, err := grpc.Dial("localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()

    // Create client
    client := iot.NewSensorServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get all devices
    allDevices, err := client.GetAllDevice(ctx, &iot.GetAllDeviceRequest{})
    if err != nil {
        log.Fatalf("GetAllDevice failed: %v", err)
    }
    log.Printf("Found %d devices", len(allDevices.GetDevice()))

    // Get specific device
    device, err := client.GetDevice(ctx, &iot.GetDeviceByIDRequest{
        DeviceId: "device-001",
    })
    if err != nil {
        log.Fatalf("GetDevice failed: %v", err)
    }
    log.Printf("Device: %s at %s", device.GetDevice().GetDeviceId(),
        device.GetDevice().GetLocation())

    // Get sensor readings (paginated)
    readings, err := client.GetSensorReadingByDeviceID(ctx,
        &iot.GetSensorReadingByDeviceIDRequest{
            DeviceId: "device-001",
            PageSize: 10,
        })
    if err != nil {
        log.Fatalf("GetSensorReadingByDeviceID failed: %v", err)
    }
    log.Printf("Found %d readings", len(readings.GetReading()))
    for _, r := range readings.GetReading() {
        log.Printf("  Temp: %.2f°C, Humidity: %.2f%%, Battery: %.2f%%",
            r.GetTemperature(), r.GetHumidity(), r.GetBatteryLevel())
    }

    // Pagination example
    pageToken := readings.GetNextPageToken()
    for pageToken != "" {
        readings, err = client.GetSensorReadingByDeviceID(ctx,
            &iot.GetSensorReadingByDeviceIDRequest{
                DeviceId:  "device-001",
                PageSize:  10,
                PageToken: pageToken,
            })
        if err != nil {
            log.Fatalf("Pagination failed: %v", err)
        }
        log.Printf("Next page: %d readings", len(readings.GetReading()))
        pageToken = readings.GetNextPageToken()
    }
}
```

### Using Python Client

Install dependencies:
```bash
pip install grpcio grpcio-tools
```

Generate Python stubs:
```bash
python -m grpc_tools.protoc \
  -I api/proto \
  --python_out=. \
  --grpc_python_out=. \
  api/proto/sensor.proto
```

Python client:
```python
import grpc
import sensor_pb2
import sensor_pb2_grpc

def main():
    # Connect to backend
    channel = grpc.insecure_channel('localhost:50051')
    client = sensor_pb2_grpc.SensorServiceStub(channel)

    # Get all devices
    response = client.GetAllDevice(sensor_pb2.GetAllDeviceRequest())
    print(f"Found {len(response.device)} devices")
    for device in response.device:
        print(f"  {device.device_id}: {device.location}")

    # Get specific device
    response = client.GetDevice(sensor_pb2.GetDeviceByIDRequest(
        device_id="device-001"
    ))
    device = response.device
    print(f"Device: {device.device_id} at {device.location}")

    # Get sensor readings
    response = client.GetSensorReadingByDeviceID(
        sensor_pb2.GetSensorReadingByDeviceIDRequest(
            device_id="device-001",
            page_size=10
        )
    )
    print(f"Found {len(response.reading)} readings")
    for reading in response.reading:
        print(f"  Temp: {reading.temperature}°C, "
              f"Humidity: {reading.humidity}%, "
              f"Battery: {reading.battery_level}%")

if __name__ == '__main__':
    main()
```

## Rate Limiting

**Current Status**: No rate limiting implemented (demo purposes)

**Production Recommendations**:
- Implement rate limiting per client
- Use token bucket algorithm
- Limit: 100 requests/minute per IP
- Return `RESOURCE_EXHAUSTED` (code 8) when limit exceeded

## Authentication

**Current Status**: No authentication (demo purposes)

**Production Recommendations**:
- Implement JWT-based authentication
- Add `Authorization` metadata to requests
- Validate token on each request
- Return `UNAUTHENTICATED` (code 16) for invalid tokens

Example with auth metadata:
```go
ctx := metadata.AppendToOutgoingContext(context.Background(),
    "authorization", "Bearer "+token)
resp, err := client.GetAllDevice(ctx, &iot.GetAllDeviceRequest{})
```

## Monitoring

### Metrics

The backend exposes Prometheus metrics at `http://localhost:9090/metrics`:

```promql
# gRPC request rate
rate(demo_app_backend_grpc_requests_total[5m])

# Request duration (p95)
histogram_quantile(0.95, rate(demo_app_backend_grpc_request_duration_seconds_bucket[5m]))

# Error rate
rate(demo_app_backend_grpc_requests_total{status="error"}[5m])

# In-flight requests
demo_app_backend_grpc_requests_in_flight
```

### Logging

All gRPC requests are logged with structured logging:

```json
{
  "time": "2025-10-17T10:30:45Z",
  "level": "INFO",
  "msg": "grpc request",
  "method": "/iot.SensorService/GetDevice",
  "device_id": "device-001",
  "duration_ms": 15,
  "status": "OK"
}
```

## Code Generation

Regenerate Go code from protobuf:

```bash
# Install protoc compiler
brew install protobuf  # macOS

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  api/proto/*.proto

# Copy to pkg/iot
cp api/proto/*.pb.go pkg/iot/
```

## Next Steps

- [Deployment Guide](./deployment.md) - Deploy the backend service
- [Monitoring Guide](./monitoring.md) - Set up API monitoring
- [Development Guide](./development.md) - Extend the API
- [Testing Guide](./testing.md) - Test API endpoints
