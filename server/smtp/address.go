package smtp

import (
	"errors"
	"strings"
)

type Address struct {
	Name string
	Host string
}

func (a *Address) Format() string {
	return a.Name + "@" + a.Host
}

func parseAddress(addr string) (*Address, error) {
	pos := strings.Index(addr, "@")
	if pos < 0 {
		return nil, errors.New("Invalid email address")
	}
	name := addr[:pos]
	host := addr[pos+1:]
	return &Address{name, host}, nil
}
