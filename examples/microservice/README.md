# Microservices Architecture - EchoNext Example

A distributed system demonstrating how to build microservices with EchoNext, including service-to-service communication, event-driven patterns, and observability.

## 🎯 What You'll Learn

- Microservices architecture patterns
- Service-to-service communication
- Event-driven architecture
- Distributed tracing
- Service discovery
- API Gateway pattern
- Circuit breakers and resilience

## 🏗️ Architecture

```
┌─────────────┐
│  API Gateway │ ──┐
└─────────────┘   │
                  │
    ┌─────────────┼─────────────┬─────────────┐
    │             │             │             │
┌───▼────┐  ┌────▼───┐  ┌──────▼──┐  ┌──────▼─────┐
│  User  │  │ Order  │  │ Product │  │ Notification│
│Service │  │Service │  │ Service │  │  Service    │
└───┬────┘  └────┬───┘  └────┬────┘  └──────┬─────┘
    │            │            │              │
    └────────────┴────────────┴──────────────┘
                     │
              ┌──────▼──────┐
              │ Message Bus │
              │ (Redis/NATS)│
              └─────────────┘
```

## 🏗️ Build This Example

### 1. Create All Services

```bash
# User Service
echonext init user-service --module=github.com/yourusername/user-service
cd user-service
echonext generate domain user
echonext db init

# Order Service
echonext init order-service --module=github.com/yourusername/order-service
cd ../order-service
echonext generate domain order
echonext generate domain orderitem
echonext db init

# Product Service
echonext init product-service --module=github.com/yourusername/product-service
cd ../product-service
echonext generate domain product
echonext generate domain inventory
echonext db init

# Notification Service
echonext init notification-service --module=github.com/yourusername/notification-service
cd ../notification-service
echonext generate domain notification
echonext db init
```

### 2. Service Communication

Each service can communicate via:
- **HTTP** - Direct service-to-service calls
- **Events** - Asynchronous messaging
- **gRPC** - For performance-critical paths

## 🔑 Key Patterns

### 1. Service-to-Service Communication
```go
// In Order Service - call User Service
type UserServiceClient struct {
    baseURL string
}

func (c *UserServiceClient) GetUser(userID uint) (*User, error) {
    resp, err := http.Get(fmt.Sprintf("%s/users/%d", c.baseURL, userID))
    // Handle response
}
```

### 2. Event-Driven Communication
```go
// Publish events
type OrderCreatedEvent struct {
    OrderID    uint
    UserID     uint
    TotalAmount float64
}

func (s *OrderService) CreateOrder(order *Order) error {
    // Create order
    if err := s.db.Create(order).Error; err != nil {
        return err
    }

    // Publish event
    event := OrderCreatedEvent{
        OrderID:     order.ID,
        UserID:      order.UserID,
        TotalAmount: order.TotalAmount,
    }
    s.eventBus.Publish("orders.created", event)

    return nil
}

// Subscribe to events (in Notification Service)
func (s *NotificationService) Start() {
    s.eventBus.Subscribe("orders.created", func(event OrderCreatedEvent) {
        // Send notification email
        s.sendOrderConfirmation(event)
    })
}
```

### 3. Circuit Breaker Pattern
```go
import "github.com/sony/gobreaker"

type ResilientClient struct {
    client  *http.Client
    breaker *gobreaker.CircuitBreaker
}

func (c *ResilientClient) CallService(url string) (*Response, error) {
    result, err := c.breaker.Execute(func() (interface{}, error) {
        return c.client.Get(url)
    })

    if err != nil {
        return nil, err
    }

    return result.(*Response), nil
}
```

### 4. Distributed Tracing
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (h *Handler) CreateOrder(c echo.Context, req CreateOrderRequest) (OrderResponse, error) {
    ctx := c.Request().Context()
    tracer := otel.Tracer("order-service")

    ctx, span := tracer.Start(ctx, "CreateOrder")
    defer span.End()

    // Call user service (propagates trace context)
    user, err := h.userClient.GetUser(ctx, req.UserID)

    // Call product service (propagates trace context)
    product, err := h.productClient.GetProduct(ctx, req.ProductID)

    // Create order
    order, err := h.service.CreateOrder(ctx, req)

    return ToOrderResponse(order), nil
}
```

### 5. Service Discovery
```go
// Using Consul for service discovery
type ServiceRegistry struct {
    consul *consulapi.Client
}

