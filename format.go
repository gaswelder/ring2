package main

/*
 * Parsing and formatting functions
 */

import (
	"errors"
	"fmt"
	"net"
	"time"
)

/*
 * Send a formatted response to the client.
 */
func (s *session) send(code int, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	fmt.Fprintf(s.conn, "%d %s\r\n", code, line)
}

type smtpWriter struct {
	code     int
	lastLine string
	conn     net.Conn
}

func (s *session) begin(code int) *smtpWriter {
	w := new(smtpWriter)
	w.code = code
	w.conn = s.conn
	return w
}

func (w *smtpWriter) send(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if w.lastLine != "" {
		fmt.Fprintf(w.conn, "%d-%s\r\n", w.code, w.lastLine)
	}
	w.lastLine = line
}

func (w *smtpWriter) end() {
	if w.lastLine == "" {
		return
	}
	fmt.Fprintf(w.conn, "%d %s\r\n", w.code, w.lastLine)
	w.lastLine = ""
}

/*
 * Parse a command line
 */
func parseCommand(line string) (*command, error) {
	var err error
	var name, arg string

	r := newScanner(line)

	// Command name: a sequence of ASCII alphabetic characters.
	for isAlpha(r.next()) {
		name += string(toUpper(r.get()))
	}

	// If space follows, read the argument
	if r.next() == ' ' {
		r.get()
		for r.more() && r.next() != '\r' {
			arg += string(r.get())
		}
	}

	// Expect "\r\n"
	if r.get() != '\r' || r.get() != '\n' {
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
func parsePath(s string) (*path, error) {

	p := new(path)
	p.hosts = make([]string, 0)

	r := newScanner(s)
	if !r.expect('<') {
		return nil, r.err
	}

	if r.next() == '@' {
		for {
			r.expect('@')
			host := readName(r)
			p.hosts = append(p.hosts, host)

			ch := r.get()
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
	r.expect('@')
	addr += readName(r)
	r.expect('>')

	p.address = addr
	return p, r.err
}

func readName(r *scanner) string {
	name := ""
	for {
		ch := r.next()
		if isAlpha(ch) || isDigit(ch) || ch == '.' || ch == '-' {
			name += string(ch)
			r.get()
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
