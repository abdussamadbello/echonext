---
name: echonext-domain
description: >-
  Scaffold and wire up a complete EchoNext domain (model + service + handler +
  DTO) following the framework's conventional layout. Use when adding a new
  resource, feature, or entity to an echonext project (e.g. "add a users
  endpoint", "create a products domain").
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext Domains

A "domain" is the framework's unit of feature organization: one Go package
under `domain/<name>/` containing a GORM model, a service (business logic), an
HTTP handler (type-safe routes), and DTOs (request/response structs). **Always
scaffold with the generator, then fill in the TODOs** — don't hand-write the
package from scratch.

## 1. Generate

```bash
echonext generate domain user
```

This creates (package `user`):

```
domain/user/
  model.go     # GORM model: User{ ID, CreatedAt, UpdatedAt, DeletedAt, ... }
  service.go   # Service{ db *gorm.DB } with Create/GetByID/List/Update/Delete/Count
  handler.go   # Handler{ service *Service } with RegisterRoutes + typed handlers
  dto.go       # CreateUserRequest / UpdateUserRequest / ListUsersRequest / UserResponse
```

To regenerate just one layer, use `generate model|service|handler|dto <name>`
(see the `echonext-cli` skill).

## 2. The conventional shape

**Model** (`model.go`) — a GORM struct with a `TableName()`:

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    // TODO: add fields, e.g.
    Email string `gorm:"unique;not null"`
    Name  string `gorm:"not null"`
}

func (User) TableName() string { return "users" }
```

**Service** (`service.go`) — owns `*gorm.DB` and the business logic:

```go
type Service struct{ db *gorm.DB }
func NewService(db *gorm.DB) *Service { return &Service{db: db} }
// Create / GetByID(id uint) / List(limit, offset int) / Update / Delete / Count
```

**Handler** (`handler.go`) — type-safe routes that call the service. It exposes
`RegisterRoutes(app *echonext.App)`:

```go
type Handler struct{ service *Service }
func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) RegisterRoutes(app *echonext.App) {
    app.POST("/users", h.Create, echonext.Route{Summary: "Create user", Tags: []string{"Users"}})
    app.GET("/users/:id", h.Get,    echonext.Route{Summary: "Get user",    Tags: []string{"Users"}})
    app.GET("/users", h.List,       echonext.Route{Summary: "List users",  Tags: []string{"Users"}})
    app.PUT("/users/:id", h.Update, echonext.Route{Summary: "Update user", Tags: []string{"Users"}})
    app.DELETE("/users/:id", h.Delete, echonext.Route{Summary: "Delete user", Tags: []string{"Users"}})
}

func (h *Handler) Create(c *echo.Context, req CreateUserRequest) (UserResponse, error) { /* ... */ }
```

The handler methods are ordinary echonext handlers — see the
`echonext-handlers` skill for the signature rules, binding, validation, and
`Response[T]` wrapping.

**DTOs** (`dto.go`) — request structs carry `validate:` tags; response structs
shape the JSON output. Map model → response in the service or handler.

## 3. Wire it into the app

Construct the layers and register routes in `main.go`:

```go
app := echonext.New()

userSvc := user.NewService(db)
userHandler := user.NewHandler(userSvc)
userHandler.RegisterRoutes(app)
```

## 4. Track the schema

After editing `model.go`, regenerate migrations so the database matches:

```bash
echonext db migrate:diff add_users
echonext db migrate
```

## Checklist

- Scaffold with `echonext generate domain <name>` (singular, lowercase).
- Fill in model fields + `validate:` tags on the request DTOs.
- Implement business rules in the service; keep handlers thin.
- Call `<name>.NewService(db)` → `NewHandler(svc)` → `RegisterRoutes(app)`.
- Regenerate migrations with `db migrate:diff` after model changes.

For the GORM `Repository[T]` alternative to the hand-rolled `db *gorm.DB`
service, see the `echonext-database` skill.
