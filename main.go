package main

import (
	"fmt"
	"os"

	"github.com/crockeo/schoner/pkg/phases"
)

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainImpl() error {
	for _, root := range os.Args[1:] {
		declarations, err := phases.FindDeclarations(root, phases.WithIgnoreDirs(".git"))
		if err != nil {
			return err
		}

		references, err := phases.FindReferences(root, declarations, phases.WithIgnoreDirs(".git"))
		if err != nil {
			return err
		}

		_ = references

		for filename, fileDecls := range declarations {
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

		for filename, fileReferences := range references {
			fmt.Println(filename)
			for from, tos := range fileReferences.References {
				fmt.Println("  ", from, "->")
				for to := range tos {
					fmt.Println("    ", to)
				}
			}
		}
	}

	return nil
}
