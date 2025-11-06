package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newInitCmd returns the init command
func newInitCmd() *cobra.Command {
	var template string
	var module string

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new EchoNext project",
		Long: `Initialize a new EchoNext project with the specified template.

Available templates:
  standard    - Full-featured API with multiple domains (default)
  minimal     - Basic API with single domain  
  microservice- Production microservice template
  monolith    - Enterprise application template

Examples:
  echonext init myapi
  echonext init blog-api --template=standard
  echonext init payment-service --template=microservice`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			
			// Validate project name
			if !isValidProjectName(projectName) {
				return fmt.Errorf("invalid project name: %s. Use lowercase letters, numbers, and hyphens only", projectName)
			}

			// Check if directory exists
			if _, err := os.Stat(projectName); !os.IsNotExist(err) {
				return fmt.Errorf("directory %s already exists", projectName)
			}

			// Set default module name if not provided
			if module == "" {
				module = projectName
			}

			// Create project
			generator := &ProjectGenerator{
				Name:     projectName,
				Module:   module,
				Template: template,
			}

			return generator.Generate()
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "standard", "Project template (standard, minimal, microservice, monolith)")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Go module name (defaults to project name)")

	return cmd
}

// isValidProjectName validates the project name
func isValidProjectName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	
	return true
}

// ProjectGenerator handles project generation
type ProjectGenerator struct {
	Name     string
	Module   string
	Template string
}

