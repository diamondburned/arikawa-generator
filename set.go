package main

type Set[K comparable] map[K]struct{}

func NewSet[K comparable](items ...K) Set[K] {
	s := make(Set[K])
	s.Add(items...)
	return s
}

func (s Set[K]) Add(items ...K) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Set[K]) Remove(items ...K) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s Set[K]) Has(item K) bool {
	_, ok := s[item]
	return ok
}
