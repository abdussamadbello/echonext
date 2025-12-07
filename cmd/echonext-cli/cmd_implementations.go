package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Implementation functions for generate commands

func runGenerateDomain(cmd *cobra.Command, args []string) {
	name := args[0]
	domainName := strings.ToLower(name)

	fmt.Printf("🔨 Generating domain '%s'...\n", domainName)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project. Run this command from your project root.")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("❌ Failed to read module name: %v", err)
	}

	domainDir := filepath.Join("domain", domainName)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create domain directory: %v", err)
	}

	generator := &DomainGenerator{
		Name:   domainName,
		Module: module,
	}

	files := map[string]string{
		filepath.Join(domainDir, "model.go"):   generator.generateModel(),
		filepath.Join(domainDir, "service.go"): generator.generateService(),
		filepath.Join(domainDir, "handler.go"): generator.generateHandler(),
		filepath.Join(domainDir, "dto.go"):     generator.generateDTO(),
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			log.Fatalf("❌ Failed to write %s: %v", path, err)
		}
		fmt.Printf("  ✅ Created %s\n", path)
	}

	fmt.Printf("\n✨ Domain '%s' generated successfully!\n", domainName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Update domain/%s/model.go with your fields\n", domainName)
	fmt.Printf("  2. Implement business logic in domain/%s/service.go\n", domainName)
	fmt.Printf("  3. Register routes in your API server\n")
}

func runGenerateHandler(cmd *cobra.Command, args []string) {
	name := args[0]
	domainName := strings.ToLower(name)

	fmt.Printf("🔨 Generating handler for '%s'...\n", domainName)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("❌ Failed to read module name: %v", err)
	}

	domainDir := filepath.Join("domain", domainName)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create directory: %v", err)
	}

	generator := &DomainGenerator{Name: domainName, Module: module}
	handlerPath := filepath.Join(domainDir, "handler.go")

	if err := os.WriteFile(handlerPath, []byte(generator.generateHandler()), 0644); err != nil {
		log.Fatalf("❌ Failed to write handler: %v", err)
	}

	fmt.Printf("✅ Created %s\n", handlerPath)
}

func runGenerateService(cmd *cobra.Command, args []string) {
	name := args[0]
	domainName := strings.ToLower(name)

	fmt.Printf("🔨 Generating service for '%s'...\n", domainName)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("❌ Failed to read module name: %v", err)
	}

	domainDir := filepath.Join("domain", domainName)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create directory: %v", err)
	}

	generator := &DomainGenerator{Name: domainName, Module: module}
	servicePath := filepath.Join(domainDir, "service.go")

	if err := os.WriteFile(servicePath, []byte(generator.generateService()), 0644); err != nil {
		log.Fatalf("❌ Failed to write service: %v", err)
	}

	fmt.Printf("✅ Created %s\n", servicePath)
}

func runGenerateModel(cmd *cobra.Command, args []string) {
	name := args[0]
	domainName := strings.ToLower(name)

	fmt.Printf("🔨 Generating model for '%s'...\n", domainName)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("❌ Failed to read module name: %v", err)
	}

	domainDir := filepath.Join("domain", domainName)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create directory: %v", err)
	}

	generator := &DomainGenerator{Name: domainName, Module: module}
	modelPath := filepath.Join(domainDir, "model.go")

	if err := os.WriteFile(modelPath, []byte(generator.generateModel()), 0644); err != nil {
		log.Fatalf("❌ Failed to write model: %v", err)
	}

	fmt.Printf("✅ Created %s\n", modelPath)
}

func runGenerateDTO(cmd *cobra.Command, args []string) {
	name := args[0]
	domainName := strings.ToLower(name)

	fmt.Printf("🔨 Generating DTOs for '%s'...\n", domainName)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("❌ Failed to read module name: %v", err)
	}

	domainDir := filepath.Join("domain", domainName)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create directory: %v", err)
	}

	generator := &DomainGenerator{Name: domainName, Module: module}
	dtoPath := filepath.Join(domainDir, "dto.go")

	if err := os.WriteFile(dtoPath, []byte(generator.generateDTO()), 0644); err != nil {
		log.Fatalf("❌ Failed to write DTO: %v", err)
	}

	fmt.Printf("✅ Created %s\n", dtoPath)
}

