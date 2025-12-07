# EchoNext Example Projects

This directory contains example projects demonstrating how to use EchoNext CLI to build production-ready APIs.

## 📚 Available Examples

### 0. Quickstart (Running Example)
**Complexity:** Beginner
**Path:** `example/main.go` (at repository root)
**Type:** Working code example

A complete, runnable Todo API demonstrating all EchoNext features. Great for understanding how everything works together.

**Run it:**
```bash
# From repository root
go run example/main.go

# Visit http://localhost:8080/api/docs
```

See [quickstart/](quickstart/) for detailed documentation.

---

### 1. Todo List API (Build Your Own)
**Complexity:** Beginner
**Path:** `examples/todo-api/`
**Features:**
- Simple CRUD operations
- Single domain (todos)
- Basic validation
- OpenAPI documentation

**Build it yourself:**
```bash
# Initialize project
echonext init todo-api --module=github.com/yourusername/todo-api

cd todo-api

# Generate todo domain
echonext generate domain todo

# Initialize database
echonext db init

# Run the API
go mod tidy
go run ./cmd/api
```

### 2. Blog API (Medium Complexity)
**Complexity:** Intermediate
**Path:** `examples/blog-api/`
**Features:**
- Multiple domains (posts, comments, users)
- Domain relationships
- Authentication patterns
- File uploads (images)
- Search and filtering

**Build it yourself:**
```bash
# Initialize project
echonext init blog-api --module=github.com/yourusername/blog-api

cd blog-api

# Generate domains
echonext generate domain post
echonext generate domain comment
echonext generate domain user
echonext generate domain category

# Add custom middleware
echonext generate middleware auth
echonext generate middleware ratelimit

# Initialize database
echonext db init

# Run the API
go mod tidy
go run ./cmd/api
```

### 3. E-commerce API (Advanced)
**Complexity:** Advanced
**Path:** `examples/ecommerce-api/`
**Features:**
- Complex domain model (products, orders, payments, inventory)
- Transaction handling
- Payment integration patterns
- Order workflow
- Admin panel endpoints
- Real-time inventory tracking

**Build it yourself:**
```bash
# Initialize project
echonext init ecommerce-api --module=github.com/yourusername/ecommerce-api

cd ecommerce-api

# Generate all domains
echonext generate domain product
echonext generate domain category
echonext generate domain order
echonext generate domain payment
echonext generate domain inventory
echonext generate domain customer
echonext generate domain cart

# Generate middleware
echonext generate middleware auth
echonext generate middleware adminauth
echonext generate middleware payment

# Initialize database
echonext db init

# Run the API
go mod tidy
go run ./cmd/api
```

### 4. Microservice Template
**Complexity:** Expert
**Path:** `examples/microservice/`
**Features:**
- Service-to-service communication
- Event-driven architecture
- Message queue integration
- Distributed tracing
- Service discovery patterns
- Health checks and monitoring

**Build it yourself:**
```bash
# Initialize services
echonext init user-service --module=github.com/yourusername/user-service
echonext init order-service --module=github.com/yourusername/order-service
echonext init notification-service --module=github.com/yourusername/notification-service

# Each service follows the same pattern
cd user-service
echonext generate domain user
echonext db init
```

### 5. OpenTelemetry Demo (Distributed Tracing)
**Complexity:** Intermediate
**Path:** `examples/otel-demo/`
**Type:** Working code example
**Features:**
- OTEL initialization and configuration
- Automatic incoming request tracing
- Traced outgoing HTTP requests
- Span events and error recording
- Request ID correlation
- Trace context propagation

**Run it:**
```bash
# Start Jaeger (trace viewer)
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one

# Run the example
go run examples/otel-demo/main.go

# Visit http://localhost:8080/api/docs for API
# Visit http://localhost:16686 for Jaeger UI
```

**Add OTEL to your project:**
```bash
# Generate OTEL setup in existing project
echonext generate otel

# Then use in your main.go
shutdown := otel.MustInit(ctx, otel.DefaultConfig())
defer shutdown()
app.Use(middleware.OTELMiddleware("your-service"))
```

See [otel-demo/README.md](otel-demo/README.md) for detailed documentation.

## 🚀 Quick Start Workflow

### General Pattern for Any Project

```bash
# 1. Initialize project
echonext init myproject

# 2. Generate your domains
echonext generate domain <entity>  # Generates model, service, handler, DTOs

# 3. Or generate components individually
echonext generate model <entity>
echonext generate service <entity>
echonext generate handler <entity>
echonext generate dto <entity>

# 4. Add custom middleware
echonext generate middleware <name>

# 5. Setup database
echonext db init

# 6. Run your API
go mod tidy
go run ./cmd/api
```

## 📁 Generated Project Structure

When you run `echonext init myproject`, you get:

```
myproject/
├── cmd/
│   ├── api/        # HTTP API server
│   ├── worker/     # Background worker
│   ├── cli/        # CLI tool
│   └── migration/  # Database migrations
├── domain/         # Business domains (generated with echonext generate)
├── internal/
│   ├── config/     # Configuration
│   ├── database/   # Database setup
│   ├── middleware/ # Custom middleware
│   └── server/     # Server setup
├── infrastructure/ # Docker, deployment
├── configs/        # Config files
├── migrations/     # SQL migrations (after db init)
└── tests/          # Tests
```

## 🎯 Domain Generation Example

When you run `echonext generate domain user`:

```
domain/user/
├── model.go    # GORM model with timestamps
├── service.go  # Business logic (CRUD operations)
├── handler.go  # EchoNext HTTP handlers
└── dto.go      # Request/Response DTOs
```

All files are ready to use with:
- ✅ Type-safe handlers
- ✅ Automatic validation
- ✅ OpenAPI documentation
- ✅ Error handling
- ✅ CRUD operations

## 🔧 Customization

After generation, customize the TODO comments in:

1. **model.go** - Add your database fields
2. **service.go** - Implement business logic
3. **handler.go** - Add custom endpoints
4. **dto.go** - Define request/response structures

## 📖 Learn More

- [Core Package Documentation](../README.md)
- [CLI Tool Guide](../cmd/echonext-cli/README.md)
- [Contrib Packages](../pkg/contrib/README.md)
- [Testing Guide](../pkg/contrib/testing/doc.go)

## 💡 Tips

1. **Start Simple**: Begin with the Todo API example
2. **Use Contrib Packages**: Leverage database, config, and testing helpers
3. **Follow Domain-Driven Design**: One domain = one business entity
4. **Generate, Then Customize**: Use CLI for scaffolding, then add business logic
5. **Test Early**: Use the testing contrib package from the start

## 🤝 Contributing Examples

Have a cool example? PRs welcome!

```bash
# Create your example
echonext init myexample
# ... build something awesome ...
# Submit PR with your example/
```
