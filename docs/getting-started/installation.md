# Installation

This guide will help you install EchoNext and set up your development environment.

## Prerequisites

Before installing EchoNext, ensure you have:

- **Go 1.24.0 or later** - [Download Go](https://golang.org/dl/)
- **Git** - For version control
- A text editor or IDE (VS Code, GoLand, etc.)

## Quick Install

The fastest way to install the EchoNext CLI:

### Linux / macOS

```bash
curl -sSL https://raw.githubusercontent.com/abdussamadbello/echonext/master/install.sh | bash
```

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/abdussamadbello/echonext/master/install.ps1 | iex
```

This installs `echonext` to `~/.local/bin` (Unix) or `%LOCALAPPDATA%\bin` (Windows) and automatically configures your PATH.

## Installing EchoNext

### Install the Library

Add EchoNext to your Go project:

```bash
go get github.com/abdussamadbello/echonext
```

This will install the core EchoNext library.

### Install the CLI Tool (Optional but Recommended)

The EchoNext CLI helps you generate projects and code quickly:

```bash
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest
```

Verify the installation:

```bash
echonext --version
```

## Quick Setup

### Option 1: Start with a New Project (Recommended)

Use the CLI to create a new project with best practices:

```bash
# Create a new project
echonext init myapp --module=github.com/yourusername/myapp

# Navigate to the project
cd myapp

# Install dependencies
go mod tidy

# Run the application
go run ./cmd/api
```

Visit http://localhost:8080/api/docs to see your API documentation.

### Option 2: Add to Existing Project

Add EchoNext to an existing Go project:

```bash
# In your existing project directory
go get github.com/abdussamadbello/echonext

# Create a basic main.go
cat > main.go << 'EOF'
package main

import (
    "github.com/abdussamadbello/echonext"
)

func main() {
    app := echonext.New()
    app.SetInfo("My API", "1.0.0", "My EchoNext API")
    
    app.GET("/health", func(c *echo.Context) error {
        return c.JSON(200, map[string]string{"status": "ok"})
    })
    
    app.Start(":8080")
}
EOF

# Run it
go run main.go
```

## Installing Optional Contrib Packages

EchoNext provides optional helper packages. Install them as needed:

### Database Helpers

```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres  # or driver/mysql, driver/sqlite, etc.
```

### Config Management

```bash
go get github.com/spf13/viper
go get github.com/fsnotify/fsnotify
```

The contrib packages are already part of the EchoNext module, so you can import them directly:

```go
import (
    "github.com/abdussamadbello/echonext/pkg/contrib/database"
    "github.com/abdussamadbello/echonext/pkg/contrib/config"
    "github.com/abdussamadbello/echonext/pkg/contrib/testing"
    "github.com/abdussamadbello/echonext/pkg/contrib/middleware"
)
```

## Verifying Your Installation

Create a simple test to verify everything works:

```go
// test.go
package main

import (
    "testing"
    "github.com/abdussamadbello/echonext"
)

func TestEchoNext(t *testing.T) {
    app := echonext.New()
    if app == nil {
        t.Fatal("Failed to create EchoNext app")
    }
}
```

Run the test:

```bash
go test test.go
```

## Development Tools (Optional)

Consider installing these tools for a better development experience:

### Air - Hot Reload

```bash
go install github.com/air-verse/air@latest
```

Create `.air.toml` for hot reload:

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/api"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor"]
```

Run with hot reload:

```bash
air
```

### golangci-lint - Code Quality

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run linter:

```bash
golangci-lint run
```

## Next Steps

Now that EchoNext is installed:

1. Follow the [Quick Start Guide](./quickstart.md) to build your first API
2. Learn about [Core Concepts](./concepts.md)
3. Explore [Example Projects](../examples/README.md)

## Troubleshooting

### "command not found: echonext"

If you used the install script, ensure `~/.local/bin` is in your PATH:

```bash
# Linux/macOS
export PATH="$PATH:$HOME/.local/bin"
```

If you used `go install` directly, ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add the appropriate line to your `~/.bashrc` or `~/.zshrc` to make it permanent.

**Windows:** The install script automatically adds `%LOCALAPPDATA%\bin` to your PATH. Restart your terminal after installation.

### "package github.com/abdussamadbello/echonext: cannot find package"

Make sure you're in a Go module:

```bash
go mod init github.com/yourusername/yourproject
go get github.com/abdussamadbello/echonext
```

### Version Conflicts

If you encounter version conflicts, try:

```bash
go get -u github.com/abdussamadbello/echonext
go mod tidy
```

## Getting Help

- [Troubleshooting Guide](../troubleshooting.md)
- [FAQ](../faq.md)
- [GitHub Issues](https://github.com/abdussamadbello/echonext/issues)
