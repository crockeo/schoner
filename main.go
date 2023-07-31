package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/graph"
	"github.com/crockeo/schoner/pkg/phases/fileinfo"
	"github.com/crockeo/schoner/pkg/phases/references"
	"github.com/crockeo/schoner/pkg/set"
	"github.com/crockeo/schoner/pkg/visualize"
	"github.com/crockeo/schoner/pkg/walk"
)

func main() {
	if err := mainImpl(); err != nil {
		errMsg := strings.TrimSpace(err.Error())
		fmt.Fprintln(os.Stderr, errMsg)
		os.Exit(1)
	}
}

type args struct {
	Visualize   visualizeArgs   `cmd:"" help:"Visualize references in a project."`
	Unreachable unreachableArgs `cmd:"" help:"List all unreachable declarations in a project."`
}

type visualizeArgs struct {
	OutputDir string   `name:"output-dir" help:"The directory in which .svg files will be generated."`
	Paths     []string `arg:"" name:"path" help:"List of projects to visualize." type:"path"`
}

type unreachableArgs struct {
	Paths []string `arg:"" name:"path" help:"List of projects to analyze." type:"path"`
}

func mainImpl() error {
	args := args{}
	ctx := kong.Parse(&args)
	switch ctx.Command() {
	case "visualize <path>":
		return visualizeMain(args.Visualize)
	case "unreachable <path>":
		return unreachableMain(args.Unreachable)
	default:
		panic("unreachable")
	}
}

func visualizeMain(args visualizeArgs) error {
	outputDir, err := filepath.Abs(args.OutputDir)
	if err != nil {
		return err
	}
	// TODO: check that outputDir is actually a directory

	for _, path := range args.Paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		// TODO: check that path is a directory

		analysis, err := analyzeProject(path)
		if err != nil {
			return err
		}
		visualize.Visualize(
			fmt.Sprintf("%s.svg", filepath.Join(outputDir, filepath.Base(path))),
			analysis.ReferenceGraph,
			analysis.Entrypoints,
			analysis.Unreachable,
		)
	}
	return nil
}

func unreachableMain(args unreachableArgs) error {
	for _, path := range args.Paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		// TODO: check that path is a directory

		analysis, err := analyzeProject(path)
		if err != nil {
			return err
		}

		unreachableByName := map[string]fileinfo.Declaration{}
		unreachableNames := make([]string, 0, len(analysis.Unreachable))
		for decl := range analysis.Unreachable {
			filename := decl.Parent.Filename
			if !strings.HasPrefix(filename, path) {
				return fmt.Errorf("file %s does not begin with expected path %s", filename, path)
			}
			filename = filename[len(path):]
			filename = strings.TrimPrefix(filename, "/")
			unreachableName := astutil.Qualify(filename, decl.Name)

			unreachableByName[unreachableName] = decl
			unreachableNames = append(unreachableNames, unreachableName)
		}
		sort.Strings(unreachableNames)
		for _, unreachableName := range unreachableNames {
			fmt.Println(unreachableName)
		}
		return nil
	}
	return nil
}

type analysis struct {
	FileInfos      map[string]*fileinfo.FileInfo
	ReferenceGraph graph.Graph[fileinfo.Declaration]
	Unreachable    set.Set[fileinfo.Declaration]
	Entrypoints    set.Set[fileinfo.Declaration]
}

func analyzeProject(path string) (analysis, error) {
	// TODO: make these configurable?
	walkOptions := walk.WithOptions(
		walk.WithIgnoreDirs(".git"),
		// walk.WithIgnoreTests(true),
	)

	fileInfos, err := fileinfo.FindFileInfos(path, walkOptions)
	if err != nil {
		return analysis{}, err
	}

	referenceGraph, err := references.BuildReferenceGraph(path, fileInfos, walkOptions)
	if err != nil {
		return analysis{}, err
	}

	unreachable := set.NewSet[fileinfo.Declaration]()
	entrypoints := set.NewSet[fileinfo.Declaration]()
	for decl := range referenceGraph {
		unreachable.Add(decl)
		if decl.Parent.Entrypoints.Contains(decl.Name) {
			entrypoints.Add(decl)
		}
	}
	_ = referenceGraph.DFS(entrypoints.ToSlice(), func(node fileinfo.Declaration) error {
		unreachable.Remove(node)
		return nil
	})

	return analysis{
		fileInfos,
		referenceGraph,
		unreachable,
		entrypoints,
	}, nil
}
