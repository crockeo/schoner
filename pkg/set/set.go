package set

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](elements ...T) Set[T] {
	set := Set[T](make(map[T]struct{}))
	for _, element := range elements {
		set.Add(element)
	}
	return set
}

func (s Set[T]) Add(v T) bool {
	if _, ok := s[v]; ok {
		return false
	}
	s[v] = struct{}{}
	return true
}

func (s Set[T]) UnionInPlace(other Set[T]) bool {
	addedAny := false
	for value := range other {
		if s.Add(value) {
			addedAny = true
		}
	}
	return addedAny
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
