---
name: echonext-cli
description: >-
  Use the EchoNext CLI to scaffold projects and code, run the dev loop, build,
  test, and manage the database. Use when creating a new echonext project,
  generating code (domain/handler/service/model/dto/middleware/openapi/graphql/
  websocket/upload), or running dev/build/test/db commands.
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext CLI

The `echonext` CLI (a Cobra app under `cmd/echonext-cli`) scaffolds projects,
generates code from templates, runs a hot-reload dev server, builds, tests, and
drives Atlas database migrations. **Prefer the generators over hand-writing
boilerplate** — they emit code in the framework's conventional layout.

## Project lifecycle

```bash
echonext init <name>                 # create a new project
echonext init <name> -t microservice # template: standard|minimal|microservice|monolith
echonext init <name> -m github.com/me/app   # set the Go module name

echonext dev      # dev server with hot reload
echonext test     # run tests (pass extra go flags after --, e.g. -- -run TestX)
echonext build    # production build
echonext docs     # generate project documentation
```

## Code generation: `echonext generate <kind> [name]`

| Command | Generates |
|---------|-----------|
| `generate domain <name>` | A complete domain: model + service + handler + DTO |
| `generate handler <name>` | HTTP handler boilerplate |
| `generate service <name>` | Business-logic service |
| `generate model <name>` | GORM model |
| `generate dto <name>` | Request/response DTOs |
| `generate middleware <name>` | Custom middleware |
| `generate otel` | OpenTelemetry instrumentation setup |
| `generate openapi <spec-file>` | Code from an OpenAPI spec |
| `generate graphql` | GraphQL boilerplate (gqlgen) |
| `generate websocket <name>` | WebSocket handler boilerplate |
| `generate upload <name>` | File-upload handler boilerplate |

For adding a feature/resource, start with `generate domain` — see the
`echonext-domain` skill. For OpenAPI/auth details see
`echonext-openapi-security`; for the integration generators see
`echonext-integrations`.

## Database: `echonext db <command>` (Atlas)

```bash
echonext db init            # initialize Atlas migration setup
echonext db migrate         # apply pending migrations
echonext db migrate:status  # show migration status
echonext db migrate:new <name>   # create a new empty migration file
echonext db migrate:diff <name>  # generate a migration from model/schema changes
echonext db migrate:down    # roll back migrations
echonext db migrate:lint    # lint migrations for issues
echonext db schema:inspect  # inspect the current database schema
echonext db seed            # seed the database with sample data
```

See the `echonext-database` skill for how models and the `Repository[T]`
pattern relate to these migration commands.

## Notes

- `echonext init` creates a **new project**; `echonext db init` initializes
  **Atlas migrations** inside an existing project — they are different commands.
- The dev server (`echonext dev`) watches for changes and rebuilds; use it
  instead of `go run` while iterating.
- After `generate domain`/`model`, regenerate migrations with
  `db migrate:diff <name>` so the schema tracks your Go models.
