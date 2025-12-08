# API Development Guide

Learn how to build production-ready APIs with EchoNext.

## Table of Contents

- [Project Setup](#project-setup)
- [Domain-Driven Design](#domain-driven-design)
- [CRUD Operations](#crud-operations)
- [Advanced Queries](#advanced-queries)
- [Authentication & Authorization](#authentication--authorization)
- [File Uploads](#file-uploads)
- [Pagination](#pagination)
- [Filtering & Search](#filtering--search)
- [API Versioning](#api-versioning)
- [Best Practices](#best-practices)

## Project Setup

### Using CLI (Recommended)

```bash
# Create new project
echonext init myapi --module=github.com/username/myapi
cd myapi

# Generate your first domain
echonext generate domain user

# Initialize database
echonext db init

# Run the API
go mod tidy
go run ./cmd/api
```

### Manual Setup

```go
package main

import (
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/pkg/contrib/database"
    "gorm.io/driver/postgres"
)

func main() {
    // Initialize database
    db, err := database.Connect(
        postgres.Open("postgres://user:pass@localhost/dbname"),
        database.DefaultConfig(),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Create app
    app := echonext.New()
    app.SetInfo("My API", "1.0.0", "Production API")
    
    // Add middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recover())
    app.Use(middleware.CORS())
    
    // Register domains
    userHandler := user.NewHandler(user.NewService(db))
    userHandler.Register(app)
    
    // Serve OpenAPI docs
    app.ServeOpenAPISpec("/api/openapi.json")
    app.ServeSwaggerUI("/api/docs", "/api/openapi.json")
    
    // Start server
    app.Start(":8080")
}
```

## Domain-Driven Design

Organize your API around business domains:

```
domain/
├── user/
│   ├── model.go      # Database model
│   ├── service.go    # Business logic
│   ├── handler.go    # HTTP endpoints
│   └── dto.go        # Request/Response types
├── product/
│   ├── model.go
│   ├── service.go
│   ├── handler.go
│   └── dto.go
└── order/
    ├── model.go
    ├── service.go
    ├── handler.go
    └── dto.go
```

### Model Layer

```go
// domain/user/model.go
package user

import (
    "gorm.io/gorm"
    "time"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Name     string `gorm:"not null" json:"name"`
    Email    string `gorm:"unique;not null" json:"email"`
    Password string `gorm:"not null" json:"-"`
    Role     string `gorm:"default:'user'" json:"role"`
}
```

### Service Layer

```go
// domain/user/service.go
package user

import (
    "errors"
    "gorm.io/gorm"
    "golang.org/x/crypto/bcrypt"
)

type Service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}

func (s *Service) Create(req CreateUserRequest) (*User, error) {
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return nil, err
    }
    
    user := &User{
        Name:     req.Name,
        Email:    req.Email,
        Password: string(hashedPassword),
    }
    
    if err := s.db.Create(user).Error; err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *Service) GetByID(id uint) (*User, error) {
    var user User
    if err := s.db.First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("user not found")
        }
        return nil, err
    }
    return &user, nil
}

func (s *Service) List(page, limit int, search string) ([]User, int64, error) {
    var users []User
    var total int64
    
    query := s.db.Model(&User{})
    
    // Apply search filter
    if search != "" {
        query = query.Where("name ILIKE ? OR email ILIKE ?", 
            "%"+search+"%", "%"+search+"%")
    }
    
    // Count total
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Paginate
    offset := (page - 1) * limit
    if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
        return nil, 0, err
    }
    
    return users, total, nil
}

func (s *Service) Update(id uint, req UpdateUserRequest) (*User, error) {
    user, err := s.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    if req.Name != nil {
        user.Name = *req.Name
    }
    if req.Email != nil {
        user.Email = *req.Email
    }
    
    if err := s.db.Save(user).Error; err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *Service) Delete(id uint) error {
    if err := s.db.Delete(&User{}, id).Error; err != nil {
        return err
    }
    return nil
}
```

### Handler Layer

```go
// domain/user/handler.go
package user

import (
    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v4"
    "strconv"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Register(app *echonext.App) {
    app.POST("/users", h.Create, echonext.Route{
        Summary:       "Create a new user",
        Description:   "Creates a new user account",
        Tags:          []string{"Users"},
        SuccessStatus: 201,
    })
    
    app.GET("/users", h.List, echonext.Route{
        Summary:     "List users",
        Description: "Returns a paginated list of users",
        Tags:        []string{"Users"},
    })
    
    app.GET("/users/:id", h.Get, echonext.Route{
        Summary:     "Get user by ID",
        Description: "Returns a single user by their ID",
        Tags:        []string{"Users"},
    })
    
    app.PUT("/users/:id", h.Update, echonext.Route{
        Summary:     "Update user",
        Description: "Updates a user's information",
        Tags:        []string{"Users"},
    })
    
    app.DELETE("/users/:id", h.Delete, echonext.Route{
        Summary:       "Delete user",
        Description:   "Deletes a user account",
        Tags:          []string{"Users"},
        SuccessStatus: 204,
    })
}

func (h *Handler) Create(c echo.Context, req CreateUserRequest) (UserResponse, error) {
    user, err := h.service.Create(req)
    if err != nil {
        return UserResponse{}, echo.NewHTTPError(500, err.Error())
    }
    return ToUserResponse(user), nil
}

func (h *Handler) List(c echo.Context, req ListUsersRequest) (ListUsersResponse, error) {
    users, total, err := h.service.List(req.Page, req.Limit, req.Search)
    if err != nil {
        return ListUsersResponse{}, echo.NewHTTPError(500, err.Error())
    }
    
    return ListUsersResponse{
        Users: ToUserResponses(users),
        Total: total,
        Page:  req.Page,
        Limit: req.Limit,
    }, nil
}

func (h *Handler) Get(c echo.Context) (UserResponse, error) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    
    user, err := h.service.GetByID(uint(id))
    if err != nil {
        return UserResponse{}, echo.NewHTTPError(404, "user not found")
    }
    
    return ToUserResponse(user), nil
}

func (h *Handler) Update(c echo.Context, req UpdateUserRequest) (UserResponse, error) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    
    user, err := h.service.Update(uint(id), req)
    if err != nil {
        return UserResponse{}, echo.NewHTTPError(500, err.Error())
    }
    
    return ToUserResponse(user), nil
}

func (h *Handler) Delete(c echo.Context) error {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    
    if err := h.service.Delete(uint(id)); err != nil {
        return echo.NewHTTPError(500, err.Error())
    }
    
    return nil
}
```

### DTO Layer

```go
// domain/user/dto.go
package user

import "time"

// Request DTOs
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

type ListUsersRequest struct {
    Page   int    `query:"page" validate:"min=1"`
    Limit  int    `query:"limit" validate:"min=1,max=100"`
    Search string `query:"search" validate:"omitempty,min=2"`
}

// Response DTOs
type UserResponse struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

type ListUsersResponse struct {
    Users []UserResponse `json:"users"`
    Total int64          `json:"total"`
    Page  int            `json:"page"`
    Limit int            `json:"limit"`
}

// Mapper functions
func ToUserResponse(user *User) UserResponse {
    return UserResponse{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Role:      user.Role,
        CreatedAt: user.CreatedAt,
    }
}

func ToUserResponses(users []User) []UserResponse {
    responses := make([]UserResponse, len(users))
    for i, user := range users {
        responses[i] = ToUserResponse(&user)
    }
    return responses
}
```

## CRUD Operations

### Create

```go
app.POST("/resources", create, echonext.Route{
    Summary:       "Create resource",
    SuccessStatus: 201,
})

func create(c echo.Context, req CreateRequest) (Response, error) {
    resource, err := service.Create(req)
    if err != nil {
        return Response{}, echo.NewHTTPError(500, err.Error())
    }
    return toResponse(resource), nil
}
```

### Read (Single)

```go
app.GET("/resources/:id", get, echonext.Route{
    Summary: "Get resource by ID",
})

func get(c echo.Context) (Response, error) {
    id := c.Param("id")
    resource, err := service.GetByID(id)
    if err != nil {
        return Response{}, echo.NewHTTPError(404, "not found")
    }
    return toResponse(resource), nil
}
```

### Read (List)

```go
app.GET("/resources", list, echonext.Route{
    Summary: "List resources",
})

func list(c echo.Context, req ListRequest) (ListResponse, error) {
    resources, total, err := service.List(req.Page, req.Limit)
    if err != nil {
        return ListResponse{}, echo.NewHTTPError(500, err.Error())
    }
    return ListResponse{
        Items: resources,
        Total: total,
        Page:  req.Page,
    }, nil
}
```

### Update

```go
app.PUT("/resources/:id", update, echonext.Route{
    Summary: "Update resource",
})

func update(c echo.Context, req UpdateRequest) (Response, error) {
    id := c.Param("id")
    resource, err := service.Update(id, req)
    if err != nil {
        return Response{}, echo.NewHTTPError(500, err.Error())
    }
    return toResponse(resource), nil
}
```

### Delete

```go
app.DELETE("/resources/:id", delete, echonext.Route{
    Summary:       "Delete resource",
    SuccessStatus: 204,
})

func delete(c echo.Context) error {
    id := c.Param("id")
    if err := service.Delete(id); err != nil {
        return echo.NewHTTPError(500, err.Error())
    }
    return nil
}
```

## Advanced Queries

### Filtering

```go
type ListRequest struct {
    Status string `query:"status" validate:"omitempty,oneof=active inactive"`
    Type   string `query:"type" validate:"omitempty,oneof=type1 type2"`
}

func (s *Service) List(req ListRequest) ([]Resource, error) {
    query := s.db.Model(&Resource{})
    
    if req.Status != "" {
        query = query.Where("status = ?", req.Status)
    }
    
    if req.Type != "" {
        query = query.Where("type = ?", req.Type)
    }
    
    var resources []Resource
    if err := query.Find(&resources).Error; err != nil {
        return nil, err
    }
    
    return resources, nil
}
```

### Sorting

```go
type ListRequest struct {
    Sort  string `query:"sort" validate:"omitempty,oneof=name created_at"`
    Order string `query:"order" validate:"omitempty,oneof=asc desc"`
}

func (s *Service) List(req ListRequest) ([]Resource, error) {
    query := s.db.Model(&Resource{})
    
    if req.Sort != "" {
        order := "asc"
        if req.Order != "" {
            order = req.Order
        }
        query = query.Order(req.Sort + " " + order)
    }
    
    var resources []Resource
    query.Find(&resources)
    return resources, nil
}
```

### Search

```go
type ListRequest struct {
    Search string `query:"search" validate:"omitempty,min=2"`
}

func (s *Service) List(req ListRequest) ([]Resource, error) {
    query := s.db.Model(&Resource{})
    
    if req.Search != "" {
        query = query.Where(
            "name ILIKE ? OR description ILIKE ?",
            "%"+req.Search+"%",
            "%"+req.Search+"%",
        )
    }
    
    var resources []Resource
    query.Find(&resources)
    return resources, nil
}
```

## Pagination

```go
type PaginatedRequest struct {
    Page  int `query:"page" validate:"min=1"`
    Limit int `query:"limit" validate:"min=1,max=100"`
}

type PaginatedResponse struct {
    Items      []Resource `json:"items"`
    Total      int64      `json:"total"`
    Page       int        `json:"page"`
    Limit      int        `json:"limit"`
    TotalPages int        `json:"total_pages"`
}

func (s *Service) List(req PaginatedRequest) (PaginatedResponse, error) {
    var total int64
    var resources []Resource
    
    // Count total
    s.db.Model(&Resource{}).Count(&total)
    
    // Calculate offset
    offset := (req.Page - 1) * req.Limit
    
    // Fetch page
    s.db.Offset(offset).Limit(req.Limit).Find(&resources)
    
    // Calculate total pages
    totalPages := int(total) / req.Limit
    if int(total)%req.Limit != 0 {
        totalPages++
    }
    
    return PaginatedResponse{
        Items:      resources,
        Total:      total,
        Page:       req.Page,
        Limit:      req.Limit,
        TotalPages: totalPages,
    }, nil
}
```

## Authentication & Authorization

### JWT Middleware

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4/middleware"
)

// Add JWT middleware
app.Use(middleware.JWTWithConfig(middleware.JWTConfig{
    SigningKey: []byte("secret"),
    Skipper: func(c echo.Context) bool {
        // Skip auth for public endpoints
        return c.Path() == "/login" || c.Path() == "/register"
    },
}))

// Protected endpoint
app.GET("/profile", getProfile, echonext.Route{
    Summary: "Get user profile",
    Security: []echonext.Security{{Type: "bearer"}},
})
```

### Role-Based Access

```go
func RequireRole(role string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            user := c.Get("user").(*jwt.Token)
            claims := user.Claims.(jwt.MapClaims)
            userRole := claims["role"].(string)
            
            if userRole != role {
                return echo.NewHTTPError(403, "insufficient permissions")
            }
            
            return next(c)
        }
    }
}

// Admin-only endpoint
app.DELETE("/users/:id", deleteUser, echonext.Route{
    Summary: "Delete user (admin only)",
}, RequireRole("admin"))
```

## File Uploads

```go
func upload(c echo.Context) (UploadResponse, error) {
    // Get file from request
    file, err := c.FormFile("file")
    if err != nil {
        return UploadResponse{}, echo.NewHTTPError(400, "file required")
    }
    
    // Validate file
    if file.Size > 10*1024*1024 { // 10MB
        return UploadResponse{}, echo.NewHTTPError(400, "file too large")
    }
    
    // Open file
    src, err := file.Open()
    if err != nil {
        return UploadResponse{}, err
    }
    defer src.Close()
    
    // Save file
    dst, err := os.Create("uploads/" + file.Filename)
    if err != nil {
        return UploadResponse{}, err
    }
    defer dst.Close()
    
    if _, err = io.Copy(dst, src); err != nil {
        return UploadResponse{}, err
    }
    
    return UploadResponse{
        Filename: file.Filename,
        Size:     file.Size,
        URL:      "/uploads/" + file.Filename,
    }, nil
}
```

## API Versioning

```go
// Create version groups
v1 := app.Group("/api/v1")
v2 := app.Group("/api/v2")

// V1 handlers
v1.GET("/users", listUsersV1, echonext.Route{...})
v1.POST("/users", createUserV1, echonext.Route{...})

// V2 handlers (with breaking changes)
v2.GET("/users", listUsersV2, echonext.Route{...})
v2.POST("/users", createUserV2, echonext.Route{...})
```

## Best Practices

1. **Use DTOs** - Separate request/response types from models
2. **Validate Everything** - Use validation tags consistently
3. **Handle Errors** - Return appropriate HTTP status codes
4. **Document Routes** - Add summaries and descriptions
5. **Use Transactions** - For operations that modify multiple records
6. **Paginate Lists** - Don't return unbounded result sets
7. **Version Your API** - Plan for breaking changes
8. **Log Requests** - Use middleware for observability
9. **Test Thoroughly** - Write tests for all endpoints
10. **Secure Endpoints** - Use authentication and authorization

## Next Steps

- [Validation Guide](./validation.md)
- [Testing Guide](./testing.md)
- [Deployment Guide](./deployment.md)
- [Example Projects](../examples/README.md)