// Generate creates the project structure
func (g *ProjectGenerator) Generate() error {
	fmt.Printf("🚀 Creating EchoNext project '%s' with template '%s'...\n", g.Name, g.Template)

	// Create project directory
	if err := os.MkdirAll(g.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate based on template
	switch g.Template {
	case "minimal":
		return g.generateMinimalProject()
	case "microservice":
		return g.generateMicroserviceProject()
	case "monolith":
		return g.generateMonolithProject()
	default: // "standard"
		return g.generateStandardProject()
	}
}

// generateStandardProject creates a standard project with multiple executables
func (g *ProjectGenerator) generateStandardProject() error {
	// Create directory structure
	dirs := []string{
		"cmd/api",
		"cmd/worker", 
		"cmd/cli",
		"cmd/migration",
		"domain/user",
		"internal/config",
		"internal/database",
		"internal/server",
		"tests/integration",
		"tests/fixtures",
		"tests/helpers",
		"configs",
		"migrations",
		"scripts",
	}

	for _, dir := range dirs {
		path := filepath.Join(g.Name, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	// Generate files
	files := map[string]string{
		"go.mod":                    g.generateGoMod(),
		"README.md":                 g.generateReadme(),
		"Makefile":                  g.generateMakefile(),
		"docker-compose.yml":        g.generateDockerCompose(),
		"Dockerfile.api":            g.generateDockerfileAPI(),
		"Dockerfile.worker":         g.generateDockerfileWorker(),
		".gitignore":                g.generateGitignore(),
		"cmd/api/main.go":           g.generateAPIMain(),
		"cmd/worker/main.go":        g.generateWorkerMain(),
		"cmd/cli/main.go":           g.generateCLIMain(),
		"cmd/migration/main.go":     g.generateMigrationMain(),
		"internal/config/config.go": g.generateConfig(),
		"internal/database/database.go": g.generateDatabase(),
		"internal/server/server.go": g.generateServer(),
		"configs/development.yaml":  g.generateConfigYAML("development"),
		"configs/production.yaml":   g.generateConfigYAML("production"),
		"configs/test.yaml":         g.generateConfigYAML("test"),
		"domain/user/model.go":      g.generateUserModel(),
		"domain/user/service.go":    g.generateUserService(),
		"domain/user/handler.go":    g.generateUserHandler(),
		"domain/user/dto.go":        g.generateUserDTO(),
		"domain/user/service_test.go": g.generateUserTest(),
	}

	for filePath, content := range files {
		fullPath := filepath.Join(g.Name, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
	}

	fmt.Printf("✅ Project '%s' created successfully!\n\n", g.Name)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", g.Name)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  make run-api\n\n")
	fmt.Printf("Documentation will be available at http://localhost:8080/api/docs\n")

	return nil
}

// Helper methods for other template types
func (g *ProjectGenerator) generateMinimalProject() error {
	// TODO: Implement minimal template
	fmt.Println("⚠️  Minimal template not yet implemented, using standard template")
	return g.generateStandardProject()
}

func (g *ProjectGenerator) generateMicroserviceProject() error {
	// TODO: Implement microservice template
	fmt.Println("⚠️  Microservice template not yet implemented, using standard template") 
	return g.generateStandardProject()
}

func (g *ProjectGenerator) generateMonolithProject() error {
	// TODO: Implement monolith template
	fmt.Println("⚠️  Monolith template not yet implemented, using standard template")
	return g.generateStandardProject()
}

// Template generation methods
func (g *ProjectGenerator) generateGoMod() string {
	return fmt.Sprintf(`module %s

go 1.21

require (
	github.com/abdussamadbello/echonext v0.0.0-00010101000000-000000000000
	github.com/labstack/echo/v4 v4.11.4
	github.com/spf13/viper v1.18.2
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
)

replace github.com/abdussamadbello/echonext => ../../
`, g.Module)
}

func (g *ProjectGenerator) generateReadme() string {
	return fmt.Sprintf(`# %s

A type-safe API built with [EchoNext](https://github.com/abdussamadbello/echonext).

## Features

- 🚀 Type-safe HTTP handlers
- 📝 Automatic OpenAPI generation
- 🔧 Request/response validation
- 🗄️ Database integration with GORM
- 🐳 Docker support
- 🧪 Comprehensive testing

## Quick Start

### Development

`+"```"+`bash
# Install dependencies
go mod tidy

# Start API server
make run-api

# Start background worker (in separate terminal)
make run-worker

# Run tests
make test
`+"```"+`

### API Documentation

Visit http://localhost:8080/api/docs for interactive API documentation.

### Build & Deploy

`+"```"+`bash
# Build all executables
make build

# Build Docker images
make docker-build

# Deploy with Docker Compose
docker-compose up -d
`+"```"+`

## Project Structure

`+"```"+`
%s/
├── cmd/                    # Executable entry points
│   ├── api/               # HTTP API server
│   ├── worker/            # Background worker
│   ├── cli/               # CLI tool
│   └── migration/         # Database migrations
├── domain/                # Business domains
│   └── user/              # User domain
├── internal/              # Private packages
│   ├── config/            # Configuration
│   ├── database/          # Database setup
│   └── server/            # Server setup
├── tests/                 # Test files
├── configs/               # Configuration files
└── migrations/            # SQL migrations
`+"```"+`

## Configuration

Configuration is managed via YAML files in the `+"` configs/ `"+` directory and can be overridden with environment variables.

Example environment variables:
`+"```"+`bash
export MYAPP_APP_PORT=9090
export MYAPP_DATABASE_DSN="postgres://user:pass@localhost/dbname?sslmode=disable"
`+"```"+`

## API Endpoints

- `+"` GET /health `"+` - Health check
- `+"` POST /users `"+` - Create user
- `+"` GET /users `"+` - List users
- `+"` GET /users/:id `"+` - Get user by ID
- `+"` PUT /users/:id `"+` - Update user
- `+"` DELETE /users/:id `"+` - Delete user

## Generated by EchoNext CLI

This project was generated using:
`+"```"+`bash
echonext init %s --template=standard
`+"```"+`

For more information, visit https://github.com/abdussamadbello/echonext
`, strings.ToTitle(g.Name), g.Name, g.Name)
}

func (g *ProjectGenerator) generateMakefile() string {
	return fmt.Sprintf(`# Makefile for %s

.PHONY: help run-api run-worker run-cli build test clean docker-build docker-push

# Default target
help:
	@echo "Available commands:"
	@echo "  run-api      - Start API server"
	@echo "  run-worker   - Start background worker"
	@echo "  run-cli      - Run CLI tool"
	@echo "  build        - Build all executables"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker images"

# Development
run-api:
	@echo "Starting API server..."
	go run ./cmd/api

run-worker:
	@echo "Starting background worker..."
	go run ./cmd/worker

run-cli:
	@echo "Running CLI..."
	go run ./cmd/cli

# Build
build:
	@echo "Building executables..."
	@mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/cli ./cmd/cli
	go build -o bin/migration ./cmd/migration

# Testing
test:
	@echo "Running tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Utilities
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database
db-migrate:
	@echo "Running database migrations..."
	go run ./cmd/migration up

db-seed:
	@echo "Seeding database..."
	go run ./cmd/migration seed

# Docker
docker-build:
	@echo "Building Docker images..."
	docker build -f Dockerfile.api -t %s-api .
	docker build -f Dockerfile.worker -t %s-worker .

docker-push:
	@echo "Pushing Docker images..."
	docker push %s-api
	docker push %s-worker

# Development database
dev-db:
	@echo "Starting development database..."
	docker-compose up -d postgres redis
`, g.Name, g.Name, g.Name, g.Name, g.Name)
}

// Additional methods would continue here...
// This is getting quite long, so I'll create separate files for the templates