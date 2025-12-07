package database

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns 10, got %d", cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns != 100 {
		t.Errorf("Expected MaxOpenConns 100, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxLifetime != time.Hour {
		t.Errorf("Expected MaxLifetime 1h, got %v", cfg.MaxLifetime)
	}
}

func TestConnect(t *testing.T) {
	// Use SQLite in-memory for testing
	cfg := DefaultConfig()
	db, err := Connect(sqlite.Open(":memory:"), cfg)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Verify connection
	if err := Ping(db); err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// Clean up
	if err := Close(db); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestWithRetry(t *testing.T) {
	cfg := DefaultConfig()

	// Should succeed with SQLite
	db, err := WithRetry(sqlite.Open(":memory:"), cfg, 3, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("WithRetry failed: %v", err)
	}

	if err := Close(db); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestPing(t *testing.T) {
	cfg := DefaultConfig()
	db, err := Connect(sqlite.Open(":memory:"), cfg)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer Close(db)

	if err := Ping(db); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null"`
}

func TestAutoMigrate(t *testing.T) {
	cfg := DefaultConfig()
	db, err := Connect(sqlite.Open(":memory:"), cfg)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer Close(db)

	if err := AutoMigrate(db, &TestModel{}); err != nil {
		t.Errorf("AutoMigrate failed: %v", err)
	}

	// Verify table exists
	if !TableExists(db, &TestModel{}) {
		t.Error("Expected table to exist after AutoMigrate")
	}
}

func TestTableExists(t *testing.T) {
	cfg := DefaultConfig()
	db, err := Connect(sqlite.Open(":memory:"), cfg)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer Close(db)

	// Table should not exist initially
	if TableExists(db, &TestModel{}) {
		t.Error("Expected table to not exist initially")
	}

	// Create table
	AutoMigrate(db, &TestModel{})

	// Table should now exist
	if !TableExists(db, &TestModel{}) {
		t.Error("Expected table to exist after migration")
	}
}
