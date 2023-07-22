package visualize

import (
	"fmt"

	"github.com/crockeo/schoner/pkg/graph"
	"github.com/crockeo/schoner/pkg/phases/declarations"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func Visualize(outputPath string, graph graph.Graph[declarations.Declaration]) error {
	gviz := graphviz.New()
	rootGraph, err := gviz.Graph(graphviz.Directed)
	if err != nil {
		return err
	}

	subgraphs := map[string]*cgraph.Graph{}
	nodes := map[string]*cgraph.Node{}
	for decl, targets := range graph {
		if err := registerDecl(rootGraph, subgraphs, nodes, decl); err != nil {
			return err
		}
		for target := range targets {
			if err := registerDecl(rootGraph, subgraphs, nodes, target); err != nil {
				return err
			}
		}
	}

	for decl, targets := range graph {
		subgraph := subgraphs[decl.Filename]
		from := nodes[decl.QualifiedName()]
		for target := range targets {
			to := nodes[target.QualifiedName()]
			_, err := subgraph.CreateEdge(
				fmt.Sprintf("%s-%s", decl.QualifiedName(), target.QualifiedName()),
				from,
				to,
			)
			if err != nil {
				return err
			}
		}
	}

	return gviz.RenderFilename(rootGraph, graphviz.SVG, outputPath)
}

func registerDecl(
	rootGraph *cgraph.Graph,
	subgraphs map[string]*cgraph.Graph,
	nodes map[string]*cgraph.Node,
	decl declarations.Declaration,
) error {
	if _, ok := subgraphs[decl.Filename]; !ok {
		subgraph := rootGraph.SubGraph(decl.Filename, 1)
		if subgraph == nil {
			return fmt.Errorf("unexpected nil sub graph for %s", decl.Filename)
		}

		subgraph = subgraph.SetLabel(decl.Filename)
		subgraph = subgraph.SetBackgroundColor("lightgrey")
		subgraphs[decl.Filename] = subgraph
	}

	nodeName := decl.Name
	if nodeName == "" {
		nodeName = decl.Filename
	}
	if _, ok := nodes[decl.QualifiedName()]; !ok {
		subgraph := subgraphs[decl.Filename]
		node, err := subgraph.CreateNode(decl.QualifiedName())
		if err != nil {
			return err
		}
		if node == nil {
			return fmt.Errorf("unexpected nil node for %s", decl.QualifiedName())
		}
		node.SetLabel(nodeName)
		node.SetGroup(decl.Filename)
		nodes[decl.QualifiedName()] = node
	}

	return nil
}
