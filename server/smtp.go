package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const smtpAuthOK = 235
const smtpParameterSyntaxError = 501
const smtpBadSequenceOfCommands = 503
const smtpParameterNotImplemented = 504
const smtpAuthInvalid = 535

func processSMTP(conn net.Conn, config *Config) {
	s := newSession(conn, config)
	s.send(220, "%s ready", config.Hostname)

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
	config     *Config
}

func newSession(conn net.Conn, config *Config) *session {
	s := new(session)
	s.conn = conn
	s.r = bufio.NewReader(s.conn)
	s.config = config
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

type smtpWriter struct {
	code     int
	lastLine string
	conn     net.Conn
}

func (s *session) begin(code int) *smtpWriter {
	w := new(smtpWriter)
	w.code = code
	w.conn = s.conn
	return w
}

func (w *smtpWriter) send(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if w.lastLine != "" {
		debMsg("> %d-%s", w.code, w.lastLine)
		fmt.Fprintf(w.conn, "%d-%s\r\n", w.code, w.lastLine)
	}
	w.lastLine = line
}

func (w *smtpWriter) end() {
	if w.lastLine == "" {
		return
	}
	debMsg("> %d %s", w.code, w.lastLine)
	fmt.Fprintf(w.conn, "%d %s\r\n", w.code, w.lastLine)
	w.lastLine = ""
}