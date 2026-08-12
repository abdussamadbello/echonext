package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/abdussamadbello/echonext/cmd/echonext-cli/generator"
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPIGenerator generates Go code from OpenAPI specifications
type OpenAPIGenerator struct {
	Spec            *openapi3.T
	OutputDir       string
	PackageName     string
	Engine          *generator.Engine
	ShouldGenTests  bool
	SpecPath        string
}

// GeneratedFiles tracks all generated file paths
type GeneratedFiles struct {
	Models   []string
	Handlers []string
	DTOs     []string
	Routes   string
	Tests    []string
}

// SchemaInfo contains information about a Go type to generate
type SchemaInfo struct {
	Name        string
	Description string
	GoType      string
	Properties  []PropertyInfo
	Imports     []string
	IsEnum      bool
	EnumValues  []string
}

// PropertyInfo describes a struct field
type PropertyInfo struct {
	Name        string
	GoName      string
	Type        string
	JSONTag     string
	ValidateTag string
	Required    bool
	Description string
	Example     string
}

// EndpointInfo describes an API endpoint
type EndpointInfo struct {
	Method       string
	Path         string
	OperationID  string
	FuncName     string
	Summary      string
	Description  string
	Tags         []string
	RequestType  string
	ResponseType string
	PathParams   []ParamInfo
	QueryParams  []ParamInfo
	HeaderParams []ParamInfo
	Security     []string
	HasRequest   bool
	HasResponse  bool
}

// ParamInfo describes a parameter
type ParamInfo struct {
	Name        string
	GoName      string
	Type        string
	Required    bool
	Description string
	In          string // path, query, header
}

// GeneratorOptions for configuring the generator
type GeneratorOptions struct {
	OutputDir     string
	PackageName   string
	GenerateTests bool
	TemplateDir   string
}

// NewOpenAPIGenerator creates a new generator from an OpenAPI spec file
func NewOpenAPIGenerator(specPath string, opts GeneratorOptions) (*OpenAPIGenerator, error) {
	// Load and parse OpenAPI spec
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate spec
	if err := spec.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	// Create template engine
	engine, err := generator.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create template engine: %w", err)
	}

	return &OpenAPIGenerator{
		Spec:           spec,
		OutputDir:      opts.OutputDir,
		PackageName:    opts.PackageName,
		Engine:         engine,
		ShouldGenTests: opts.GenerateTests,
		SpecPath:       specPath,
	}, nil
}

// Generate runs the complete code generation process
func (g *OpenAPIGenerator) Generate() (*GeneratedFiles, error) {
	files := &GeneratedFiles{}

	// Create output directory
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate models from schemas
	models, err := g.GenerateModels()
	if err != nil {
		return nil, fmt.Errorf("failed to generate models: %w", err)
	}
	files.Models = models

	// Generate DTOs (request/response types)
	dtos, err := g.GenerateDTOs()
	if err != nil {
		return nil, fmt.Errorf("failed to generate DTOs: %w", err)
	}
	files.DTOs = dtos

	// Generate handlers
	handlers, err := g.GenerateHandlers()
	if err != nil {
		return nil, fmt.Errorf("failed to generate handlers: %w", err)
	}
	files.Handlers = handlers

	// Generate routes
	routes, err := g.GenerateRoutes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate routes: %w", err)
	}
	files.Routes = routes

	// Generate tests if enabled
	if g.ShouldGenTests {
		tests, err := g.GenerateTests()
		if err != nil {
			return nil, fmt.Errorf("failed to generate tests: %w", err)
		}
		files.Tests = tests
	}

	return files, nil
}

