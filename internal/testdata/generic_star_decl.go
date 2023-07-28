package testdata

import (
	"fmt"

	"github.com/crockeo/schoner/internal/testdata/otherpkg"
)

var Something string = otherpkg.Value

type Struct[T any] struct{}

func (s *Struct[T]) Asdf() {}

func (s *Struct[T]) Asdf2() {
	fmt.Println(Something)
}

type PlainStruct struct {}
type ArrayOfPlainStruct struct {
	Elements []PlainStruct
}
