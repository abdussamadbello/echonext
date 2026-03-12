# Deployment Guide

Learn how to deploy your EchoNext applications to production.

## Table of Contents

- [Building for Production](#building-for-production)
- [Docker Deployment](#docker-deployment)
- [Cloud Platforms](#cloud-platforms)
- [Environment Configuration](#environment-configuration)
- [Database Migrations](#database-migrations)
- [Security Best Practices](#security-best-practices)
- [Monitoring and Logging](#monitoring-and-logging)
- [Performance Optimization](#performance-optimization)

## Building for Production

### Compile Your Application

```bash
# Build for current platform
go build -o api ./cmd/api

# Build with optimizations
go build -ldflags="-s -w" -o api ./cmd/api

# Cross-compile for Linux (from Mac/Windows)
GOOS=linux GOARCH=amd64 go build -o api ./cmd/api
```

### Build Flags Explained

- `-ldflags="-s -w"` - Strip debug info and reduce binary size
- `-trimpath` - Remove file system paths from binary
- `-tags` - Include/exclude build tags

### Production Build Script

```bash
#!/bin/bash
# build.sh

VERSION=$(git describe --tags --always --dirty)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

go build \
  -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
  -o api \
  ./cmd/api

echo "Built version: ${VERSION}"
```

## Docker Deployment

### Multi-Stage Dockerfile

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o api \
    ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/api .

# Copy config files (if needed)
COPY configs/ ./configs/

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run application
CMD ["./api"]
```

### Docker Compose for Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:password@db:5432/mydb
      - ENVIRONMENT=development
    depends_on:
      - db
    volumes:
      - ./configs:/root/configs

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=mydb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### Build and Run with Docker

```bash
# Build image
docker build -t myapi:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  --name myapi \
  myapi:latest

# View logs
docker logs -f myapi

# With docker-compose
docker-compose up -d
docker-compose logs -f
```

## Cloud Platforms

### Deploy to AWS ECS

1. **Build and push to ECR:**

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

# Build and tag
docker build -t myapi .
docker tag myapi:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/myapi:latest

# Push
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/myapi:latest
```

2. **Create ECS task definition:**

```json
{
  "family": "myapi",
  "containerDefinitions": [
    {
      "name": "myapi",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/myapi:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DATABASE_URL",
          "value": "postgres://..."
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/myapi",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ],
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512"
}
```

### Deploy to Google Cloud Run

```bash
# Build and submit
gcloud builds submit --tag gcr.io/PROJECT-ID/myapi

# Deploy
gcloud run deploy myapi \
  --image gcr.io/PROJECT-ID/myapi \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars DATABASE_URL="postgres://..."
```

### Deploy to Heroku

```bash
# Login
heroku login

# Create app
heroku create myapi

# Add PostgreSQL
heroku addons:create heroku-postgresql:hobby-dev

# Deploy
git push heroku main

# Scale
heroku ps:scale web=1

# View logs
heroku logs --tail
```

### Deploy to DigitalOcean App Platform

```yaml
# .do/app.yaml
name: myapi
services:
  - name: api
    source_dir: /
    github:
      repo: username/myapi
      branch: main
      deploy_on_push: true
    build_command: go build -o api ./cmd/api
    run_command: ./api
    envs:
      - key: DATABASE_URL
        value: ${db.DATABASE_URL}
    http_port: 8080
databases:
  - name: db
    engine: PG
    version: "15"
```

## Environment Configuration

### Using Environment Variables

```go
// internal/config/config.go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port         string
    DatabaseURL  string
    Environment  string
    JWTSecret    string
    LogLevel     string
}

func Load() *Config {
    return &Config{
        Port:         getEnv("PORT", "8080"),
        DatabaseURL:  getEnv("DATABASE_URL", ""),
        Environment:  getEnv("ENVIRONMENT", "development"),
        JWTSecret:    getEnv("JWT_SECRET", ""),
        LogLevel:     getEnv("LOG_LEVEL", "info"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Using Configuration Files

```yaml
# configs/production.yaml
app:
  name: "My API"
  version: "1.0.0"
  environment: "production"
  port: 8080
  debug: false

database:
  driver: "postgres"
  dsn: "${DATABASE_URL}"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

cache:
  driver: "redis"
  address: "${REDIS_URL}"
  default_ttl: 3600

logger:
  level: "info"
  format: "json"
  output: "stdout"
```

Load config:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/config"

var cfg Config
if err := config.LoadFromFile(&cfg, "configs/production.yaml"); err != nil {
    log.Fatal(err)
}
```

### .env File Support

```bash
# .env
PORT=8080
DATABASE_URL=postgres://user:pass@localhost/db
JWT_SECRET=your-secret-key
LOG_LEVEL=info
ENVIRONMENT=production
```

Load with:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }
    
    // Use environment variables
    config := config.Load()
}
```

## Database Migrations

### Using golang-migrate

```bash
# Install migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq create_users_table

# Run migrations
migrate -database "postgres://user:pass@localhost/db?sslmode=disable" \
        -path migrations up

# Rollback
migrate -database "postgres://user:pass@localhost/db?sslmode=disable" \
        -path migrations down 1
```

### Migration in Code

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(databaseURL string) error {
    m, err := migrate.New(
        "file://migrations",
        databaseURL,
    )
    if err != nil {
        return err
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

### Startup Migration Check

```go
func main() {
    // Load config
    cfg := config.Load()
    
    // Run migrations
    if err := runMigrations(cfg.DatabaseURL); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    
    // Connect to database
    db, err := database.Connect(...)
    
    // Start server
    app := echonext.New()
    // ...
}
```

## Security Best Practices

### 1. Use HTTPS

```go
// Force HTTPS redirect
app.Pre(middleware.HTTPSRedirect())

// Or start with TLS
app.StartTLS(":443", "cert.pem", "key.pem")
```

### 2. Set Security Headers

```go
import "github.com/labstack/echo/v5/middleware"

app.Use(middleware.SecureWithConfig(middleware.SecureConfig{
    XSSProtection:         "1; mode=block",
    ContentTypeNosniff:    "nosniff",
    XFrameOptions:         "SAMEORIGIN",
    HSTSMaxAge:            31536000,
    ContentSecurityPolicy: "default-src 'self'",
}))
```

### 3. Rate Limiting

```go
app.Use(middleware.RateLimiter(
    middleware.NewRateLimiterMemoryStore(20), // 20 requests per second
))
```

### 4. CORS Configuration

```go
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"https://yourdomain.com"},
    AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
}))
```

### 5. Environment-Specific Settings

```go
if cfg.Environment == "production" {
    app.Debug = false
    app.Use(middleware.Recover())
    // Production-only middleware
} else {
    app.Debug = true
    // Development-only middleware
}
```

## Monitoring and Logging

### Structured Logging

```go
import "github.com/labstack/gommon/log"

app.Logger.SetLevel(log.INFO)
app.Logger.SetOutput(os.Stdout)

// In handlers
app.Logger.Info("User created", "user_id", user.ID)
app.Logger.Error("Failed to create user", "error", err)
```

### Request Logging

```go
app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Format: `{"time":"${time_rfc3339}","method":"${method}","uri":"${uri}",` +
            `"status":${status},"latency":"${latency_human}"}` + "\n",
}))
```

### Health Check Endpoint

```go
app.GET("/health", func(c echo.Context) error {
    // Check database
    if err := db.Ping(); err != nil {
        return c.JSON(503, map[string]string{
            "status": "unhealthy",
            "error":  err.Error(),
        })
    }
    
    return c.JSON(200, map[string]string{
        "status": "healthy",
    })
})
```

### Metrics with Prometheus

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Expose metrics
app.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
```