// GenerateModels generates Go structs from OpenAPI schemas
func (g *OpenAPIGenerator) GenerateModels() ([]string, error) {
	if g.Spec.Components == nil || g.Spec.Components.Schemas == nil {
		return nil, nil
	}

	var generatedFiles []string
	modelsDir := filepath.Join(g.OutputDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return nil, err
	}

	// Collect all schema names and sort for consistent output
	var schemaNames []string
	for name := range g.Spec.Components.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	// Generate a single models.go file with all models
	var schemas []SchemaInfo
	imports := make(map[string]bool)

	for _, name := range schemaNames {
		schemaRef := g.Spec.Components.Schemas[name]
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}

		schemaInfo := g.schemaToType(name, schemaRef.Value)
		schemas = append(schemas, schemaInfo)

		for _, imp := range schemaInfo.Imports {
			imports[imp] = true
		}
	}

	// Convert imports map to slice
	var importList []string
	for imp := range imports {
		importList = append(importList, imp)
	}
	sort.Strings(importList)

	// Generate models file
	data := map[string]interface{}{
		"PackageName": g.PackageName,
		"Schemas":     schemas,
		"Imports":     importList,
	}

	content, err := g.executeTemplate("openapi/models.go.tmpl", data)
	if err != nil {
		// Fallback to inline generation if template doesn't exist
		content = g.generateModelsInline(schemas, importList)
	}

	modelsFile := filepath.Join(modelsDir, "models.go")
	if err := os.WriteFile(modelsFile, []byte(content), 0644); err != nil {
		return nil, err
	}

	generatedFiles = append(generatedFiles, modelsFile)
	return generatedFiles, nil
}

// GenerateDTOs generates request/response DTOs from operations
func (g *OpenAPIGenerator) GenerateDTOs() ([]string, error) {
	var generatedFiles []string
	dtosDir := filepath.Join(g.OutputDir, "dto")
	if err := os.MkdirAll(dtosDir, 0755); err != nil {
		return nil, err
	}

	// Collect request/response types from all operations
	var dtos []SchemaInfo
	imports := make(map[string]bool)

	for path, pathItem := range g.Spec.Paths.Map() {
		operations := map[string]*openapi3.Operation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"PATCH":  pathItem.Patch,
			"DELETE": pathItem.Delete,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}

			operationID := op.OperationID
			if operationID == "" {
				operationID = generateOperationID(method, path)
			}

			// Generate request DTO
			if op.RequestBody != nil && op.RequestBody.Value != nil {
				for _, mediaType := range op.RequestBody.Value.Content {
					if mediaType.Schema != nil && mediaType.Schema.Value != nil {
						reqName := toPascalCase(operationID) + "Request"
						reqSchema := g.schemaToType(reqName, mediaType.Schema.Value)
						dtos = append(dtos, reqSchema)
						for _, imp := range reqSchema.Imports {
							imports[imp] = true
						}
					}
					break // Only process first content type
				}
			}

			// Generate response DTO
			for statusCode, responseRef := range op.Responses.Map() {
				if !strings.HasPrefix(statusCode, "2") {
					continue // Only success responses
				}
				if responseRef.Value != nil && responseRef.Value.Content != nil {
					for _, mediaType := range responseRef.Value.Content {
						if mediaType.Schema != nil && mediaType.Schema.Value != nil {
							respName := toPascalCase(operationID) + "Response"
							respSchema := g.schemaToType(respName, mediaType.Schema.Value)
							dtos = append(dtos, respSchema)
							for _, imp := range respSchema.Imports {
								imports[imp] = true
							}
						}
						break // Only process first content type
					}
				}
				break // Only process first success response
			}
		}
	}

	if len(dtos) == 0 {
		return nil, nil
	}

	// Convert imports map to slice
	var importList []string
	for imp := range imports {
		importList = append(importList, imp)
	}
	sort.Strings(importList)

	// Generate DTOs file
	data := map[string]interface{}{
		"PackageName": g.PackageName,
		"DTOs":        dtos,
		"Imports":     importList,
	}

	content, err := g.executeTemplate("openapi/dto.go.tmpl", data)
	if err != nil {
		content = g.generateDTOsInline(dtos, importList)
	}

	dtosFile := filepath.Join(dtosDir, "dto.go")
	if err := os.WriteFile(dtosFile, []byte(content), 0644); err != nil {
		return nil, err
	}

	generatedFiles = append(generatedFiles, dtosFile)
	return generatedFiles, nil
}

