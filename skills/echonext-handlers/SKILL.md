---
name: echonext-handlers
description: >-
  Write and register type-safe EchoNext HTTP handlers, request structs, and
  routes. Use when adding, editing, or debugging an echonext handler function,
  a request/response struct, validation tags, or a route registration
  (GET/POST/PUT/PATCH/DELETE).
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext Handlers

EchoNext wraps Echo with type-safe handlers. You write a plain function with
typed input/output; the framework uses reflection to parse the request,
validate it, call your function, and serialize the result. There is **no**
manual `c.Bind`, `c.Validate`, or `c.JSON` in your handler.

## The three valid handler shapes

```go
// 1. No typed input — read params off the context yourself, return a value.
func(c *echo.Context) (T, error)

// 2. Typed input — req is parsed + validated for you.
func(c *echo.Context, req R) (T, error)

// 3. No response body (side effects only).
func(c *echo.Context) error
```

`T` and `R` are any of your structs. Returning `(T, error)` is the common case.

## How the request is parsed

The framework picks binding based on the HTTP method (see
`echonext.go` `createEchoHandler`):

- **GET / DELETE** → binds **query params** (`query:"..."`) and **path values**
  (`param:"..."`).
- **POST / PUT / PATCH** → binds the **JSON body** (`json:"..."`) and **path
  values** (`param:"..."`).

After binding, the struct is validated with go-playground/validator using its
`validate:"..."` tags. A validation failure returns 400 automatically — you
never see the handler called.

### Request struct examples

```go
// Body (POST/PUT/PATCH)
type CreateTodoRequest struct {
    Title       string `json:"title" validate:"required,min=3,max=200"`
    Description string `json:"description" validate:"max=1000"`
}

// Query (GET/DELETE)
type ListTodosRequest struct {
    Page  int    `query:"page"  validate:"min=1"`
    Limit int    `query:"limit" validate:"min=1,max=100"`
    Sort  string `query:"sort"  validate:"omitempty,oneof=created_at title"`
}

// Path value: route "/todos/:id" → field tagged param:"id"
type GetTodoRequest struct {
    ID string `param:"id" validate:"required"`
}
```

Common validator tags: `required`, `min=`, `max=`, `email`, `oneof=a b c`,
`omitempty`, `uuid`, `gt=`, `lte=`. Use a pointer (`*bool`, `*int`) for
optional fields that must distinguish "absent" from "zero".

## Responses are wrapped automatically

Whatever you return is wrapped in `Response[T]` (`echonext.go:131`):

```go
type Response[T any] struct {
    Data    T      `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
    Success bool   `json:"success"`
}
```

- Returning `(value, nil)` →
  `{"data": value, "success": true}` with the route's success status.
- Returning `(zero, err)` → an error response. Use
  `echo.NewHTTPError(status, msg)` to control the status code:

```go
func getTodo(c *echo.Context, req GetTodoRequest) (Todo, error) {
    todo, ok := store[req.ID]
    if !ok {
        return Todo{}, echo.NewHTTPError(http.StatusNotFound, "todo not found")
    }
    return *todo, nil
}
```

Helpers exist (`echonext.go:1487`): `echonext.Success[T](data)`,
`echonext.Error[T](msg)`, `echonext.NoContent()` — but for the common
`(T, error)` shape you just return the value and let the wrapper do the work.

## Registering routes

Methods exist on both `*App` and `*Group` with the same signature
`(path string, handler interface{}, opts ...Route)`:

```go
app := echonext.New()

app.GET("/todos", listTodos)
app.GET("/todos/:id", getTodo)
app.POST("/todos", createTodo, echonext.Route{
    Summary:       "Create a todo",
    Tags:          []string{"Todos"},
    SuccessStatus: http.StatusCreated, // default is 200
})
app.PUT("/todos/:id", updateTodo)
app.PATCH("/todos/:id", patchTodo)
app.DELETE("/todos/:id", deleteTodo)

// Groups share a prefix and middleware
v1 := app.Group("/api/v1")
v1.GET("/users", listUsers)
```

The optional `Route{}` carries OpenAPI metadata (summary, description, tags,
success status, headers, examples) and security — see the
`echonext-openapi-security` skill for auth and `Route.Security`.

## Quick checklist

- Handler is `func(c *echo.Context[, req R]) (T, error)` or
  `func(c *echo.Context) error`.
- Tags match the method: `json:` for bodies, `query:` for GET/DELETE query,
  `param:` for path values.
- Validation lives in `validate:` tags, not in the handler body.
- Return `echo.NewHTTPError(...)` for non-2xx; return the value for success.
- Don't call `c.JSON`/`c.Bind` — the wrapper handles serialization.

See [`references/handler-patterns.md`](references/handler-patterns.md) for
fuller end-to-end examples (CRUD set, pagination, mixing raw Echo handlers).
