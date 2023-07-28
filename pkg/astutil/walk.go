package astutil

import (
	"errors"
	"fmt"
	"go/ast"
)

var ErrAmbiguousOuterDecl = errors.New("ambiguous outer declaration name")

func Walk(fileAst *ast.File, fn func([]ast.Node, ast.Node) error) error {
	path := []ast.Node{}
	var err error
	ast.Inspect(fileAst, func(node ast.Node) bool {
		if err != nil {
			return false
		}
		if node == nil {
			path = path[:len(path)-1]
			return true
		}
		defer func() { path = append(path, node) }()
		err = fn(path, node)
		return true
	})
	return err
}

func OuterDeclName(path []ast.Node) (string, error) {
	for _, node := range path {
		switch node := node.(type) {
		case *ast.FuncDecl:
			return FunctionName(node)
		case *ast.TypeSpec:
			return node.Name.Name, nil
		case *ast.ValueSpec:
			if len(node.Names) != 1 {
				return "", fmt.Errorf("ValueSpec contains more than 1 name: %w", ErrAmbiguousOuterDecl)
			}
			return node.Names[0].Name, nil
		}
	}
	return "", fmt.Errorf("no decl")
}
