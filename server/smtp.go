package server

import (
	"io"
	"net"

	"github.com/gaswelder/ring2/server/smtp"
)

func processSMTP(conn net.Conn, config *Config) {
	s := newSession(conn, config)
	s.Send(220, "%s ready", config.Hostname)

	/*
	 * Go allows to organize the processing in a linear manner, but the
	 * SMTP standard was written around implementations of that time
	 * which maintained explicit state and thus allowed different
	 * commands like "HELP" to be issued out of context.
	 *
	 * Therefore we read commands here and dispatch them to separate
	 * command functions, passing them a pointer to the current state.
	 */
	for {
		cmd, err := s.ReadCommand()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.Send(500, err.Error())
			continue
		}

		if cmd.Name == "QUIT" {
			s.Send(221, "So long, Bob")
			break
		}

		if !processCmd(s, cmd) {
			s.Send(500, "Unknown command")
		}
	}
}

/*
 * A user session, or context.
 */
type session struct {
	*smtp.ReadWriter
	senderHost string
	draft      *smtp.Mail
	user       *UserRec
	config     *Config
}

func newSession(conn net.Conn, config *Config) *session {
	s := new(session)
	s.ReadWriter = smtp.NewWriter(conn)
	s.config = config
	return s
}
