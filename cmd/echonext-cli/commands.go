package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newGenerateCmd returns the generate command with subcommands
func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate code from templates",
		Long: `Generate various components for your EchoNext project.

Available generators:
  domain      - Complete domain with model, service, handler, and DTOs
  handler     - HTTP handler for a domain
  service     - Business service layer
  model       - GORM model
  dto         - Request/Response DTOs
  middleware  - Custom middleware
  otel        - OpenTelemetry instrumentation setup
  openapi     - Generate code from OpenAPI specification
  graphql     - Generate GraphQL boilerplate with gqlgen

Examples:
  echonext generate domain user
  echonext generate handler product
  echonext generate otel
  echonext generate openapi api.yaml
  echonext generate graphql`,
		Aliases: []string{"gen", "g"},
	}

	// Add subcommands
	cmd.AddCommand(
		newGenerateDomainCmd(),
		newGenerateHandlerCmd(),
		newGenerateServiceCmd(),
		newGenerateModelCmd(),
		newGenerateDTOCmd(),
		newGenerateMiddlewareCmd(),
		newGenerateOtelCmd(),
		newGenerateOpenAPICmd(),
		newGenerateGraphQLCmd(),
	)

	return cmd
}

// newDBCmd returns the database command with subcommands
func newDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
		Long: `Manage database operations using Atlas for migrations.

Atlas is used for schema management and migrations.
See https://atlasgo.io for more information.

Examples:
  echonext db init              - Initialize Atlas migration setup
  echonext db migrate           - Apply pending migrations
  echonext db migrate:status    - Show migration status
  echonext db migrate:new NAME  - Create a new migration
  echonext db migrate:diff NAME - Generate migration from schema changes
  echonext db migrate:down      - Rollback last migration
  echonext db seed              - Seed database with sample data`,
	}

	cmd.AddCommand(
		newDBInitCmd(),
		newDBMigrateCmd(),
		newDBMigrateStatusCmd(),
		newDBMigrateNewCmd(),
		newDBMigrateDiffCmd(),
		newDBMigrateDownCmd(),
		newDBMigrateLintCmd(),
		newDBSchemaInspectCmd(),
		newDBSeedCmd(),
	)

	return cmd
}

// newDevCmd returns the development server command
func newDevCmd() *cobra.Command {
	var port int
	var watch bool
	var target string

	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start development server with hot reload",
		Long: `Start the API server in development mode with hot reload.

The server will automatically restart when Go files change.

Examples:
  echonext dev                    # Start API server on port 8080
  echonext dev --port=3000        # Start on port 3000
  echonext dev --target=worker    # Start worker instead of api
  echonext dev --watch=false      # Run without watching`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDev(port, watch, target)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port")
	cmd.Flags().BoolVarP(&watch, "watch", "w", true, "Enable file watching")
	cmd.Flags().StringVarP(&target, "target", "t", "api", "Build target (api, worker, cli, migration)")

	return cmd
}

// newDocsCmd returns the documentation command
func newDocsCmd() *cobra.Command {
	var output string
	var format string

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate project documentation",
		Long: `Generate documentation for your EchoNext project.

Supports multiple output formats:
  html     - HTML documentation
  markdown - Markdown files
  pdf      - PDF documentation (requires pandoc)

Examples:
  echonext docs
  echonext docs --format=markdown --output=./docs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("📚 Generating %s documentation to %s...\n", format, output)
			// TODO: Implement documentation generation
			return fmt.Errorf("docs command not yet implemented")
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "./docs", "Output directory")
	cmd.Flags().StringVarP(&format, "format", "f", "html", "Output format (html, markdown, pdf)")

	return cmd
}

// newTestCmd returns the test command
func newTestCmd() *cobra.Command {
	var coverage bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "test [flags] [-- additional go test flags]",
		Short: "Run tests with enhanced features",
		Long: `Run tests with additional features like coverage reporting and parallel execution.

Examples:
  echonext test                    # Run all tests
  echonext test --coverage         # Run with coverage report
  echonext test --verbose          # Run with verbose output
  echonext test -- -run TestUser   # Pass flags to go test`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTest(coverage, verbose, args)
		},
	}

	cmd.Flags().BoolVarP(&coverage, "coverage", "c", false, "Generate coverage report")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

