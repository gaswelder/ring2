package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func processSMTP(conn net.Conn) {

	log.Printf("%s connected\n", conn.RemoteAddr().String())
	s := newSession(conn)
	s.send(220, "%s ready", config.hostname)

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

		fmt.Print(line)

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
}

func newSession(conn net.Conn) *session {
	s := new(session)
	s.conn = conn
	s.r = bufio.NewReader(s.conn)
	return s
}

/*
 * Send a formatted response to the client.
 */
func (s *session) send(code int, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
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
		fmt.Fprintf(w.conn, "%d-%s\r\n", w.code, w.lastLine)
	}
	w.lastLine = line
}

func (w *smtpWriter) end() {
	if w.lastLine == "" {
		return
	}
	fmt.Fprintf(w.conn, "%d %s\r\n", w.code, w.lastLine)
	w.lastLine = ""
}

func createDir(path string) error {

	stat, err := os.Stat(path)

	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		stat, err = os.Stat(config.maildir)
	}

	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}
