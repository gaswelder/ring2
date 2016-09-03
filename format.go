package main

/*
 * Parsing and formatting functions
 */
 
import (
	"fmt"
)


/*
 * Send a formatted response to the client.
 */
func (s *session) send(code int, format string, args... interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintf(s.conn, "%d %s\r\n", code, line)
}

/*
 * Read a command from the client.
 */
func (s *session) readCommand() (*command, error) {
	var err error
	var ch byte
	var name, arg string

	line, err := s.r.ReadString('\n')

	r := newScanner(line)

	// Command name: a sequence of ASCII alphabetic characters.
	for isAlpha(r.next()) {
		name += toUpper(r.get())
	}

	// If space follows, read the argument
	if r.next() == ' ' {
		for r.more() && r.next() != '\r' {
			arg += r.get()
		}
	}

	// Expect "\r\n"
	if r.Get() != '\r' || r.Get() != '\n' {
		err = errors.New("<CRLF> expected")
	}

	if r.err() != nil {
		err = s.r.err()
	}

	if err != nil {
		return nil, err
	}

	return &command{name, arg}, nil
}

func formatDate() string {
	return time.Now().Format(RFC822)
}

func formatPath(p *path) string {
	s := "<"

	if len(p.hosts) > 0 {
		for i, host := range(p.hosts) {
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
func parsePath(s string) (*path, error) {

	p := new(path)
	p.hosts = make([]string, 0)

	r := newScanner(s)
	r.expect('<')
	if r.next() == '@' {
		for {
			r.expect('@')
			host := readName(r)
			p.hosts = append(p.hosts, host)

			ch = r.get()
			if ch == ',' {
				continue
			}
			if ch == ':' {
				break
			}

			return p, fmt.Errorf("Unexpected character: %c", ch)
		}
	}

	addr := readName(p) + "@"
	p.expect('@')
	addr += readName(p)
	r.expect('>')

	path.address = addr
	return path, p.err()
}

func readName(r *scanner) string {
	name := ""
	for {
		ch := r.next()
		if isAlpha(ch) || isDigit(ch) || ch == '.' || ch == '-' {
			name += ch
			r.get()
		}
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
