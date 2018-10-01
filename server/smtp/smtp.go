package smtp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gaswelder/ring2/server/mailbox"
)

const AuthOK = 235
const ParameterSyntaxError = 501
const BadSequenceOfCommands = 503
const ParameterNotImplemented = 504
const AuthInvalid = 535

type AuthFunc func(name, password string) error
type MailboxLookupFunc func(name string) ([]*mailbox.Mailbox, error)

type session struct {
	*ReadWriter
	senderHost string
	draft      *Mail
	auth       bool
	authorize  AuthFunc
	lookup     MailboxLookupFunc
	recipients []*mailbox.Mailbox
}

type ReadWriter struct {
	conn io.Writer
	r    *bufio.Reader
}

func NewWriter(conn io.ReadWriter) *ReadWriter {
	return &ReadWriter{
		conn: conn,
		r:    bufio.NewReader(conn),
	}
}

func (w *ReadWriter) ReadCommand() (*Command, error) {
	line, err := w.r.ReadString('\n')
	log.Println(line)
	if err != nil {
		return nil, err
	}
	// debMsg("< " + line)
	return parseCommand(line)
}

func (w *ReadWriter) ReadLine() (string, error) {
	line, err := w.r.ReadString('\n')
	log.Println(line)
	return line, err
}

func (w *ReadWriter) Send(code int, format string, args ...interface{}) {
	line := fmt.Sprintf("%d %s", code, fmt.Sprintf(format, args...))
	log.Println(line)
	fmt.Fprintf(w.conn, "%s\r\n", line)
}

func (w *ReadWriter) BeginBatch(code int) *BatchWriter {
	return newBatchWriter(code, w.conn)
}

type BatchWriter struct {
	code     int
	lastLine string
	conn     io.Writer
}

func newBatchWriter(code int, conn io.Writer) *BatchWriter {
	w := new(BatchWriter)
	w.code = code
	w.conn = conn
	return w
}

func (w *BatchWriter) Send(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if w.lastLine != "" {
		// debMsg("> %d-%s", w.code, w.lastLine)
		log.Printf("%d-%s\r\n", w.code, w.lastLine)
		fmt.Fprintf(w.conn, "%d-%s\r\n", w.code, w.lastLine)
	}
	w.lastLine = line
}

func (w *BatchWriter) End() {
	if w.lastLine == "" {
		return
	}
	// debMsg("> %d %s", w.code, w.lastLine)
	fmt.Fprintf(w.conn, "%d %s\r\n", w.code, w.lastLine)
	w.lastLine = ""
}

func Process(conn io.ReadWriter, auth AuthFunc, lookup MailboxLookupFunc) {
	s := &session{
		ReadWriter: NewWriter(conn),
		authorize:  auth,
		lookup:     lookup,
		recipients: make([]*mailbox.Mailbox, 0),
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("couldn't get hostname: %s", err.Error())
		hostname = "localhost"
	}
	s.Send(220, "%s ready", hostname)

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
