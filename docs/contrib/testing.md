# Testing

EchoNext provides optional testing utilities through the `pkg/contrib/testing` package. These utilities make it easy to write comprehensive tests for your EchoNext applications.

## Overview

The testing package provides:

- **APIClient** - Fluent HTTP client for testing API endpoints
- **Response** - Rich response object with assertions
- **Suite** - Base test suite with common setup/teardown
- **IntegrationSuite** - Suite with transaction rollback support
- **FixtureManager** - Manage test data in the database
- **Factory** - Generic factory pattern for creating test entities

## Installation

```bash
go get github.com/abdussamadbello/echonext/pkg/contrib/testing
```

## APIClient

The `APIClient` provides a convenient, fluent interface for testing HTTP endpoints.

### Basic Usage

```go
import (
    "testing"
    "github.com/abdussamadbello/echonext"
    echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"
)

func TestAPI(t *testing.T) {
    app := echonext.New()
    // ... register routes ...

    client := echonexttest.NewAPIClient(app)

    // Make a GET request
    resp := client.GET("/users")

    // Check response
    if resp.Status() != 200 {
        t.Errorf("Expected 200, got %d", resp.Status())
    }
}
```

### HTTP Methods

```go
client := echonexttest.NewAPIClient(app)

// GET request
resp := client.GET("/users")

// POST request with body
resp := client.POST("/users", map[string]string{
    "name":  "John",
    "email": "john@example.com",
})

// PUT request with body
resp := client.PUT("/users/123", map[string]string{
    "name": "John Updated",
})

// PATCH request with body
resp := client.PATCH("/users/123", map[string]string{
    "email": "new@example.com",
})

// DELETE request
resp := client.DELETE("/users/123")
```

### Authentication

```go
// Bearer token authentication
client := echonexttest.NewAPIClient(app).WithAuth("my-jwt-token")
resp := client.GET("/protected")

// Basic authentication
client := echonexttest.NewAPIClient(app).WithBasicAuth("user", "pass")
resp := client.GET("/protected")
```

### Custom Headers

```go
client := echonexttest.NewAPIClient(app).
    WithHeader("X-API-Key", "secret").
    WithHeader("X-Custom", "value")

resp := client.GET("/api/data")
```

### Chaining

```go
client := echonexttest.NewAPIClient(app).
    WithAuth("token123").
    WithHeader("X-Request-ID", "test-123").
    WithHeader("Accept-Language", "en-US")

resp := client.POST("/orders", orderRequest)
```

## Response

The `Response` object provides methods for inspecting and asserting on HTTP responses.

### Basic Methods

```go
resp := client.GET("/users")

// Get status code
status := resp.Status()

// Get response body as string
body := resp.String()

// Check for errors
if err := resp.Error(); err != nil {
    t.Fatal(err)
}
```

### JSON Parsing

```go
resp := client.GET("/users")

// Parse into struct
var result echonext.Response[[]User]
err := resp.JSON(&result)
if err != nil {
    t.Fatal(err)
}

fmt.Println(result.Data[0].Name)
```

### Status Checks

```go
resp := client.GET("/users")

// Check if successful (2xx)
if resp.IsSuccess() {
    // Handle success
}

// Check if error (4xx or 5xx)
if resp.IsError() {
    // Handle error
}
```

### Headers

```go
resp := client.GET("/users")

// Get specific header
contentType := resp.GetHeader("Content-Type")
requestID := resp.GetHeader("X-Request-ID")
```

### Assertions

Use built-in assertions for cleaner tests:

```go
resp := client.GET("/users")

// Assert status code
resp.AssertStatus(t, 200)

// Assert successful response (2xx)
resp.AssertSuccess(t)

// Assert error response (4xx or 5xx)
resp.AssertError(t)

// Assert JSON content matches
expected := echonext.Response[[]User]{
    Success: true,
    Data: []User{
        {ID: "1", Name: "John", Email: "john@example.com"},
    },
}
resp.AssertJSON(t, expected)
```

### Chained Assertions

```go
client.GET("/users").
    AssertStatus(t, 200).
    AssertSuccess(t)

client.POST("/users", invalidData).
    AssertStatus(t, 400).
    AssertError(t)
```