// GenerateHandlers generates handler functions
func (g *OpenAPIGenerator) GenerateHandlers() ([]string, error) {
	var generatedFiles []string
	handlersDir := filepath.Join(g.OutputDir, "handlers")
	if err := os.MkdirAll(handlersDir, 0755); err != nil {
		return nil, err
	}

	endpoints := g.collectEndpoints()

	data := map[string]interface{}{
		"PackageName": g.PackageName,
		"Endpoints":   endpoints,
	}

	content, err := g.executeTemplate("openapi/handlers.go.tmpl", data)
	if err != nil {
		content = g.generateHandlersInline(endpoints)
	}

	handlersFile := filepath.Join(handlersDir, "handlers.go")
	if err := os.WriteFile(handlersFile, []byte(content), 0644); err != nil {
		return nil, err
	}

	generatedFiles = append(generatedFiles, handlersFile)
	return generatedFiles, nil
}

// GenerateRoutes generates route registration code
func (g *OpenAPIGenerator) GenerateRoutes() (string, error) {
	endpoints := g.collectEndpoints()

	data := map[string]interface{}{
		"PackageName": g.PackageName,
		"Endpoints":   endpoints,
	}

	content, err := g.executeTemplate("openapi/routes.go.tmpl", data)
	if err != nil {
		content = g.generateRoutesInline(endpoints)
	}

	routesFile := filepath.Join(g.OutputDir, "routes.go")
	if err := os.WriteFile(routesFile, []byte(content), 0644); err != nil {
		return "", err
	}

	return routesFile, nil
}

// GenerateTests generates test files
func (g *OpenAPIGenerator) GenerateTests() ([]string, error) {
	var generatedFiles []string
	testsDir := filepath.Join(g.OutputDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return nil, err
	}

	endpoints := g.collectEndpoints()

	data := map[string]interface{}{
		"PackageName": g.PackageName,
		"Endpoints":   endpoints,
	}

	content, err := g.executeTemplate("openapi/handlers_test.go.tmpl", data)
	if err != nil {
		content = g.generateTestsInline(endpoints)
	}

	testsFile := filepath.Join(testsDir, "handlers_test.go")
	if err := os.WriteFile(testsFile, []byte(content), 0644); err != nil {
		return nil, err
	}

	generatedFiles = append(generatedFiles, testsFile)
	return generatedFiles, nil
}

