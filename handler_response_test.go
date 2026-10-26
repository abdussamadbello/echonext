package echonext_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdussamadbello/echonext"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// serve runs one request against app and returns the recorder plus everything
// Echo logged while handling it
func serve(app *echonext.App, method, path string) (*httptest.ResponseRecorder, string) {
	var logs bytes.Buffer
	app.Echo.Logger = slog.New(slog.NewTextHandler(&logs, nil))

	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec, logs.String()
}

func TestPlainHandlerWritingItsOwnResponse(t *testing.T) {
	app := echonext.New()
	app.GET("/raw", func(c *echo.Context) error {
		return c.JSON(http.StatusAccepted, map[string]string{"status": "ok"})
	})
	app.DELETE("/items/:id", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	rec, logs := serve(app, http.MethodGet, "/raw")
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
	assert.NotContains(t, logs, "response already written")

	rec, logs = serve(app, http.MethodDelete, "/items/1")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.NotContains(t, logs, "response already written")
}

func TestPlainHandlerWithoutResponse(t *testing.T) {
	app := echonext.New()
	app.POST("/ping", func(c *echo.Context) error {
		return nil
	})

	rec, _ := serve(app, http.MethodPost, "/ping")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestPlainHandlerError(t *testing.T) {
	app := echonext.New()
	app.DELETE("/missing", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "item not found")
	})
	app.DELETE("/broken", func(c *echo.Context) error {
		return errors.New("database unavailable")
	})

	for path, want := range map[string]struct {
		code int
		msg  string
	}{
		"/missing": {http.StatusNotFound, "item not found"},
		"/broken":  {http.StatusInternalServerError, "database unavailable"},
	} {
		rec, _ := serve(app, http.MethodDelete, path)
		assert.Equal(t, want.code, rec.Code, path)

		var response echonext.Response[any]
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response), path)
		assert.False(t, response.Success, path)
		assert.Equal(t, want.msg, response.Error, path)
	}
}

func TestTypedHandlerWritingItsOwnResponse(t *testing.T) {
	app := echonext.New()
	app.GET("/csv", func(c *echo.Context) (TestUser, error) {
		return TestUser{}, c.Blob(http.StatusOK, "text/csv", []byte("id,name\n1,John\n"))
	})

	rec, logs := serve(app, http.MethodGet, "/csv")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "id,name\n1,John\n", rec.Body.String())
	assert.NotContains(t, logs, "response already written")
}
