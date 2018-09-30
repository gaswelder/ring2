package smtp

import (
	"fmt"
	"net"
)

const AuthOK = 235
const ParameterSyntaxError = 501
const BadSequenceOfCommands = 503
const ParameterNotImplemented = 504
const AuthInvalid = 535

type Writer struct {
	conn net.Conn
}

func NewWriter(conn net.Conn) *Writer {
	return &Writer{conn}
}

func (w *Writer) Send(code int, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	// debMsg("> %d %s", code, line)
	fmt.Fprintf(w.conn, "%d %s\r\n", code, line)
}

func (w *Writer) BeginBatch(code int) *BatchWriter {
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