// collectEndpoints extracts endpoint information from the spec
func (g *OpenAPIGenerator) collectEndpoints() []EndpointInfo {
	var endpoints []EndpointInfo

	// Sort paths for consistent output
	var paths []string
	for path := range g.Spec.Paths.Map() {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := g.Spec.Paths.Value(path)
		operations := map[string]*openapi3.Operation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"PATCH":  pathItem.Patch,
			"DELETE": pathItem.Delete,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}

			operationID := op.OperationID
			if operationID == "" {
				operationID = generateOperationID(method, path)
			}

			endpoint := EndpointInfo{
				Method:      method,
				Path:        path,
				OperationID: operationID,
				FuncName:    toPascalCase(operationID),
				Summary:     op.Summary,
				Description: op.Description,
				Tags:        op.Tags,
			}

			// Extract parameters
			for _, paramRef := range op.Parameters {
				if paramRef.Value == nil {
					continue
				}
				param := paramRef.Value
				paramInfo := ParamInfo{
					Name:        param.Name,
					GoName:      toPascalCase(param.Name),
					Type:        openAPITypeToGo(param.Schema),
					Required:    param.Required,
					Description: param.Description,
					In:          param.In,
				}

				switch param.In {
				case "path":
					endpoint.PathParams = append(endpoint.PathParams, paramInfo)
				case "query":
					endpoint.QueryParams = append(endpoint.QueryParams, paramInfo)
				case "header":
					endpoint.HeaderParams = append(endpoint.HeaderParams, paramInfo)
				}
			}

			// Check for request body
			if op.RequestBody != nil && op.RequestBody.Value != nil {
				endpoint.HasRequest = true
				endpoint.RequestType = toPascalCase(operationID) + "Request"
			}

			// Check for response
			for statusCode := range op.Responses.Map() {
				if strings.HasPrefix(statusCode, "2") {
					endpoint.HasResponse = true
					endpoint.ResponseType = toPascalCase(operationID) + "Response"
					break
				}
			}

			// Extract security
			if op.Security != nil {
				for _, secReq := range *op.Security {
					for schemeName := range secReq {
						endpoint.Security = append(endpoint.Security, schemeName)
					}
				}
			}

			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

// schemaToType converts an OpenAPI schema to a Go type
func (g *OpenAPIGenerator) schemaToType(name string, schema *openapi3.Schema) SchemaInfo {
	info := SchemaInfo{
		Name:        toPascalCase(name),
		Description: schema.Description,
	}

	// Handle enum types
	if len(schema.Enum) > 0 {
		info.IsEnum = true
		for _, v := range schema.Enum {
			if str, ok := v.(string); ok {
				info.EnumValues = append(info.EnumValues, str)
			}
		}
		info.GoType = "string"
		return info
	}

	// Handle object types
	if schema.Type.Is("object") || len(schema.Properties) > 0 {
		info.GoType = "struct"

		// Track required fields
		requiredMap := make(map[string]bool)
		for _, req := range schema.Required {
			requiredMap[req] = true
		}

		for propName, propRef := range schema.Properties {
			if propRef.Value == nil {
				continue
			}

			prop := propRef.Value
			propInfo := PropertyInfo{
				Name:        propName,
				GoName:      toPascalCase(propName),
				Type:        openAPITypeToGo(propRef),
				JSONTag:     propName,
				Required:    requiredMap[propName],
				Description: prop.Description,
			}

			// Build validation tag
			var validateParts []string
			if propInfo.Required {
				validateParts = append(validateParts, "required")
			}
			if prop.MinLength > 0 {
				validateParts = append(validateParts, fmt.Sprintf("min=%d", prop.MinLength))
			}
			if prop.MaxLength != nil && *prop.MaxLength > 0 {
				validateParts = append(validateParts, fmt.Sprintf("max=%d", *prop.MaxLength))
			}
			if prop.Format == "email" {
				validateParts = append(validateParts, "email")
			}
			if len(prop.Enum) > 0 {
				var enumStrs []string
				for _, e := range prop.Enum {
					if str, ok := e.(string); ok {
						enumStrs = append(enumStrs, str)
					}
				}
				if len(enumStrs) > 0 {
					validateParts = append(validateParts, "oneof="+strings.Join(enumStrs, " "))
				}
			}
			propInfo.ValidateTag = strings.Join(validateParts, ",")

			// Add example if present
			if prop.Example != nil {
				propInfo.Example = fmt.Sprintf("%v", prop.Example)
			}

			info.Properties = append(info.Properties, propInfo)

			// Track imports needed
			if strings.Contains(propInfo.Type, "time.Time") {
				info.Imports = append(info.Imports, "time")
			}
		}

		// Sort properties for consistent output
		sort.Slice(info.Properties, func(i, j int) bool {
			return info.Properties[i].Name < info.Properties[j].Name
		})
	} else {
		info.GoType = schemaTypeToGo(schema)
	}

	return info
}

// executeTemplate executes a template with the given data
func (g *OpenAPIGenerator) executeTemplate(templateName string, data interface{}) (string, error) {
	return g.Engine.Execute(templateName, data)
}

// Helper functions

func generateOperationID(method, path string) string {
	// Convert path to camelCase operation ID
	// /users/{id} -> getUsersById
	path = strings.ReplaceAll(path, "{", "By")
	path = strings.ReplaceAll(path, "}", "")
	path = strings.ReplaceAll(path, "/", " ")
	path = strings.TrimSpace(path)

	words := strings.Fields(path)
	result := strings.ToLower(method)
	for _, word := range words {
		result += toPascalCase(word)
	}

	return result
}

func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	// Handle special cases
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")

	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	result := strings.Join(words, "")

	// Handle common acronyms
	acronyms := []string{"ID", "URL", "HTTP", "JSON", "XML", "API", "UUID"}
	for _, acronym := range acronyms {
		lower := strings.ToLower(acronym)
		// Replace only if at word boundary
		re := regexp.MustCompile(`(?i)` + lower + `\b`)
		result = re.ReplaceAllString(result, acronym)
	}

	return result
}

