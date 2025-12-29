package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/abdussamadbello/echonext/cmd/echonext-cli/generator"
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

func runGenerateOtel(cmd *cobra.Command, args []string) {
	fmt.Println("🔨 Generating OpenTelemetry setup...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project. Run this command from your project root.")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("❌ Failed to read module name: %v", err)
	}

	// Extract project name from module (last segment)
	parts := strings.Split(module, "/")
	projectName := parts[len(parts)-1]

	otelDir := filepath.Join("internal", "otel")
	if err := os.MkdirAll(otelDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create otel directory: %v", err)
	}

	// Use template generator
	gen := &ProjectGenerator{
		Name:     projectName,
		Module:   module,
		Template: "standard",
	}

	// Initialize the template generator
	if err := gen.initTemplateGenerator(); err != nil {
		log.Fatalf("❌ Failed to initialize template generator: %v", err)
	}

	otelPath := filepath.Join(otelDir, "otel.go")
	content := gen.generateOTEL()

	if err := os.WriteFile(otelPath, []byte(content), 0644); err != nil {
		log.Fatalf("❌ Failed to write otel.go: %v", err)
	}
	fmt.Printf("  ✅ Created %s\n", otelPath)

	fmt.Println("\n✨ OpenTelemetry setup generated!")
	fmt.Println("\n📝 Usage in your main.go:")
	fmt.Println("")
	fmt.Printf("  import \"%s/internal/otel\"\n", module)
	fmt.Println("  import \"github.com/abdussamadbello/echonext/pkg/contrib/middleware\"")
	fmt.Println("")
	fmt.Println("  func main() {")
	fmt.Println("      ctx := context.Background()")
	fmt.Println("      shutdown := otel.MustInit(ctx, otel.DefaultConfig())")
	fmt.Println("      defer shutdown()")
	fmt.Println("")
	fmt.Println("      app := echonext.New()")
	fmt.Println("      app.Use(middleware.OTELMiddleware(\"your-service\"))")
	fmt.Println("      // ...")
	fmt.Println("  }")
	fmt.Println("")
	fmt.Println("🌍 Environment variables:")
	fmt.Println("  OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317")
	fmt.Println("  OTEL_SERVICE_NAME=your-service")
}

// Database command implementations

