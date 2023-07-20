package main

import (
	"fmt"
	"os"

	"github.com/crockeo/schoner/pkg/phases/decls"
)

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainImpl() error {
	for _, root := range os.Args[1:] {
		projectDecls, err := decls.FindProjectDeclarations(root, decls.WithIgnoreDirs(".git"))
		if err != nil {
			return err
		}

		for filename, fileDecls := range projectDecls {
			fmt.Println(filename)
			for _, decl := range fileDecls.Symbols {
				fmt.Println("  ", decl)
			}
			fmt.Println("   ---")
			for _, importDecl := range fileDecls.Imports {
				fmt.Println("  ", fmt.Sprintf("`%s`", importDecl.Name), importDecl.Path)
			}
			fmt.Println()
			fmt.Println()
		}
	}

	return nil
}
