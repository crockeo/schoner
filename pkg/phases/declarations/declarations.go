package declarations

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/set"
)

type Declarations struct {
	Symbols set.Set[string]
	Imports set.Set[ImportDeclaration]
}

type ImportDeclaration struct {
	Name string
	Path string
}

func FindDeclarations(context struct{}, fileset *token.FileSet, filename string, contents []byte) (*Declarations, error) {
	fileAst, err := parser.ParseFile(fileset, filename, contents, 0)
	if err != nil {
		return nil, err
	}

	// TODO: this could probably be simplified by just looping through fileAst.Decls
	declarations := &Declarations{
		Symbols: set.NewSet[string](),
		Imports: set.NewSet[ImportDeclaration](),
	}
	for _, decl := range fileAst.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			declNames, importDecls, err := getDeclNames(decl)
			if err != nil {
				return nil, err
			}
			for _, declName := range declNames {
				declarations.Symbols.Add(declName)
			}
			for _, importDecl := range importDecls {
				declarations.Imports.Add(importDecl)
			}
		case *ast.FuncDecl:
			funcName, err := astutil.FunctionName(decl)
			if err != nil {
				return nil, err
			}
			declarations.Symbols.Add(funcName)
		}
	}
	if err != nil {
		return nil, err
	}

	return declarations, nil
}

func getDeclNames(decl *ast.GenDecl) ([]string, []ImportDeclaration, error) {
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