func runDBInit(cmd *cobra.Command, args []string) {
	fmt.Println("🔨 Initializing Atlas migration setup...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("❌ Failed to read module name: %v", err)
	}

	// Extract project name from module
	parts := strings.Split(module, "/")
	projectName := parts[len(parts)-1]

	// Check if Atlas is installed
	if !isAtlasInstalled() {
		fmt.Println("⚠️  Atlas CLI is not installed.")
		fmt.Println("\nInstall Atlas using one of these methods:")
		fmt.Println("  macOS:  brew install ariga/tap/atlas")
		fmt.Println("  Linux:  curl -sSf https://atlasgo.sh | sh")
		fmt.Println("  Docker: docker pull arigaio/atlas")
		fmt.Println("\nFor more info: https://atlasgo.io/getting-started/")
		fmt.Println("\n⏳ Continuing with file generation...")
	}

	// Create migrations directory
	migrationsDir := "migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create migrations directory: %v", err)
	}
	fmt.Printf("  ✅ Created %s/\n", migrationsDir)

	// Create atlas.hcl configuration
	atlasConfig := generateAtlasConfig(projectName)
	if err := os.WriteFile("atlas.hcl", []byte(atlasConfig), 0644); err != nil {
		log.Fatalf("❌ Failed to write atlas.hcl: %v", err)
	}
	fmt.Println("  ✅ Created atlas.hcl")

	// Create schema.hcl
	schemaHCL := generateSchemaHCL(projectName)
	if err := os.WriteFile("schema.hcl", []byte(schemaHCL), 0644); err != nil {
		log.Fatalf("❌ Failed to write schema.hcl: %v", err)
	}
	fmt.Println("  ✅ Created schema.hcl")

	// Create initial migration
	initialMigration := generateInitialMigration(projectName)
	migrationFile := filepath.Join(migrationsDir, "00001_initial.sql")
	if err := os.WriteFile(migrationFile, []byte(initialMigration), 0644); err != nil {
		log.Fatalf("❌ Failed to write initial migration: %v", err)
	}
	fmt.Printf("  ✅ Created %s\n", migrationFile)

	// Create atlas.sum placeholder
	sumFile := filepath.Join(migrationsDir, "atlas.sum")
	if err := os.WriteFile(sumFile, []byte("h1:placeholder\n"), 0644); err != nil {
		log.Fatalf("❌ Failed to write atlas.sum: %v", err)
	}
	fmt.Printf("  ✅ Created %s\n", sumFile)

	// Create seeds directory
	seedsDir := filepath.Join("internal", "database", "seeds")
	if err := os.MkdirAll(seedsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create seeds directory: %v", err)
	}
	fmt.Printf("  ✅ Created %s/\n", seedsDir)

	sampleSeed := filepath.Join(seedsDir, "seeds.go")
	if _, err := os.Stat(sampleSeed); os.IsNotExist(err) {
		seedContent := `package seeds

import (
	"log"

	"gorm.io/gorm"
)

// Run executes all seed functions
func Run(db *gorm.DB) error {
	log.Println("Seeding database...")

	// Add your seed functions here
	// if err := seedUsers(db); err != nil {
	// 	return err
	// }

	log.Println("Database seeded successfully")
	return nil
}
`
		if err := os.WriteFile(sampleSeed, []byte(seedContent), 0644); err != nil {
			log.Fatalf("❌ Failed to write seed file: %v", err)
		}
		fmt.Printf("  ✅ Created %s\n", sampleSeed)
	}

	fmt.Println("\n✨ Atlas migration setup initialized!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Set DATABASE_URL environment variable")
	fmt.Println("     export DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'")
	fmt.Println("  2. Update schema.hcl with your database schema")
	fmt.Println("  3. Generate migrations from schema changes:")
	fmt.Println("     echonext db migrate:diff my_changes")
	fmt.Println("  4. Apply migrations:")
	fmt.Println("     echonext db migrate")
	fmt.Println("\nUseful commands:")
	fmt.Println("  echonext db migrate:status  - Check migration status")
	fmt.Println("  echonext db migrate:lint    - Lint migrations for issues")
	fmt.Println("  echonext db schema:inspect  - View current database schema")
}

func runDBMigrate(cmd *cobra.Command, args []string, dryRun bool, env string) {
	fmt.Println("🔄 Running database migrations...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	if !isAtlasInstalled() {
		log.Fatal("❌ Atlas CLI is not installed. Run 'echonext db init' for installation instructions.")
	}

	atlasArgs := []string{"migrate", "apply", "--dir", "file://migrations", "--env", env}

	if dryRun {
		atlasArgs = append(atlasArgs, "--dry-run")
		fmt.Println("📋 Dry run mode - no changes will be applied")
	}

	// Check for atlas.hcl
	if _, err := os.Stat("atlas.hcl"); err == nil {
		atlasArgs = append(atlasArgs, "-c", "file://atlas.hcl")
	}

	if err := runAtlasCommand(atlasArgs...); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	if !dryRun {
		fmt.Println("✅ Migrations applied successfully")
	}
}

func runDBMigrateStatus(cmd *cobra.Command, args []string, env string) {
	fmt.Println("📊 Migration status:")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	if !isAtlasInstalled() {
		log.Fatal("❌ Atlas CLI is not installed. Run 'echonext db init' for installation instructions.")
	}

	atlasArgs := []string{"migrate", "status", "--dir", "file://migrations", "--env", env}

	if _, err := os.Stat("atlas.hcl"); err == nil {
		atlasArgs = append(atlasArgs, "-c", "file://atlas.hcl")
	}

	if err := runAtlasCommand(atlasArgs...); err != nil {
		log.Fatalf("❌ Failed to get migration status: %v", err)
	}
}

