package testdata

import "github.com/crockeo/schoner/internal/testdata/otherpkg"

var Something string = otherpkg.Value

type Struct[T any] struct{}

func (s *String[T]) Asdf() {}

func (s *String[T]) Asdf2() {
	fmt.Println(Something)
}
