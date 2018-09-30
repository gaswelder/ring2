package server

import (
	"errors"
	"net"

	"github.com/gaswelder/ring2/server/pop"
)

/*
 * User POP session
 */
type popState struct {
	userName string
	server   *Server
	inbox    *inboxView
	*pop.ReadWriter
}

func newPopSession(c net.Conn, server *Server) *popState {
	s := new(popState)
	s.ReadWriter = pop.NewReadWriter(c)
	s.server = server
	return s
}

func (s *popState) begin(user *UserRec) error {
	if s.inbox != nil {
		return errors.New("Session already started")
	}
	box, err := user.mailbox(s.server.config)
	if err != nil {
		return err
	}

	m, err := newInboxView(box)
	if err != nil {
		return err
	}

	s.inbox = m
	return nil
}
