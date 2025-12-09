package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AtlasConfig holds configuration for Atlas migrations
type AtlasConfig struct {
	// Dir is the path to the migrations directory
	Dir string
	// URL is the database connection URL
	URL string
	// DevURL is the dev database URL for schema calculations
	DevURL string
	// Env is the Atlas environment to use (local, staging, production)
	Env string
	// ConfigFile is the path to atlas.hcl configuration file
	ConfigFile string
	// DryRun enables dry-run mode for migrations
	DryRun bool
	// Verbose enables verbose output
	Verbose bool
}

// DefaultAtlasConfig returns a default Atlas configuration
func DefaultAtlasConfig() *AtlasConfig {
	return &AtlasConfig{
		Dir:        "migrations",
		Env:        "local",
		ConfigFile: "atlas.hcl",
	}
}

// Atlas provides methods for running Atlas migrations
type Atlas struct {
	config *AtlasConfig
}

// NewAtlas creates a new Atlas instance with the given configuration
func NewAtlas(config *AtlasConfig) *Atlas {
	if config == nil {
		config = DefaultAtlasConfig()
	}
	return &Atlas{config: config}
}

// MigrationStatus represents the status of a single migration
type MigrationStatus struct {
	Version     string `json:"Version"`
	Description string `json:"Description"`
	Applied     bool   `json:"Applied"`
	AppliedAt   string `json:"AppliedAt,omitempty"`
}

// AtlasStatus represents the overall migration status
type AtlasStatus struct {
	Current    string            `json:"Current"`
	Next       string            `json:"Next,omitempty"`
	Pending    []MigrationStatus `json:"Pending,omitempty"`
	Applied    []MigrationStatus `json:"Applied,omitempty"`
	TotalCount int               `json:"TotalCount"`
}

// Status returns the current migration status
func (a *Atlas) Status(ctx context.Context) (*AtlasStatus, error) {
	args := a.buildBaseArgs("migrate", "status")
	args = append(args, "--format", "{{ json . }}")

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	var status AtlasStatus
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		// If JSON parsing fails, return raw output as error
		return nil, fmt.Errorf("failed to parse status output: %s", output)
	}

	return &status, nil
}

// Apply runs pending migrations
func (a *Atlas) Apply(ctx context.Context) error {
	args := a.buildBaseArgs("migrate", "apply")

	if a.config.DryRun {
		args = append(args, "--dry-run")
	}

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w\n%s", err, output)
	}

	if a.config.Verbose {
		fmt.Println(output)
	}

	return nil
}

// ApplyN applies N migrations
func (a *Atlas) ApplyN(ctx context.Context, n int) error {
	args := a.buildBaseArgs("migrate", "apply")
	args = append(args, fmt.Sprintf("%d", n))

	if a.config.DryRun {
		args = append(args, "--dry-run")
	}

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w\n%s", err, output)
	}

	if a.config.Verbose {
		fmt.Println(output)
	}

	return nil
}

// Down reverts the last migration
func (a *Atlas) Down(ctx context.Context) error {
	return a.DownN(ctx, 1)
}

// DownN reverts N migrations
func (a *Atlas) DownN(ctx context.Context, n int) error {
	args := a.buildBaseArgs("migrate", "down")
	args = append(args, fmt.Sprintf("%d", n))

	if a.config.DryRun {
		args = append(args, "--dry-run")
	}

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to revert migrations: %w\n%s", err, output)
	}

	if a.config.Verbose {
		fmt.Println(output)
	}

	return nil
}

// Diff generates a new migration by comparing the schema to the database
func (a *Atlas) Diff(ctx context.Context, name string) (string, error) {
	args := a.buildBaseArgs("migrate", "diff", name)

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate migration diff: %w\n%s", err, output)
	}

	return output, nil
}

// Hash updates the atlas.sum file
func (a *Atlas) Hash(ctx context.Context) error {
	args := a.buildBaseArgs("migrate", "hash")

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to update hash: %w\n%s", err, output)
	}

	return nil
}

