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

func (s *popState) SetUserName(name string) error {
	if s.inbox != nil {
		return errors.New("already authorized")
	}
	if name == "" {
		return errors.New("empty username")
	}
	s.userName = name
	return nil
}

func (s *popState) Inbox() *pop.InboxView {
	return s.inbox
}

func (s *popState) Open(password string) error {
	if s.inbox != nil {
		return errors.New("Session already started")
	}
	if s.userName == "" {
		return errors.New("Wrong commands order")
	}

	user := s.config.findUser(s.userName, password)
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

func (s *popState) Close() error {
	return s.inbox.Commit()
}
