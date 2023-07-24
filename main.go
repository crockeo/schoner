package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Command int

const (
	CommandVisualize Command = iota
	CommandUnreachable
)

func mainImpl() error {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return printHelp("")
	}

	commandStr := args[0]
	var command Command
	switch commandStr {
	case "visualize":
		command = CommandVisualize
	case "unreachable":
		command = CommandUnreachable
	default:
		return printHelp(fmt.Sprintf("unrecognized command: %s", commandStr))
	}

	roots := args[1:]

	walkOptions := walk.WithOptions(
		walk.WithIgnoreDirs(".git"),
		walk.WithIgnoreTests(true),
	)
	for _, root := range roots {
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

		unreachable := set.NewSet[references.Declaration]()
		entrypoints := set.NewSet[references.Declaration]()
		for decl := range referenceGraph {
			unreachable.Add(decl)
			if decl.Parent.Entrypoints.Contains(decl.Name) {
				entrypoints.Add(decl)
			}
		}
		_ = referenceGraph.DFS(entrypoints.ToSlice(), func(node references.Declaration) error {
			unreachable.Remove(node)
			return nil
		})

		switch command {
		case CommandVisualize:
			if err := cmdVisualize(root, referenceGraph, entrypoints, unreachable); err != nil {
				return err
			}
		case CommandUnreachable:
			if err := cmdUnreachable(root, unreachable); err != nil {
				return err
			}
		default:
			panic("unreachable")
		}

	}
	return nil
}

func cmdVisualize(
	root string,
	referenceGraph graph.Graph[references.Declaration],
	unreachable set.Set[references.Declaration],
	entrypoints set.Set[references.Declaration],
) error {
	return visualize.Visualize(
		fmt.Sprintf("%s.svg", filepath.Base(root)),
		referenceGraph,
		entrypoints,
		unreachable,
	)
}

func cmdUnreachable(root string, unreachable set.Set[references.Declaration]) error {
	unreachableNames := make([]string, 0, len(unreachable))
	for decl := range unreachable {
		filename := decl.Parent.Filename
		if !strings.HasPrefix(filename, root) {
			return fmt.Errorf("file %s does not begin with expected root %s", filename, root)
		}
		filename = filename[len(root):]
		filename = strings.TrimPrefix(filename, "/")
		unreachableNames = append(unreachableNames, astutil.Qualify(filename, decl.Name))
	}
	sort.Strings(unreachableNames)
	for _, unreachableName := range unreachableNames {
		fmt.Println(unreachableName)
	}
	return nil
}

func printHelp(extraMessage string) error {
	return errors.New(extraMessage)
}
