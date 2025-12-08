# Architecture Overview

This document explains how EchoNext works under the hood.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      EchoNext App                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │   Type-Safe  │  │   OpenAPI    │  │   Validation    │  │
│  │   Handlers   │  │  Generation  │  │     Engine      │  │
│  └──────────────┘  └──────────────┘  └─────────────────┘  │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                     Echo Framework                          │
├─────────────────────────────────────────────────────────────┤
│            Standard Library (net/http)                      │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. App Wrapper

The `App` struct wraps `*echo.Echo` and adds EchoNext functionality:

```go
type App struct {
    *echo.Echo                    // Embedded Echo instance
    spec      *openapi3.T         // OpenAPI specification
    validator *validator.Validate // Validation engine
    routes    []RouteInfo         // Route metadata
}
```

**Key responsibilities:**
- Manage OpenAPI spec
- Configure validation
- Track route metadata
- Provide type-safe route registration

### 2. Type-Safe Handler Wrapper

When you register a typed handler, EchoNext creates a wrapper:

```go
// Your handler
func createUser(c echo.Context, req CreateUserRequest) (UserResponse, error)

// EchoNext creates this wrapper
func wrapper(c echo.Context) error {
    // 1. Parse and validate request
    var req CreateUserRequest
    if err := parseAndValidate(c, &req); err != nil {
        return err
    }
    
    // 2. Call your handler
    resp, err := createUser(c, req)
    if err != nil {
        return err
    }
    
    // 3. Serialize response
    return c.JSON(getStatusCode(), wrapResponse(resp))
}
```

### 3. Request Processing Pipeline

```
HTTP Request
    ↓
Echo Middleware
    ↓
EchoNext Wrapper
    ↓
1. Bind JSON/Query params → Request struct
    ↓
2. Validate struct → Check constraints
    ↓
3. Call handler → Execute business logic
    ↓
4. Wrap response → Standard format
    ↓
5. Serialize → JSON response
    ↓
HTTP Response
```

### 4. OpenAPI Generation

OpenAPI specs are generated during route registration:

```go
func (app *App) POST(path string, handler interface{}, config Route) {
    // 1. Extract types from handler signature
    reqType := getRequestType(handler)
    respType := getResponseType(handler)
    
    // 2. Generate OpenAPI schema
    operation := &openapi3.Operation{
        Summary:     config.Summary,
        Description: config.Description,
        Tags:        config.Tags,
        RequestBody: generateRequestBody(reqType),
        Responses:   generateResponses(respType),
    }
    
    // 3. Add to spec
    app.spec.Paths[path] = &openapi3.PathItem{
        Post: operation,
    }
    
    // 4. Register with Echo
    app.Echo.POST(path, wrapHandler(handler))
}
```

## Type Inspection

EchoNext uses reflection to inspect handler signatures at registration time (not at request time):

```go
func analyzeHandler(handler interface{}) HandlerInfo {
    handlerType := reflect.TypeOf(handler)
    
    // Check function signature
    if handlerType.Kind() != reflect.Func {
        panic("handler must be a function")
    }
    
    // Extract input types
    numIn := handlerType.NumIn()
    if numIn < 1 {
        panic("handler must accept echo.Context")
    }
    
    // First param must be echo.Context
    if handlerType.In(0) != contextType {
        panic("first param must be echo.Context")
    }
    
    // Second param (if exists) is request type
    var reqType reflect.Type
    if numIn == 2 {
        reqType = handlerType.In(1)
    }
    
    // Extract output types
    var respType reflect.Type
    if handlerType.NumOut() == 2 {
        respType = handlerType.Out(0)
    }
    
    return HandlerInfo{
        RequestType:  reqType,
        ResponseType: respType,
    }
}
```

## Request Binding

### JSON Body Binding

```go
func bindJSONBody(c echo.Context, req interface{}) error {
    // 1. Read body
    body, err := ioutil.ReadAll(c.Request().Body)
    if err != nil {
        return err
    }
    
    // 2. Unmarshal JSON
    if err := json.Unmarshal(body, req); err != nil {
        return echo.NewHTTPError(400, "invalid JSON")
    }
    
    return nil
}
```

### Query Parameter Binding

```go
func bindQueryParams(c echo.Context, req interface{}) error {
    // Use reflection to find fields with `query` tag
    reqValue := reflect.ValueOf(req).Elem()
    reqType := reqValue.Type()
    
    for i := 0; i < reqType.NumField(); i++ {
        field := reqType.Field(i)
        queryTag := field.Tag.Get("query")
        
        if queryTag != "" {
            // Get query param value
            value := c.QueryParam(queryTag)
            
            // Convert and set field value
            setFieldValue(reqValue.Field(i), value)
        }
    }
    
    return nil
}
```

