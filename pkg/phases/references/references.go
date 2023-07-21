package references

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

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
	ProjectRoot  string
	GoModRoot    string
	Declarations map[string]*declarations.Declarations
}

type References struct {
	References map[string]set.Set[string]
}

func (r *References) Add(from string, to string) {
	if from == to {
		return
	}

	if _, ok := r.References[from]; !ok {
		r.References[from] = set.NewSet[string]()
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

	fileAst, err := parser.ParseFile(fileset, filename, contents, 0)
	if err != nil {
		return nil, err
	}

	references := &References{References: map[string]set.Set[string]{}}
	path := []ast.Node{}
	err = astutil.Walk(fileAst, func(node ast.Node) error {
		if node == nil {
			path = path[:len(path)-1]
			return nil
		}
		defer func() { path = append(path, node) }()

		from := "global"
		container, err := astutil.OuterDeclName(path)
		if err == nil {
			from = container
		}

		if node, ok := node.(*ast.Ident); ok {
			if !ourDeclarations.Symbols.Contains(node.Name) {
				return nil
			}
			references.Add(from, node.Name)
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
) (string, bool) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	selectorName := ident.Name
	importDecl, ok := findImportDecl(ourDeclarations.Imports, selectorName)
	if !ok {
		return "", false
	}

	prefixLen := len(ctx.GoModRoot)
	if prefixLen > len(importDecl.Path) {
		prefixLen = len(importDecl.Path)
	}
	if importDecl.Path[:prefixLen] != ctx.GoModRoot {
		return "", false
	}
	pathSuffix := importDecl.Path[prefixLen:]
	pathSuffix = strings.TrimPrefix(pathSuffix, "/")

	searchTarget := ctx.ProjectRoot
	if pathSuffix != "" {
		searchTarget = filepath.Join(searchTarget, pathSuffix)
	}
	for filename, declarations := range ctx.Declarations {
		if !strings.HasPrefix(filename, searchTarget) {
			continue
		}
		if declarations.Symbols.Contains(selector.Sel.Name) {
			return astutil.Qualify(filename, selector.Sel.Name), true
		}
	}
	return "", false
}

func findImportDecl(imports set.Set[declarations.ImportDeclaration], selectorName string) (*declarations.ImportDeclaration, bool) {
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
