# Contributing to EchoNext

Thank you for your interest in contributing to EchoNext! This guide will help you get started.

## Ways to Contribute

- 🐛 **Report bugs** - Help us find and fix issues
- 💡 **Suggest features** - Share your ideas for improvements
- 📝 **Improve documentation** - Help others learn EchoNext
- 🔧 **Submit code** - Fix bugs or add features
- 🎓 **Create examples** - Share how you use EchoNext
- 💬 **Help others** - Answer questions in issues and discussions

## Getting Started

### 1. Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/echonext.git
cd echonext

# Add upstream remote
git remote add upstream https://github.com/abdussamadbello/echonext.git
```

### 2. Install Dependencies

```bash
# Install Go 1.24+
go version

# Install dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Create a Branch

```bash
# Update your fork
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/my-feature

# Or a bugfix branch
git checkout -b fix/my-bugfix
```

## Making Changes

### Code Guidelines

1. **Follow Go conventions**
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - Use meaningful variable names

2. **Write clear code**
   - Add comments for complex logic
   - Keep functions small and focused
   - Use early returns for error handling

3. **Maintain backwards compatibility**
   - Don't break existing APIs
   - Deprecate before removing
   - Document breaking changes

### Example of Good Code

```go
// CreateUser creates a new user with the provided information.
// Returns an error if the email is already in use.
func (s *Service) CreateUser(req CreateUserRequest) (*User, error) {
    // Validate email uniqueness
    exists, err := s.emailExists(req.Email)
    if err != nil {
        return nil, fmt.Errorf("checking email: %w", err)
    }
    if exists {
        return nil, echo.NewHTTPError(409, "email already in use")
    }
    
    // Create user
    user := &User{
        Name:  req.Name,
        Email: req.Email,
    }
    
    if err := s.db.Create(user).Error; err != nil {
        return nil, fmt.Errorf("creating user: %w", err)
    }
    
    return user, nil
}
```

### Testing

Always add tests for your changes:

```go
func TestCreateUser(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    service := NewService(db)
    
    // Test successful creation
    req := CreateUserRequest{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    user, err := service.CreateUser(req)
    if err != nil {
        t.Fatalf("CreateUser failed: %v", err)
    }
    
    if user.Name != req.Name {
        t.Errorf("expected name %s, got %s", req.Name, user.Name)
    }
    
    // Test duplicate email
    _, err = service.CreateUser(req)
    if err == nil {
        t.Error("expected error for duplicate email")
    }
}
```

Run tests:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestCreateUser
```

### Linting

Ensure your code passes linting:

```bash
# Run linter
golangci-lint run

# Auto-fix issues when possible
golangci-lint run --fix
```

## Submitting Changes

### 1. Commit Your Changes

Write clear commit messages:

```bash
# Good commit messages
git commit -m "feat: add user authentication middleware"
git commit -m "fix: resolve validation error for empty arrays"
git commit -m "docs: update quickstart guide with examples"

# Commit message format
# type: description
#
# Types:
# - feat: New feature
# - fix: Bug fix
# - docs: Documentation changes
# - test: Adding or updating tests
# - refactor: Code refactoring
# - perf: Performance improvements
# - chore: Maintenance tasks
```

### 2. Push to Your Fork

```bash
git push origin feature/my-feature
```

### 3. Create a Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Select your branch
4. Fill in the PR template:

```markdown
## Description
Brief description of what this PR does.

## Changes
- List of changes made
- Another change
- And another

## Testing
How to test these changes.

## Related Issues
Closes #123
```

### 4. Respond to Review

- Address reviewer comments
- Make requested changes
- Push updates to your branch
- PR will update automatically

## Types of Contributions

### Bug Fixes

1. **Create an issue** first (unless it's a typo or obvious fix)
2. **Describe the bug**:
   - What happened
   - What you expected
   - Steps to reproduce
   - Your environment
3. **Submit a fix** with tests

### New Features

1. **Open a discussion** first for major features
2. **Explain the use case**
3. **Get feedback** before coding
4. **Submit PR** with:
   - Implementation
   - Tests
   - Documentation
   - Examples

### Documentation

Documentation improvements are always welcome!

```bash
# Edit docs
vim docs/guides/my-guide.md

# Preview locally (if using markdown viewer)
# Or just check it on GitHub after pushing

# Submit PR
git add docs/
git commit -m "docs: improve validation guide"
git push origin docs/validation-improvements
```

### Examples

Share how you use EchoNext:

1. Create a complete example project
2. Add a detailed README
3. Include comments in code
4. Submit to `examples/` directory

## Development Workflow

### Typical workflow for a bug fix:

```bash
# 1. Sync with upstream
git checkout main
git pull upstream main

# 2. Create branch
git checkout -b fix/validation-bug

# 3. Make changes
vim echonext.go

# 4. Add tests
vim echonext_test.go

# 5. Run tests
go test ./...

# 6. Lint
golangci-lint run

# 7. Commit
git add .
git commit -m "fix: resolve validation bug for nested structs"

# 8. Push
git push origin fix/validation-bug

# 9. Create PR on GitHub
```

### Typical workflow for a feature:

```bash
# 1. Open issue/discussion first
# Get feedback on the feature

# 2. Sync and branch
git checkout main
git pull upstream main
git checkout -b feature/custom-validators

# 3. Implement
vim echonext.go

# 4. Add tests
vim echonext_test.go

# 5. Add docs
vim docs/guides/validation.md

# 6. Add example
vim examples/custom-validation/main.go

# 7. Test everything
go test ./...
golangci-lint run

# 8. Commit and push
git add .
git commit -m "feat: add support for custom validators"
git push origin feature/custom-validators

# 9. Create PR with detailed description
```

## Code Review Process

### What to expect:

1. **Initial review** - Within a few days
2. **Feedback** - Suggestions and questions
3. **Discussion** - May iterate on approach
4. **Approval** - Once everything looks good
5. **Merge** - Maintainer will merge

### Good PR practices:

- Keep PRs focused and small
- One feature/fix per PR
- Include tests
- Update documentation
- Respond to feedback promptly

## Community Guidelines

- **Be respectful** - We're all here to learn and improve
- **Be patient** - Maintainers are volunteers
- **Be constructive** - Provide helpful feedback
- **Be collaborative** - Work together toward solutions

## Questions?

- Check existing [issues](https://github.com/abdussamadbello/echonext/issues)
- Ask in [discussions](https://github.com/abdussamadbello/echonext/discussions)
- Read the [documentation](../README.md)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Your contributions make EchoNext better for everyone. Thank you for taking the time to contribute! 🎉
