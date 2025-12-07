// Package testing provides optional testing utilities for EchoNext applications.
// This is a contrib package and is completely optional - you can use standard testing if preferred.
package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/abdussamadbello/echonext"
)

// APIClient provides a convenient way to test EchoNext APIs
type APIClient struct {
	app     *echonext.App
	headers map[string]string
	token   string
}

// NewAPIClient creates a new API testing client
func NewAPIClient(app *echonext.App) *APIClient {
	return &APIClient{
		app:     app,
		headers: make(map[string]string),
	}
}

// WithHeader adds a header to all requests
func (c *APIClient) WithHeader(key, value string) *APIClient {
	c.headers[key] = value
	return c
}

// WithAuth sets the Authorization header with a Bearer token
func (c *APIClient) WithAuth(token string) *APIClient {
	c.token = token
	return c
}

// WithBasicAuth sets the Authorization header with Basic auth
func (c *APIClient) WithBasicAuth(username, password string) *APIClient {
	req := &http.Request{}
	req.SetBasicAuth(username, password)
	c.headers["Authorization"] = req.Header.Get("Authorization")
	return c
}

// GET sends a GET request and returns the response
func (c *APIClient) GET(path string) *Response {
	return c.request(http.MethodGet, path, nil)
}

// POST sends a POST request with the given body
func (c *APIClient) POST(path string, body interface{}) *Response {
	return c.request(http.MethodPost, path, body)
}

// PUT sends a PUT request with the given body
func (c *APIClient) PUT(path string, body interface{}) *Response {
	return c.request(http.MethodPut, path, body)
}

// PATCH sends a PATCH request with the given body
func (c *APIClient) PATCH(path string, body interface{}) *Response {
	return c.request(http.MethodPatch, path, body)
}

// DELETE sends a DELETE request
func (c *APIClient) DELETE(path string) *Response {
	return c.request(http.MethodDelete, path, nil)
}

// request sends an HTTP request and returns a Response
func (c *APIClient) request(method, path string, body interface{}) *Response {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return &Response{err: fmt.Errorf("failed to marshal request body: %w", err)}
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)

	// Set content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Add auth token if set
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	rec := httptest.NewRecorder()
	c.app.ServeHTTP(rec, req)

	return &Response{
		StatusCode: rec.Code,
		Body:       rec.Body.Bytes(),
		Header:     rec.Header(),
		recorder:   rec,
	}
}

// Response represents an HTTP response for testing
type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
	recorder   *httptest.ResponseRecorder
	err        error
}

// Error returns any error that occurred during the request
func (r *Response) Error() error {
	return r.err
}

// JSON unmarshals the response body into the target
func (r *Response) JSON(target interface{}) error {
	if r.err != nil {
		return r.err
	}
	if err := json.Unmarshal(r.Body, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

// String returns the response body as a string
func (r *Response) String() string {
	return string(r.Body)
}

// Status returns the HTTP status code
func (r *Response) Status() int {
	return r.StatusCode
}

// IsSuccess returns true if the status code is 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError returns true if the status code is 4xx or 5xx
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}

// GetHeader returns the value of a response header
func (r *Response) GetHeader(key string) string {
	return r.Header.Get(key)
}

// AssertStatus checks if the response has the expected status code
func (r *Response) AssertStatus(t TestingT, expected int) *Response {
	if r.StatusCode != expected {
		t.Errorf("Expected status %d, got %d. Body: %s", expected, r.StatusCode, r.String())
	}
	return r
}

// AssertSuccess checks if the response is successful (2xx)
func (r *Response) AssertSuccess(t TestingT) *Response {
	if !r.IsSuccess() {
		t.Errorf("Expected successful response, got %d. Body: %s", r.StatusCode, r.String())
	}
	return r
}

// AssertError checks if the response is an error (4xx or 5xx)
func (r *Response) AssertError(t TestingT) *Response {
	if !r.IsError() {
		t.Errorf("Expected error response, got %d. Body: %s", r.StatusCode, r.String())
	}
	return r
}

// AssertJSON checks if the response body matches the expected JSON
func (r *Response) AssertJSON(t TestingT, expected interface{}) *Response {
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("Failed to marshal expected JSON: %v", err)
		return r
	}

	var actualData, expectedData interface{}
	if err := json.Unmarshal(r.Body, &actualData); err != nil {
		t.Errorf("Failed to unmarshal actual response: %v", err)
		return r
	}
	if err := json.Unmarshal(expectedJSON, &expectedData); err != nil {
		t.Errorf("Failed to unmarshal expected data: %v", err)
		return r
	}

	// Compare as strings for simplicity
	actualStr, _ := json.Marshal(actualData)
	expectedStr, _ := json.Marshal(expectedData)
	if string(actualStr) != string(expectedStr) {
		t.Errorf("JSON mismatch.\nExpected: %s\nGot: %s", expectedStr, actualStr)
	}

	return r
}

// TestingT is an interface that matches the standard testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
	Failed() bool
}
