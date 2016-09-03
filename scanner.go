package main

type scanner struct {
	str string
	pos int
	len int
}

func newScanner(s string) *scanner {
	return &scanner{s, 0, len(s)}
}

func (s *scanner) more() bool {
	return s.pos < s.len
}

func (s *scanner) next() byte {
	if !more() {
		return -1
	}
	return s.str[s.pos]
}

func (s *scanner) get() byte {
	if !more() {
		return -1
	}
	ch := s.str[s.pos]
	s.pos++
	return ch
}

func (s *scanner) SkipStri(s string) bool {
	for _, ch := range(s) {
		if s.get() != ch {
			return false
		}
	}
	return true
}

func (s *scanner) rest() string {
	return s.str[s.pos:]
}
