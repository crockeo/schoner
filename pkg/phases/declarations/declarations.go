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
	Declarations set.Set[Declaration]
	Imports      set.Set[Import]
}

type Declaration struct {
	Filename string
	Name     string
}

type Import struct {
	Filename string
	Name     string
	Path     string
}

func FindDeclarations(context struct{}, fileset *token.FileSet, filename string, contents []byte) (*Declarations, error) {
	fileAst, err := parser.ParseFile(fileset, filename, contents, 0)
	if err != nil {
		return nil, err
	}

	// TODO: this could probably be simplified by just looping through fileAst.Decls
	declarations := &Declarations{
		Declarations: set.NewSet[Declaration](),
		Imports:      set.NewSet[Import](),
	}
	for _, decl := range fileAst.Decls {
		if decl, ok := decl.(*ast.FuncDecl); ok {
			funcName, err := astutil.FunctionName(decl)
			if err != nil {
				return nil, err
			}
			declarations.Declarations.Add(Declaration{
				Filename: filename,
				Name:     funcName,
			})
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
				importDecl := Import{
					Path: strings.Trim(spec.Path.Value, "\""),
				}
				if spec.Name != nil {
					importDecl.Name = spec.Name.Name
				}
				declarations.Imports.Add(importDecl)

			case *ast.TypeSpec:
				declarations.Declarations.Add(Declaration{
					Filename: filename,
					Name:     spec.Name.Name,
				})

			case *ast.ValueSpec:
				for _, name := range spec.Names {
					declarations.Declarations.Add(Declaration{
						Filename: filename,
						Name:     name.Name,
					})
				}
			}
		}
	}

	return declarations, nil
}
