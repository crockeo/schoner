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
		}
	}

	return nil
}
