package scan

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/set"
)

var (
	ErrMissingMain         = errors.New("main package, but does not contain main")
	ErrImportPathNotString = errors.New("import path is not a string")
)

type Import struct {
	Path string
	Name string
}

type Reference struct {
	Package string
	Name    string
}

type FileInfo struct {
	Package      string
	Filename     string
	Entrypoints  set.Set[string]
	Declarations set.Set[string]
	References   set.Set[Reference]
	Imports      set.Set[Import]
}

func NewFileInfo() *FileInfo {
	return &FileInfo{
		Entrypoints:  set.NewSet[string](),
		Declarations: set.NewSet[string](),
		References:   set.NewSet[Reference](),
		Imports:      set.NewSet[Import](),
	}
}

func (fi *FileInfo) String() string {
	builder := strings.Builder{}
	fmt.Fprintln(&builder, "Filename:", fi.Filename)
	fmt.Fprintln(&builder, "Package:", fi.Package)

	if len(fi.Entrypoints) > 0 {
		fmt.Fprintln(&builder, "Entrypoints:")
		for entrypoint := range fi.Entrypoints {
			fmt.Fprintln(&builder, "  ", entrypoint)
		}
	}

	if len(fi.Declarations) > 0 {
		fmt.Fprintln(&builder, "Declarations:")
		for declaration := range fi.Declarations {
			fmt.Fprintln(&builder, "  ", declaration)
		}
	}

	if len(fi.References) > 0 {
		fmt.Fprintln(&builder, "References:")
		for reference := range fi.References {
			fmt.Fprintln(&builder, "  ", reference.Package, reference.Name)
		}
	}

	if len(fi.Imports) > 0 {
		fmt.Fprintln(&builder, "Imports:")
		for importDecl := range fi.Imports {
			fmt.Fprintln(&builder, "  ", importDecl.Name, "from", importDecl.Path)
		}
	}

	return builder.String()
}

type walkFilesOptions struct {
	ignoreDirs set.Set[string]
}

type Option func(*walkFilesOptions)

func WithIgnoreDirs(dirs ...string) func(*walkFilesOptions) {
	return func(wfo *walkFilesOptions) {
		for _, dir := range dirs {
			wfo.ignoreDirs.Add(dir)
		}
	}
}

func ScanFileInfo(root string, options ...Option) (map[string]*FileInfo, error) {
	realizedOptions := walkFilesOptions{
		ignoreDirs: set.NewSet[string](),
	}
	for _, option := range options {
		option(&realizedOptions)
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	fileset := token.NewFileSet()
	fileInfos := map[string]*FileInfo{}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if realizedOptions.ignoreDirs.Contains(filepath.Base(path)) {
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
		fileAst, err := parser.ParseFile(fileset, path, contents, 0)
		if err != nil {
			return fmt.Errorf("failed to parse AST for `%s`: %w", path, err)
		}
		fileInfo, err := parseFileInfo(path, fileAst)
		if err != nil {
			return fmt.Errorf("failed to file info for `%s`: %w", path, err)
		}
		fileInfos[path] = fileInfo
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileInfos, nil
}

func parseFileInfo(fileName string, fileAst *ast.File) (*FileInfo, error) {
	fileInfo := NewFileInfo()
	fileInfo.Package = fileAst.Name.Name
	fileInfo.Filename = fileName

	if err := populateDeclarations(fileInfo, fileAst); err != nil {
		return nil, err
	}

	// TODO: references
	// References   set.Set[string]
	path := []ast.Node{}
	ast.Inspect(fileAst, func(node ast.Node) bool {
		if node == nil {
			path = path[:len(path)-1]
			return true
		}
		defer func() { path = append(path, node) }()

		switch node := node.(type) {
		case *ast.Ident:
			fileInfo.References.Add(Reference{
				// TODO: this is wrong; it should be
				// the full import path of the package,
				// not just its name
				Package: fileInfo.Package,
				Name:    node.Name,
			})
			return false
		case *ast.SelectorExpr:
			ident, ok := node.X.(*ast.Ident)
			if !ok {
				return true
			}
			selectorName := ident.Name
			for importDecl := range fileInfo.Imports {
				if selectorName != importDecl.Name {
					continue
				}
				fileInfo.References.Add(Reference{
					Package: importDecl.Path,
					Name:    node.Sel.Name,
				})
				return false
			}
			return true
		}

		return true
	})

	// TODO: populate entrypoints for tests
	// Entrypoints  set.Set[string]
	if fileInfo.Package == "main" {
		if !fileInfo.Declarations.Contains("main") {
			return nil, ErrMissingMain
		}
		fileInfo.Entrypoints.Add("main")
	}

	return fileInfo, nil
}

func populateDeclarations(fileInfo *FileInfo, fileAst *ast.File) error {
	for _, decl := range fileAst.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			name, err := astutil.FunctionName(decl)
			if err != nil {
				return err
			}
			fileInfo.Declarations.Add(name)

		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					if spec.Path.Kind != token.STRING {
						return ErrImportPathNotString
					}
					path := strings.Trim(spec.Path.Value, "\"")
					name := filepath.Base(path)
					if spec.Name != nil {
						name = spec.Name.Name
					}
					fileInfo.Imports.Add(Import{
						Name: name,
						Path: path,
					})
				case *ast.TypeSpec:
					fileInfo.Declarations.Add(spec.Name.Name)
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						fileInfo.Declarations.Add(name.Name)
					}
				}
			}
		}
	}
	return nil
}

func findImportDecl(imports set.Set[Import], selectorName string) (*Import, bool) {
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
