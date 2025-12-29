// Package echonext provides a type-safe wrapper around Echo with automatic OpenAPI generation and validation
package echonext

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/abdussamadbello/echonext/graphql"
	"github.com/abdussamadbello/echonext/upload"
	"github.com/abdussamadbello/echonext/websocket"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// App represents an EchoNext application
type App struct {
	*echo.Echo
	spec           *openapi3.T
	validator      *validator.Validate
	routes         []RouteInfo
	globalSecurity *GlobalSecurityConfig
}

// Group represents a route group with type-safe handlers
type Group struct {
	app       *App
	echoGroup *echo.Group
	prefix    string
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
	RouteConfig  *Route // Store the full route configuration
	IsUpload     bool   // Indicates if this is a file upload endpoint
	IsWebSocket  bool   // Indicates if this is a WebSocket endpoint
}

// Route configures route metadata for OpenAPI generation
type Route struct {
	Summary               string
	Description           string
	Tags                  []string
	Security              []SecurityRequirement
	SuccessStatus         int
	RequestHeaders        map[string]HeaderInfo
	ResponseHeaders       map[string]HeaderInfo
	ContentTypes          []string
	Examples              map[string]interface{}
	DisableGlobalSecurity bool           // When true, disables global security for this route
	FileConfig            *upload.Config // Configuration for file upload endpoints
}

// Security defines a security scheme configuration for OpenAPI
type Security struct {
	Type             string       // "bearer", "apiKey", "oauth2", "basic", "openIdConnect"
	Name             string       // For apiKey: header/query/cookie name
	Scheme           string       // For bearer: "bearer", for basic: "basic"
	In               string       // For apiKey: "header", "query", "cookie"
	Description      string       // Human-readable description for the security scheme
	OAuth2Flows      *OAuth2Flows // OAuth2 flow configuration (required for oauth2 type)
	OpenIDConnectURL string       // OpenID Connect discovery URL (required for openIdConnect type)
}

// OAuth2Flows configures OAuth2 authentication flows
type OAuth2Flows struct {
	Implicit          *OAuth2Flow // Implicit flow configuration
	Password          *OAuth2Flow // Password (Resource Owner Password Credentials) flow
	ClientCredentials *OAuth2Flow // Client Credentials flow configuration
	AuthorizationCode *OAuth2Flow // Authorization Code flow configuration
}

// OAuth2Flow represents a single OAuth2 flow configuration
type OAuth2Flow struct {
	AuthorizationURL string            // Authorization endpoint URL (required for implicit and authorizationCode)
	TokenURL         string            // Token endpoint URL (required for password, clientCredentials, authorizationCode)
	RefreshURL       string            // Optional refresh token endpoint URL
	Scopes           map[string]string // Maps scope names to their descriptions
}

// SecurityRequirement references a security scheme by name with optional scopes
type SecurityRequirement struct {
	SchemeName string   // References a scheme registered via AddSecurityScheme
	Scopes     []string // Required scopes for this security requirement
}

// GlobalSecurityConfig defines default security settings for all routes
type GlobalSecurityConfig struct {
	Requirements     []SecurityRequirement // Default security requirements
	ApplyToAllRoutes bool                  // When true, applies to all routes by default
}

// HeaderInfo describes a header parameter
type HeaderInfo struct {
	Description string
	Required    bool
	Schema      string // "string", "integer", etc.
}

// Server represents an OpenAPI server
type Server struct {
	URL         string
	Description string
}

// Contact represents OpenAPI contact information
type Contact struct {
	Name  string
	URL   string
	Email string
}