func runDBMigrateNew(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Printf("📝 Creating new migration '%s'...\n", name)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	if !isAtlasInstalled() {
		log.Fatal("❌ Atlas CLI is not installed. Run 'echonext db init' for installation instructions.")
	}

	atlasArgs := []string{"migrate", "new", name, "--dir", "file://migrations"}

	if err := runAtlasCommand(atlasArgs...); err != nil {
		log.Fatalf("❌ Failed to create migration: %v", err)
	}

	fmt.Println("✅ Migration file created")
	fmt.Println("\n💡 Edit the migration file in migrations/ directory")
	fmt.Println("   Then run 'atlas migrate hash' to update the checksum")
}

func runDBMigrateDiff(cmd *cobra.Command, args []string, env string) {
	name := args[0]
	fmt.Printf("🔍 Generating migration '%s' from schema diff...\n", name)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	if !isAtlasInstalled() {
		log.Fatal("❌ Atlas CLI is not installed. Run 'echonext db init' for installation instructions.")
	}

	// Check for schema.hcl
	if _, err := os.Stat("schema.hcl"); os.IsNotExist(err) {
		log.Fatal("❌ schema.hcl not found. Run 'echonext db init' first.")
	}

	atlasArgs := []string{"migrate", "diff", name, "--dir", "file://migrations", "--env", env}

	if _, err := os.Stat("atlas.hcl"); err == nil {
		atlasArgs = append(atlasArgs, "-c", "file://atlas.hcl")
	}

	if err := runAtlasCommand(atlasArgs...); err != nil {
		log.Fatalf("❌ Failed to generate migration: %v", err)
	}

	fmt.Println("✅ Migration generated from schema diff")
	fmt.Println("\n💡 Review the generated migration in migrations/ directory")
}

func runDBMigrateDown(cmd *cobra.Command, args []string, count int, env string) {
	fmt.Printf("⬇️  Rolling back %d migration(s)...\n", count)

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	if !isAtlasInstalled() {
		log.Fatal("❌ Atlas CLI is not installed. Run 'echonext db init' for installation instructions.")
	}

	atlasArgs := []string{"migrate", "down", fmt.Sprintf("%d", count), "--dir", "file://migrations", "--env", env}

	if _, err := os.Stat("atlas.hcl"); err == nil {
		atlasArgs = append(atlasArgs, "-c", "file://atlas.hcl")
	}

	if err := runAtlasCommand(atlasArgs...); err != nil {
		log.Fatalf("❌ Rollback failed: %v", err)
	}

	fmt.Println("✅ Migration(s) rolled back successfully")
}

func runDBMigrateLint(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Linting migrations...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	if !isAtlasInstalled() {
		log.Fatal("❌ Atlas CLI is not installed. Run 'echonext db init' for installation instructions.")
	}

	atlasArgs := []string{"migrate", "lint", "--dir", "file://migrations", "--latest", "1"}

	if _, err := os.Stat("atlas.hcl"); err == nil {
		atlasArgs = append(atlasArgs, "-c", "file://atlas.hcl")
	}

	if err := runAtlasCommand(atlasArgs...); err != nil {
		log.Fatalf("❌ Lint check failed: %v", err)
	}

	fmt.Println("✅ No issues found")
}

