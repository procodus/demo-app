package backend

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig holds the database configuration.
type DBConfig struct {
	Logger   *slog.Logger
	Host     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Port     int
}

// NewDB creates a new database connection and runs migrations.
func NewDB(cfg *DBConfig) (*gorm.DB, error) {
	if cfg == nil {
		return nil, errors.New("database config cannot be nil")
	}

	if cfg.Logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	cfg.Logger.Info("connecting to database",
		"host", cfg.Host,
		"port", cfg.Port,
		"dbname", cfg.DBName,
	)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Use slog instead of GORM's logger
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database to verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	cfg.Logger.Info("database connection established")

	// Run migrations
	if err := runMigrations(db, cfg.Logger); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations runs database migrations for all models.
func runMigrations(db *gorm.DB, logger *slog.Logger) error {
	logger.Info("running database migrations")

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&SensorReading{},
		&IoTDevice{},
	); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	logger.Info("database migrations completed successfully")
	return nil
}

// CloseDB closes the database connection.
func CloseDB(db *gorm.DB, logger *slog.Logger) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	logger.Info("closing database connection")
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	logger.Info("database connection closed")
	return nil
}
