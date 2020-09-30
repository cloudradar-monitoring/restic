package render

import (
	"bytes"
	"html/template"
	"net/url"
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

func (ns Nodes) Find(name string) (node *Node, index int) {
	for i, n := range ns {
		if n.Name == name {
			return &n, i
		}
	}

	return nil, 0
}

func (n *Nodes) Add(pathItem PathItem) {
	pathParts := strings.Split(pathItem.Path, "/")

	curNodes := n
	pathPartsTillCurrent := make([]string, 0, len(pathParts))
	for i, pathPart := range pathParts {
		if pathPart == "" {
			continue
		}
		pathPartsTillCurrent = append(pathPartsTillCurrent, pathPart)

		if i == len(pathParts)-1 {
			*curNodes = append(*curNodes, Node{
				PathChain:  "/" + strings.Join(pathPartsTillCurrent, "/"),
				PathItem:   pathItem,
				Nodes:      &Nodes{},
				IsExpanded: false,
			})
			continue
		}

		curNode, index := curNodes.Find(pathPart)
		if curNode == nil {
			curNode = &Node{
				PathChain: "/" + strings.Join(pathPartsTillCurrent, "/"),
				PathItem: PathItem{
					Name: pathPart,
					Path: pathItem.Path,
					Type: "dir",
				},
				Nodes:      &Nodes{},
				IsExpanded: true,
			}
			*curNodes = append(*curNodes, *curNode)
		} else {
			curNode.IsExpanded = true
			curNodesIn := *curNodes
			curNodesIn[index] = *curNode
		}

		curNodes = curNode.Nodes
	}
}

func (n Nodes) Render(data NodesContext) template.HTML {
	nodesStrs := []string{}
	for _, node := range n {
		nodesStrs = append(nodesStrs, node.Render(data))
	}

	return template.HTML(strings.Join(nodesStrs, ""))
}

type Node struct {
	PathItem
	Nodes      *Nodes
	IsExpanded bool
	PathChain  string
}

type NodesContext struct {
	Params     map[string]string
	SnapshotID string
	Curpath    string
	Url        *url.URL
}

func (n Node) IsLeaf() bool {
	return n.PathChain == n.Path
}

func (n Node) Render(data NodesContext) string {
	u := data.Url

	q := u.Query()
	q.Set("dir", n.PathChain)
	u.RawQuery = q.Encode()

	data.Url = u

	var buf bytes.Buffer
	err := Render("dirtree", &buf, struct {
		Node         Node
		NodesContext NodesContext
	}{Node: n, NodesContext: data})
	if err != nil {
		return err.Error()
	}

	return buf.String()
}

func (n Node) IsDir() bool {
	return n.Type == "dir"
}
