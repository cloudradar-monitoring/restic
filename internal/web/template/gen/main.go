package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type externalLinkType int

const (
	css externalLinkType = iota + 1
	js
)

type externalLink struct {
	link string
	hash string
	name string
	typ  externalLinkType
}

var externalLinks = []externalLink{
	{
		link: "https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css",
		hash: "sha384-25c29bf2ade2a89eb580d57d2866fcb614ac363a522f49fc3c0467f47b993a72313748683fe56698318c379b7d509d19",
		name: "bootstrap.css",
		typ:  css,
	},
	{
		link: "https://code.jquery.com/jquery-3.5.1.slim.min.js",
		hash: "sha384-0df5ddcf686d3c7d25b124ace67093a6e8ffcf2e02f8e1a96a6a05572dfc31506713e21b6d56147b0f8eac25da4647e3",
		name: "jqueryjs",
		typ:  js,
	},
	{
		link: "https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js",
		hash: "sha384-07882dd63ac60bb261e008133d2754b4e06f7cef2c86e7f9ec16a086a15f3e5631a1fbcbf3f29411f62224e30c57ead5",
		name: "bootstrapjs",
		typ:  js,
	},
}

// This program generates go files from html templates. It can be invoked by running `go generate internal/web/template/gen/main.go`
//go:generate go run main.go
//go:generate go fmt ../gohtml
func main() {
	templateNames := make(map[string]string, 0)

	err := filepath.Walk("../source", func(path string, info os.FileInfo, err error) error {
		ex := filepath.Ext(path)
		if ex != ".gohtml" {
			return nil
		}
		templateNameRaw := strings.TrimSuffix(info.Name(), ex)
		templateNameSanitized := sanitiseTemplateNameVariable(templateNameRaw)

		fileReader, err := os.Open(path)
		die(err)

		err = generateTemplateFile(fileReader, templateNameSanitized)
		die(err)

		templateNames[templateNameRaw] = templateNameSanitized

		return nil
	})
	die(err)

	for _, el := range externalLinks {
		body, err := readLink(el)
		die(err)

		templateNameSanitized := sanitiseTemplateNameVariable(el.name)

		var templateReader io.Reader
		if el.typ == css {
			templateFormat := `{{define "%s"}}
<style type="text/css">%s</style>
{{end}}
`
			templateReader = bytes.NewBufferString(fmt.Sprintf(templateFormat, el.name, string(body)))
		} else if el.typ == js {
			templateFormat := `{{define "%s"}}
<script type="application/javascript">%s</script>
{{end}}
`
			templateReader = bytes.NewBufferString(fmt.Sprintf(templateFormat, el.name, string(body)))
		} else {
			templateReader = bytes.NewBuffer(body)
		}

		err = generateTemplateFile(templateReader, templateNameSanitized)
		die(err)

		templateNames[el.name] = templateNameSanitized
	}

	err = generateTemplatesMap(templateNames)
	die(err)
}

func sanitiseTemplateNameVariable(templateName string) string {
	return regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(templateName, "_")
}

func readLink(el externalLink) (data []byte, err error) {
	resp, err := http.Get(el.link)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	if el.hash != "" {
		buf := bytes.NewBuffer(body)
		err = CheckHash(el.hash, el.link, buf)
		if err != nil {
			return data, err
		}
	}

	return body, nil
}

func generateTemplatesMap(templateNames map[string]string) error {
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
		TemplateNames map[string]string
	}{TemplateNames: templateNames})

	if err != nil {
		return err
	}

	return nil
}

func generateTemplateFile(reader io.Reader, name string) error {
	f, err := os.Create("../gohtml/" + name + ".template.go")
	if err != nil {
		return err
	}
	defer closeSecure(f)

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	t, err := template.ParseFiles("template.gohtml")
	if err != nil {
		return err
	}

	err = t.Execute(f, struct {
		Content string
		Name    string
	}{Content: string(content), Name: name})
	if err != nil {
		return err
	}

	return nil
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

func CheckHash(hashStr, resourceName string, content io.Reader) (err error) {
	expectedHashAlgoName, expectedHashSum, err := ParseHashAlgoAndSum(hashStr)
	if err != nil {
		return err
	}

	hashAlgo, err := ExtractHashAlgo(expectedHashAlgoName)
	if err != nil {
		return err
	}

	if _, err := io.Copy(hashAlgo, content); err != nil {
		return err
	}

	actualHashSum := fmt.Sprintf("%x", hashAlgo.Sum(nil))

	if expectedHashSum != actualHashSum {
		return fmt.Errorf(
			"unexpected hash sum '%s' for algo '%s' from '%s', expected hash sum is '%s'",
			actualHashSum,
			expectedHashAlgoName,
			resourceName,
			hashStr,
		)
	}
	return nil
}

func ParseHashAlgoAndSum(hashStr string) (algoName, sum string, err error) {
	const expectedRegexParts = 3
	reg := regexp.MustCompile(`^(\w*)-(.+)$`)
	regParts := reg.FindStringSubmatch(hashStr)

	if len(regParts) != expectedRegexParts {
		return "", "", fmt.Errorf("invalid hash string '%s'", hashStr)
	}

	return regParts[1], regParts[2], nil
}

func ExtractHashAlgo(hashAlgoName string) (hChecker hash.Hash, err error) {
	switch hashAlgoName {
	case "sha512":
		return sha512.New(), nil
	case "sha384":
		return sha512.New384(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha224":
		return sha256.New224(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	default:
		return nil, fmt.Errorf("unknown hash algorithm '%s'", hashAlgoName)
	}
}