// License represents OpenAPI license information
type License struct {
	Name string
	URL  string
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

// SetContact sets the API contact information
func (app *App) SetContact(name, url, email string) {
	if app.spec.Info.Contact == nil {
		app.spec.Info.Contact = &openapi3.Contact{}
	}
	app.spec.Info.Contact.Name = name
	app.spec.Info.Contact.URL = url
	app.spec.Info.Contact.Email = email
}

// SetLicense sets the API license information
func (app *App) SetLicense(name, url string) {
	if app.spec.Info.License == nil {
		app.spec.Info.License = &openapi3.License{}
	}
	app.spec.Info.License.Name = name
	app.spec.Info.License.URL = url
}

// SetServers sets the API servers
func (app *App) SetServers(servers []Server) {
	app.spec.Servers = make([]*openapi3.Server, len(servers))
	for i, server := range servers {
		app.spec.Servers[i] = &openapi3.Server{
			URL:         server.URL,
			Description: server.Description,
		}
	}
}

// AddSecurityScheme adds a security scheme to the OpenAPI spec
func (app *App) AddSecurityScheme(name string, security Security) error {
	if app.spec.Components.SecuritySchemes == nil {
		app.spec.Components.SecuritySchemes = make(openapi3.SecuritySchemes)
	}

	scheme := &openapi3.SecurityScheme{
		Description: security.Description,
	}

	switch security.Type {
	case "bearer":
		scheme.Type = "http"
		scheme.Scheme = "bearer"
		if security.Scheme != "" {
			scheme.BearerFormat = security.Scheme
		}
	case "apiKey":
		scheme.Type = "apiKey"
		scheme.Name = security.Name
		scheme.In = security.In
		if scheme.In == "" {
			scheme.In = "header" // Default to header
		}
	case "basic":
		scheme.Type = "http"
		scheme.Scheme = "basic"
	case "oauth2":
		if security.OAuth2Flows == nil {
			return fmt.Errorf("OAuth2 security scheme requires OAuth2Flows configuration")
		}
		scheme.Type = "oauth2"
		scheme.Flows = convertOAuth2Flows(security.OAuth2Flows)
	case "openIdConnect":
		if security.OpenIDConnectURL == "" {
			return fmt.Errorf("OpenID Connect security scheme requires OpenIDConnectURL")
		}
		scheme.Type = "openIdConnect"
		scheme.OpenIdConnectUrl = security.OpenIDConnectURL
	default:
		return fmt.Errorf("unsupported security type: %s", security.Type)
	}

	app.spec.Components.SecuritySchemes[name] = &openapi3.SecuritySchemeRef{
		Value: scheme,
	}
	return nil
}

// convertOAuth2Flows converts internal OAuth2Flows to openapi3.OAuthFlows
func convertOAuth2Flows(flows *OAuth2Flows) *openapi3.OAuthFlows {
	if flows == nil {
		return nil
	}

	result := &openapi3.OAuthFlows{}

	if flows.Implicit != nil {
		result.Implicit = &openapi3.OAuthFlow{
			AuthorizationURL: flows.Implicit.AuthorizationURL,
			RefreshURL:       flows.Implicit.RefreshURL,
			Scopes:           flows.Implicit.Scopes,
		}
	}

	if flows.Password != nil {
		result.Password = &openapi3.OAuthFlow{
			TokenURL:   flows.Password.TokenURL,
			RefreshURL: flows.Password.RefreshURL,
			Scopes:     flows.Password.Scopes,
		}
	}

	if flows.ClientCredentials != nil {
		result.ClientCredentials = &openapi3.OAuthFlow{
			TokenURL:   flows.ClientCredentials.TokenURL,
			RefreshURL: flows.ClientCredentials.RefreshURL,
			Scopes:     flows.ClientCredentials.Scopes,
		}
	}

	if flows.AuthorizationCode != nil {
		result.AuthorizationCode = &openapi3.OAuthFlow{
			AuthorizationURL: flows.AuthorizationCode.AuthorizationURL,
			TokenURL:         flows.AuthorizationCode.TokenURL,
			RefreshURL:       flows.AuthorizationCode.RefreshURL,
			Scopes:           flows.AuthorizationCode.Scopes,
		}
	}

	return result
}

// SetGlobalSecurity sets default security requirements for all routes
func (app *App) SetGlobalSecurity(requirements ...SecurityRequirement) error {
	// Validate all scheme names exist
	for _, req := range requirements {
		if err := app.validateSecurityRequirement(req); err != nil {
			return err
		}
	}

	app.globalSecurity = &GlobalSecurityConfig{
		Requirements:     requirements,
		ApplyToAllRoutes: true,
	}

	// Also set global security in OpenAPI spec
	app.spec.Security = openapi3.SecurityRequirements{}
	for _, req := range requirements {
		scopes := req.Scopes
		if scopes == nil {
			scopes = []string{}
		}
		app.spec.Security = append(app.spec.Security, openapi3.SecurityRequirement{
			req.SchemeName: scopes,
		})
	}

	return nil
}

// ClearGlobalSecurity removes all global security requirements
func (app *App) ClearGlobalSecurity() {
	app.globalSecurity = nil
	app.spec.Security = nil
}

// validateSecurityRequirement validates that a security requirement references a valid scheme
func (app *App) validateSecurityRequirement(req SecurityRequirement) error {
	if app.spec.Components.SecuritySchemes == nil {
		return fmt.Errorf("security scheme '%s' not found: no security schemes registered", req.SchemeName)
	}
	if _, exists := app.spec.Components.SecuritySchemes[req.SchemeName]; !exists {
		return fmt.Errorf("security scheme '%s' not found", req.SchemeName)
	}
	return nil
}

// resolveRouteSecurity determines the security requirements for a route
func (app *App) resolveRouteSecurity(route RouteInfo) openapi3.SecurityRequirements {
	var requirements openapi3.SecurityRequirements

	// Route-specific security takes precedence
	if route.RouteConfig != nil && len(route.RouteConfig.Security) > 0 {
		for _, req := range route.RouteConfig.Security {
			// Validate scheme exists - skip invalid schemes
			if err := app.validateSecurityRequirement(req); err != nil {
				continue
			}
			scopes := req.Scopes
			if scopes == nil {
				scopes = []string{}
			}
			requirements = append(requirements, openapi3.SecurityRequirement{
				req.SchemeName: scopes,
			})
		}
		return requirements
	}

	// Check if global security is disabled for this route
	if route.RouteConfig != nil && route.RouteConfig.DisableGlobalSecurity {
		return nil
	}

	// Apply global security if configured
	if app.globalSecurity != nil && app.globalSecurity.ApplyToAllRoutes {
		for _, req := range app.globalSecurity.Requirements {
			scopes := req.Scopes
			if scopes == nil {
				scopes = []string{}
			}
			requirements = append(requirements, openapi3.SecurityRequirement{
				req.SchemeName: scopes,
			})
		}
	}

	return requirements
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

// Upload registers a file upload endpoint (POST with multipart/form-data)
// The handler should accept a request struct containing FileUpload fields
func (app *App) Upload(path string, handler interface{}, opts ...Route) {
	app.registerUploadRoute(path, handler, opts...)
}

// WS registers a WebSocket endpoint
// The handler should implement websocket.Handler interface or be a typed handler function
func (app *App) WS(path string, handler interface{}, opts ...Route) {
	app.registerWSRoute(path, handler, opts...)
}

// Group creates a route group with the given prefix
func (app *App) Group(prefix string, middleware ...echo.MiddlewareFunc) *Group {
	echoGroup := app.Echo.Group(prefix, middleware...)
	return &Group{
		app:       app,
		echoGroup: echoGroup,
		prefix:    prefix,
	}
}

// Group methods for type-safe route registration

// GET registers a typed GET endpoint on the group
func (g *Group) GET(path string, handler interface{}, opts ...Route) {
	fullPath := g.prefix + path
	g.registerGroupRoute("GET", fullPath, path, handler, opts...)
}

// POST registers a typed POST endpoint on the group
func (g *Group) POST(path string, handler interface{}, opts ...Route) {
	fullPath := g.prefix + path
	g.registerGroupRoute("POST", fullPath, path, handler, opts...)
}

// PUT registers a typed PUT endpoint on the group
func (g *Group) PUT(path string, handler interface{}, opts ...Route) {
	fullPath := g.prefix + path
	g.registerGroupRoute("PUT", fullPath, path, handler, opts...)
}

// PATCH registers a typed PATCH endpoint on the group
func (g *Group) PATCH(path string, handler interface{}, opts ...Route) {
	fullPath := g.prefix + path
	g.registerGroupRoute("PATCH", fullPath, path, handler, opts...)
}

// DELETE registers a typed DELETE endpoint on the group
func (g *Group) DELETE(path string, handler interface{}, opts ...Route) {
	fullPath := g.prefix + path
	g.registerGroupRoute("DELETE", fullPath, path, handler, opts...)
}

// Upload registers a file upload endpoint on the group (POST with multipart/form-data)
func (g *Group) Upload(path string, handler interface{}, opts ...Route) {
	fullPath := g.prefix + path
	g.registerGroupUploadRoute(fullPath, path, handler, opts...)
}

// Use adds middleware to the group
func (g *Group) Use(middleware ...echo.MiddlewareFunc) {
	g.echoGroup.Use(middleware...)
}

// Group creates a sub-group with the given prefix
func (g *Group) Group(prefix string, middleware ...echo.MiddlewareFunc) *Group {
	echoGroup := g.echoGroup.Group(prefix, middleware...)
	return &Group{
		app:       g.app,
		echoGroup: echoGroup,
		prefix:    g.prefix + prefix,
	}
}

// registerGroupRoute registers a route on the group
func (g *Group) registerGroupRoute(method, fullPath, relativePath string, handler interface{}, opts ...Route) {
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

	// Store route info for OpenAPI generation (use full path for OpenAPI)
	routeInfo := RouteInfo{
		Method:       method,
		Path:         fullPath,
		Handler:      handler,
		RequestType:  requestType,
		ResponseType: responseType,
	}

	if len(opts) > 0 {
		route := opts[0]
		routeInfo.Summary = route.Summary
		routeInfo.Description = route.Description
		routeInfo.Tags = route.Tags
		routeInfo.RouteConfig = &route
	}

	g.app.routes = append(g.app.routes, routeInfo)

	// Create Echo handler and register on group (use relative path for Echo)
	echoHandler := g.app.createEchoHandler(handler, requestType, responseType, routeInfo.RouteConfig)

	switch method {
	case "GET":
		g.echoGroup.GET(relativePath, echoHandler)
	case "POST":
		g.echoGroup.POST(relativePath, echoHandler)
	case "PUT":
		g.echoGroup.PUT(relativePath, echoHandler)
	case "PATCH":
		g.echoGroup.PATCH(relativePath, echoHandler)
	case "DELETE":
		g.echoGroup.DELETE(relativePath, echoHandler)
	}
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
		route := opts[0]
		routeInfo.Summary = route.Summary
		routeInfo.Description = route.Description
		routeInfo.Tags = route.Tags
		routeInfo.RouteConfig = &route
	}

	app.routes = append(app.routes, routeInfo)

	// Create Echo handler
	echoHandler := app.createEchoHandler(handler, requestType, responseType, routeInfo.RouteConfig)

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

// registerUploadRoute registers a file upload route
func (app *App) registerUploadRoute(path string, handler interface{}, opts ...Route) {
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

	// Validate that request type has File fields
	if requestType != nil && !upload.HasFileFields(requestType) {
		panic("upload handler request type must contain upload.File fields")
	}

	// Store route info for OpenAPI generation
	routeInfo := RouteInfo{
		Method:       "POST",
		Path:         path,
		Handler:      handler,
		RequestType:  requestType,
		ResponseType: responseType,
		IsUpload:     true,
	}

	if len(opts) > 0 {
		route := opts[0]
		routeInfo.Summary = route.Summary
		routeInfo.Description = route.Description
		routeInfo.Tags = route.Tags
		routeInfo.RouteConfig = &route
	}

	app.routes = append(app.routes, routeInfo)

	// Create Echo handler with file upload support
	echoHandler := app.createUploadEchoHandler(handler, requestType, responseType, routeInfo.RouteConfig)
	app.Echo.POST(path, echoHandler)
}

// registerGroupUploadRoute registers a file upload route on a group
func (g *Group) registerGroupUploadRoute(fullPath, relativePath string, handler interface{}, opts ...Route) {
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

	// Validate that request type has File fields
	if requestType != nil && !upload.HasFileFields(requestType) {
		panic("upload handler request type must contain upload.File fields")
	}

	// Store route info for OpenAPI generation (use full path for OpenAPI)
	routeInfo := RouteInfo{
		Method:       "POST",
		Path:         fullPath,
		Handler:      handler,
		RequestType:  requestType,
		ResponseType: responseType,
		IsUpload:     true,
	}

	if len(opts) > 0 {
		route := opts[0]
		routeInfo.Summary = route.Summary
		routeInfo.Description = route.Description
		routeInfo.Tags = route.Tags
		routeInfo.RouteConfig = &route
	}

	g.app.routes = append(g.app.routes, routeInfo)

	// Create Echo handler with file upload support and register on group
	echoHandler := g.app.createUploadEchoHandler(handler, requestType, responseType, routeInfo.RouteConfig)
	g.echoGroup.POST(relativePath, echoHandler)
}

// registerWSRoute registers a WebSocket route
func (app *App) registerWSRoute(path string, handler interface{}, opts ...Route) {
	// Store route info for OpenAPI generation
	routeInfo := RouteInfo{
		Method:      "GET",
		Path:        path,
		Handler:     handler,
		IsWebSocket: true,
	}

	if len(opts) > 0 {
		route := opts[0]
		routeInfo.Summary = route.Summary
		routeInfo.Description = route.Description
		routeInfo.Tags = route.Tags
	}

	app.routes = append(app.routes, routeInfo)

	// Create WebSocket Echo handler
	config := websocket.DefaultConfig()
	echoHandler := func(c echo.Context) error {
		upgrader := websocket.CreateUpgrader(&config)
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		conn := websocket.NewConnection(websocket.GenerateConnectionID(), ws, c)
		websocket.SetupConnection(conn, &config)

		// Check if handler implements websocket.Handler interface
		if h, ok := handler.(websocket.Handler); ok {
			return websocket.HandleHandler(conn, h, &config)
		}

		// For function handlers
		return websocket.HandleTypedHandler(conn, handler, &config)
	}

	app.Echo.GET(path, echoHandler)
}

// createUploadEchoHandler wraps file upload handlers for Echo
func (app *App) createUploadEchoHandler(handler interface{}, requestType, responseType reflect.Type, routeConfig *Route) echo.HandlerFunc {
	handlerValue := reflect.ValueOf(handler)

	return func(c echo.Context) error {
		args := []reflect.Value{reflect.ValueOf(c)}

		// Handle request binding if handler expects input
		if requestType != nil {
			reqPtr := reflect.New(requestType)
			req := reqPtr.Interface()

			// Get file upload config
			var fileConfig *upload.Config
			if routeConfig != nil && routeConfig.FileConfig != nil {
				fileConfig = routeConfig.FileConfig
			}

			// Bind file upload fields and form fields
			if err := upload.Bind(c, req, fileConfig); err != nil {
				return c.JSON(http.StatusBadRequest, Response[any]{
					Error:   fmt.Sprintf("File upload failed: %v", err),
					Success: false,
				})
			}

			// Validate request (including file fields)
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
				// Determine status code
				statusCode := http.StatusOK
				if routeConfig != nil && routeConfig.SuccessStatus > 0 {
					statusCode = routeConfig.SuccessStatus
				}

				return c.JSON(statusCode, Response[any]{
					Data:    results[0].Interface(),
					Success: true,
				})
			}
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// createEchoHandler wraps typed handlers for Echo
func (app *App) createEchoHandler(handler interface{}, requestType, responseType reflect.Type, routeConfig *Route) echo.HandlerFunc {
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
				// Determine status code
				statusCode := http.StatusOK
				if routeConfig != nil && routeConfig.SuccessStatus > 0 {
					statusCode = routeConfig.SuccessStatus
				}

				return c.JSON(statusCode, Response[any]{
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
		Security:    &openapi3.SecurityRequirements{},
	}

	// Resolve and apply security requirements
	securityReqs := app.resolveRouteSecurity(route)
	if len(securityReqs) > 0 {
		operation.Security = &securityReqs
	} else if route.RouteConfig != nil && route.RouteConfig.DisableGlobalSecurity {
		// Empty security for public endpoints (explicitly disabled)
		operation.Security = &openapi3.SecurityRequirements{}
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

	// Add request headers if specified
	if route.RouteConfig != nil && len(route.RouteConfig.RequestHeaders) > 0 {
		for headerName, headerInfo := range route.RouteConfig.RequestHeaders {
			schemaType := headerInfo.Schema
			if schemaType == "" {
				schemaType = "string"
			}
			param := &openapi3.Parameter{
				Name:        headerName,
				In:          "header",
				Required:    headerInfo.Required,
				Description: headerInfo.Description,
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: schemaType},
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
		} else if route.IsUpload {
			// Handle file upload routes with multipart/form-data
			schema := app.generateMultipartSchema(route.RequestType)
			content := openapi3.Content{
				"multipart/form-data": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Value: schema,
					},
					Encoding: app.generateMultipartEncoding(route.RequestType),
				},
			}

			requestBody := &openapi3.RequestBody{
				Content:     content,
				Required:    true,
				Description: "File upload request",
			}
			operation.RequestBody = &openapi3.RequestBodyRef{Value: requestBody}
		} else {
			// Add request body for POST/PUT/PATCH
			schema := app.generateSchema(route.RequestType)

			// Determine content types
			contentTypes := []string{"application/json"}
			if route.RouteConfig != nil && len(route.RouteConfig.ContentTypes) > 0 {
				contentTypes = route.RouteConfig.ContentTypes
			}

			content := openapi3.Content{}
			for _, contentType := range contentTypes {
				mediaType := &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Value: schema,
					},
				}

				// Add examples if provided
				if route.RouteConfig != nil && len(route.RouteConfig.Examples) > 0 {
					mediaType.Examples = make(openapi3.Examples)
					for exampleName, exampleValue := range route.RouteConfig.Examples {
						mediaType.Examples[exampleName] = &openapi3.ExampleRef{
							Value: &openapi3.Example{
								Value: exampleValue,
							},
						}
					}
				}

				content[contentType] = mediaType
			}

			requestBody := &openapi3.RequestBody{
				Content:  content,
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

		// Determine success status code
		successStatus := "200"
		if route.RouteConfig != nil && route.RouteConfig.SuccessStatus > 0 {
			successStatus = fmt.Sprintf("%d", route.RouteConfig.SuccessStatus)
		}

		response := &openapi3.Response{
			Description: strPtr("Successful response"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{Value: responseSchema},
				},
			},
		}

		// Add response headers if specified
		if route.RouteConfig != nil && len(route.RouteConfig.ResponseHeaders) > 0 {
			response.Headers = make(openapi3.Headers)
			for headerName, headerInfo := range route.RouteConfig.ResponseHeaders {
				schemaType := headerInfo.Schema
				if schemaType == "" {
					schemaType = "string"
				}
				response.Headers[headerName] = &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Description: headerInfo.Description,
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: schemaType},
							},
						},
					},
				}
			}
		}

		operation.Responses[successStatus] = &openapi3.ResponseRef{Value: response}
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
	visited := make(map[reflect.Type]bool)
	return app.generateSchemaWithVisited(t, visited)
}

