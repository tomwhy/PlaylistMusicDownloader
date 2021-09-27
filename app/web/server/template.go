package server

import (
	"html/template"
	"io"
	"path"

	"github.com/labstack/echo"
)

type TemplateRenderer struct {
	template *template.Template
}

func NewTemplateRenderer(rootPath string) *TemplateRenderer {
	renderer := new(TemplateRenderer)

	renderer.template = template.Must(template.ParseGlob(path.Join(rootPath, "*.html")))
	return renderer
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.template.ExecuteTemplate(w, name, data)
}
