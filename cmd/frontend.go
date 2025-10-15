package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"procodus.dev/demo-app/internal/frontend"
)

var frontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Run the frontend server",
	Long: `Run the frontend web server that:
- Serves the web UI for viewing IoT devices and sensor data
- Connects to backend gRPC API
- Uses htmx for dynamic updates
- Provides real-time data visualization`,
	RunE: runFrontend,
}

func init() {
	rootCmd.AddCommand(frontendCmd)

	// Frontend-specific flags
	frontendCmd.Flags().Int("http-port", 8080, "HTTP server port")
	frontendCmd.Flags().String("backend-addr", "localhost:9090", "Backend gRPC server address")

	// Bind flags to viper
	_ = viper.BindPFlag("frontend.http.port", frontendCmd.Flags().Lookup("http-port"))
	_ = viper.BindPFlag("frontend.backend.addr", frontendCmd.Flags().Lookup("backend-addr"))
}

func runFrontend(_ *cobra.Command, _ []string) error {
	logger := GetLogger()
	logger.Info("starting frontend service")

	// Create frontend configuration from viper
	config := &frontend.ServerConfig{
		Logger:          logger,
		HTTPPort:        viper.GetInt("frontend.http.port"),
		BackendGRPCAddr: viper.GetString("frontend.backend.addr"),
	}

	// Create and run server
	server, err := frontend.NewServer(config)
	if err != nil {
		logger.Error("failed to create frontend server", "error", err)
		return err
	}

	logger.Info("frontend server configuration",
		"http_port", config.HTTPPort,
		"backend_addr", config.BackendGRPCAddr,
	)

	if err := server.Run(context.Background()); err != nil {
		logger.Error("frontend server error", "error", err)
		return err
	}

	logger.Info("frontend server stopped")
	return nil
}
