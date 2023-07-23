package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crockeo/schoner/pkg/graph"
	"github.com/crockeo/schoner/pkg/phases"
	"github.com/crockeo/schoner/pkg/phases/declarations"
	"github.com/crockeo/schoner/pkg/phases/references"
	"github.com/crockeo/schoner/pkg/walk"
	"github.com/crockeo/schoner/pkg/visualize"
)

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainImpl() error {
	for _, root := range os.Args[1:] {
		root, err := filepath.Abs(root)
		if err != nil {
			return err
		}

		decls, err := phases.FindDeclarations(root, walk.WithIgnoreDirs(".git"))
		if err != nil {
			return err
		}

		refs, err := phases.FindReferences(root, decls, walk.WithIgnoreDirs(".git"))
		if err != nil {
			return err
		}

		// renderDecls(decls)
		// renderRefs(refs)

		referenceGraph := graph.NewGraph[declarations.Declaration]()
		for _, fileRefs := range refs {
			for decl, targets := range fileRefs.References {
				for target := range targets {
					referenceGraph.AddEdge(decl, target)
				}
			}
		}

		if err := visualize.Visualize(
			fmt.Sprintf("%s.svg", filepath.Base(root)),
			referenceGraph,
		); err != nil {
			return err
		}
	}

	return nil
}

func renderDecls(decls map[string]*declarations.Declarations) {
	for filename, fileDecls := range decls {
		fmt.Println(filename)
		for decl := range fileDecls.Declarations {
			fmt.Println("  ", decl)
		}
		fmt.Println("   ---")
		for importDecl := range fileDecls.Imports {
			fmt.Println("  ", fmt.Sprintf("`%s`", importDecl.Name), importDecl.Path)
		}
		fmt.Println()
		fmt.Println()
	}
}

func renderRefs(refs map[string]*references.References) {
	for filename, fileReferences := range refs {
		fmt.Println(filename)
		for from, tos := range fileReferences.References {
			fmt.Println("  ", from, "->")
			for to := range tos {
				fmt.Println("    ", to)
			}
		}
	}
}
