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
		if decl, ok := decl.(*ast.FuncDecl); ok {
			funcName, err := astutil.FunctionName(decl)
			if err != nil {
				return nil, err
			}
			declarations.Symbols.Add(funcName)
			continue
		}

		decl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range decl.Specs {
			switch spec := spec.(type) {
			case *ast.ImportSpec:
				if spec.Path.Kind != token.STRING {
					return nil, fmt.Errorf("import path is not a string token")
				}
				importDecl := ImportDeclaration{
					Path: strings.Trim(spec.Path.Value, "\""),
				}
				if spec.Name != nil {
					importDecl.Name = spec.Name.Name
				}
				declarations.Imports.Add(importDecl)

			case *ast.TypeSpec:
				declarations.Symbols.Add(spec.Name.Name)

			case *ast.ValueSpec:
				for _, name := range spec.Names {
					declarations.Symbols.Add(name.Name)
				}
			}
		}
	}

	return declarations, nil
}
