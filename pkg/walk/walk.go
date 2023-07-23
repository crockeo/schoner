package walk

import (
	"os"
	"path/filepath"

	"github.com/crockeo/schoner/pkg/set"
)

type walkFilesOptions struct {
	ignoreDirs set.Set[string]
}

type Option func(*walkFilesOptions)

func WithIgnoreDirs(dirs ...string) Option {
	return func(wfo *walkFilesOptions) {
		for _, dir := range dirs {
			wfo.ignoreDirs.Add(dir)
		}
	}
}

func WithOptions(options ...Option) Option {
	return func(wfo *walkFilesOptions) {
		for _, option := range options {
			option(wfo)
		}
	}
}

// GoFiles walks through the directory `root` and calls the visitor on every `.go` file.
// See available Options for more configuration.
func GoFiles(root string, option Option, visitor func(path string) error) error {
	options := walkFilesOptions{ignoreDirs: set.NewSet[string]()}
	option(&options)

	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if options.ignoreDirs.Contains(filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}
		return visitor(path)
	})
}
