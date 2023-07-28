package visualize

import (
	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/graph"
	"github.com/crockeo/schoner/pkg/phases/fileinfo"
	"github.com/crockeo/schoner/pkg/set"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func Visualize(
	outputPath string,
	graph graph.Graph[fileinfo.Declaration],
	entrypoints set.Set[fileinfo.Declaration],
	unreachable set.Set[fileinfo.Declaration],
) error {
	gviz := graphviz.New()
	g, err := gviz.Graph(graphviz.Directed)
	if err != nil {
		return err
	}

	nodes := map[fileinfo.Declaration]*cgraph.Node{}
	for decl := range graph {
		node, err := g.CreateNode(astutil.Qualify(decl.Parent.Filename, decl.Name))
		node.SetLabel(decl.Name)
		if entrypoints.Contains(decl) {
			node.SetStyle(cgraph.FilledNodeStyle)
			node.SetFillColor("chartreuse")
		} else if unreachable.Contains(decl) {
			node.SetStyle(cgraph.FilledNodeStyle)
			node.SetFillColor("gray")
		}
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
