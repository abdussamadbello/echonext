---
name: echonext-database
description: >-
  Work with EchoNext's database layer: GORM models, the generic Repository[T]
  pattern from pkg/contrib/database, and Atlas migrations/seeds via the CLI. Use
  when defining models, querying/persisting data, or creating and applying
  migrations in an echonext project.
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext Database

EchoNext uses GORM for persistence and Atlas for migrations. The optional
`pkg/contrib/database` package adds a generic, type-safe `Repository[T]` so you
don't rewrite CRUD per model.

## Models

Plain GORM structs (see the `echonext-domain` skill for the conventional layout):

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Email     string         `gorm:"unique;not null"`
    Name      string         `gorm:"not null"`
}

func (User) TableName() string { return "users" }
```

## Repository[T]

`pkg/contrib/database` provides a generic repository. Construct it with the
model type and a `*gorm.DB`:

```go
import "github.com/abdussamadbello/echonext/pkg/contrib/database"

repo := database.NewRepository[User](db)
```

Core methods (from the `Repository[T]` interface):

```go
repo.Create(&user)            // INSERT
user, err := repo.Find(1)     // by primary key -> (*User, error)
users, err := repo.FindAll()  // -> ([]*User, error)
repo.Update(&user)            // UPDATE
repo.Delete(1)                // DELETE by id
count, err := repo.Count()    // -> (int64, error)
```

Chainable query builders return the repository so you can compose, then call a
terminal method:

```go
// All active users, newest first, page 2 of 20.
recent, err := repo.
    Where("active = ?", true).
    Order("created_at DESC").
    Limit(20).
    Offset(20).
    FindAll()

// Run inside a transaction:
err := db.Transaction(func(tx *gorm.DB) error {
    return repo.WithTx(tx).Create(&user)
})
```

Available builders: `Where(query, args...)`, `Order(value)`, `Limit(n)`,
`Offset(n)`, `WithTx(tx)`; `DB()` returns the underlying `*gorm.DB` for anything
the repository doesn't cover.

> Use `Repository[T]` for generic CRUD. The default `echonext generate domain`
> service holds a raw `*gorm.DB` instead — either is fine; pick one per project
> and stay consistent.

## Migrations (Atlas via the CLI)

```bash
echonext db init             # one-time: set up Atlas migration files
echonext db migrate:diff add_users  # diff Go models -> a new migration
echonext db migrate          # apply pending migrations
echonext db migrate:status   # what's applied / pending
echonext db migrate:down     # roll back
echonext db migrate:new add_index   # hand-write an empty migration
echonext db migrate:lint     # check migrations for issues
echonext db schema:inspect   # dump the current DB schema
```

Typical loop: edit `model.go` → `db migrate:diff <name>` → review the generated
SQL → `db migrate`.

## Seeding

```bash
echonext db seed   # insert sample/seed data
```

For loading fixtures in **tests**, use the contrib testing helpers instead (see
the `echonext-testing` skill), not the seed command.

## Checklist

- Models are GORM structs with a `TableName()`; use pointers/`omitempty` only
  where semantics require it.
- Prefer `database.NewRepository[T](db)` for CRUD; drop to `repo.DB()` for
  bespoke queries.
- After any model change, run `db migrate:diff <name>` then `db migrate`.
- Review generated migration SQL before applying it.
