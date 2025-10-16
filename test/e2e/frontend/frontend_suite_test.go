package frontend_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"procodus.dev/demo-app/internal/backend"
	"procodus.dev/demo-app/internal/frontend"
	"procodus.dev/demo-app/pkg/iot"
	e2econtainers "procodus.dev/demo-app/test/e2e/testcontainers"
)

func TestFrontend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Frontend E2E Suite")
}

var (
	// Infrastructure containers.
	pgContainer testcontainers.Container
	pgDSN       string

	// Backend components.
	testDB     *gorm.DB
	grpcServer *grpc.Server
	grpcAddr   string

	// Frontend server.
	frontendServer *frontend.Server
	frontendPort   int

	// Shared logger.
	logger *slog.Logger

	// Context for cleanup.
	ctx context.Context
)

var _ = BeforeSuite(func() {
	ctx = context.Background()

	// Create logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("setting up frontend E2E test suite")

	// Start PostgreSQL container
	logger.Info("starting PostgreSQL container")
	var err error
	pgContainer, pgDSN, err = e2econtainers.StartPostgres(ctx, &e2econtainers.PostgresConfig{
		User:          "frontendtest",
		Password:      "frontendtest",
		Database:      "frontend_e2e_db",
		ContainerName: "postgres-frontend-e2e",
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(pgContainer).NotTo(BeNil())
	Expect(pgDSN).NotTo(BeEmpty())

	logger.Info("PostgreSQL container started", "dsn", pgDSN)

	// Initialize database
	logger.Info("initializing database with DSN")
	db, err := gorm.Open(postgres.Open(pgDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(db).NotTo(BeNil())

	// Run migrations
	logger.Info("running database migrations")
	err = db.AutoMigrate(&backend.IoTDevice{}, &backend.SensorReading{})
	Expect(err).NotTo(HaveOccurred())

	testDB = db

	// Create gRPC service implementation
	logger.Info("creating gRPC service")
	iotService, err := backend.NewIoTService(logger, testDB)
	Expect(err).NotTo(HaveOccurred())

	// Start gRPC server
	logger.Info("starting gRPC server")
	listener, err := net.Listen("tcp", ":0") // Use random port
	Expect(err).NotTo(HaveOccurred())

	grpcAddr = listener.Addr().String()
	logger.Info("gRPC server listening", "address", grpcAddr)

	grpcServer = grpc.NewServer()
	iot.RegisterIoTServiceServer(grpcServer, iotService)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	// Wait for gRPC server to be ready
	time.Sleep(500 * time.Millisecond)

	// Create frontend server
	logger.Info("creating frontend server")
	frontendPort = 8180 // Fixed port for testing
	frontendCfg := &frontend.ServerConfig{
		BackendGRPCAddr: grpcAddr,
		HTTPPort:        frontendPort,
		Logger:          logger,
	}
	frontendServer, err = frontend.NewServer(frontendCfg)
	Expect(err).NotTo(HaveOccurred())

	// Start frontend server in background
	go func() {
		ctx := context.Background()
		if err := frontendServer.Run(ctx); err != nil {
			logger.Error("frontend server error", "error", err)
		}
	}()

	// Wait for frontend server to be ready
	time.Sleep(1 * time.Second)

	logger.Info("frontend E2E test suite setup complete")
})

var _ = AfterSuite(func() {
	logger.Info("tearing down frontend E2E test suite")

	// Shutdown frontend server
	if frontendServer != nil {
		logger.Info("shutting down frontend server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := frontendServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown frontend server", "error", err)
		}
	}

	// Stop gRPC server
	if grpcServer != nil {
		logger.Info("stopping gRPC server")
		grpcServer.GracefulStop()
	}

	// Close database
	if testDB != nil {
		logger.Info("closing database")
		if err := backend.CloseDB(testDB, logger); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}

	// Terminate PostgreSQL container
	if pgContainer != nil {
		logger.Info("terminating PostgreSQL container")
		if err := pgContainer.Terminate(ctx); err != nil {
			logger.Error("failed to terminate PostgreSQL container", "error", err)
		}
	}

	logger.Info("frontend E2E test suite teardown complete")
})

// Helper function to create test devices in the database.
func createTestDevice(_ context.Context, deviceID string) *iot.IoTDevice {
	device := &iot.IoTDevice{
		DeviceId:   deviceID,
		Timestamp:  time.Now().Unix(),
		Location:   "Test Location",
		MacAddress: "00:11:22:33:44:55",
		IpAddress:  "192.168.1.100",
		Firmware:   "v1.0.0",
		Latitude:   37.7749,
		Longitude:  -122.4194,
	}

	// Save to database via gRPC (simulating device creation)
	dbDevice := &backend.IoTDevice{
		DeviceID:   device.GetDeviceId(),
		Location:   device.GetLocation(),
		MACAddress: device.GetMacAddress(),
		IPAddress:  device.GetIpAddress(),
		Firmware:   device.GetFirmware(),
		Latitude:   device.GetLatitude(),
		Longitude:  device.GetLongitude(),
		LastSeen:   time.Unix(device.GetTimestamp(), 0),
	}

	err := testDB.Create(dbDevice).Error
	Expect(err).NotTo(HaveOccurred())

	return device
}

// Helper function to create test sensor reading.
func createTestSensorReading(_ context.Context, deviceID string, timestamp time.Time) *iot.SensorReading {
	reading := &iot.SensorReading{
		DeviceId:     deviceID,
		Timestamp:    timestamp.Unix(),
		Temperature:  25.5,
		Humidity:     65.0,
		Pressure:     1013.25,
		BatteryLevel: 85.0,
	}

	// Save to database
	dbReading := &backend.SensorReading{
		DeviceID:     reading.GetDeviceId(),
		Timestamp:    timestamp,
		Temperature:  reading.GetTemperature(),
		Humidity:     reading.GetHumidity(),
		Pressure:     reading.GetPressure(),
		BatteryLevel: reading.GetBatteryLevel(),
	}

	err := testDB.Create(dbReading).Error
	Expect(err).NotTo(HaveOccurred())

	return reading
}

// Helper function to get the base URL for the frontend.
func getFrontendURL(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", frontendPort, path)
}
