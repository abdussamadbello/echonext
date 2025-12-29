# Changelog

All notable changes to EchoNext will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2024-12-29

### Added

#### File Upload Support
- ✅ Type-safe file uploads with `upload.File` type
- ✅ `app.Upload()` method for file upload route registration
- ✅ `file.SaveTo()` method for easy file persistence
- ✅ File validation (size, type, extension)
- ✅ Multiple file upload support
- ✅ OpenAPI documentation for file endpoints
- ✅ CLI generator: `echonext generate upload [name]`
- ✅ Complete example: `examples/upload-demo/`

#### WebSocket Support
- ✅ Type-safe WebSocket handlers with `websocket.Connection`
- ✅ `app.WS()` method for WebSocket route registration
- ✅ Hub pattern for connection management and broadcasting
- ✅ `IsWebSocket` field in RouteInfo for OpenAPI support
- ✅ CLI generator: `echonext generate websocket [name]`
- ✅ Complete example: `examples/websocket-demo/`

#### GraphQL Integration
- ✅ Seamless gqlgen integration via `graphql.Config`
- ✅ `app.GraphQL()` method for GraphQL endpoint registration
- ✅ Echo context access in resolvers with `graphql.GetEchoContext()`
- ✅ GraphQL Playground integration
- ✅ Query caching with generic LRU (gqlgen v0.17.85)
- ✅ CLI generator: `echonext generate graphql`
- ✅ Complete example: `examples/graphql-demo/`

#### Developer Experience
- ✅ Hot reload development server (`echonext dev`)
- ✅ Enhanced test runner (`echonext test`)
- ✅ Build automation (`echonext build`)

#### Code Generation
- ✅ Generate from OpenAPI spec (`echonext generate openapi`)
- ✅ WebSocket handler templates (handler, hub, message)
- ✅ Upload handler templates (handler, dto)

### Changed
- Updated gqlgen dependency to v0.17.85 with generic LRU cache support

## [1.3.0] - 2024-12-09

### Added

