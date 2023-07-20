package phases

import (
	"go/token"
	"os"
	"path/filepath"

	"github.com/crockeo/schoner/pkg/set"
)

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

func WalkFiles[Context any, Result any](
	ctx Context,
	root string,
	visitor func(ctx Context, fileset *token.FileSet, filename string, contents []byte) (Result, error),
	options ...Option,
) (map[string]Result, error) {
	realizedOptions := walkFilesOptions{
		ignoreDirs: set.NewSet[string](),
	}
	for _, option := range options {
		option(&realizedOptions)
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(root) {
		root = filepath.Join(wd, root)
	}

	fileset := token.NewFileSet()
	results := map[string]Result{}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
		result, err := visitor(ctx, fileset, path, contents)
		if err != nil {
			return err
		}
		results[path] = result
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}
