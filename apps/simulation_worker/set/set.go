package set

type Set[T comparable] struct {
	data map[T]struct{}
}

func New[T comparable]() *Set[T] {
	s := Set[T]{}
	s.data = make(map[T]struct{})

	return &s
}

func (s *Set[T]) Has(value T) bool {
	if s.IsZero() {
		s.init()
		return false
	}
	_, exists := s.data[value]
	return exists
}

// Add adds a new item to the set if not present.
// The return value indicates whether or not a new
// value was actually added.
//
// Value was already in the set? returns false
// Value was not already in the set? returns true.
func (s *Set[T]) Add(value T) bool {
	if s.data == nil {
		s.init()
		s.data[value] = struct{}{}

		return true // early return bc we just initialized so it must be new
	}

	_, alreadyHas := s.data[value]
	if alreadyHas {
		return false
	}

	s.data[value] = struct{}{}

	return true
}

// IsZero returns whether or not the set is equal to the zero value
// as in, whether or not its initialized. True if zero, false if not.
// If s itself is nil, then we also return true.
func (s *Set[T]) IsZero() bool {
	return s == nil || s.data == nil
}

func (s *Set[T]) init() {
	if s.data != nil {
		panic("init called when s.data was not nil")
	}
	s.data = make(map[T]struct{})
}
