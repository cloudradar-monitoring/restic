package render

import (
	"bytes"
	"html/template"
	"os"
	"strings"
	"time"
)

type PathItem struct {
	Name       string
	Type       string
	Path       string
	UID        uint32
	GID        uint32
	Size       uint64
	Mode       os.FileMode
	ModTime    time.Time
	AccessTime time.Time
	ChangeTime time.Time
	StructType string
}

type Node struct {
	PathItem
	Nodes Nodes
}

func (n Node) Render(data interface{}) string {
	var buf bytes.Buffer

	err := Render("dirtree", &buf, struct {
		Node Node
		Data interface{}
	}{Node: n, Data: data})
	if err != nil {
		return err.Error()
	}

	return buf.String()
}

func (n Node) IsDir() bool {
	return n.Type == "dir"
}

type Nodes map[string]Node

func (n Nodes) Add(pathItem PathItem) {
	pathParts := strings.Split(pathItem.Path, string(os.PathSeparator))

	curNodes := n
	for i, pathPart := range pathParts {
		if pathPart == "" {
			continue
		}

		if i == len(pathParts)-1 {
			n.addPathPart(curNodes, pathPart, pathItem)
			continue
		}
		emptyPathItem := PathItem{
			Path: pathItem.Path,
			Type: "dir",
		}
		n.addPathPart(curNodes, pathPart, emptyPathItem)
		curNodes = curNodes[pathPart].Nodes
	}
}

func (n Nodes) Render(data interface{}) template.HTML {
	nodesStrs := []string{}
	for _, node := range n {
		nodesStrs = append(nodesStrs, node.Render(data))
	}

	return template.HTML(strings.Join(nodesStrs, ""))
}

func (n Nodes) addPathPart(nodes Nodes, pathPart string, pathItem PathItem) {
	if _, ok := nodes[pathPart]; !ok {
		nodes[pathPart] = Node{
			PathItem: pathItem,
			Nodes:    Nodes{},
		}
	}
}