#### Atlas Integration for Database Migrations
- ✅ Full [Atlas](https://atlasgo.io) integration for declarative schema management
- ✅ `pkg/contrib/database/atlas.go` - Go wrapper for Atlas CLI
- ✅ Declarative schema definition with `schema.hcl`
- ✅ Environment-based configuration via `atlas.hcl`
- ✅ CLI commands for migration management:
  - `echonext db init` - Initialize Atlas setup
  - `echonext db migrate` - Apply migrations with dry-run support
  - `echonext db migrate:status` - Check migration status
  - `echonext db migrate:new` - Create empty migration
  - `echonext db migrate:diff` - Generate migrations from schema changes
  - `echonext db migrate:down` - Rollback migrations
  - `echonext db migrate:lint` - Lint for destructive changes
  - `echonext db schema:inspect` - Inspect database schema

#### Comprehensive Testing Utilities
- ✅ **APIClient** - Fluent HTTP client for testing endpoints
  - Support for GET, POST, PUT, PATCH, DELETE methods
  - Bearer token and Basic authentication
  - Custom header support with method chaining
- ✅ **Response** - Rich response object with built-in assertions
  - `AssertStatus`, `AssertSuccess`, `AssertError`, `AssertJSON`
  - JSON parsing and header access
- ✅ **Comprehensive test coverage** for all contrib packages:
  - APIClient tests (all HTTP methods, auth, headers)
  - Middleware tests (RequestID, Metrics, StructuredLogger)
  - Response assertion tests

## [1.0.0] - 2024-12-01

### Added

#### Core Features
- ✅ Type-safe handler wrappers around Echo
- ✅ Automatic OpenAPI 3.0 specification generation
- ✅ Built-in request validation using struct tags
- ✅ Swagger UI integration
- ✅ Multiple handler signature support
- ✅ Full Echo middleware compatibility
- ✅ Query parameter binding with `query` tags
- ✅ Automatic response wrapping
- ✅ Comprehensive error handling

#### CLI Tool
- ✅ `echonext init` - Project initialization with complete structure
- ✅ `echonext generate domain` - Complete domain generation (model, service, handler, DTOs)
- ✅ `echonext generate handler` - HTTP handler generation
- ✅ `echonext generate service` - Service layer generation
- ✅ `echonext generate model` - GORM model generation
- ✅ `echonext generate dto` - Request/Response DTO generation
- ✅ `echonext generate middleware` - Custom middleware generation
- ✅ `echonext generate otel` - OpenTelemetry setup generation
- ✅ `echonext db init` - Database migration initialization
- ✅ `echonext db migrate` - Run database migrations
- ✅ `echonext db seed` - Seed database with test data

#### Contrib Packages
- ✅ **Database** - GORM helpers, repository pattern, and Atlas integration
  - Connection management with retry logic
  - Repository[T] with CRUD operations
  - Transaction utilities (WithTx, WithTxResult)
  - Migration helpers
  - Connection pool configuration
  - Atlas CLI wrapper for declarative migrations

- ✅ **Config** - Viper-based configuration management
  - Generic config loading with Load[T]
  - Environment variable binding
  - Hot reload support with Watch[T]
  - Standard config structures (AppConfig, DatabaseConfig, etc.)
  - Multiple config file format support

- ✅ **Testing** - Comprehensive testing utilities
  - APIClient for testing HTTP endpoints with fluent interface
  - Response object with built-in assertions
  - FixtureManager for test data management
  - Factory[T] pattern for test entity creation
  - Suite base class with setup/teardown
  - IntegrationSuite with transaction rollback
  - Full test coverage for all utilities

- ✅ **Middleware** - Additional Echo middleware helpers
  - RequestID for request correlation (with comprehensive tests)
  - Metrics collection and exposure (with comprehensive tests)
  - Structured logging with context (with comprehensive tests)
  - OpenTelemetry instrumentation
  - Traced HTTP client for outgoing requests

#### OpenTelemetry Support
- ✅ Automatic request tracing
- ✅ Traced outgoing HTTP requests
- ✅ Span events and attributes
- ✅ Request ID correlation
- ✅ Trace context propagation
- ✅ Environment-based configuration
- ✅ Integration with Jaeger, Zipkin, etc.

#### Example Projects
- ✅ Quickstart (running Todo API example)
- ✅ Todo List API (beginner example)
- ✅ Blog API (intermediate example with relationships)
- ✅ E-commerce API (advanced example with transactions)
- ✅ Microservices template (expert example)
- ✅ OpenTelemetry demo (observability example)

#### Documentation
- ✅ Comprehensive README
- ✅ Getting Started guide
- ✅ Quick Start tutorial
- ✅ Core Concepts documentation
- ✅ API Development guide
- ✅ Validation guide
- ✅ CLI tool documentation
- ✅ Example projects documentation
- ✅ Contrib packages documentation
- ✅ Architecture documentation
- ✅ Contributing guide
- ✅ FAQ
- ✅ Troubleshooting guide
- ✅ Deployment guide

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- Built on top of secure and battle-tested Echo framework
- Automatic request validation prevents injection attacks
- CORS support for cross-origin security
- Rate limiting middleware available
- Security headers middleware available

## Version History

### Release Strategy

EchoNext follows Semantic Versioning:

- **Major version** (X.0.0) - Breaking changes
- **Minor version** (0.X.0) - New features, backwards compatible
- **Patch version** (0.0.X) - Bug fixes, backwards compatible

### Upgrade Guide

When upgrading between versions, check the release notes for:
- Breaking changes
- New features
- Deprecation notices
- Migration steps

## Contributing

See our [Contributing Guide](./contributing/guide.md) for details on how to contribute to EchoNext.

## License

EchoNext is released under the MIT License. See [LICENSE](../LICENSE) file for details.
