package smtp

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const AuthOK = 235
const ParameterSyntaxError = 501
const BadSequenceOfCommands = 503
const ParameterNotImplemented = 504
const AuthInvalid = 535

type ReadWriter struct {
	conn net.Conn
	r    *bufio.Reader
}

func NewWriter(conn net.Conn) *ReadWriter {
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
	conn     net.Conn
}

func newBatchWriter(code int, conn net.Conn) *BatchWriter {
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
