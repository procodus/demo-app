# Database Schema

This document describes the PostgreSQL database schema used by the Demo App.

## Table of Contents

- [Overview](#overview)
- [Schema Diagram](#schema-diagram)
- [Tables](#tables)
- [Indexes](#indexes)
- [Relationships](#relationships)
- [Migrations](#migrations)
- [Queries](#queries)

## Overview

The application uses PostgreSQL 16+ with GORM ORM for database operations. The schema consists of two main tables with a one-to-many relationship.

**Database**: `iot_db`

**Tables**:
- `iot_devices` - Device metadata
- `sensor_readings` - Time-series sensor data

**Key Features**:
- Foreign key constraints for data integrity
- Indexes for query performance
- Soft deletes for devices
- Automatic timestamps

## Schema Diagram

```
┌─────────────────────────────────────┐
│         iot_devices                 │
├─────────────────────────────────────┤
│ PK │ id               SERIAL        │
│ UK │ device_id        VARCHAR(255)  │
│    │ location         VARCHAR(255)  │
│    │ mac_address      VARCHAR(255)  │
│    │ ip_address       VARCHAR(255)  │
│    │ firmware         VARCHAR(255)  │
│    │ latitude         FLOAT         │
│    │ longitude        FLOAT         │
│    │ last_seen        TIMESTAMP     │
│    │ created_at       TIMESTAMP     │
│    │ updated_at       TIMESTAMP     │
│    │ deleted_at       TIMESTAMP     │
└─────────────────────────────────────┘
                │
                │ 1:N
                │
                ▼
┌─────────────────────────────────────┐
│       sensor_readings               │
├─────────────────────────────────────┤
│ PK │ id               SERIAL        │
│ FK │ device_id        VARCHAR(255)  │────┐
│    │ timestamp        TIMESTAMP     │    │
│    │ temperature      DOUBLE        │    │ References
│    │ humidity         DOUBLE        │    │ iot_devices
│    │ pressure         DOUBLE        │    │ (device_id)
│    │ battery_level    DOUBLE        │    │
│    │ created_at       TIMESTAMP     │    │
│    │ updated_at       TIMESTAMP     │────┘
└─────────────────────────────────────┘
```

## Tables

### iot_devices

Stores IoT device metadata and configuration.

**SQL Schema**:
```sql
CREATE TABLE iot_devices (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL UNIQUE,
    location VARCHAR(255) NOT NULL,
    mac_address VARCHAR(255) NOT NULL,
    ip_address VARCHAR(255) NOT NULL,
    firmware VARCHAR(255) NOT NULL,
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_device_id ON iot_devices(device_id);
CREATE INDEX idx_last_seen ON iot_devices(last_seen);
CREATE INDEX idx_deleted_at ON iot_devices(deleted_at);
```

**GORM Model**:
```go
type IoTDevice struct {
    ID             uint            `gorm:"primaryKey"`
    DeviceID       string          `gorm:"uniqueIndex;not null"`
    Location       string          `gorm:"not null"`
    MACAddress     string          `gorm:"not null"`
    IPAddress      string          `gorm:"not null"`
    Firmware       string          `gorm:"not null"`
    Latitude       float32         `gorm:"not null"`
    Longitude      float32         `gorm:"not null"`
    LastSeen       time.Time       `gorm:"index:idx_last_seen"`
    CreatedAt      time.Time       `gorm:"autoCreateTime"`
    UpdatedAt      time.Time       `gorm:"autoUpdateTime"`
    DeletedAt      gorm.DeletedAt  `gorm:"index"`
    SensorReadings []SensorReading `gorm:"foreignKey:DeviceID;references:DeviceID"`
}
```

**Columns**:

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| `id` | SERIAL | No | Auto-incrementing primary key |
| `device_id` | VARCHAR(255) | No | Unique device identifier (e.g., "device-001") |
| `location` | VARCHAR(255) | No | Physical location (e.g., "San Francisco") |
| `mac_address` | VARCHAR(255) | No | Network MAC address |
| `ip_address` | VARCHAR(255) | No | IP address (IPv4 or IPv6) |
| `firmware` | VARCHAR(255) | No | Firmware version |
| `latitude` | FLOAT | No | GPS latitude (-90 to 90) |
| `longitude` | FLOAT | No | GPS longitude (-180 to 180) |
| `last_seen` | TIMESTAMP | No | Last time device sent data |
| `created_at` | TIMESTAMP | No | Record creation timestamp |
| `updated_at` | TIMESTAMP | No | Record last update timestamp |
| `deleted_at` | TIMESTAMP | Yes | Soft delete timestamp (NULL if active) |

**Constraints**:
- `device_id` must be unique across all devices
- Soft delete support (deleted_at NULL = active)
- All fields except deleted_at are required

**Typical Queries**:
```sql
-- Find all active devices
SELECT * FROM iot_devices WHERE deleted_at IS NULL;

-- Get device by ID
SELECT * FROM iot_devices WHERE device_id = 'device-001' AND deleted_at IS NULL;

-- Get devices seen in last hour
SELECT * FROM iot_devices
WHERE last_seen > NOW() - INTERVAL '1 hour'
AND deleted_at IS NULL;

-- Update last_seen timestamp
UPDATE iot_devices
SET last_seen = NOW(), updated_at = NOW()
WHERE device_id = 'device-001';
```

---

### sensor_readings

Stores time-series sensor data from IoT devices.

**SQL Schema**:
```sql
CREATE TABLE sensor_readings (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    temperature DOUBLE PRECISION NOT NULL,
    humidity DOUBLE PRECISION NOT NULL,
    pressure DOUBLE PRECISION NOT NULL,
    battery_level DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_sensor_readings_device
        FOREIGN KEY (device_id)
        REFERENCES iot_devices(device_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_device_timestamp ON sensor_readings(device_id, timestamp);
CREATE INDEX idx_timestamp ON sensor_readings(timestamp);
```

**GORM Model**:
```go
type SensorReading struct {
    ID           uint      `gorm:"primaryKey"`
    DeviceID     string    `gorm:"index:idx_device_timestamp;not null"`
    Timestamp    time.Time `gorm:"index:idx_device_timestamp;index:idx_timestamp;not null"`
    Temperature  float64   `gorm:"not null"`
    Humidity     float64   `gorm:"not null"`
    Pressure     float64   `gorm:"not null"`
    BatteryLevel float64   `gorm:"not null"`
    CreatedAt    time.Time `gorm:"autoCreateTime"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
```

**Columns**:

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| `id` | SERIAL | No | Auto-incrementing primary key |
| `device_id` | VARCHAR(255) | No | Foreign key to iot_devices |
| `timestamp` | TIMESTAMP | No | When the reading was taken |
| `temperature` | DOUBLE PRECISION | No | Temperature in Celsius (-40 to 85) |
| `humidity` | DOUBLE PRECISION | No | Relative humidity (0-100%) |
| `pressure` | DOUBLE PRECISION | No | Atmospheric pressure (300-1100 hPa) |
| `battery_level` | DOUBLE PRECISION | No | Battery level (0-100%) |
| `created_at` | TIMESTAMP | No | Record creation timestamp |
| `updated_at` | TIMESTAMP | No | Record last update timestamp |

**Constraints**:
- `device_id` must reference existing device
- Foreign key cascade delete (readings deleted when device deleted)
- All sensor values are required

**Typical Queries**:
```sql
-- Get latest readings for device
SELECT * FROM sensor_readings
WHERE device_id = 'device-001'
ORDER BY timestamp DESC
LIMIT 10;

-- Get readings in time range
SELECT * FROM sensor_readings
WHERE device_id = 'device-001'
AND timestamp BETWEEN '2025-10-17 00:00:00' AND '2025-10-17 23:59:59'
ORDER BY timestamp DESC;

-- Get average temperature per device
SELECT device_id, AVG(temperature) as avg_temp
FROM sensor_readings
GROUP BY device_id;

-- Get readings with low battery
SELECT device_id, battery_level, timestamp
FROM sensor_readings
WHERE battery_level < 20
ORDER BY timestamp DESC;

-- Count readings per device
SELECT device_id, COUNT(*) as reading_count
FROM sensor_readings
GROUP BY device_id
ORDER BY reading_count DESC;
```

## Indexes

### Primary Indexes

| Table | Column | Type | Purpose |
|-------|--------|------|---------|
| `iot_devices` | `id` | PRIMARY KEY | Auto-incrementing ID |
| `sensor_readings` | `id` | PRIMARY KEY | Auto-incrementing ID |

### Unique Indexes

| Table | Column | Purpose |
|-------|--------|---------|
| `iot_devices` | `device_id` | Ensure device uniqueness |

### Performance Indexes

| Table | Columns | Purpose |
|-------|---------|---------|
| `iot_devices` | `last_seen` | Query devices by last activity |
| `iot_devices` | `deleted_at` | Soft delete queries |
| `sensor_readings` | `device_id, timestamp` | Query readings by device and time |
| `sensor_readings` | `timestamp` | Time-range queries |

### Index Usage Examples

```sql
-- Uses idx_device_id (unique)
SELECT * FROM iot_devices WHERE device_id = 'device-001';

-- Uses idx_device_timestamp (composite)
SELECT * FROM sensor_readings
WHERE device_id = 'device-001'
AND timestamp > '2025-10-17 00:00:00';

-- Uses idx_timestamp
SELECT * FROM sensor_readings
WHERE timestamp BETWEEN '2025-10-17 00:00:00' AND '2025-10-17 23:59:59';

-- Uses idx_last_seen
SELECT * FROM iot_devices
WHERE last_seen > NOW() - INTERVAL '1 hour';
```

## Relationships

### One-to-Many: Device → Sensor Readings

**Relationship**: Each device can have multiple sensor readings

**Implementation**:
```sql
CONSTRAINT fk_sensor_readings_device
    FOREIGN KEY (device_id)
    REFERENCES iot_devices(device_id)
    ON DELETE CASCADE
```

**Behavior**:
- Sensor readings cannot exist without a device
- Deleting a device cascades to its readings
- Foreign key ensures referential integrity

**GORM Association**:
```go
// In IoTDevice model
SensorReadings []SensorReading `gorm:"foreignKey:DeviceID;references:DeviceID"`

// Query with preload
db.Preload("SensorReadings").First(&device, "device_id = ?", deviceID)
```

**Query Examples**:
```sql
-- Get device with all readings (JOIN)
SELECT d.*, r.*
FROM iot_devices d
LEFT JOIN sensor_readings r ON d.device_id = r.device_id
WHERE d.device_id = 'device-001';

-- Get devices with reading count
SELECT d.device_id, d.location, COUNT(r.id) as reading_count
FROM iot_devices d
LEFT JOIN sensor_readings r ON d.device_id = r.device_id
GROUP BY d.id, d.device_id, d.location;

-- Get devices with no readings
SELECT d.*
FROM iot_devices d
LEFT JOIN sensor_readings r ON d.device_id = r.device_id
WHERE r.id IS NULL;
```

## Migrations

### Automatic Migrations

The backend service automatically migrates the database on startup using GORM's AutoMigrate:

```go
func InitDatabase(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Auto-migrate models
    if err := db.AutoMigrate(&IoTDevice{}, &SensorReading{}); err != nil {
        return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
    }

    return db, nil
}
```

**Features**:
- Idempotent (safe to run multiple times)
- Creates tables if they don't exist
- Adds missing columns
- Creates indexes
- Does NOT delete columns or tables

### Manual Migrations

For production, consider using migration tools like:
- **golang-migrate**: https://github.com/golang-migrate/migrate
- **goose**: https://github.com/pressly/goose

Example migration with golang-migrate:

```sql
-- 001_create_tables.up.sql
CREATE TABLE iot_devices (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL UNIQUE,
    -- ... other columns
);

CREATE TABLE sensor_readings (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    -- ... other columns
    CONSTRAINT fk_sensor_readings_device
        FOREIGN KEY (device_id)
        REFERENCES iot_devices(device_id)
        ON DELETE CASCADE
);

-- 001_create_tables.down.sql
DROP TABLE IF EXISTS sensor_readings;
DROP TABLE IF EXISTS iot_devices;
```

Run migrations:
```bash
migrate -path migrations -database "postgres://user:pass@localhost:5432/iot_db?sslmode=disable" up
```

## Queries

### Common Query Patterns

**Get Device with Latest Reading**:
```sql
SELECT d.*, r.temperature, r.humidity, r.timestamp
FROM iot_devices d
LEFT JOIN LATERAL (
    SELECT temperature, humidity, timestamp
    FROM sensor_readings
    WHERE device_id = d.device_id
    ORDER BY timestamp DESC
    LIMIT 1
) r ON true
WHERE d.deleted_at IS NULL;
```

**Get Hourly Average Temperature**:
```sql
SELECT
    device_id,
    DATE_TRUNC('hour', timestamp) as hour,
    AVG(temperature) as avg_temp,
    MIN(temperature) as min_temp,
    MAX(temperature) as max_temp
FROM sensor_readings
WHERE device_id = 'device-001'
AND timestamp > NOW() - INTERVAL '24 hours'
GROUP BY device_id, hour
ORDER BY hour DESC;
```

**Get Devices with Anomalies**:
```sql
SELECT device_id, temperature, humidity, timestamp
FROM sensor_readings
WHERE (temperature < -20 OR temperature > 70)
   OR (humidity < 10 OR humidity > 95)
   OR (battery_level < 10)
ORDER BY timestamp DESC;
```

**Pagination (Cursor-Based)**:
```sql
-- First page
SELECT * FROM sensor_readings
WHERE device_id = 'device-001'
ORDER BY timestamp DESC
LIMIT 50;

-- Next page (using last timestamp as cursor)
SELECT * FROM sensor_readings
WHERE device_id = 'device-001'
AND timestamp < '2025-10-17 10:30:00'
ORDER BY timestamp DESC
LIMIT 50;
```

## Performance Tips

### Query Optimization

1. **Use Indexes**: Ensure queries use appropriate indexes
   ```sql
   EXPLAIN ANALYZE SELECT * FROM sensor_readings WHERE device_id = 'device-001';
   ```

2. **Limit Result Sets**: Always use LIMIT for large tables
   ```sql
   SELECT * FROM sensor_readings LIMIT 100;
   ```

3. **Avoid SELECT ***: Select only needed columns
   ```sql
   SELECT device_id, temperature, timestamp FROM sensor_readings;
   ```

4. **Use Composite Indexes**: For multi-column WHERE clauses
   ```sql
   -- Uses idx_device_timestamp
   WHERE device_id = 'device-001' AND timestamp > '2025-10-17'
   ```

### Database Maintenance

**Vacuum**: Clean up dead tuples
```sql
VACUUM ANALYZE sensor_readings;
VACUUM ANALYZE iot_devices;
```

**Reindex**: Rebuild indexes if fragmented
```sql
REINDEX TABLE sensor_readings;
REINDEX TABLE iot_devices;
```

**Analyze**: Update statistics for query planner
```sql
ANALYZE sensor_readings;
ANALYZE iot_devices;
```

## Backup and Restore

### Backup

```bash
# Full database backup
pg_dump -h localhost -U postgres -d iot_db -F c -f iot_db_backup.dump

# Schema only
pg_dump -h localhost -U postgres -d iot_db -s > schema.sql

# Data only
pg_dump -h localhost -U postgres -d iot_db -a > data.sql

# Single table
pg_dump -h localhost -U postgres -d iot_db -t sensor_readings > sensor_readings.sql
```

### Restore

```bash
# Restore from custom format
pg_restore -h localhost -U postgres -d iot_db -c iot_db_backup.dump

# Restore from SQL file
psql -h localhost -U postgres -d iot_db < schema.sql
```

## Next Steps

- [API Reference](./api.md) - Query via gRPC API
- [Architecture](./architecture.md) - System design
- [Development Guide](./development.md) - Work with the database
- [Troubleshooting](./troubleshooting.md) - Fix database issues
