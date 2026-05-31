---
name: echonext-testing
description: >-
  Write tests for EchoNext APIs using the pkg/contrib/testing helpers —
  APIClient for in-process HTTP requests, Suite for setup/teardown, and
  FixtureManager for seeding test data. Use when writing or fixing tests for
  echonext handlers, services, or endpoints.
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext Testing

`pkg/contrib/testing` provides in-process HTTP testing for echonext apps: build
your `*echonext.App`, drive it with an `APIClient` (no real network), and assert
on typed responses. A `Suite` bundles the app, DB, client, and fixtures with
setup/teardown.

## APIClient — request an endpoint

```go
import (
    "testing"

    echotest "github.com/abdussamadbello/echonext/pkg/contrib/testing"
)

func TestCreateTodo(t *testing.T) {
    app := buildApp() // your function that wires routes onto echonext.New()
    client := echotest.NewAPIClient(app)

    resp := client.POST("/todos", map[string]any{"title": "write tests"})
    if resp.Status() != 201 {
        t.Fatalf("got %d: %s", resp.Status(), resp.String())
    }

    var out struct {
        Data    Todo `json:"data"`
        Success bool `json:"success"`
    }
    if err := resp.JSON(&out); err != nil {
        t.Fatal(err)
    }
    if out.Data.Title != "write tests" {
        t.Fatalf("unexpected title: %q", out.Data.Title)
    }
}
```

`APIClient` methods:

- Requests: `GET(path)`, `POST(path, body)`, `PUT(path, body)`,
  `PATCH(path, body)`, `DELETE(path)` — each returns a `*Response`.
- Chainable setup: `WithHeader(key, value)`, `WithAuth(token)` (Bearer),
  `WithBasicAuth(user, pass)`.
- `*Response`: `Status() int`, `JSON(target any) error`, `String() string`,
  `Error() error`.

Remember responses are wrapped in `Response[T]` — decode into a struct with
`Data`/`Success`/`Error` fields (see the `echonext-handlers` skill).

### Authenticated requests

```go
resp := client.
    WithAuth("test-jwt").
    GET("/me")
```

## Suite — app + DB + fixtures with lifecycle

```go
func TestUsersSuite(t *testing.T) {
    app := buildApp()
    db := openTestDB(t) // *gorm.DB, e.g. in-memory sqlite

    s := echotest.NewSuite(app, db)
    if err := s.Setup(); err != nil {
        t.Fatal(err)
    }
    defer s.Teardown()

    // Seed rows for this test:
    if err := s.LoadFixtures(
        &User{ID: 1, Email: "a@example.com"},
        &User{ID: 2, Email: "b@example.com"},
    ); err != nil {
        t.Fatal(err)
    }

    resp := s.Client.GET("/users")           // shared client
    authed := s.WithAuth("token").GET("/me") // per-request auth
    _ = authed
    if resp.Status() != 200 {
        t.Fatalf("got %d", resp.Status())
    }
}
```

`Suite` fields: `App`, `DB`, `Client *APIClient`, `Fixtures *FixtureManager`.
Methods: `Setup()`, `Teardown()`, `LoadFixtures(records ...any)`,
`WithAuth(token) *APIClient`.

## FixtureManager — standalone data seeding

When you don't need a full suite:

```go
fm := echotest.NewFixtureManager(db)
fm.Load(&User{ID: 1, Email: "a@example.com"}) // insert one/several records
fm.LoadMany(users)                            // insert a slice
fm.ClearTable("users")                        // truncate one table
fm.ClearAll()                                 // clear all known fixtures
```

## Running tests

```bash
echonext test                 # framework test runner
echonext test -- -run TestX   # pass-through flags after --
go test ./...                 # or plain Go
```

## Checklist

- Build the same `*echonext.App` your `main` builds, then `NewAPIClient(app)`.
- Decode into a `{ Data, Success, Error }` wrapper, not the bare model.
- Assert on `resp.Status()`; print `resp.String()` on failure for context.
- Use a `Suite` when you need a DB + fixtures + teardown; otherwise just the
  `APIClient`.
- Seed test data with fixtures, not `echonext db seed`.
