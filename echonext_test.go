package echonext_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abdussamadbello/echonext"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Test models
type TestUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

func TestEchoNextRoutes(t *testing.T) {
	// Create app
	app := echonext.New()

	// Register test route
	app.POST("/users", func(c echo.Context, req CreateUserRequest) (TestUser, error) {
		return TestUser{
			ID:    "123",
			Name:  req.Name,
			Email: req.Email,
		}, nil
	}, echonext.Route{
		Summary: "Create user",
		Tags:    []string{"Users"},
	})

	t.Run("successful request", func(t *testing.T) {
		// Create request
		reqBody := CreateUserRequest{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		// Execute request
		app.ServeHTTP(rec, req)

		// Assert response
		assert.Equal(t, http.StatusOK, rec.Code)

		var response echonext.Response[TestUser]
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "John Doe", response.Data.Name)
		assert.Equal(t, "john@example.com", response.Data.Email)
	})

	t.Run("validation error", func(t *testing.T) {
		// Create invalid request (missing email)
		reqBody := map[string]string{
			"name": "John",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		// Execute request
		app.ServeHTTP(rec, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response echonext.Response[any]
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Validation failed")
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		// Execute request
		app.ServeHTTP(rec, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response echonext.Response[any]
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "Invalid request body")
	})
}

func TestOpenAPIGeneration(t *testing.T) {
	app := echonext.New()
	app.SetInfo("Test API", "1.0.0", "Test API Description")

	// Register routes
	app.GET("/users", func(c echo.Context) ([]TestUser, error) {
		return []TestUser{{ID: "1", Name: "John", Email: "john@example.com"}}, nil
	}, echonext.Route{
		Summary: "List users",
		Tags:    []string{"Users"},
	})

	app.POST("/users", func(c echo.Context, req CreateUserRequest) (TestUser, error) {
		return TestUser{ID: "1", Name: req.Name, Email: req.Email}, nil
	}, echonext.Route{
		Summary: "Create user",
		Tags:    []string{"Users"},
	})

	// Generate spec
	spec := app.GenerateOpenAPISpec()

	// Assert spec
	assert.Equal(t, "3.0.0", spec.OpenAPI)
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "1.0.0", spec.Info.Version)

	// Check paths
	assert.NotNil(t, spec.Paths["/users"])
	assert.NotNil(t, spec.Paths["/users"].Get)
	assert.NotNil(t, spec.Paths["/users"].Post)

	// Check operation details
	assert.Equal(t, "List users", spec.Paths["/users"].Get.Summary)
	assert.Equal(t, []string{"Users"}, spec.Paths["/users"].Get.Tags)

	// Check request body for POST
	assert.NotNil(t, spec.Paths["/users"].Post.RequestBody)
}

func TestQueryParameters(t *testing.T) {
	app := echonext.New()

	type ListRequest struct {
		Page  int    `query:"page" validate:"min=1"`
		Limit int    `query:"limit" validate:"min=1,max=100"`
		Sort  string `query:"sort"`
	}

	app.GET("/items", func(c echo.Context, req ListRequest) (map[string]interface{}, error) {
		return map[string]interface{}{
			"page":  req.Page,
			"limit": req.Limit,
			"sort":  req.Sort,
		}, nil
	})

	// Test with query parameters
	req := httptest.NewRequest(http.MethodGet, "/items?page=2&limit=20&sort=name", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response echonext.Response[map[string]interface{}]
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, float64(2), response.Data["page"])
	assert.Equal(t, float64(20), response.Data["limit"])
	assert.Equal(t, "name", response.Data["sort"])
}

func TestErrorHandling(t *testing.T) {
	app := echonext.New()

	app.GET("/error", func(c echo.Context) (TestUser, error) {
		return TestUser{}, echo.NewHTTPError(404, "user not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response echonext.Response[any]
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "user not found", response.Error)
}

// Benchmark example
func BenchmarkEchoNext(b *testing.B) {
	app := echonext.New()

	app.POST("/users", func(c echo.Context, req CreateUserRequest) (TestUser, error) {
		return TestUser{
			ID:    "123",
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	reqBody := CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)
	}
}

// Integration test example
func TestIntegration(t *testing.T) {
	// Create a full app
	app := createTestApp()

	// Test creating a user
	createReq := CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
	}
	body, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var createResp echonext.Response[TestUser]
	json.Unmarshal(rec.Body.Bytes(), &createResp)
	userID := createResp.Data.ID

	// Test getting the created user
	req = httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
	rec = httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var getResp echonext.Response[TestUser]
	json.Unmarshal(rec.Body.Bytes(), &getResp)
	assert.Equal(t, "Alice", getResp.Data.Name)
}

// Helper to create test app
func createTestApp() *echonext.App {
	app := echonext.New()

	// In-memory storage
	users := make(map[string]*TestUser)

	app.POST("/users", func(c echo.Context, req CreateUserRequest) (TestUser, error) {
		user := TestUser{
			ID:    generateTestID(),
			Name:  req.Name,
			Email: req.Email,
		}
		users[user.ID] = &user
		return user, nil
	})

	app.GET("/users/:id", func(c echo.Context) (TestUser, error) {
		id := c.Param("id")
		user, exists := users[id]
		if !exists {
			return TestUser{}, echo.NewHTTPError(404, "user not found")
		}
		return *user, nil
	})

	return app
}

func generateTestID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func TestAdvancedOpenAPIFeatures(t *testing.T) {
	app := echonext.New()
	app.SetInfo("Advanced API", "2.0.0", "API with advanced OpenAPI features")
	app.SetContact("API Team", "https://example.com/support", "api@example.com")
	app.SetLicense("MIT", "https://opensource.org/licenses/MIT")
	app.SetServers([]echonext.Server{
		{URL: "https://api.example.com/v1", Description: "Production"},
		{URL: "https://staging.example.com/v1", Description: "Staging"},
	})

	// Add security schemes
	app.AddSecurityScheme("bearerAuth", echonext.Security{
		Type:   "bearer",
		Scheme: "JWT",
	})
	app.AddSecurityScheme("apiKey", echonext.Security{
		Type: "apiKey",
		Name: "X-API-Key",
		In:   "header",
	})

	type AdvancedRequest struct {
		Name string `json:"name" validate:"required" example:"John Doe"`
		Age  int    `json:"age" validate:"min=1,max=120" example:"30"`
	}

	type AdvancedResponse struct {
		ID   string `json:"id" example:"123"`
		Name string `json:"name" example:"John Doe"`
	}

	// Register route with advanced features
	app.POST("/advanced", func(c echo.Context, req AdvancedRequest) (AdvancedResponse, error) {
		return AdvancedResponse{
			ID:   "123",
			Name: req.Name,
		}, nil
	}, echonext.Route{
		Summary:       "Advanced endpoint",
		Description:   "Demonstrates advanced OpenAPI features",
		Tags:          []string{"Advanced"},
		SuccessStatus: 201,
		Security: []echonext.Security{
			{Type: "bearer"},
			{Type: "apiKey", Name: "X-API-Key"},
		},
		RequestHeaders: map[string]echonext.HeaderInfo{
			"X-Request-ID": {
				Description: "Unique request identifier",
				Required:    true,
				Schema:      "string",
			},
		},
		ResponseHeaders: map[string]echonext.HeaderInfo{
			"X-Rate-Limit": {
				Description: "Requests remaining",
				Schema:      "integer",
			},
		},
		ContentTypes: []string{"application/json", "application/xml"},
		Examples: map[string]interface{}{
			"example1": map[string]interface{}{
				"name": "John Doe",
				"age":  30,
			},
		},
	})

	// Generate spec and test
	spec := app.GenerateOpenAPISpec()

	// Test basic info
	assert.Equal(t, "Advanced API", spec.Info.Title)
	assert.Equal(t, "2.0.0", spec.Info.Version)
	assert.NotNil(t, spec.Info.Contact)
	assert.Equal(t, "API Team", spec.Info.Contact.Name)
	assert.NotNil(t, spec.Info.License)
	assert.Equal(t, "MIT", spec.Info.License.Name)

	// Test servers
	assert.Len(t, spec.Servers, 2)
	assert.Equal(t, "https://api.example.com/v1", spec.Servers[0].URL)

	// Test security schemes
	assert.NotNil(t, spec.Components.SecuritySchemes)
	assert.Contains(t, spec.Components.SecuritySchemes, "bearerAuth")
	assert.Contains(t, spec.Components.SecuritySchemes, "apiKey")

	// Test route features
	assert.NotNil(t, spec.Paths["/advanced"])
	post := spec.Paths["/advanced"].Post
	assert.NotNil(t, post)

	// Test security requirements
	assert.NotNil(t, post.Security)
	assert.Greater(t, len(*post.Security), 0)

	// Test response status
	assert.Contains(t, post.Responses, "201")
	assert.NotContains(t, post.Responses, "200")

	// Test request body examples
	assert.NotNil(t, post.RequestBody)
	jsonContent := post.RequestBody.Value.Content["application/json"]
	assert.NotNil(t, jsonContent)
	assert.NotNil(t, jsonContent.Examples)
}

func TestCustomStatusCodes(t *testing.T) {
	app := echonext.New()

	app.POST("/create", func(c echo.Context, req TestUser) (TestUser, error) {
		return req, nil
	}, echonext.Route{
		SuccessStatus: 201,
	})

	reqBody := TestUser{Name: "John", Email: "john@example.com"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	// Should return 201 Created instead of 200 OK
	assert.Equal(t, 201, rec.Code)
}
