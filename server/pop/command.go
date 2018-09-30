package pop

import (
	"errors"

	"github.com/gaswelder/ring2/scanner"
)

/*
 * Client's command
 */
type Command struct {
	Name string
	Arg  string
}

/*
 * Parse a command line
 */
func parseCommand(line string) (*Command, error) {
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

	return &Command{name, arg}, nil
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
