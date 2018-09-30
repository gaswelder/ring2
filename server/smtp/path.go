package smtp

import (
	"fmt"

	"github.com/gaswelder/ring2/scanner"
)

/*
 * Forward or reverse path
 */
type Path struct {
	// zero or more lists of hostnames like foo.com
	Hosts []string
	// address endpoint, like bob@example.net
	Addr *Address
}

func (p *Path) Format() string {
	s := "<"
	if len(p.Hosts) > 0 {
		for i, host := range p.Hosts {
			if i > 0 {
				s += ","
				s += "@" + host
			}
		}
		s += ":"
	}
	s += p.Addr.Format() + ">"
	return s
}

// "<@ONE,@TWO:JOE@THREE>"
// "<joe@three>"
func ParsePath(r *scanner.Scanner) (*Path, error) {

	p := new(Path)
	p.Hosts = make([]string, 0)

	if !r.Expect('<') {
		return nil, r.Err()
	}

	if r.Next() == '@' {
		for {
			r.Expect('@')
			host := readName(r)
			p.Hosts = append(p.Hosts, host)

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

	user := readName(r) + "@"
	r.Expect('@')
	host := readName(r)
	r.Expect('>')

	p.Addr = &Address{user, host}
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

// func isAlpha(c byte) bool {
// 	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
// }

// func toUpper(c byte) byte {
// 	if c >= 'a' && c <= 'z' {
// 		c -= 'a' - 'A'
// 	}
// 	return c
// }
