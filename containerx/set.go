package containerx

type Empty struct{}

type Set[T comparable] map[T]Empty

func NewSet[T comparable](items ...T) Set[T] {
	set := Set[T]{}
	set.Insert(items...)
	return set
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) Insert(items ...T) {
	for _, item := range items {
		s[item] = Empty{}
	}
}

func (s Set[T]) Delete(items ...T) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s Set[T]) Has(item T) bool {
	_, contained := s[item]
	return contained
}

func (s Set[T]) HasAll(items ...T) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

func (s Set[T]) HasAny(items ...T) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

func (s Set[T]) Slice() []T {
	slice := make([]T, 0, len(s))
	for item := range s {
		slice = append(slice, item)
	}
	return slice
}
