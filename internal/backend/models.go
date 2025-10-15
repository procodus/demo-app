// Package backend provides the backend service for consuming IoT data from RabbitMQ
// and persisting it to PostgreSQL.
package backend

import (
	"time"

	"gorm.io/gorm"
)

// SensorReading represents a sensor reading stored in the database.
// This model maps to the IoT sensor data received from RabbitMQ.
type SensorReading struct {
	Timestamp    time.Time `gorm:"index:idx_device_timestamp;index:idx_timestamp;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	DeviceID     string    `gorm:"index:idx_device_timestamp;not null"`
	Temperature  float64   `gorm:"not null"`
	Humidity     float64   `gorm:"not null"`
	Pressure     float64   `gorm:"not null"`
	BatteryLevel float64   `gorm:"not null"`
	ID           uint      `gorm:"primaryKey"`
}

// TableName specifies the table name for SensorReading model.
func (SensorReading) TableName() string {
	return "sensor_readings"
}

// IoTDevice represents an IoT device stored in the database.
type IoTDevice struct {
	DeviceID       string          `gorm:"uniqueIndex;not null"`
	Location       string          `gorm:"not null"`
	MACAddress     string          `gorm:"not null"`
	IPAddress      string          `gorm:"not null"`
	Firmware       string          `gorm:"not null"`
	LastSeen       time.Time       `gorm:"index:idx_last_seen"`
	CreatedAt      time.Time       `gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt  `gorm:"index"`
	Latitude       float32         `gorm:"not null"`
	Longitude      float32         `gorm:"not null"`
	ID             uint            `gorm:"primaryKey"`
	SensorReadings []SensorReading `gorm:"foreignKey:DeviceID;references:DeviceID"`
}

// TableName specifies the table name for IoTDevice model.
func (IoTDevice) TableName() string {
	return "iot_devices"
}
