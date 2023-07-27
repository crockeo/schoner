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

type Import struct {
	Name string
	Path string
}

type FileInfo struct {
	Filename     string
	Package      string
	Entrypoints  set.Set[string]
	Declarations set.Set[string]
	Imports      set.Set[Import]
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

func parseFileInfo(filename string, fileAst *ast.File) (*FileInfo, error) {
	fileInfo := &FileInfo{
		Package:      fileAst.Name.Name,
		Filename:     filename,
		Entrypoints:  set.NewSet[string](),
		Declarations: set.NewSet[string](),
		Imports:      set.NewSet[Import](),
	}

	for _, decl := range fileAst.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			name, err := astutil.FunctionName(decl)
			if err != nil {
				return nil, err
			}
			if astutil.IsQualified(name) {
				continue
			}
			fileInfo.Declarations.Add(name)
			isMainFunc := fileInfo.Package == "main" && name == "main"
			if isMainFunc || isTestFuncDecl(decl) {
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
					fileInfo.Declarations.Add(spec.Name.Name)
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						fileInfo.Declarations.Add(name.Name)
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
