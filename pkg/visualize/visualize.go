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
	rootGraph, err := gviz.Graph(graphviz.Directed)
	if err != nil {
		return err
	}

	nodes := map[string]*cgraph.Node{}
	for decl, targets := range graph {
		if err := registerDecl(rootGraph, nodes, decl); err != nil {
			return err
		}
		for target := range targets {
			if err := registerDecl(rootGraph, nodes, target); err != nil {
				return err
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

			_, err := rootGraph.CreateEdge("a", from, to)
			if err != nil {
				return err
			}
		}
	}

	return gviz.RenderFilename(rootGraph, graphviz.SVG, outputPath)
}

func registerDecl(
	rootGraph *cgraph.Graph,
	nodes map[string]*cgraph.Node,
	decl declarations.Declaration,
) error {
	nodeName := decl.Name
	if nodeName == "" {
		nodeName = "root"
	}
	if _, ok := nodes[nodeName]; !ok {
		node, err := rootGraph.CreateNode(nodeName)
		if err != nil {
			return err
		}
		if node == nil {
			return fmt.Errorf("unexpected nil node for %s", decl.QualifiedName())
		}
		node.SetGroup(decl.Filename)
		nodes[decl.QualifiedName()] = node
	}

	return nil
}
