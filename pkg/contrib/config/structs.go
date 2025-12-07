package config

import "time"

// Standard configuration structures that can be embedded in your application config

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	Debug       bool   `mapstructure:"debug"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver      string `mapstructure:"driver"`
	DSN         string `mapstructure:"dsn"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
	LogQueries  bool   `mapstructure:"log_queries"`
	MaxIdle     int    `mapstructure:"max_idle_conns"`
	MaxOpen     int    `mapstructure:"max_open_conns"`
	MaxLifetime string `mapstructure:"max_lifetime"` // Duration string like "1h"
}

// ParseMaxLifetime parses the MaxLifetime duration string
func (c *DatabaseConfig) ParseMaxLifetime() (time.Duration, error) {
	if c.MaxLifetime == "" {
		return time.Hour, nil // Default to 1 hour
	}
	return time.ParseDuration(c.MaxLifetime)
}

// CacheConfig holds cache configuration (Redis, in-memory, etc.)
type CacheConfig struct {
	Driver     string `mapstructure:"driver"`      // redis, memory, etc.
	Address    string `mapstructure:"address"`     // host:port
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	DefaultTTL int    `mapstructure:"default_ttl"` // seconds
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
	Output string `mapstructure:"output"` // stdout, stderr, file path
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	ReadTimeout     string `mapstructure:"read_timeout"`
	WriteTimeout    string `mapstructure:"write_timeout"`
	ShutdownTimeout string `mapstructure:"shutdown_timeout"`
}

// ParseReadTimeout parses the ReadTimeout duration string
func (c *ServerConfig) ParseReadTimeout() (time.Duration, error) {
	if c.ReadTimeout == "" {
		return 30 * time.Second, nil
	}
	return time.ParseDuration(c.ReadTimeout)
}

// ParseWriteTimeout parses the WriteTimeout duration string
func (c *ServerConfig) ParseWriteTimeout() (time.Duration, error) {
	if c.WriteTimeout == "" {
		return 30 * time.Second, nil
	}
	return time.ParseDuration(c.WriteTimeout)
}

// ParseShutdownTimeout parses the ShutdownTimeout duration string
func (c *ServerConfig) ParseShutdownTimeout() (time.Duration, error) {
	if c.ShutdownTimeout == "" {
		return 10 * time.Second, nil
	}
	return time.ParseDuration(c.ShutdownTimeout)
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
	Issuer          string `mapstructure:"issuer"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// Example of a complete application configuration that uses these structs
//
// Example usage:
//
//	type MyAppConfig struct {
//	    App      config.AppConfig      `mapstructure:"app"`
//	    Database config.DatabaseConfig `mapstructure:"database"`
//	    Cache    config.CacheConfig    `mapstructure:"cache"`
//	    Logger   config.LoggerConfig   `mapstructure:"logger"`
//	    Server   config.ServerConfig   `mapstructure:"server"`
//	    JWT      config.JWTConfig      `mapstructure:"jwt"`
//	    CORS     config.CORSConfig     `mapstructure:"cors"`
//	}
//
//	var cfg MyAppConfig
//	if err := config.LoadSimple(&cfg); err != nil {
//	    log.Fatal(err)
//	}
