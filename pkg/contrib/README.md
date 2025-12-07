# EchoNext Contrib Packages

Optional helper packages for EchoNext applications. These packages are completely optional - you can use the libraries directly if you prefer, or use these helpers to reduce boilerplate code.

## 📦 Available Packages

### 1. Database (`pkg/contrib/database`)

GORM integration helpers for database operations.

**Features:**
- Connection management with retry logic
- Generic Repository[T] pattern with CRUD operations
- Transaction helpers (WithTx, WithTxResult)
- Migration utilities
- Connection pool configuration

**Example:**

```go
import (
    "github.com/abdussamadbello/echonext/pkg/contrib/database"
    "gorm.io/driver/postgres"
)

// Connect to database
cfg := database.DefaultConfig()
cfg.DSN = "postgres://user:pass@localhost/mydb"
db, err := database.Connect(postgres.Open(cfg.DSN), cfg)
if err != nil {
    log.Fatal(err)
}

// Use repository pattern
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"not null"`
    Email string `gorm:"unique;not null"`
}

userRepo := database.NewRepository[User](db)

// Create
user := &User{Name: "John", Email: "john@example.com"}
err = userRepo.Create(user)

// Find
user, err = userRepo.Find(1)

// Find all with conditions
users, err := userRepo.Where("name LIKE ?", "John%").FindAll()

// Update
user.Name = "John Doe"
err = userRepo.Update(user)

// Delete
err = userRepo.Delete(1)

// Use transactions
err = database.WithTx(db, func(tx *gorm.DB) error {
    repo := userRepo.WithTx(tx)
    return repo.Create(&user)
})

// Auto-migrate
err = database.AutoMigrate(db, &User{})
```

### 2. Config (`pkg/contrib/config`)

Viper integration helpers for configuration management.

**Features:**
- Generic config loading with Load[T]
- Environment variable binding
- Hot reload with Watch[T]
- Standard config structures (AppConfig, DatabaseConfig, etc.)
- Multiple config file formats (YAML, JSON, TOML)

**Example:**

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/config"

// Define your config structure using standard types
type MyConfig struct {
    App      config.AppConfig      `mapstructure:"app"`
    Database config.DatabaseConfig `mapstructure:"database"`
    Cache    config.CacheConfig    `mapstructure:"cache"`
    Logger   config.LoggerConfig   `mapstructure:"logger"`
}

// Load from file
var cfg MyConfig
if err := config.LoadSimple(&cfg); err != nil {
    log.Fatal(err)
}

// Load with environment variable prefix
// Environment variables: MYAPP_APP_PORT, MYAPP_DATABASE_DSN, etc.
if err := config.LoadWithEnv(&cfg, "MYAPP"); err != nil {
    log.Fatal(err)
}

// Load from specific file
if err := config.LoadFromFile(&cfg, "./config.yaml"); err != nil {
    log.Fatal(err)
}

// Watch for changes
watcher, err := config.Watch(&cfg, config.DefaultLoadOptions(), func(updated *MyConfig) {
    log.Println("Config reloaded!")
    // Update your services with new config
})
```

**Config file example (config.yaml):**

```yaml
app:
  name: "myapp"
  version: "1.0.0"
  environment: "development"
  port: 8080
  debug: true

database:
  driver: "postgres"
  dsn: "postgres://user:pass@localhost/mydb?sslmode=disable"
  auto_migrate: true
  log_queries: true

cache:
  driver: "redis"
  address: "localhost:6379"
  default_ttl: 3600

logger:
  level: "info"
  format: "json"
  output: "stdout"
```

### 3. Testing (`pkg/contrib/testing`)

Testing utilities for EchoNext applications.

**Features:**
- APIClient for testing HTTP endpoints
- FixtureManager for managing test data
- Factory pattern for creating test entities
- Suite base class with setup/teardown
- IntegrationSuite with transaction rollback

**Example:**

```go
import (
    "testing"
    "github.com/abdussamadbello/echonext"
    echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"
)

func TestUserAPI(t *testing.T) {
    app := echonext.New()
    // Register your routes...

    client := echonexttest.NewAPIClient(app)

    // Test GET request
    resp := client.GET("/users/1")
    resp.AssertStatus(t, 200)

    var user User
    if err := resp.JSON(&user); err != nil {
        t.Fatal(err)
    }

    // Test POST request
    newUser := CreateUserRequest{Name: "John", Email: "john@example.com"}
    resp = client.POST("/users", newUser)
    resp.AssertStatus(t, 201).AssertSuccess(t)

    // Test with authentication
    resp = client.WithAuth("token123").GET("/protected")
    resp.AssertStatus(t, 200)
}

func TestWithFixtures(t *testing.T) {
    fixtures := echonexttest.NewFixtureManager(db)
    defer fixtures.Clear()

    // Load test data
    fixtures.Load(
        &User{ID: 1, Name: "Alice", Email: "alice@example.com"},
        &User{ID: 2, Name: "Bob", Email: "bob@example.com"},
    )

    // Run your tests...
}

func TestWithSuite(t *testing.T) {
    suite := echonexttest.NewSuite(app, db)

    // Load fixtures
    suite.LoadFixtures(
        &User{Name: "Test User", Email: "test@example.com"},
    )

    // Test with suite client
    resp := suite.Client.GET("/users")
    resp.AssertStatus(t, 200)

    // Assert database state
    suite.AssertRecordExists(t, &User{}, "email = ?", "test@example.com")
    suite.AssertRecordCount(t, &User{}, 1)

    // Cleanup happens automatically
}

func TestFactory(t *testing.T) {
    // Create a factory for generating test users
    userFactory := echonexttest.NewFactory(db, func() User {
        return User{
            Name:  "Test User",
            Email: fmt.Sprintf("user%d@example.com", rand.Int()),
        }
    })

    // Create single user
    user, err := userFactory.Create()

    // Create multiple users
    users, err := userFactory.CreateMany(10)
}
```

## 🎯 Philosophy

The contrib packages follow these principles:

1. **Completely Optional** - You can use the underlying libraries directly
2. **Zero Lock-in** - No vendor lock-in, use what you need
3. **Type-Safe** - Leverage Go generics for type safety
4. **Minimal Overhead** - Thin wrappers that don't hide complexity
5. **Best Practices** - Encourage good patterns without forcing them

## 📦 Installation

These packages are part of the EchoNext repository:

```bash
go get github.com/abdussamadbello/echonext
```

Then import what you need:

```go
import (
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/pkg/contrib/database"
    "github.com/abdussamadbello/echonext/pkg/contrib/config"
    echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"
)
```

## 🔧 Requirements

Each package has its own dependencies:

**Database:**
- `gorm.io/gorm`
- Database driver (e.g., `gorm.io/driver/postgres`)

**Config:**
- `github.com/spf13/viper`
- `github.com/fsnotify/fsnotify`

**Testing:**
- `github.com/abdussamadbello/echonext`
- `gorm.io/gorm` (optional, for database testing)

## 🧪 Testing

Run tests for all contrib packages:

```bash
go test ./pkg/contrib/...
```

Run tests for a specific package:

```bash
go test ./pkg/contrib/database
go test ./pkg/contrib/config
go test ./pkg/contrib/testing
```

## 📝 License

Same as EchoNext core - MIT License
