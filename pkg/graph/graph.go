package graph

import "github.com/crockeo/schoner/pkg/set"

type Graph[T comparable] map[T]set.Set[T]

func NewGraph[T comparable]() Graph[T] {
	return map[T]set.Set[T]{}
}

func (g Graph[T]) ContainsNode(node T) bool {
	_, ok := g[node]
	return ok
}

func (g Graph[T]) AddNode(node T) bool {
	if g.ContainsNode(node) {
		return false
	}
	g[node] = set.NewSet[T]()
	return true
}

func (g Graph[T]) ContainsEdge(from T, to T) bool {
	set, ok := g[from]
	return ok && set.Contains(to)
}

func (g Graph[T]) AddEdge(from T, to T) bool {
	if g.ContainsEdge(from, to) {
		return false
	}
	g.AddNode(from)
	g.AddNode(to)
	g[from].Add(to)
	return true
}

func (g Graph[T]) DFS(roots []T, visitor func(T) error) error {
	visited := set.NewSet[T]()
	for _, root := range roots {
		if visited.Contains(root) {
			continue
		}

		stack := []T{root}
		for len(stack) > 0 {
			next := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if err := visitor(next); err != nil {
				return err
			}
			for child := range g[next] {
				if visited.Contains(child) {
					continue
				}
				stack = append(stack, child)
				visited.Add(child)
			}
		}
	}
	return nil
}
