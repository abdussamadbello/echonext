// Package database provides optional GORM integration helpers for EchoNext applications.
// This is a contrib package and is completely optional - you can use GORM directly if preferred.
package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database connection configuration
type Config struct {
	Driver      string        // Database driver (postgres, mysql, sqlite, etc.)
	DSN         string        // Data Source Name connection string
	MaxIdleConns int          // Maximum idle connections in pool
	MaxOpenConns int          // Maximum open connections in pool
	MaxLifetime time.Duration // Maximum connection lifetime
	LogLevel    logger.LogLevel // GORM log level
}

// DefaultConfig returns sensible defaults for database configuration
func DefaultConfig() Config {
	return Config{
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour,
		LogLevel:     logger.Silent,
	}
}

// Connect establishes a database connection with the given dialector and config
func Connect(dialector gorm.Dialector, cfg Config) (*gorm.DB, error) {
	// Create GORM config with logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(cfg.LogLevel),
	}

	// Open database connection
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	return db, nil
}

// WithRetry wraps database connection with retry logic
func WithRetry(dialector gorm.Dialector, cfg Config, maxRetries int, retryDelay time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = Connect(dialector, cfg)
		if err == nil {
			return db, nil
		}

		log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

// Ping checks if the database connection is alive
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}