## Performance Optimization

### 1. Enable Gzip Compression

```go
app.Use(middleware.Gzip())
```

### 2. Database Connection Pooling

```go
cfg := database.DefaultConfig()
cfg.MaxOpenConns = 25
cfg.MaxIdleConns = 5
cfg.ConnMaxLifetime = 5 * time.Minute

db, err := database.Connect(driver, cfg)
```

### 3. Caching

```go
import "github.com/go-redis/redis/v8"

// Redis cache
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Cache handler responses
func getCachedUser(c echo.Context) error {
    id := c.Param("id")
    
    // Try cache first
    cached, err := rdb.Get(ctx, "user:"+id).Result()
    if err == nil {
        return c.JSONBlob(200, []byte(cached))
    }
    
    // Fetch from database
    user, err := service.GetUser(id)
    if err != nil {
        return err
    }
    
    // Cache for 5 minutes
    data, _ := json.Marshal(user)
    rdb.Set(ctx, "user:"+id, data, 5*time.Minute)
    
    return c.JSON(200, user)
}
```

### 4. Graceful Shutdown

```go
func main() {
    app := echonext.New()
    // ... setup routes
    
    // Start server in goroutine
    go func() {
        if err := app.Start(":8080"); err != nil && err != http.ErrServerClosed {
            app.Logger.Fatal(err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    
    // Graceful shutdown with 10 second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := app.Shutdown(ctx); err != nil {
        app.Logger.Fatal(err)
    }
}
```

## Deployment Checklist

- [ ] Build optimized binary
- [ ] Set environment variables
- [ ] Configure database connection
- [ ] Run database migrations
- [ ] Enable HTTPS
- [ ] Set security headers
- [ ] Configure CORS
- [ ] Enable rate limiting
- [ ] Set up logging
- [ ] Add health checks
- [ ] Configure monitoring
- [ ] Set up backups
- [ ] Document deployment process
- [ ] Test in staging environment
- [ ] Plan rollback strategy

## Next Steps

- [Security Best Practices](../advanced/security.md)
- [Performance Optimization](../advanced/performance.md)
- [Observability Guide](../advanced/observability.md)
- [Monitoring with OpenTelemetry](../advanced/observability.md)
