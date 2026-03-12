# Troubleshooting Guide

This guide helps you solve common issues with EchoNext.

## Installation Issues

### "command not found: echonext"

**Problem:** CLI tool is not in your PATH.

**Solution:** Add Go's bin directory to your PATH:

```bash
# Find your GOPATH
go env GOPATH

# Add to PATH (add to ~/.bashrc or ~/.zshrc for persistence)
export PATH=$PATH:$(go env GOPATH)/bin

# Verify
echonext --version
```

### "cannot find package"

**Problem:** EchoNext package not installed or Go module not initialized.

**Solution:**

```bash
# Make sure you're in a Go module
go mod init github.com/username/project

# Install EchoNext
go get github.com/abdussamadbello/echonext

# Tidy dependencies
go mod tidy
```

### "version conflict"

**Problem:** Dependency version conflicts.

**Solution:**

```bash
# Update EchoNext
go get -u github.com/abdussamadbello/echonext

# Clean module cache
go clean -modcache

# Tidy dependencies
go mod tidy
```

## Runtime Issues

### Port Already in Use

**Problem:** Cannot bind to port 8080 (or other port).

**Error:**
```
bind: address already in use
```

**Solution:**

```bash
# Find what's using the port (Linux/Mac)
lsof -i :8080

# Kill the process
kill -9 PID

# Or use a different port in your code
app.Start(":3000")
```

### Validation Errors

**Problem:** Request validation fails unexpectedly.

**Common causes:**

1. **Missing required fields:**
```go
// Request
{"name": "John"}  // Missing 'email'

// Handler expects
type Request struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required"`  // Missing!
}
```

**Solution:** Provide all required fields or make them optional:
```go
Email string `json:"email" validate:"omitempty,email"`
```

2. **Wrong field names:**
```go
// Request uses wrong case
{"Name": "John"}  // Capital N

// Expected
{"name": "John"}  // lowercase n
```

**Solution:** Match JSON tag names exactly.

3. **Type mismatch:**
```go
// Sending string instead of number
{"age": "25"}  // String

// Expected
{"age": 25}  // Number
```

### Handler Not Found (404)

**Problem:** Request returns 404 Not Found.

**Debugging steps:**

1. **Check route path:**
```go
// Registered as
app.GET("/users/:id", handler)

// Must request with matching path
GET /users/123  // ✅ OK
GET /user/123   // ❌ 404 - wrong path
GET /users      // ❌ 404 - missing :id
```

2. **Check HTTP method:**
```go
// Registered as
app.POST("/users", handler)

// Must use POST
POST /users  // ✅ OK
GET /users   // ❌ 404 - wrong method
```

3. **Check route registration order:**
```go
// Routes must be registered before app.Start()
app.GET("/users", handler)  // ✅ OK
app.Start(":8080")
app.GET("/posts", handler)  // ❌ Never registered
```

### Empty OpenAPI Spec

**Problem:** OpenAPI spec is empty or missing routes.

**Solutions:**

1. **Call SetInfo() before registering routes:**
```go
app := echonext.New()
app.SetInfo("My API", "1.0.0", "Description")  // Before routes
app.GET("/users", handler, echonext.Route{...})
```

2. **Use echonext.Route configuration:**
```go
// ❌ No OpenAPI metadata
app.GET("/users", handler)

// ✅ With metadata
app.GET("/users", handler, echonext.Route{
    Summary: "List users",
    Tags:    []string{"Users"},
})
```

3. **Serve the spec:**
```go
app.ServeOpenAPISpec("/api/openapi.json")
app.ServeSwaggerUI("/api/docs", "/api/openapi.json")
```

### Middleware Not Working

**Problem:** Middleware doesn't seem to execute.

**Solutions:**

1. **Use middleware before routes:**
```go
app := echonext.New()
app.Use(middleware.Logger())  // Before routes
app.GET("/users", handler)
```

2. **Check middleware return:**
```go
// ❌ Doesn't call next
func myMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Do something
        return nil  // Stops here!
    }
}

// ✅ Calls next
func myMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Do something
        return next(c)  // Continue to next handler
    }
}
```

### CORS Errors

**Problem:** Browser blocks requests with CORS errors.

**Error:**
```
Access to XMLHttpRequest has been blocked by CORS policy
```

**Solution:**

```go
import "github.com/labstack/echo/v5/middleware"

app := echonext.New()

// Basic CORS (allows all origins)
app.Use(middleware.CORS())

// Custom CORS
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
}))
```

## Type Safety Issues

### "cannot use func literal as type"

**Problem:** Handler signature doesn't match expected types.

**Error:**
```
cannot use func literal (type func(echo.Context) error) as type ...
```

**Solution:** Use supported handler signatures:

```go
// ✅ Supported signatures
func(c echo.Context) (Response, error)
func(c echo.Context, req Request) (Response, error)
func(c echo.Context) error
func(c echo.Context, req Request) error

