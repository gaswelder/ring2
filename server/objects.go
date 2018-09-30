package server

import (
	"io/ioutil"
)

/*
 * User record
 */
type UserRec struct {
	Name     string
	Pwhash   string
	Password string
	Lists    []string
}

/*
 * Client's command
 */
type command struct {
	name string
	arg  string
}

/*
 * Forward or reverse path
 */
type path struct {
	// zero or more lists of hostnames like foo.com
	hosts []string
	// address endpoint, like bob@example.net
	address string
}

/*
 * A mail draft
 */
type mail struct {
	sender     *path
	recipients []*path
}

func newDraft(from *path) *mail {
	return &mail{
		from,
		make([]*path, 0),
	}
}

type message struct {
	id       int
	size     int64
	deleted  bool
	path     string
	filename string
}

// Returns contents of the message
func (m *message) Content() (string, error) {
	v, err := ioutil.ReadFile(m.path)
	if err != nil {
		return "", err
	}
	return string(v), nil
}