## Test Suites

### Basic Suite

The `Suite` provides a base for organizing tests with common setup/teardown:

```go
import (
    "testing"
    "github.com/abdussamadbello/echonext"
    echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"
)

type UserSuite struct {
    *echonexttest.Suite
}

func NewUserSuite(app *echonext.App, db *gorm.DB) *UserSuite {
    return &UserSuite{
        Suite: echonexttest.NewSuite(app, db),
    }
}

func (s *UserSuite) Setup() error {
    // Load test fixtures
    return s.LoadFixtures(
        &User{ID: 1, Name: "Test User", Email: "test@example.com"},
    )
}

func (s *UserSuite) Teardown() error {
    return s.Fixtures.Clear()
}

func TestUserAPI(t *testing.T) {
    app, db := setupTestApp()
    suite := NewUserSuite(app, db)

    suite.Setup()
    defer suite.Teardown()

    // Use the client
    resp := suite.Client.GET("/users")
    resp.AssertSuccess(t)
}
```

### IntegrationSuite

The `IntegrationSuite` wraps each test in a transaction that rolls back:

```go
type UserIntegrationSuite struct {
    *echonexttest.IntegrationSuite
}

func NewUserIntegrationSuite(app *echonext.App, db *gorm.DB) *UserIntegrationSuite {
    return &UserIntegrationSuite{
        IntegrationSuite: echonexttest.NewIntegrationSuite(app, db),
    }
}

func TestUserIntegration(t *testing.T) {
    app, db := setupTestApp()
    suite := NewUserIntegrationSuite(app, db)

    // Start transaction
    suite.BeginTx()
    defer suite.RollbackTx() // Always rolls back

    // Load fixtures (within transaction)
    suite.LoadFixtures(&User{Name: "Test"})

    // Run test
    resp := suite.Client.POST("/users", createUserRequest)
    resp.AssertSuccess(t)

    // Verify record exists
    suite.AssertRecordExists(t, &User{}, "name = ?", "Test")

    // Transaction rolls back after test
}
```

### Suite Helper Methods

```go
// Check if record exists
suite.AssertRecordExists(t, &User{}, "email = ?", "john@example.com")

// Check if record does NOT exist
suite.AssertRecordNotExists(t, &User{}, "email = ?", "deleted@example.com")

// Check record count
suite.AssertRecordCount(t, &User{}, 5) // Expect 5 users
suite.AssertRecordCount(t, &User{}, 2, "active = ?", true) // Expect 2 active users

// Use authenticated client
authClient := suite.WithAuth("admin-token")
resp := authClient.GET("/admin/users")
```

## FixtureManager

Manage test data in the database:

```go
fixtures := echonexttest.NewFixtureManager(db)

// Load individual records
err := fixtures.Load(
    &User{Name: "John", Email: "john@example.com"},
    &User{Name: "Jane", Email: "jane@example.com"},
)

// Load multiple records of same type
users := []User{
    {Name: "User 1"},
    {Name: "User 2"},
}
err := fixtures.LoadMany(&users)

// Clear all loaded fixtures
err := fixtures.Clear()

// Clear specific table
err := fixtures.ClearTable("users")

// Clear ALL tables (use with caution)
err := fixtures.ClearAll()
```

### Retrieving Fixtures

```go
// Get loaded fixture by table and index
user, err := fixtures.Get("users", 0)

// Get count of loaded fixtures
count := fixtures.Count("users")
```

## Factory Pattern

Use factories for creating test entities with sensible defaults:

```go
import echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"

// Create a factory
userFactory := echonexttest.NewFactory(db, func() User {
    return User{
        Name:  "Default Name",
        Email: fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()),
        Active: true,
    }
})

// Create and persist entity
user, err := userFactory.Create()

// Create multiple entities
users, err := userFactory.CreateMany(5)

// Build without persisting
user := userFactory.Build()
users := userFactory.BuildMany(10)
```

### Factory with Overrides

```go
// Override specific fields
user, err := userFactory.
    With("Name", "Custom Name").
    With("Active", false).
    Create()
```

