# E-commerce API - EchoNext Example

An advanced e-commerce platform demonstrating complex domain models, transactions, and real-world business logic.

## 🎯 What You'll Learn

- Complex domain relationships
- Transaction handling
- Payment processing patterns
- Inventory management
- Order workflows
- Admin panel patterns
- Event-driven updates

## 🏗️ Build This Example

```bash
# Initialize
echonext init ecommerce-api --module=github.com/yourusername/ecommerce-api
cd ecommerce-api

# Generate all domains
echonext generate domain product
echonext generate domain category
echonext generate domain order
echonext generate domain orderitem
echonext generate domain payment
echonext generate domain inventory
echonext generate domain customer
echonext generate domain cart

# Generate middleware
echonext generate middleware auth
echonext generate middleware adminauth
echonext generate middleware ratelimit

# Setup database
echonext db init

# Run
go mod tidy
go run ./cmd/api
```

## 📡 API Endpoints

### Products
- `GET /products` - List products (search, filter, sort)
- `GET /products/:id` - Get product details
- `POST /admin/products` - Create product (admin)
- `PUT /admin/products/:id` - Update product (admin)
- `DELETE /admin/products/:id` - Delete product (admin)

### Cart
- `GET /cart` - Get cart (auth required)
- `POST /cart/items` - Add item to cart
- `PUT /cart/items/:id` - Update cart item quantity
- `DELETE /cart/items/:id` - Remove from cart
- `POST /cart/checkout` - Proceed to checkout

### Orders
- `POST /orders` - Create order (from cart)
- `GET /orders` - List customer orders
- `GET /orders/:id` - Get order details
- `POST /orders/:id/cancel` - Cancel order
- `GET /admin/orders` - List all orders (admin)
- `PUT /admin/orders/:id/status` - Update order status (admin)

### Payments
- `POST /payments` - Process payment
- `GET /payments/:id` - Get payment status
- `POST /payments/:id/refund` - Refund payment (admin)

### Inventory
- `GET /admin/inventory` - Check inventory levels (admin)
- `PUT /admin/inventory/:product_id` - Update stock (admin)
- `GET /admin/inventory/low-stock` - Low stock alerts (admin)

## 🔑 Key Features

### 1. Complex Transactions
```go
func (s *OrderService) CreateOrder(cartID uint) (*Order, error) {
    return database.WithTxResult(s.db, func(tx *gorm.DB) (*Order, error) {
        // 1. Get cart items
        cart, err := s.getCart(tx, cartID)

        // 2. Check inventory
        if err := s.checkInventory(tx, cart.Items); err != nil {
            return nil, err
        }

        // 3. Create order
        order := s.createOrderFromCart(cart)
        if err := tx.Create(&order).Error; err != nil {
            return nil, err
        }

        // 4. Reserve inventory
        if err := s.reserveInventory(tx, order.Items); err != nil {
            return nil, err
        }

        // 5. Clear cart
        if err := tx.Delete(&cart).Error; err != nil {
            return nil, err
        }

        return &order, nil
    })
}
```

### 2. Inventory Management
```go
func (s *InventoryService) ReserveStock(productID uint, quantity int) error {
    return database.WithTx(s.db, func(tx *gorm.DB) error {
        var inventory Inventory
        // Lock row for update
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&inventory, "product_id = ?", productID).Error; err != nil {
            return err
        }

        if inventory.Available < quantity {
            return ErrInsufficientStock
        }

        inventory.Available -= quantity
        inventory.Reserved += quantity

        return tx.Save(&inventory).Error
    })
}
```

### 3. Order State Machine
```go
type OrderStatus string

const (
    OrderPending    OrderStatus = "pending"
    OrderProcessing OrderStatus = "processing"
    OrderShipped    OrderStatus = "shipped"
    OrderDelivered  OrderStatus = "delivered"
    OrderCancelled  OrderStatus = "cancelled"
)

func (s *OrderService) UpdateStatus(orderID uint, newStatus OrderStatus) error {
    // Validate state transition
    if !s.isValidTransition(currentStatus, newStatus) {
        return ErrInvalidStateTransition
    }

    // Update and trigger events
    // ...
}
```

### 4. Payment Integration Pattern
```go
type PaymentProvider interface {
    ProcessPayment(amount float64, currency string, metadata map[string]string) (*PaymentResult, error)
    RefundPayment(paymentID string, amount float64) error
    GetPaymentStatus(paymentID string) (*PaymentStatus, error)
}

// Stripe implementation
type StripeProvider struct {
    apiKey string
}

// PayPal implementation
type PayPalProvider struct {
    clientID     string
    clientSecret string
}
```

## 📦 Database Schema Highlights

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    category_id INTEGER REFERENCES categories(id),
    sku VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id),
    status VARCHAR(20) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    payment_id INTEGER REFERENCES payments(id),
    shipping_address TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_at_time DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) UNIQUE,
    available INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP
);
```

## 🧪 Example Workflow

### 1. Add Items to Cart
```bash
curl -X POST http://localhost:8080/cart/items \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 123,
    "quantity": 2
  }'
```

### 2. Checkout
```bash
curl -X POST http://localhost:8080/cart/checkout \
  -H "Authorization: Bearer <token>"
```

### 3. Create Order and Process Payment
```bash
curl -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "shipping_address": "123 Main St",
    "payment_method": "stripe",
    "payment_token": "tok_visa"
  }'
```

### 4. Track Order
```bash
curl http://localhost:8080/orders/456 \
  -H "Authorization: Bearer <token>"
```

## 🎨 Advanced Patterns

### Event-Driven Architecture
```go
type OrderEvent struct {
    Type      string
    OrderID   uint
    Timestamp time.Time
    Data      map[string]interface{}
}

// Publish events for order state changes
func (s *OrderService) publishEvent(event OrderEvent) {
    // Send to message queue (Redis, RabbitMQ, Kafka)
    s.eventPublisher.Publish("orders", event)
}

// Consumers can react to events
// - Send email notifications
// - Update analytics
// - Trigger inventory updates
```

### Caching Strategy
```go
// Cache product catalog
func (s *ProductService) GetProduct(id uint) (*Product, error) {
    // Check cache first
    if product := s.cache.Get(fmt.Sprintf("product:%d", id)); product != nil {
        return product.(*Product), nil
    }

    // Fetch from database
    product, err := s.fetchFromDB(id)
    if err != nil {
        return nil, err
    }

    // Cache for 1 hour
    s.cache.Set(fmt.Sprintf("product:%d", id), product, time.Hour)

    return product, nil
}
```

## 📊 Admin Panel Features

- Order management and fulfillment
- Inventory tracking and alerts
- Product catalog management
- Customer management
- Sales analytics and reporting
- Payment reconciliation

## 🚀 Production Considerations

1. **Payment Security**: PCI compliance, tokenization
2. **Inventory Consistency**: Distributed locks, eventual consistency
3. **Order Processing**: Idempotency, retry logic
4. **Performance**: Caching, database indexing, query optimization
5. **Monitoring**: Order success rate, payment failures, stock levels

## 📚 Learn More

- [Simple Example: Todo API](../todo-api/)
- [Medium Example: Blog API](../blog-api/)
- [Microservices Example](../microservice/)
