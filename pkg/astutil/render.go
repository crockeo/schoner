package astutil

import (
	"fmt"
	"go/ast"
	"strings"
)

func Qualify(names ...string) string {
	return strings.Join(names, "::")
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

func RenderDecl(decl ast.Decl) (string, error) {
	switch decl := decl.(type) {
	case *ast.GenDecl:
		return renderGenDecl(decl)
	case *ast.FuncDecl:
		return renderFuncDecl(decl)
	default:
		return "", fmt.Errorf("unknown declaration type: %T", decl)
	}
}

func renderGenDecl(decl *ast.GenDecl) (string, error) {
	// TODO: handle more than one spec
	switch decl := decl.Specs[0].(type) {
	case *ast.ImportSpec:
		panic("not implemented")
	case *ast.TypeSpec:
		return decl.Name.Name, nil
	case *ast.ValueSpec:
		// TODO: do we ever need to handle multiple elements? probably yes...
		return decl.Names[0].Name, nil
	default:
		return "", fmt.Errorf("unknown GenDecl type: %T", decl)
	}
}

func renderFuncDecl(decl *ast.FuncDecl) (string, error) {
	funcName := decl.Name.Name
	if decl.Recv != nil {
		recvName, ok := ExprName(decl.Recv.List[0].Type)
		if !ok {
			return "", fmt.Errorf("failed to get receiver name for %s", decl.Name.Name)
		}
		funcName = Qualify(recvName, funcName)
	}
	return funcName, nil
}
