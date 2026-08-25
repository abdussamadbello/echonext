package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/abdussamadbello/echonext/cmd/echonext-cli/generator"
	"github.com/spf13/cobra"
)

// newInitCmd returns the init command
func newInitCmd() *cobra.Command {
	var template string
	var module string
	var withSkills bool

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new EchoNext project",
		Long: `Initialize a new EchoNext project with the specified template.

Available templates:
  standard    - Full-featured API with multiple domains (default)
  minimal     - Basic API with single domain  
  microservice- Production microservice template
  monolith    - Enterprise application template

Examples:
  echonext init myapi
  echonext init blog-api --template=standard
  echonext init payment-service --template=microservice
  echonext init myapi --with-skills

With --with-skills, the EchoNext agent skills are installed into the new project
so AI coding assistants know the framework's conventions. This requires npx and
writes skills-lock.json, which should be committed.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			
			// Validate project name
			if !isValidProjectName(projectName) {
				return fmt.Errorf("invalid project name: %s. Use lowercase letters, numbers, and hyphens only", projectName)
			}

			// Check if directory exists
			if _, err := os.Stat(projectName); !os.IsNotExist(err) {
				return fmt.Errorf("directory %s already exists", projectName)
			}

			// Set default module name if not provided
			if module == "" {
				module = projectName
			}

			// Create project
			generator := &ProjectGenerator{
				Name:         projectName,
				Module:       module,
				Template:     template,
				EchoNextPath: detectEchoNextPath(),
				WithSkills:   withSkills,
			}

			return generator.Generate()
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "standard", "Project template (standard, minimal, microservice, monolith)")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Go module name (defaults to project name)")
	cmd.Flags().BoolVar(&withSkills, "with-skills", false, "Install the EchoNext agent skills into the project (requires npx)")

	return cmd
}

// detectEchoNextPath finds the local echonext path for development
func detectEchoNextPath() string {
	// For now, use a simple heuristic - check if we're in echonext directory
	if _, err := os.Stat("echonext.go"); err == nil {
		return "." // We're in the echonext root directory
	}
	
	// Check if echonext.go exists in parent directories (up to 3 levels)
	for _, path := range []string{"../", "../../", "../../../"} {
		if _, err := os.Stat(path + "echonext.go"); err == nil {
			return path
		}
	}
	
	// Default to empty string - will use normal module resolution
	return ""
}

// isValidProjectName validates the project name
func isValidProjectName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	
	return true
}

// ProjectGenerator handles project generation
type ProjectGenerator struct {
	Name             string
	Module           string
	Template         string
	EchoNextPath     string // Path to echonext for local development
	WithSkills       bool   // Install agent skills into the generated project
	templateGenerator *generator.ProjectGenerator
}

// initTemplateGenerator initializes the template generator
func (g *ProjectGenerator) initTemplateGenerator() error {
	var err error
	g.templateGenerator, err = generator.NewProjectGenerator(&generator.ProjectData{
		Name:         g.Name,
		Module:       g.Module,
		Template:     g.Template,
		EchoNextPath: g.EchoNextPath,
	})
	return err
}

// Generate creates the project structure
func (g *ProjectGenerator) Generate() error {
	fmt.Printf("🚀 Creating EchoNext project '%s' with template '%s'...\n", g.Name, g.Template)

	// Initialize template generator
	if err := g.initTemplateGenerator(); err != nil {
		return fmt.Errorf("failed to initialize template generator: %w", err)
	}

	// Create project directory
	if err := os.MkdirAll(g.Name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate based on template
	switch g.Template {
	case "minimal":
		return g.generateMinimalProject()
	case "microservice":
		return g.generateMicroserviceProject()
	case "monolith":
		return g.generateMonolithProject()
	default: // "standard"
		return g.generateStandardProject()
	}
}

// generateStandardProject creates a standard project with multiple executables
func (g *ProjectGenerator) generateStandardProject() error {
	// Create directory structure
	dirs := []string{
		"cmd/api",
		"cmd/worker",
		"cmd/cli",
		"cmd/migration",
		"domain/user",
		"internal/config",
		"internal/database",
		"internal/server",
		"tests/integration",
		"tests/fixtures",
		"tests/helpers",
		"configs",
		"migrations",
		"scripts",
		"infrastructure",
	}

	for _, dir := range dirs {
		path := filepath.Join(g.Name, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	// Generate files
	files := map[string]string{
		"go.mod":                             g.generateGoMod(),
		"README.md":                          g.generateReadme(),
		"Makefile":                           g.generateMakefile(),
		"atlas.hcl":                          g.generateAtlasConfig(),
		"schema.hcl":                         g.generateSchemaHCL(),
		"infrastructure/docker-compose.yml": g.generateDockerCompose(),
		"infrastructure/Dockerfile.api":     g.generateDockerfileAPI(),
		"infrastructure/Dockerfile.worker":  g.generateDockerfileWorker(),
		".gitignore":                         g.generateGitignore(),
		".env.example":                       g.generateEnvExample(),
		"cmd/api/main.go":                    g.generateAPIMain(),
		"cmd/worker/main.go":                 g.generateWorkerMain(),
		"cmd/cli/main.go":                    g.generateCLIMain(),
		"cmd/migration/main.go":              g.generateMigrationMain(),
		"internal/config/config.go":          g.generateConfig(),
		"internal/database/database.go":      g.generateDatabase(),
		"internal/server/server.go":          g.generateServer(),
		"configs/development.yaml":           g.generateConfigYAML("development"),
		"configs/production.yaml":            g.generateConfigYAML("production"),
		"configs/test.yaml":                  g.generateConfigYAML("test"),
		"domain/user/model.go":               g.generateUserModel(),
		"domain/user/service.go":             g.generateUserService(),
		"domain/user/handler.go":             g.generateUserHandler(),
		"domain/user/dto.go":                 g.generateUserDTO(),
		"domain/user/service_test.go":        g.generateUserTest(),
		"migrations/00001_initial.sql":       g.generateInitialMigration(),
		"migrations/atlas.sum":               g.generateAtlasSum(),
	}

	for filePath, content := range files {
		fullPath := filepath.Join(g.Name, filePath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
	}

	// Run go mod tidy to resolve dependencies
	fmt.Println("Running go mod tidy...")
	modTidyCmd := exec.Command("go", "mod", "tidy")
	modTidyCmd.Dir = g.Name
	modTidyCmd.Stdout = os.Stdout
	modTidyCmd.Stderr = os.Stderr
	if err := modTidyCmd.Run(); err != nil {
		fmt.Printf("Warning: go mod tidy failed: %v\n", err)
		fmt.Println("You may need to run 'go mod tidy' manually.")
	}

	// Install agent skills so AI assistants know the framework's conventions
	if g.WithSkills {
		g.installSkills()
	}

	fmt.Printf("\n✅ Project '%s' created successfully!\n\n", g.Name)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", g.Name)
	fmt.Printf("  export DATABASE_URL='postgres://user:pass@localhost:5432/%s_dev?sslmode=disable'\n", g.Name)
	fmt.Printf("  make dev-db      # Start PostgreSQL with Docker\n")
	fmt.Printf("  make db-migrate  # Run database migrations\n")
	fmt.Printf("  make run-api     # Start the API server\n\n")
	fmt.Printf("Or use the dev server with hot-reload:\n")
	fmt.Printf("  echonext dev     # Start with file watching\n\n")
	fmt.Printf("Documentation will be available at http://localhost:8080/api/docs\n")

	return nil
}

// installSkills installs the EchoNext agent skills into the generated project.
// Skills are optional tooling, so every failure here is a warning: the project
// itself is already complete by the time this runs.
func (g *ProjectGenerator) installSkills() {
	const manual = "npx skills add abdussamadbello/echonext"

	if _, err := exec.LookPath("npx"); err != nil {
		fmt.Println("\n⚠️  Skipping agent skills: npx not found on PATH.")
		fmt.Printf("   Install them later with: %s\n", manual)
		return
	}

	fmt.Println("\nInstalling EchoNext agent skills...")
	skillsCmd := exec.Command("npx", "-y", "skills", "add", "abdussamadbello/echonext", "-s", "*", "-y")
	skillsCmd.Dir = g.Name
	skillsCmd.Stdout = os.Stdout
	skillsCmd.Stderr = os.Stderr
	if err := skillsCmd.Run(); err != nil {
		fmt.Printf("⚠️  Agent skill install failed: %v\n", err)
		fmt.Printf("   Install them later with: %s\n", manual)
		return
	}

	fmt.Println("✅ Agent skills installed. Commit skills-lock.json so collaborators")
	fmt.Println("   can restore them with: npx skills experimental_install")
}

// Helper methods for other template types
func (g *ProjectGenerator) generateMinimalProject() error {
	// TODO: Implement minimal template
	fmt.Println("⚠️  Minimal template not yet implemented, using standard template")
	return g.generateStandardProject()
}

func (g *ProjectGenerator) generateMicroserviceProject() error {
	// TODO: Implement microservice template
	fmt.Println("⚠️  Microservice template not yet implemented, using standard template") 
	return g.generateStandardProject()
}

func (g *ProjectGenerator) generateMonolithProject() error {
	// TODO: Implement monolith template
	fmt.Println("⚠️  Monolith template not yet implemented, using standard template")
	return g.generateStandardProject()
}

// Template generation methods
func (g *ProjectGenerator) generateGoMod() string {
	content, err := g.templateGenerator.GenerateGoMod()
	if err != nil {
		log.Fatalf("Failed to generate go.mod: %v", err)
	}
	return content
}

func (g *ProjectGenerator) generateReadme() string {
	content, err := g.templateGenerator.GenerateReadme()
	if err != nil {
		log.Fatalf("Failed to generate README: %v", err)
	}
	return content
}

func (g *ProjectGenerator) generateMakefile() string {
	content, err := g.templateGenerator.GenerateMakefile()
	if err != nil {
		log.Fatalf("Failed to generate Makefile: %v", err)
	}
	return content
}

// Additional methods would continue here...
// This is getting quite long, so I'll create separate files for the templates