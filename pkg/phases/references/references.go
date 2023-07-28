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

func BuildReferenceGraph(
	root string,
	fileInfos map[string]*fileinfo.FileInfo,
	option walk.Option,
) (graph.Graph[fileinfo.Declaration], error) {
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
	FileInfos         map[string]*fileinfo.FileInfo
	ReferenceGraph    graph.Graph[fileinfo.Declaration]
	Root              string
	RootModule        string
	DeclarationLookup map[string]map[string]fileinfo.Declaration
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
		FileInfos:         fileInfos,
		ReferenceGraph:    graph.NewGraph[fileinfo.Declaration](),
		Root:              root,
		RootModule:        modulePath,
		DeclarationLookup: declarationLookup,
	}, nil
}

func (rgb *referenceGraphBuilder) Visit(filename string, fileInfo *fileinfo.FileInfo, fileAst *ast.File) error {
	for _, decl := range fileInfo.Declarations {
		rgb.ReferenceGraph.AddNode(decl)
	}

	ourModule, err := fileInfoModule(fileInfo, rgb.Root, rgb.RootModule)
	if err != nil {
		return fmt.Errorf("failed to determine current module: %w", err)
	}

	err = astutil.Walk(fileAst, func(path []ast.Node, node ast.Node) error {
		container, err := astutil.OuterDeclName(path)
		if err != nil {
			return nil
		}
		if astutil.IsQualified(container) {
			parts := astutil.Unqualify(container)
			container = parts[0]
		}
		from, ok := rgb.identReference(ourModule, container)
		if !ok {
			return nil
		}

		from.Name = container
		if astutil.IsQualified(from.Name) {
			parts := astutil.Unqualify(from.Name)
			from.Name = parts[0]
		}
		if from.Name == "_" {
			// TODO: unify this and the other branch in fileinfo.go
			// Typically values named `_` are intentionally unused,
			// and are used to assert that structs abide by interfaces.
			return nil
		}

		switch node := node.(type) {
		case *ast.Ident:
			target, ok := rgb.identReference(ourModule, node.Name)
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

func (rgb *referenceGraphBuilder) identReference(ourModule string, name string) (fileinfo.Declaration, bool) {
	declarations, ok := rgb.DeclarationLookup[ourModule]
	if !ok {
		return fileinfo.Declaration{}, false
	}
	declaration, ok := declarations[name]
	return declaration, ok
}

func (rgb *referenceGraphBuilder) selectorReference(currentFileInfo *fileinfo.FileInfo, selector *ast.SelectorExpr) (fileinfo.Declaration, bool) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return fileinfo.Declaration{}, false
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
		return fileinfo.Declaration{}, false
	}

	moduleDecls, ok := rgb.DeclarationLookup[importDecl.Path]
	if !ok {
		return fileinfo.Declaration{}, false
	}
	decl, ok := moduleDecls[selector.Sel.Name]
	return decl, ok
}

func makeDeclarationLookup(
	root string,
	rootModule string,
	fileInfos map[string]*fileinfo.FileInfo,
) (map[string]map[string]fileinfo.Declaration, error) {
	// module -> symbol -> declaration
	declarationLookup := map[string]map[string]fileinfo.Declaration{}
	for _, fileInfo := range fileInfos {
		module, err := fileInfoModule(fileInfo, root, rootModule)
		if err != nil {
			return nil, err
		}

		if _, ok := declarationLookup[module]; !ok {
			declarationLookup[module] = map[string]fileinfo.Declaration{}
		}

		for declName, decl := range fileInfo.Declarations {
			declarationLookup[module][declName] = decl
		}
	}
	return declarationLookup, nil
}

func fileInfoModule(fileInfo *fileinfo.FileInfo, root string, rootModule string) (string, error) {
	if !strings.HasPrefix(fileInfo.Filename, root) {
		return "", fmt.Errorf("file %s does not start with root %s", fileInfo.Filename, root)
	}
	suffix := fileInfo.Filename[len(root):]
	suffix = strings.TrimPrefix(suffix, "/")

	module := rootModule
	if suffix != "" {
		module = filepath.Join(module, filepath.Dir(suffix))
	}

	return module, nil
}