// ❌ Not supported
func(c echo.Context) Response  // Missing error
func(req Request) (Response, error)  // Missing context
```

### "validation for 'X' failed on the 'Y' tag"

**Problem:** Validation tag constraint not met.

**Examples:**

```go
// min constraint
Name string `validate:"min=3"`
// "John" ✅ OK (length 4)
// "Jo" ❌ fails (length 2)

// email constraint
Email string `validate:"email"`
// "john@example.com" ✅ OK
// "invalid-email" ❌ fails

// oneof constraint
Status string `validate:"oneof=active inactive"`
// "active" ✅ OK
// "pending" ❌ fails (not in list)
```

**Solution:** Ensure request data meets validation constraints.

## Database Issues

### "failed to connect to database"

**Problem:** Database connection fails.

**Solutions:**

1. **Check connection string:**
```go
// PostgreSQL
dsn := "host=localhost user=myuser password=mypass dbname=mydb port=5432 sslmode=disable"

// MySQL
dsn := "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True"

// SQLite
dsn := "test.db"
```

2. **Verify database is running:**
```bash
# PostgreSQL
pg_isready

# MySQL
mysqladmin ping
```

3. **Check network/firewall:**
```bash
# Test connection
telnet localhost 5432  # PostgreSQL
telnet localhost 3306  # MySQL
```

### "table doesn't exist"

**Problem:** Database table not created.

**Solution:** Run migrations:

```go
// Auto-migrate
db.AutoMigrate(&User{}, &Product{}, &Order{})

// Or use migration files
echonext db migrate
```

### "foreign key constraint fails"

**Problem:** Foreign key reference doesn't exist.

**Solution:** Ensure referenced records exist:

```go
// Create user first
user := &User{Name: "John"}
db.Create(user)

// Then create related record
order := &Order{UserID: user.ID}
db.Create(order)
```

## Performance Issues

### Slow Response Times

**Debugging:**

1. **Add logging middleware:**
```go
app.Use(middleware.Logger())
```

2. **Check database queries:**
```go
// Enable query logging
db.Debug().Find(&users)
```

3. **Profile your code:**
```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.
go tool pprof cpu.prof
```

**Common fixes:**

- Add database indexes
- Use pagination
- Cache frequently accessed data
- Use connection pooling

### High Memory Usage

**Solutions:**

1. **Limit query results:**
```go
// ❌ Loads all records
var users []User
db.Find(&users)

// ✅ Paginate
db.Limit(100).Offset(0).Find(&users)
```

2. **Close resources:**
```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()  // Important!
```

## CLI Tool Issues

### "echonext init" fails

**Problem:** Project initialization fails.

**Solutions:**

1. **Check directory doesn't exist:**
```bash
# Remove existing directory
rm -rf myproject

# Then init
echonext init myproject
```

2. **Provide module name:**
```bash
echonext init myproject --module=github.com/user/myproject
```

### "echonext generate" produces errors

**Problem:** Generated code has compilation errors.

**Solution:**

1. **Run in correct directory:**
```bash
cd myproject  # Must be in project root
echonext generate domain user
```

2. **Ensure dependencies installed:**
```bash
go mod tidy
```

3. **Update CLI tool:**
```bash
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest
```

## Testing Issues

### Tests Hang or Timeout

**Problem:** Tests don't complete.

**Solution:**

```go
// ❌ Server runs forever
func TestAPI(t *testing.T) {
    app := echonext.New()
    app.Start(":8080")  // Blocks forever!
}

// ✅ Use test server
func TestAPI(t *testing.T) {
    app := echonext.New()
    server := httptest.NewServer(app)
    defer server.Close()
    // Test with server.URL
}
```

### Test Database Issues

**Solution:** Use separate test database:

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }
    db.AutoMigrate(&User{}, &Product{})
    return db
}
```

## Common Error Messages

### "panic: runtime error: invalid memory address"

**Cause:** Nil pointer dereference.

**Solution:** Check for nil before using:

```go
// ❌ May panic
user := getUser()
fmt.Println(user.Name)

// ✅ Safe
user := getUser()
if user != nil {
    fmt.Println(user.Name)
}
```

### "context canceled"

**Cause:** Request context was canceled (client disconnected).

**Solution:** Handle gracefully:

```go
func handler(c echo.Context) error {
    select {
    case <-c.Request().Context().Done():
        return c.Request().Context().Err()
    case result := <-processChan:
        return c.JSON(200, result)
    }
}
```

## Getting More Help

### Enable Debug Logging

```go
import "github.com/labstack/echo/v5/middleware"

app := echonext.New()
app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
}))
```

### Check EchoNext Version

```bash
# CLI version
echonext --version

# Package version
go list -m github.com/abdussamadbello/echonext
```

### Still Stuck?

1. Check [FAQ](./faq.md)
2. Review [Example Projects](./examples/README.md)
3. Search [GitHub Issues](https://github.com/abdussamadbello/echonext/issues)
4. Open a [new issue](https://github.com/abdussamadbello/echonext/issues/new) with:
   - EchoNext version
   - Go version
   - Minimal reproducible example
   - Error messages
   - What you've tried
