package routes

func (r *Routes) {{ .funcName }}(c echo.Context) error {
    return views.{{ .funcName }}().Render(c.Request().Context(), c.Response().Writer)
}
