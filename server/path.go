package server

/*
 * Forward or reverse path
 */
type path struct {
	// zero or more lists of hostnames like foo.com
	hosts []string
	// address endpoint, like bob@example.net
	address string
}

func (p *path) format() string {
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
