package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestIndexHandler(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test
	err := IndexGET(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// The actual response body will be HTML from the templ component
	assert.Contains(t, rec.Body.String(), "<html")
}

func TestBlogSlugHandler(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/blog/test-post", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/blog/:slug")
	c.SetParamNames("slug")
	c.SetParamValues("test-post")

	// Test
	err := BlogSlugGET(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// The actual response body will be HTML from the templ component
	assert.Contains(t, rec.Body.String(), "<html")
}

func TestRegisterRoutes(t *testing.T) {
	// Setup
	e := echo.New()

	// Register routes
	RegisterRoutes(e)

	// Get registered routes
	routes := e.Routes()

	// Expected routes
	expectedRoutes := map[string]string{
		"/":           "GET",
		"/blog/:slug": "GET",
	}

	// Verify all expected routes are registered
	foundRoutes := make(map[string]bool)
	for _, r := range routes {
		if method, exists := expectedRoutes[r.Path]; exists {
			assert.Equal(t, method, r.Method, "Unexpected method for route %s", r.Path)
			foundRoutes[r.Path] = true
		}
	}

	// Verify all expected routes were found
	for path := range expectedRoutes {
		assert.True(t, foundRoutes[path], "Expected route %s was not registered", path)
	}
}