func toLowerCamel(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func openAPITypeToGo(schemaRef *openapi3.SchemaRef) string {
	if schemaRef == nil || schemaRef.Value == nil {
		return "interface{}"
	}
	return schemaTypeToGo(schemaRef.Value)
}

func schemaTypeToGo(schema *openapi3.Schema) string {
	// kin-openapi models a schema's type as a list (OpenAPI 3.1 allows several),
	// so match on a single type via Is rather than comparing to a string.
	switch {
	case schema.Type.Is("string"):
		switch schema.Format {
		case "date-time":
			return "time.Time"
		case "date":
			return "string" // Could use civil.Date
		case "uuid":
			return "string" // Could use uuid.UUID
		case "binary":
			return "[]byte"
		default:
			return "string"
		}
	case schema.Type.Is("integer"):
		switch schema.Format {
		case "int32":
			return "int32"
		case "int64":
			return "int64"
		default:
			return "int"
		}
	case schema.Type.Is("number"):
		switch schema.Format {
		case "float":
			return "float32"
		default:
			return "float64"
		}
	case schema.Type.Is("boolean"):
		return "bool"
	case schema.Type.Is("array"):
		if schema.Items != nil {
			return "[]" + openAPITypeToGo(schema.Items)
		}
		return "[]interface{}"
	case schema.Type.Is("object"):
		if schema.AdditionalProperties.Schema != nil {
			return "map[string]" + openAPITypeToGo(schema.AdditionalProperties.Schema)
		}
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}

// Inline generation fallbacks (used when templates don't exist)

func (g *OpenAPIGenerator) generateModelsInline(schemas []SchemaInfo, imports []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))

	if len(imports) > 0 {
		sb.WriteString("import (\n")
		for _, imp := range imports {
			sb.WriteString(fmt.Sprintf("\t%q\n", imp))
		}
		sb.WriteString(")\n\n")
	}

	for _, schema := range schemas {
		if schema.Description != "" {
			sb.WriteString(fmt.Sprintf("// %s %s\n", schema.Name, schema.Description))
		}

		if schema.IsEnum {
			sb.WriteString(fmt.Sprintf("type %s string\n\n", schema.Name))
			sb.WriteString("const (\n")
			for _, val := range schema.EnumValues {
				constName := schema.Name + toPascalCase(val)
				sb.WriteString(fmt.Sprintf("\t%s %s = %q\n", constName, schema.Name, val))
			}
			sb.WriteString(")\n\n")
		} else if schema.GoType == "struct" {
			sb.WriteString(fmt.Sprintf("type %s struct {\n", schema.Name))
			for _, prop := range schema.Properties {
				tags := fmt.Sprintf("`json:\"%s", prop.JSONTag)
				if !prop.Required {
					tags += ",omitempty"
				}
				tags += "\""
				if prop.ValidateTag != "" {
					tags += fmt.Sprintf(" validate:\"%s\"", prop.ValidateTag)
				}
				tags += "`"

				comment := ""
				if prop.Description != "" {
					comment = " // " + prop.Description
				}

				sb.WriteString(fmt.Sprintf("\t%s %s %s%s\n", prop.GoName, prop.Type, tags, comment))
			}
			sb.WriteString("}\n\n")
		}
	}

	return sb.String()
}

