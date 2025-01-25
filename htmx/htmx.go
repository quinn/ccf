package htmx

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Is(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func Redirect(c echo.Context, path string) error {
	if Is(c) {
		c.Response().Header().Set("HX-Redirect", path)
		return c.NoContent(http.StatusOK)
	}

	return c.Redirect(http.StatusFound, path)
}

func Refresh(c echo.Context) error {
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}
