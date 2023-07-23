package references

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/graph"
	"github.com/crockeo/schoner/pkg/phases/fileinfo"
	"github.com/crockeo/schoner/pkg/walk"
	"golang.org/x/mod/modfile"
)

// TODO: find all references in a file
//
// 3 kinds of references:
//
// - [x] static local references, where an ast.Ident references a delcaration in the current file
//
// - [x] static imported references, where an ast.SelectorExpr references a declaration in an imported file
//
// - [ ] dynamic references, where an ast.SelectorExpr references a field or method on another

type Declaration struct {
	Parent *fileinfo.FileInfo
	Name   string
}

func BuildReferenceGraph(
	root string,
	fileInfos map[string]*fileinfo.FileInfo,
	option walk.Option,
) (graph.Graph[Declaration], error) {
	builder, err := newReferenceGraphBuilder(root, fileInfos)
	if err != nil {
		return nil, err
	}
	fileset := token.NewFileSet()
	for path, fileInfo := range fileInfos {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fileAst, err := parser.ParseFile(fileset, path, contents, 0)
		if err != nil {
			return nil, err
		}
		if err := builder.Visit(path, fileInfo, fileAst); err != nil {
			return nil, err
		}
	}
	return builder.ReferenceGraph, nil
}

type referenceGraphBuilder struct {
	ReferenceGraph    graph.Graph[Declaration]
	RootModule        string
	DeclarationLookup map[string]map[string]Declaration
}

func newReferenceGraphBuilder(root string, fileInfos map[string]*fileinfo.FileInfo) (*referenceGraphBuilder, error) {
	goModContents, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("failed to find go module root: %w", err)
	}
	modulePath := modfile.ModulePath(goModContents)
	if modulePath == "" {
		return nil, fmt.Errorf("failed to parse module path from go module: %w", err)
	}
	declarationLookup, err := makeDeclarationLookup(root, modulePath, fileInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to create declaration lookup: %w", err)
	}
	return &referenceGraphBuilder{
		ReferenceGraph:    graph.NewGraph[Declaration](),
		RootModule:        modulePath,
		DeclarationLookup: declarationLookup,
	}, nil
}

func (rgb *referenceGraphBuilder) Visit(filename string, fileInfo *fileinfo.FileInfo, fileAst *ast.File) error {
	path := []ast.Node{}
	err := astutil.Walk(fileAst, func(node ast.Node) error {
		if node == nil {
			path = path[:len(path)-1]
			return nil
		}
		defer func() { path = append(path, node) }()

		from := Declaration{
			Parent: fileInfo,
			Name:   "root",
		}
		container, err := astutil.OuterDeclName(path)
		if err == nil {
			from.Name = container
		}

		switch node := node.(type) {
		case *ast.Ident:
			target, ok := rgb.identReference(fileInfo, node)
			if ok && from != target {
				rgb.ReferenceGraph.AddEdge(from, target)
			}
		case *ast.SelectorExpr:
			target, ok := rgb.selectorReference(fileInfo, node)
			if ok && from != target {
				rgb.ReferenceGraph.AddEdge(from, target)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (rgb *referenceGraphBuilder) identReference(currentFileInfo *fileinfo.FileInfo, ident *ast.Ident) (Declaration, bool) {
	if !currentFileInfo.Declarations.Contains(ident.Name) {
		return Declaration{}, false
	}
	return Declaration{
		Parent: currentFileInfo,
		Name: ident.Name,
	}, true
}

func (rgb *referenceGraphBuilder) selectorReference(currentFileInfo *fileinfo.FileInfo, selector *ast.SelectorExpr) (Declaration, bool) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return Declaration{}, false
	}
	selectorName := ident.Name

	foundImport := false
	var importDecl fileinfo.Import
	for importDecl = range currentFileInfo.Imports {
		if selectorName == importDecl.Name {
			foundImport = true
			break
		}
	}
	if !foundImport {
		return Declaration{}, false
	}

	moduleDecls, ok := rgb.DeclarationLookup[importDecl.Path]
	if !ok {
		return Declaration{}, false
	}
	decl, ok := moduleDecls[selector.Sel.Name]
	return decl, ok
}

func makeDeclarationLookup(
	root string,
	rootModule string,
	fileInfos map[string]*fileinfo.FileInfo,
) (map[string]map[string]Declaration, error) {
	// module -> symbol -> declaration
	declarationLookup := map[string]map[string]Declaration{}
	for filename, fileInfo := range fileInfos {
		if !strings.HasPrefix(filename, root) {
			return nil, fmt.Errorf("file %s does not start with root %s", filename, root)
		}
		suffix := filename[len(root):]
		suffix = strings.TrimPrefix(suffix, "/")

		module := rootModule
		if suffix != "" {
			module = filepath.Join(module, filepath.Dir(suffix))
		}

		if _, ok := declarationLookup[module]; !ok {
			declarationLookup[module] = map[string]Declaration{}
		}

		for decl := range fileInfo.Declarations {
			declarationLookup[module][decl] = Declaration{
				Parent: fileInfo,
				Name:   decl,
			}
		}
	}
	return declarationLookup, nil
}
