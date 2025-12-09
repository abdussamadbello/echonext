# Database

EchoNext provides optional database utilities through the `pkg/contrib/database` package, including GORM helpers, repository pattern, and **Atlas** integration for declarative schema migrations.

## Overview

The database package provides:

- **Connection Management** - Configure and connect to databases with retry logic
- **Repository Pattern** - Generic CRUD operations with `Repository[T]`
- **Transactions** - Utilities for managing database transactions
- **Migrations** - Traditional migration helpers
- **Atlas Integration** - Declarative schema management with Atlas

## Installation

```bash
go get github.com/abdussamadbello/echonext/pkg/contrib/database
```

## Connection Management

### Basic Connection

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

config := &database.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    Database: "myapp",
    SSLMode:  "disable",
}

db, err := database.Connect(config)
if err != nil {
    log.Fatal(err)
}
```

### Connection with Pool Configuration

```go
config := &database.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    Database: "myapp",

    // Connection pool settings
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: 30 * time.Minute,
}

db, err := database.Connect(config)
```

## Repository Pattern

The generic `Repository[T]` provides type-safe CRUD operations:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

// Define your model
type User struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string `gorm:"not null"`
    Email     string `gorm:"unique;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Create repository
userRepo := database.NewRepository[User](db)

// Create
user := &User{Name: "John", Email: "john@example.com"}
err := userRepo.Create(user)

// Find by ID
user, err := userRepo.FindByID(1)

// Find all
users, err := userRepo.FindAll()

// Find with conditions
users, err := userRepo.FindWhere("email = ?", "john@example.com")

// Update
user.Name = "John Doe"
err := userRepo.Update(user)

// Delete
err := userRepo.Delete(user)
```

### Custom Queries

```go
// First matching record
user, err := userRepo.First("email = ?", "john@example.com")

// Count records
count, err := userRepo.Count("active = ?", true)

// Exists check
exists, err := userRepo.Exists("email = ?", "john@example.com")
```

## Transactions

### Using WithTx

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

err := database.WithTx(db, func(tx *gorm.DB) error {
    // All operations use the same transaction
    userRepo := database.NewRepository[User](tx)
    orderRepo := database.NewRepository[Order](tx)

    user := &User{Name: "Jane"}
    if err := userRepo.Create(user); err != nil {
        return err // Transaction rolls back
    }

    order := &Order{UserID: user.ID, Total: 100}
    if err := orderRepo.Create(order); err != nil {
        return err // Transaction rolls back
    }

    return nil // Transaction commits
})
```

### Using WithTxResult

```go
result, err := database.WithTxResult(db, func(tx *gorm.DB) (*Order, error) {
    userRepo := database.NewRepository[User](tx)
    orderRepo := database.NewRepository[Order](tx)

    user := &User{Name: "Jane"}
    if err := userRepo.Create(user); err != nil {
        return nil, err
    }

    order := &Order{UserID: user.ID, Total: 100}
    if err := orderRepo.Create(order); err != nil {
        return nil, err
    }

    return order, nil
})
```

## Atlas Integration

EchoNext integrates with [Atlas](https://atlasgo.io) for declarative, version-controlled database schema management.

### Why Atlas?

- **Declarative Schema** - Define desired state in HCL, generate migrations automatically
- **Versioned Migrations** - SQL files with checksum verification
- **Migration Linting** - Detect destructive changes before they happen
- **Multi-Environment** - Separate configs for local, staging, production

### Prerequisites

Install Atlas CLI:

```bash
# macOS
brew install ariga/tap/atlas

# Linux
curl -sSf https://atlasgo.sh | sh

# Docker
docker pull arigaio/atlas
```

### Basic Setup

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

// Check if Atlas is installed
if !database.IsAtlasInstalled() {
    fmt.Println(database.InstallAtlas())
    return
}

// Create Atlas instance
atlas := database.NewAtlas(&database.AtlasConfig{
    Dir:        "migrations",
    ConfigFile: "atlas.hcl",
    Env:        "local",
})
```

### Configuration

**AtlasConfig Options:**

```go
type AtlasConfig struct {
    Dir        string // Path to migrations directory (default: "migrations")
    URL        string // Database connection URL
    DevURL     string // Dev database URL for schema calculations
    Env        string // Atlas environment (local, staging, production)
    ConfigFile string // Path to atlas.hcl (default: "atlas.hcl")
    DryRun     bool   // Enable dry-run mode
    Verbose    bool   // Enable verbose output
}
```

**Using Defaults:**

```go
atlas := database.NewAtlas(database.DefaultAtlasConfig())
```

### Migration Operations

#### Check Status

```go
ctx := context.Background()
status, err := atlas.Status(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Current: %s\n", status.Current)
fmt.Printf("Pending: %d migrations\n", len(status.Pending))
```

#### Apply Migrations

```go
// Apply all pending migrations
err := atlas.Apply(ctx)

// Apply N migrations
err := atlas.ApplyN(ctx, 2)
```

#### Rollback Migrations

```go
// Rollback last migration
err := atlas.Down(ctx)

// Rollback N migrations
err := atlas.DownN(ctx, 3)
```

#### Generate New Migration

```go
// Generate migration from schema changes
output, err := atlas.Diff(ctx, "add_users_table")
if err != nil {
    log.Fatal(err)
}
fmt.Println(output)
```

#### Create Empty Migration

```go
// Create empty migration file
path, err := atlas.New(ctx, "add_index")
fmt.Printf("Created: %s\n", path)
```

#### Lint Migrations

```go
// Check for issues
err := atlas.Lint(ctx)
if err != nil {
    fmt.Printf("Lint issues: %v\n", err)
}
```

#### Validate Migrations

```go
// Validate migration directory
err := atlas.Validate(ctx)
```

#### Inspect Schema

```go
// Get current database schema as SQL
schema, err := atlas.SchemaInspect(ctx)
fmt.Println(schema)
```

### Schema Definition (schema.hcl)

Define your database schema declaratively:

```hcl
// schema.hcl
table "users" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }

  column "email" {
    null = false
    type = varchar(255)
  }

  column "name" {
    null = false
    type = varchar(100)
  }

  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_users_email" {
    unique  = true
    columns = [column.email]
  }
}

