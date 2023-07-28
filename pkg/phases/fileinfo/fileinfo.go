package fileinfo

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/set"
	"github.com/crockeo/schoner/pkg/walk"
)

var (
	ErrImportPathNotString = errors.New("import path is not a string")
)

type FileInfo struct {
	Filename     string
	Package      string
	Entrypoints  set.Set[string]
	Declarations map[string]Declaration
	Imports      set.Set[Import]
}

type Declaration struct {
	Parent *FileInfo
	Name   string
	Pos    token.Position
	End    token.Position
}

type Import struct {
	Name string
	Path string
}

func FindFileInfos(root string, option walk.Option) (map[string]*FileInfo, error) {
	fileset := token.NewFileSet()
	fileInfos := map[string]*FileInfo{}
	err := walk.GoFiles(root, option, func(path string) error {
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fileAst, err := parser.ParseFile(fileset, path, contents, 0)
		if err != nil {
			return fmt.Errorf("failed to parse AST for `%s`: %w", path, err)
		}
		fileInfo, err := parseFileInfo(fileset, path, fileAst)
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

func parseFileInfo(fileset *token.FileSet, filename string, fileAst *ast.File) (*FileInfo, error) {
	fileInfo := &FileInfo{
		Package:      fileAst.Name.Name,
		Filename:     filename,
		Entrypoints:  set.NewSet[string](),
		Declarations: map[string]Declaration{},
		Imports:      set.NewSet[Import](),
	}

	for _, decl := range fileAst.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			name, err := astutil.FunctionName(decl)
			if err != nil {
				return nil, err
			}
			fileInfo.Declarations[name] = Declaration{
				Parent: fileInfo,
				Name:   name,
				Pos:    fileset.Position(decl.Pos()),
				End:    fileset.Position(decl.End()),
			}
			isInitFunc := name == "init"
			isMainFunc := fileInfo.Package == "main" && name == "main"
			if isInitFunc || isMainFunc || isTestFuncDecl(decl) {
				fileInfo.Entrypoints.Add(name)
			}

		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					if spec.Path.Kind != token.STRING {
						return nil, ErrImportPathNotString
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
					fileInfo.Declarations[spec.Name.Name] = Declaration{
						Parent: fileInfo,
						Name:   spec.Name.Name,
						Pos:    fileset.Position(spec.Pos()),
						End:    fileset.Position(spec.End()),
					}
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						if name.Name == "_" {
							// TODO: unify this and the other branch in references.go
							// Typically values named `_` are intentionally unused,
							// and are used to assert that structs abide by interfaces.
							continue
						}
						fileInfo.Declarations[name.Name] = Declaration{
							Parent: fileInfo,
							Name:   name.Name,
							Pos:    fileset.Position(spec.Pos()),
							End:    fileset.Position(spec.End()),
						}
					}
				}
			}
		}
	}

	return fileInfo, nil
}

func isTestFuncDecl(decl *ast.FuncDecl) bool {
	name := decl.Name.Name
	if !strings.HasPrefix(name, "Test") {
		return false
	}
	name = name[len("Test"):]

	var firstChar rune
	for _, firstChar = range name {
		break
	}
	if !unicode.IsUpper(firstChar) {
		return false
	}

	return true
}
