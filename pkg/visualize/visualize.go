package visualize

import (
	"fmt"

	"github.com/crockeo/schoner/pkg/phases/declarations"
	"github.com/crockeo/schoner/pkg/set"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func Visualize(outputPath string, graph map[declarations.Declaration]set.Set[declarations.Declaration]) error {
	gviz := graphviz.New()
	g, err := gviz.Graph()
	if err != nil {
		return err
	}

	nodes := map[string]*cgraph.Node{}
	for decl, targets := range graph {
		if _, ok := nodes[decl.QualifiedName()]; !ok {
			node, err := g.CreateNode(decl.QualifiedName())
			if err != nil {
				return err
			}
			if node == nil {
				return fmt.Errorf("unexpected nil node %s", decl.QualifiedName())
			}
			nodes[decl.QualifiedName()] = node
		}

		for target := range targets {
			if _, ok := nodes[target.QualifiedName()]; !ok {
				node, err := g.CreateNode(target.QualifiedName())
				if err != nil {
					return err
				}
				if node == nil {
					return fmt.Errorf("unexpected nil node %s", target.QualifiedName())
				}
				nodes[target.QualifiedName()] = node
			}
		}
	}

	for decl, targets := range graph {
		for target := range targets {
			from := nodes[decl.QualifiedName()]
			to := nodes[target.QualifiedName()]
			if from == nil || to == nil {
				panic(fmt.Sprintf("%s %v %s %v", decl.QualifiedName(), from, target.QualifiedName(), to))
			}

			_, err := g.CreateEdge("a", from, to)
			if err != nil {
				return err
			}
		}
	}

	return gviz.RenderFilename(g, graphviz.SVG, outputPath)
}
