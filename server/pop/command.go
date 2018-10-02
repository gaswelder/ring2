package pop

import (
	"errors"

	"github.com/gaswelder/ring2/scanner"
)

/*
 * Client's command
 */
type command struct {
	name string
	arg  string
}

/*
 * Parse a command line
 */
func parseCommand(line string) (*command, error) {
	var name, arg string
	r := scanner.New(line)

	// Command name: a sequence of ASCII alphabetic characters.
	for isAlpha(r.Next()) {
		name += string(r.Get())
	}
	if name == "" {
		return nil, errors.New("command name expected")
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
		return nil, errors.New("<CRLF> expected")
	}

	return &command{name, arg}, nil
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
