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
	projectRoot string,
	context Context,
	visitor func(context Context, fileset *token.FileSet, filepath string, contents []byte) (Result, error),
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
	if !filepath.IsAbs(projectRoot) {
		projectRoot = filepath.Join(wd, projectRoot)
	}

	fileset := token.NewFileSet()
	results := map[string]Result{}
	if err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
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
		result, err := visitor(context, fileset, path, contents)
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