func (g *OpenAPIGenerator) generateDTOsInline(dtos []SchemaInfo, imports []string) string {
	return g.generateModelsInline(dtos, imports)
}

func (g *OpenAPIGenerator) generateHandlersInline(endpoints []EndpointInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/labstack/echo/v5\"\n")
	sb.WriteString(")\n\n")

	for _, ep := range endpoints {
		// Write handler comment
		if ep.Summary != "" {
			sb.WriteString(fmt.Sprintf("// %s %s\n", ep.FuncName, ep.Summary))
		}
		if ep.Description != "" {
			sb.WriteString(fmt.Sprintf("// %s\n", ep.Description))
		}

		// Write function signature
		sig := fmt.Sprintf("func %s(c *echo.Context", ep.FuncName)
		if ep.HasRequest {
			sig += fmt.Sprintf(", req %s", ep.RequestType)
		}
		sig += ") ("
		if ep.HasResponse {
			sig += ep.ResponseType + ", "
		}
		sig += "error) {\n"
		sb.WriteString(sig)

		// Write TODO body
		sb.WriteString(fmt.Sprintf("\t// TODO: Implement %s\n", ep.OperationID))
		if ep.HasResponse {
			sb.WriteString(fmt.Sprintf("\treturn %s{}, nil\n", ep.ResponseType))
		} else {
			sb.WriteString("\treturn nil\n")
		}
		sb.WriteString("}\n\n")
	}

	return sb.String()
}

func (g *OpenAPIGenerator) generateRoutesInline(endpoints []EndpointInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/abdussamadbello/echonext\"\n")
	sb.WriteString(")\n\n")
	sb.WriteString("// RegisterRoutes registers all generated routes\n")
	sb.WriteString("func RegisterRoutes(app *echonext.App) {\n")

	for _, ep := range endpoints {
		tags := "[]string{"
		for i, tag := range ep.Tags {
			if i > 0 {
				tags += ", "
			}
			tags += fmt.Sprintf("%q", tag)
		}
		tags += "}"

		sb.WriteString(fmt.Sprintf("\tapp.%s(%q, %s, echonext.Route{\n", ep.Method, ep.Path, ep.FuncName))
		sb.WriteString(fmt.Sprintf("\t\tSummary: %q,\n", ep.Summary))
		sb.WriteString(fmt.Sprintf("\t\tTags:    %s,\n", tags))
		sb.WriteString("\t})\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *OpenAPIGenerator) generateTestsInline(endpoints []EndpointInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"net/http\"\n")
	sb.WriteString("\t\"net/http/httptest\"\n")
	sb.WriteString("\t\"testing\"\n\n")
	sb.WriteString("\t\"github.com/abdussamadbello/echonext\"\n")
	sb.WriteString("\t\"github.com/stretchr/testify/assert\"\n")
	sb.WriteString(")\n\n")

	for _, ep := range endpoints {
		testName := "Test" + ep.FuncName
		sb.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
		sb.WriteString("\tapp := echonext.New()\n")
		sb.WriteString("\tRegisterRoutes(app)\n\n")
		sb.WriteString(fmt.Sprintf("\treq := httptest.NewRequest(%q, %q, nil)\n", ep.Method, ep.Path))
		sb.WriteString("\trec := httptest.NewRecorder()\n")
		sb.WriteString("\tapp.ServeHTTP(rec, req)\n\n")
		sb.WriteString("\t// TODO: Add assertions\n")
		sb.WriteString("\tassert.NotEqual(t, http.StatusNotFound, rec.Code)\n")
		sb.WriteString("}\n\n")
	}

	return sb.String()
}
