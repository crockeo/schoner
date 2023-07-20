package decls

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Declarations struct {
	Symbols []string
	Imports []ImportDeclaration
}

type ImportDeclaration struct {
	Name string
	Path string
}

func FindFileDeclarations(context struct{}, fileset *token.FileSet, filename string, contents []byte) (*Declarations, error) {
	fileAst, err := parser.ParseFile(fileset, filename, contents, 0)
	if err != nil {
		return nil, err
	}

	symbols := []string{}
	imports := []ImportDeclaration{}
	path := []ast.Node{}
	err = walk(fileAst, func(node ast.Node) error {
		if node == nil {
			path = path[:len(path)-1]
			return nil
		}

		switch node := node.(type) {
		case *ast.GenDecl:
			declNames, importDecls, err := getDeclNames(path, node)
			if err != nil {
				return err
			}
			symbols = append(symbols, declNames...)
			imports = append(imports, importDecls...)
		case *ast.FuncDecl:
			funcName, err := getFunctionName(path, node)
			if err != nil {
				return err
			}
			symbols = append(symbols, funcName)
		}
		path = append(path, node)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Declarations{
		Symbols: symbols,
		Imports: imports,
	}, nil
}

func getDeclNames(path []ast.Node, decl *ast.GenDecl) ([]string, []ImportDeclaration, error) {
	lastNode := path[len(path)-1]
	if _, ok := lastNode.(*ast.File); !ok {
		// We only care about top-level declarations,
		// because only those can be used by other packages.
		return nil, nil, nil
	}

	names := []string{}
	importDecls := []ImportDeclaration{}
	for _, spec := range decl.Specs {
		switch spec := spec.(type) {
		case *ast.ImportSpec:
			// Intentionally ignoring ImportSpec in this phase,
			// because we only care about declarations which are defined
			// inside of this file.
			if spec.Path.Kind != token.STRING {
				return nil, nil, fmt.Errorf("import path is not a string token")
			}
			importDecl := ImportDeclaration{
				Path: strings.Trim(spec.Path.Value, "\""),
			}
			if spec.Name != nil {
				importDecl.Name = spec.Name.Name
			}
			importDecls = append(importDecls, importDecl)

		case *ast.TypeSpec:
			names = append(names, spec.Name.Name)
		case *ast.ValueSpec:
			for _, name := range spec.Names {
				names = append(names, name.Name)
			}
		default:
			return nil, nil, fmt.Errorf("unknown spec type %T", spec)
		}
	}
	return names, importDecls, nil
}

func getFunctionName(path []ast.Node, fn *ast.FuncDecl) (string, error) {
	funcName := fn.Name.Name
	if fn.Recv != nil {
		recvName, ok := exprName(fn.Recv.List[0].Type)
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
	case *ast.IndexExpr:
		return typename.X.(*ast.Ident).Name, true
	default:
		return "", false
	}
}

func exprName(expr ast.Expr) (string, bool) {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name, true
	case *ast.StarExpr:
		return exprName(expr.X)
	case *ast.IndexExpr:
		return exprName(expr.X)
	default:
		return "", false
	}
}

func walk(node ast.Node, fn func(ast.Node) error) error {
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