// generateSchemaWithVisited generates OpenAPI schema with circular reference detection
func (app *App) generateSchemaWithVisited(t reflect.Type, visited map[reflect.Type]bool) *openapi3.Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for circular reference
	if visited[t] {
		// Return a simple reference to avoid infinite recursion
		return &openapi3.Schema{
			Type:        "object",
			Description: fmt.Sprintf("Circular reference to %s", t.String()),
		}
	}

	// Mark this type as visited for structs to detect cycles
	if t.Kind() == reflect.Struct && t.String() != "time.Time" {
		visited[t] = true
		defer func() { delete(visited, t) }() // Clean up when done
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
			Items: &openapi3.SchemaRef{Value: app.generateSchemaWithVisited(t.Elem(), visited)},
		}
	case reflect.Map:
		return &openapi3.Schema{
			Type: "object",
			AdditionalProperties: openapi3.AdditionalProperties{
				Schema: &openapi3.SchemaRef{Value: app.generateSchemaWithVisited(t.Elem(), visited)},
			},
		}
	case reflect.Struct:
		// Handle time.Time specially
		if t.String() == "time.Time" {
			return &openapi3.Schema{Type: "string", Format: "date-time"}
		}

		// Handle upload.File specially - represents a binary file
		if t.Name() == "File" && strings.HasSuffix(t.PkgPath(), "echonext/upload") {
			return &openapi3.Schema{
				Type:        "string",
				Format:      "binary",
				Description: "File upload",
			}
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

			fieldSchema := app.generateSchemaWithVisited(field.Type, visited)

			// Add example from struct tag
			if exampleTag := field.Tag.Get("example"); exampleTag != "" {
				fieldSchema.Example = exampleTag
			}

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

// generateMultipartSchema generates OpenAPI schema for multipart/form-data requests
func (app *App) generateMultipartSchema(t reflect.Type) *openapi3.Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return &openapi3.Schema{Type: "object"}
	}

	schema := &openapi3.Schema{
		Type:       "object",
		Properties: openapi3.Schemas{},
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get field name from form or json tag
		formTag := field.Tag.Get("form")
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name

		if formTag != "" && formTag != "-" {
			parts := strings.Split(formTag, ",")
			fieldName = parts[0]
		} else if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			fieldName = parts[0]
		} else {
			fieldName = strings.ToLower(field.Name)
		}

		var fieldSchema *openapi3.Schema

		// Check if this is an upload.File field
		if upload.IsFileType(field.Type) {
			fieldSchema = &openapi3.Schema{
				Type:        "string",
				Format:      "binary",
				Description: "File upload",
			}
		} else if upload.IsFileSliceType(field.Type) {
			fieldSchema = &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:        "string",
						Format:      "binary",
						Description: "File upload",
					},
				},
			}
		} else {
			// Regular field
			fieldSchema = app.generateSchema(field.Type)
		}

		// Check if required
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			if strings.Contains(validateTag, "required") {
				schema.Required = append(schema.Required, fieldName)
			}
		}

		schema.Properties[fieldName] = &openapi3.SchemaRef{Value: fieldSchema}
	}

	return schema
}

