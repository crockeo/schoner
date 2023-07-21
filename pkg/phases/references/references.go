package references

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/phases/declarations"
	"github.com/crockeo/schoner/pkg/set"
)

// find all references in a file
//
// 3 kinds of references:
//
// - static local references, where an ast.Ident references a delcaration in the current file
//
// - static imorted references, where an ast.SelectorExpr references a declaration in an imported file
//
// - dynamic references, where an ast.SelectorExpr references a field or method on another

type ReferencesContext struct {
	ProjectRoot       string
	GoModRoot         string
	Declarations      map[string]*declarations.Declarations
	DeclarationLookup map[string]map[string]declarations.Declaration
}

type References struct {
	References map[declarations.Declaration]set.Set[declarations.Declaration]
}

func (r *References) Add(from declarations.Declaration, to declarations.Declaration) {
	if from == to {
		return
	}

	if _, ok := r.References[from]; !ok {
		r.References[from] = set.NewSet[declarations.Declaration]()
	}
	r.References[from].Add(to)
}

func FindReferences(
	ctx *ReferencesContext,
	fileset *token.FileSet,
	filename string,
	contents []byte,
) (*References, error) {
	ourDeclarations, ok := ctx.Declarations[filename]
	if !ok {
		return nil, fmt.Errorf("no declarations for file %s", filename)
	}
	ourDeclarationsByName := map[string]declarations.Declaration{}
	for decl := range ourDeclarations.Declarations {
		ourDeclarationsByName[decl.Name] = decl
	}

	fileAst, err := parser.ParseFile(fileset, filename, contents, 0)
	if err != nil {
		return nil, err
	}

	references := &References{References: map[declarations.Declaration]set.Set[declarations.Declaration]{}}
	path := []ast.Node{}
	err = astutil.Walk(fileAst, func(node ast.Node) error {
		if node == nil {
			path = path[:len(path)-1]
			return nil
		}
		defer func() { path = append(path, node) }()

		from := declarations.Declaration{Filename: filename}
		container, err := astutil.OuterDeclName(path)
		if err == nil {
			from.Name = container
		}

		if node, ok := node.(*ast.Ident); ok {
			if _, ok := ourDeclarationsByName[node.Name]; !ok {
				return nil
			}
			references.Add(from, declarations.Declaration{
				Filename: filename,
				Name:     node.Name,
			})
			return nil
		}

		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return nil
		}
		selectorTarget, ok := findSelectorTarget(ctx, ourDeclarations, selector)
		if !ok {
			return nil
		}
		references.Add(from, selectorTarget)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return references, nil
}

func findSelectorTarget(
	ctx *ReferencesContext,
	ourDeclarations *declarations.Declarations,
	selector *ast.SelectorExpr,
) (declarations.Declaration, bool) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return declarations.Declaration{}, false
	}
	selectorName := ident.Name
	importDecl, ok := findImportDecl(ourDeclarations.Imports, selectorName)
	if !ok {
		return declarations.Declaration{}, false
	}

	moduleDecls, ok := ctx.DeclarationLookup[importDecl.Path]
	if !ok {
		return declarations.Declaration{}, false
	}
	decl, ok := moduleDecls[selector.Sel.Name]
	return decl, ok
}

func findImportDecl(imports set.Set[declarations.Import], selectorName string) (*declarations.Import, bool) {
	for importDecl := range imports {
		if selectorName == importDecl.Name {
			return &importDecl, true
		}

		lastPart := filepath.Base(importDecl.Path)
		if selectorName == lastPart {
			return &importDecl, true
		}
	}
	return nil, false
}
