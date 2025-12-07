// Package config provides optional Viper integration helpers for EchoNext applications.
//
// This is a contrib package and is completely optional. You can use Viper directly
// if you prefer, or use these helpers to reduce boilerplate code.
//
// Features:
//   - Generic config loading with Load[T]
//   - Environment variable binding
//   - Hot reload with Watch[T]
//   - Standard config structures (AppConfig, DatabaseConfig, etc.)
//   - Multiple config file formats (YAML, JSON, TOML)
//
// Example usage:
//
//	import "github.com/abdussamadbello/echonext/pkg/contrib/config"
//
//	type MyConfig struct {
//	    App      config.AppConfig      `mapstructure:"app"`
//	    Database config.DatabaseConfig `mapstructure:"database"`
//	}
//
//	var cfg MyConfig
//	if err := config.LoadSimple(&cfg); err != nil {
//	    log.Fatal(err)
//	}
//
// With environment variables:
//
//	export MYAPP_APP_PORT=9090
//	export MYAPP_DATABASE_DSN="postgres://..."
//
//	if err := config.LoadWithEnv(&cfg, "MYAPP"); err != nil {
//	    log.Fatal(err)
//	}
//
// With hot reload:
//
//	watcher, err := config.Watch(&cfg, config.DefaultLoadOptions(), func(updated *MyConfig) {
//	    log.Println("Config reloaded!")
//	})
package config
