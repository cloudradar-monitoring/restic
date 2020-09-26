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

type Nodes []Node

func (ns Nodes) Find(name string) (node Node, found bool) {
	for _, n := range ns {
		if n.Name == name {
			return n, true
		}
	}

	return
}

func (n *Nodes) Add(pathItem PathItem) {
	pathParts := strings.Split(pathItem.Path, string(os.PathSeparator))

	curNodes := n
	for i, pathPart := range pathParts {
		if pathPart == "" {
			continue
		}

		if i == len(pathParts)-1 {
			*curNodes = append(*curNodes, Node{
				PathItem: pathItem,
				Nodes:    &Nodes{},
			})
			continue
		}

		curNode, found := curNodes.Find(pathPart)
		if !found {
			curNode = Node{
				PathItem: PathItem{
					Path: pathItem.Path,
					Type: "dir",
				},
			}
			*curNodes = append(*curNodes, curNode)
		}

		curNodes = curNode.Nodes
	}
}

func (n Nodes) Render(data interface{}) template.HTML {
	nodesStrs := []string{}
	for _, node := range n {
		nodesStrs = append(nodesStrs, node.Render(data))
	}

	return template.HTML(strings.Join(nodesStrs, ""))
}

type Node struct {
	PathItem
	Nodes *Nodes
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