## Validation

EchoNext uses `go-playground/validator`:

```go
func validateRequest(req interface{}) error {
    validate := validator.New()
    
    if err := validate.Struct(req); err != nil {
        // Convert validation errors to user-friendly format
        var errors []string
        for _, err := range err.(validator.ValidationErrors) {
            errors = append(errors, fmt.Sprintf(
                "%s: %s",
                err.Field(),
                err.Tag(),
            ))
        }
        return echo.NewHTTPError(422, strings.Join(errors, ", "))
    }
    
    return nil
}
```

## Response Wrapping

All responses are wrapped in a standard format:

```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Error   string      `json:"error"`
}

func wrapResponse(data interface{}, err error) APIResponse {
    if err != nil {
        return APIResponse{
            Success: false,
            Data:    nil,
            Error:   err.Error(),
        }
    }
    
    return APIResponse{
        Success: true,
        Data:    data,
        Error:   "",
    }
}
```

## OpenAPI Schema Generation

### From Struct Types

```go
func generateSchema(t reflect.Type) *openapi3.Schema {
    schema := &openapi3.Schema{
        Type: "object",
        Properties: make(map[string]*openapi3.SchemaRef),
    }
    
    // Iterate through struct fields
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        
        // Get JSON tag
        jsonTag := field.Tag.Get("json")
        if jsonTag == "-" {
            continue
        }
        
        // Get validation tag
        validateTag := field.Tag.Get("validate")
        
        // Generate field schema
        fieldSchema := generateFieldSchema(field.Type, validateTag)
        
        // Add to properties
        schema.Properties[jsonTag] = &openapi3.SchemaRef{
            Value: fieldSchema,
        }
        
        // Mark as required if needed
        if strings.Contains(validateTag, "required") {
            schema.Required = append(schema.Required, jsonTag)
        }
    }
    
    return schema
}
```

### From Validation Tags

```go
func applyValidationConstraints(schema *openapi3.Schema, tag string) {
    // Parse validation tag
    rules := strings.Split(tag, ",")
    
    for _, rule := range rules {
        parts := strings.Split(rule, "=")
        constraint := parts[0]
        
        switch constraint {
        case "min":
            if len(parts) > 1 {
                min := parseFloat(parts[1])
                schema.Min = &min
            }
        case "max":
            if len(parts) > 1 {
                max := parseFloat(parts[1])
                schema.Max = &max
            }
        case "email":
            schema.Format = "email"
        case "url":
            schema.Format = "uri"
        }
    }
}
```

## Performance Considerations

### Minimal Runtime Overhead

- **Type inspection happens once** - During route registration, not per request
- **Reflection is minimal** - Only for binding and validation
- **No runtime code generation** - All wrappers created at startup
- **Direct Echo integration** - Minimal abstraction overhead

### Benchmarks

Typical overhead compared to raw Echo:

```
BenchmarkRawEcho       100000    10234 ns/op
BenchmarkEchoNext      95000     10789 ns/op  (~5% overhead)
```

## Design Decisions

### Why Wrap Echo?

- **Leverage maturity** - Echo is battle-tested
- **Maintain compatibility** - Use existing ecosystem
- **Focus on value-add** - Type safety + OpenAPI
- **Avoid reinventing** - Don't rewrite HTTP server

### Why Runtime Type Inspection?

- **No code generation** - Simpler workflow
- **No build step** - Standard Go build
- **Better IDE support** - Native Go types
- **Easier debugging** - No generated code

### Why Standard Response Format?

- **Consistency** - Predictable API responses
- **Error handling** - Uniform error format
- **Client simplification** - Easy to parse
- **Backwards compatible** - Can be unwrapped

## Extension Points

### Custom Binders

```go
app.Binder = &CustomBinder{
    DefaultBinder: app.Binder,
}
```

### Custom Validators

```go
validate := validator.New()
validate.RegisterValidation("custom", customFunc)
app.Validator = validate
```

### Custom Response Wrapper

```go
// Override default wrapper
app.ResponseWrapper = customWrapperFunc
```

## See Also

- [Design Philosophy](./philosophy.md)
- [Type System](./types.md)
- [OpenAPI Generation](./openapi-generation.md)
