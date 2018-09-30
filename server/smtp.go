package server

import (
	"log"
	"net"

	"github.com/gaswelder/ring2/server/smtp"
)

func processSMTP(conn net.Conn, server *Server) {
	s := newSession(conn, server)
	s.Send(220, "%s ready", server.config.Hostname)

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

	conn.Close()
	log.Printf("%s disconnected\n", conn.RemoteAddr().String())
}

/*
 * A user session, or context.
 */
type session struct {
	*smtp.ReadWriter
	senderHost string
	draft      *mail
	user       *UserRec
	server     *Server
}

func newSession(conn net.Conn, server *Server) *session {
	s := new(session)
	s.ReadWriter = smtp.NewWriter(conn)
	s.server = server
	return s
}
