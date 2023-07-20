package decls

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/crockeo/schoner/pkg/set"
)

type FileDeclaration struct {
	Symbols []string
	Imports []ImportDeclaration
}

type ImportDeclaration struct {
	Name string
	Path string
}

type option struct {
	ignoreDirs set.Set[string]
}

type Option func(*option)

func WithIgnoreDirs(dirs ...string) Option {
	return func (o *option) {
		for _, dir := range dirs {
			o.ignoreDirs.Add(dir)
		}
	}
}

func FindProjectDeclarations(projectRoot string, options ...Option) (map[string]*FileDeclaration, error) {
	o := option{ignoreDirs: set.NewSet[string]()}
	for _, option := range options {
		option(&o)
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(projectRoot) {
		projectRoot = filepath.Join(wd, projectRoot)
	}

	fileset := token.NewFileSet()
	fileDeclarations := map[string]*FileDeclaration{}
	err = filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if o.ignoreDirs.Contains(filepath.Base(path)) {
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
		fileDeclaration, err := findFileDeclaration(fileset, path, contents)
		if err != nil {
			return err
		}
		fileDeclarations[path] = fileDeclaration
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileDeclarations, nil
}

func FindFileDeclarations(filename string) (*FileDeclaration, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(wd, filename)
	}
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	fileset := token.NewFileSet()
	return findFileDeclaration(fileset, filename, contents)
}

func findFileDeclaration(fileset *token.FileSet, filename string, contents []byte) (*FileDeclaration, error) {
	fileAst, err := parser.ParseFile(fileset, filename, contents, 0)
	if err != nil {
		return nil, err
	}
	symbols, err := collectNames(fileAst)
	if err != nil {
		return nil, err
	}
	return &FileDeclaration{
		Symbols: symbols,
	}, nil
}

func collectNames(fileAst *ast.File) ([]string, error) {
	names := []string{}
	path := []ast.Node{}
	return names, walk(fileAst, func(node ast.Node) error {
		if node == nil {
			path = path[:len(path)-1]
			return nil
		}

		switch node := node.(type) {
		case *ast.GenDecl:
			declNames, err := getDeclNames(path, node)
			if err != nil {
				return err
			}
			names = append(names, declNames...)
		case *ast.FuncDecl:
			funcName, err := getFunctionName(path, node)
			if err != nil {
				return err
			}
			names = append(names, funcName)
		}
		path = append(path, node)
		return nil
	})
}

func getDeclNames(path []ast.Node, decl *ast.GenDecl) ([]string, error) {
	lastNode := path[len(path)-1]
	if _, ok := lastNode.(*ast.File); !ok {
		// We only care about top-level declarations,
		// because only those can be used by other packages.
		return nil, nil
	}

	names := []string{}
	for _, spec := range decl.Specs {
		switch spec := spec.(type) {
		case *ast.ImportSpec:
			// Intentionally ignoring ImportSpec in this phase,
			// because we only care about declarations which are defined
			// inside of this file.
		case *ast.TypeSpec:
			names = append(names, spec.Name.Name)
		case *ast.ValueSpec:
			for _, name := range spec.Names {
				names = append(names, name.Name)
			}
		default:
			return nil, fmt.Errorf("unknown spec type %T", spec)
		}
	}
	return names, nil
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
