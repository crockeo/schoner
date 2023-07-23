package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crockeo/schoner/pkg/phases/fileinfo"
	"github.com/crockeo/schoner/pkg/phases/references"
	"github.com/crockeo/schoner/pkg/set"
	"github.com/crockeo/schoner/pkg/visualize"
	"github.com/crockeo/schoner/pkg/walk"
)

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainImpl() error {
	walkOptions := walk.WithIgnoreDirs(".git")
	for _, root := range os.Args[1:] {
		root, err := filepath.Abs(root)
		if err != nil {
			return err
		}

		fileInfos, err := fileinfo.FindFileInfos(root, walkOptions)
		if err != nil {
			return err
		}

		referenceGraph, err := references.BuildReferenceGraph(root, fileInfos, walkOptions)
		if err != nil {
			return err
		}

		unreachable := set.NewSet[references.Declaration]()
		entrypoints := set.NewSet[references.Declaration]()
		for decl := range referenceGraph {
			unreachable.Add(decl)
			if decl.Parent.Entrypoints.Contains(decl.Name) {
				entrypoints.Add(decl)
			}
		}
		_ = referenceGraph.DFS(entrypoints.ToSlice(), func(node references.Declaration) error {
			unreachable.Remove(node)
			return nil
		})

		if err := visualize.Visualize(
			fmt.Sprintf("%s.svg", filepath.Base(root)),
			referenceGraph,
			entrypoints,
			unreachable,
		); err != nil {
			return err
		}
	}

	return nil
}
