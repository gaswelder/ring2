package scanner

import (
	"fmt"
)

type Scanner struct {
	str string
	pos int
	len int
	err error
}

func New(s string) *Scanner {
	return &Scanner{s, 0, len(s), nil}
}

func (s *Scanner) More() bool {
	if s.err != nil {
		return false
	}
	return s.pos < s.len
}

func (s *Scanner) Next() byte {
	if !s.More() {
		return 0
	}
	return s.str[s.pos]
}

func (s *Scanner) Get() byte {
	if !s.More() {
		return 0
	}
	ch := s.str[s.pos]
	s.pos++
	return ch
}

func (s *Scanner) SkipStri(str string) bool {
	if s.err != nil {
		return false
	}

	n := len(str)
	for i := 0; i < n; i++ {
		if toUpper(s.Get()) != toUpper(str[i]) {
			s.err = fmt.Errorf("expected '%s'", str)
			return false
		}
	}
	return true
}

func (s *Scanner) Rest() string {
	return s.str[s.pos:]
}

func (s *Scanner) Expect(ch byte) bool {
	if s.err != nil {
		return false
	}
	n := s.Get()
	if n != ch {
		s.err = fmt.Errorf("expected %c, got %c", ch, n)
		return false
	}
	return true
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) ReadName() string {
	name := ""
	if !isAlpha(s.Next()) {
		return name
	}

	for isAlpha(s.Next()) || isDigit(s.Next()) {
		name += string(s.Get())
	}
	return name
}

func toUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	return c
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