func runGenerateMiddleware(cmd *cobra.Command, args []string) {
	name := args[0]
	middlewareName := strings.ToLower(name)

	fmt.Printf("🔨 Generating middleware '%s'...\n", middlewareName)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	middlewareDir := "internal/middleware"
	if err := os.MkdirAll(middlewareDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create middleware directory: %v", err)
	}

	generator := &MiddlewareGenerator{Name: middlewareName}
	middlewarePath := filepath.Join(middlewareDir, middlewareName+".go")

	if err := os.WriteFile(middlewarePath, []byte(generator.generate()), 0644); err != nil {
		log.Fatalf("❌ Failed to write middleware: %v", err)
	}

	fmt.Printf("✅ Created %s\n", middlewarePath)
}

// Database command implementations

func runDBInit(cmd *cobra.Command, args []string) {
	fmt.Println("🔨 Initializing database configuration...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	migrationsDir := "migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create migrations directory: %v", err)
	}
	fmt.Printf("  ✅ Created %s/\n", migrationsDir)

	sampleMigration := filepath.Join(migrationsDir, "001_initial_schema.sql")
	migrationContent := `-- Initial schema migration
-- Add your SQL migrations here

-- Example:
-- CREATE TABLE IF NOT EXISTS users (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     email VARCHAR(255) UNIQUE NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- CREATE INDEX idx_users_email ON users(email);
`

	if err := os.WriteFile(sampleMigration, []byte(migrationContent), 0644); err != nil {
		log.Fatalf("❌ Failed to write migration file: %v", err)
	}
	fmt.Printf("  ✅ Created %s\n", sampleMigration)

	seedsDir := filepath.Join("internal", "database", "seeds")
	if err := os.MkdirAll(seedsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create seeds directory: %v", err)
	}
	fmt.Printf("  ✅ Created %s/\n", seedsDir)

	sampleSeed := filepath.Join(seedsDir, "seeds.go")
	seedContent := `package seeds

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Run executes all seed functions
func Run(db *gorm.DB) error {
	log.Println("🌱 Seeding database...")

	// Add your seed functions here
	// if err := seedUsers(db); err != nil {
	// 	return err
	// }

	log.Println("✅ Database seeded successfully")
	return nil
}
`

	if err := os.WriteFile(sampleSeed, []byte(seedContent), 0644); err != nil {
		log.Fatalf("❌ Failed to write seed file: %v", err)
	}
	fmt.Printf("  ✅ Created %s\n", sampleSeed)

	fmt.Println("\n✨ Database configuration initialized!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Update your database configuration in configs/development.yaml")
	fmt.Println("  2. Add your migrations to migrations/")
	fmt.Println("  3. Run 'go run ./cmd/migration up' to apply migrations")
	fmt.Println("  4. Run 'go run ./cmd/migration seed' to seed data")
}

func runDBMigrate(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Running database migrations...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	migrationCmd := filepath.Join("cmd", "migration", "main.go")
	if _, err := os.Stat(migrationCmd); os.IsNotExist(err) {
		fmt.Println("⚠️  Migration command not found.")
		fmt.Println("\n💡 Use: go run ./cmd/migration up")
		return
	}

	fmt.Println("💡 Run: go run ./cmd/migration up")
	fmt.Println("\nAvailable migration commands:")
	fmt.Println("  go run ./cmd/migration up      - Run all pending migrations")
	fmt.Println("  go run ./cmd/migration down    - Rollback last migration")
	fmt.Println("  go run ./cmd/migration status  - Check migration status")
}

func runDBSeed(cmd *cobra.Command, args []string) {
	fmt.Println("🌱 Seeding database...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	migrationCmd := filepath.Join("cmd", "migration", "main.go")
	if _, err := os.Stat(migrationCmd); os.IsNotExist(err) {
		fmt.Println("⚠️  Migration command not found")
		fmt.Println("\n💡 Use: go run ./cmd/migration seed")
		return
	}

	fmt.Println("💡 Run: go run ./cmd/migration seed")
}

// Helper functions

func isEchoNextProject() bool {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return false
	}
	return true
}

func getModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}