// newBuildCmd returns the build command
func newBuildCmd() *cobra.Command {
	var target string
	var output string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build project for production",
		Long: `Build the project for production deployment.

Available targets:
  all        - Build all executables (default)
  api        - Build only API server
  worker     - Build only background worker
  cli        - Build only CLI tool
  migration  - Build only migration tool

Examples:
  echonext build
  echonext build --target=api
  echonext build --output=./dist`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(target, output)
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "all", "Build target (all, api, worker, cli, migration)")
	cmd.Flags().StringVarP(&output, "output", "o", "./bin", "Output directory")

	return cmd
}

// Generate subcommands
func newGenerateDomainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "domain [name]",
		Short: "Generate a complete domain",
		Long: `Generate a complete domain with model, service, handler, DTOs, and tests.

Example:
  echonext generate domain user`,
		Args: cobra.ExactArgs(1),
		Run:  runGenerateDomain,
	}
}

func newGenerateHandlerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "handler [name]",
		Short: "Generate HTTP handler",
		Args:  cobra.ExactArgs(1),
		Run:   runGenerateHandler,
	}
}

func newGenerateServiceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "service [name]",
		Short: "Generate business service",
		Args:  cobra.ExactArgs(1),
		Run:   runGenerateService,
	}
}

func newGenerateModelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "model [name]",
		Short: "Generate GORM model",
		Args:  cobra.ExactArgs(1),
		Run:   runGenerateModel,
	}
}

func newGenerateDTOCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dto [name]",
		Short: "Generate request/response DTOs",
		Args:  cobra.ExactArgs(1),
		Run:   runGenerateDTO,
	}
}

func newGenerateMiddlewareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "middleware [name]",
		Short: "Generate custom middleware",
		Args:  cobra.ExactArgs(1),
		Run:   runGenerateMiddleware,
	}
}

func newGenerateOtelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "otel",
		Short: "Generate OpenTelemetry instrumentation setup",
		Long: `Generate OpenTelemetry configuration and initialization code.

Creates internal/otel/otel.go with:
  - OTEL configuration struct
  - Initialization helpers (Init, MustInit)
  - Default configuration with environment variable support
  - Traced HTTP client factory

Example:
  echonext generate otel

Usage in your main.go:
  shutdown := otel.MustInit(ctx, otel.DefaultConfig())
  defer shutdown()

  app.Use(middleware.OTELMiddleware("your-service"))`,
		Run: runGenerateOtel,
	}
}

