package main

import (
	"fmt"
	"strings"
)

// generateAPIMain creates the main.go file for the API server
func (g *ProjectGenerator) generateAPIMain() string {
	return fmt.Sprintf(`package main

import (
	"log"

	"%s/domain/user"
	"%s/internal/config"
	"%s/internal/database"
	"%s/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %%v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %%v", err)
	}

	// Create server
	app := server.New(cfg)

	// Register domains
	userHandler := user.NewHandler(user.NewService(user.NewRepository(db)))
	userHandler.RegisterRoutes(app)

	// Start server
	log.Printf("🚀 API server starting on port %%d", cfg.App.Port)
	log.Printf("📖 API docs available at http://localhost:%%d/api/docs", cfg.App.Port)
	
	if err := app.Start(fmt.Sprintf(":%%d", cfg.App.Port)); err != nil {
		log.Fatalf("Failed to start server: %%v", err)
	}
}
`, g.Module, g.Module, g.Module, g.Module)
}

// generateWorkerMain creates the main.go file for the background worker
func (g *ProjectGenerator) generateWorkerMain() string {
	return fmt.Sprintf(`package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"%s/internal/config"
	"%s/internal/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %%v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %%v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker
	worker := NewWorker(db, cfg)
	
	go func() {
		log.Println("🔄 Background worker starting...")
		if err := worker.Start(ctx); err != nil {
			log.Printf("Worker error: %%v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("🛑 Shutting down worker...")
	
	// Graceful shutdown
	cancel()
	time.Sleep(5 * time.Second)
	log.Println("✅ Worker stopped")
}

// Worker handles background jobs
type Worker struct {
	db  interface{} // Database connection
	cfg *config.Config
}

// NewWorker creates a new worker instance
func NewWorker(db interface{}, cfg *config.Config) *Worker {
	return &Worker{
		db:  db,
		cfg: cfg,
	}
}

// Start begins processing background jobs
func (w *Worker) Start(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// TODO: Process background jobs
			log.Println("🔄 Processing background jobs...")
			
			// Example job processing
			if err := w.processEmailQueue(); err != nil {
				log.Printf("Failed to process email queue: %%v", err)
			}
		}
	}
}

// processEmailQueue processes pending email jobs
func (w *Worker) processEmailQueue() error {
	// TODO: Implement email queue processing
	log.Println("📧 Processing email queue...")
	return nil
}
`, g.Module, g.Module)
}

// generateCLIMain creates the main.go file for the CLI tool
func (g *ProjectGenerator) generateCLIMain() string {
	return fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"%s/internal/config"
	"%s/internal/database"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "%s",
		Short: "%s CLI tool",
		Long:  "Command-line interface for %s application",
	}

	// Add commands
	rootCmd.AddCommand(
		newMigrateCmd(),
		newSeedCmd(),
		newUserCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v\n", err)
		os.Exit(1)
	}
}

// newMigrateCmd returns the migrate command
func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %%w", err)
			}

			db, err := database.Connect(cfg.Database)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %%w", err)
			}

			fmt.Println("Running database migrations...")
			return database.RunMigrations(db)
		},
	}
}

// newSeedCmd returns the seed command
func newSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed database with sample data",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %%w", err)
			}

			db, err := database.Connect(cfg.Database)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %%w", err)
			}

			fmt.Println("Seeding database...")
			return database.SeedData(db)
		},
	}
}

// newUserCmd returns the user management commands
func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User management commands",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "create [email] [name]",
			Short: "Create a new user",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				email, name := args[0], args[1]
				fmt.Printf("Creating user: %%s (%%s)\n", name, email)
				// TODO: Implement user creation
				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all users",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("Listing users...")
				// TODO: Implement user listing
				return nil
			},
		},
	)

	return cmd
}
`, g.Module, g.Module, g.Name, strings.ToTitle(g.Name), g.Name)
}

// generateMigrationMain creates the main.go file for database migrations
func (g *ProjectGenerator) generateMigrationMain() string {
	return fmt.Sprintf(`package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"%s/internal/config"
	"%s/internal/database"
)

