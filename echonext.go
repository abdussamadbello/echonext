// Package echonext provides a type-safe wrapper around Echo with automatic OpenAPI generation and validation
package echonext

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// App represents an EchoNext application
type App struct {
	*echo.Echo
	spec      *openapi3.T
	validator *validator.Validate
	routes    []RouteInfo
}

// RouteInfo stores metadata about a route for OpenAPI generation
type RouteInfo struct {
	Method       string
	Path         string
	Handler      interface{}
	Summary      string
	Description  string
	Tags         []string
	RequestType  reflect.Type
	ResponseType reflect.Type
}

// Route configures route metadata for OpenAPI generation
type Route struct {
	Summary     string
	Description string
	Tags        []string
}

// Response wraps API responses with a standard structure
type Response[T any] struct {
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

// New creates a new EchoNext application
func New() *App {
	e := echo.New()
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "API",
			Version: "1.0.0",
		},
		Paths: openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{},
		},
	}

	return &App{
		Echo:      e,
		spec:      spec,
		validator: validator.New(),
		routes:    []RouteInfo{},
	}
}

// SetInfo sets the API information for OpenAPI spec
func (app *App) SetInfo(title, version, description string) {
	app.spec.Info.Title = title
	app.spec.Info.Version = version
	app.spec.Info.Description = description
}

// GET registers a typed GET endpoint
func (app *App) GET(path string, handler interface{}, opts ...Route) {
	app.registerRoute("GET", path, handler, opts...)
}

// POST registers a typed POST endpoint
func (app *App) POST(path string, handler interface{}, opts ...Route) {
	app.registerRoute("POST", path, handler, opts...)
}

// PUT registers a typed PUT endpoint
func (app *App) PUT(path string, handler interface{}, opts ...Route) {
	app.registerRoute("PUT", path, handler, opts...)
}

// PATCH registers a typed PATCH endpoint
func (app *App) PATCH(path string, handler interface{}, opts ...Route) {
	app.registerRoute("PATCH", path, handler, opts...)
}

// DELETE registers a typed DELETE endpoint
func (app *App) DELETE(path string, handler interface{}, opts ...Route) {
	app.registerRoute("DELETE", path, handler, opts...)
}

// registerRoute registers a route with type information
func (app *App) registerRoute(method, path string, handler interface{}, opts ...Route) {
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		panic("handler must be a function")
	}

	// Extract request and response types
	var requestType, responseType reflect.Type
	if handlerType.NumIn() > 1 {
		requestType = handlerType.In(1)
	}
	if handlerType.NumOut() > 0 {
		responseType = handlerType.Out(0)
	}

	// Store route info for OpenAPI generation
	routeInfo := RouteInfo{
		Method:       method,
		Path:         path,
		Handler:      handler,
		RequestType:  requestType,
		ResponseType: responseType,
	}

	if len(opts) > 0 {
		routeInfo.Summary = opts[0].Summary
		routeInfo.Description = opts[0].Description
		routeInfo.Tags = opts[0].Tags
	}

	app.routes = append(app.routes, routeInfo)

	// Create Echo handler
	echoHandler := app.createEchoHandler(handler, requestType, responseType)

	switch method {
	case "GET":
		app.Echo.GET(path, echoHandler)
	case "POST":
		app.Echo.POST(path, echoHandler)
	case "PUT":
		app.Echo.PUT(path, echoHandler)
	case "PATCH":
		app.Echo.PATCH(path, echoHandler)
	case "DELETE":
		app.Echo.DELETE(path, echoHandler)
	}
}

