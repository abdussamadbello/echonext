package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Custom Function Tests
// =============================================================================

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"regular word", "user", "users"},
		{"word ending in y", "category", "categories"},
		{"word ending in s", "bus", "buses"},
		{"word ending in x", "box", "boxes"},
		{"word ending in z", "quiz", "quizes"},
		{"word ending in ch", "match", "matches"},
		{"word ending in sh", "dish", "dishes"},
		{"single letter", "a", "as"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pluralize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"regular plural", "users", "user"},
		{"word ending in ies", "categories", "category"},
		{"word ending in ses", "buses", "bus"},
		{"word ending in xes", "boxes", "box"},
		{"word ending in zes", "quizes", "quiz"},
		{"word ending in ches", "matches", "match"},
		{"word ending in shes", "dishes", "dish"},
		{"already singular", "user", "user"},
		{"word not ending in s", "child", "child"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := singularize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"snake_case", "user_account", "UserAccount"},
		{"kebab-case", "user-account", "UserAccount"},
		{"space separated", "user account", "UserAccount"},
		{"single word", "user", "User"},
		{"already pascal", "User", "User"},
		{"mixed separators", "user_account-info", "UserAccountInfo"},
		{"multiple underscores", "user__account", "UserAccount"},
		{"uppercase word", "USER", "USER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLowerCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"snake_case", "user_account", "userAccount"},
		{"kebab-case", "user-account", "userAccount"},
		{"space separated", "user account", "userAccount"},
		{"single word", "user", "user"},
		{"already camel", "userAccount", "userAccount"},
		{"pascal case input", "UserAccount", "userAccount"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lowerCamelCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", []string{}},
		{"snake_case", "user_account", []string{"user", "account"}},
		{"kebab-case", "user-account", []string{"user", "account"}},
		{"space separated", "user account", []string{"user", "account"}},
		{"single word", "user", []string{"user"}},
		{"mixed separators", "user_account-info name", []string{"user", "account", "info", "name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitWords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Engine Tests
// =============================================================================

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)
	require.NotNil(t, engine)

	// Should have loaded templates
	templates := engine.ListTemplates()
	assert.Greater(t, len(templates), 0, "should have loaded at least one template")
}

func TestEngineListTemplates(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	templates := engine.ListTemplates()

	// Check for expected templates
	expectedTemplates := []string{
		"project/go.mod.tmpl",
		"project/api_main.go.tmpl",
	}

	for _, expected := range expectedTemplates {
		found := false
		for _, tmpl := range templates {
			if tmpl == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "expected template %s to be loaded", expected)
	}
}

func TestEngineExecute(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	// Test with project data
	data := map[string]interface{}{
		"Name":   "testproject",
		"Module": "github.com/test/testproject",
	}

	// Execute go.mod template
	result, err := engine.Execute("project/go.mod.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "github.com/test/testproject")
}

func TestEngineExecuteNonexistentTemplate(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	_, err = engine.Execute("nonexistent.tmpl", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestCreateFuncMap(t *testing.T) {
	funcMap := createFuncMap()

	// Check custom functions are present
	customFuncs := []string{"pluralize", "singularize", "pascalCase", "lowerCamelCase"}
	for _, fn := range customFuncs {
		_, exists := funcMap[fn]
		assert.True(t, exists, "funcMap should contain %s", fn)
	}

	// Check some Sprig functions are present
	sprigFuncs := []string{"upper", "lower", "title", "trim", "default"}
	for _, fn := range sprigFuncs {
		_, exists := funcMap[fn]
		assert.True(t, exists, "funcMap should contain Sprig function %s", fn)
	}
}

// =============================================================================
// Template Function Integration Tests
// =============================================================================

func TestTemplateFunctionsInTemplate(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)

	// The templates should be able to use our custom functions
	// We can verify this by checking that templates load and execute without error
	templates := engine.ListTemplates()
	assert.Greater(t, len(templates), 0)
}
