---
name: echonext-integrations
description: >-
  Add WebSocket, GraphQL, or file-upload endpoints to an EchoNext app. Use when
  building real-time/WebSocket handlers (Hub pattern), wiring a gqlgen GraphQL
  schema, or accepting multipart file uploads with validation.
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext Integrations

Beyond standard JSON handlers, EchoNext ships first-class support for three
integration types, each with its own package and route method. Scaffold any of
them with `echonext generate websocket|graphql|upload` (see `echonext-cli`),
then fill in the logic using the references below.

## WebSocket — `app.WS(...)`

Implement the `websocket.Handler` interface (connect/message/disconnect
lifecycle) and register it. A `Hub` tracks connections and broadcasts.

```go
app.WS("/ws/chat", chatHandler, echonext.Route{
    Summary: "Chat WebSocket",
    Tags:    []string{"Chat"},
})
```

The interface (note `OnMessage` takes a `messageType int`):

```go
type Handler interface {
    OnConnect(conn *websocket.Connection) error
    OnMessage(conn *websocket.Connection, messageType int, data []byte) error
    OnDisconnect(conn *websocket.Connection, err error)
}
```

Full Hub/Connection API and a chat example:
[`references/websocket.md`](references/websocket.md).

## GraphQL — `app.GraphQL(...)`

Pass a gqlgen `ExecutableSchema` via `graphql.Config`; the Echo context is
injected so resolvers can read auth/request data.

```go
app.GraphQL(graphql.Config{
    Path:                "/graphql",
    PlaygroundPath:      "/playground",
    Schema:              generated.NewExecutableSchema(generated.Config{Resolvers: r}),
    EnableIntrospection: true,
}, graphql.Route{Summary: "GraphQL API", Tags: []string{"GraphQL"}})
```

Config/Route fields and resolver context access:
[`references/graphql.md`](references/graphql.md).

## File upload — `app.Upload(...)`

Use a request struct with `*upload.File` (or `[]*upload.File`) fields tagged
`form:"..."`; the framework parses multipart, enforces `upload.Config` limits,
and validates.

```go
type AvatarRequest struct {
    File *upload.File `form:"file" validate:"required"`
}

app.Upload("/avatar", uploadAvatar, echonext.Route{
    Summary:    "Upload avatar",
    Tags:       []string{"Avatar"},
    FileConfig: &upload.Config{MaxFileSize: 5 << 20, AllowedExtensions: []string{".png", ".jpg"}},
})
```

`File` methods (`SaveTo`, `Open`, `Read`, `Extension`) and multi-file uploads:
[`references/upload.md`](references/upload.md).

## Choosing

- **WebSocket** — real-time, bidirectional (chat, live updates).
- **GraphQL** — flexible client-driven queries over a typed schema.
- **Upload** — multipart file intake with size/type validation.

All three still register routes on the same `*App` and appear in the generated
OpenAPI spec (upload as multipart; GraphQL/WS documented per their `Route`).
