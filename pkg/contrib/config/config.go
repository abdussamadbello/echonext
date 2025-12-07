// Package config provides optional Viper integration helpers for EchoNext applications.
// This is a contrib package and is completely optional - you can use Viper directly if preferred.
package config

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// LoadOptions configures how configuration should be loaded
type LoadOptions struct {
	ConfigName     string   // Config file name (without extension)
	ConfigType     string   // Config file type (yaml, json, toml, etc.)
	ConfigPaths    []string // Paths to search for config file
	EnvPrefix      string   // Prefix for environment variables
	EnvKeyReplacer *strings.Replacer // Replace characters in keys for env vars
	AutomaticEnv   bool     // Automatically bind environment variables
	WatchConfig    bool     // Watch config file for changes
}

// DefaultLoadOptions returns sensible defaults for configuration loading
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		ConfigName:     "config",
		ConfigType:     "yaml",
		ConfigPaths:    []string{"./configs", "."},
		EnvKeyReplacer: strings.NewReplacer(".", "_"),
		AutomaticEnv:   true,
		WatchConfig:    false,
	}
}

// Load reads configuration from file and environment variables into the target struct
func Load[T any](target *T, opts LoadOptions) error {
	v := viper.New()

	// Configure file properties
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}

	// Configure environment variables
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
	}
	if opts.EnvKeyReplacer != nil {
		v.SetEnvKeyReplacer(opts.EnvKeyReplacer)
	}
	if opts.AutomaticEnv {
		v.AutomaticEnv()
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, continue with env vars only
	}

	// Watch for changes if enabled
	if opts.WatchConfig {
		v.WatchConfig()
	}

	// Unmarshal into target struct
	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// LoadSimple is a convenience function that loads config with default options
func LoadSimple[T any](target *T) error {
	return Load(target, DefaultLoadOptions())
}

// LoadWithEnv loads config with a custom environment variable prefix
func LoadWithEnv[T any](target *T, envPrefix string) error {
	opts := DefaultLoadOptions()
	opts.EnvPrefix = envPrefix
	return Load(target, opts)
}

// LoadFromFile loads config from a specific file path
func LoadFromFile[T any](target *T, filePath string) error {
	v := viper.New()
	v.SetConfigFile(filePath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Watcher watches for configuration changes and calls the callback when changes occur
type Watcher[T any] struct {
	viper    *viper.Viper
	target   *T
	callback func(*T)
}

// Watch creates a new configuration watcher
func Watch[T any](target *T, opts LoadOptions, callback func(*T)) (*Watcher[T], error) {
	v := viper.New()

	// Configure file properties
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}

	// Configure environment variables
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
	}
	if opts.EnvKeyReplacer != nil {
		v.SetEnvKeyReplacer(opts.EnvKeyReplacer)
	}
	if opts.AutomaticEnv {
		v.AutomaticEnv()
	}

	// Read initial config
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := v.Unmarshal(target); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	watcher := &Watcher[T]{
		viper:    v,
		target:   target,
		callback: callback,
	}

	// Watch for changes
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := v.Unmarshal(target); err != nil {
			// Log error but don't crash
			fmt.Printf("Error reloading config: %v\n", err)
			return
		}
		callback(target)
	})

	v.WatchConfig()

	return watcher, nil
}

// Stop stops watching for configuration changes
func (w *Watcher[T]) Stop() {
	// Viper doesn't provide a way to stop watching, so this is a no-op
	// but we provide it for API consistency
}
