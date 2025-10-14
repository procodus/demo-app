package logger_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/pkg/logger"
)

var _ = Describe("Logger", func() {
	Describe("New", func() {
		Context("with default config", func() {
			It("should create a non-nil logger", func() {
				log := logger.New(logger.DefaultConfig())
				Expect(log).NotTo(BeNil())
			})
		})

		Context("with nil config", func() {
			It("should create a non-nil logger with defaults", func() {
				log := logger.New(nil)
				Expect(log).NotTo(BeNil())
			})
		})

		Context("with custom level", func() {
			It("should create a logger with the specified level", func() {
				cfg := &logger.Config{
					Level:  slog.LevelDebug,
					Output: &bytes.Buffer{},
				}
				log := logger.New(cfg)
				Expect(log).NotTo(BeNil())
			})
		})

		Context("with add source enabled", func() {
			It("should create a logger that includes source information", func() {
				cfg := &logger.Config{
					Level:     slog.LevelInfo,
					Output:    &bytes.Buffer{},
					AddSource: true,
				}
				log := logger.New(cfg)
				Expect(log).NotTo(BeNil())
			})
		})
	})

	Describe("NewDefault", func() {
		It("should create a non-nil logger with default settings", func() {
			log := logger.NewDefault()
			Expect(log).NotTo(BeNil())
		})
	})

	Describe("NewWithLevel", func() {
		DescribeTable("should create loggers with different levels",
			func(level slog.Level) {
				log := logger.NewWithLevel(level)
				Expect(log).NotTo(BeNil())
			},
			Entry("debug level", slog.LevelDebug),
			Entry("info level", slog.LevelInfo),
			Entry("warn level", slog.LevelWarn),
			Entry("error level", slog.LevelError),
		)
	})

	Describe("ParseLevel", func() {
		DescribeTable("should parse level strings correctly",
			func(input string, expected slog.Level) {
				level := logger.ParseLevel(input)
				Expect(level).To(Equal(expected))
			},
			Entry("debug", "debug", slog.LevelDebug),
			Entry("info", "info", slog.LevelInfo),
			Entry("warn", "warn", slog.LevelWarn),
			Entry("warning", "warning", slog.LevelWarn),
			Entry("error", "error", slog.LevelError),
			Entry("invalid defaults to info", "invalid", slog.LevelInfo),
			Entry("empty string defaults to info", "", slog.LevelInfo),
		)
	})

	Describe("Logger Output Format", func() {
		var (
			buf *bytes.Buffer
			log *slog.Logger
		)

		BeforeEach(func() {
			buf = &bytes.Buffer{}
			cfg := &logger.Config{
				Level:  slog.LevelInfo,
				Output: buf,
			}
			log = logger.New(cfg)
		})

		It("should output valid JSON", func() {
			log.Info("test message")

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should include required fields", func() {
			log.Info("test message")

			var logEntry map[string]interface{}
			json.Unmarshal(buf.Bytes(), &logEntry)

			Expect(logEntry).To(HaveKey("time"))
			Expect(logEntry).To(HaveKey("level"))
			Expect(logEntry).To(HaveKey("msg"))
		})

		It("should include the correct message", func() {
			log.Info("test message")

			var logEntry map[string]interface{}
			json.Unmarshal(buf.Bytes(), &logEntry)

			Expect(logEntry["msg"]).To(Equal("test message"))
		})

		It("should include custom fields", func() {
			log.Info("test message", "key", "value", "count", 42)

			var logEntry map[string]interface{}
			json.Unmarshal(buf.Bytes(), &logEntry)

			Expect(logEntry).To(HaveKeyWithValue("key", "value"))
			Expect(logEntry).To(HaveKeyWithValue("count", float64(42)))
		})
	})

	Describe("Logger Levels", func() {
		DescribeTable("should respect log level filtering",
			func(level slog.Level, logFunc func(*slog.Logger), shouldAppear bool) {
				buf := &bytes.Buffer{}
				cfg := &logger.Config{
					Level:  level,
					Output: buf,
				}
				log := logger.New(cfg)

				logFunc(log)

				output := buf.String()
				hasOutput := len(strings.TrimSpace(output)) > 0
				Expect(hasOutput).To(Equal(shouldAppear))
			},
			Entry("debug logged when level is debug",
				slog.LevelDebug,
				func(l *slog.Logger) { l.Debug("debug message") },
				true,
			),
			Entry("debug not logged when level is info",
				slog.LevelInfo,
				func(l *slog.Logger) { l.Debug("debug message") },
				false,
			),
			Entry("info logged when level is info",
				slog.LevelInfo,
				func(l *slog.Logger) { l.Info("info message") },
				true,
			),
			Entry("warn logged when level is info",
				slog.LevelInfo,
				func(l *slog.Logger) { l.Warn("warn message") },
				true,
			),
			Entry("error logged when level is error",
				slog.LevelError,
				func(l *slog.Logger) { l.Error("error message") },
				true,
			),
			Entry("info not logged when level is error",
				slog.LevelError,
				func(l *slog.Logger) { l.Info("info message") },
				false,
			),
		)
	})

	Describe("WithContext", func() {
		var (
			buf *bytes.Buffer
			log *slog.Logger
		)

		BeforeEach(func() {
			buf = &bytes.Buffer{}
			cfg := &logger.Config{
				Level:  slog.LevelInfo,
				Output: buf,
			}
			log = logger.New(cfg)
		})

		It("should add context fields to log messages", func() {
			contextLogger := logger.WithContext(log,
				slog.String("service", "test-service"),
				slog.String("version", "1.0.0"),
			)
			contextLogger.Info("test message")

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			Expect(err).NotTo(HaveOccurred())

			Expect(logEntry).To(HaveKeyWithValue("service", "test-service"))
			Expect(logEntry).To(HaveKeyWithValue("version", "1.0.0"))
		})
	})

	Describe("DefaultConfig", func() {
		It("should return a non-nil config", func() {
			cfg := logger.DefaultConfig()
			Expect(cfg).NotTo(BeNil())
		})

		It("should have Info level by default", func() {
			cfg := logger.DefaultConfig()
			Expect(cfg.Level).To(Equal(slog.LevelInfo))
		})

		It("should have AddSource disabled by default", func() {
			cfg := logger.DefaultConfig()
			Expect(cfg.AddSource).To(BeFalse())
		})
	})
})
