# Roadmap

This document outlines the future development plans for EchoNext.

## Current Status (v1.0.0)

EchoNext is production-ready with comprehensive features:

✅ **Core Framework**
- Type-safe handlers with automatic OpenAPI generation
- Built-in validation and error handling
- Full Echo compatibility
- Comprehensive middleware support

✅ **CLI Tool**
- Project initialization
- Code generation (domains, handlers, services, etc.)
- Database management commands

✅ **Contrib Packages**
- Database helpers (GORM integration)
- Configuration management (Viper integration)
- Testing utilities
- Middleware helpers with OpenTelemetry support

✅ **Documentation & Examples**
- Comprehensive guides
- Multiple example projects
- Full API reference

## Short Term (Next 3-6 months)

### Developer Experience

- [ ] **Hot Reload Development Server**
  ```bash
  echonext dev --watch
  ```
  - Watch for file changes
  - Auto-rebuild and restart
  - LiveReload integration for frontend

- [ ] **Enhanced Test Runner**
  ```bash
  echonext test --coverage --verbose
  ```
  - Run tests with better reporting
  - Code coverage analysis
  - Test result formatting

- [ ] **Build Automation**
  ```bash
  echonext build --optimize --platform linux/amd64
  ```
  - Optimized production builds
  - Cross-platform compilation
  - Binary size optimization

### Code Generation

- [ ] **Custom Templates**
  - User-defined code generation templates
  - Template marketplace/sharing
  - Project-specific scaffolding

- [ ] **More Generators**
  - Repository patterns
  - Use case/interactor patterns
  - Event handlers
  - Scheduled jobs

### Documentation

- [ ] **Interactive Tutorial**
  - Step-by-step web-based tutorial
  - Interactive code examples
  - Video guides

- [ ] **API Playground**
  - Live API testing environment
  - Share and test API definitions
  - Mock server generation

## Medium Term (6-12 months)

### File Handling

- [ ] **File Upload Support in OpenAPI**
  - Automatic multipart/form-data handling
  - File validation (size, type)
  - Upload progress tracking
  - OpenAPI documentation for file endpoints

### Real-time Features

- [ ] **WebSocket Support**
  ```go
  app.WS("/chat", chatHandler, echonext.WebSocket{
      Summary: "Real-time chat",
  })
  ```
  - Type-safe WebSocket handlers
  - Event-based messaging
  - Connection management
  - Auto-reconnection support

- [ ] **Server-Sent Events (SSE)**
  ```go
  app.SSE("/events", eventsHandler, echonext.SSE{
      Summary: "Live updates",
  })
  ```
  - Type-safe SSE handlers
  - Event streaming
  - Client reconnection

### GraphQL Integration

- [ ] **GraphQL Support**
  ```go
  app.GraphQL("/graphql", schema, echonext.GraphQL{
      Summary: "GraphQL API",
  })
  ```
  - Type-safe resolvers
  - Schema generation from structs
  - GraphQL playground integration
  - Subscriptions support

### Advanced Features

- [ ] **Background Jobs**
  - Job queue integration
  - Scheduled tasks
  - Retry mechanisms
  - Job monitoring

- [ ] **Caching Layer**
  - Built-in caching middleware
  - Redis integration
  - Cache invalidation strategies
  - Response caching

## Long Term (12+ months)

### Code Generation from Specs

- [ ] **Generate from OpenAPI**
  ```bash
  echonext import openapi.yaml
  ```
  - Generate handlers from OpenAPI spec
  - Generate client SDKs
  - Keep code and spec in sync

- [ ] **Generate from Database Schema**
  ```bash
  echonext generate from-db --connection postgres://...
  ```
  - Auto-generate models from existing database
  - Create CRUD handlers
  - Relationship detection

### Advanced Middleware

- [ ] **Circuit Breaker**
  - Automatic failure detection
  - Graceful degradation
  - Recovery mechanisms

- [ ] **Advanced Rate Limiting**
  - Distributed rate limiting
  - Per-user rate limits
  - Token bucket algorithm
  - Sliding window

### Multi-Protocol Support

- [ ] **gRPC Integration**
  - Type-safe gRPC handlers
  - Automatic proto generation
  - Gateway for HTTP/JSON
  - Streaming support

- [ ] **Message Queue Integration**
  - RabbitMQ support
  - Kafka support
  - NATS support
  - Event-driven patterns

### Microservices Features

- [ ] **Service Mesh Integration**
  - Istio compatibility
  - Linkerd integration
  - Automatic service discovery

- [ ] **Distributed Tracing Enhancements**
  - Advanced span relationships
  - Trace sampling strategies
  - Performance insights
  - Error tracking

### Developer Tools

- [ ] **VS Code Extension**
  - Syntax highlighting for EchoNext patterns
  - Code snippets
  - Route visualization
  - OpenAPI preview

- [ ] **CLI Plugin System**
  - Custom command plugins
  - Community command marketplace
  - Plugin development SDK

### Testing & Quality

- [ ] **Load Testing Tools**
  ```bash
  echonext bench --duration 60s --rps 1000
  ```
  - Built-in load testing
  - Performance profiling
  - Bottleneck identification

- [ ] **Contract Testing**
  - Consumer-driven contracts
  - Provider verification
  - Contract evolution

## Community Requests

Have an idea? We want to hear it!

1. [Open an issue](https://github.com/abdussamadbello/echonext/issues/new) with the `feature-request` label
2. Describe your use case
3. Explain why it would benefit others
4. Discuss implementation approaches

## Contributing to the Roadmap

Want to help implement features?

1. Check [GitHub Issues](https://github.com/abdussamadbello/echonext/issues) for features marked `help-wanted`
2. Comment on the issue to express interest
3. Follow the [Contributing Guide](./contributing/guide.md)
4. Submit your PR!

## Feature Voting

We prioritize features based on:

- **Community demand** - GitHub issue reactions and comments
- **Use case value** - Real-world problem solving
- **Implementation complexity** - Balance of effort vs. benefit
- **Ecosystem fit** - Alignment with Go and Echo community

Vote on features by:
- 👍 Reacting to feature request issues
- 💬 Commenting with your use case
- 🌟 Starring the repository

## Versioning Strategy

- **Patch releases (1.0.x)** - Bug fixes, documentation, minor improvements
- **Minor releases (1.x.0)** - New features, backwards compatible
- **Major releases (x.0.0)** - Breaking changes, major features

## Stay Updated

- Watch the [GitHub repository](https://github.com/abdussamadbello/echonext)
- Follow release notes
- Join discussions
- Subscribe to announcements

## Questions?

- Check the [FAQ](./faq.md)
- Open a [GitHub Discussion](https://github.com/abdussamadbello/echonext/discussions)
- Read the [Contributing Guide](./contributing/guide.md)

---

**Note:** This roadmap is subject to change based on community feedback, priorities, and resources. Features may be added, removed, or rescheduled.

Last updated: December 2024