// Lint checks migrations for issues
func (a *Atlas) Lint(ctx context.Context) error {
	args := a.buildBaseArgs("migrate", "lint")
	args = append(args, "--latest", "1")

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("migration lint failed: %w\n%s", err, output)
	}

	if a.config.Verbose {
		fmt.Println(output)
	}

	return nil
}

// Validate validates the migrations directory
func (a *Atlas) Validate(ctx context.Context) error {
	args := a.buildBaseArgs("migrate", "validate")

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("migration validation failed: %w\n%s", err, output)
	}

	return nil
}

// New creates a new empty migration file
func (a *Atlas) New(ctx context.Context, name string) (string, error) {
	args := a.buildBaseArgs("migrate", "new", name)

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to create migration: %w\n%s", err, output)
	}

	// Parse output to find the created file path
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, ".sql") {
			return strings.TrimSpace(line), nil
		}
	}

	return output, nil
}

// SchemaInspect inspects the current database schema
func (a *Atlas) SchemaInspect(ctx context.Context) (string, error) {
	args := []string{"schema", "inspect"}

	if a.config.URL != "" {
		args = append(args, "--url", a.config.URL)
	} else if a.config.Env != "" {
		args = append(args, "--env", a.config.Env)
	}

	if a.config.ConfigFile != "" {
		if _, err := os.Stat(a.config.ConfigFile); err == nil {
			args = append(args, "-c", fmt.Sprintf("file://%s", a.config.ConfigFile))
		}
	}

	args = append(args, "--format", "{{ sql . }}")

	output, err := a.runCommand(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to inspect schema: %w\n%s", err, output)
	}

	return output, nil
}

// buildBaseArgs constructs the base arguments for Atlas commands
func (a *Atlas) buildBaseArgs(command ...string) []string {
	args := command

	// Add migration directory
	if a.config.Dir != "" {
		args = append(args, "--dir", fmt.Sprintf("file://%s", a.config.Dir))
	}

	// Add database URL
	if a.config.URL != "" {
		args = append(args, "--url", a.config.URL)
	}

	// Add dev URL
	if a.config.DevURL != "" {
		args = append(args, "--dev-url", a.config.DevURL)
	}

	// Add environment
	if a.config.Env != "" {
		args = append(args, "--env", a.config.Env)
	}

	// Add config file
	if a.config.ConfigFile != "" {
		if _, err := os.Stat(a.config.ConfigFile); err == nil {
			args = append(args, "-c", fmt.Sprintf("file://%s", a.config.ConfigFile))
		}
	}

	return args
}

// runCommand executes an Atlas CLI command
func (a *Atlas) runCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "atlas", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if a.config.Verbose {
		fmt.Printf("Running: atlas %s\n", strings.Join(args, " "))
	}

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return stderr.String(), err
		}
		return stdout.String(), err
	}

	return stdout.String(), nil
}

// IsInstalled checks if Atlas CLI is installed
func IsAtlasInstalled() bool {
	_, err := exec.LookPath("atlas")
	return err == nil
}

// InstallAtlas provides instructions for installing Atlas
func InstallAtlas() string {
	return `Atlas CLI is not installed. Install it using one of the following methods:

macOS:
  brew install ariga/tap/atlas

Linux:
  curl -sSf https://atlasgo.sh | sh

Docker:
  docker pull arigaio/atlas

For more information, visit: https://atlasgo.io/getting-started/
`
}

// EnsureAtlasInstalled checks if Atlas is installed and returns an error if not
func EnsureAtlasInstalled() error {
	if !IsAtlasInstalled() {
		return fmt.Errorf("atlas CLI is not installed.\n\n%s", InstallAtlas())
	}
	return nil
}

// InitMigrationDir initializes a new migration directory structure
func InitMigrationDir(dir string) error {
	// Create migrations directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create initial atlas.sum file
	sumFile := filepath.Join(dir, "atlas.sum")
	if _, err := os.Stat(sumFile); os.IsNotExist(err) {
		if err := os.WriteFile(sumFile, []byte("h1:placeholder\n"), 0644); err != nil {
			return fmt.Errorf("failed to create atlas.sum file: %w", err)
		}
	}

	return nil
}
