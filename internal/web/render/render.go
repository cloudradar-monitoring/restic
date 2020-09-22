package render

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/restic/restic/internal/web/template/gohtml"
	"html/template"
	"io"
	"net/url"
	"strings"
)

func RenderWithLayout(name string, w io.Writer, data interface{}) error {
	cssTemplateBody, ok := gohtml.TemplatesMap["css"]
	if !ok {
		return errors.New("unknown/unregistered template 'css'")
	}

	baseTemplateBody, ok := gohtml.TemplatesMap["base"]
	if !ok {
		return errors.New("unknown/unregistered template 'base'")
	}

	templateBody, ok := gohtml.TemplatesMap[name]
	if !ok {
		return fmt.Errorf("unknown/unregistered template '%s'", name)
	}

	templ := template.New(name)
	addFuncs(templ)

	pageTemplate, err := templ.Parse(templateBody)
	if err != nil {
		return err
	}

	pageTemplate, err = pageTemplate.Parse(baseTemplateBody)
	if err != nil {
		return err
	}

	pageTemplate, err = pageTemplate.Parse(cssTemplateBody)
	if err != nil {
		return err
	}

	err = pageTemplate.Execute(w, data)
	return err
}

func Render(name string, w io.Writer, data interface{}) error {
	templateBody, ok := gohtml.TemplatesMap[name]
	if !ok {
		return fmt.Errorf("unknown/unregistered template '%s'", name)
	}

	templ := template.New(name)
	addFuncs(templ)

	pageTemplate, err := templ.Parse(templateBody)
	if err != nil {
		return err
	}

	err = pageTemplate.Execute(w, data)
	return err
}

func addFuncs(t *template.Template) {
	t.Funcs(template.FuncMap{
		"joinStrings":      strings.Join,
		"unescapeHtml":     UnescapeHtml,
		"encodeToBase64":   EncodeToBase64,
		"decodeFromBase64": DecodeFromBase64,
		"escapeUrl":        url.QueryEscape,
	})
}

func UnescapeHtml(s string) template.HTML {
	return template.HTML(s)
}

func EncodeToBase64(s string) string {
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)

	_, err := encoder.Write([]byte(s))
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return buf.String()
}

func DecodeFromBase64(s string) string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return string(data)
}
