package astutil

import (
	"errors"
	"fmt"
	"go/ast"
)

var ErrAmbiguousOuterDecl = errors.New("ambiguous outer declaration name")

func Walk(node ast.Node, fn func(ast.Node) error) error {
	var err error
	ast.Inspect(node, func(node ast.Node) bool {
		if err != nil {
			return false
		}
		err = fn(node)
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