## Complete Example

Here's a complete test file demonstrating all features:

```go
package user_test

import (
    "testing"

    "github.com/abdussamadbello/echonext"
    echonexttest "github.com/abdussamadbello/echonext/pkg/contrib/testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func setupTestApp() (*echonext.App, *gorm.DB) {
    app := echonext.New()
    db := setupTestDB()

    // Register routes
    app.GET("/users", func(c echo.Context) ([]User, error) {
        var users []User
        db.Find(&users)
        return users, nil
    })

    app.POST("/users", func(c echo.Context, req CreateUserRequest) (User, error) {
        user := User{Name: req.Name, Email: req.Email}
        db.Create(&user)
        return user, nil
    })

    return app, db
}

func TestListUsers(t *testing.T) {
    app, db := setupTestApp()
    client := echonexttest.NewAPIClient(app)
    fixtures := echonexttest.NewFixtureManager(db)

    // Setup
    fixtures.Load(
        &User{Name: "John", Email: "john@example.com"},
        &User{Name: "Jane", Email: "jane@example.com"},
    )
    defer fixtures.Clear()

    // Test
    resp := client.GET("/users")

    // Assert
    resp.AssertStatus(t, 200)
    resp.AssertSuccess(t)

    var result echonext.Response[[]User]
    require.NoError(t, resp.JSON(&result))
    assert.Len(t, result.Data, 2)
}

func TestCreateUser(t *testing.T) {
    app, db := setupTestApp()
    suite := echonexttest.NewIntegrationSuite(app, db)

    suite.BeginTx()
    defer suite.RollbackTx()

    // Test
    resp := suite.Client.POST("/users", CreateUserRequest{
        Name:  "New User",
        Email: "new@example.com",
    })

    // Assert
    resp.AssertStatus(t, 200)
    suite.AssertRecordExists(t, &User{}, "email = ?", "new@example.com")
}

func TestCreateUserValidation(t *testing.T) {
    app, _ := setupTestApp()
    client := echonexttest.NewAPIClient(app)

    // Test with missing required fields
    resp := client.POST("/users", map[string]string{
        "name": "", // Empty name
    })

    // Assert
    resp.AssertStatus(t, 400)
    resp.AssertError(t)
}

func TestAuthenticatedEndpoint(t *testing.T) {
    app, _ := setupTestApp()

    // Test without auth
    client := echonexttest.NewAPIClient(app)
    resp := client.GET("/admin/users")
    resp.AssertStatus(t, 401)

    // Test with auth
    authClient := client.WithAuth("valid-token")
    resp = authClient.GET("/admin/users")
    resp.AssertStatus(t, 200)
}

func TestUserFactory(t *testing.T) {
    _, db := setupTestApp()

    factory := echonexttest.NewFactory(db, func() User {
        return User{
            Name:  "Factory User",
            Email: fmt.Sprintf("factory-%d@example.com", time.Now().UnixNano()),
        }
    })

    // Create 10 test users
    users, err := factory.CreateMany(10)
    require.NoError(t, err)
    assert.Len(t, users, 10)

    // Clean up
    for _, u := range users {
        db.Delete(u)
    }
}
```

## Best Practices

### Test Organization

1. **Use suites** for related tests that share setup
2. **Use IntegrationSuite** for database tests to ensure isolation
3. **Clear fixtures** in teardown to prevent test pollution

### Assertions

1. **Prefer built-in assertions** (`AssertStatus`, `AssertSuccess`) for clarity
2. **Chain assertions** where it makes sense
3. **Use `require`** for critical assertions that should stop the test

### Fixtures

1. **Create minimal fixtures** - only what's needed for the test
2. **Use factories** for complex entities with many fields
3. **Always clean up** fixtures after tests

### API Testing

1. **Test both success and error cases**
2. **Test validation** with invalid inputs
3. **Test authentication** for protected endpoints
4. **Verify response bodies** not just status codes

## See Also

- [API Development Guide](../guides/api-development.md)
- [Example Projects](../examples/README.md)
- [Validation Guide](../guides/validation.md)
