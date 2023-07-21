package references

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/crockeo/schoner/pkg/astutil"
	"github.com/crockeo/schoner/pkg/phases/declarations"
	"github.com/crockeo/schoner/pkg/set"
)

type References struct {
	References map[string]set.Set[string]
}

func (r *References) Add(from string, to string) {
	if _, ok := r.References[from]; !ok {
		r.References[from] = set.NewSet[string]()
	}
	r.References[from].Add(to)
}

func FindReferences(
	context map[string]*declarations.Declarations,
	fileset *token.FileSet,
	filename string,
	contents []byte,
) (*References, error) {
	_, ok := context[filename]
	if !ok {
		return nil, fmt.Errorf("no declarations for file %s", filename)
	}

	fileAst, err := parser.ParseFile(fileset, filename, contents, 0)
	if err != nil {
		return nil, err
	}

	references := &References{References: map[string]set.Set[string]{}}
	path := []ast.Node{}
	err = astutil.Walk(fileAst, func(node ast.Node) error {
		if node == nil {
			path = path[:len(path)-1]
			return nil
		}
		defer func() { path = append(path, node) }()

		from := filename
		container, err := astutil.OuterDeclName(path)
		if err == nil {
			from = astutil.Qualify(from, container)
		}

		var to string
		switch node := node.(type) {
		case *ast.Ident:
			to = astutil.Qualify(filename, node.Name)
		default:
			return nil
		}

		references.Add(from, to)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return references, nil
}
