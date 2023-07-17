package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

var skipDirs = map[string]struct{}{
	".git": {},
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainImpl() error {
	for _, root := range os.Args[1:] {
		if err := processRoot(root); err != nil {
			return fmt.Errorf("failed to process root %s: %w", root, err)
		}
	}

	return nil
}

func processRoot(root string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !filepath.IsAbs(root) {
		root = filepath.Join(wd, root)
	}

	isGoProject, err := checkIsGoProject(root)
	if err != nil {
		return err

	}
	if !isGoProject {
		fmt.Fprintf(os.Stderr, "Skipping %s because it doesn't have a go.mod. Is it a Go project?\n", root)
		return nil
	}

	fileset := token.NewFileSet()
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if _, ok := skipDirs[filepath.Base(path)]; ok {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return processFile(fileset, path, contents)
	})
	return nil
}

func processFile(fileset *token.FileSet, filename string, contents []byte) error {
	fileAst, err := parser.ParseFile(fileset, filename, contents, parser.ParseComments)
	if err != nil {
		return err
	}

	names, err := collectNames(fileAst)
	if err != nil {
		return err
	}

	fmt.Println(filename)
	for _, name := range names {
		fmt.Println("  ", name)
	}
	fmt.Println()
	return nil
}

func collectNames(fileAst *ast.File) ([]string, error) {
	names := []string{}

	var walkErr error
	ast.Inspect(fileAst, func(node ast.Node) bool {
		if walkErr != nil {
			return false
		}

		switch node := node.(type) {
		case *ast.GenDecl:
			// names = append(names, node.Specs[0].(*ast.ValueSpec).Names[0].Name)
			// fmt.Println(node)
		case *ast.FuncDecl:
			funcName, err := getFunctionName(node)
			if err != nil {
				walkErr = err
				return false
			}
			names = append(names, funcName)
		}
		return true
	})

	return names, walkErr
}

func getDeclName(decl ast.GenDecl) (string, error) {
	panic("not implemented")
}

func getFunctionName(fn *ast.FuncDecl) (string, error) {
	funcName := fn.Name.Name
	if fn.Recv != nil {
		recvName, ok := getReceiverName(fn.Recv)
		if !ok {
			return "", fmt.Errorf("failed to get receiver name for %s", fn.Name.Name)
		}
		funcName = fmt.Sprintf("%s::%s", recvName, funcName)
	}
	return funcName, nil
}

func getReceiverName(recv *ast.FieldList) (string, bool) {
	switch typename := recv.List[0].Type.(type) {
	case *ast.Ident:
		return typename.Name, true
	case *ast.StarExpr:
		return typename.X.(*ast.Ident).Name, true
	default:
		return "", false
	}
}

func checkIsGoProject(root string) (bool, error) {
	_, err := os.Stat(root)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