func main() {
	var action string
	flag.StringVar(&action, "action", "up", "Migration action: up, down, or status")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %%v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %%v", err)
	}

	// Execute migration based on action
	switch action {
	case "up":
		fmt.Println("⬆️  Running migrations...")
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("Migration failed: %%v", err)
		}
		fmt.Println("✅ Migrations completed successfully")
		
	case "down":
		fmt.Println("⬇️  Rolling back migrations...")
		if err := database.RollbackMigrations(db); err != nil {
			log.Fatalf("Rollback failed: %%v", err)
		}
		fmt.Println("✅ Rollback completed successfully")
		
	case "status":
		fmt.Println("📊 Migration status:")
		if err := database.MigrationStatus(db); err != nil {
			log.Fatalf("Failed to get migration status: %%v", err)
		}
		
	default:
		fmt.Printf("Unknown action: %%s\n", action)
		fmt.Println("Available actions: up, down, status")
		os.Exit(1)
	}
}
`, g.Module, g.Module)
}

// generateConfig creates the configuration package
func (g *ProjectGenerator) generateConfig() string {
	return fmt.Sprintf(`package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig      ` + "`mapstructure:\"app\"`" + `
	Database DatabaseConfig ` + "`mapstructure:\"database\"`" + `
	Cache    CacheConfig    ` + "`mapstructure:\"cache\"`" + `
	Logger   LoggerConfig   ` + "`mapstructure:\"logger\"`" + `
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string ` + "`mapstructure:\"name\"`" + `
	Version     string ` + "`mapstructure:\"version\"`" + `
	Environment string ` + "`mapstructure:\"environment\"`" + `
	Port        int    ` + "`mapstructure:\"port\"`" + `
	Debug       bool   ` + "`mapstructure:\"debug\"`" + `
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver      string ` + "`mapstructure:\"driver\"`" + `
	DSN         string ` + "`mapstructure:\"dsn\"`" + `
	AutoMigrate bool   ` + "`mapstructure:\"auto_migrate\"`" + `
	LogQueries  bool   ` + "`mapstructure:\"log_queries\"`" + `
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Driver  string ` + "`mapstructure:\"driver\"`" + `
	Address string ` + "`mapstructure:\"address\"`" + `
	TTL     int    ` + "`mapstructure:\"default_ttl\"`" + `
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string ` + "`mapstructure:\"level\"`" + `
	Format string ` + "`mapstructure:\"format\"`" + `
	Output string ` + "`mapstructure:\"output\"`" + `
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	// Set config file properties
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Environment variables
	viper.SetEnvPrefix("%s")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "%s")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.debug", true)

	// Database defaults
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.log_queries", false)

	// Cache defaults
	viper.SetDefault("cache.driver", "memory")
	viper.SetDefault("cache.default_ttl", 3600)

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")
}
`, strings.ToUpper(g.Name), g.Name)
}

// Continue with more template methods...
func (g *ProjectGenerator) generateDatabase() string {
	return fmt.Sprintf(`package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"%s/domain/user"
	"time"
	"log"
	"os"
)

// Connect creates a database connection
func Connect(cfg DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM logger
	logLevel := logger.Silent
	if cfg.LogQueries {
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %%w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %%w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate if enabled
	if cfg.AutoMigrate {
		if err := RunMigrations(db); err != nil {
			return nil, fmt.Errorf("auto migration failed: %%w", err)
		}
	}

	return db, nil
}

// RunMigrations runs database migrations
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&user.User{},
		// Add other models here
	)
}

// RollbackMigrations rolls back database migrations
func RollbackMigrations(db *gorm.DB) error {
	// Note: GORM doesn't support automatic rollback
	// You'll need to implement this manually or use a migration tool
	fmt.Println("⚠️  Manual rollback required - GORM doesn't support automatic rollback")
	return nil
}

// MigrationStatus shows current migration status
func MigrationStatus(db *gorm.DB) error {
	// Check if tables exist
	tables := []string{"users"}
	
	for _, table := range tables {
		if db.Migrator().HasTable(table) {
			fmt.Printf("✅ Table '%%s' exists\n", table)
		} else {
			fmt.Printf("❌ Table '%%s' missing\n", table)
		}
	}
	
	return nil
}

// SeedData seeds the database with sample data
func SeedData(db *gorm.DB) error {
	// Create sample users
	users := []user.User{
		{
			Name:  "Admin User",
			Email: "admin@%s.com",
		},
		{
			Name:  "Test User",
			Email: "test@%s.com",
		},
	}

	for _, u := range users {
		// Check if user already exists
		var existing user.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			fmt.Printf("👤 User '%%s' already exists, skipping\n", u.Email)
			continue
		}

		if err := db.Create(&u).Error; err != nil {
			return fmt.Errorf("failed to create user %%s: %%w", u.Email, err)
		}
		
		fmt.Printf("👤 Created user: %%s (%%s)\n", u.Name, u.Email)
	}

	return nil
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Driver      string
	DSN         string
	AutoMigrate bool
	LogQueries  bool
}
`, g.Module, g.Name, g.Name)
}

// Generate more template files...
func (g *ProjectGenerator) generateServer() string {
	return fmt.Sprintf(`package server

import (
	"github.com/abdussamadbello/echonext"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// New creates a new EchoNext server with middleware
func New(cfg *Config) *echonext.App {
	// Create EchoNext app
	app := echonext.New()

	// Configure API info
	app.SetInfo(
		cfg.App.Name,
		cfg.App.Version,
		"API built with EchoNext framework",
	)

	// Add middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORS())

	// Add health check
	app.GET("/health", func(c echo.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			"status":      "healthy",
			"service":     cfg.App.Name,
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
		}, nil
	}, echonext.Route{
		Summary: "Health check",
		Tags:    []string{"System"},
	})

	// Serve OpenAPI documentation
	app.ServeOpenAPISpec("/api/openapi.json")
	app.ServeSwaggerUI("/api/docs", "/api/openapi.json")

	return app
}

// Config represents server configuration
type Config struct {
	App AppConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Port        int
	Debug       bool
}
`, g.Module)
}