// createEchoHandler wraps typed handlers for Echo
func (app *App) createEchoHandler(handler interface{}, requestType, responseType reflect.Type) echo.HandlerFunc {
	handlerValue := reflect.ValueOf(handler)

	return func(c echo.Context) error {
		args := []reflect.Value{reflect.ValueOf(c)}

		// Handle request binding if handler expects input
		if requestType != nil {
			reqPtr := reflect.New(requestType)
			req := reqPtr.Interface()

			// Bind based on content type and method
			if c.Request().Method == "GET" || c.Request().Method == "DELETE" {
				// Bind query parameters
				if err := (&echo.DefaultBinder{}).BindQueryParams(c, req); err != nil {
					return c.JSON(http.StatusBadRequest, Response[any]{
						Error:   fmt.Sprintf("Invalid query parameters: %v", err),
						Success: false,
					})
				}
			} else {
				// Bind JSON body for POST/PUT/PATCH
				if err := c.Bind(req); err != nil {
					return c.JSON(http.StatusBadRequest, Response[any]{
						Error:   fmt.Sprintf("Invalid request body: %v", err),
						Success: false,
					})
				}
			}

			// Bind path parameters
			if err := (&echo.DefaultBinder{}).BindPathParams(c, req); err != nil {
				return c.JSON(http.StatusBadRequest, Response[any]{
					Error:   fmt.Sprintf("Invalid path parameters: %v", err),
					Success: false,
				})
			}

			// Validate request
			if err := app.validator.Struct(req); err != nil {
				return c.JSON(http.StatusBadRequest, Response[any]{
					Error:   fmt.Sprintf("Validation failed: %v", err),
					Success: false,
				})
			}

			args = append(args, reqPtr.Elem())
		}

		// Call handler
		results := handlerValue.Call(args)

		// Handle response
		if len(results) > 0 {
			// Check if last result is an error
			if len(results) > 1 {
				if err, ok := results[len(results)-1].Interface().(error); ok && err != nil {
					// Handle echo.HTTPError specially
					if he, ok := err.(*echo.HTTPError); ok {
						return c.JSON(he.Code, Response[any]{
							Error:   fmt.Sprintf("%v", he.Message),
							Success: false,
						})
					}
					return c.JSON(http.StatusInternalServerError, Response[any]{
						Error:   err.Error(),
						Success: false,
					})
				}
			}

			// Return successful response
			if results[0].IsValid() && !results[0].IsZero() {
				return c.JSON(http.StatusOK, Response[any]{
					Data:    results[0].Interface(),
					Success: true,
				})
			}
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// GenerateOpenAPISpec generates OpenAPI specification from registered routes
func (app *App) GenerateOpenAPISpec() *openapi3.T {
	for _, route := range app.routes {
		app.addRouteToSpec(route)
	}
	return app.spec
}

// addRouteToSpec adds a route to the OpenAPI specification
func (app *App) addRouteToSpec(route RouteInfo) {
	path := route.Path
	// Convert Echo path params to OpenAPI format
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	path = strings.Join(parts, "/")

	if app.spec.Paths[path] == nil {
		app.spec.Paths[path] = &openapi3.PathItem{}
	}

	operation := &openapi3.Operation{
		Summary:     route.Summary,
		Description: route.Description,
		Tags:        route.Tags,
		Responses:   openapi3.Responses{},
		Parameters:  openapi3.Parameters{},
	}

	// Extract path parameters
	pathParts := strings.Split(route.Path, "/")
	for _, part := range pathParts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			param := &openapi3.Parameter{
				Name:     paramName,
				In:       "path",
				Required: true,
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: "string"},
				},
			}
			operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: param})
		}
	}

	// Add request body schema if applicable
	if route.RequestType != nil {
		if route.Method == "GET" || route.Method == "DELETE" {
			// Add query parameters
			app.addQueryParameters(operation, route.RequestType)
		} else {
			// Add request body for POST/PUT/PATCH
			schema := app.generateSchema(route.RequestType)
			requestBody := &openapi3.RequestBody{
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{
							Value: schema,
						},
					},
				},
				Required: true,
			}
			operation.RequestBody = &openapi3.RequestBodyRef{Value: requestBody}
		}
	}

	// Add response schema
	if route.ResponseType != nil {
		schema := app.generateSchema(route.ResponseType)
		responseSchema := &openapi3.Schema{
			Type: "object",
			Properties: openapi3.Schemas{
				"success": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: "boolean"},
				},
				"data": &openapi3.SchemaRef{
					Value: schema,
				},
				"error": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: "string"},
				},
			},
		}

		operation.Responses["200"] = &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: strPtr("Successful response"),
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{Value: responseSchema},
					},
				},
			},
		}
	}

	// Add error responses
	errorSchema := &openapi3.Schema{
		Type: "object",
		Properties: openapi3.Schemas{
			"success": &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: "boolean", Default: false},
			},
			"error": &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: "string"},
			},
		},
	}

	operation.Responses["400"] = &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: strPtr("Bad request"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{Value: errorSchema},
				},
			},
		},
	}

	operation.Responses["500"] = &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: strPtr("Internal server error"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{Value: errorSchema},
				},
			},
		},
	}

	// Set operation on the path
	switch route.Method {
	case "GET":
		app.spec.Paths[path].Get = operation
	case "POST":
		app.spec.Paths[path].Post = operation
	case "PUT":
		app.spec.Paths[path].Put = operation
	case "PATCH":
		app.spec.Paths[path].Patch = operation
	case "DELETE":
		app.spec.Paths[path].Delete = operation
	}
}

