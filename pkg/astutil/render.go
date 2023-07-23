package astutil

import (
	"fmt"
	"go/ast"
	"strings"
)

func Qualify(names ...string) string {
	return strings.Join(names, "::")
}

func Unqualify(name string) []string {
	return strings.Split(name, "::")
}

func IsQualified(name string) bool {
	return strings.Contains(name, "::")
}

func ExprName(expr ast.Expr) (string, bool) {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name, true
	case *ast.StarExpr:
		return ExprName(expr.X)
	case *ast.IndexExpr:
		return ExprName(expr.X)
	default:
		return "", false
	}
}

func FunctionName(fn *ast.FuncDecl) (string, error) {
	funcName := fn.Name.Name
	if fn.Recv != nil {
		recvName, ok := ExprName(fn.Recv.List[0].Type)
		if !ok {
			return "", fmt.Errorf("failed to get receiver name for %s", fn.Name.Name)
		}
		funcName = Qualify(recvName, funcName)
	}
	return funcName, nil
}
