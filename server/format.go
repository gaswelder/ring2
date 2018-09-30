package server

/*
 * Parsing and formatting functions
 */

import (
	"errors"
	"fmt"
	"time"

	"github.com/gaswelder/ring2/scanner"
)

/*
 * Parse a command line
 */
func parseCommand(line string) (*command, error) {
	var err error
	var name, arg string

	r := scanner.New(line)

	// Command name: a sequence of ASCII alphabetic characters.
	for isAlpha(r.Next()) {
		name += string(toUpper(r.Get()))
	}

	// If space follows, read the argument
	if r.Next() == ' ' {
		r.Get()
		for r.More() && r.Next() != '\r' {
			arg += string(r.Get())
		}
	}

	// Expect "\r\n"
	if r.Get() != '\r' || r.Get() != '\n' {
		err = errors.New("<CRLF> expected")
	}

	if err != nil {
		return nil, err
	}

	return &command{name, arg}, nil
}

func formatDate() string {
	return time.Now().Format(time.RFC822)
}

func formatPath(p *path) string {
	s := "<"

	if len(p.hosts) > 0 {
		for i, host := range p.hosts {
			if i > 0 {
				s += ","
				s += "@" + host
			}
		}
		s += ":"
	}
	s += p.address + ">"
	return s
}

// "<@ONE,@TWO:JOE@THREE>"
// "<joe@three>"
func parsePath(r *scanner.Scanner) (*path, error) {

	p := new(path)
	p.hosts = make([]string, 0)

	if !r.Expect('<') {
		return nil, r.Err()
	}

	if r.Next() == '@' {
		for {
			r.Expect('@')
			host := readName(r)
			p.hosts = append(p.hosts, host)

			ch := r.Get()
			if ch == ',' {
				continue
			}
			if ch == ':' {
				break
			}

			return p, fmt.Errorf("Unexpected character: %c", ch)
		}
	}

	addr := readName(r) + "@"
	r.Expect('@')
	addr += readName(r)
	r.Expect('>')

	p.address = addr
	return p, r.Err()
}

func readName(r *scanner.Scanner) string {
	name := ""
	for {
		ch := r.Next()
		if isAlpha(ch) || isDigit(ch) || ch == '.' || ch == '-' {
			name += string(ch)
			r.Get()
			continue
		}
		break
	}
	return name
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func toUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	return c
}
