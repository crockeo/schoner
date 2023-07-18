package set

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(map[T]struct{})
}

func (s Set[T]) Add(v T) bool {
	if _, ok := s[v]; ok {
		return false
	}
	s[v] = struct{}{}
	return true
}

func (s Set[T]) Remove(v T) bool {
	if _, ok := s[v]; !ok {
		return false
	}
	delete(s, v)
	return true
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(s))
	for v := range s {
		slice = append(slice, v)
	}
	return slice
}