// addQueryParameters adds query parameters to operation from struct
func (app *App) addQueryParameters(operation *openapi3.Operation, t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		queryTag := field.Tag.Get("query")
		if queryTag == "" || queryTag == "-" {
			continue
		}

		required := false
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			required = strings.Contains(validateTag, "required")
		}

		param := &openapi3.Parameter{
			Name:     queryTag,
			In:       "query",
			Required: required,
			Schema: &openapi3.SchemaRef{
				Value: app.generateSchema(field.Type),
			},
		}

		operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: param})
	}
}

// generateSchema generates OpenAPI schema from Go type
func (app *App) generateSchema(t reflect.Type) *openapi3.Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &openapi3.Schema{Type: "string"}
	case reflect.Int, reflect.Int32, reflect.Int64:
		return &openapi3.Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &openapi3.Schema{Type: "number"}
	case reflect.Bool:
		return &openapi3.Schema{Type: "boolean"}
	case reflect.Slice:
		return &openapi3.Schema{
			Type:  "array",
			Items: &openapi3.SchemaRef{Value: app.generateSchema(t.Elem())},
		}
	case reflect.Map:
		return &openapi3.Schema{
			Type: "object",
			AdditionalProperties: openapi3.AdditionalProperties{
				Schema: &openapi3.SchemaRef{Value: app.generateSchema(t.Elem())},
			},
		}
	case reflect.Struct:
		// Handle time.Time specially
		if t.String() == "time.Time" {
			return &openapi3.Schema{Type: "string", Format: "date-time"}
		}

		schema := &openapi3.Schema{
			Type:       "object",
			Properties: openapi3.Schemas{},
			Required:   []string{},
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			fieldName := field.Name
			omitempty := false
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				fieldName = parts[0]
				for _, part := range parts[1:] {
					if part == "omitempty" {
						omitempty = true
					}
				}
			}

			fieldSchema := app.generateSchema(field.Type)

			// Add validation from struct tags
			if validateTag := field.Tag.Get("validate"); validateTag != "" {
				if strings.Contains(validateTag, "required") && !omitempty {
					schema.Required = append(schema.Required, fieldName)
				}

				// Parse additional validations
				validations := strings.Split(validateTag, ",")
				for _, v := range validations {
					if strings.HasPrefix(v, "min=") {
						if val := strings.TrimPrefix(v, "min="); val != "" {
							if fieldSchema.Type == "string" {
								if minLen, err := strconv.Atoi(val); err == nil {
									fieldSchema.MinLength = uint64(minLen)
								}
							} else if fieldSchema.Type == "integer" || fieldSchema.Type == "number" {
								if min, err := strconv.ParseFloat(val, 64); err == nil {
									fieldSchema.Min = &min
								}
							}
						}
					}
					if strings.HasPrefix(v, "max=") {
						if val := strings.TrimPrefix(v, "max="); val != "" {
							if fieldSchema.Type == "string" {
								if maxLen, err := strconv.Atoi(val); err == nil {
									maxLenValue := uint64(maxLen)
									fieldSchema.MaxLength = &maxLenValue
								}
							} else if fieldSchema.Type == "integer" || fieldSchema.Type == "number" {
								if max, err := strconv.ParseFloat(val, 64); err == nil {
									fieldSchema.Max = &max
								}
							}
						}
					}
					if v == "email" {
						fieldSchema.Format = "email"
					}
					if strings.HasPrefix(v, "oneof=") {
						values := strings.Split(strings.TrimPrefix(v, "oneof="), " ")
						enums := make([]interface{}, len(values))
						for i, val := range values {
							enums[i] = val
						}
						fieldSchema.Enum = enums
					}
				}
			}

			schema.Properties[fieldName] = &openapi3.SchemaRef{Value: fieldSchema}
		}

		return schema
	default:
		return &openapi3.Schema{Type: "object"}
	}
}

// ServeOpenAPISpec serves the OpenAPI specification
func (app *App) ServeOpenAPISpec(path string) {
	app.Echo.GET(path, func(c echo.Context) error {
		return c.JSON(http.StatusOK, app.GenerateOpenAPISpec())
	})
}

// ServeSwaggerUI serves Swagger UI for API documentation
func (app *App) ServeSwaggerUI(path string, specPath string) {
	app.Echo.GET(path, func(c echo.Context) error {
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s - API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "%s",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.presets.standalone
                ],
                layout: "BaseLayout",
                deepLinking: true
            });
        }
    </script>
</body>
</html>`, app.spec.Info.Title, specPath)
		return c.HTML(http.StatusOK, html)
	})
}

// Helper functions
func strPtr(s string) *string {
	return &s
}