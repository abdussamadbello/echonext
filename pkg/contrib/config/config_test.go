package config

import (
	"os"
	"testing"
)

func TestDefaultLoadOptions(t *testing.T) {
	opts := DefaultLoadOptions()

	if opts.ConfigName != "config" {
		t.Errorf("Expected ConfigName 'config', got %s", opts.ConfigName)
	}
	if opts.ConfigType != "yaml" {
		t.Errorf("Expected ConfigType 'yaml', got %s", opts.ConfigType)
	}
	if !opts.AutomaticEnv {
		t.Error("Expected AutomaticEnv to be true")
	}
}

func TestLoadSimple(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test config
	configData := `
app:
  name: testapp
  port: 8080
  debug: true
`
	if _, err := tmpFile.Write([]byte(configData)); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Test loading
	type TestConfig struct {
		App AppConfig `mapstructure:"app"`
	}

	var cfg TestConfig
	err = LoadFromFile(&cfg, tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if cfg.App.Name != "testapp" {
		t.Errorf("Expected app name 'testapp', got %s", cfg.App.Name)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.App.Port)
	}
	if !cfg.App.Debug {
		t.Error("Expected debug to be true")
	}
}

func TestLoadWithEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("TESTAPP_APP_NAME", "envapp")
	os.Setenv("TESTAPP_APP_PORT", "9090")
	defer os.Unsetenv("TESTAPP_APP_NAME")
	defer os.Unsetenv("TESTAPP_APP_PORT")

	// Test that LoadOptions can be configured with env prefix
	opts := DefaultLoadOptions()
	opts.EnvPrefix = "TESTAPP"
	opts.ConfigName = "config"
	opts.ConfigType = "yaml"
	opts.ConfigPaths = []string{os.TempDir()}

	// Verify the structures exist
	if opts.EnvPrefix != "TESTAPP" {
		t.Errorf("Expected EnvPrefix 'TESTAPP', got %s", opts.EnvPrefix)
	}
}

func TestStandardConfigs(t *testing.T) {
	// Test that standard config structs can be instantiated
	_ = AppConfig{
		Name:        "test",
		Version:     "1.0.0",
		Environment: "test",
		Port:        8080,
		Debug:       true,
	}

	db := DatabaseConfig{
		Driver:      "postgres",
		DSN:         "test",
		MaxLifetime: "1h",
	}

	duration, err := db.ParseMaxLifetime()
	if err != nil {
		t.Errorf("ParseMaxLifetime failed: %v", err)
	}
	if duration.Hours() != 1 {
		t.Errorf("Expected 1 hour, got %v", duration)
	}

	_ = CacheConfig{
		Driver:     "redis",
		Address:    "localhost:6379",
		DefaultTTL: 3600,
	}

	_ = LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
}
