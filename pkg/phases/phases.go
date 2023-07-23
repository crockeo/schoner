package phases

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/crockeo/schoner/pkg/phases/declarations"
	"github.com/crockeo/schoner/pkg/phases/references"
	"github.com/crockeo/schoner/pkg/walk"
	"golang.org/x/mod/modfile"
)

func WalkFiles[Context any, Result any](
	ctx Context,
	root string,
	option walk.Option,
	visitor func(ctx Context, fileset *token.FileSet, filename string, contents []byte) (Result, error),
) (map[string]Result, error) {
	fileset := token.NewFileSet()
	results := map[string]Result{}
	err := walk.GoFiles(root, option, func(path string) error {
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result, err := visitor(ctx, fileset, path, contents)
		if err != nil {
			return err
		}
		results[path] = result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FindDeclarations(root string, option walk.Option) (map[string]*declarations.Declarations, error) {
	return WalkFiles(
		struct{}{},
		root,
		option,
		declarations.FindDeclarations,
	)
}

func FindReferences(root string, decls map[string]*declarations.Declarations, option walk.Option) (map[string]*references.References, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize the root")
	}
	goModContents, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("failed to find go module root: %w", err)
	}
	modulePath := modfile.ModulePath(goModContents)
	if modulePath == "" {
		return nil, fmt.Errorf("failed to parse module path from go module: %w", err)
	}
	declarationLookup, err := makeDeclarationLookup(root, modulePath, decls)
	if err != nil {
		return nil, fmt.Errorf("failed to produce declaration lookup: %w", err)
	}

	ctx := &references.ReferencesContext{
		ProjectRoot:       root,
		GoModRoot:         modulePath,
		Declarations:      decls,
		DeclarationLookup: declarationLookup,
	}
	return WalkFiles(
		ctx,
		root,
		option,
		references.FindReferences,
	)
}

func makeDeclarationLookup(
	root string,
	rootModule string,
	projectDecls map[string]*declarations.Declarations,
) (map[string]map[string]declarations.Declaration, error) {
	// module -> symbol -> declaration
	declarationLookup := map[string]map[string]declarations.Declaration{}
	for filename, decls := range projectDecls {
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
			declarationLookup[module] = map[string]declarations.Declaration{}
		}

		for decl := range decls.Declarations {
			declarationLookup[module][decl.Name] = decl
		}
	}
	return declarationLookup, nil
}
