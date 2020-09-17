package render

import (
	"errors"
	"fmt"
	"github.com/restic/restic/internal/web/template/gohtml"
	"html/template"
	"io"
	"strings"
)

func Render(name string, w io.Writer, data interface{}) error {
	baseTemplateBody, ok := gohtml.TemplatesMap["base"]
	if !ok {
		return errors.New("unknown/unregistered template 'base'")
	}

	templateBody, ok := gohtml.TemplatesMap[name]
	if !ok {
		return fmt.Errorf("unknown/unregistered template '%s'", name)
	}

	pageTemplate, err := template.New(name).Funcs(template.FuncMap{
		"joinStrings":  strings.Join,
		"unescapeHtml": unescapeHtml,
	}).Parse(templateBody)
	if err != nil {
		return err
	}

	pageTemplate, err = pageTemplate.Parse(baseTemplateBody)
	if err != nil {
		return err
	}

	err = pageTemplate.Execute(w, data)
	return err
}

func unescapeHtml(s string) template.HTML {
	return template.HTML(s)
}
