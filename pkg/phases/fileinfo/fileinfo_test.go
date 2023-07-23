package fileinfo

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/crockeo/schoner/pkg/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fileContents string = `
package main

import (
    "fmt"

    namedimport "github.com/crockeo/schoner/something"
    "github.com/crockeo/schoner/unnamedimport"
)

func main() {
    otherThing()
}

func otherThing() { }

type StructType struct {}

var variable string = "string"
const constant string = "string2"
`

func TestParseFileInfo(t *testing.T) {
	fileset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fileset, "/fake/file", fileContents, 0)
	require.NoError(t, err)
	fileInfo, err := parseFileInfo("/fake/file", fileAst)
	require.NoError(t, err)

	assert.Equal(
		t,
		&FileInfo{
			Filename:     "/fake/file",
			Package:      "main",
			Entrypoints:  set.NewSet("main"),
			Declarations: set.NewSet("main", "otherThing", "StructType", "variable", "constant"),
			Imports: set.NewSet(
				Import{
					Name: "fmt",
					Path: "fmt",
				},
				Import{
					Name: "namedimport",
					Path: "github.com/crockeo/schoner/something",
				},
				Import{
					Name: "unnamedimport",
					Path: "github.com/crockeo/schoner/unnamedimport",
				},
			),
		},
		fileInfo,
	)
}
