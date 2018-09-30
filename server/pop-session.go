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
	config   *Config
	inbox    *pop.InboxView
	*pop.ReadWriter
}

func newPopSession(c net.Conn, config *Config) *popState {
	s := new(popState)
	s.ReadWriter = pop.NewReadWriter(c)
	s.config = config
	return s
}

func (s *popState) open(username, password string) error {
	if s.inbox != nil {
		return errors.New("Session already started")
	}

	user := s.config.findUser(username, password)
	if user == nil {
		return errors.New("auth failed")
	}
	box, err := s.config.mailbox(user)
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
