package backend

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"procodus.dev/demo-app/pkg/iot"
)

// IoTServiceImpl implements the gRPC IoTService interface.
type IoTServiceImpl struct {
	iot.UnimplementedIoTServiceServer
	logger *slog.Logger
	db     *gorm.DB
}

// NewIoTService creates a new IoTServiceImpl instance.
func NewIoTService(logger *slog.Logger, db *gorm.DB) (*IoTServiceImpl, error) {
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	if db == nil {
		return nil, errors.New("database cannot be nil")
	}

	return &IoTServiceImpl{
		logger: logger,
		db:     db,
	}, nil
}

// GetAllDevice returns all IoT devices from the database.
func (s *IoTServiceImpl) GetAllDevice(ctx context.Context, _ *iot.GetAllDevicesRequest) (*iot.GetAllDevicesResponse, error) {
	s.logger.Info("GetAllDevice called")

	var devices []IoTDevice
	if err := s.db.WithContext(ctx).Find(&devices).Error; err != nil {
		s.logger.Error("failed to fetch devices", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch devices: %v", err)
	}

	// Convert database models to proto messages
	protoDevices := make([]*iot.IoTDevice, len(devices))
	for i, device := range devices {
		protoDevices[i] = &iot.IoTDevice{
			DeviceId:   device.DeviceID,
			Timestamp:  device.LastSeen.Unix(),
			Location:   device.Location,
			MacAddress: device.MACAddress,
			IpAddress:  device.IPAddress,
			Firmware:   device.Firmware,
			Latitude:   device.Latitude,
			Longitude:  device.Longitude,
		}
	}

	s.logger.Info("fetched devices", "count", len(devices))

	return &iot.GetAllDevicesResponse{
		Devices: protoDevices,
	}, nil
}

// GetDevice returns a specific IoT device by device ID.
func (s *IoTServiceImpl) GetDevice(ctx context.Context, req *iot.GetDeviceByIDRequest) (*iot.GetDeviceByIDResponse, error) {
	if req.GetDeviceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id cannot be empty")
	}

	s.logger.Info("GetDevice called", "device_id", req.GetDeviceId())

	var device IoTDevice
	if err := s.db.WithContext(ctx).Where("device_id = ?", req.GetDeviceId()).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("device not found", "device_id", req.GetDeviceId())
			return nil, status.Errorf(codes.NotFound, "device not found: %s", req.GetDeviceId())
		}
		s.logger.Error("failed to fetch device", "device_id", req.GetDeviceId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch device: %v", err)
	}

	protoDevice := &iot.IoTDevice{
		DeviceId:   device.DeviceID,
		Timestamp:  device.LastSeen.Unix(),
		Location:   device.Location,
		MacAddress: device.MACAddress,
		IpAddress:  device.IPAddress,
		Firmware:   device.Firmware,
		Latitude:   device.Latitude,
		Longitude:  device.Longitude,
	}

	s.logger.Info("fetched device", "device_id", req.GetDeviceId())

	return &iot.GetDeviceByIDResponse{
		Device: protoDevice,
	}, nil
}

// GetSensorReadingByDeviceID returns sensor readings for a specific device with pagination.
func (s *IoTServiceImpl) GetSensorReadingByDeviceID(ctx context.Context, req *iot.GetSensorReadingByDeviceIDRequest) (*iot.GetSensorReadingByDeviceIDResponse, error) {
	if req.GetDeviceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id cannot be empty")
	}

	s.logger.Info("GetSensorReadingByDeviceID called", "device_id", req.GetDeviceId())

	const pageSize = 100

	// Parse page token (offset)
	offset := 0
	if req.GetPageToken() != "" {
		var err error
		offset, err = strconv.Atoi(req.GetPageToken())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid page_token")
		}
	}

	// Query sensor readings with pagination
	var readings []SensorReading
	query := s.db.WithContext(ctx).
		Where("device_id = ?", req.GetDeviceId()).
		Order("timestamp DESC").
		Limit(pageSize + 1). // Fetch one extra to determine if there's a next page
		Offset(offset)

	if err := query.Find(&readings).Error; err != nil {
		s.logger.Error("failed to fetch sensor readings", "device_id", req.GetDeviceId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch sensor readings: %v", err)
	}

	// Determine if there's a next page
	hasNextPage := len(readings) > pageSize
	if hasNextPage {
		readings = readings[:pageSize]
	}

	// Convert database models to proto messages
	protoReadings := make([]*iot.SensorReading, len(readings))
	for i, reading := range readings {
		protoReadings[i] = &iot.SensorReading{
			DeviceId:     reading.DeviceID,
			Timestamp:    reading.Timestamp.Unix(),
			Temperature:  reading.Temperature,
			Humidity:     reading.Humidity,
			Pressure:     reading.Pressure,
			BatteryLevel: reading.BatteryLevel,
		}
	}

	// Generate next page token
	nextPageToken := ""
	if hasNextPage {
		nextPageToken = strconv.Itoa(offset + pageSize)
	}

	s.logger.Info("fetched sensor readings",
		"device_id", req.GetDeviceId(),
		"count", len(protoReadings),
		"has_next_page", hasNextPage,
	)

	return &iot.GetSensorReadingByDeviceIDResponse{
		Reading:       protoReadings,
		NextPageToken: nextPageToken,
	}, nil
}
