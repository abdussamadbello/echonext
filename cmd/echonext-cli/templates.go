package main

import (
	"log"
)

// generateAPIMain creates the main.go file for the API server
func (g *ProjectGenerator) generateAPIMain() string {
	content, err := g.templateGenerator.GenerateAPIMain()
	if err != nil {
		log.Fatalf("Failed to generate API main: %v", err)
	}
	return content
}

// generateWorkerMain creates the main.go file for the background worker
func (g *ProjectGenerator) generateWorkerMain() string {
	content, err := g.templateGenerator.GenerateWorkerMain()
	if err != nil {
		log.Fatalf("Failed to generate worker main: %v", err)
	}
	return content
}

// generateCLIMain creates the main.go file for the CLI tool
func (g *ProjectGenerator) generateCLIMain() string {
	content, err := g.templateGenerator.GenerateCLIMain()
	if err != nil {
		log.Fatalf("Failed to generate CLI main: %v", err)
	}
	return content
}

// generateMigrationMain creates the main.go file for database migrations
func (g *ProjectGenerator) generateMigrationMain() string {
	content, err := g.templateGenerator.GenerateMigrationMain()
	if err != nil {
		log.Fatalf("Failed to generate migration main: %v", err)
	}
	return content
}

// generateConfig creates the configuration package
func (g *ProjectGenerator) generateConfig() string {
	content, err := g.templateGenerator.GenerateConfig()
	if err != nil {
		log.Fatalf("Failed to generate config: %v", err)
	}
	return content
}

// generateDatabase creates the database package
func (g *ProjectGenerator) generateDatabase() string {
	content, err := g.templateGenerator.GenerateDatabase()
	if err != nil {
		log.Fatalf("Failed to generate database: %v", err)
	}
	return content
}

// generateServer creates the server package
func (g *ProjectGenerator) generateServer() string {
	content, err := g.templateGenerator.GenerateServer()
	if err != nil {
		log.Fatalf("Failed to generate server: %v", err)
	}
	return content
}

// generateOTEL creates the OpenTelemetry setup package
func (g *ProjectGenerator) generateOTEL() string {
	content, err := g.templateGenerator.GenerateOTEL()
	if err != nil {
		log.Fatalf("Failed to generate OTEL: %v", err)
	}
	return content
}

// generateAtlasConfig creates the atlas.hcl configuration file
func (g *ProjectGenerator) generateAtlasConfig() string {
	content, err := g.templateGenerator.GenerateAtlasConfig()
	if err != nil {
		log.Fatalf("Failed to generate atlas.hcl: %v", err)
	}
	return content
}

// generateSchemaHCL creates the schema.hcl file for Atlas declarative schema
func (g *ProjectGenerator) generateSchemaHCL() string {
	content, err := g.templateGenerator.GenerateSchemaHCL()
	if err != nil {
		log.Fatalf("Failed to generate schema.hcl: %v", err)
	}
	return content
}

// generateInitialMigration creates the initial SQL migration file
func (g *ProjectGenerator) generateInitialMigration() string {
	content, err := g.templateGenerator.GenerateInitialMigration()
	if err != nil {
		log.Fatalf("Failed to generate initial migration: %v", err)
	}
	return content
}

// generateAtlasSum creates the atlas.sum file for migration checksums
func (g *ProjectGenerator) generateAtlasSum() string {
	content, err := g.templateGenerator.GenerateAtlasSum()
	if err != nil {
		log.Fatalf("Failed to generate atlas.sum: %v", err)
	}
	return content
}

// generateEnvExample creates the .env.example file
func (g *ProjectGenerator) generateEnvExample() string {
	content, err := g.templateGenerator.GenerateEnvExample()
	if err != nil {
		log.Fatalf("Failed to generate .env.example: %v", err)
	}
	return content
}