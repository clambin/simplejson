package set

type Set[K comparable] map[K]struct{}

func Create[K comparable](values []K) Set[K] {
	s := make(Set[K])

	for _, value := range values {
		s[value] = struct{}{}
	}
	return s
}

func (s Set[K]) Has(value K) bool {
	_, found := s[value]
	return found
}
