package server

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/gaswelder/ring2/server/smtp"
)

func processSMTP(conn net.Conn, server *Server) {
	s := newSession(conn, server)
	s.send(220, "%s ready", server.config.Hostname)

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
		line, err := s.r.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		debMsg("< " + line)

		cmd, err := parseCommand(line)

		if err != nil {
			s.send(500, err.Error())
			continue
		}

		if cmd.name == "QUIT" {
			s.send(221, "So long, Bob")
			break
		}

		if !processCmd(s, cmd) {
			s.send(500, "Unknown command")
		}
	}

	conn.Close()
	log.Printf("%s disconnected\n", conn.RemoteAddr().String())
}

/*
 * A user session, or context.
 */
type session struct {
	senderHost string
	conn       net.Conn
	r          *bufio.Reader
	draft      *mail
	user       *UserRec
	server     *Server
}

func newSession(conn net.Conn, server *Server) *session {
	s := new(session)
	s.conn = conn
	s.r = bufio.NewReader(s.conn)
	s.server = server
	return s
}

/*
 * Send a formatted response to the client.
 */
func (s *session) send(code int, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	debMsg("> %d %s", code, line)
	fmt.Fprintf(s.conn, "%d %s\r\n", code, line)
}

func (s *session) begin(code int) *smtp.BatchWriter {
	return smtp.NewWriter(code, s.conn)
}
