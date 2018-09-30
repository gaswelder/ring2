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
	inbox    *pop.InboxView
	*pop.ReadWriter
}

func newPopSession(c net.Conn, server *Server) *popState {
	s := new(popState)
	s.ReadWriter = pop.NewReadWriter(c)
	s.server = server
	return s
}

func (s *popState) open(username, password string) error {
	if s.inbox != nil {
		return errors.New("Session already started")
	}

	user := s.server.config.findUser(username, password)
	if user == nil {
		return errors.New("auth failed")
	}
	box, err := s.server.config.mailbox(user)
	if err != nil {
		return err
	}

	m, err := pop.NewInboxView(box)
	if err != nil {
		return err
	}

	s.inbox = m
	return nil
}

func (s *popState) end() error {
	return s.inbox.Commit()
}
