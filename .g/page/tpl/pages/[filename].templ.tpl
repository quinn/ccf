package pages

type {{ .templName }}Data = []string


func {{ .templName }}Handler(c echo.Context{{ .templParams }}) ({{ .templName }}Data, error) {
    return []string{}, nil
}

templ {{ .templName }}(data {{ .templName }}Data) {
    @ui.Layout() {
    }
}