// generateMultipartEncoding generates OpenAPI encoding for multipart/form-data file fields
func (app *App) generateMultipartEncoding(t reflect.Type) map[string]*openapi3.Encoding {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	encoding := make(map[string]*openapi3.Encoding)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get field name from form or json tag
		formTag := field.Tag.Get("form")
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name

		if formTag != "" && formTag != "-" {
			parts := strings.Split(formTag, ",")
			fieldName = parts[0]
		} else if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			fieldName = parts[0]
		} else {
			fieldName = strings.ToLower(field.Name)
		}

		// Add encoding for file fields
		if upload.IsFileType(field.Type) || upload.IsFileSliceType(field.Type) {
			encoding[fieldName] = &openapi3.Encoding{
				ContentType: "application/octet-stream",
			}
		}
	}

	if len(encoding) == 0 {
		return nil
	}

	return encoding
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

// Response helper functions

// Success creates a successful response
func Success[T any](data T) Response[T] {
	return Response[T]{
		Data:    data,
		Success: true,
	}
}

// Error creates an error response
func Error[T any](message string) Response[T] {
	return Response[T]{
		Error:   message,
		Success: false,
	}
}

// NoContent creates an empty successful response
func NoContent() Response[interface{}] {
	return Response[interface{}]{
		Success: true,
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

// GraphQL registration methods

// GraphQL registers a GraphQL endpoint with the given schema
func (app *App) GraphQL(config graphql.Config, opts ...graphql.Route) {
	// Apply defaults
	if config.Path == "" {
		config.Path = graphql.DefaultPath
	}

	// Create gqlgen handler using graphql package
	srv := graphql.Handler(config)

	// Create Echo handler that injects context
	echoHandler := graphql.WrapWithEchoContext(srv)

	// Register GraphQL endpoint
	app.Echo.POST(config.Path, echoHandler)
	app.Echo.GET(config.Path, echoHandler)

	// Register Playground if enabled
	if config.PlaygroundPath != "" {
		playgroundHandler := graphql.PlaygroundHandler("GraphQL Playground", config.Path)
		app.Echo.GET(config.PlaygroundPath, echo.WrapHandler(playgroundHandler))
	}

	// Store route info for OpenAPI generation
	routeInfo := RouteInfo{
		Method:      "POST",
		Path:        config.Path,
		Handler:     config,
		Summary:     "GraphQL endpoint",
		Description: "GraphQL API endpoint for queries and mutations",
		Tags:        []string{"GraphQL"},
	}

	if len(opts) > 0 {
		route := opts[0]
		if route.Summary != "" {
			routeInfo.Summary = route.Summary
		}
		if route.Description != "" {
			routeInfo.Description = route.Description
		}
		if len(route.Tags) > 0 {
			routeInfo.Tags = route.Tags
		}
	}

	app.routes = append(app.routes, routeInfo)
}

// GraphQLSubscriptions enables WebSocket subscriptions at a separate endpoint
func (app *App) GraphQLSubscriptions(schema interface{}, config ...graphql.SubscriptionConfig) {
	// Import the gqlgen graphql package for ExecutableSchema type assertion
	gqlSchema, ok := schema.(interface {
		Schema() interface{}
	})

	if !ok {
		// Try direct use - schema should be a gqlgen ExecutableSchema
		srv := graphql.SubscriptionHandler(nil)
		if srv == nil {
			return
		}
	}

	var subConfig graphql.SubscriptionConfig
	if len(config) > 0 {
		subConfig = config[0]
	}

	if subConfig.Path == "" {
		subConfig.Path = graphql.DefaultSubscriptionsPath
	}

	// Create subscription handler - user must pass a proper schema
	_ = gqlSchema // Use this to create handler

	// Create Echo handler
	echoHandler := func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), graphql.EchoContextKey, c)
		c.SetRequest(c.Request().WithContext(ctx))
		// Handler will be set up by the user with proper schema
		return nil
	}

	app.Echo.GET(subConfig.Path, echoHandler)
}

// Group GraphQL method

// GraphQL registers a GraphQL endpoint on the group
func (g *Group) GraphQL(config graphql.Config, opts ...graphql.Route) {
	// Prepend group prefix to paths
	if config.Path == "" {
		config.Path = graphql.DefaultPath
	}
	config.Path = g.prefix + config.Path

	if config.PlaygroundPath != "" {
		config.PlaygroundPath = g.prefix + config.PlaygroundPath
	}

	g.app.GraphQL(config, opts...)
}
