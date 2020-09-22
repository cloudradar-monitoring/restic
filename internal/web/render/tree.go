package render

import (
	"bytes"
	"html/template"
	"strings"
)

type Node struct {
	Name   string
	IsLeaf bool
	Nodes  Nodes
}

func (n Node) Render() string {
	var buf bytes.Buffer

	err := Render("dirtree", &buf, struct{ Node Node }{Node: n})
	if err != nil {
		return err.Error()
	}

	return buf.String()
}

func (n Node) IsDir() bool {
	return !n.IsLeaf
}

type Nodes map[string]Node

func (n Nodes) Add(pathSep, path, name string, isLeaf bool) {
	pathParts := strings.Split(path, pathSep)

	curNodes := n
	for i, pathPart := range pathParts {
		if pathPart == "" {
			continue
		}

		if i == len(pathParts)-1 {
			n.addPathPart(curNodes, pathPart, name, isLeaf)
			continue
		}
		n.addPathPart(curNodes, pathPart, "", false)
		curNodes = curNodes[pathPart].Nodes
	}
}

func (n Nodes) Render() template.HTML {
	nodesStrs := []string{}
	for _, node := range n {
		nodesStrs = append(nodesStrs, node.Render())
	}

	return template.HTML(strings.Join(nodesStrs, ""))
}

func (n Nodes) addPathPart(nodes Nodes, pathPart, name string, isLeaf bool) {
	if _, ok := nodes[pathPart]; !ok {
		nodes[pathPart] = Node{
			Name:   name,
			IsLeaf: isLeaf,
			Nodes:  Nodes{},
		}
	}
}
