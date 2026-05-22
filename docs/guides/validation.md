# Request Validation Guide

Learn how to validate HTTP requests in EchoNext using struct tags.

## Table of Contents

- [Basics](#basics)
- [Common Validation Tags](#common-validation-tags)
- [String Validation](#string-validation)
- [Numeric Validation](#numeric-validation)
- [Array and Slice Validation](#array-and-slice-validation)
- [Nested Struct Validation](#nested-struct-validation)
- [Custom Validation](#custom-validation)
- [Error Messages](#error-messages)
- [Best Practices](#best-practices)

## Basics

EchoNext uses [go-playground/validator](https://github.com/go-playground/validator) for validation.

### Simple Example

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=18,max=120"`
}

func createUser(c *echo.Context, req CreateUserRequest) (UserResponse, error) {
    // req is already validated!
    // If validation fails, client gets 422 with error details
    user := service.Create(req)
    return user, nil
}
```

### How It Works

1. Client sends JSON request
2. EchoNext binds JSON to struct
3. EchoNext validates using `validate` tags
4. If validation fails, returns 422 with errors
5. If validation passes, calls your handler

## Common Validation Tags

### Required Fields

```go
type Request struct {
    Name string `validate:"required"`  // Must be provided
    Age  int    `validate:"required"`  // Must be provided
}
```

### Optional Fields

```go
type Request struct {
    Name  string `validate:"required"`           // Required
    Email string `validate:"omitempty,email"`    // Optional, but if provided must be email
    Phone string `json:"phone"`                  // Optional, no validation
}
```

### Multiple Constraints

```go
type Request struct {
    // Multiple constraints separated by commas
    Username string `validate:"required,min=3,max=20,alphanum"`
    Password string `validate:"required,min=8,containsany=!@#$%"`
}
```

## String Validation

### Length

```go
type Request struct {
    ShortCode  string `validate:"len=6"`        // Exactly 6 characters
    Username   string `validate:"min=3,max=20"` // Between 3 and 20
    Password   string `validate:"min=8"`        // At least 8
    Description string `validate:"max=500"`     // At most 500
}
```

### Format

```go
type Request struct {
    Email   string `validate:"email"`                    // Email format
    URL     string `validate:"url"`                      // URL format
    UUID    string `validate:"uuid"`                     // UUID format
    IP      string `validate:"ip"`                       // IP address
    IPv4    string `validate:"ipv4"`                     // IPv4 address
    IPv6    string `validate:"ipv6"`                     // IPv6 address
    MAC     string `validate:"mac"`                      // MAC address
    Hex     string `validate:"hexadecimal"`              // Hexadecimal
    Base64  string `validate:"base64"`                   // Base64 encoded
    ISBN    string `validate:"isbn"`                     // ISBN number
    CreditCard string `validate:"credit_card"`           // Credit card number
}
```

### Content

```go
type Request struct {
    Alpha      string `validate:"alpha"`        // Only letters
    Alphanum   string `validate:"alphanum"`     // Letters and numbers
    Numeric    string `validate:"numeric"`      // Only numbers
    Lowercase  string `validate:"lowercase"`    // Only lowercase
    Uppercase  string `validate:"uppercase"`    // Only uppercase
    
    // Contains specific characters
    Password string `validate:"containsany=!@#$%"`
    
    // Starts/ends with
    Code string `validate:"startswith=ABC"`
    File string `validate:"endswith=.pdf"`
}
```

### Allowed Values

```go
type Request struct {
    Status string `validate:"oneof=active inactive pending"`
    Role   string `validate:"oneof=admin user guest"`
    Type   string `validate:"oneof=type1 type2 type3"`
}
```

## Numeric Validation

### Range

```go
type Request struct {
    Age      int     `validate:"min=18,max=120"`     // Between 18 and 120
    Rating   float64 `validate:"min=0,max=5"`        // Between 0.0 and 5.0
    Quantity int     `validate:"min=1"`              // At least 1
    Discount int     `validate:"max=100"`            // At most 100
}
```

### Comparison

```go
type Request struct {
    Price     float64 `validate:"gt=0"`              // Greater than 0
    Stock     int     `validate:"gte=0"`             // Greater than or equal to 0
    Discount  int     `validate:"lt=100"`            // Less than 100
    MaxUsers  int     `validate:"lte=1000"`          // Less than or equal to 1000
}
```

### Special

```go
type Request struct {
    Port     int `validate:"min=1,max=65535"`       // Valid port number
    Percent  int `validate:"min=0,max=100"`         // Percentage
    Positive int `validate:"gt=0"`                  // Positive number
}
```

## Array and Slice Validation

### Array Length

```go
type Request struct {
    Tags       []string `validate:"min=1,max=5"`         // Between 1 and 5 items
    Categories []string `validate:"required,min=1"`      // At least 1 item
    Options    []string `validate:"max=10"`              // At most 10 items
}
```

### Element Validation

Use `dive` to validate array elements:

```go
type Request struct {
    // Each tag must be 2-20 characters
    Tags []string `validate:"required,min=1,max=5,dive,min=2,max=20"`
    
    // Each email must be valid
    Emails []string `validate:"dive,email"`
    
    // Each number must be positive
    Numbers []int `validate:"dive,gt=0"`
}
```

### Nested Arrays

```go
type Request struct {
    // Array of arrays
    Matrix [][]int `validate:"dive,dive,min=0,max=100"`
}
```

## Nested Struct Validation

### Simple Nesting

```go
type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    ZipCode string `validate:"required,numeric,len=5"`
}

type CreateUserRequest struct {
    Name    string  `validate:"required"`
    Email   string  `validate:"required,email"`
    Address Address `validate:"required"` // Validates nested struct
}
```

### Optional Nesting

```go
type Request struct {
    // Address is optional, but if provided must be valid
    Address *Address `validate:"omitempty"`
}
```

### Array of Structs

```go
type Item struct {
    Name     string  `validate:"required"`
    Quantity int     `validate:"required,min=1"`
    Price    float64 `validate:"required,gt=0"`
}

type OrderRequest struct {
    Items []Item `validate:"required,min=1,dive"`
}
```

## Custom Validation

### Register Custom Validator

```go
import "github.com/go-playground/validator/v10"

// Custom validation function
func validatePhoneNumber(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    // Your validation logic
    return len(phone) >= 10 && len(phone) <= 15
}

// Register in main.go
func main() {
    app := echonext.New()
    
    // Get validator instance
    validate := validator.New()
    
    // Register custom validation
    validate.RegisterValidation("phone", validatePhoneNumber)
    
    // Set validator
    app.Validator = validate
}

// Use in struct
type Request struct {
    Phone string `validate:"required,phone"`
}
```

### Cross-Field Validation

Validate one field against another:

```go
type Request struct {
    Password        string `validate:"required,min=8"`
    ConfirmPassword string `validate:"required,eqfield=Password"`
    
    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required,gtfield=StartDate"`
    
    MinPrice float64 `validate:"required"`
    MaxPrice float64 `validate:"required,gtefield=MinPrice"`
}
```

## Error Messages

### Default Error Format

When validation fails, EchoNext returns:

```json
{
  "success": false,
  "data": null,
  "error": "Validation failed: Name: min, Email: email, Age: min"
}
```

### Custom Error Messages

```go
import (
    "github.com/go-playground/validator/v10"
    "strings"
)

func customErrorMessage(err error) string {
    var messages []string
    
    for _, err := range err.(validator.ValidationErrors) {
        field := err.Field()
        tag := err.Tag()
        
        var message string
        switch tag {
        case "required":
            message = field + " is required"
        case "email":
            message = field + " must be a valid email"
        case "min":
            message = field + " must be at least " + err.Param() + " characters"
        case "max":
            message = field + " must be at most " + err.Param() + " characters"
        default:
            message = field + " is invalid"
        }
        
        messages = append(messages, message)
    }
    
    return strings.Join(messages, ", ")
}
```

## Validation Examples

### User Registration

```go
type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
    Terms    bool   `json:"terms" validate:"required,eq=true"`
}
```

### Product Creation

```go
type CreateProductRequest struct {
    Name        string   `json:"name" validate:"required,min=3,max=200"`
    Description string   `json:"description" validate:"required,min=10,max=1000"`
    Price       float64  `json:"price" validate:"required,gt=0"`
    SKU         string   `json:"sku" validate:"required,alphanum,len=8"`
    Stock       int      `json:"stock" validate:"required,gte=0"`
    Categories  []string `json:"categories" validate:"required,min=1,max=5,dive,min=2,max=50"`
    Tags        []string `json:"tags" validate:"max=10,dive,min=2,max=30"`
    Active      bool     `json:"active" validate:"required"`
}
```

### Search/Filter Request

```go
type SearchRequest struct {
    Query  string `query:"q" validate:"omitempty,min=2,max=100"`
    Page   int    `query:"page" validate:"min=1"`
    Limit  int    `query:"limit" validate:"min=1,max=100"`
    Sort   string `query:"sort" validate:"omitempty,oneof=name price created_at"`
    Order  string `query:"order" validate:"omitempty,oneof=asc desc"`
    Status string `query:"status" validate:"omitempty,oneof=active inactive all"`
}
```

### Date Range Request

```go
type ReportRequest struct {
    StartDate time.Time `json:"start_date" validate:"required"`
    EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
    Type      string    `json:"type" validate:"required,oneof=daily weekly monthly"`
}
```

### Address Validation

```go
type AddressRequest struct {
    Street     string `json:"street" validate:"required,min=5,max=200"`
    City       string `json:"city" validate:"required,min=2,max=100"`
    State      string `json:"state" validate:"required,len=2,uppercase"`
    ZipCode    string `json:"zip_code" validate:"required,numeric,len=5"`
    Country    string `json:"country" validate:"required,len=2,uppercase"`
    IsDefault  bool   `json:"is_default"`
}
```

## Best Practices

### 1. Be Specific

```go
// ❌ Too generic
type Request struct {
    Data string `validate:"required"`
}

// ✅ Specific constraints
type Request struct {
    Email string `validate:"required,email"`
    Phone string `validate:"required,e164"`
}
```

### 2. Use Appropriate Types

```go
// ❌ String when number is expected
type Request struct {
    Age string `validate:"required"`
}

// ✅ Use correct type
type Request struct {
    Age int `validate:"required,min=18,max=120"`
}
```

### 3. Make Optional Fields Clear

```go
// ❌ Unclear if optional
type Request struct {
    Phone string `validate:"e164"`
}

// ✅ Explicitly optional
type Request struct {
    Phone string `validate:"omitempty,e164"`
}
```

### 4. Group Related Validations

```go
type Request struct {
    // User identity
    Username string `validate:"required,min=3,max=20,alphanum"`
    Email    string `validate:"required,email"`
    
    // Security
    Password string `validate:"required,min=8,containsany=!@#$%"`
    
    // Profile
    Age      int    `validate:"required,min=18,max=120"`
    Bio      string `validate:"omitempty,max=500"`
}
```

### 5. Document Complex Validations

```go
type Request struct {
    // SKU must be exactly 8 alphanumeric characters
    SKU string `validate:"required,alphanum,len=8"`
    
    // Price must be positive and up to 2 decimal places
    Price float64 `validate:"required,gt=0"`
}
```

## Complete Validation Reference

Common tags:

- `required` - Field must be provided
- `omitempty` - Skip validation if empty
- `min=n` - Minimum length/value
- `max=n` - Maximum length/value
- `len=n` - Exact length
- `eq=value` - Equal to value
- `ne=value` - Not equal to value
- `gt=n` - Greater than
- `gte=n` - Greater than or equal
- `lt=n` - Less than
- `lte=n` - Less than or equal
- `oneof=a b c` - One of listed values
- `email` - Email format
- `url` - URL format
- `uuid` - UUID format
- `alpha` - Alphabetic only
- `alphanum` - Alphanumeric only
- `numeric` - Numeric string
- `dive` - Validate array/slice elements

See the [full list](https://pkg.go.dev/github.com/go-playground/validator/v10) in the validator documentation.

## Next Steps

- [API Development Guide](./api-development.md)
- [Error Handling Guide](./error-handling.md)
- [Testing Guide](./testing.md)
- [API Reference](../api-reference/validation.md)
