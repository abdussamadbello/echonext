package main

import (
	"fmt"
	"log"
	"strings"
)

// generateConfigYAML creates configuration files for different environments
func (g *ProjectGenerator) generateConfigYAML(environment string) string {
	switch environment {
	case "development":
		return g.generateDevelopmentConfig()
	case "production":
		return g.generateProductionConfig()
	case "test":
		return g.generateTestConfig()
	default:
		return g.generateDevelopmentConfig()
	}
}

// generateDevelopmentConfig creates development configuration
func (g *ProjectGenerator) generateDevelopmentConfig() string {
	content, err := g.templateGenerator.GenerateDevelopmentConfig()
	if err != nil {
		log.Fatalf("Failed to generate development config: %v", err)
	}
	return content
}

// generateProductionConfig creates production configuration
func (g *ProjectGenerator) generateProductionConfig() string {
	content, err := g.templateGenerator.GenerateProductionConfig()
	if err != nil {
		log.Fatalf("Failed to generate production config: %v", err)
	}
	return content
}

// generateTestConfig creates test configuration
func (g *ProjectGenerator) generateTestConfig() string {
	content, err := g.templateGenerator.GenerateTestConfig()
	if err != nil {
		log.Fatalf("Failed to generate test config: %v", err)
	}
	return content
}

// generateDockerCompose creates docker-compose.yml
func (g *ProjectGenerator) generateDockerCompose() string {
	return fmt.Sprintf(`version: '3.8'

services:
  # API Server
  api:
    build:
      context: ..
      dockerfile: infrastructure/Dockerfile.api
    ports:
      - "8080:8080"
    environment:
      - %s_DATABASE_DSN=postgres://postgres:password@postgres:5432/%s_dev?sslmode=disable
      - %s_CACHE_ADDRESS=redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  # Background Worker
  worker:
    build:
      context: ..
      dockerfile: infrastructure/Dockerfile.worker
    environment:
      - %s_DATABASE_DSN=postgres://postgres:password@postgres:5432/%s_dev?sslmode=disable
      - %s_CACHE_ADDRESS=redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  # PostgreSQL Database
  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=%s_dev
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ../scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    restart: unless-stopped

  # Redis Cache
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  # Migration (run once)
  migration:
    build:
      context: ..
      dockerfile: infrastructure/Dockerfile.api
    command: ["./migration", "up"]
    environment:
      - %s_DATABASE_DSN=postgres://postgres:password@postgres:5432/%s_dev?sslmode=disable
    depends_on:
      - postgres
    restart: "no"

volumes:
  postgres_data:
  redis_data:
`, 
	toEnvVar(g.Name), g.Name,
	toEnvVar(g.Name),
	toEnvVar(g.Name), g.Name,
	toEnvVar(g.Name),
	g.Name,
	toEnvVar(g.Name), g.Name)
}

// generateDockerfileAPI creates Dockerfile for API server
func (g *ProjectGenerator) generateDockerfileAPI() string {
	return fmt.Sprintf(`# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migration ./cmd/migration

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/migration .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN adduser -D -s /bin/sh %s
USER %s

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the API
CMD ["./api"]
`, g.Name, g.Name)
}

// generateDockerfileWorker creates Dockerfile for background worker
func (g *ProjectGenerator) generateDockerfileWorker() string {
	return fmt.Sprintf(`# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the worker binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/worker .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN adduser -D -s /bin/sh %s
USER %s

# Run the worker
CMD ["./worker"]
`, g.Name, g.Name)
}

// generateGitignore creates .gitignore file
func (g *ProjectGenerator) generateGitignore() string {
	return `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out
coverage.html

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Environment variables
.env
.env.local
.env.*.local

# Log files
*.log
logs/

# Database files
*.db
*.sqlite3

# Cache directories
.cache/

# Temporary files
tmp/
temp/

# Build artifacts
/api
/worker
/cli
/migration
`
}

// Helper function to convert project name to environment variable format
func toEnvVar(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}