func runDBSchemaInspect(cmd *cobra.Command, args []string, env string) {
	fmt.Println("🔍 Inspecting database schema...")

	if !isEchoNextProject() {
		log.Fatal("❌ Not in an EchoNext project")
	}

	if !isAtlasInstalled() {
		log.Fatal("❌ Atlas CLI is not installed. Run 'echonext db init' for installation instructions.")
	}

	atlasArgs := []string{"schema", "inspect", "--env", env, "--format", "{{ sql . }}"}

	if _, err := os.Stat("atlas.hcl"); err == nil {
		atlasArgs = append(atlasArgs, "-c", "file://atlas.hcl")
	}

	if err := runAtlasCommand(atlasArgs...); err != nil {
		log.Fatalf("❌ Schema inspection failed: %v", err)
	}
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

// Atlas helper functions

func isAtlasInstalled() bool {
	_, err := exec.LookPath("atlas")
	return err == nil
}

func runAtlasCommand(args ...string) error {
	cmd := exec.Command("atlas", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func generateAtlasConfig(projectName string) string {
	return fmt.Sprintf(`// Atlas configuration for %s
// Documentation: https://atlasgo.io/atlas-schema/hcl

// Define development environment
env "local" {
  // Use PostgreSQL as the database driver
  src = "file://schema.hcl"

  // Development database URL - loaded from environment variable
  url = getenv("DATABASE_URL")

  // Dev database for schema diff calculations
  dev = "docker://postgres/16/dev?search_path=public"

  // Migration directory
  migration {
    dir = "file://migrations"
  }

  // Format migration files
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

// Define staging environment
env "staging" {
  src = "file://schema.hcl"
  url = getenv("STAGING_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}

// Define production environment
env "production" {
  src = "file://schema.hcl"
  url = getenv("PRODUCTION_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }

  // Require approval for destructive changes in production
  diff {
    skip {
      drop_column = true
      drop_table  = true
    }
  }
}

// Lint configuration for migration validation
lint {
  // Detect destructive changes
  destructive {
    error = true
  }

  // Detect data-dependent changes
  data_depend {
    error = true
  }

  // Naming conventions
  naming {
    match   = "^[a-z]+(_[a-z]+)*$"
    message = "must be lowercase with underscores"
  }
}
`, projectName)
}

func generateSchemaHCL(projectName string) string {
	return fmt.Sprintf(`// Database schema for %s
// This file is the source of truth for your database schema.
// Atlas will use this file to generate migrations.
//
// Documentation: https://atlasgo.io/atlas-schema/hcl

// Users table
table "users" {
  schema = schema.public

  column "id" {
    null    = false
    type    = bigserial
  }

  column "name" {
    null = false
    type = varchar(255)
  }

  column "email" {
    null = false
    type = varchar(255)
  }

  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }

  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }

  column "deleted_at" {
    null = true
    type = timestamptz
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_users_email" {
    unique  = true
    columns = [column.email]
    where   = "deleted_at IS NULL"
  }

  index "idx_users_deleted_at" {
    columns = [column.deleted_at]
  }
}

// Schema definition
schema "public" {
  comment = "%s database schema"
}
`, projectName, projectName)
}

func generateInitialMigration(projectName string) string {
	return fmt.Sprintf(`-- Migration: Initial schema for %s
-- Created by: echonext db init

-- Create users table
CREATE TABLE IF NOT EXISTS "users" (
    "id" bigserial NOT NULL,
    "name" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id")
);

-- Create unique index on email for active users
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email") WHERE "deleted_at" IS NULL;

-- Create index on deleted_at for soft delete queries
CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");
`, projectName)
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

// Dev, Test, and Build command implementations

// runDev starts the development server with hot reload
func runDev(port int, watch bool, target string) error {
	fmt.Printf("Starting development server for %s on port %d...\n", target, port)
	if watch {
		fmt.Println("File watching enabled")
	}

	server := NewDevServer(port, target, watch)
	return server.Run()
}

// runTest runs the test suite with enhanced features
func runTest(coverage bool, verbose bool, args []string) error {
	if !isEchoNextProject() {
		return fmt.Errorf("not in an EchoNext project. Run from project root")
	}

	fmt.Println("Running tests...")

	// Build the go test command arguments
	testArgs := []string{"test"}

	if verbose {
		testArgs = append(testArgs, "-v")
	}

	if coverage {
		coverFile := "coverage.out"
		testArgs = append(testArgs, "-coverprofile="+coverFile)
	}

	// Add default package pattern
	testArgs = append(testArgs, "./...")

	// Add any extra args passed by user
	testArgs = append(testArgs, args...)

	// Run go test
	cmd := exec.Command("go", testArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	// Show coverage summary if enabled
	if coverage {
		showCoverageSummary("coverage.out")
	}

	// Print result indicator
	if err != nil {
		fmt.Println("\nTests FAILED")
		return err
	}

	fmt.Println("\nTests PASSED")
	return nil
}

// showCoverageSummary displays a summary of test coverage
func showCoverageSummary(coverFile string) {
	if _, err := os.Stat(coverFile); os.IsNotExist(err) {
		return
	}

	fmt.Println("\nCoverage report generated: coverage.out")

	cmd := exec.Command("go", "tool", "cover", "-func="+coverFile)

	// Capture output to show only the total line
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			fmt.Println(line)
			break
		}
	}
}

// BuildTarget represents a build target
type BuildTarget struct {
	Name   string
	Path   string
	Output string
}

// runBuild builds the project for production
func runBuild(target string, output string) error {
	if !isEchoNextProject() {
		return fmt.Errorf("not in an EchoNext project. Run from project root")
	}

	// Define available targets
	allTargets := []BuildTarget{
		{Name: "api", Path: "./cmd/api", Output: "api"},
		{Name: "worker", Path: "./cmd/worker", Output: "worker"},
		{Name: "cli", Path: "./cmd/cli", Output: "cli"},
		{Name: "migration", Path: "./cmd/migration", Output: "migration"},
	}

	// Filter targets based on selection
	var targets []BuildTarget
	if target == "all" {
		targets = allTargets
	} else {
		for _, t := range allTargets {
			if t.Name == target {
				targets = append(targets, t)
				break
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf("unknown target: %s. Available: all, api, worker, cli, migration", target)
		}
	}

	// Create output directory
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get version info for ldflags
	version := getVersionFromGit()
	buildTime := fmt.Sprintf("%d", getCurrentTimestamp())

	// Build ldflags for production
	ldflags := fmt.Sprintf("-s -w -X main.Version=%s -X main.BuildTime=%s", version, buildTime)

	fmt.Printf("Building %d target(s) to %s/\n", len(targets), output)
	fmt.Printf("Version: %s\n", version)
	fmt.Println()

	// Build each target
	successCount := 0
	for _, t := range targets {
		// Check if target path exists
		if _, err := os.Stat(t.Path); os.IsNotExist(err) {
			fmt.Printf("  Skipping %s (not found at %s)\n", t.Name, t.Path)
			continue
		}

		outputPath := filepath.Join(output, t.Output)
		fmt.Printf("  Building %s...", t.Name)

		startTime := getCurrentTimestamp()

		// Build command
		cmd := exec.Command("go", "build",
			"-ldflags", ldflags,
			"-trimpath",
			"-o", outputPath,
			t.Path,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf(" FAILED\n")
			fmt.Printf("    Error: %v\n", err)
			continue
		}

		// Get binary size
		info, _ := os.Stat(outputPath)
		duration := getCurrentTimestamp() - startTime

		fmt.Printf(" OK (%s, %dms)\n", formatSize(info.Size()), duration)
		successCount++
	}

	fmt.Println()
	if successCount == len(targets) {
		fmt.Printf("Build complete: %d/%d targets built successfully\n", successCount, len(targets))
	} else if successCount > 0 {
		fmt.Printf("Build partial: %d/%d targets built\n", successCount, len(targets))
	} else {
		return fmt.Errorf("build failed: no targets built successfully")
	}

	return nil
}

// getVersionFromGit gets version information from git
func getVersionFromGit() string {
	// Try git describe first
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	// Fall back to short commit hash
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	return "dev"
}

// getCurrentTimestamp returns current unix timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// formatSize formats a file size in human-readable form
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// runGenerateOpenAPI generates code from an OpenAPI specification
func runGenerateOpenAPI(specPath, outputDir, packageName string, generateTests bool) error {
	fmt.Printf("Generating code from OpenAPI spec: %s\n", specPath)
	fmt.Printf("Output directory: %s\n", outputDir)
	fmt.Printf("Package name: %s\n", packageName)
	fmt.Println()

	// Create generator
	gen, err := NewOpenAPIGenerator(specPath, GeneratorOptions{
		OutputDir:     outputDir,
		PackageName:   packageName,
		GenerateTests: generateTests,
	})
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	// Run generation
	files, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Report generated files
	fmt.Println("Generated files:")
	for _, f := range files.Models {
		fmt.Printf("  - %s\n", f)
	}
	for _, f := range files.DTOs {
		fmt.Printf("  - %s\n", f)
	}
	for _, f := range files.Handlers {
		fmt.Printf("  - %s\n", f)
	}
	if files.Routes != "" {
		fmt.Printf("  - %s\n", files.Routes)
	}
	for _, f := range files.Tests {
		fmt.Printf("  - %s\n", f)
	}

	fmt.Println()
	fmt.Println("Code generation complete!")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Review generated code in %s\n", outputDir)
	fmt.Printf("  2. Implement handler logic in handlers/handlers.go\n")
	fmt.Printf("  3. Register routes in your main.go:\n")
	fmt.Printf("     %s.RegisterRoutes(app)\n", packageName)

	return nil
}

// runGenerateGraphQL generates GraphQL boilerplate for gqlgen integration
func runGenerateGraphQL(cmd *cobra.Command, args []string) {
	fmt.Println("Generating GraphQL boilerplate...")

	if !isEchoNextProject() {
		log.Fatal("Not in an EchoNext project. Run this command from your project root.")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("Failed to read module name: %v", err)
	}

	// Create template engine
	engine, err := generator.NewEngine()
	if err != nil {
		log.Fatalf("Failed to create template engine: %v", err)
	}

	// Template data
	data := map[string]string{
		"Module": module,
	}

	// Create directories
	graphDir := "graph"
	if err := os.MkdirAll(graphDir, 0755); err != nil {
		log.Fatalf("Failed to create graph directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(graphDir, "model"), 0755); err != nil {
		log.Fatalf("Failed to create model directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(graphDir, "examples"), 0755); err != nil {
		log.Fatalf("Failed to create examples directory: %v", err)
	}
	if err := os.MkdirAll("tools", 0755); err != nil {
		log.Fatalf("Failed to create tools directory: %v", err)
	}

	// Generate files from templates
	templates := []struct {
		template string
		output   string
	}{
		{"graphql/gqlgen.yml.tmpl", "gqlgen.yml"},
		{"graphql/schema.graphqls.tmpl", filepath.Join(graphDir, "schema.graphqls")},
		{"graphql/resolver.go.tmpl", filepath.Join(graphDir, "resolver.go")},
		{"graphql/server.go.tmpl", filepath.Join(graphDir, "examples", "server_example.go.txt")},
		{"graphql/tools.go.tmpl", filepath.Join("tools", "tools.go")},
	}

	for _, t := range templates {
		content, err := engine.Execute(t.template, data)
		if err != nil {
			log.Fatalf("Failed to execute template %s: %v", t.template, err)
		}
		if err := os.WriteFile(t.output, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write %s: %v", t.output, err)
		}
		fmt.Printf("  Created %s\n", t.output)
	}

	fmt.Println()
	fmt.Println("GraphQL boilerplate generated!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Install gqlgen: go get github.com/99designs/gqlgen")
	fmt.Println("  2. Edit graph/schema.graphqls with your schema")
	fmt.Println("  3. Run: go run github.com/99designs/gqlgen generate")
	fmt.Println("  4. Implement resolvers in graph/*.resolvers.go")
	fmt.Println("  5. See graph/examples/server_example.go.txt for integration")
	fmt.Println()
	fmt.Println("Documentation: https://gqlgen.com/getting-started/")
}

// runGenerateWebSocket generates WebSocket handler boilerplate
func runGenerateWebSocket(cmd *cobra.Command, args []string) {
	name := strings.ToLower(args[0])
	fmt.Printf("Generating WebSocket handler '%s'...\n", name)

	if !isEchoNextProject() {
		log.Fatal("Not in an EchoNext project. Run this command from your project root.")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("Failed to read module name: %v", err)
	}

	// Create template engine
	engine, err := generator.NewEngine()
	if err != nil {
		log.Fatalf("Failed to create template engine: %v", err)
	}

	// Template data
	data := map[string]interface{}{
		"Module":     module,
		"Name":       name,
		"PascalName": toPascalCase(name),
	}

	// Create directories
	wsDir := filepath.Join("internal", "ws", name)
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		log.Fatalf("Failed to create ws directory: %v", err)
	}

	// Generate files from templates
	templates := []struct {
		template string
		output   string
	}{
		{"websocket/handler.go.tmpl", filepath.Join(wsDir, "handler.go")},
		{"websocket/hub.go.tmpl", filepath.Join(wsDir, "hub.go")},
		{"websocket/message.go.tmpl", filepath.Join(wsDir, "message.go")},
	}

	for _, t := range templates {
		content, err := engine.Execute(t.template, data)
		if err != nil {
			log.Fatalf("Failed to execute template %s: %v", t.template, err)
		}
		if err := os.WriteFile(t.output, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write %s: %v", t.output, err)
		}
		fmt.Printf("  Created %s\n", t.output)
	}

	fmt.Println()
	fmt.Println("WebSocket handler generated!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Review and customize internal/ws/%s/handler.go\n", name)
	fmt.Println("  2. Register the WebSocket route in your main.go:")
	fmt.Printf("     import ws%s \"%s/internal/ws/%s\"\n", toPascalCase(name), module, name)
	fmt.Printf("     app.WS(\"/ws/%s\", ws%s.NewHandler())\n", name, toPascalCase(name))
	fmt.Println()
	fmt.Println("Documentation: https://github.com/abdussamadbello/echonext#websocket")
}

// runGenerateUpload generates file upload handler boilerplate
func runGenerateUpload(cmd *cobra.Command, args []string) {
	name := strings.ToLower(args[0])
	fmt.Printf("Generating upload handler '%s'...\n", name)

	if !isEchoNextProject() {
		log.Fatal("Not in an EchoNext project. Run this command from your project root.")
	}

	module, err := getModuleName()
	if err != nil {
		log.Fatalf("Failed to read module name: %v", err)
	}

	// Create template engine
	engine, err := generator.NewEngine()
	if err != nil {
		log.Fatalf("Failed to create template engine: %v", err)
	}

	// Template data
	data := map[string]interface{}{
		"Module":     module,
		"Name":       name,
		"PascalName": toPascalCase(name),
	}

	// Create directories
	uploadDir := filepath.Join("internal", "upload", name)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Generate files from templates
	templates := []struct {
		template string
		output   string
	}{
		{"upload/handler.go.tmpl", filepath.Join(uploadDir, "handler.go")},
		{"upload/dto.go.tmpl", filepath.Join(uploadDir, "dto.go")},
	}

	for _, t := range templates {
		content, err := engine.Execute(t.template, data)
		if err != nil {
			log.Fatalf("Failed to execute template %s: %v", t.template, err)
		}
		if err := os.WriteFile(t.output, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to write %s: %v", t.output, err)
		}
		fmt.Printf("  Created %s\n", t.output)
	}

	fmt.Println()
	fmt.Println("Upload handler generated!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Review and customize internal/upload/%s/handler.go\n", name)
	fmt.Println("  2. Register the upload route in your main.go:")
	fmt.Printf("     import upload%s \"%s/internal/upload/%s\"\n", toPascalCase(name), module, name)
	fmt.Printf("     app.Upload(\"/%s/upload\", upload%s.Handler)\n", name, toPascalCase(name))
	fmt.Println()
	fmt.Println("Documentation: https://github.com/abdussamadbello/echonext#file-uploads")
}
