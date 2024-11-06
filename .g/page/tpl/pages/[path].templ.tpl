package pages

type {{ .funcName }}Data = []string


func {{ .funcName }}Handler(c echo.Context) ({{ .funcName }}Data, error) {
    return {{ .funcName }}Data{}, nil
}

templ {{ .funcName }}(data {{ .funcName }}Data) {
    @ui.Layout() {
    }
}
