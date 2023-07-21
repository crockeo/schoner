package astutil

import (
	"fmt"
	"go/ast"
)

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
		decl, ok := node.(ast.Decl)
		if !ok {
			continue
		}
		return RenderDecl(decl)
	}
	return "", fmt.Errorf("no containing declaration")
}
