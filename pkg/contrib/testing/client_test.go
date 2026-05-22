package testing

import (
	"net/http"
	"testing"

	"github.com/abdussamadbello/echonext"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test types
type TestUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// =============================================================================
// APIClient Tests
// =============================================================================

func setupTestApp() *echonext.App {
	app := echonext.New()

	// GET endpoint
	app.GET("/users", func(c *echo.Context) ([]TestUser, error) {
		return []TestUser{
			{ID: "1", Name: "John", Email: "john@example.com"},
		}, nil
	})

	// POST endpoint
	app.POST("/users", func(c *echo.Context, req CreateUserRequest) (TestUser, error) {
		return TestUser{
			ID:    "2",
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	// PUT endpoint
	app.PUT("/users/:id", func(c *echo.Context, req CreateUserRequest) (TestUser, error) {
		return TestUser{
			ID:    c.Param("id"),
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	// DELETE endpoint
	app.DELETE("/users/:id", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// Authenticated endpoint
	app.GET("/me", func(c *echo.Context) (map[string]string, error) {
		auth := c.Request().Header.Get("Authorization")
		return map[string]string{"auth": auth}, nil
	})

	// Custom header endpoint
	app.GET("/headers", func(c *echo.Context) (map[string]string, error) {
		return map[string]string{
			"x-custom": c.Request().Header.Get("X-Custom"),
		}, nil
	})

	return app
}

func TestNewAPIClient(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	require.NotNil(t, client)
	assert.NotNil(t, client.app)
	assert.NotNil(t, client.headers)
}

func TestAPIClientGET(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp := client.GET("/users")

	assert.Nil(t, resp.Error())
	assert.Equal(t, http.StatusOK, resp.Status())
	assert.True(t, resp.IsSuccess())
	assert.False(t, resp.IsError())
}

func TestAPIClientPOST(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp := client.POST("/users", CreateUserRequest{
		Name:  "Jane",
		Email: "jane@example.com",
	})

	assert.Nil(t, resp.Error())
	assert.Equal(t, http.StatusOK, resp.Status())

	var result echonext.Response[TestUser]
	err := resp.JSON(&result)
	require.NoError(t, err)
	assert.Equal(t, "Jane", result.Data.Name)
}

func TestAPIClientPUT(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp := client.PUT("/users/123", CreateUserRequest{
		Name:  "Updated",
		Email: "updated@example.com",
	})

	assert.Nil(t, resp.Error())
	assert.Equal(t, http.StatusOK, resp.Status())
}

func TestAPIClientPATCH(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	// PATCH will get 405 Method Not Allowed since we didn't define PATCH
	resp := client.PATCH("/users/123", CreateUserRequest{
		Name:  "Patched",
		Email: "patched@example.com",
	})

	// Will be 405 since we didn't define PATCH endpoint
	assert.Equal(t, http.StatusMethodNotAllowed, resp.Status())
}

func TestAPIClientDELETE(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp := client.DELETE("/users/123")

	assert.Nil(t, resp.Error())
	assert.Equal(t, http.StatusNoContent, resp.Status())
}

func TestAPIClientWithAuth(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app).WithAuth("my-token")

	resp := client.GET("/me")

	assert.Nil(t, resp.Error())
	assert.Contains(t, resp.String(), "Bearer my-token")
}

func TestAPIClientWithBasicAuth(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app).WithBasicAuth("user", "pass")

	resp := client.GET("/me")

	assert.Nil(t, resp.Error())
	assert.Contains(t, resp.String(), "Basic")
}

func TestAPIClientWithHeader(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app).WithHeader("X-Custom", "custom-value")

	resp := client.GET("/headers")

	assert.Nil(t, resp.Error())
	assert.Contains(t, resp.String(), "custom-value")
}

func TestAPIClientChaining(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app).
		WithHeader("X-Custom", "value1").
		WithHeader("X-Another", "value2").
		WithAuth("token123")

	resp := client.GET("/me")

	assert.Nil(t, resp.Error())
	assert.Contains(t, resp.String(), "Bearer token123")
}

// =============================================================================
// Response Tests
// =============================================================================

func TestResponseJSON(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp := client.GET("/users")

	var result echonext.Response[[]TestUser]
	err := resp.JSON(&result)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "John", result.Data[0].Name)
}

func TestResponseString(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp := client.GET("/users")

	str := resp.String()
	assert.Contains(t, str, "John")
	assert.Contains(t, str, "john@example.com")
}

func TestResponseIsSuccess(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	tests := []struct {
		name      string
		path      string
		isSuccess bool
	}{
		{"200 OK", "/users", true},
		{"404 Not Found", "/nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := client.GET(tt.path)
			assert.Equal(t, tt.isSuccess, resp.IsSuccess())
		})
	}
}

func TestResponseIsError(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp404 := client.GET("/nonexistent")
	assert.True(t, resp404.IsError())

	resp200 := client.GET("/users")
	assert.False(t, resp200.IsError())
}

func TestResponseGetHeader(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	resp := client.GET("/users")

	contentType := resp.GetHeader("Content-Type")
	assert.Contains(t, contentType, "application/json")
}

// =============================================================================
// Response Assertion Tests
// =============================================================================

// MockT implements TestingT for testing assertion methods
type MockT struct {
	errors []string
	failed bool
}

func (m *MockT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, format)
}

func (m *MockT) FailNow() {
	m.failed = true
}

func (m *MockT) Failed() bool {
	return m.failed
}

func TestResponseAssertStatus(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)
	resp := client.GET("/users")

	mockT := &MockT{}

	// Should pass
	resp.AssertStatus(mockT, 200)
	assert.Empty(t, mockT.errors)

	// Should fail
	resp.AssertStatus(mockT, 404)
	assert.NotEmpty(t, mockT.errors)
}

func TestResponseAssertSuccess(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	mockT := &MockT{}

	// Should pass
	resp200 := client.GET("/users")
	resp200.AssertSuccess(mockT)
	assert.Empty(t, mockT.errors)

	// Should fail
	mockT2 := &MockT{}
	resp404 := client.GET("/nonexistent")
	resp404.AssertSuccess(mockT2)
	assert.NotEmpty(t, mockT2.errors)
}

func TestResponseAssertError(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	mockT := &MockT{}

	// Should pass for 404
	resp404 := client.GET("/nonexistent")
	resp404.AssertError(mockT)
	assert.Empty(t, mockT.errors)

	// Should fail for 200
	mockT2 := &MockT{}
	resp200 := client.GET("/users")
	resp200.AssertError(mockT2)
	assert.NotEmpty(t, mockT2.errors)
}

func TestResponseAssertJSON(t *testing.T) {
	app := setupTestApp()
	client := NewAPIClient(app)

	mockT := &MockT{}

	resp := client.GET("/users")

	// This will compare JSON structure
	expected := echonext.Response[[]TestUser]{
		Success: true,
		Data: []TestUser{
			{ID: "1", Name: "John", Email: "john@example.com"},
		},
	}

	resp.AssertJSON(mockT, expected)
	assert.Empty(t, mockT.errors)
}
