# Handler Patterns — Full Examples

End-to-end patterns that build on the `echonext-handlers` skill.

## A complete CRUD set

```go
package main

import (
    "net/http"

    "github.com/abdussamadbello/echonext"
    "github.com/labstack/echo/v5"
)

type Todo struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}

type CreateTodoRequest struct {
    Title string `json:"title" validate:"required,min=3,max=200"`
}

type UpdateTodoRequest struct {
    ID    string `param:"id"    validate:"required"`
    Title string `json:"title"  validate:"omitempty,min=3,max=200"`
    Done  *bool  `json:"done"`
}

type ListTodosRequest struct {
    Page  int `query:"page"  validate:"min=1"`
    Limit int `query:"limit" validate:"min=1,max=100"`
}

var store = map[string]*Todo{}

func listTodos(c *echo.Context, req ListTodosRequest) ([]Todo, error) {
    out := make([]Todo, 0, len(store))
    for _, t := range store {
        out = append(out, *t)
    }
    return out, nil
}

func getTodo(c *echo.Context) (Todo, error) {
    id := c.Param("id")
    t, ok := store[id]
    if !ok {
        return Todo{}, echo.NewHTTPError(http.StatusNotFound, "not found")
    }
    return *t, nil
}

func createTodo(c *echo.Context, req CreateTodoRequest) (Todo, error) {
    t := &Todo{ID: newID(), Title: req.Title}
    store[t.ID] = t
    return *t, nil
}

func updateTodo(c *echo.Context, req UpdateTodoRequest) (Todo, error) {
    t, ok := store[req.ID]
    if !ok {
        return Todo{}, echo.NewHTTPError(http.StatusNotFound, "not found")
    }
    if req.Title != "" {
        t.Title = req.Title
    }
    if req.Done != nil {
        t.Done = *req.Done
    }
    return *t, nil
}

func deleteTodo(c *echo.Context) error {
    delete(store, c.Param("id"))
    return c.NoContent(http.StatusNoContent)
}

func register(app *echonext.App) {
    app.GET("/todos", listTodos)
    app.GET("/todos/:id", getTodo)
    app.POST("/todos", createTodo, echonext.Route{
        Summary:       "Create a todo",
        Tags:          []string{"Todos"},
        SuccessStatus: http.StatusCreated,
    })
    app.PUT("/todos/:id", updateTodo)
    app.DELETE("/todos/:id", deleteTodo)
}
```

## Two ways to read path params

Both work — pick one consistently:

```go
// A) Off the context (handler shape #1, no typed request).
func getTodo(c *echo.Context) (Todo, error) {
    id := c.Param("id")
    // ...
}

// B) As a typed field (handler shape #2) — gets you validation for free.
type GetTodoRequest struct {
    ID string `param:"id" validate:"required,uuid"`
}
func getTodo(c *echo.Context, req GetTodoRequest) (Todo, error) {
    // req.ID is already validated as a UUID
}
```

## Pagination response

Return a struct that includes both items and metadata; it's wrapped in
`Response[T]` automatically:

```go
type Page[T any] struct {
    Items []T `json:"items"`
    Page  int `json:"page"`
    Total int `json:"total"`
}

func listTodos(c *echo.Context, req ListTodosRequest) (Page[Todo], error) {
    items := paginate(req.Page, req.Limit)
    return Page[Todo]{Items: items, Page: req.Page, Total: len(store)}, nil
}
```

## Mixing raw Echo handlers

`*App` embeds `*echo.Echo`, so you can drop down to plain Echo for static
files, health checks, or anything outside the type-safe model:

```go
app.Static("/assets", "public")
app.Echo.GET("/healthz", func(c *echo.Context) error {
    return c.String(http.StatusOK, "ok")
})
app.Use(middleware.RequestLogger()) // any Echo middleware
```

Note: routes registered via raw Echo are **not** added to the generated
OpenAPI spec — only the type-safe `app.GET/POST/...` routes are.
