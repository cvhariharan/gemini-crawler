package utils

// Set represents the in-memory visited set. Lock is not required since the crawler
// goroutines already use channels so only one crawler gets a path at any time
// and also lock is already used in the crawler goroutine
type Set struct {
	m map[string]bool
}

func NewMemSet() VisitedSet {
	return &Set{
		m: make(map[string]bool),
	}
}

func (s *Set) IsVisited(path string) bool {
	return s.m[path]
}

func (s *Set) Visit(path string) error {
	s.m[path] = true
	return nil
}
