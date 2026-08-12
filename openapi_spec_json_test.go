package echonext_test

import (
	"encoding/json"
	"testing"

	"github.com/abdussamadbello/echonext"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPISpecJSONShape guards the serialized shape of the generated spec.
//
// kin-openapi models a schema's type as a list to support OpenAPI 3.1, so a
// schema type could regress from "string" to ["string"] and break consumers
// without failing any structural assertion.
func TestOpenAPISpecJSONShape(t *testing.T) {
	type CreateReq struct {
		Name string `json:"name" validate:"required,min=2,max=40"`
		Age  int    `json:"age" validate:"min=1,max=120"`
	}
	type CreateResp struct {
		ID   string   `json:"id"`
		Tags []string `json:"tags"`
	}

	app := echonext.New()
	app.POST("/things/:id", func(c *echo.Context, req CreateReq) (CreateResp, error) {
		return CreateResp{}, nil
	}, echonext.Route{Summary: "create"})

	raw, err := json.Marshal(app.GenerateOpenAPISpec())
	require.NoError(t, err)

	var spec map[string]any
	require.NoError(t, json.Unmarshal(raw, &spec))

	assert.Equal(t, "3.0.0", spec["openapi"])

	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok, "paths must serialize as an object")
	pathItem, ok := paths["/things/{id}"].(map[string]any)
	require.True(t, ok, "echo-style :id must serialize as OpenAPI {id}")
	post, ok := pathItem["post"].(map[string]any)
	require.True(t, ok)

	// Path parameter schema type must stay a scalar string.
	params, ok := post["parameters"].([]any)
	require.True(t, ok)
	require.Len(t, params, 1)
	param := params[0].(map[string]any)
	assert.Equal(t, "id", param["name"])
	assert.Equal(t, "path", param["in"])
	assert.Equal(t, "string", param["schema"].(map[string]any)["type"],
		"schema type must serialize as a string, not a list")

	// Request body schema: types and validation constraints.
	reqSchema := post["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	assert.Equal(t, "object", reqSchema["type"])
	props := reqSchema["properties"].(map[string]any)

	name := props["name"].(map[string]any)
	assert.Equal(t, "string", name["type"])
	assert.EqualValues(t, 2, name["minLength"])
	assert.EqualValues(t, 40, name["maxLength"])

	age := props["age"].(map[string]any)
	assert.Equal(t, "integer", age["type"])
	assert.EqualValues(t, 1, age["minimum"])
	assert.EqualValues(t, 120, age["maximum"])

	assert.Equal(t, []any{"name"}, reqSchema["required"])

	// Responses, including the array-typed field and the error responses.
	responses := post["responses"].(map[string]any)
	for _, status := range []string{"200", "400", "500"} {
		assert.Contains(t, responses, status)
	}

	data := responses["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)["properties"].(map[string]any)["data"].(map[string]any)
	assert.Equal(t, "object", data["type"])
	tags := data["properties"].(map[string]any)["tags"].(map[string]any)
	assert.Equal(t, "array", tags["type"])
	assert.Equal(t, "string", tags["items"].(map[string]any)["type"])
}
