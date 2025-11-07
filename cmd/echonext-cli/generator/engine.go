package generator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

//go:embed templates/**/*.tmpl
var embeddedTemplates embed.FS

// Engine handles template parsing and execution with Sprig functions
type Engine struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewEngine creates a new template engine with Sprig functions
func NewEngine() (*Engine, error) {
	engine := &Engine{
		templates: make(map[string]*template.Template),
		funcMap:   createFuncMap(),
	}

	// Load embedded templates
	if err := engine.loadEmbeddedTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load embedded templates: %w", err)
	}

	return engine, nil
}

// createFuncMap creates a template function map with Sprig + custom functions
func createFuncMap() template.FuncMap {
	// Start with all Sprig functions
	funcMap := sprig.TxtFuncMap()

	// Add custom functions beyond Sprig
	funcMap["pluralize"] = pluralize
	funcMap["singularize"] = singularize
	funcMap["pascalCase"] = pascalCase
	funcMap["lowerCamelCase"] = lowerCamelCase

	return funcMap
}

// loadEmbeddedTemplates loads all templates from embedded filesystem
func (e *Engine) loadEmbeddedTemplates() error {
	// Walk through embedded templates
	return walkEmbeddedDir("templates", func(path string, isDir bool) error {
		if isDir {
			return nil
		}

		// Only process .tmpl files
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		// Read template content
		content, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}

		// Template name is the path without "templates/" prefix
		name := strings.TrimPrefix(path, "templates/")

		// Parse template with function map
		tmpl, err := template.New(name).Funcs(e.funcMap).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		e.templates[name] = tmpl
		return nil
	})
}

// walkEmbeddedDir recursively walks through embedded directory
func walkEmbeddedDir(dir string, fn func(path string, isDir bool) error) error {
	entries, err := embeddedTemplates.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := fn(path, true); err != nil {
				return err
			}
			if err := walkEmbeddedDir(path, fn); err != nil {
				return err
			}
		} else {
			if err := fn(path, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// Execute executes a template with the given data
func (e *Engine) Execute(templateName string, data interface{}) (string, error) {
	// Check for filesystem override first (for user customization)
	if customTmpl, err := e.loadCustomTemplate(templateName); err == nil {
		return e.executeTemplate(customTmpl, data)
	}

	// Fall back to embedded template
	tmpl, ok := e.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	return e.executeTemplate(tmpl, data)
}

// executeTemplate executes a template and returns the result
func (e *Engine) executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}

// loadCustomTemplate loads a custom template from filesystem
// Checks in ~/.echonext/templates/ for user overrides
func (e *Engine) loadCustomTemplate(templateName string) (*template.Template, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	customPath := filepath.Join(homeDir, ".echonext", "templates", templateName)
	content, err := os.ReadFile(customPath)
	if err != nil {
		return nil, err
	}

	return template.New(templateName).Funcs(e.funcMap).Parse(string(content))
}

// ListTemplates returns all available template names
func (e *Engine) ListTemplates() []string {
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}

// Custom template functions beyond Sprig

// pluralize converts a word to its plural form (simple English rules)
func pluralize(word string) string {
	if word == "" {
		return word
	}

	// Simple pluralization rules
	switch {
	case strings.HasSuffix(word, "y"):
		return strings.TrimSuffix(word, "y") + "ies"
	case strings.HasSuffix(word, "s"), strings.HasSuffix(word, "x"),
		strings.HasSuffix(word, "z"), strings.HasSuffix(word, "ch"),
		strings.HasSuffix(word, "sh"):
		return word + "es"
	default:
		return word + "s"
	}
}

// singularize converts a word to its singular form (simple English rules)
func singularize(word string) string {
	if word == "" {
		return word
	}

	switch {
	case strings.HasSuffix(word, "ies"):
		return strings.TrimSuffix(word, "ies") + "y"
	case strings.HasSuffix(word, "ses"), strings.HasSuffix(word, "xes"),
		strings.HasSuffix(word, "zes"), strings.HasSuffix(word, "ches"),
		strings.HasSuffix(word, "shes"):
		return strings.TrimSuffix(word, "es")
	case strings.HasSuffix(word, "s"):
		return strings.TrimSuffix(word, "s")
	default:
		return word
	}
}

// pascalCase converts a string to PascalCase (e.g., "user_account" -> "UserAccount")
func pascalCase(s string) string {
	if s == "" {
		return s
	}

	// Split by common separators
	words := splitWords(s)

	// Capitalize first letter of each word
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, "")
}

// lowerCamelCase converts a string to lowerCamelCase (e.g., "user_account" -> "userAccount")
func lowerCamelCase(s string) string {
	if s == "" {
		return s
	}

	pascal := pascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}

	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// splitWords splits a string by common separators (_, -, space)
func splitWords(s string) []string {
	// Replace separators with space
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Split by space and filter empty strings
	words := strings.Fields(s)
	return words
}
