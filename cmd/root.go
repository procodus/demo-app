// Package main provides the unified CLI entry point for the demo-app services.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "demo-app",
		Short: "IoT data pipeline application",
		Long: `A demonstration IoT data pipeline application with three components:
- generator: Generates synthetic IoT sensor data
- backend: Processes and stores IoT data
- frontend: Web interface for viewing IoT data`,
		Version: "1.0.0",
	}
)

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml or /etc/demo-app/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")

	// Bind flags to viper
	if err := viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Fatalf("failed to bind log-level flag: %v", err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := InitConfig(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Log config file being used
	if viper.ConfigFileUsed() != "" {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}
