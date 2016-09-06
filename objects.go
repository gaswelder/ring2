package main

import (
	"bufio"
	"io/ioutil"
	"net"
)

/*
 * User record
 */
type userRec struct {
	name   string
	remote string
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

/*
 * A user session, or context.
 */
type session struct {
	senderHost string
	conn       net.Conn
	r          *bufio.Reader
	draft      *mail
}

func newSession(conn net.Conn) *session {
	s := new(session)
	s.conn = conn
	s.r = bufio.NewReader(s.conn)
	return s
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

/*
 * User POP session
 */
type popState struct {
	userName string
	user     *userRec
	box      *mailbox
	lastId   int
	conn     net.Conn
	r        *bufio.Reader
}

func newPopSession(c net.Conn) *popState {
	s := new(popState)
	s.conn = c
	s.r = bufio.NewReader(c)
	return s
}
