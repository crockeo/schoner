package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crockeo/schoner/pkg/phases/fileinfo"
	"github.com/crockeo/schoner/pkg/phases/references"
	"github.com/crockeo/schoner/pkg/walk"
)

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainImpl() error {
	walkOptions := walk.WithIgnoreDirs(".git")
	for _, root := range os.Args[1:] {
		root, err := filepath.Abs(root)
		if err != nil {
			return err
		}

		fileInfos, err := fileinfo.FindFileInfos(root, walkOptions)
		if err != nil {
			return err
		}

		referenceGraph, err := references.BuildReferenceGraph(root, fileInfos, walkOptions)
		if err != nil {
			return err
		}

		_ = referenceGraph
	}

	return nil
}
