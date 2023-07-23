package visualize

import (
	"github.com/crockeo/schoner/pkg/graph"
	"github.com/crockeo/schoner/pkg/phases/references"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func Visualize(outputPath string, graph graph.Graph[references.Declaration]) error {
	gviz := graphviz.New()
	g, err := gviz.Graph(graphviz.Directed)
	if err != nil {
		return err
	}

	nodes := map[references.Declaration]*cgraph.Node{}
	for decl := range graph {
		node, err := g.CreateNode(decl.Name)
		if err != nil {
			return err
		}
		nodes[decl] = node
	}

	for decl := range graph {
		for peer := range graph[decl] {
			_, err := g.CreateEdge("", nodes[decl], nodes[peer])
			if err != nil {
				return err
			}
		}
	}

	return gviz.RenderFilename(g, graphviz.SVG, outputPath)
}
