package frontend_test

import (
	"context"
	"log/slog"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"procodus.dev/demo-app/internal/frontend"
)

var _ = Describe("Frontend Server", func() {
	var (
		logger *slog.Logger
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	})

	Describe("NewServer", func() {
		Context("with valid configuration", func() {
			It("should create a server", func() {
				config := &frontend.ServerConfig{
					Logger:          logger,
					HTTPPort:        8080,
					BackendGRPCAddr: "localhost:9090",
				}

				server, err := frontend.NewServer(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(server).NotTo(BeNil())
			})

			It("should create server with different HTTP ports", func() {
				ports := []int{8080, 8081, 8082, 3000}

				for _, port := range ports {
					config := &frontend.ServerConfig{
						Logger:          logger,
						HTTPPort:        port,
						BackendGRPCAddr: "localhost:9090",
					}

					server, err := frontend.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})

			It("should create server with different backend addresses", func() {
				addresses := []string{
					"localhost:9090",
					"127.0.0.1:9090",
					"backend.example.com:9090",
					"backend:9090",
				}

				for _, addr := range addresses {
					config := &frontend.ServerConfig{
						Logger:          logger,
						HTTPPort:        8080,
						BackendGRPCAddr: addr,
					}

					server, err := frontend.NewServer(config)
					Expect(err).NotTo(HaveOccurred())
					Expect(server).NotTo(BeNil())
				}
			})
		})

		Context("with invalid configuration", func() {
			It("should return error when config is nil", func() {
				server, err := frontend.NewServer(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("config cannot be nil"))
				Expect(server).To(BeNil())
			})

			It("should return error when logger is nil", func() {
				config := &frontend.ServerConfig{
					Logger:          nil,
					HTTPPort:        8080,
					BackendGRPCAddr: "localhost:9090",
				}

				server, err := frontend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("logger"))
				Expect(server).To(BeNil())
			})

			It("should return error when HTTP port is zero", func() {
				config := &frontend.ServerConfig{
					Logger:          logger,
					HTTPPort:        0,
					BackendGRPCAddr: "localhost:9090",
				}

				server, err := frontend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("HTTP port"))
				Expect(server).To(BeNil())
			})

			It("should return error when HTTP port is negative", func() {
				config := &frontend.ServerConfig{
					Logger:          logger,
					HTTPPort:        -1,
					BackendGRPCAddr: "localhost:9090",
				}

				server, err := frontend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("HTTP port"))
				Expect(server).To(BeNil())
			})

			It("should return error when backend gRPC address is empty", func() {
				config := &frontend.ServerConfig{
					Logger:          logger,
					HTTPPort:        8080,
					BackendGRPCAddr: "",
				}

				server, err := frontend.NewServer(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("backend gRPC address"))
				Expect(server).To(BeNil())
			})
		})
	})

	Describe("Server Run", func() {
		Context("with context cancellation", func() {
			It("should shutdown when context is canceled", func() {
				config := &frontend.ServerConfig{
					Logger:          logger,
					HTTPPort:        8081,
					BackendGRPCAddr: "invalid:9090", // Invalid to prevent actual connection
				}

				server, err := frontend.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				// Should complete within reasonable time after context cancellation
				Eventually(done, 2*time.Second).Should(Receive())
			})

			It("should shutdown immediately with pre-canceled context", func() {
				config := &frontend.ServerConfig{
					Logger:          logger,
					HTTPPort:        8082,
					BackendGRPCAddr: "invalid:9090",
				}

				server, err := frontend.NewServer(config)
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel before Run

				done := make(chan error, 1)
				go func() {
					done <- server.Run(ctx)
				}()

				// Should complete very quickly
				Eventually(done, 1*time.Second).Should(Receive())
			})
		})
	})

	Describe("Server Shutdown", func() {
		It("should shutdown cleanly with no initialized components", func() {
			config := &frontend.ServerConfig{
				Logger:          logger,
				HTTPPort:        8083,
				BackendGRPCAddr: "localhost:9090",
			}

			server, err := frontend.NewServer(config)
			Expect(err).NotTo(HaveOccurred())

			err = server.Shutdown()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle multiple shutdown calls", func() {
			config := &frontend.ServerConfig{
				Logger:          logger,
				HTTPPort:        8084,
				BackendGRPCAddr: "localhost:9090",
			}

			server, err := frontend.NewServer(config)
			Expect(err).NotTo(HaveOccurred())

			err1 := server.Shutdown()
			Expect(err1).NotTo(HaveOccurred())

			err2 := server.Shutdown()
			Expect(err2).NotTo(HaveOccurred())
		})
	})

	Describe("Concurrent Server Creation", func() {
		It("should handle concurrent NewServer calls", func() {
			results := make(chan error, 5)

			for i := 0; i < 5; i++ {
				go func(index int) {
					config := &frontend.ServerConfig{
						Logger:          logger,
						HTTPPort:        8090 + index,
						BackendGRPCAddr: "localhost:9090",
					}

					_, err := frontend.NewServer(config)
					results <- err
				}(i)
			}

			for i := 0; i < 5; i++ {
				Eventually(results).Should(Receive(BeNil()))
			}
		})
	})
})
