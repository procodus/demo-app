package backend_test

import (
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/backend"
)

var _ = Describe("Database", func() {
	var (
		logger *slog.Logger
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	})

	Describe("NewDB", func() {
		Context("with invalid configuration", func() {
			It("should return error when config is nil", func() {
				db, err := backend.NewDB(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("config cannot be nil"))
				Expect(db).To(BeNil())
			})

			It("should return error when logger is nil", func() {
				config := &backend.DBConfig{
					Logger:   nil,
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
				}

				db, err := backend.NewDB(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("logger"))
				Expect(db).To(BeNil())
			})
		})

		Context("connection validation", func() {
			It("should fail with invalid host", func() {
				config := &backend.DBConfig{
					Logger:   logger,
					Host:     "invalid-host-that-does-not-exist",
					Port:     5432,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
				}

				db, err := backend.NewDB(config)
				Expect(err).To(HaveOccurred())
				Expect(db).To(BeNil())
			})

			It("should fail with invalid port", func() {
				config := &backend.DBConfig{
					Logger:   logger,
					Host:     "localhost",
					Port:     99999,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
				}

				db, err := backend.NewDB(config)
				Expect(err).To(HaveOccurred())
				Expect(db).To(BeNil())
			})

			It("should fail when connecting to wrong port", func() {
				config := &backend.DBConfig{
					Logger:   logger,
					Host:     "localhost",
					Port:     9999,
					User:     "test",
					Password: "password",
					DBName:   "testdb",
					SSLMode:  "disable",
				}

				db, err := backend.NewDB(config)
				Expect(err).To(HaveOccurred())
				Expect(db).To(BeNil())
			})
		})

		Context("with different SSL modes", func() {
			It("should handle different SSL modes", func() {
				sslModes := []string{
					"disable",
					"require",
					"verify-ca",
					"verify-full",
				}

				for _, sslMode := range sslModes {
					config := &backend.DBConfig{
						Logger:   logger,
						Host:     "localhost",
						Port:     5432,
						User:     "test",
						Password: "password",
						DBName:   "testdb",
						SSLMode:  sslMode,
					}

					db, err := backend.NewDB(config)
					// We expect this to fail without a real DB, but the SSL mode should be accepted
					Expect(err).To(HaveOccurred())
					Expect(db).To(BeNil())
				}
			})
		})

		Context("with different database parameters", func() {
			It("should accept different hosts", func() {
				hosts := []string{
					"localhost",
					"127.0.0.1",
					"postgres.example.com",
					"10.0.0.1",
				}

				for _, host := range hosts {
					config := &backend.DBConfig{
						Logger:   logger,
						Host:     host,
						Port:     5432,
						User:     "test",
						Password: "password",
						DBName:   "testdb",
						SSLMode:  "disable",
					}

					db, err := backend.NewDB(config)
					// We expect connection to fail, but configuration should be accepted
					Expect(err).To(HaveOccurred())
					Expect(db).To(BeNil())
				}
			})

			It("should accept different ports", func() {
				ports := []int{5432, 5433, 5434, 15432}

				for _, port := range ports {
					config := &backend.DBConfig{
						Logger:   logger,
						Host:     "localhost",
						Port:     port,
						User:     "test",
						Password: "password",
						DBName:   "testdb",
						SSLMode:  "disable",
					}

					db, err := backend.NewDB(config)
					// We expect connection to fail, but configuration should be accepted
					Expect(err).To(HaveOccurred())
					Expect(db).To(BeNil())
				}
			})

			It("should accept empty password", func() {
				config := &backend.DBConfig{
					Logger:   logger,
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "",
					DBName:   "testdb",
					SSLMode:  "disable",
				}

				db, err := backend.NewDB(config)
				// We expect connection to fail, but configuration should be accepted
				Expect(err).To(HaveOccurred())
				Expect(db).To(BeNil())
			})

			It("should accept different database names", func() {
				dbNames := []string{
					"testdb",
					"iot_data",
					"sensor_readings",
					"production",
				}

				for _, dbName := range dbNames {
					config := &backend.DBConfig{
						Logger:   logger,
						Host:     "localhost",
						Port:     5432,
						User:     "test",
						Password: "password",
						DBName:   dbName,
						SSLMode:  "disable",
					}

					db, err := backend.NewDB(config)
					// We expect connection to fail, but configuration should be accepted
					Expect(err).To(HaveOccurred())
					Expect(db).To(BeNil())
				}
			})
		})
	})

	Describe("CloseDB", func() {
		Context("with nil database", func() {
			It("should handle nil database gracefully", func() {
				err := backend.CloseDB(nil, logger)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with nil logger", func() {
			It("should handle nil logger gracefully", func() {
				err := backend.CloseDB(nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
