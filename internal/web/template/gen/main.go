package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// This program generates go files from html templates. It can be invoked by running `go generate internal/web/template/gen/main.go`
//go:generate go run main.go
//go:generate go fmt ../gohtml
func main() {
	templateNames := make([]string, 0)

	err := filepath.Walk("../html", func(path string, info os.FileInfo, err error) error {
		templateName, err := generateTemplateFile(path, info)
		die(err)

		if templateName != "" {
			templateNames = append(templateNames, templateName)
		}
		return nil
	})
	die(err)

	err = generateTemplatesMap(templateNames)
	die(err)
}

func generateTemplatesMap(templateNames []string) error {
	t, err := template.ParseFiles("map.gohtml")
	if err != nil {
		return err
	}

	f, err := os.Create("../gohtml/templates.go")
	if err != nil {
		return err
	}

	defer closeSecure(f)

	err = t.Execute(f, struct {
		TemplateNames []string
	}{TemplateNames: templateNames})

	if err != nil {
		return err
	}

	return nil
}

func generateTemplateFile(path string, info os.FileInfo) (string, error) {
	ex := filepath.Ext(path)
	if ex != ".gohtml" {
		return "", nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	fileNameWithoutExt := strings.TrimSuffix(info.Name(), ex)

	f, err := os.Create("../gohtml/" + fileNameWithoutExt + ".template.go")
	if err != nil {
		return "", err
	}
	defer closeSecure(f)

	t, err := template.ParseFiles("template.gohtml")
	if err != nil {
		return "", err
	}

	err = t.Execute(f, struct {
		Content string
		Name    string
	}{Content: string(content), Name: fileNameWithoutExt})
	if err != nil {
		return "", err
	}

	return fileNameWithoutExt, nil
}

func closeSecure(cl io.Closer) {
	err := cl.Close()
	if err != nil {
		log.Println(err)
	}
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
