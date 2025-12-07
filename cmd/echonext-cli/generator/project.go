package generator

// ProjectData holds data for project template generation
type ProjectData struct {
	Name         string // Project name (e.g., "my-api")
	Module       string // Go module name (e.g., "github.com/user/my-api")
	Template     string // Template type (standard, minimal, etc.)
	EchoNextPath string // Path to echonext for local development
}

// ProjectGenerator uses the template engine to generate project files
type ProjectGenerator struct {
	engine *Engine
	data   *ProjectData
}

// NewProjectGenerator creates a new project generator with template engine
func NewProjectGenerator(data *ProjectData) (*ProjectGenerator, error) {
	engine, err := NewEngine()
	if err != nil {
		return nil, err
	}

	return &ProjectGenerator{
		engine: engine,
		data:   data,
	}, nil
}

// GenerateAPIMain generates the API main.go file content
func (g *ProjectGenerator) GenerateAPIMain() (string, error) {
	return g.engine.Execute("project/api_main.go.tmpl", g.data)
}

// GenerateUserModel generates the user model file content
func (g *ProjectGenerator) GenerateUserModel() (string, error) {
	return g.engine.Execute("domain/model.go.tmpl", g.data)
}

// GenerateDevelopmentConfig generates the development.yaml config file content
func (g *ProjectGenerator) GenerateDevelopmentConfig() (string, error) {
	return g.engine.Execute("config/development.yaml.tmpl", g.data)
}

// GenerateWorkerMain generates the worker main.go file content
func (g *ProjectGenerator) GenerateWorkerMain() (string, error) {
	return g.engine.Execute("project/worker_main.go.tmpl", g.data)
}

// GenerateCLIMain generates the CLI main.go file content
func (g *ProjectGenerator) GenerateCLIMain() (string, error) {
	return g.engine.Execute("project/cli_main.go.tmpl", g.data)
}

// GenerateMigrationMain generates the migration main.go file content
func (g *ProjectGenerator) GenerateMigrationMain() (string, error) {
	return g.engine.Execute("project/migration_main.go.tmpl", g.data)
}

// GenerateGoMod generates the go.mod file content
func (g *ProjectGenerator) GenerateGoMod() (string, error) {
	return g.engine.Execute("project/go.mod.tmpl", g.data)
}

// GenerateReadme generates the README.md file content
func (g *ProjectGenerator) GenerateReadme() (string, error) {
	return g.engine.Execute("project/README.md.tmpl", g.data)
}

// GenerateMakefile generates the Makefile content
func (g *ProjectGenerator) GenerateMakefile() (string, error) {
	return g.engine.Execute("project/Makefile.tmpl", g.data)
}

// GenerateProductionConfig generates the production.yaml config file content
func (g *ProjectGenerator) GenerateProductionConfig() (string, error) {
	return g.engine.Execute("config/production.yaml.tmpl", g.data)
}

// GenerateTestConfig generates the test.yaml config file content
func (g *ProjectGenerator) GenerateTestConfig() (string, error) {
	return g.engine.Execute("config/test.yaml.tmpl", g.data)
}

// GenerateConfig generates the internal/config/config.go file content
func (g *ProjectGenerator) GenerateConfig() (string, error) {
	return g.engine.Execute("internal/config.go.tmpl", g.data)
}

// GenerateDatabase generates the internal/database/database.go file content
func (g *ProjectGenerator) GenerateDatabase() (string, error) {
	return g.engine.Execute("internal/database.go.tmpl", g.data)
}

// GenerateServer generates the internal/server/server.go file content
func (g *ProjectGenerator) GenerateServer() (string, error) {
	return g.engine.Execute("internal/server.go.tmpl", g.data)
}

// GenerateOTEL generates the internal/otel/otel.go file content
func (g *ProjectGenerator) GenerateOTEL() (string, error) {
	return g.engine.Execute("internal/otel.go.tmpl", g.data)
}