table "posts" {
  schema = schema.public

  column "id" {
    null = false
    type = bigserial
  }

  column "user_id" {
    null = false
    type = bigint
  }

  column "title" {
    null = false
    type = varchar(255)
  }

  column "content" {
    null = false
    type = text
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "fk_posts_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
  }
}

schema "public" {}
```

### Atlas Configuration (atlas.hcl)

Configure environments and settings:

```hcl
// atlas.hcl

// Local development
env "local" {
  src = "file://schema.hcl"
  url = "postgres://postgres:password@localhost:5432/myapp?sslmode=disable"
  dev = "docker://postgres/16/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }
}

// Staging environment
env "staging" {
  src = "file://schema.hcl"
  url = getenv("STAGING_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}

// Production with safety checks
env "production" {
  src = "file://schema.hcl"
  url = getenv("PRODUCTION_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }

  // Prevent destructive changes
  diff {
    skip {
      drop_column = true
      drop_table  = true
    }
  }
}

// Lint configuration
lint {
  destructive {
    error = true
  }

  data_depend {
    error = true
  }
}
```

### CLI Commands

EchoNext CLI provides Atlas integration commands:

```bash
# Initialize Atlas setup
echonext db init

# Apply migrations
echonext db migrate
echonext db migrate --dry-run
echonext db migrate --env=production

# Check status
echonext db migrate:status

# Create new migration
echonext db migrate:new add_posts_table

# Generate from schema changes
echonext db migrate:diff add_email_index

# Rollback
echonext db migrate:down --count=1

# Lint for issues
echonext db migrate:lint

# Inspect current schema
echonext db schema:inspect
```

### Migration Workflow

**Recommended workflow:**

1. **Modify schema.hcl** with your desired changes
2. **Generate migration**: `echonext db migrate:diff describe_change`
3. **Review** the generated SQL in `migrations/`
4. **Apply**: `echonext db migrate`

**Example:**

```bash
# Add a new column to users table in schema.hcl
# ...edit schema.hcl...

# Generate migration
echonext db migrate:diff add_phone_column

# Review generated migration
cat migrations/XXXXXX_add_phone_column.sql

# Apply
echonext db migrate
```

### Environment Variables

Configure Atlas via environment variables:

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/myapp"
export STAGING_DATABASE_URL="postgres://user:pass@staging:5432/myapp"
export PRODUCTION_DATABASE_URL="postgres://user:pass@prod:5432/myapp"
```

### Initialization Helper

Initialize a new migration directory:

```go
err := database.InitMigrationDir("migrations")
if err != nil {
    log.Fatal(err)
}
```

### Error Handling

```go
// Check Atlas installation before running commands
if err := database.EnsureAtlasInstalled(); err != nil {
    log.Fatal(err)
}

// Handle migration errors
if err := atlas.Apply(ctx); err != nil {
    // Check if it's a connection error, migration conflict, etc.
    log.Printf("Migration failed: %v", err)
}
```

### Dry Run Mode

Test migrations without applying:

```go
atlas := database.NewAtlas(&database.AtlasConfig{
    Dir:     "migrations",
    Env:     "local",
    DryRun:  true,
    Verbose: true,
})

// Shows what would be applied without making changes
err := atlas.Apply(ctx)
```

## Best Practices

### Database Migrations

1. **Always use version control** for migrations
2. **Never modify existing migrations** - create new ones
3. **Test migrations locally** before applying to staging/production
4. **Use dry-run mode** to preview changes
5. **Enable linting** to catch destructive operations

### Repository Pattern

1. **One repository per model** for clear separation
2. **Use transactions** for multi-table operations
3. **Handle errors appropriately** - check for `gorm.ErrRecordNotFound`
4. **Use proper indexing** in your models

### Connections

1. **Configure pool sizes** based on workload
2. **Set connection timeouts** to prevent hanging
3. **Use connection pooling** in production
4. **Close connections** gracefully on shutdown

## See Also

- [Atlas Documentation](https://atlasgo.io/docs)
- [GORM Documentation](https://gorm.io/docs/)
- [CLI Database Commands](../cli/overview.md#database-management)
- [Example Projects](../examples/README.md)