func newGenerateOpenAPICmd() *cobra.Command {
	var outputDir string
	var packageName string
	var generateTests bool

	cmd := &cobra.Command{
		Use:   "openapi [spec-file]",
		Short: "Generate code from OpenAPI specification",
		Long: `Generate Go code (models, handlers, DTOs, routes) from an OpenAPI specification.

Supports OpenAPI 3.0 and 3.1 specifications in YAML or JSON format.
Supports $ref resolution for external files and URLs.

Generated files:
  - models/models.go     - Data models from schema components
  - dto/dto.go           - Request/Response DTOs
  - handlers/handlers.go - Handler function stubs
  - routes.go            - Route registration using EchoNext

Examples:
  echonext generate openapi api.yaml
  echonext generate openapi api.yaml --output=./generated
  echonext generate openapi api.yaml --package=api
  echonext generate openapi https://api.example.com/openapi.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateOpenAPI(args[0], outputDir, packageName, generateTests)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "./generated", "Output directory")
	cmd.Flags().StringVarP(&packageName, "package", "p", "api", "Go package name")
	cmd.Flags().BoolVar(&generateTests, "tests", true, "Generate test files")

	return cmd
}

func newGenerateGraphQLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "graphql",
		Short: "Generate GraphQL boilerplate with gqlgen",
		Long: `Generate GraphQL boilerplate files for gqlgen integration.

Creates:
  - gqlgen.yml          - gqlgen configuration
  - graph/schema.graphqls - Base GraphQL schema
  - graph/resolver.go   - Resolver base struct
  - graph/model/        - Model directory for generated types
  - tools/tools.go      - Tool dependency for go generate

After generation:
  1. Install gqlgen: go get github.com/99designs/gqlgen
  2. Edit graph/schema.graphqls with your schema
  3. Run: go run github.com/99designs/gqlgen generate
  4. Implement resolvers in graph/*.resolvers.go

Example:
  echonext generate graphql

The generated code integrates with EchoNext's GraphQL support:
  app.GraphQL(echonext.GraphQLConfig{
      Schema: graph.NewExecutableSchema(graph.Config{
          Resolvers: graph.NewResolver(),
      }),
  })`,
		Run: runGenerateGraphQL,
	}
}

// Database subcommands
func newDBInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Atlas migration setup",
		Long: `Initialize Atlas migration configuration for your project.

Creates:
  - atlas.hcl     - Atlas configuration file
  - schema.hcl    - Database schema definition
  - migrations/   - Migration directory with initial migration

Example:
  echonext db init`,
		Run: runDBInit,
	}
}

func newDBMigrateCmd() *cobra.Command {
	var dryRun bool
	var env string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending migrations",
		Long: `Apply all pending migrations using Atlas.

Example:
  echonext db migrate
  echonext db migrate --dry-run
  echonext db migrate --env=production`,
		Run: func(cmd *cobra.Command, args []string) {
			runDBMigrate(cmd, args, dryRun, env)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview migrations without applying")
	cmd.Flags().StringVar(&env, "env", "local", "Atlas environment (local, staging, production)")

	return cmd
}

func newDBMigrateStatusCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "migrate:status",
		Short: "Show migration status",
		Long: `Display the current migration status.

Example:
  echonext db migrate:status`,
		Run: func(cmd *cobra.Command, args []string) {
			runDBMigrateStatus(cmd, args, env)
		},
	}

	cmd.Flags().StringVar(&env, "env", "local", "Atlas environment")

	return cmd
}

func newDBMigrateNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:new [name]",
		Short: "Create a new empty migration file",
		Long: `Create a new empty migration file.

Example:
  echonext db migrate:new add_users_table`,
		Args: cobra.ExactArgs(1),
		Run:  runDBMigrateNew,
	}
}

func newDBMigrateDiffCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "migrate:diff [name]",
		Short: "Generate migration from schema changes",
		Long: `Generate a new migration by comparing schema.hcl to the current database.

This is the recommended way to create migrations when using declarative schema.

Example:
  echonext db migrate:diff add_email_index`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDBMigrateDiff(cmd, args, env)
		},
	}

	cmd.Flags().StringVar(&env, "env", "local", "Atlas environment")

	return cmd
}

func newDBMigrateDownCmd() *cobra.Command {
	var count int
	var env string

	cmd := &cobra.Command{
		Use:   "migrate:down",
		Short: "Rollback migrations",
		Long: `Rollback the last N migrations (default: 1).

Example:
  echonext db migrate:down
  echonext db migrate:down --count=2`,
		Run: func(cmd *cobra.Command, args []string) {
			runDBMigrateDown(cmd, args, count, env)
		},
	}

	cmd.Flags().IntVar(&count, "count", 1, "Number of migrations to rollback")
	cmd.Flags().StringVar(&env, "env", "local", "Atlas environment")

	return cmd
}

func newDBMigrateLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:lint",
		Short: "Lint migrations for issues",
		Long: `Check migrations for common issues like destructive changes.

Example:
  echonext db migrate:lint`,
		Run: runDBMigrateLint,
	}
}

func newDBSchemaInspectCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "schema:inspect",
		Short: "Inspect current database schema",
		Long: `Inspect and display the current database schema.

Example:
  echonext db schema:inspect`,
		Run: func(cmd *cobra.Command, args []string) {
			runDBSchemaInspect(cmd, args, env)
		},
	}

	cmd.Flags().StringVar(&env, "env", "local", "Atlas environment")

	return cmd
}

func newDBSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed database with sample data",
		Run:   runDBSeed,
	}
}