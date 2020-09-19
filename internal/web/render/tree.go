package render

import "strings"

type Node struct {
	Name   string
	IsLeaf bool
	Nodes  Nodes
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

func (n Nodes) addPathPart(nodes Nodes, pathPart, name string, isLeaf bool) {
	if _, ok := nodes[pathPart]; !ok {
		nodes[pathPart] = Node{
			Name:   name,
			IsLeaf: isLeaf,
			Nodes:  Nodes{},
		}
	}
}