func (r *ServiceRegistry) RegisterService(name string, port int) error {
    registration := &consulapi.AgentServiceRegistration{
        Name:    name,
        Port:    port,
        Check: &consulapi.AgentServiceCheck{
            HTTP:     fmt.Sprintf("http://localhost:%d/health", port),
            Interval: "10s",
            Timeout:  "1s",
        },
    }
    return r.consul.Agent().ServiceRegister(registration)
}

func (r *ServiceRegistry) DiscoverService(name string) (string, error) {
    services, _, err := r.consul.Health().Service(name, "", true, nil)
    if err != nil {
        return "", err
    }

    if len(services) == 0 {
        return "", fmt.Errorf("service not found: %s", name)
    }

    // Return first healthy service
    service := services[0]
    return fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port), nil
}
```

## 📡 Service Definitions

### User Service
**Port:** 8081
**Endpoints:**
- `POST /users` - Register user
- `GET /users/:id` - Get user
- `PUT /users/:id` - Update user
- `GET /users/:id/orders` - Get user orders (calls Order Service)

### Order Service
**Port:** 8082
**Endpoints:**
- `POST /orders` - Create order
- `GET /orders/:id` - Get order
- `GET /orders` - List orders

**Events Published:**
- `orders.created`
- `orders.completed`
- `orders.cancelled`

### Product Service
**Port:** 8083
**Endpoints:**
- `GET /products` - List products
- `GET /products/:id` - Get product
- `POST /products/:id/reserve` - Reserve inventory

### Notification Service
**Port:** 8084
**Endpoints:**
- `POST /notifications/send` - Send notification

**Events Consumed:**
- `orders.created` → Send order confirmation
- `orders.completed` → Send completion email
- `users.registered` → Send welcome email

## 🐳 Docker Compose Setup

```yaml
version: '3.8'

services:
  user-service:
    build: ./user-service
    ports: ["8081:8081"]
    environment:
      DATABASE_URL: postgres://user:pass@postgres/users
      REDIS_URL: redis:6379

  order-service:
    build: ./order-service
    ports: ["8082:8082"]
    environment:
      DATABASE_URL: postgres://user:pass@postgres/orders
      USER_SERVICE_URL: http://user-service:8081
      PRODUCT_SERVICE_URL: http://product-service:8083

  product-service:
    build: ./product-service
    ports: ["8083:8083"]
    environment:
      DATABASE_URL: postgres://user:pass@postgres/products

  notification-service:
    build: ./notification-service
    ports: ["8084:8084"]
    environment:
      REDIS_URL: redis:6379
      EMAIL_SMTP: smtp.example.com

  redis:
    image: redis:alpine
    ports: ["6379:6379"]

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## 🚀 Running the Example

```bash
# Start all services
docker-compose up

# Or run individually for development
cd user-service && go run ./cmd/api &
cd order-service && go run ./cmd/api &
cd product-service && go run ./cmd/api &
cd notification-service && go run ./cmd/api &
```

## 🧪 Testing the Flow

```bash
# 1. Create a user
curl -X POST http://localhost:8081/users \
  -d '{"name": "John", "email": "john@example.com"}'

# 2. Create a product
curl -X POST http://localhost:8083/products \
  -d '{"name": "Widget", "price": 29.99, "stock": 100}'

# 3. Create an order
curl -X POST http://localhost:8082/orders \
  -d '{"user_id": 1, "items": [{"product_id": 1, "quantity": 2}]}'

# 4. Check notifications were sent (logs)
docker-compose logs notification-service
```

## 📊 Observability

### Health Checks
Each service exposes:
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness check
- `GET /metrics` - Prometheus metrics

### Distributed Tracing
Using OpenTelemetry:
- Trace IDs propagate across services
- View traces in Jaeger UI
- Identify bottlenecks and failures

### Logging
Structured logging with correlation IDs:
```json
{
  "level": "info",
  "service": "order-service",
  "trace_id": "abc123",
  "msg": "Order created",
  "order_id": 456
}
```

## 🎨 Advanced Patterns

1. **Saga Pattern**: Distributed transactions across services
2. **CQRS**: Separate read and write models
3. **Event Sourcing**: Store events instead of current state
4. **API Gateway**: Single entry point with routing
5. **Sidecar Pattern**: Inject capabilities (logging, metrics, etc.)

## 📚 Learn More

- [Simple Example: Todo API](../todo-api/)
- [Medium Example: Blog API](../blog-api/)
- [Advanced Example: E-commerce API](../ecommerce-api/)
- [EchoNext Documentation](../../README.md)
