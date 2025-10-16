package backend

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"procodus.dev/demo-app/pkg/iot"
	"procodus.dev/demo-app/pkg/metrics"
)

// IoTServiceImpl implements the gRPC IoTService interface.
type IoTServiceImpl struct {
	iot.UnimplementedIoTServiceServer
	logger  *slog.Logger
	db      *gorm.DB
	metrics *metrics.BackendMetrics // Optional metrics
}

// NewIoTService creates a new IoTServiceImpl instance.
func NewIoTService(logger *slog.Logger, db *gorm.DB, m *metrics.BackendMetrics) (*IoTServiceImpl, error) {
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	if db == nil {
		return nil, errors.New("database cannot be nil")
	}

	return &IoTServiceImpl{
		logger:  logger,
		db:      db,
		metrics: m,
	}, nil
}

// GetAllDevice returns all IoT devices from the database.
func (s *IoTServiceImpl) GetAllDevice(ctx context.Context, _ *iot.GetAllDevicesRequest) (*iot.GetAllDevicesResponse, error) {
	// Track in-flight requests
	if s.metrics != nil {
		s.metrics.GRPCRequestsInFlight.WithLabelValues("GetAllDevice").Inc()
		defer s.metrics.GRPCRequestsInFlight.WithLabelValues("GetAllDevice").Dec()
	}

	// Track duration
	var timer *prometheus.Timer
	if s.metrics != nil {
		timer = prometheus.NewTimer(s.metrics.GRPCRequestDuration.WithLabelValues("GetAllDevice"))
		defer timer.ObserveDuration()
	}

	s.logger.Info("GetAllDevice called")

	var devices []IoTDevice
	if err := s.db.WithContext(ctx).Find(&devices).Error; err != nil {
		s.logger.Error("failed to fetch devices", "error", err)

		// Track error
		if s.metrics != nil {
			s.metrics.GRPCRequestsTotal.WithLabelValues("GetAllDevice", "error").Inc()
		}

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

	// Track success
	if s.metrics != nil {
		s.metrics.GRPCRequestsTotal.WithLabelValues("GetAllDevice", "success").Inc()
	}

	return &iot.GetAllDevicesResponse{
		Devices: protoDevices,
	}, nil
}

// GetDevice returns a specific IoT device by device ID.
func (s *IoTServiceImpl) GetDevice(ctx context.Context, req *iot.GetDeviceByIDRequest) (*iot.GetDeviceByIDResponse, error) {
	// Track in-flight requests
	if s.metrics != nil {
		s.metrics.GRPCRequestsInFlight.WithLabelValues("GetDevice").Inc()
		defer s.metrics.GRPCRequestsInFlight.WithLabelValues("GetDevice").Dec()
	}

	// Track duration
	var timer *prometheus.Timer
	if s.metrics != nil {
		timer = prometheus.NewTimer(s.metrics.GRPCRequestDuration.WithLabelValues("GetDevice"))
		defer timer.ObserveDuration()
	}

	if req.GetDeviceId() == "" {
		// Track error
		if s.metrics != nil {
			s.metrics.GRPCRequestsTotal.WithLabelValues("GetDevice", "error").Inc()
		}
		return nil, status.Error(codes.InvalidArgument, "device_id cannot be empty")
	}

	s.logger.Info("GetDevice called", "device_id", req.GetDeviceId())

	var device IoTDevice
	if err := s.db.WithContext(ctx).Where("device_id = ?", req.GetDeviceId()).First(&device).Error; err != nil {
		// Track error
		if s.metrics != nil {
			s.metrics.GRPCRequestsTotal.WithLabelValues("GetDevice", "error").Inc()
		}

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

	// Track success
	if s.metrics != nil {
		s.metrics.GRPCRequestsTotal.WithLabelValues("GetDevice", "success").Inc()
	}

	return &iot.GetDeviceByIDResponse{
		Device: protoDevice,
	}, nil
}

// GetSensorReadingByDeviceID returns sensor readings for a specific device with pagination.
func (s *IoTServiceImpl) GetSensorReadingByDeviceID(ctx context.Context, req *iot.GetSensorReadingByDeviceIDRequest) (*iot.GetSensorReadingByDeviceIDResponse, error) {
	// Track in-flight requests
	if s.metrics != nil {
		s.metrics.GRPCRequestsInFlight.WithLabelValues("GetSensorReadingByDeviceID").Inc()
		defer s.metrics.GRPCRequestsInFlight.WithLabelValues("GetSensorReadingByDeviceID").Dec()
	}

	// Track duration
	var timer *prometheus.Timer
	if s.metrics != nil {
		timer = prometheus.NewTimer(s.metrics.GRPCRequestDuration.WithLabelValues("GetSensorReadingByDeviceID"))
		defer timer.ObserveDuration()
	}

	if req.GetDeviceId() == "" {
		// Track error
		if s.metrics != nil {
			s.metrics.GRPCRequestsTotal.WithLabelValues("GetSensorReadingByDeviceID", "error").Inc()
		}
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
			// Track error
			if s.metrics != nil {
				s.metrics.GRPCRequestsTotal.WithLabelValues("GetSensorReadingByDeviceID", "error").Inc()
			}
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

		// Track error
		if s.metrics != nil {
			s.metrics.GRPCRequestsTotal.WithLabelValues("GetSensorReadingByDeviceID", "error").Inc()
		}

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

	// Track success
	if s.metrics != nil {
		s.metrics.GRPCRequestsTotal.WithLabelValues("GetSensorReadingByDeviceID", "success").Inc()
	}

	return &iot.GetSensorReadingByDeviceIDResponse{
		Reading:       protoReadings,
		NextPageToken: nextPageToken,
	}, nil
}
