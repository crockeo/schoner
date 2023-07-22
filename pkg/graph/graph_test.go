package graph

import (
	"testing"

	"github.com/crockeo/schoner/pkg/set"
	"github.com/stretchr/testify/assert"
)

func TestGraph_ContainsNode(t *testing.T) {
	graph := NewGraph[string]()
	graph["a"] = set.NewSet[string]()
	assert.True(t, graph.ContainsNode("a"))
	assert.False(t, graph.ContainsNode("b"))
}

func TestGraph_AddNode(t *testing.T) {
	graph := NewGraph[string]()
	assert.True(t, graph.AddNode("a"))
	assert.False(t, graph.AddNode("a"))
	assert.True(t, graph.ContainsNode("a"))
}

func TestGraph_ContainsEdge(t *testing.T) {
	graph := NewGraph[string]()
	graph["a"] = set.NewSet[string]()
	graph["b"] = set.NewSet[string]()
	graph["a"].Add("b")
	assert.True(t, graph.ContainsEdge("a", "b"))
	assert.False(t, graph.ContainsEdge("b", "a"))
}

func TestGraph_AddEdge(t *testing.T) {
	graph := NewGraph[string]()
	assert.True(t, graph.AddEdge("a", "b"))
	assert.False(t, graph.AddEdge("a", "b"))
	assert.True(t, graph.ContainsEdge("a", "b"))
}

func TestGraph_DFS(t *testing.T) {
	graph := NewGraph[string]()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("a", "d")
	graph.AddEdge("e", "f")

	visited := set.NewSet[string]()
	graph.DFS([]string{"a"}, func(node string) error {
		visited.Add(node)
		return nil
	})

	assert.Equal(
		t,
		set.Set[string](map[string]struct{}{
			"a": {},
			"b": {},
			"c": {},
			"d": {},
		}),
		visited,
	)
}
