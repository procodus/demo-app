package backend_test

import (
	"context"
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/backend"
	"procodus.dev/demo-app/pkg/iot"
)

var _ = Describe("gRPC Service", func() {
	var (
		logger *slog.Logger
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	})

	Describe("NewIoTService", func() {
		Context("with valid configuration", func() {
			It("should create a service with valid logger and DB", func() {
				// Create a mock DB
				dbCfg := &backend.DBConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
					Logger:   logger,
				}
				db, dbErr := backend.NewDB(dbCfg)
				// We don't expect this to succeed without a real DB

				if db != nil && dbErr == nil {
					defer backend.CloseDB(db, logger)

					service, err := backend.NewIoTService(logger, db, nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(service).NotTo(BeNil())
				}
			})
		})

		Context("with invalid configuration", func() {
			It("should return error when logger is nil", func() {
				dbCfg := &backend.DBConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
					Logger:   logger,
				}
				db, _ := backend.NewDB(dbCfg)
				if db != nil {
					defer backend.CloseDB(db, logger)
				}

				service, err := backend.NewIoTService(nil, db, nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("logger"))
				Expect(service).To(BeNil())
			})

			It("should return error when database is nil", func() {
				service, err := backend.NewIoTService(logger, nil, nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database"))
				Expect(service).To(BeNil())
			})
		})
	})

	Describe("GetDevice", func() {
		Context("with invalid request", func() {
			It("should return error when device_id is empty", func() {
				dbCfg := &backend.DBConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
					Logger:   logger,
				}
				db, err := backend.NewDB(dbCfg)
				if err != nil || db == nil {
					Skip("skipping test: database not available")
				}
				defer backend.CloseDB(db, logger)

				service, err := backend.NewIoTService(logger, db, nil)
				Expect(err).NotTo(HaveOccurred())

				ctx := context.Background()
				req := &iot.GetDeviceByIDRequest{
					DeviceId: "",
				}

				resp, err := service.GetDevice(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})
	})

	Describe("GetSensorReadingByDeviceID", func() {
		Context("with invalid request", func() {
			It("should return error when device_id is empty", func() {
				dbCfg := &backend.DBConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
					Logger:   logger,
				}
				db, err := backend.NewDB(dbCfg)
				if err != nil || db == nil {
					Skip("skipping test: database not available")
				}
				defer backend.CloseDB(db, logger)

				service, err := backend.NewIoTService(logger, db, nil)
				Expect(err).NotTo(HaveOccurred())

				ctx := context.Background()
				req := &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId: "",
				}

				resp, err := service.GetSensorReadingByDeviceID(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})

			It("should return error when page_token is invalid", func() {
				dbCfg := &backend.DBConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
					Logger:   logger,
				}
				db, err := backend.NewDB(dbCfg)
				if err != nil || db == nil {
					Skip("skipping test: database not available")
				}
				defer backend.CloseDB(db, logger)

				service, err := backend.NewIoTService(logger, db, nil)
				Expect(err).NotTo(HaveOccurred())

				ctx := context.Background()
				req := &iot.GetSensorReadingByDeviceIDRequest{
					DeviceId:  "device-001",
					PageToken: "invalid-token",
				}

				resp, err := service.GetSensorReadingByDeviceID(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})
	})
})
